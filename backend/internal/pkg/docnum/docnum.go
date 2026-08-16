// Package docnum menghasilkan nomor dokumen berurutan per tahun
// (TRP-2026-001, ORD-2026-0001, INV-2026-0001, CUS-0001).
package docnum

import (
	"context"
	"fmt"

	"github.com/ipoool/jastipin/backend/internal/db"
)

// Type mendefinisikan awalan dan lebar padding sebuah jenis dokumen.
type Type struct {
	Prefix string
	Width  int
	// PerYear false berarti nomor berjalan terus tanpa reset tiap tahun dan
	// tahun tidak ikut dicetak (dipakai untuk kode customer).
	PerYear bool
}

var (
	Trip     = Type{Prefix: "TRP", Width: 3, PerYear: true}
	Order    = Type{Prefix: "ORD", Width: 4, PerYear: true}
	Invoice  = Type{Prefix: "INV", Width: 4, PerYear: true}
	Customer = Type{Prefix: "CUS", Width: 4, PerYear: false}
)

// Next mengambil nomor urut berikutnya dan merangkainya jadi nomor dokumen.
//
// Panggil ini di dalam transaksi yang sama dengan INSERT dokumennya. Upsert di
// bawah mengunci baris counter sampai transaksi selesai, sehingga dua request
// bersamaan tidak mungkin mendapat nomor yang sama; dan kalau transaksi
// di-rollback, nomornya ikut dikembalikan sehingga tidak ada nomor bolong.
func Next(ctx context.Context, q db.Querier, t Type, year int) (string, error) {
	counterYear := year
	if !t.PerYear {
		// Counter tanpa reset tahunan tetap butuh satu baris; tahun 0 dipakai
		// sebagai penanda "berlaku sepanjang masa".
		counterYear = 0
	}

	var seq int
	err := q.QueryRow(ctx, `
		INSERT INTO document_counters (doc_type, year, last_number)
		VALUES ($1, $2, 1)
		ON CONFLICT (doc_type, year)
		DO UPDATE SET last_number = document_counters.last_number + 1,
		              updated_at  = now()
		RETURNING last_number`,
		t.Prefix, counterYear,
	).Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("ambil nomor %s: %w", t.Prefix, err)
	}

	if t.PerYear {
		return fmt.Sprintf("%s-%d-%0*d", t.Prefix, year, t.Width, seq), nil
	}
	return fmt.Sprintf("%s-%0*d", t.Prefix, t.Width, seq), nil
}
