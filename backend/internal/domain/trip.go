package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Status perjalanan, berurutan dari perencanaan sampai selesai dibukukan.
const (
	TripDraft     = "draft"      // masih disusun, belum dibuka untuk order
	TripOpen      = "open"       // sudah diposting, order boleh masuk
	TripClosed    = "closed"     // pendaftaran order ditutup
	TripShopping  = "shopping"   // tripper sedang belanja di negara tujuan
	TripInTransit = "in_transit" // tripper dalam perjalanan pulang
	TripArrived   = "arrived"    // barang sudah sampai Indonesia
	TripSettled   = "settled"    // semua order selesai, profit sudah dibukukan
	TripCancelled = "cancelled"
)

// tripTransitions memetakan status trip ke daftar status berikutnya yang sah.
// Peta ini satu-satunya sumber kebenaran alur trip.
var tripTransitions = map[string][]string{
	TripDraft:     {TripOpen, TripCancelled},
	TripOpen:      {TripClosed, TripShopping, TripCancelled},
	TripClosed:    {TripShopping, TripOpen, TripCancelled},
	TripShopping:  {TripInTransit, TripArrived, TripCancelled},
	TripInTransit: {TripArrived},
	TripArrived:   {TripSettled},
	TripSettled:   {},
	TripCancelled: {},
}

func IsValidTripStatus(status string) bool {
	_, ok := tripTransitions[status]
	return ok
}

// CanTransitionTrip menjawab apakah trip boleh berpindah dari satu status ke
// status lain.
func CanTransitionTrip(from, to string) bool {
	for _, allowed := range tripTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// NextTripStatuses dipakai UI untuk menampilkan tombol aksi yang relevan saja.
func NextTripStatuses(from string) []string {
	next := tripTransitions[from]
	out := make([]string, len(next))
	copy(out, next)
	return out
}

// TripAcceptsOrder menentukan apakah order baru masih boleh dicatat.
// Order tetap diizinkan saat status shopping karena di lapangan sering ada
// customer yang menyusul titip saat tripper sudah di toko.
func TripAcceptsOrder(status string) bool {
	return status == TripOpen || status == TripShopping
}

type Trip struct {
	ID            uuid.UUID       `db:"id"              json:"id"`
	Code          string          `db:"code"            json:"code"`
	Title         string          `db:"title"           json:"title"`
	Country       string          `db:"country"         json:"country"`
	City          *string         `db:"city"            json:"city"`
	TripperUserID *uuid.UUID      `db:"tripper_user_id" json:"tripper_user_id"`
	DepartDate    time.Time       `db:"depart_date"     json:"depart_date"`
	ReturnDate    time.Time       `db:"return_date"     json:"return_date"`
	OrderDeadline *time.Time      `db:"order_deadline"  json:"order_deadline"`
	Currency      string          `db:"currency"        json:"currency"`
	ExchangeRate  decimal.Decimal `db:"exchange_rate"   json:"exchange_rate"`
	Status        string          `db:"status"          json:"status"`
	Notes         *string         `db:"notes"           json:"notes"`
	CreatedBy     *uuid.UUID      `db:"created_by"      json:"created_by"`
	CreatedAt     time.Time       `db:"created_at"      json:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"      json:"updated_at"`
}

// TripDetail menambahkan data turunan yang selalu dibutuhkan halaman detail.
type TripDetail struct {
	Trip
	TripperName    *string `db:"tripper_name"     json:"tripper_name"`
	TotalOrders    int     `db:"total_orders"     json:"total_orders"`
	TotalCustomers int     `db:"total_customers"  json:"total_customers"`
	CatalogItems   int     `db:"catalog_items"    json:"catalog_items"`
}

// TripItem adalah satu produk dalam katalog sebuah trip beserta harganya.
type TripItem struct {
	ID           uuid.UUID       `db:"id"             json:"id"`
	TripID       uuid.UUID       `db:"trip_id"        json:"trip_id"`
	ProductID    uuid.UUID       `db:"product_id"     json:"product_id"`
	CostPrice    decimal.Decimal `db:"cost_price"     json:"cost_price"`
	CostPriceIDR decimal.Decimal `db:"cost_price_idr" json:"cost_price_idr"`
	MarkupType   string          `db:"markup_type"    json:"markup_type"`
	MarkupValue  decimal.Decimal `db:"markup_value"   json:"markup_value"`
	SellPrice    decimal.Decimal `db:"sell_price"     json:"sell_price"`
	MaxQty       *int            `db:"max_qty"        json:"max_qty"`
	IsActive     bool            `db:"is_active"      json:"is_active"`
	Notes        *string         `db:"notes"          json:"notes"`
	CreatedAt    time.Time       `db:"created_at"     json:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"     json:"updated_at"`
}

type TripItemDetail struct {
	TripItem
	ProductName  string  `db:"product_name"  json:"product_name"`
	ProductSKU   string  `db:"product_sku"   json:"product_sku"`
	Brand        *string `db:"brand"         json:"brand"`
	ImageURL     *string `db:"image_url"     json:"image_url"`
	WeightGram   int     `db:"weight_gram"   json:"weight_gram"`
	CategoryName *string `db:"category_name" json:"category_name"`
	// QtyOrdered memberi tahu admin berapa yang sudah dipesan customer untuk
	// produk ini, penting saat max_qty dibatasi.
	QtyOrdered int `db:"qty_ordered" json:"qty_ordered"`
}

// Kategori biaya perjalanan.
const (
	ExpenseTiket     = "tiket"
	ExpenseBagasi    = "bagasi"
	ExpenseAkomodasi = "akomodasi"
	ExpenseTransport = "transport"
	ExpenseVisa      = "visa"
	ExpenseLainnya   = "lainnya"
)

func IsValidExpenseCategory(c string) bool {
	switch c {
	case ExpenseTiket, ExpenseBagasi, ExpenseAkomodasi, ExpenseTransport, ExpenseVisa, ExpenseLainnya:
		return true
	default:
		return false
	}
}

type TripExpense struct {
	ID          uuid.UUID       `db:"id"          json:"id"`
	TripID      uuid.UUID       `db:"trip_id"     json:"trip_id"`
	Category    string          `db:"category"    json:"category"`
	Description string          `db:"description" json:"description"`
	Amount      decimal.Decimal `db:"amount"      json:"amount"`
	SpentAt     time.Time       `db:"spent_at"    json:"spent_at"`
	ReceiptURL  *string         `db:"receipt_url" json:"receipt_url"`
	CreatedBy   *uuid.UUID      `db:"created_by"  json:"created_by"`
	CreatedAt   time.Time       `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"  json:"updated_at"`
}
