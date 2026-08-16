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
	AmountPaid  decimal.Decimal `db:"amount_paid"  json:"amount_paid"`
	AmountDue   decimal.Decimal `db:"amount_due"   json:"amount_due"`

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

type ShipmentListItem struct {
	Shipment
	OrderNumber     string          `db:"order_number"     json:"order_number"`
	CustomerName    string          `db:"customer_name"    json:"customer_name"`
	RecipientName   string          `db:"recipient_name"   json:"recipient_name"`
	RecipientPhone  string          `db:"recipient_phone"  json:"recipient_phone"`
	ShippingCity    string          `db:"shipping_city"    json:"shipping_city"`
	OrderStatus     string          `db:"order_status"     json:"order_status"`
	OrderBalanceDue decimal.Decimal `db:"order_balance_due" json:"order_balance_due"`
}
