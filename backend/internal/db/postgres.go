// Package db menangani koneksi PostgreSQL dan helper transaksi.
//
// Seluruh repository menerima Querier, bukan *pgxpool.Pool, sehingga fungsi
// repository yang sama bisa dipakai di dalam maupun di luar transaksi. Service
// yang perlu menulis ke beberapa tabel sekaligus membungkusnya dengan WithTx.
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ipoool/jastipin/backend/internal/config"
)

// Querier dipenuhi oleh *pgxpool.Pool maupun pgx.Tx.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Connect membuka connection pool dan memastikan database benar-benar bisa
// dihubungi sebelum aplikasi dinyatakan siap.
func Connect(ctx context.Context, cfg config.DB) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = time.Minute

	// Daftarkan tipe NUMERIC <-> decimal.Decimal pada setiap koneksi baru.
	// Tanpa ini semua nominal uang harus di-scan sebagai string secara manual.
	poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("buat connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// WithTx menjalankan fn di dalam satu transaksi. Transaksi di-rollback jika fn
// mengembalikan error atau panic, dan di-commit jika fn sukses.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("mulai transaksi: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback dulu, baru teruskan panic ke atas supaya koneksi tidak
			// tertinggal dalam keadaan transaksi terbuka.
			_ = tx.Rollback(context.WithoutCancel(ctx))
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		// Gunakan context tanpa cancel supaya rollback tetap terkirim walau
		// context request sudah keburu dibatalkan.
		if rbErr := tx.Rollback(context.WithoutCancel(ctx)); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			return fmt.Errorf("%w (rollback gagal: %v)", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaksi: %w", err)
	}
	return nil
}
