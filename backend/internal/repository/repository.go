// Package repository berisi akses data memakai pgx v5 secara langsung.
//
// Tidak ada ORM di sini: setiap query ditulis sebagai SQL biasa, lalu hasilnya
// dipetakan ke struct dengan pgx.RowToStructByName yang mencocokkan nama kolom
// dengan tag `db` pada struct domain. Konsekuensinya, daftar kolom pada SELECT
// harus persis sama dengan field struct tujuannya.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
)

// Kode error PostgreSQL yang perlu dibedakan penanganannya.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
)

// collectRows menjalankan query dan memetakan seluruh baris ke []T.
func collectRows[T any](ctx context.Context, q db.Querier, sql string, args ...any) ([]T, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, wrapPgError(err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, wrapPgError(err)
	}
	return out, nil
}

// collectRowsLax sama seperti collectRows tapi mengizinkan struct punya field
// yang tidak ada kolomnya di hasil query. Dipakai untuk struct gabungan yang
// sebagian fieldnya diisi terpisah.
func collectRowsLax[T any](ctx context.Context, q db.Querier, sql string, args ...any) ([]T, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, wrapPgError(err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, wrapPgError(err)
	}
	return out, nil
}

// collectOne mengembalikan tepat satu baris. Kalau tidak ada, error-nya adalah
// domain.NotFound dengan nama entitas yang diberikan.
func collectOne[T any](ctx context.Context, q db.Querier, entity, sql string, args ...any) (*T, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, wrapPgError(err)
	}
	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound(entity)
		}
		return nil, wrapPgError(err)
	}
	return &row, nil
}

func collectOneLax[T any](ctx context.Context, q db.Querier, entity, sql string, args ...any) (*T, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, wrapPgError(err)
	}
	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound(entity)
		}
		return nil, wrapPgError(err)
	}
	return &row, nil
}

// exec menjalankan perintah dan mengembalikan jumlah baris yang terpengaruh.
func exec(ctx context.Context, q db.Querier, sql string, args ...any) (int64, error) {
	tag, err := q.Exec(ctx, sql, args...)
	if err != nil {
		return 0, wrapPgError(err)
	}
	return tag.RowsAffected(), nil
}

// execExpectOne menjalankan UPDATE/DELETE yang seharusnya mengenai satu baris,
// dan mengubah "tidak ada baris terkena" menjadi error not found.
func execExpectOne(ctx context.Context, q db.Querier, entity, sql string, args ...any) error {
	affected, err := exec(ctx, q, sql, args...)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.NotFound(entity)
	}
	return nil
}

// wrapPgError menerjemahkan pelanggaran constraint database menjadi error
// domain yang enak dibaca pengguna. Nama constraint dipakai untuk menebak
// pesannya karena constraint-lah yang paling tahu aturan apa yang dilanggar.
func wrapPgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case pgUniqueViolation:
		return domain.Conflict("%s", uniqueMessage(pgErr.ConstraintName))
	case pgForeignKeyViolation:
		return domain.Conflict("%s", foreignKeyMessage(pgErr.ConstraintName))
	case pgCheckViolation:
		return domain.Validationf("%s", checkMessage(pgErr.ConstraintName))
	default:
		return err
	}
}

func uniqueMessage(constraint string) string {
	switch constraint {
	case "users_email_key":
		return "email sudah dipakai pengguna lain"
	case "customers_code_key":
		return "kode customer sudah dipakai"
	case "idx_customers_phone_unique":
		return "nomor WhatsApp ini sudah terdaftar pada customer lain — pakai customer yang sudah ada supaya laporan per customer tidak terpecah"
	case "products_sku_key":
		return "SKU produk sudah dipakai"
	case "product_categories_slug_key":
		return "slug kategori sudah dipakai"
	case "trips_code_key":
		return "kode trip sudah dipakai"
	case "trip_items_unique_product":
		return "produk ini sudah ada di katalog trip"
	case "orders_order_number_key":
		return "nomor order sudah dipakai"
	case "invoices_invoice_number_key":
		return "nomor invoice sudah dipakai"
	case "shipments_order_id_key":
		return "order ini sudah punya data pengiriman"
	case "stock_items_product_id_key":
		return "produk ini sudah punya catatan stok"
	default:
		return "data yang sama sudah ada"
	}
}

func foreignKeyMessage(constraint string) string {
	switch {
	case strings.Contains(constraint, "product_id"):
		return "produk masih dipakai transaksi lain atau tidak ditemukan"
	case strings.Contains(constraint, "customer_id"):
		return "customer masih punya order atau tidak ditemukan"
	case strings.Contains(constraint, "trip_id"):
		return "trip masih punya data terkait atau tidak ditemukan"
	case strings.Contains(constraint, "order_id"):
		return "order tidak ditemukan"
	default:
		return "data terkait masih dipakai atau tidak ditemukan"
	}
}

func checkMessage(constraint string) string {
	switch constraint {
	case "trips_date_order":
		return "tanggal pulang tidak boleh lebih awal dari tanggal berangkat"
	case "shipments_tracking_required":
		return "nomor resi wajib diisi sebelum status pengiriman diubah menjadi terkirim"
	default:
		return fmt.Sprintf("nilai yang dikirim melanggar aturan %s", constraint)
	}
}

// buildWhere merangkai potongan kondisi menjadi klausa WHERE.
// Potongan kosong diabaikan supaya pemanggil bisa menambahkan filter secara
// kondisional tanpa mengurus tanda hubungnya.
func buildWhere(conditions []string) string {
	filtered := conditions[:0]
	for _, c := range conditions {
		if strings.TrimSpace(c) != "" {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(filtered, " AND ")
}
