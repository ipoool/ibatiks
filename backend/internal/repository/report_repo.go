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

type ReportRepo struct{}

func NewReportRepo() *ReportRepo { return &ReportRepo{} }

// TripFinancials adalah angka mentah laporan trip. Penjumlahan dan pembagian
// akhirnya dikerjakan di service memakai decimal, bukan di SQL, supaya
// pembulatannya konsisten dengan perhitungan uang di tempat lain.
type TripFinancials struct {
	Revenue              decimal.Decimal `db:"revenue"`
	COGS                 decimal.Decimal `db:"cogs"`
	TripExpenses         decimal.Decimal `db:"trip_expenses"`
	PurchaseTotal        decimal.Decimal `db:"purchase_total"`
	ShippingFeeCollected decimal.Decimal `db:"shipping_fee_collected"`
	ShippingCostPaid     decimal.Decimal `db:"shipping_cost_paid"`
	DiscountGiven        decimal.Decimal `db:"discount_given"`
	PaymentReceived      decimal.Decimal `db:"payment_received"`
	Outstanding          decimal.Decimal `db:"outstanding"`
	SurplusStockQty      int             `db:"surplus_stock_qty"`
	SurplusStockValue    decimal.Decimal `db:"surplus_stock_value"`
	OrderCount           int             `db:"order_count"`
	CustomerCount        int             `db:"customer_count"`
	ItemQty              int             `db:"item_qty"`
}

// TripFinancials menghitung seluruh angka keuangan satu trip dalam sekali
// query. Dipisah per CTE supaya tiap komponen bisa dibaca dan diverifikasi
// sendiri-sendiri saat angka laporan terasa aneh.
// TripFinancials merangkum angka keuangan sebuah trip, atau seluruh trip
// sekaligus bila tripID nil.
//
// Penyaringnya ditulis sebagai `($1::uuid IS NULL OR trip_id = $1)` supaya satu
// kueri melayani keduanya. Menggandakan kuerinya untuk kasus "semua trip"
// berarti dua definisi laba yang harus dijaga tetap sama — dan cepat atau
// lambat salah satunya tertinggal saat rumusnya berubah.
func (r *ReportRepo) TripFinancials(ctx context.Context, q db.Querier, tripID *uuid.UUID) (*TripFinancials, error) {
	return collectOne[TripFinancials](ctx, q, "laporan trip", `
		WITH trip_orders AS (
			-- Order draft dan batal tidak dihitung sebagai omzet.
			SELECT id, customer_id, total, shipping_fee, discount, paid_amount, balance_due
			FROM orders
			WHERE ($1::uuid IS NULL OR trip_id = $1) AND status <> 'cancelled'
		),
		order_totals AS (
			SELECT
				COALESCE(sum(total), 0)                     AS revenue,
				COALESCE(sum(shipping_fee), 0)              AS shipping_fee_collected,
				COALESCE(sum(discount), 0)                  AS discount_given,
				COALESCE(sum(paid_amount), 0)               AS payment_received,
				COALESCE(sum(GREATEST(balance_due, 0)), 0)  AS outstanding,
				count(*)::int                               AS order_count,
				count(DISTINCT customer_id)::int            AS customer_count
			FROM trip_orders
		),
		ordered_qty AS (
			SELECT COALESCE(sum(oi.qty), 0)::int AS item_qty
			FROM order_items oi
			JOIN trip_orders o ON o.id = oi.order_id
		),
		cogs AS (
			-- HPP riil: biaya belanja yang benar-benar dialokasikan ke pesanan.
			SELECT COALESCE(sum(pa.qty * pa.unit_cost_idr), 0) AS cogs
			FROM purchase_allocations pa
			JOIN order_items oi ON oi.id = pa.order_item_id
			JOIN trip_orders o  ON o.id = oi.order_id
		),
		surplus AS (
			-- Belanja yang tidak dipesan siapa pun; jadi aset stok, bukan beban trip.
			SELECT
				COALESCE(sum(pa.qty), 0)::int                 AS surplus_stock_qty,
				COALESCE(sum(pa.qty * pa.unit_cost_idr), 0)   AS surplus_stock_value
			FROM purchase_allocations pa
			JOIN purchases pu ON pu.id = pa.purchase_id
			WHERE ($1::uuid IS NULL OR pu.trip_id = $1) AND pa.order_item_id IS NULL
		),
		expenses AS (
			SELECT COALESCE(sum(amount), 0) AS trip_expenses
			FROM trip_expenses WHERE ($1::uuid IS NULL OR trip_id = $1)
		),
		purchase_total AS (
			SELECT COALESCE(sum(total_cost_idr), 0) AS purchase_total
			FROM purchases WHERE ($1::uuid IS NULL OR trip_id = $1)
		),
		shipping AS (
			SELECT COALESCE(sum(s.shipping_cost), 0) AS shipping_cost_paid
			FROM shipments s
			JOIN trip_orders o ON o.id = s.order_id
		)
		SELECT
			ot.revenue, cg.cogs, ex.trip_expenses, pt.purchase_total,
			ot.shipping_fee_collected, sh.shipping_cost_paid, ot.discount_given,
			ot.payment_received, ot.outstanding,
			su.surplus_stock_qty, su.surplus_stock_value,
			ot.order_count, ot.customer_count, oq.item_qty
		FROM order_totals ot, cogs cg, expenses ex, purchase_total pt,
		     shipping sh, surplus su, ordered_qty oq`, tripID)
}

func (r *ReportRepo) ExpenseBreakdown(ctx context.Context, q db.Querier, tripID *uuid.UUID) ([]domain.ExpenseBreakdown, error) {
	return collectRows[domain.ExpenseBreakdown](ctx, q, `
		SELECT category, sum(amount) AS amount
		FROM trip_expenses
		WHERE ($1::uuid IS NULL OR trip_id = $1)
		GROUP BY category
		ORDER BY sum(amount) DESC`, tripID)
}

// OrderProfits menghitung margin tiap order. Order yang HPP-nya belum tercatat
// akan tampil dengan cogs 0, yang justru berguna sebagai penanda bahwa
// pembelian untuk order itu belum diinput.
func (r *ReportRepo) OrderProfits(ctx context.Context, q db.Querier, p pagination.Params, tripID *uuid.UUID) ([]domain.OrderProfit, int64, error) {
	conditions := []string{"o.status <> 'cancelled'"}
	args := []any{}

	if tripID != nil {
		args = append(args, *tripID)
		conditions = append(conditions, fmt.Sprintf("o.trip_id = $%d", len(args)))
	}
	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf("(o.order_number ILIKE $%d OR c.name ILIKE $%d)", n, n))
	}
	where := buildWhere(conditions)

	var total int64
	countQuery := `SELECT count(*) FROM orders o JOIN customers c ON c.id = o.customer_id` + where
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT
			o.id           AS order_id,
			o.order_number AS order_number,
			c.name         AS customer_name,
			t.code         AS trip_code,
			o.status       AS status,
			o.order_date   AS order_date,
			o.total        AS revenue,
			COALESCE(cg.cogs, 0) AS cogs,
			(o.total - COALESCE(cg.cogs, 0)) AS profit,
			CASE WHEN o.total > 0
			     THEN round((o.total - COALESCE(cg.cogs, 0)) / o.total * 100, 2)
			     ELSE 0 END AS margin_pct
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN trips t     ON t.id = o.trip_id
		LEFT JOIN LATERAL (
			SELECT sum(pa.qty * pa.unit_cost_idr) AS cogs
			FROM purchase_allocations pa
			JOIN order_items oi ON oi.id = pa.order_item_id
			WHERE oi.order_id = o.id
		) cg ON TRUE%s
		ORDER BY o.order_date DESC, o.created_at DESC
		LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))

	profits, err := collectRows[domain.OrderProfit](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return profits, total, nil
}

// Receivables mendaftar order yang masih punya sisa tagihan, diurutkan dari
// yang paling lama menunggak.
func (r *ReportRepo) Receivables(ctx context.Context, q db.Querier, p pagination.Params) ([]domain.Receivable, int64, error) {
	conditions := []string{"o.balance_due > 0", "o.status <> 'cancelled'"}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf("(o.order_number ILIKE $%d OR c.name ILIKE $%d)", n, n))
	}
	where := buildWhere(conditions)

	var total int64
	countQuery := `SELECT count(*) FROM orders o JOIN customers c ON c.id = o.customer_id` + where
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT
			o.id           AS order_id,
			o.order_number AS order_number,
			c.id           AS customer_id,
			c.name         AS customer_name,
			c.phone_wa     AS customer_phone,
			t.code         AS trip_code,
			o.status       AS status,
			o.order_date   AS order_date,
			o.total        AS total,
			o.paid_amount  AS paid_amount,
			o.balance_due  AS balance_due,
			(CURRENT_DATE - o.order_date)::int AS days_outstanding
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN trips t     ON t.id = o.trip_id%s
		ORDER BY o.order_date ASC
		LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))

	receivables, err := collectRows[domain.Receivable](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return receivables, total, nil
}

// ProductSales merangkum performa tiap produk pada rentang tanggal tertentu.
func (r *ReportRepo) ProductSales(ctx context.Context, q db.Querier, limit int, tripID *uuid.UUID, from, to *time.Time) ([]domain.ProductSales, error) {
	conditions := []string{"o.status <> 'cancelled'"}
	args := []any{}

	if tripID != nil {
		args = append(args, *tripID)
		conditions = append(conditions, fmt.Sprintf("o.trip_id = $%d", len(args)))
	}
	if from != nil {
		args = append(args, *from)
		conditions = append(conditions, fmt.Sprintf("o.order_date >= $%d", len(args)))
	}
	if to != nil {
		args = append(args, *to)
		conditions = append(conditions, fmt.Sprintf("o.order_date <= $%d", len(args)))
	}
	where := buildWhere(conditions)

	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT
			p.id   AS product_id,
			p.name AS product_name,
			p.sku  AS product_sku,
			c.name AS category_name,
			sum(oi.qty)::int          AS qty_sold,
			count(DISTINCT o.id)::int AS order_count,
			sum(oi.subtotal)          AS revenue,
			COALESCE(sum(alloc.cogs), 0) AS cogs,
			(sum(oi.subtotal) - COALESCE(sum(alloc.cogs), 0)) AS profit
		FROM order_items oi
		JOIN orders o   ON o.id = oi.order_id
		JOIN products p ON p.id = oi.product_id
		LEFT JOIN product_categories c ON c.id = p.category_id
		LEFT JOIN LATERAL (
			SELECT sum(pa.qty * pa.unit_cost_idr) AS cogs
			FROM purchase_allocations pa
			WHERE pa.order_item_id = oi.id
		) alloc ON TRUE%s
		GROUP BY p.id, p.name, p.sku, c.name
		ORDER BY sum(oi.qty) DESC
		LIMIT $%d`, where, len(args))

	return collectRows[domain.ProductSales](ctx, q, query, args...)
}

// CustomerSales merangkum belanja tiap customer. HPP diambil dari alokasi
// pembelian yang nyata, sama seperti laporan lainnya, supaya angka profitnya
// konsisten di seluruh aplikasi.
func (r *ReportRepo) CustomerSales(ctx context.Context, q db.Querier, p pagination.Params, tripID *uuid.UUID) ([]domain.CustomerSales, int64, error) {
	conditions := []string{"o.status <> 'cancelled'"}
	args := []any{}

	if tripID != nil {
		args = append(args, *tripID)
		conditions = append(conditions, fmt.Sprintf("o.trip_id = $%d", len(args)))
	}
	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf("(c.name ILIKE $%d OR c.phone_wa ILIKE $%d)", n, n))
	}
	where := buildWhere(conditions)

	var total int64
	countQuery := `
		SELECT count(DISTINCT o.customer_id)
		FROM orders o JOIN customers c ON c.id = o.customer_id` + where
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"revenue":     "sum(o.total)",
		"profit":      "sum(o.total) - COALESCE(sum(cg.cogs), 0)",
		"order_count": "count(DISTINCT o.id)",
		"name":        "c.name",
	}, "sum(o.total)")

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT
			c.id       AS customer_id,
			c.code     AS customer_code,
			c.name     AS customer_name,
			c.phone_wa AS customer_phone,
			c.city     AS city,
			count(DISTINCT o.id)::int                      AS order_count,
			COALESCE(sum(iq.qty), 0)::int                  AS item_qty,
			sum(o.total)                                   AS revenue,
			COALESCE(sum(cg.cogs), 0)                      AS cogs,
			(sum(o.total) - COALESCE(sum(cg.cogs), 0))     AS profit,
			COALESCE(sum(GREATEST(o.balance_due, 0)), 0)   AS outstanding,
			round(sum(o.total) / count(DISTINCT o.id), 2)  AS avg_order_value,
			min(o.order_date)                              AS first_order_at,
			max(o.order_date)                              AS last_order_at
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		LEFT JOIN LATERAL (
			SELECT sum(pa.qty * pa.unit_cost_idr) AS cogs
			FROM purchase_allocations pa
			JOIN order_items oi ON oi.id = pa.order_item_id
			WHERE oi.order_id = o.id
		) cg ON TRUE
		LEFT JOIN LATERAL (
			SELECT sum(oi.qty) AS qty FROM order_items oi WHERE oi.order_id = o.id
		) iq ON TRUE%s
		GROUP BY c.id, c.code, c.name, c.phone_wa, c.city
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`, where, sortCol, p.Order, len(args)-1, len(args))

	rows, err := collectRows[domain.CustomerSales](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// CustomerSalesByChannel memecah rekap customer menjadi satu baris per kanal.
//
// Tanpa paginasi: hasilnya dipakai untuk ekspor CSV, dan ekspor yang berhenti
// di halaman pertama diam-diam kehilangan sebagian besar datanya tanpa ada
// tanda apa pun di berkasnya.
func (r *ReportRepo) CustomerSalesByChannel(ctx context.Context, q db.Querier, tripID *uuid.UUID) ([]domain.CustomerChannelSales, error) {
	conditions := []string{"o.status <> 'cancelled'"}
	args := []any{}

	if tripID != nil {
		args = append(args, *tripID)
		conditions = append(conditions, fmt.Sprintf("o.trip_id = $%d", len(args)))
	}
	where := buildWhere(conditions)

	// Baris tiap customer dikelompokkan berdampingan dan customer dengan omzet
	// terbesar di atas, mengikuti urutan yang sama seperti di layar. Jendela
	// dipakai atas hasil agregat, jadi sum-nya bertingkat.
	query := fmt.Sprintf(`
		SELECT
			c.code     AS customer_code,
			c.name     AS customer_name,
			c.phone_wa AS customer_phone,
			c.city     AS city,
			o.order_source                                 AS order_source,
			count(DISTINCT o.id)::int                      AS order_count,
			COALESCE(sum(iq.qty), 0)::int                  AS item_qty,
			sum(o.total)                                   AS revenue,
			COALESCE(sum(cg.cogs), 0)                      AS cogs,
			(sum(o.total) - COALESCE(sum(cg.cogs), 0))     AS profit,
			COALESCE(sum(GREATEST(o.balance_due, 0)), 0)   AS outstanding,
			min(o.order_date)                              AS first_order_at,
			max(o.order_date)                              AS last_order_at
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		LEFT JOIN LATERAL (
			SELECT sum(pa.qty * pa.unit_cost_idr) AS cogs
			FROM purchase_allocations pa
			JOIN order_items oi ON oi.id = pa.order_item_id
			WHERE oi.order_id = o.id
		) cg ON TRUE
		LEFT JOIN LATERAL (
			SELECT sum(oi.qty) AS qty FROM order_items oi WHERE oi.order_id = o.id
		) iq ON TRUE%s
		GROUP BY c.id, c.code, c.name, c.phone_wa, c.city, o.order_source
		ORDER BY sum(sum(o.total)) OVER (PARTITION BY c.id) DESC, c.name, o.order_source`, where)

	return collectRows[domain.CustomerChannelSales](ctx, q, query, args...)
}

// ChannelSales merangkum penjualan per asal order. Jumlah barisnya selalu
// sedikit (satu per kanal), jadi tidak perlu dihalaman.
func (r *ReportRepo) ChannelSales(ctx context.Context, q db.Querier, tripID *uuid.UUID, from, to *time.Time) ([]domain.ChannelSales, error) {
	conditions := []string{"o.status <> 'cancelled'"}
	args := []any{}

	if tripID != nil {
		args = append(args, *tripID)
		conditions = append(conditions, fmt.Sprintf("o.trip_id = $%d", len(args)))
	}
	if from != nil {
		args = append(args, *from)
		conditions = append(conditions, fmt.Sprintf("o.order_date >= $%d", len(args)))
	}
	if to != nil {
		args = append(args, *to)
		conditions = append(conditions, fmt.Sprintf("o.order_date <= $%d", len(args)))
	}
	where := buildWhere(conditions)

	query := fmt.Sprintf(`
		SELECT
			o.order_source                                 AS order_source,
			count(DISTINCT o.id)::int                      AS order_count,
			count(DISTINCT o.customer_id)::int             AS customer_count,
			COALESCE(sum(iq.qty), 0)::int                  AS item_qty,
			sum(o.total)                                   AS revenue,
			COALESCE(sum(cg.cogs), 0)                      AS cogs,
			(sum(o.total) - COALESCE(sum(cg.cogs), 0))     AS profit,
			round(sum(o.total) / count(DISTINCT o.id), 2)  AS avg_order_value
		FROM orders o
		LEFT JOIN LATERAL (
			SELECT sum(pa.qty * pa.unit_cost_idr) AS cogs
			FROM purchase_allocations pa
			JOIN order_items oi ON oi.id = pa.order_item_id
			WHERE oi.order_id = o.id
		) cg ON TRUE
		LEFT JOIN LATERAL (
			SELECT sum(oi.qty) AS qty FROM order_items oi WHERE oi.order_id = o.id
		) iq ON TRUE%s
		GROUP BY o.order_source
		ORDER BY sum(o.total) DESC`, where)

	// Varian lax dipakai karena RevenueShare tidak punya kolom di query; porsi
	// omzet dihitung di service setelah seluruh baris terkumpul.
	return collectRowsLax[domain.ChannelSales](ctx, q, query, args...)
}

// DashboardCounters adalah angka ringkas untuk halaman depan.
type DashboardCounters struct {
	ActiveTrips      int             `db:"active_trips"`
	OpenOrders       int             `db:"open_orders"`
	PendingShipment  int             `db:"pending_shipment"`
	Outstanding      decimal.Decimal `db:"outstanding"`
	RevenueThisMonth decimal.Decimal `db:"revenue_this_month"`
	COGSThisMonth    decimal.Decimal `db:"cogs_this_month"`
	OrdersThisMonth  int             `db:"orders_this_month"`
	StockValue       decimal.Decimal `db:"stock_value"`
	StockQty         int             `db:"stock_qty"`
	CustomerCount    int             `db:"customer_count"`
}

func (r *ReportRepo) DashboardCounters(ctx context.Context, q db.Querier) (*DashboardCounters, error) {
	return collectOne[DashboardCounters](ctx, q, "dashboard", `
		WITH month_orders AS (
			SELECT id, total
			FROM orders
			WHERE status <> 'cancelled'
			  AND order_date >= date_trunc('month', CURRENT_DATE)
		)
		SELECT
			-- Trip aktif = yang masih menerima order. Trip yang sudah ditutup
			-- pekerjaannya diikuti lewat order dan pembeliannya sendiri.
			(SELECT count(*) FROM trips WHERE status = 'open')::int AS active_trips,
			(SELECT count(*) FROM orders
			 WHERE status NOT IN ('completed', 'cancelled'))::int AS open_orders,
			(SELECT count(*) FROM orders WHERE status = 'paid')::int AS pending_shipment,
			(SELECT COALESCE(sum(balance_due), 0) FROM orders
			 WHERE balance_due > 0 AND status <> 'cancelled') AS outstanding,
			(SELECT COALESCE(sum(total), 0) FROM month_orders) AS revenue_this_month,
			(SELECT COALESCE(sum(pa.qty * pa.unit_cost_idr), 0)
			 FROM purchase_allocations pa
			 JOIN order_items oi  ON oi.id = pa.order_item_id
			 JOIN month_orders mo ON mo.id = oi.order_id) AS cogs_this_month,
			(SELECT count(*) FROM month_orders)::int AS orders_this_month,
			(SELECT COALESCE(sum(qty_on_hand * avg_cost_idr), 0) FROM stock_items) AS stock_value,
			(SELECT COALESCE(sum(qty_on_hand), 0) FROM stock_items)::int AS stock_qty,
			(SELECT count(*) FROM customers WHERE deleted_at IS NULL)::int AS customer_count`)
}
