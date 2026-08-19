package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Jenis invoice.
const (
	InvoiceDP    = "dp"    // tagihan uang muka saat order dikonfirmasi
	InvoiceFinal = "final" // tagihan pelunasan saat barang sudah sampai
)

// Status invoice.
const (
	InvoiceDraft = "draft"
	InvoiceSent  = "sent"
	InvoicePaid  = "paid"
	InvoiceVoid  = "void"
)

// Kanal pengiriman invoice ke customer.
const (
	ChannelWA     = "wa"
	ChannelEmail  = "email"
	ChannelManual = "manual"
)

func IsValidInvoiceType(t string) bool {
	return t == InvoiceDP || t == InvoiceFinal
}

func IsValidSentChannel(c string) bool {
	switch c {
	case ChannelWA, ChannelEmail, ChannelManual:
		return true
	default:
		return false
	}
}

// Invoice menyimpan salinan nominal saat diterbitkan. Kalau order diedit
// setelah invoice terkirim, invoice lama tetap memuat angka yang sudah terlanjur
// dilihat customer, dan admin menerbitkan invoice pengganti.
type Invoice struct {
	ID            uuid.UUID  `db:"id"             json:"id"`
	InvoiceNumber string     `db:"invoice_number" json:"invoice_number"`
	OrderID       uuid.UUID  `db:"order_id"       json:"order_id"`
	Type          string     `db:"type"           json:"type"`
	IssueDate     time.Time  `db:"issue_date"     json:"issue_date"`
	DueDate       *time.Time `db:"due_date"       json:"due_date"`

	Subtotal    decimal.Decimal `db:"subtotal"     json:"subtotal"`
	Discount    decimal.Decimal `db:"discount"     json:"discount"`
	ShippingFee decimal.Decimal `db:"shipping_fee" json:"shipping_fee"`
	Total       decimal.Decimal `db:"total"        json:"total"`
	// DPAmount adalah uang muka pesanan ini: pada invoice DP ia yang ditagih,
	// pada invoice pelunasan ia yang dikurangkan dari total.
	DPAmount   decimal.Decimal `db:"dp_amount"   json:"dp_amount"`
	AmountPaid decimal.Decimal `db:"amount_paid" json:"amount_paid"`
	AmountDue  decimal.Decimal `db:"amount_due"   json:"amount_due"`

	Status      string     `db:"status"       json:"status"`
	PDFPath     *string    `db:"pdf_path"     json:"pdf_path"`
	SentChannel *string    `db:"sent_channel" json:"sent_channel"`
	SentAt      *time.Time `db:"sent_at"      json:"sent_at"`
	PaidAt      *time.Time `db:"paid_at"      json:"paid_at"`
	Notes       *string    `db:"notes"        json:"notes"`
	CreatedBy   *uuid.UUID `db:"created_by"   json:"created_by"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
}

type InvoiceListItem struct {
	Invoice
	OrderNumber   string `db:"order_number"   json:"order_number"`
	CustomerName  string `db:"customer_name"  json:"customer_name"`
	CustomerPhone string `db:"customer_phone" json:"customer_phone"`
	TripCode      string `db:"trip_code"      json:"trip_code"`
}

// Status paket kiriman.
const (
	ShipmentPacking   = "packing"
	ShipmentReady     = "ready"
	ShipmentShipped   = "shipped"
	ShipmentDelivered = "delivered"
	ShipmentReturned  = "returned"
)

func IsValidShipmentStatus(s string) bool {
	switch s {
	case ShipmentPacking, ShipmentReady, ShipmentShipped, ShipmentDelivered, ShipmentReturned:
		return true
	default:
		return false
	}
}

// Layanan JNE yang lazim dipakai. Nilai bebas lain tetap diterima karena
// sesekali ada layanan promo baru.
var JNEServices = []string{"REG", "YES", "OKE", "JTR"}

type Shipment struct {
	ID                 uuid.UUID       `db:"id"                   json:"id"`
	OrderID            uuid.UUID       `db:"order_id"             json:"order_id"`
	Courier            string          `db:"courier"              json:"courier"`
	Service            string          `db:"service"              json:"service"`
	TrackingNumber     *string         `db:"tracking_number"      json:"tracking_number"`
	WeightGram         int             `db:"weight_gram"          json:"weight_gram"`
	LengthCM           int             `db:"length_cm"            json:"length_cm"`
	WidthCM            int             `db:"width_cm"             json:"width_cm"`
	HeightCM           int             `db:"height_cm"            json:"height_cm"`
	EstimatedCost      decimal.Decimal `db:"estimated_cost"       json:"estimated_cost"`
	ShippingCost       decimal.Decimal `db:"shipping_cost"        json:"shipping_cost"`
	Status             string          `db:"status"               json:"status"`
	PackedAt           *time.Time      `db:"packed_at"            json:"packed_at"`
	PackedBy           *uuid.UUID      `db:"packed_by"            json:"packed_by"`
	ShippedAt          *time.Time      `db:"shipped_at"           json:"shipped_at"`
	DeliveredAt        *time.Time      `db:"delivered_at"         json:"delivered_at"`
	CustomerNotifiedAt *time.Time      `db:"customer_notified_at" json:"customer_notified_at"`
	Notes              *string         `db:"notes"                json:"notes"`
	CreatedAt          time.Time       `db:"created_at"           json:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at"           json:"updated_at"`
}

// ShippingQueueItem adalah satu baris di menu Pengiriman.
//
// Yang didaftar adalah order, bukan paket. Paket baru terbentuk setelah admin
// mengisi data kemasan, sementara pekerjaannya justru dimulai sebelum itu —
// order yang DP-nya sudah masuk itulah yang menunggu ditimbang, dihitung
// ongkirnya, lalu diserahkan ke kurir. Mendaftar paket berarti menyembunyikan
// justru pekerjaan yang belum dikerjakan.
//
// Kolom paket bertipe penunjuk karena memang boleh belum ada.
type ShippingQueueItem struct {
	OrderID        uuid.UUID       `db:"order_id"        json:"order_id"`
	OrderNumber    string          `db:"order_number"    json:"order_number"`
	OrderStatus    string          `db:"order_status"    json:"order_status"`
	OrderDate      time.Time       `db:"order_date"      json:"order_date"`
	TripCode       string          `db:"trip_code"       json:"trip_code"`
	CustomerName   string          `db:"customer_name"   json:"customer_name"`
	RecipientName  string          `db:"recipient_name"  json:"recipient_name"`
	RecipientPhone string          `db:"recipient_phone" json:"recipient_phone"`
	ShippingCity   string          `db:"shipping_city"   json:"shipping_city"`
	TotalQty       int             `db:"total_qty"       json:"total_qty"`
	Total          decimal.Decimal `db:"total"           json:"total"`
	BalanceDue     decimal.Decimal `db:"balance_due"     json:"balance_due"`
	// ShippingFee adalah ongkir yang ditagihkan ke customer, tersimpan di
	// order. Nol berarti layanan kurirnya belum dipilih.
	ShippingFee decimal.Decimal `db:"shipping_fee" json:"shipping_fee"`

	ShipmentID     *uuid.UUID `db:"shipment_id"     json:"shipment_id"`
	Courier        *string    `db:"courier"         json:"courier"`
	Service        *string    `db:"service"         json:"service"`
	WeightGram     *int       `db:"weight_gram"     json:"weight_gram"`
	LengthCM       *int       `db:"length_cm"       json:"length_cm"`
	WidthCM        *int       `db:"width_cm"        json:"width_cm"`
	HeightCM       *int       `db:"height_cm"       json:"height_cm"`
	TrackingNumber *string    `db:"tracking_number" json:"tracking_number"`
	ShipmentStatus *string    `db:"shipment_status" json:"shipment_status"`
	ShipmentNotes  *string    `db:"shipment_notes"  json:"shipment_notes"`
	PackedAt       *time.Time `db:"packed_at"       json:"packed_at"`
	ShippedAt      *time.Time `db:"shipped_at"      json:"shipped_at"`
	// ShippingCost adalah ongkos yang benar-benar dibayar ke kurir, diisi saat
	// resi dicatat. Dipisah dari ShippingFee karena toko boleh menanggung
	// selisihnya.
	ShippingCost       *decimal.Decimal `db:"shipping_cost"        json:"shipping_cost"`
	CustomerNotifiedAt *time.Time       `db:"customer_notified_at" json:"customer_notified_at"`
}

// Tahap pekerjaan di menu Pengiriman. Bukan status yang disimpan, melainkan
// cara menyaring daftar menurut apa yang belum dikerjakan.
const (
	// ShippingStagePacking: paketnya belum ditimbang atau ongkirnya belum
	// ditetapkan, jadi invoice pelunasannya pun belum bisa terbit.
	ShippingStagePacking = "perlu_kemas"
	// ShippingStageReady: sudah lunas dan tinggal diserahkan ke kurir.
	ShippingStageReady = "siap_kirim"
	// ShippingStageSent: nomor resi sudah tercatat.
	ShippingStageSent = "terkirim"
)

func IsValidShippingStage(s string) bool {
	switch s {
	case ShippingStagePacking, ShippingStageReady, ShippingStageSent:
		return true
	default:
		return false
	}
}
