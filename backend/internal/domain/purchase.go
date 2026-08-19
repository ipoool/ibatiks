package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Purchase mencatat realisasi belanja tripper di toko, bukan rencananya.
// Selisih antara harga di sini dan sell_price pada TripItem adalah margin
// sesungguhnya dari sebuah trip.
type Purchase struct {
	ID              uuid.UUID       `db:"id"                json:"id"`
	TripID          uuid.UUID       `db:"trip_id"           json:"trip_id"`
	ProductID       uuid.UUID       `db:"product_id"        json:"product_id"`
	PurchaseDate    time.Time       `db:"purchase_date"     json:"purchase_date"`
	Qty             int             `db:"qty"               json:"qty"`
	UnitCostForeign decimal.Decimal `db:"unit_cost_foreign" json:"unit_cost_foreign"`
	Currency        string          `db:"currency"          json:"currency"`
	ExchangeRate    decimal.Decimal `db:"exchange_rate"     json:"exchange_rate"`
	UnitCostIDR     decimal.Decimal `db:"unit_cost_idr"     json:"unit_cost_idr"`
	TotalCostIDR    decimal.Decimal `db:"total_cost_idr"    json:"total_cost_idr"`
	StoreName       *string         `db:"store_name"        json:"store_name"`
	ReceiptURL      *string         `db:"receipt_url"       json:"receipt_url"`
	Notes           *string         `db:"notes"             json:"notes"`
	PurchasedBy     *uuid.UUID      `db:"purchased_by"      json:"purchased_by"`
	CreatedAt       time.Time       `db:"created_at"        json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"        json:"updated_at"`
}

type PurchaseDetail struct {
	Purchase
	ProductName   string  `db:"product_name"     json:"product_name"`
	ProductSKU    string  `db:"product_sku"      json:"product_sku"`
	PurchaserName *string `db:"purchaser_name"  json:"purchaser_name"`
	QtyToOrders   int     `db:"qty_to_orders"    json:"qty_to_orders"`
	QtyToStock    int     `db:"qty_to_stock"     json:"qty_to_stock"`
}

// PurchaseAllocation menjawab pertanyaan "unit ini dibeli untuk siapa".
// OrderItemID nil berarti unit tersebut tidak dipesan siapa pun dan masuk stok.
type PurchaseAllocation struct {
	ID          uuid.UUID       `db:"id"            json:"id"`
	PurchaseID  uuid.UUID       `db:"purchase_id"   json:"purchase_id"`
	OrderItemID *uuid.UUID      `db:"order_item_id" json:"order_item_id"`
	Qty         int             `db:"qty"           json:"qty"`
	UnitCostIDR decimal.Decimal `db:"unit_cost_idr" json:"unit_cost_idr"`
	CreatedAt   time.Time       `db:"created_at"    json:"created_at"`
}

type PurchaseAllocationDetail struct {
	PurchaseAllocation
	// OrderID menyertai nomornya supaya daftar alokasi bisa menautkan langsung
	// ke detail ordernya; tanpa ini nomor order di layar cuma teks yang harus
	// disalin lalu dicari sendiri di menu Order.
	OrderID      *uuid.UUID `db:"order_id"      json:"order_id"`
	OrderNumber  *string    `db:"order_number"  json:"order_number"`
	CustomerName *string    `db:"customer_name" json:"customer_name"`
	ProductName  string     `db:"product_name"  json:"product_name"`
}

// ShoppingListEntry adalah satu baris daftar belanja yang dibawa tripper.
// Angka-angkanya diagregasi langsung dari order sehingga selalu mencerminkan
// pesanan terkini, termasuk perubahan qty yang baru saja diedit admin.
type ShoppingListEntry struct {
	ProductID    uuid.UUID `db:"product_id"      json:"product_id"`
	ProductName  string    `db:"product_name"    json:"product_name"`
	ProductSKU   string    `db:"product_sku"     json:"product_sku"`
	Brand        *string   `db:"brand"           json:"brand"`
	StoreName    *string   `db:"store_name"      json:"store_name"`
	ImageURL     *string   `db:"image_url"       json:"image_url"`
	CategoryName *string   `db:"category_name"   json:"category_name"`
	// QtyOrdered hanya menghitung order yang DP-nya sudah masuk, karena itulah
	// yang benar-benar boleh dibelanjakan tripper.
	QtyOrdered int `db:"qty_ordered"      json:"qty_ordered"`
	// QtyAwaitingDP adalah permintaan yang masih menunggu DP. Ditampilkan
	// terpisah supaya tripper tahu ada potensi tambahan, tanpa ikut terbeli.
	QtyAwaitingDP int             `db:"qty_awaiting_dp" json:"qty_awaiting_dp"`
	QtyPurchased  int             `db:"qty_purchased"   json:"qty_purchased"`
	QtyRemaining  int             `db:"qty_remaining"   json:"qty_remaining"`
	OrderCount    int             `db:"order_count"     json:"order_count"`
	EstCostIDR    decimal.Decimal `db:"est_cost_idr"    json:"est_cost_idr"`
	CostPrice     decimal.Decimal `db:"cost_price"      json:"cost_price"`
	SellPriceIDR  decimal.Decimal `db:"sell_price_idr"  json:"sell_price_idr"`
}

// StockItem adalah posisi stok berjalan hasil surplus belanja.
type StockItem struct {
	ID         uuid.UUID       `db:"id"           json:"id"`
	ProductID  uuid.UUID       `db:"product_id"   json:"product_id"`
	QtyOnHand  int             `db:"qty_on_hand"  json:"qty_on_hand"`
	AvgCostIDR decimal.Decimal `db:"avg_cost_idr" json:"avg_cost_idr"`
	Location   *string         `db:"location"     json:"location"`
	UpdatedAt  time.Time       `db:"updated_at"   json:"updated_at"`
}

type StockItemDetail struct {
	StockItem
	ProductName  string          `db:"product_name"  json:"product_name"`
	ProductSKU   string          `db:"product_sku"   json:"product_sku"`
	Brand        *string         `db:"brand"         json:"brand"`
	ImageURL     *string         `db:"image_url"     json:"image_url"`
	CategoryName *string         `db:"category_name" json:"category_name"`
	StockValue   decimal.Decimal `db:"stock_value"   json:"stock_value"`
}

// Jenis pergerakan stok.
const (
	StockInPurchase     = "in_purchase"     // surplus belanja masuk gudang
	StockOutOrder       = "out_order"       // stok dipakai memenuhi order
	StockOutMarketplace = "out_marketplace" // terjual di marketplace
	StockAdjustment     = "adjustment"      // koreksi hasil stock opname
)

func IsValidStockMovementType(t string) bool {
	switch t {
	case StockInPurchase, StockOutOrder, StockOutMarketplace, StockAdjustment:
		return true
	default:
		return false
	}
}

type StockMovement struct {
	ID           uuid.UUID        `db:"id"             json:"id"`
	ProductID    uuid.UUID        `db:"product_id"     json:"product_id"`
	Type         string           `db:"type"           json:"type"`
	Qty          int              `db:"qty"            json:"qty"`
	UnitCostIDR  decimal.Decimal  `db:"unit_cost_idr"  json:"unit_cost_idr"`
	SalePriceIDR *decimal.Decimal `db:"sale_price_idr" json:"sale_price_idr"`
	TripID       *uuid.UUID       `db:"trip_id"        json:"trip_id"`
	RefType      *string          `db:"ref_type"       json:"ref_type"`
	RefID        *uuid.UUID       `db:"ref_id"         json:"ref_id"`
	Note         *string          `db:"note"           json:"note"`
	CreatedBy    *uuid.UUID       `db:"created_by"     json:"created_by"`
	CreatedAt    time.Time        `db:"created_at"     json:"created_at"`
}

type StockMovementDetail struct {
	StockMovement
	ProductName string  `db:"product_name" json:"product_name"`
	ProductSKU  string  `db:"product_sku"  json:"product_sku"`
	ActorName   *string `db:"actor_name"  json:"actor_name"`
}
