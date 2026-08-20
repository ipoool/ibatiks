package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TripProfitReport adalah laporan keuangan satu trip.
//
// Rumusnya:
//
//	laba kotor  = omzet − HPP riil barang pesanan
//	laba bersih = laba kotor − biaya perjalanan
//
// HPP yang dipakai adalah biaya belanja yang benar-benar dialokasikan ke order
// (purchase_allocations), bukan estimasi saat order dibuat. Surplus belanja
// yang masuk stok sengaja tidak dibebankan ke trip ini: uangnya memang keluar,
// tapi nilainya masih dipegang sebagai barang, dan akan menjadi HPP ketika
// stok itu terjual. Nilainya tetap ditampilkan sebagai SurplusStockValue agar
// arus kas trip tetap terbaca.
type TripProfitReport struct {
	// Kosong saat laporannya mencakup seluruh trip: tidak ada satu trip yang
	// bisa disebut, dan mengarang salah satunya lebih menyesatkan daripada
	// membiarkannya kosong.
	TripID   *uuid.UUID `json:"trip_id"`
	TripCode string     `json:"trip_code"`
	Title    string     `json:"title"`
	Country  string     `json:"country"`
	Status   string     `json:"status"`

	Revenue      decimal.Decimal `json:"revenue"`        // total order yang dihitung sebagai omzet
	COGS         decimal.Decimal `json:"cogs"`           // HPP riil barang yang dipesan
	GrossProfit  decimal.Decimal `json:"gross_profit"`   // revenue − cogs
	TripExpenses decimal.Decimal `json:"trip_expenses"`  // tiket, bagasi, akomodasi, dll
	NetProfit    decimal.Decimal `json:"net_profit"`     // gross_profit − trip_expenses
	MarginPct    decimal.Decimal `json:"margin_percent"` // net_profit / revenue × 100

	ShippingFeeCollected decimal.Decimal `json:"shipping_fee_collected"`
	ShippingCostPaid     decimal.Decimal `json:"shipping_cost_paid"`
	DiscountGiven        decimal.Decimal `json:"discount_given"`

	SurplusStockQty   int             `json:"surplus_stock_qty"`
	SurplusStockValue decimal.Decimal `json:"surplus_stock_value"`

	TotalCapitalOut decimal.Decimal `json:"total_capital_out"` // seluruh uang keluar: belanja + biaya trip
	PaymentReceived decimal.Decimal `json:"payment_received"`
	Outstanding     decimal.Decimal `json:"outstanding"` // sisa tagihan yang belum masuk

	OrderCount    int `json:"order_count"`
	CustomerCount int `json:"customer_count"`
	ItemQty       int `json:"item_qty"`

	ExpenseBreakdown []ExpenseBreakdown `json:"expense_breakdown"`
}

type ExpenseBreakdown struct {
	Category string          `db:"category" json:"category"`
	Amount   decimal.Decimal `db:"amount"   json:"amount"`
}

// OrderProfit adalah margin per order, dipakai untuk menemukan order yang
// ternyata merugi karena harga di toko naik.
type OrderProfit struct {
	OrderID      uuid.UUID       `db:"order_id"      json:"order_id"`
	OrderNumber  string          `db:"order_number"  json:"order_number"`
	CustomerName string          `db:"customer_name" json:"customer_name"`
	TripCode     string          `db:"trip_code"     json:"trip_code"`
	Status       string          `db:"status"        json:"status"`
	OrderDate    time.Time       `db:"order_date"    json:"order_date"`
	Revenue      decimal.Decimal `db:"revenue"       json:"revenue"`
	COGS         decimal.Decimal `db:"cogs"          json:"cogs"`
	Profit       decimal.Decimal `db:"profit"        json:"profit"`
	MarginPct    decimal.Decimal `db:"margin_pct"    json:"margin_percent"`
}

// Receivable adalah order yang masih punya sisa tagihan.
type Receivable struct {
	OrderID         uuid.UUID       `db:"order_id"       json:"order_id"`
	OrderNumber     string          `db:"order_number"   json:"order_number"`
	CustomerID      uuid.UUID       `db:"customer_id"    json:"customer_id"`
	CustomerName    string          `db:"customer_name"  json:"customer_name"`
	CustomerPhone   string          `db:"customer_phone" json:"customer_phone"`
	TripCode        string          `db:"trip_code"      json:"trip_code"`
	Status          string          `db:"status"         json:"status"`
	OrderDate       time.Time       `db:"order_date"     json:"order_date"`
	Total           decimal.Decimal `db:"total"          json:"total"`
	PaidAmount      decimal.Decimal `db:"paid_amount"    json:"paid_amount"`
	BalanceDue      decimal.Decimal `db:"balance_due"    json:"balance_due"`
	DaysOutstanding int             `db:"days_outstanding" json:"days_outstanding"`
}

// CustomerSales merangkum belanja seorang customer lintas trip. Dipakai untuk
// mengenali pelanggan yang paling bernilai dan yang paling sering menunggak.
type CustomerSales struct {
	CustomerID    uuid.UUID       `db:"customer_id"    json:"customer_id"`
	CustomerCode  string          `db:"customer_code"  json:"customer_code"`
	CustomerName  string          `db:"customer_name"  json:"customer_name"`
	CustomerPhone string          `db:"customer_phone" json:"customer_phone"`
	City          *string         `db:"city"           json:"city"`
	OrderCount    int             `db:"order_count"    json:"order_count"`
	ItemQty       int             `db:"item_qty"       json:"item_qty"`
	Revenue       decimal.Decimal `db:"revenue"        json:"revenue"`
	COGS          decimal.Decimal `db:"cogs"           json:"cogs"`
	Profit        decimal.Decimal `db:"profit"         json:"profit"`
	Outstanding   decimal.Decimal `db:"outstanding"    json:"outstanding"`
	AvgOrderValue decimal.Decimal `db:"avg_order_value" json:"avg_order_value"`
	FirstOrderAt  *time.Time      `db:"first_order_at" json:"first_order_at"`
	LastOrderAt   *time.Time      `db:"last_order_at"  json:"last_order_at"`
}

// CustomerChannelSales memecah rekap per customer menjadi per kanal asal order.
//
// Hanya dipakai untuk ekspor CSV. Di layar, satu baris per customer sudah cukup
// untuk menjawab siapa pembelanja terbesar; pemecahan per kanal berguna justru
// setelah datanya dibawa ke spreadsheet — untuk menilai kanal mana yang
// mendatangkan pelanggan bernilai, bukan sekadar kanal mana yang ramai.
type CustomerChannelSales struct {
	CustomerCode  string          `db:"customer_code"  json:"customer_code"`
	CustomerName  string          `db:"customer_name"  json:"customer_name"`
	CustomerPhone string          `db:"customer_phone" json:"customer_phone"`
	City          *string         `db:"city"           json:"city"`
	Source        string          `db:"order_source"   json:"order_source"`
	OrderCount    int             `db:"order_count"    json:"order_count"`
	ItemQty       int             `db:"item_qty"       json:"item_qty"`
	Revenue       decimal.Decimal `db:"revenue"        json:"revenue"`
	COGS          decimal.Decimal `db:"cogs"           json:"cogs"`
	Profit        decimal.Decimal `db:"profit"         json:"profit"`
	Outstanding   decimal.Decimal `db:"outstanding"    json:"outstanding"`
	FirstOrderAt  *time.Time      `db:"first_order_at" json:"first_order_at"`
	LastOrderAt   *time.Time      `db:"last_order_at"  json:"last_order_at"`
}

// ChannelSales merangkum penjualan per asal order, misalnya WhatsApp versus
// Instagram, untuk menilai kanal mana yang paling menghasilkan.
type ChannelSales struct {
	Source        string          `db:"order_source"    json:"order_source"`
	OrderCount    int             `db:"order_count"     json:"order_count"`
	CustomerCount int             `db:"customer_count"  json:"customer_count"`
	ItemQty       int             `db:"item_qty"        json:"item_qty"`
	Revenue       decimal.Decimal `db:"revenue"         json:"revenue"`
	COGS          decimal.Decimal `db:"cogs"            json:"cogs"`
	Profit        decimal.Decimal `db:"profit"          json:"profit"`
	AvgOrderValue decimal.Decimal `db:"avg_order_value" json:"avg_order_value"`
	// RevenueShare adalah porsi omzet kanal ini terhadap total, dalam persen.
	RevenueShare decimal.Decimal `json:"revenue_share"`
}

// ProductSales merangkum performa produk lintas trip.
type ProductSales struct {
	ProductID    uuid.UUID       `db:"product_id"    json:"product_id"`
	ProductName  string          `db:"product_name"  json:"product_name"`
	ProductSKU   string          `db:"product_sku"   json:"product_sku"`
	CategoryName *string         `db:"category_name" json:"category_name"`
	QtySold      int             `db:"qty_sold"      json:"qty_sold"`
	OrderCount   int             `db:"order_count"   json:"order_count"`
	Revenue      decimal.Decimal `db:"revenue"       json:"revenue"`
	COGS         decimal.Decimal `db:"cogs"          json:"cogs"`
	Profit       decimal.Decimal `db:"profit"        json:"profit"`
}

// DashboardSummary adalah angka-angka yang tampil di halaman depan.
type DashboardSummary struct {
	ActiveTrips     int             `json:"active_trips"`
	OpenOrders      int             `json:"open_orders"`
	PendingShipment int             `json:"pending_shipment"`
	Outstanding     decimal.Decimal `json:"outstanding"`

	RevenueThisMonth decimal.Decimal `json:"revenue_this_month"`
	ProfitThisMonth  decimal.Decimal `json:"profit_this_month"`
	OrdersThisMonth  int             `json:"orders_this_month"`

	StockValue    decimal.Decimal `json:"stock_value"`
	StockQty      int             `json:"stock_qty"`
	CustomerCount int             `json:"customer_count"`

	RecentOrders  []OrderListItem `json:"recent_orders"`
	UpcomingTrips []Trip          `json:"upcoming_trips"`
	TopProducts   []ProductSales  `json:"top_products"`
}
