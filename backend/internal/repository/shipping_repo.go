package repository

import (
	"context"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
)

// ShippingRepo menyimpan pemetaan alamat ke ID tujuan milik RajaOngkir.
//
// Dulu ia juga memegang tabel tarif per kota yang diisi sendiri oleh toko.
// Tabel itu dilepas: satu-satunya sumber ongkir sekarang adalah kurir, dan
// tarif yang diisi tangan selalu tertinggal dari yang sebenarnya berlaku.
type ShippingRepo struct{}

func NewShippingRepo() *ShippingRepo { return &ShippingRepo{} }

// --- Cache tujuan RajaOngkir ------------------------------------------------

const shippingDestinationColumns = `id, query, destination_id, label, city_name,
	                                province_name, zip_code, created_at, updated_at`

// FindDestination membaca hasil pencarian tujuan yang sudah pernah disimpan.
// Mengembalikan nil tanpa error kalau belum ada, supaya pemanggil tahu harus
// bertanya ke RajaOngkir.
func (r *ShippingRepo) FindDestination(ctx context.Context, q db.Querier, query string) (*domain.ShippingDestination, error) {
	found, err := collectRows[domain.ShippingDestination](ctx, q,
		`SELECT `+shippingDestinationColumns+` FROM shipping_destinations WHERE query = $1 LIMIT 1`,
		domain.NormalizeCity(query))
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	return &found[0], nil
}

// SaveDestination menyimpan hasil pencarian supaya alamat yang sama tidak
// memicu panggilan API berulang.
func (r *ShippingRepo) SaveDestination(ctx context.Context, q db.Querier, in domain.ShippingDestination) (*domain.ShippingDestination, error) {
	return collectOne[domain.ShippingDestination](ctx, q, "tujuan kirim", `
		INSERT INTO shipping_destinations (query, destination_id, label, city_name, province_name, zip_code)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (query) DO UPDATE SET
			destination_id = EXCLUDED.destination_id,
			label          = EXCLUDED.label,
			city_name      = EXCLUDED.city_name,
			province_name  = EXCLUDED.province_name,
			zip_code       = EXCLUDED.zip_code
		RETURNING `+shippingDestinationColumns,
		domain.NormalizeCity(in.Query), in.DestinationID, in.Label,
		in.CityName, in.ProvinceName, in.ZipCode)
}
