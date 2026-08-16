package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductCategory struct {
	ID          uuid.UUID `db:"id"          json:"id"`
	Name        string    `db:"name"        json:"name"`
	Slug        string    `db:"slug"        json:"slug"`
	Description *string   `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

// Product adalah master katalog. Harga yang benar-benar dijual ke customer
// ditentukan per trip lewat TripItem, karena kurs dan harga toko berubah tiap
// perjalanan.
type Product struct {
	ID           uuid.UUID       `db:"id"            json:"id"`
	SKU          string          `db:"sku"           json:"sku"`
	Name         string          `db:"name"          json:"name"`
	CategoryID   *uuid.UUID      `db:"category_id"   json:"category_id"`
	Brand        *string         `db:"brand"         json:"brand"`
	StoreName    *string         `db:"store_name"    json:"store_name"`
	BaseCurrency string          `db:"base_currency" json:"base_currency"`
	BasePrice    decimal.Decimal `db:"base_price"    json:"base_price"`
	MarkupType   string          `db:"markup_type"   json:"markup_type"`
	MarkupValue  decimal.Decimal `db:"markup_value"  json:"markup_value"`
	WeightGram   int             `db:"weight_gram"   json:"weight_gram"`
	ImageURL     *string         `db:"image_url"     json:"image_url"`
	Notes        *string         `db:"notes"         json:"notes"`
	IsActive     bool            `db:"is_active"     json:"is_active"`
	CreatedAt    time.Time       `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"    json:"updated_at"`
	DeletedAt    *time.Time      `db:"deleted_at"    json:"-"`
}

// ProductWithCategory dipakai pada endpoint daftar produk supaya UI tidak perlu
// memanggil endpoint kategori satu per satu.
type ProductWithCategory struct {
	Product
	CategoryName *string `db:"category_name" json:"category_name"`
}

// ProductPriceHistory adalah satu baris riwayat harga sebuah produk pada satu
// trip: berapa harga modal yang dipasang di katalog, berapa yang benar-benar
// dibayar tripper di kasir, dan berapa yang laku terjual.
//
// Inilah yang membuat produk bisa dipakai ulang antar trip tanpa mencatat harga
// dari nol. Harga modal sengaja tidak disalin otomatis ke trip berikutnya:
// negara asal dan kurs bisa berbeda, jadi angka lama adalah acuan, bukan
// jawaban.
type ProductPriceHistory struct {
	TripID       uuid.UUID       `db:"trip_id"        json:"trip_id"`
	TripCode     string          `db:"trip_code"      json:"trip_code"`
	TripTitle    string          `db:"trip_title"     json:"trip_title"`
	Country      string          `db:"country"        json:"country"`
	DepartDate   time.Time       `db:"depart_date"    json:"depart_date"`
	Currency     string          `db:"currency"       json:"currency"`
	ExchangeRate decimal.Decimal `db:"exchange_rate"  json:"exchange_rate"`

	// Harga yang dipasang admin di katalog trip, dalam mata uang trip.
	CatalogCost    decimal.Decimal `db:"catalog_cost"     json:"catalog_cost"`
	CatalogCostIDR decimal.Decimal `db:"catalog_cost_idr" json:"catalog_cost_idr"`
	SellPrice      decimal.Decimal `db:"sell_price"       json:"sell_price"`

	// Harga rata-rata yang benar-benar dibayar tripper. Kosong kalau produk ini
	// masuk katalog tapi belum sempat dibeli.
	ActualCost    decimal.Decimal `db:"actual_cost"     json:"actual_cost"`
	ActualCostIDR decimal.Decimal `db:"actual_cost_idr" json:"actual_cost_idr"`

	QtyPurchased int `db:"qty_purchased" json:"qty_purchased"`
	QtySold      int `db:"qty_sold"      json:"qty_sold"`
}
