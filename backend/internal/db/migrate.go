package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// advisoryLockID mengunci proses migrasi antar instance. Angkanya arbitrer,
// yang penting konsisten: kalau dua container start bersamaan, hanya satu yang
// menjalankan migrasi dan yang lain menunggu.
const advisoryLockID int64 = 8_412_776_301

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// LoadMigrations membaca file migrasi yang ter-embed di binary. Penamaan file
// wajib mengikuti pola: 000001_nama_migrasi.up.sql / .down.sql
func LoadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("baca direktori migrasi: %w", err)
	}

	byVersion := map[int]*Migration{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		version, name, direction, err := parseMigrationName(e.Name())
		if err != nil {
			return nil, err
		}

		content, err := migrationFS.ReadFile(path.Join("migrations", e.Name()))
		if err != nil {
			return nil, fmt.Errorf("baca %s: %w", e.Name(), err)
		}

		m, ok := byVersion[version]
		if !ok {
			m = &Migration{Version: version, Name: name}
			byVersion[version] = m
		}
		if direction == "up" {
			m.Up = string(content)
		} else {
			m.Down = string(content)
		}
	}

	out := make([]Migration, 0, len(byVersion))
	for _, m := range byVersion {
		if strings.TrimSpace(m.Up) == "" {
			return nil, fmt.Errorf("migrasi versi %d (%s) tidak punya file .up.sql", m.Version, m.Name)
		}
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

func parseMigrationName(filename string) (version int, name, direction string, err error) {
	base := strings.TrimSuffix(filename, ".sql")

	switch {
	case strings.HasSuffix(base, ".up"):
		direction = "up"
		base = strings.TrimSuffix(base, ".up")
	case strings.HasSuffix(base, ".down"):
		direction = "down"
		base = strings.TrimSuffix(base, ".down")
	default:
		return 0, "", "", fmt.Errorf("nama migrasi %q harus berakhiran .up.sql atau .down.sql", filename)
	}

	prefix, rest, found := strings.Cut(base, "_")
	if !found {
		return 0, "", "", fmt.Errorf("nama migrasi %q harus berpola <versi>_<nama>.<arah>.sql", filename)
	}

	version, err = strconv.Atoi(prefix)
	if err != nil {
		return 0, "", "", fmt.Errorf("versi pada %q bukan angka: %w", filename, err)
	}
	return version, rest, direction, nil
}

func ensureMigrationTable(ctx context.Context, q Querier) error {
	_, err := q.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     INTEGER     PRIMARY KEY,
			name        TEXT        NOT NULL,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("buat tabel schema_migrations: %w", err)
	}
	return nil
}

func appliedVersions(ctx context.Context, q Querier) (map[int]bool, error) {
	rows, err := q.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("baca schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// MigrateUp menjalankan seluruh migrasi yang belum diterapkan, satu transaksi
// per migrasi sehingga migrasi yang gagal tidak meninggalkan skema separuh jadi.
func MigrateUp(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	migrations, err := LoadMigrations()
	if err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("ambil koneksi: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockID); err != nil {
		return fmt.Errorf("ambil advisory lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, advisoryLockID)
	}()

	if err := ensureMigrationTable(ctx, conn); err != nil {
		return err
	}
	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}

	count := 0
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("mulai transaksi migrasi %d: %w", m.Version, err)
		}
		if _, err := tx.Exec(ctx, m.Up); err != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			return fmt.Errorf("jalankan migrasi %06d_%s: %w", m.Version, m.Name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.Version, m.Name); err != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			return fmt.Errorf("catat migrasi %d: %w", m.Version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migrasi %d: %w", m.Version, err)
		}

		log.Info("migrasi diterapkan", "version", m.Version, "name", m.Name)
		count++
	}

	if count == 0 {
		log.Info("skema database sudah terbaru")
	} else {
		log.Info("migrasi selesai", "jumlah", count)
	}
	return nil
}

// MigrateDown membatalkan n migrasi terakhir. n <= 0 berarti membatalkan semua.
func MigrateDown(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger, n int) error {
	migrations, err := LoadMigrations()
	if err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("ambil koneksi: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockID); err != nil {
		return fmt.Errorf("ambil advisory lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, advisoryLockID)
	}()

	if err := ensureMigrationTable(ctx, conn); err != nil {
		return err
	}
	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}

	// Turun dari versi tertinggi ke terendah.
	rolled := 0
	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		if !applied[m.Version] {
			continue
		}
		if n > 0 && rolled >= n {
			break
		}
		if strings.TrimSpace(m.Down) == "" {
			return fmt.Errorf("migrasi %06d_%s tidak punya file .down.sql, tidak bisa di-rollback", m.Version, m.Name)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("mulai transaksi rollback %d: %w", m.Version, err)
		}
		if _, err := tx.Exec(ctx, m.Down); err != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			return fmt.Errorf("rollback migrasi %06d_%s: %w", m.Version, m.Name, err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, m.Version); err != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			return fmt.Errorf("hapus catatan migrasi %d: %w", m.Version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit rollback %d: %w", m.Version, err)
		}

		log.Info("migrasi di-rollback", "version", m.Version, "name", m.Name)
		rolled++
	}

	if rolled == 0 {
		log.Info("tidak ada migrasi yang perlu di-rollback")
	}
	return nil
}

// CurrentVersion mengembalikan versi migrasi tertinggi yang sudah diterapkan.
func CurrentVersion(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	if err := ensureMigrationTable(ctx, pool); err != nil {
		return 0, err
	}

	var version int
	err := pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	return version, nil
}
