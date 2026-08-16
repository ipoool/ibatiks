package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
)

const shippingRateColumns = `id, courier, service, destination_city, province, price_per_kg,
	                         min_weight_gram, etd, created_at, updated_at`

type ShippingRepo struct{}

func NewShippingRepo() *ShippingRepo { return &ShippingRepo{} }

type ShippingRateParams struct {
	Courier         string
	Service         string
	DestinationCity string
	Province        *string
	PricePerKg      decimal.Decimal
	MinWeightGram   int
	ETD             *string
}

// FindRate mencari tarif untuk satu kota tujuan. Nama kota dinormalkan lebih
// dulu agar "Kota Jakarta Selatan" dan "jakarta selatan" sama-sama ketemu.
//
// Mengembalikan nil tanpa error kalau kotanya belum terdaftar; pemanggil yang
// memutuskan memakai tarif default.
func (r *ShippingRepo) FindRate(ctx context.Context, q db.Querier, courier, service, city string) (*domain.ShippingRate, error) {
	rates, err := collectRows[domain.ShippingRate](ctx, q, `
		SELECT `+shippingRateColumns+`
		FROM shipping_rates
		WHERE courier = $1 AND service = $2 AND destination_city = $3
		LIMIT 1`,
		courier, service, domain.NormalizeCity(city))
	if err != nil {
		return nil, err
	}
	if len(rates) == 0 {
		return nil, nil
	}
	return &rates[0], nil
}

func (r *ShippingRepo) List(ctx context.Context, q db.Querier, courier, search string) ([]domain.ShippingRate, error) {
	conditions := []string{}
	args := []any{}

	if courier != "" {
		args = append(args, courier)
		conditions = append(conditions, fmt.Sprintf("courier = $%d", len(args)))
	}
	if search != "" {
		args = append(args, "%"+domain.NormalizeCity(search)+"%")
		conditions = append(conditions, fmt.Sprintf("destination_city LIKE $%d", len(args)))
	}

	return collectRows[domain.ShippingRate](ctx, q,
		`SELECT `+shippingRateColumns+` FROM shipping_rates`+buildWhere(conditions)+
			` ORDER BY destination_city ASC, courier ASC, service ASC`, args...)
}

// Upsert menyimpan tarif. Kombinasi kurir, layanan, dan kota bersifat unik,
// jadi menyimpan kota yang sama dua kali berarti memperbarui tarifnya.
func (r *ShippingRepo) Upsert(ctx context.Context, q db.Querier, p ShippingRateParams) (*domain.ShippingRate, error) {
	return collectOne[domain.ShippingRate](ctx, q, "tarif kirim", `
		INSERT INTO shipping_rates (courier, service, destination_city, province,
		                            price_per_kg, min_weight_gram, etd)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (courier, service, destination_city) DO UPDATE
		SET province        = EXCLUDED.province,
		    price_per_kg    = EXCLUDED.price_per_kg,
		    min_weight_gram = EXCLUDED.min_weight_gram,
		    etd             = EXCLUDED.etd
		RETURNING `+shippingRateColumns,
		p.Courier, p.Service, domain.NormalizeCity(p.DestinationCity), p.Province,
		p.PricePerKg, p.MinWeightGram, p.ETD)
}

func (r *ShippingRepo) Delete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "tarif kirim", `DELETE FROM shipping_rates WHERE id = $1`, id)
}
