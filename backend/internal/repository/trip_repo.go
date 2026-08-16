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

const tripColumns = `id, code, title, country, city, tripper_user_id, depart_date, return_date,
	                 order_deadline, currency, exchange_rate, status, notes, created_by,
	                 created_at, updated_at`

const tripColumnsPrefixed = `t.id, t.code, t.title, t.country, t.city, t.tripper_user_id,
	                         t.depart_date, t.return_date, t.order_deadline, t.currency,
	                         t.exchange_rate, t.status, t.notes, t.created_by,
	                         t.created_at, t.updated_at`

type TripRepo struct{}

func NewTripRepo() *TripRepo { return &TripRepo{} }

type TripParams struct {
	Code          string
	Title         string
	Country       string
	City          *string
	TripperUserID *uuid.UUID
	DepartDate    time.Time
	ReturnDate    time.Time
	OrderDeadline *time.Time
	Currency      string
	ExchangeRate  decimal.Decimal
	Notes         *string
	CreatedBy     *uuid.UUID
}

func (r *TripRepo) Create(ctx context.Context, q db.Querier, p TripParams) (*domain.Trip, error) {
	return collectOne[domain.Trip](ctx, q, "trip", `
		INSERT INTO trips (code, title, country, city, tripper_user_id, depart_date, return_date,
		                   order_deadline, currency, exchange_rate, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING `+tripColumns,
		p.Code, p.Title, p.Country, p.City, p.TripperUserID, p.DepartDate, p.ReturnDate,
		p.OrderDeadline, p.Currency, p.ExchangeRate, p.Notes, p.CreatedBy)
}

func (r *TripRepo) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Trip, error) {
	return collectOne[domain.Trip](ctx, q, "trip",
		`SELECT `+tripColumns+` FROM trips WHERE id = $1`, id)
}

// GetDetail mengambil trip beserta angka ringkasan yang dipakai halaman detail.
func (r *TripRepo) GetDetail(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.TripDetail, error) {
	return collectOne[domain.TripDetail](ctx, q, "trip", `
		SELECT `+tripColumnsPrefixed+`,
			u.name AS tripper_name,
			(SELECT count(*)            FROM orders o WHERE o.trip_id = t.id AND o.status <> 'cancelled')::int AS total_orders,
			(SELECT count(DISTINCT o.customer_id) FROM orders o WHERE o.trip_id = t.id AND o.status <> 'cancelled')::int AS total_customers,
			(SELECT count(*)            FROM trip_items ti WHERE ti.trip_id = t.id)::int AS catalog_items
		FROM trips t
		LEFT JOIN users u ON u.id = t.tripper_user_id
		WHERE t.id = $1`, id)
}

func (r *TripRepo) List(ctx context.Context, q db.Querier, p pagination.Params, status string) ([]domain.TripDetail, int64, error) {
	conditions := []string{}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf(
			"(t.title ILIKE $%d OR t.code ILIKE $%d OR t.country ILIKE $%d)", n, n, n))
	}
	if status != "" {
		args = append(args, status)
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", len(args)))
	}
	where := buildWhere(conditions)

	var total int64
	if err := q.QueryRow(ctx, `SELECT count(*) FROM trips t`+where, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"depart_date": "t.depart_date",
		"return_date": "t.return_date",
		"code":        "t.code",
		"title":       "t.title",
		"created_at":  "t.created_at",
	}, "t.depart_date")

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT %s,
			u.name AS tripper_name,
			(SELECT count(*) FROM orders o WHERE o.trip_id = t.id AND o.status <> 'cancelled')::int AS total_orders,
			(SELECT count(DISTINCT o.customer_id) FROM orders o WHERE o.trip_id = t.id AND o.status <> 'cancelled')::int AS total_customers,
			(SELECT count(*) FROM trip_items ti WHERE ti.trip_id = t.id)::int AS catalog_items
		FROM trips t
		LEFT JOIN users u ON u.id = t.tripper_user_id%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		tripColumnsPrefixed, where, sortCol, p.Order, len(args)-1, len(args))

	trips, err := collectRows[domain.TripDetail](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return trips, total, nil
}

func (r *TripRepo) Update(ctx context.Context, q db.Querier, id uuid.UUID, p TripParams) (*domain.Trip, error) {
	return collectOne[domain.Trip](ctx, q, "trip", `
		UPDATE trips
		SET title = $2, country = $3, city = $4, tripper_user_id = $5, depart_date = $6,
		    return_date = $7, order_deadline = $8, currency = $9, exchange_rate = $10, notes = $11
		WHERE id = $1
		RETURNING `+tripColumns,
		id, p.Title, p.Country, p.City, p.TripperUserID, p.DepartDate, p.ReturnDate,
		p.OrderDeadline, p.Currency, p.ExchangeRate, p.Notes)
}

// UpdateExchangeRate mengubah kurs trip tanpa menyentuh data lainnya.
func (r *TripRepo) UpdateExchangeRate(ctx context.Context, q db.Querier, id uuid.UUID, rate decimal.Decimal) (*domain.Trip, error) {
	return collectOne[domain.Trip](ctx, q, "trip",
		`UPDATE trips SET exchange_rate = $2 WHERE id = $1 RETURNING `+tripColumns, id, rate)
}

func (r *TripRepo) UpdateStatus(ctx context.Context, q db.Querier, id uuid.UUID, status string) (*domain.Trip, error) {
	return collectOne[domain.Trip](ctx, q, "trip",
		`UPDATE trips SET status = $2 WHERE id = $1 RETURNING `+tripColumns, id, status)
}

func (r *TripRepo) Delete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "trip", `DELETE FROM trips WHERE id = $1`, id)
}

func (r *TripRepo) CountOrders(ctx context.Context, q db.Querier, tripID uuid.UUID) (int, error) {
	var count int
	err := q.QueryRow(ctx, `SELECT count(*) FROM orders WHERE trip_id = $1`, tripID).Scan(&count)
	if err != nil {
		return 0, wrapPgError(err)
	}
	return count, nil
}

// --- Katalog trip ----------------------------------------------------------

const tripItemColumns = `id, trip_id, product_id, cost_price, cost_price_idr, markup_type,
	                     markup_value, sell_price, max_qty, is_active, notes, created_at, updated_at`

const tripItemColumnsPrefixed = `ti.id, ti.trip_id, ti.product_id, ti.cost_price, ti.cost_price_idr,
	                             ti.markup_type, ti.markup_value, ti.sell_price, ti.max_qty,
	                             ti.is_active, ti.notes, ti.created_at, ti.updated_at`

type TripItemParams struct {
	TripID       uuid.UUID
	ProductID    uuid.UUID
	CostPrice    decimal.Decimal
	CostPriceIDR decimal.Decimal
	MarkupType   string
	MarkupValue  decimal.Decimal
	SellPrice    decimal.Decimal
	MaxQty       *int
	IsActive     bool
	Notes        *string
}

func (r *TripRepo) CreateItem(ctx context.Context, q db.Querier, p TripItemParams) (*domain.TripItem, error) {
	return collectOne[domain.TripItem](ctx, q, "item katalog", `
		INSERT INTO trip_items (trip_id, product_id, cost_price, cost_price_idr, markup_type,
		                        markup_value, sell_price, max_qty, is_active, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING `+tripItemColumns,
		p.TripID, p.ProductID, p.CostPrice, p.CostPriceIDR, p.MarkupType,
		p.MarkupValue, p.SellPrice, p.MaxQty, p.IsActive, p.Notes)
}

func (r *TripRepo) GetItem(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.TripItem, error) {
	return collectOne[domain.TripItem](ctx, q, "item katalog",
		`SELECT `+tripItemColumns+` FROM trip_items WHERE id = $1`, id)
}

// GetItemByProduct dipakai saat menambah order item: harga diambil dari katalog
// trip, bukan dari master produk.
func (r *TripRepo) GetItemByProduct(ctx context.Context, q db.Querier, tripID, productID uuid.UUID) (*domain.TripItem, error) {
	return collectOne[domain.TripItem](ctx, q, "item katalog",
		`SELECT `+tripItemColumns+` FROM trip_items WHERE trip_id = $1 AND product_id = $2`,
		tripID, productID)
}

// ListItems mengembalikan katalog satu trip lengkap dengan berapa yang sudah
// dipesan, supaya admin tahu sisa kuota tiap produk.
func (r *TripRepo) ListItems(ctx context.Context, q db.Querier, tripID uuid.UUID) ([]domain.TripItemDetail, error) {
	return collectRows[domain.TripItemDetail](ctx, q, `
		SELECT `+tripItemColumnsPrefixed+`,
			p.name       AS product_name,
			p.sku        AS product_sku,
			p.brand      AS brand,
			p.image_url  AS image_url,
			p.weight_gram AS weight_gram,
			c.name       AS category_name,
			COALESCE((
				SELECT sum(oi.qty)
				FROM order_items oi
				JOIN orders o ON o.id = oi.order_id
				WHERE oi.trip_item_id = ti.id AND o.status <> 'cancelled'
			), 0)::int   AS qty_ordered
		FROM trip_items ti
		JOIN products p ON p.id = ti.product_id
		LEFT JOIN product_categories c ON c.id = p.category_id
		WHERE ti.trip_id = $1
		ORDER BY p.name ASC`, tripID)
}

func (r *TripRepo) UpdateItem(ctx context.Context, q db.Querier, id uuid.UUID, p TripItemParams) (*domain.TripItem, error) {
	return collectOne[domain.TripItem](ctx, q, "item katalog", `
		UPDATE trip_items
		SET cost_price = $2, cost_price_idr = $3, markup_type = $4, markup_value = $5,
		    sell_price = $6, max_qty = $7, is_active = $8, notes = $9
		WHERE id = $1
		RETURNING `+tripItemColumns,
		id, p.CostPrice, p.CostPriceIDR, p.MarkupType, p.MarkupValue,
		p.SellPrice, p.MaxQty, p.IsActive, p.Notes)
}

func (r *TripRepo) DeleteItem(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "item katalog", `DELETE FROM trip_items WHERE id = $1`, id)
}

// CountItemOrders memberi tahu apakah item katalog sudah pernah dipesan.
func (r *TripRepo) CountItemOrders(ctx context.Context, q db.Querier, tripItemID uuid.UUID) (int, error) {
	var count int
	err := q.QueryRow(ctx, `
		SELECT COALESCE(sum(oi.qty), 0)::int
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.trip_item_id = $1 AND o.status <> 'cancelled'`, tripItemID).Scan(&count)
	if err != nil {
		return 0, wrapPgError(err)
	}
	return count, nil
}

// --- Biaya perjalanan ------------------------------------------------------

const tripExpenseColumns = `id, trip_id, category, description, amount, spent_at, receipt_url,
	                        created_by, created_at, updated_at`

type TripExpenseParams struct {
	TripID      uuid.UUID
	Category    string
	Description string
	Amount      decimal.Decimal
	SpentAt     time.Time
	ReceiptURL  *string
	CreatedBy   *uuid.UUID
}

func (r *TripRepo) CreateExpense(ctx context.Context, q db.Querier, p TripExpenseParams) (*domain.TripExpense, error) {
	return collectOne[domain.TripExpense](ctx, q, "biaya trip", `
		INSERT INTO trip_expenses (trip_id, category, description, amount, spent_at, receipt_url, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+tripExpenseColumns,
		p.TripID, p.Category, p.Description, p.Amount, p.SpentAt, p.ReceiptURL, p.CreatedBy)
}

func (r *TripRepo) ListExpenses(ctx context.Context, q db.Querier, tripID uuid.UUID) ([]domain.TripExpense, error) {
	return collectRows[domain.TripExpense](ctx, q,
		`SELECT `+tripExpenseColumns+` FROM trip_expenses WHERE trip_id = $1
		 ORDER BY spent_at DESC, created_at DESC`, tripID)
}

func (r *TripRepo) UpdateExpense(ctx context.Context, q db.Querier, id uuid.UUID, p TripExpenseParams) (*domain.TripExpense, error) {
	return collectOne[domain.TripExpense](ctx, q, "biaya trip", `
		UPDATE trip_expenses
		SET category = $2, description = $3, amount = $4, spent_at = $5, receipt_url = $6
		WHERE id = $1
		RETURNING `+tripExpenseColumns,
		id, p.Category, p.Description, p.Amount, p.SpentAt, p.ReceiptURL)
}

func (r *TripRepo) DeleteExpense(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "biaya trip", `DELETE FROM trip_expenses WHERE id = $1`, id)
}

func (r *TripRepo) TotalExpenses(ctx context.Context, q db.Querier, tripID uuid.UUID) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := q.QueryRow(ctx,
		`SELECT COALESCE(sum(amount), 0) FROM trip_expenses WHERE trip_id = $1`, tripID).Scan(&total)
	if err != nil {
		return decimal.Zero, wrapPgError(err)
	}
	return total, nil
}
