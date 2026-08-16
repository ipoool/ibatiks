package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Status order, mengikuti perjalanan satu pesanan dari dicatat sampai selesai.
const (
	OrderDraft      = "draft"       // masih disusun admin, belum ditagihkan
	OrderAwaitingDP = "awaiting_dp" // sudah dikonfirmasi, menunggu DP masuk
	OrderDPPaid     = "dp_paid"     // DP diterima, pesanan terkunci
	OrderPurchasing = "purchasing"  // barang sedang dibelikan tripper
	OrderArrived    = "arrived"     // barang sudah sampai dan dicocokkan
	OrderPacked     = "packed"      // sudah dikemas atas nama customer
	OrderInvoiced   = "invoiced"    // invoice pelunasan sudah dikirim
	OrderPaid       = "paid"        // lunas, siap kirim
	OrderShipped    = "shipped"     // sudah diserahkan ke kurir, resi terisi
	OrderCompleted  = "completed"   // diterima customer
	OrderCancelled  = "cancelled"
)

// orderTransitions adalah satu-satunya sumber kebenaran alur status order.
// Handler tidak boleh mengubah status di luar peta ini.
var orderTransitions = map[string][]string{
	OrderDraft:      {OrderAwaitingDP, OrderCancelled},
	OrderAwaitingDP: {OrderDPPaid, OrderDraft, OrderCancelled},
	OrderDPPaid:     {OrderPurchasing, OrderArrived, OrderCancelled},
	OrderPurchasing: {OrderArrived, OrderCancelled},
	OrderArrived:    {OrderPacked, OrderCancelled},
	OrderPacked:     {OrderInvoiced, OrderCancelled},
	// Order boleh langsung ke paid kalau customer melunasi sebelum invoice
	// resmi dibuat, yang sering terjadi untuk pelanggan lama.
	OrderInvoiced:  {OrderPaid, OrderPacked, OrderCancelled},
	OrderPaid:      {OrderShipped},
	OrderShipped:   {OrderCompleted},
	OrderCompleted: {},
	OrderCancelled: {},
}

func IsValidOrderStatus(status string) bool {
	_, ok := orderTransitions[status]
	return ok
}

func CanTransitionOrder(from, to string) bool {
	for _, allowed := range orderTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

func NextOrderStatuses(from string) []string {
	next := orderTransitions[from]
	out := make([]string, len(next))
	copy(out, next)
	return out
}

// OrderIsEditable menentukan apakah isi order (item dan qty) masih boleh
// diubah. Setelah barang diserahkan ke kurir, isi order dibekukan supaya
// dokumen pengiriman dan invoice tidak lagi berbeda dengan kenyataan.
func OrderIsEditable(status string) bool {
	switch status {
	case OrderShipped, OrderCompleted, OrderCancelled:
		return false
	default:
		return true
	}
}

// OrderIsConfirmed menandai order yang DP-nya sudah diverifikasi, yaitu status
// "Diproses" pada proses bisnis. Hanya order inilah yang boleh masuk daftar
// belanja tripper: membelanjakan order yang DP-nya belum masuk berarti
// menalangi pembelian dengan uang toko.
func OrderIsConfirmed(status string) bool {
	switch status {
	case OrderDraft, OrderAwaitingDP, OrderCancelled:
		return false
	default:
		return true
	}
}

// OrderCountsAsRevenue menentukan apakah order ikut dihitung sebagai omzet
// pada laporan. Order yang dibatalkan dan yang masih draft tidak dihitung.
func OrderCountsAsRevenue(status string) bool {
	return status != OrderCancelled && status != OrderDraft
}

// Asal order. Dipakai untuk rekap penjualan per channel, karena biaya promosi
// dan cara melayani customer berbeda antar kanal.
const (
	SourceWhatsApp    = "whatsapp"
	SourceInstagram   = "instagram"
	SourceTikTok      = "tiktok"
	SourceMarketplace = "marketplace"
	SourceOther       = "lainnya"
)

var OrderSources = []string{
	SourceWhatsApp, SourceInstagram, SourceTikTok, SourceMarketplace, SourceOther,
}

func IsValidOrderSource(source string) bool {
	for _, s := range OrderSources {
		if s == source {
			return true
		}
	}
	return false
}

// Status pemenuhan tiap baris pesanan.
const (
	FulfillmentPending     = "pending"
	FulfillmentPurchased   = "purchased"
	FulfillmentPartial     = "partial"
	FulfillmentUnavailable = "unavailable"
	FulfillmentRefunded    = "refunded"
)

type Order struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	OrderNumber string    `db:"order_number" json:"order_number"`
	TripID      uuid.UUID `db:"trip_id"      json:"trip_id"`
	CustomerID  uuid.UUID `db:"customer_id"  json:"customer_id"`
	OrderDate   time.Time `db:"order_date"   json:"order_date"`
	Status      string    `db:"status"       json:"status"`
	OrderSource string    `db:"order_source" json:"order_source"`

	Subtotal    decimal.Decimal `db:"subtotal"     json:"subtotal"`
	Discount    decimal.Decimal `db:"discount"     json:"discount"`
	ShippingFee decimal.Decimal `db:"shipping_fee" json:"shipping_fee"`
	Total       decimal.Decimal `db:"total"        json:"total"`
	DPRequired  decimal.Decimal `db:"dp_required"  json:"dp_required"`
	PaidAmount  decimal.Decimal `db:"paid_amount"  json:"paid_amount"`
	BalanceDue  decimal.Decimal `db:"balance_due"  json:"balance_due"`

	RecipientName      string  `db:"recipient_name"       json:"recipient_name"`
	RecipientPhone     string  `db:"recipient_phone"      json:"recipient_phone"`
	ShippingAddress    string  `db:"shipping_address"     json:"shipping_address"`
	ShippingCity       string  `db:"shipping_city"        json:"shipping_city"`
	ShippingProvince   *string `db:"shipping_province"    json:"shipping_province"`
	ShippingPostalCode *string `db:"shipping_postal_code" json:"shipping_postal_code"`

	Notes        *string    `db:"notes"         json:"notes"`
	CancelReason *string    `db:"cancel_reason" json:"cancel_reason"`
	CancelledAt  *time.Time `db:"cancelled_at"  json:"cancelled_at"`
	CreatedBy    *uuid.UUID `db:"created_by"    json:"created_by"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"    json:"updated_at"`
}

// IsFullyPaid dipakai untuk memutuskan apakah paket boleh dikirim.
func (o Order) IsFullyPaid() bool {
	return o.BalanceDue.LessThanOrEqual(decimal.Zero) && o.Total.GreaterThan(decimal.Zero)
}

// OrderListItem adalah bentuk ringkas untuk tabel daftar order.
//
// Mata uang dan kurs trip ikut dikirim supaya daftar order bisa menampilkan
// nominal dalam mata uang negara tujuan tanpa memanggil data trip satu per satu.
type OrderListItem struct {
	Order
	CustomerName     string          `db:"customer_name"      json:"customer_name"`
	CustomerCode     string          `db:"customer_code"      json:"customer_code"`
	TripCode         string          `db:"trip_code"          json:"trip_code"`
	TripTitle        string          `db:"trip_title"         json:"trip_title"`
	TripCurrency     string          `db:"trip_currency"      json:"trip_currency"`
	TripExchangeRate decimal.Decimal `db:"trip_exchange_rate" json:"trip_exchange_rate"`
	ItemCount        int             `db:"item_count"         json:"item_count"`
	TotalQty         int             `db:"total_qty"          json:"total_qty"`
}

type OrderItem struct {
	ID                uuid.UUID       `db:"id"                 json:"id"`
	OrderID           uuid.UUID       `db:"order_id"           json:"order_id"`
	ProductID         uuid.UUID       `db:"product_id"         json:"product_id"`
	TripItemID        *uuid.UUID      `db:"trip_item_id"       json:"trip_item_id"`
	ProductName       string          `db:"product_name"       json:"product_name"`
	ProductSKU        string          `db:"product_sku"        json:"product_sku"`
	Qty               int             `db:"qty"                json:"qty"`
	UnitPrice         decimal.Decimal `db:"unit_price"         json:"unit_price"`
	UnitCostEst       decimal.Decimal `db:"unit_cost_est"      json:"unit_cost_est"`
	Subtotal          decimal.Decimal `db:"subtotal"           json:"subtotal"`
	QtyPurchased      int             `db:"qty_purchased"      json:"qty_purchased"`
	QtyReceived       int             `db:"qty_received"       json:"qty_received"`
	FulfillmentStatus string          `db:"fulfillment_status" json:"fulfillment_status"`
	Notes             *string         `db:"notes"              json:"notes"`
	CreatedAt         time.Time       `db:"created_at"         json:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at"         json:"updated_at"`
}

// OrderDetail adalah bentuk lengkap yang dipakai halaman detail order.
type OrderDetail struct {
	Order
	Customer     *Customer   `json:"customer"`
	Trip         *Trip       `json:"trip"`
	Items        []OrderItem `json:"items"`
	Payments     []Payment   `json:"payments"`
	Invoices     []Invoice   `json:"invoices"`
	Shipment     *Shipment   `json:"shipment"`
	NextStatuses []string    `json:"next_statuses"`
	Editable     bool        `json:"editable"`
}

// Jenis pembayaran.
const (
	PaymentDP         = "dp"
	PaymentSettlement = "settlement"
	PaymentRefund     = "refund"
	PaymentAdjustment = "adjustment"
)

func IsValidPaymentType(t string) bool {
	switch t {
	case PaymentDP, PaymentSettlement, PaymentRefund, PaymentAdjustment:
		return true
	default:
		return false
	}
}

// SignedAmount mengembalikan nominal dengan arah: refund mengurangi uang yang
// sudah diterima, sisanya menambah.
func SignedAmount(paymentType string, amount decimal.Decimal) decimal.Decimal {
	if paymentType == PaymentRefund {
		return amount.Neg()
	}
	return amount
}

type Payment struct {
	ID         uuid.UUID       `db:"id"          json:"id"`
	OrderID    uuid.UUID       `db:"order_id"    json:"order_id"`
	Type       string          `db:"type"        json:"type"`
	Amount     decimal.Decimal `db:"amount"      json:"amount"`
	Method     string          `db:"method"      json:"method"`
	Reference  *string         `db:"reference"   json:"reference"`
	ProofURL   *string         `db:"proof_url"   json:"proof_url"`
	PaidAt     time.Time       `db:"paid_at"     json:"paid_at"`
	Notes      *string         `db:"notes"       json:"notes"`
	RecordedBy *uuid.UUID      `db:"recorded_by" json:"recorded_by"`
	CreatedAt  time.Time       `db:"created_at"  json:"created_at"`
}
