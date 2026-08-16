package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
)

const purchaseColumns = `id, trip_id, product_id, purchase_date, qty, unit_cost_foreign, currency,
	                     exchange_rate, unit_cost_idr, total_cost_idr, store_name, receipt_url,
	                     notes, purchased_by, created_at, updated_at`

const purchaseColumnsPrefixed = `pu.id, pu.trip_id, pu.product_id, pu.purchase_date, pu.qty,
	                             pu.unit_cost_foreign, pu.currency, pu.exchange_rate,
	                             pu.unit_cost_idr, pu.total_cost_idr, pu.store_name,
	                             pu.receipt_url, pu.notes, pu.purchased_by,
	                             pu.created_at, pu.updated_at`

type PurchaseRepo struct{}

func NewPurchaseRepo() *PurchaseRepo { return &PurchaseRepo{} }

type PurchaseParams struct {
	TripID          uuid.UUID
	ProductID       uuid.UUID
	PurchaseDate    time.Time
	Qty             int
	UnitCostForeign decimal.Decimal
	Currency        string
	ExchangeRate    decimal.Decimal
	UnitCostIDR     decimal.Decimal
	TotalCostIDR    decimal.Decimal
	StoreName       *string
	ReceiptURL      *string
	Notes           *string
	PurchasedBy     *uuid.UUID
}

func (r *PurchaseRepo) Create(ctx context.Context, q db.Querier, p PurchaseParams) (*domain.Purchase, error) {
	return collectOne[domain.Purchase](ctx, q, "pembelian", `
		INSERT INTO purchases (trip_id, product_id, purchase_date, qty, unit_cost_foreign, currency,
		                       exchange_rate, unit_cost_idr, total_cost_idr, store_name,
		                       receipt_url, notes, purchased_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING `+purchaseColumns,
		p.TripID, p.ProductID, p.PurchaseDate, p.Qty, p.UnitCostForeign, p.Currency,
		p.ExchangeRate, p.UnitCostIDR, p.TotalCostIDR, p.StoreName, p.ReceiptURL, p.Notes, p.PurchasedBy)
}

func (r *PurchaseRepo) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Purchase, error) {
	return collectOne[domain.Purchase](ctx, q, "pembelian",
		`SELECT `+purchaseColumns+` FROM purchases WHERE id = $1`, id)
}

func (r *PurchaseRepo) List(ctx context.Context, q db.Querier, p pagination.Params, tripID, productID *uuid.UUID) ([]domain.PurchaseDetail, int64, error) {
	conditions := []string{}
	args := []any{}

	if tripID != nil {
		args = append(args, *tripID)
		conditions = append(conditions, fmt.Sprintf("pu.trip_id = $%d", len(args)))
	}
	if productID != nil {
		args = append(args, *productID)
		conditions = append(conditions, fmt.Sprintf("pu.product_id = $%d", len(args)))
	}
	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf(
			"(pr.name ILIKE $%d OR pr.sku ILIKE $%d OR COALESCE(pu.store_name, '') ILIKE $%d)", n, n, n))
	}
	where := buildWhere(conditions)

	var total int64
	countQuery := `SELECT count(*) FROM purchases pu JOIN products pr ON pr.id = pu.product_id` + where
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"purchase_date": "pu.purchase_date",
		"qty":           "pu.qty",
		"total_cost":    "pu.total_cost_idr",
		"created_at":    "pu.created_at",
	}, "pu.purchase_date")

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT %s,
			pr.name AS product_name,
			pr.sku  AS product_sku,
			u.name  AS purchaser_name,
			COALESCE((SELECT sum(pa.qty) FROM purchase_allocations pa
			          WHERE pa.purchase_id = pu.id AND pa.order_item_id IS NOT NULL), 0)::int AS qty_to_orders,
			COALESCE((SELECT sum(pa.qty) FROM purchase_allocations pa
			          WHERE pa.purchase_id = pu.id AND pa.order_item_id IS NULL), 0)::int AS qty_to_stock
		FROM purchases pu
		JOIN products pr ON pr.id = pu.product_id
		LEFT JOIN users u ON u.id = pu.purchased_by%s
		ORDER BY %s %s, pu.created_at DESC
		LIMIT $%d OFFSET $%d`,
		purchaseColumnsPrefixed, where, sortCol, p.Order, len(args)-1, len(args))

	purchases, err := collectRows[domain.PurchaseDetail](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return purchases, total, nil
}

func (r *PurchaseRepo) Update(ctx context.Context, q db.Querier, id uuid.UUID, p PurchaseParams) (*domain.Purchase, error) {
	return collectOne[domain.Purchase](ctx, q, "pembelian", `
		UPDATE purchases
		SET purchase_date = $2, qty = $3, unit_cost_foreign = $4, currency = $5, exchange_rate = $6,
		    unit_cost_idr = $7, total_cost_idr = $8, store_name = $9, receipt_url = $10, notes = $11
		WHERE id = $1
		RETURNING `+purchaseColumns,
		id, p.PurchaseDate, p.Qty, p.UnitCostForeign, p.Currency, p.ExchangeRate,
		p.UnitCostIDR, p.TotalCostIDR, p.StoreName, p.ReceiptURL, p.Notes)
}

func (r *PurchaseRepo) Delete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "pembelian", `DELETE FROM purchases WHERE id = $1`, id)
}

// TotalCostByTrip menjumlahkan seluruh uang belanja pada satu trip, termasuk
// yang berakhir sebagai stok.
func (r *PurchaseRepo) TotalCostByTrip(ctx context.Context, q db.Querier, tripID uuid.UUID) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := q.QueryRow(ctx,
		`SELECT COALESCE(sum(total_cost_idr), 0) FROM purchases WHERE trip_id = $1`, tripID).Scan(&total)
	if err != nil {
		return decimal.Zero, wrapPgError(err)
	}
	return total, nil
}

// --- Alokasi pembelian -----------------------------------------------------

const allocationColumns = `id, purchase_id, order_item_id, qty, unit_cost_idr, created_at`

func (r *PurchaseRepo) CreateAllocation(ctx context.Context, q db.Querier, purchaseID uuid.UUID, orderItemID *uuid.UUID, qty int, unitCost decimal.Decimal) (*domain.PurchaseAllocation, error) {
	return collectOne[domain.PurchaseAllocation](ctx, q, "alokasi pembelian", `
		INSERT INTO purchase_allocations (purchase_id, order_item_id, qty, unit_cost_idr)
		VALUES ($1, $2, $3, $4)
		RETURNING `+allocationColumns,
		purchaseID, orderItemID, qty, unitCost)
}

func (r *PurchaseRepo) ListAllocations(ctx context.Context, q db.Querier, purchaseID uuid.UUID) ([]domain.PurchaseAllocationDetail, error) {
	return collectRows[domain.PurchaseAllocationDetail](ctx, q, `
		SELECT pa.id, pa.purchase_id, pa.order_item_id, pa.qty, pa.unit_cost_idr, pa.created_at,
			o.order_number AS order_number,
			c.name         AS customer_name,
			pr.name        AS product_name
		FROM purchase_allocations pa
		JOIN purchases pu       ON pu.id = pa.purchase_id
		JOIN products pr        ON pr.id = pu.product_id
		LEFT JOIN order_items oi ON oi.id = pa.order_item_id
		LEFT JOIN orders o       ON o.id = oi.order_id
		LEFT JOIN customers c    ON c.id = o.customer_id
		WHERE pa.purchase_id = $1
		ORDER BY pa.created_at ASC`, purchaseID)
}

func (r *PurchaseRepo) DeleteAllocationsByPurchase(ctx context.Context, q db.Querier, purchaseID uuid.UUID) (int64, error) {
	return exec(ctx, q, `DELETE FROM purchase_allocations WHERE purchase_id = $1`, purchaseID)
}

// ListAllocationsByOrderItem mengembalikan alokasi sebuah baris pesanan,
// diurutkan dari yang paling baru. Urutan ini dipakai saat qty pesanan
// dikurangi: alokasi termuda yang dilepas lebih dulu.
func (r *PurchaseRepo) ListAllocationsByOrderItem(ctx context.Context, q db.Querier, orderItemID uuid.UUID) ([]domain.PurchaseAllocation, error) {
	return collectRows[domain.PurchaseAllocation](ctx, q,
		`SELECT `+allocationColumns+` FROM purchase_allocations
		 WHERE order_item_id = $1
		 ORDER BY created_at DESC`, orderItemID)
}

func (r *PurchaseRepo) UpdateAllocationQty(ctx context.Context, q db.Querier, id uuid.UUID, qty int) error {
	return execExpectOne(ctx, q, "alokasi pembelian",
		`UPDATE purchase_allocations SET qty = $2 WHERE id = $1`, id, qty)
}

func (r *PurchaseRepo) DeleteAllocation(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "alokasi pembelian",
		`DELETE FROM purchase_allocations WHERE id = $1`, id)
}

// AllocatedQtyByOrderItem menghitung total unit yang sudah dialokasikan ke satu
// baris pesanan.
func (r *PurchaseRepo) AllocatedQtyByOrderItem(ctx context.Context, q db.Querier, orderItemID uuid.UUID) (int, error) {
	var qty int
	err := q.QueryRow(ctx,
		`SELECT COALESCE(sum(qty), 0)::int FROM purchase_allocations WHERE order_item_id = $1`,
		orderItemID).Scan(&qty)
	if err != nil {
		return 0, wrapPgError(err)
	}
	return qty, nil
}

// AllocatedQty menghitung berapa unit dari sebuah pembelian yang sudah
// dialokasikan, dipakai untuk memastikan alokasi tidak melebihi qty dibeli.
func (r *PurchaseRepo) AllocatedQty(ctx context.Context, q db.Querier, purchaseID uuid.UUID) (int, error) {
	var qty int
	err := q.QueryRow(ctx,
		`SELECT COALESCE(sum(qty), 0)::int FROM purchase_allocations WHERE purchase_id = $1`,
		purchaseID).Scan(&qty)
	if err != nil {
		return 0, wrapPgError(err)
	}
	return qty, nil
}

// PendingOrderItem adalah baris pesanan yang masih menunggu dibelikan.
// Diurutkan FIFO berdasarkan tanggal order supaya customer yang pesan lebih
// dulu terpenuhi lebih dulu saat barang di toko terbatas.
type PendingOrderItem struct {
	OrderItemID  uuid.UUID `db:"order_item_id"`
	OrderID      uuid.UUID `db:"order_id"`
	OrderNumber  string    `db:"order_number"`
	CustomerName string    `db:"customer_name"`
	QtyNeeded    int       `db:"qty_needed"`
}

func (r *PurchaseRepo) ListPendingOrderItems(ctx context.Context, q db.Querier, tripID, productID uuid.UUID) ([]PendingOrderItem, error) {
	return collectRows[PendingOrderItem](ctx, q, `
		SELECT
			oi.id           AS order_item_id,
			o.id            AS order_id,
			o.order_number  AS order_number,
			c.name          AS customer_name,
			(oi.qty - COALESCE(alloc.qty, 0))::int AS qty_needed
		FROM order_items oi
		JOIN orders o    ON o.id = oi.order_id
		JOIN customers c ON c.id = o.customer_id
		LEFT JOIN LATERAL (
			SELECT COALESCE(sum(pa.qty), 0)::int AS qty
			FROM purchase_allocations pa
			WHERE pa.order_item_id = oi.id
		) alloc ON TRUE
		WHERE o.trip_id = $1
		  AND oi.product_id = $2
		  AND o.status NOT IN ('cancelled', 'draft')
		  AND oi.fulfillment_status NOT IN ('unavailable', 'refunded')
		  AND oi.qty > COALESCE(alloc.qty, 0)
		ORDER BY o.order_date ASC, o.created_at ASC`, tripID, productID)
}

// --- Daftar belanja --------------------------------------------------------

// ShoppingList mengagregasi pesanan pada satu trip menjadi daftar belanja per
// produk. Sengaja dihitung dari order, bukan disimpan sebagai tabel, supaya
// edit qty pesanan langsung tercermin di daftar tripper.
//
// Yang masuk hitungan belanja hanyalah order yang DP-nya sudah diverifikasi.
// Order yang masih menunggu DP dihitung terpisah pada qty_awaiting_dp: tripper
// perlu tahu ada permintaan tertahan, tapi tidak boleh membelanjakannya karena
// uang mukanya belum ada.
func (r *PurchaseRepo) ShoppingList(ctx context.Context, q db.Querier, tripID uuid.UUID) ([]domain.ShoppingListEntry, error) {
	return collectRows[domain.ShoppingListEntry](ctx, q, `
		SELECT
			p.id         AS product_id,
			p.name       AS product_name,
			p.sku        AS product_sku,
			p.brand      AS brand,
			p.store_name AS store_name,
			p.image_url  AS image_url,
			c.name       AS category_name,
			COALESCE(sum(oi.qty) FILTER (WHERE o.status <> 'awaiting_dp'), 0)::int AS qty_ordered,
			COALESCE(sum(oi.qty) FILTER (WHERE o.status =  'awaiting_dp'), 0)::int AS qty_awaiting_dp,
			COALESCE((SELECT sum(pu.qty) FROM purchases pu
			          WHERE pu.trip_id = $1 AND pu.product_id = p.id), 0)::int AS qty_purchased,
			GREATEST(
				COALESCE(sum(oi.qty) FILTER (WHERE o.status <> 'awaiting_dp'), 0)
					- COALESCE((SELECT sum(pu.qty) FROM purchases pu
					            WHERE pu.trip_id = $1 AND pu.product_id = p.id), 0),
				0
			)::int AS qty_remaining,
			count(DISTINCT o.id) FILTER (WHERE o.status <> 'awaiting_dp')::int AS order_count,
			COALESCE(sum(oi.qty * oi.unit_cost_est) FILTER (WHERE o.status <> 'awaiting_dp'), 0) AS est_cost_idr,
			COALESCE(max(ti.cost_price), 0) AS cost_price,
			COALESCE(max(ti.sell_price), 0) AS sell_price_idr
		FROM order_items oi
		JOIN orders o    ON o.id = oi.order_id
		JOIN products p  ON p.id = oi.product_id
		LEFT JOIN product_categories c ON c.id = p.category_id
		LEFT JOIN trip_items ti ON ti.trip_id = o.trip_id AND ti.product_id = oi.product_id
		WHERE o.trip_id = $1
		  AND o.status NOT IN ('cancelled', 'draft')
		  AND oi.fulfillment_status <> 'refunded'
		GROUP BY p.id, p.name, p.sku, p.brand, p.store_name, p.image_url, c.name
		ORDER BY p.name ASC`, tripID)
}

// --- Stok ------------------------------------------------------------------

const stockColumns = `id, product_id, qty_on_hand, avg_cost_idr, location, updated_at`

// StockIn menambah stok dan memperbarui harga pokok rata-rata bergerak.
// Rata-rata dipakai (bukan FIFO) karena barang jastip identik dan admin tidak
// melacak batch per unit; ini menjaga nilai stok tetap masuk akal saat harga
// beli berbeda antar trip.
func (r *PurchaseRepo) StockIn(ctx context.Context, q db.Querier, productID uuid.UUID, qty int, unitCost decimal.Decimal) (*domain.StockItem, error) {
	return collectOne[domain.StockItem](ctx, q, "stok", `
		INSERT INTO stock_items (product_id, qty_on_hand, avg_cost_idr)
		VALUES ($1, $2, $3)
		ON CONFLICT (product_id) DO UPDATE
		SET qty_on_hand  = stock_items.qty_on_hand + EXCLUDED.qty_on_hand,
		    avg_cost_idr = CASE
		        WHEN stock_items.qty_on_hand + EXCLUDED.qty_on_hand = 0 THEN 0
		        ELSE ((stock_items.qty_on_hand * stock_items.avg_cost_idr)
		              + (EXCLUDED.qty_on_hand * EXCLUDED.avg_cost_idr))
		             / (stock_items.qty_on_hand + EXCLUDED.qty_on_hand)
		    END
		RETURNING `+stockColumns,
		productID, qty, unitCost)
}

// StockOut mengurangi stok. Syarat qty_on_hand >= qty ditaruh di WHERE supaya
// pengurangan yang melebihi stok tidak terjadi sama sekali, bukan ditolak
// belakangan oleh CHECK constraint.
func (r *PurchaseRepo) StockOut(ctx context.Context, q db.Querier, productID uuid.UUID, qty int) (*domain.StockItem, error) {
	items, err := collectRows[domain.StockItem](ctx, q, `
		UPDATE stock_items
		SET qty_on_hand = qty_on_hand - $2
		WHERE product_id = $1 AND qty_on_hand >= $2
		RETURNING `+stockColumns, productID, qty)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.Conflict("stok produk tidak mencukupi untuk dikurangi %d unit", qty)
	}
	return &items[0], nil
}

func (r *PurchaseRepo) GetStock(ctx context.Context, q db.Querier, productID uuid.UUID) (*domain.StockItem, error) {
	return collectOne[domain.StockItem](ctx, q, "stok",
		`SELECT `+stockColumns+` FROM stock_items WHERE product_id = $1`, productID)
}

func (r *PurchaseRepo) ListStock(ctx context.Context, q db.Querier, p pagination.Params, inStockOnly bool) ([]domain.StockItemDetail, int64, error) {
	conditions := []string{}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf("(pr.name ILIKE $%d OR pr.sku ILIKE $%d)", n, n))
	}
	if inStockOnly {
		conditions = append(conditions, "si.qty_on_hand > 0")
	}
	where := buildWhere(conditions)

	var total int64
	countQuery := `SELECT count(*) FROM stock_items si JOIN products pr ON pr.id = si.product_id` + where
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"name":        "pr.name",
		"qty":         "si.qty_on_hand",
		"stock_value": "(si.qty_on_hand * si.avg_cost_idr)",
		"updated_at":  "si.updated_at",
	}, "pr.name")
	order := p.Order
	if p.Sort == "" {
		order = "asc"
	}

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT si.id, si.product_id, si.qty_on_hand, si.avg_cost_idr, si.location, si.updated_at,
			pr.name      AS product_name,
			pr.sku       AS product_sku,
			pr.brand     AS brand,
			pr.image_url AS image_url,
			c.name       AS category_name,
			(si.qty_on_hand * si.avg_cost_idr) AS stock_value
		FROM stock_items si
		JOIN products pr ON pr.id = si.product_id
		LEFT JOIN product_categories c ON c.id = pr.category_id%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		where, sortCol, order, len(args)-1, len(args))

	items, err := collectRows[domain.StockItemDetail](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *PurchaseRepo) SetStockLocation(ctx context.Context, q db.Querier, productID uuid.UUID, location *string) error {
	return execExpectOne(ctx, q, "stok",
		`UPDATE stock_items SET location = $2 WHERE product_id = $1`, productID, location)
}

// AdjustStock menyetel qty stok ke angka hasil stock opname.
func (r *PurchaseRepo) AdjustStock(ctx context.Context, q db.Querier, productID uuid.UUID, newQty int) (*domain.StockItem, error) {
	return collectOne[domain.StockItem](ctx, q, "stok", `
		INSERT INTO stock_items (product_id, qty_on_hand, avg_cost_idr)
		VALUES ($1, $2, 0)
		ON CONFLICT (product_id) DO UPDATE SET qty_on_hand = EXCLUDED.qty_on_hand
		RETURNING `+stockColumns, productID, newQty)
}

// --- Pergerakan stok -------------------------------------------------------

const movementColumns = `id, product_id, type, qty, unit_cost_idr, sale_price_idr, trip_id,
	                     ref_type, ref_id, note, created_by, created_at`

type StockMovementParams struct {
	ProductID    uuid.UUID
	Type         string
	Qty          int
	UnitCostIDR  decimal.Decimal
	SalePriceIDR *decimal.Decimal
	TripID       *uuid.UUID
	RefType      *string
	RefID        *uuid.UUID
	Note         *string
	CreatedBy    *uuid.UUID
}

func (r *PurchaseRepo) CreateMovement(ctx context.Context, q db.Querier, p StockMovementParams) (*domain.StockMovement, error) {
	return collectOne[domain.StockMovement](ctx, q, "pergerakan stok", `
		INSERT INTO stock_movements (product_id, type, qty, unit_cost_idr, sale_price_idr, trip_id,
		                             ref_type, ref_id, note, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING `+movementColumns,
		p.ProductID, p.Type, p.Qty, p.UnitCostIDR, p.SalePriceIDR, p.TripID,
		p.RefType, p.RefID, p.Note, p.CreatedBy)
}

func (r *PurchaseRepo) ListMovements(ctx context.Context, q db.Querier, p pagination.Params, productID *uuid.UUID) ([]domain.StockMovementDetail, int64, error) {
	conditions := []string{}
	args := []any{}

	if productID != nil {
		args = append(args, *productID)
		conditions = append(conditions, fmt.Sprintf("sm.product_id = $%d", len(args)))
	}
	where := buildWhere(conditions)

	var total int64
	if err := q.QueryRow(ctx, `SELECT count(*) FROM stock_movements sm`+where, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT sm.id, sm.product_id, sm.type, sm.qty, sm.unit_cost_idr, sm.sale_price_idr,
		       sm.trip_id, sm.ref_type, sm.ref_id, sm.note, sm.created_by, sm.created_at,
			pr.name AS product_name,
			pr.sku  AS product_sku,
			u.name  AS actor_name
		FROM stock_movements sm
		JOIN products pr ON pr.id = sm.product_id
		LEFT JOIN users u ON u.id = sm.created_by%s
		ORDER BY sm.created_at DESC
		LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))

	movements, err := collectRows[domain.StockMovementDetail](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return movements, total, nil
}
