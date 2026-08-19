package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Status perjalanan.
//
// Hanya dua, dan keduanya menjawab satu pertanyaan yang benar-benar dipakai
// sehari-hari: trip ini masih menerima order atau tidak. Posisi barangnya —
// sedang dibelanjakan, dalam perjalanan pulang, sudah sampai — dibaca dari
// order dan pembelian yang tercatat, bukan dari status trip, supaya satu
// kejadian tidak perlu dicatat di dua tempat.
const (
	TripOpen   = "open"   // order masih boleh masuk
	TripClosed = "closed" // pendaftaran order ditutup
)

// tripTransitions memetakan status trip ke daftar status berikutnya yang sah.
// Peta ini satu-satunya sumber kebenaran alur trip.
var tripTransitions = map[string][]string{
	// Bolak-balik dibiarkan terbuka: menutup order lalu membukanya lagi karena
	// ada yang menyusul titip adalah kejadian biasa, bukan penyimpangan.
	TripOpen:   {TripClosed},
	TripClosed: {TripOpen},
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
func TripAcceptsOrder(status string) bool {
	return status == TripOpen
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

// TripDeletionImpact merangkum apa yang ikut terhapus bersama sebuah trip, dan
// apa yang justru menghalanginya.
//
// Dihitung lebih dulu lalu ditampilkan sebelum tombol hapus ditekan. Menghapus
// trip berarti membuang order, invoice yang sudah terkirim, dan catatan uang
// yang sudah diterima — angka-angkanya harus terlihat dulu, bukan disimpulkan
// sendiri oleh admin dari ingatan.
type TripDeletionImpact struct {
	TripID   uuid.UUID `json:"trip_id"`
	TripCode string    `json:"trip_code"`

	Orders   int `json:"orders"`
	Invoices int `json:"invoices"`
	// PaymentsTotal adalah uang yang sudah benar-benar diterima pada trip ini.
	// Ini angka yang paling perlu dilihat: menghapus trip membuatnya lenyap
	// dari pembukuan sementara uangnya tetap ada di rekening.
	PaymentsTotal decimal.Decimal `json:"payments_total"`
	Purchases     int             `json:"purchases"`
	PurchasesCost decimal.Decimal `json:"purchases_cost"`
	Expenses      int             `json:"expenses"`
	CatalogItems  int             `json:"catalog_items"`
	Shipments     int             `json:"shipments"`

	// ShippedOrders menghalangi penghapusan: barangnya sudah di tangan
	// customer, jadi penjualannya sudah jadi dan bukan lagi sesuatu yang bisa
	// dibatalkan dengan menghapus catatannya.
	ShippedOrders []string `json:"shipped_orders"`
	// StockOnHand menghalangi penghapusan: barang surplus dari trip ini masih
	// tersimpan di gudang. Barangnya nyata, jadi membuang catatan asalnya
	// hanya akan menyisakan stok tanpa asal-usul.
	StockOnHand []TripStockOnHand `json:"stock_on_hand"`
}

// TripStockOnHand adalah sisa barang surplus dari sebuah trip yang masih ada.
type TripStockOnHand struct {
	ProductName string `db:"product_name" json:"product_name"`
	SKU         string `db:"sku"          json:"sku"`
	Qty         int    `db:"qty"          json:"qty"`
}

// Deletable benar bila tidak ada yang menghalangi trip ini dihapus.
func (i TripDeletionImpact) Deletable() bool {
	return len(i.ShippedOrders) == 0 && len(i.StockOnHand) == 0
}

// Empty benar bila trip belum menyimpan apa pun yang berarti, sehingga
// menghapusnya tidak perlu peringatan berat.
func (i TripDeletionImpact) Empty() bool {
	return i.Orders == 0 && i.Purchases == 0 && i.Expenses == 0
}
