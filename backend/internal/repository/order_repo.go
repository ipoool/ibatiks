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

const orderColumns = `id, order_number, trip_id, customer_id, order_date, status, order_source,
	                  subtotal, discount, shipping_fee, total, dp_required, paid_amount, balance_due,
	                  recipient_name, recipient_phone, shipping_address, shipping_city,
	                  shipping_district, shipping_subdistrict,
	                  shipping_province, shipping_postal_code,
	                  notes, cancel_reason, cancelled_at, created_by, created_at, updated_at`

const orderColumnsPrefixed = `o.id, o.order_number, o.trip_id, o.customer_id, o.order_date, o.status,
	                          o.order_source,
	                          o.subtotal, o.discount, o.shipping_fee, o.total, o.dp_required,
	                          o.paid_amount, o.balance_due,
	                          o.recipient_name, o.recipient_phone, o.shipping_address, o.shipping_city,
	                          o.shipping_district, o.shipping_subdistrict,
	                          o.shipping_province, o.shipping_postal_code,
	                          o.notes, o.cancel_reason, o.cancelled_at, o.created_by,
	                          o.created_at, o.updated_at`

type OrderRepo struct{}

func NewOrderRepo() *OrderRepo { return &OrderRepo{} }

type CreateOrderParams struct {
	OrderNumber         string
	TripID              uuid.UUID
	CustomerID          uuid.UUID
	OrderDate           time.Time
	OrderSource         string
	Status              string
	Discount            decimal.Decimal
	ShippingFee         decimal.Decimal
	DPRequired          decimal.Decimal
	RecipientName       string
	RecipientPhone      string
	ShippingAddress     string
	ShippingCity        string
	ShippingDistrict    *string
	ShippingSubdistrict *string
	ShippingProvince    *string
	ShippingPostalCode  *string
	Notes               *string
	CreatedBy           *uuid.UUID
}

func (r *OrderRepo) Create(ctx context.Context, q db.Querier, p CreateOrderParams) (*domain.Order, error) {
	return collectOne[domain.Order](ctx, q, "order", `
		INSERT INTO orders (order_number, trip_id, customer_id, order_date, order_source, status,
		                    discount, shipping_fee, dp_required, recipient_name, recipient_phone,
		                    shipping_address, shipping_city, shipping_district, shipping_subdistrict,
		                    shipping_province, shipping_postal_code, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING `+orderColumns,
		p.OrderNumber, p.TripID, p.CustomerID, p.OrderDate, p.OrderSource, p.Status,
		p.Discount, p.ShippingFee, p.DPRequired, p.RecipientName, p.RecipientPhone,
		p.ShippingAddress, p.ShippingCity, p.ShippingDistrict, p.ShippingSubdistrict,
		p.ShippingProvince, p.ShippingPostalCode, p.Notes, p.CreatedBy)
}

func (r *OrderRepo) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Order, error) {
	return collectOne[domain.Order](ctx, q, "order",
		`SELECT `+orderColumns+` FROM orders WHERE id = $1`, id)
}

// GetForUpdate mengunci baris order sampai transaksi selesai. Dipakai sebelum
// mencatat pembayaran atau mengubah item, supaya dua admin yang menyimpan
// bersamaan tidak menimpa hasil hitungan satu sama lain.
func (r *OrderRepo) GetForUpdate(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Order, error) {
	return collectOne[domain.Order](ctx, q, "order",
		`SELECT `+orderColumns+` FROM orders WHERE id = $1 FOR UPDATE`, id)
}

type OrderFilter struct {
	TripID     *uuid.UUID
	CustomerID *uuid.UUID
	Status     string
	// Source menyaring berdasarkan asal order (whatsapp, instagram, dan lainnya).
	Source string
	// UnpaidOnly menyaring order yang masih punya sisa tagihan.
	UnpaidOnly bool
	// ReadyToShip menyaring order lunas yang belum diserahkan ke kurir.
	ReadyToShip bool
}

func (r *OrderRepo) List(ctx context.Context, q db.Querier, p pagination.Params, f OrderFilter) ([]domain.OrderListItem, int64, error) {
	conditions := []string{}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf(
			"(o.order_number ILIKE $%d OR c.name ILIKE $%d OR c.phone_wa ILIKE $%d OR o.recipient_name ILIKE $%d)",
			n, n, n, n))
	}
	if f.TripID != nil {
		args = append(args, *f.TripID)
		conditions = append(conditions, fmt.Sprintf("o.trip_id = $%d", len(args)))
	}
	if f.CustomerID != nil {
		args = append(args, *f.CustomerID)
		conditions = append(conditions, fmt.Sprintf("o.customer_id = $%d", len(args)))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		conditions = append(conditions, fmt.Sprintf("o.status = $%d", len(args)))
	}
	if f.Source != "" {
		args = append(args, f.Source)
		conditions = append(conditions, fmt.Sprintf("o.order_source = $%d", len(args)))
	}
	if f.UnpaidOnly {
		conditions = append(conditions, "o.balance_due > 0 AND o.status <> 'cancelled'")
	}
	if f.ReadyToShip {
		conditions = append(conditions, "o.status = 'paid'")
	}
	where := buildWhere(conditions)

	countQuery := `SELECT count(*) FROM orders o JOIN customers c ON c.id = o.customer_id` + where
	var total int64
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"order_date":   "o.order_date",
		"order_number": "o.order_number",
		"total":        "o.total",
		"balance_due":  "o.balance_due",
		"status":       "o.status",
		"created_at":   "o.created_at",
	}, "o.created_at")

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT %s,
			c.name          AS customer_name,
			c.code          AS customer_code,
			t.code          AS trip_code,
			t.title         AS trip_title,
			t.currency      AS trip_currency,
			t.exchange_rate AS trip_exchange_rate,
			(SELECT count(*)               FROM order_items oi WHERE oi.order_id = o.id)::int AS item_count,
			COALESCE((SELECT sum(oi.qty)   FROM order_items oi WHERE oi.order_id = o.id), 0)::int AS total_qty
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN trips t     ON t.id = o.trip_id%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		orderColumnsPrefixed, where, sortCol, p.Order, len(args)-1, len(args))

	orders, err := collectRows[domain.OrderListItem](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

type UpdateOrderParams struct {
	OrderDate           time.Time
	OrderSource         string
	Discount            decimal.Decimal
	ShippingFee         decimal.Decimal
	DPRequired          decimal.Decimal
	RecipientName       string
	RecipientPhone      string
	ShippingAddress     string
	ShippingCity        string
	ShippingDistrict    *string
	ShippingSubdistrict *string
	ShippingProvince    *string
	ShippingPostalCode  *string
	Notes               *string
}

func (r *OrderRepo) Update(ctx context.Context, q db.Querier, id uuid.UUID, p UpdateOrderParams) (*domain.Order, error) {
	return collectOne[domain.Order](ctx, q, "order", `
		UPDATE orders
		SET order_date = $2, order_source = $3, discount = $4, shipping_fee = $5, dp_required = $6,
		    recipient_name = $7, recipient_phone = $8, shipping_address = $9, shipping_city = $10,
		    shipping_district = $11, shipping_subdistrict = $12,
		    shipping_province = $13, shipping_postal_code = $14, notes = $15
		WHERE id = $1
		RETURNING `+orderColumns,
		id, p.OrderDate, p.OrderSource, p.Discount, p.ShippingFee, p.DPRequired,
		p.RecipientName, p.RecipientPhone, p.ShippingAddress, p.ShippingCity,
		p.ShippingDistrict, p.ShippingSubdistrict,
		p.ShippingProvince, p.ShippingPostalCode, p.Notes)
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, q db.Querier, id uuid.UUID, status string) (*domain.Order, error) {
	return collectOne[domain.Order](ctx, q, "order",
		`UPDATE orders SET status = $2 WHERE id = $1 RETURNING `+orderColumns, id, status)
}

func (r *OrderRepo) Cancel(ctx context.Context, q db.Querier, id uuid.UUID, reason *string) (*domain.Order, error) {
	return collectOne[domain.Order](ctx, q, "order", `
		UPDATE orders SET status = 'cancelled', cancel_reason = $2, cancelled_at = now()
		WHERE id = $1
		RETURNING `+orderColumns, id, reason)
}

func (r *OrderRepo) Delete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "order", `DELETE FROM orders WHERE id = $1`, id)
}

// ListForShipping mendaftar order yang sudah membayar DP beserta data
// kemasannya, untuk menu Pengiriman.
//
// LEFT JOIN, bukan JOIN: paket baru terbentuk setelah admin mengisi data
// kemasan, sementara justru order yang belum dikemas itulah yang paling perlu
// terlihat. Menyaring dengan JOIN biasa akan menyembunyikan seluruh pekerjaan
// yang belum dikerjakan.
func (r *OrderRepo) ListForShipping(
	ctx context.Context, q db.Querier, p pagination.Params, stage string, tripID *uuid.UUID,
) ([]domain.ShippingQueueItem, int64, error) {
	// Order yang masih menunggu DP belum jadi pekerjaan gudang: barangnya
	// belum tentu dibeli, dan mengemasnya berarti menalangi dengan uang toko.
	conditions := []string{"o.status IN ('dp_paid', 'paid', 'shipped', 'completed')"}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf(
			"(o.order_number ILIKE $%d OR c.name ILIKE $%d OR o.recipient_name ILIKE $%d"+
				" OR COALESCE(s.tracking_number, '') ILIKE $%d)", n, n, n, n))
	}

	switch stage {
	case domain.ShippingStagePacking:
		// Belum ditimbang atau ongkirnya belum ditetapkan — keduanya sama-sama
		// menghalangi invoice pelunasan terbit.
		conditions = append(conditions,
			"(s.id IS NULL OR s.weight_gram = 0 OR o.shipping_fee = 0)")
	case domain.ShippingStageReady:
		conditions = append(conditions, "o.status = 'paid'", "s.tracking_number IS NULL")
	case domain.ShippingStageSent:
		conditions = append(conditions, "s.tracking_number IS NOT NULL")
	}

	if tripID != nil {
		args = append(args, *tripID)
		conditions = append(conditions, fmt.Sprintf("o.trip_id = $%d", len(args)))
	}
	where := buildWhere(conditions)

	from := `
		FROM orders o
		JOIN customers c      ON c.id = o.customer_id
		JOIN trips t          ON t.id = o.trip_id
		LEFT JOIN shipments s ON s.order_id = o.id`

	var total int64
	if err := q.QueryRow(ctx, `SELECT count(*)`+from+where, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT
			o.id              AS order_id,
			o.order_number    AS order_number,
			o.status          AS order_status,
			o.order_date      AS order_date,
			t.code            AS trip_code,
			c.name            AS customer_name,
			o.recipient_name  AS recipient_name,
			o.recipient_phone AS recipient_phone,
			o.shipping_city   AS shipping_city,
			o.total           AS total,
			o.balance_due     AS balance_due,
			o.shipping_fee    AS shipping_fee,
			COALESCE((SELECT sum(oi.qty) FROM order_items oi WHERE oi.order_id = o.id), 0)::int AS total_qty,
			s.id                   AS shipment_id,
			s.courier              AS courier,
			s.service              AS service,
			s.weight_gram          AS weight_gram,
			s.length_cm            AS length_cm,
			s.width_cm             AS width_cm,
			s.height_cm            AS height_cm,
			s.tracking_number      AS tracking_number,
			s.status               AS shipment_status,
			s.notes                AS shipment_notes,
			s.packed_at            AS packed_at,
			s.shipped_at           AS shipped_at,
			s.shipping_cost        AS shipping_cost,
			s.customer_notified_at AS customer_notified_at%s%s
		ORDER BY o.order_date DESC, o.order_number DESC
		LIMIT $%d OFFSET $%d`, from, where, len(args)-1, len(args))

	items, err := collectRows[domain.ShippingQueueItem](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// DeleteByTrip menghapus seluruh order milik sebuah trip.
//
// Item, pembayaran, invoice, dan pengiriman ikut terhapus lewat cascade dari
// order. Kolom orders.trip_id sengaja tetap ON DELETE RESTRICT supaya trip
// tidak bisa dihapus diam-diam dari jalur lain — hanya pemanggil yang memang
// sudah memeriksa penghalangnya yang boleh memakai fungsi ini.
func (r *OrderRepo) DeleteByTrip(ctx context.Context, q db.Querier, tripID uuid.UUID) error {
	_, err := q.Exec(ctx, `DELETE FROM orders WHERE trip_id = $1`, tripID)
	if err != nil {
		return wrapPgError(err)
	}
	return nil
}

// SetShippingFee menuliskan ongkir yang ditagihkan ke customer.
//
// Sengaja tidak lewat Update: ongkir ditetapkan belakangan, saat paket dikemas
// dan layanan kurirnya dipilih, sementara Update menulis ulang seluruh kolom
// order termasuk DP dan alamat. Satu kolom yang berubah tidak boleh membawa
// serta belasan kolom lain yang kebetulan ikut terkirim.
func (r *OrderRepo) SetShippingFee(ctx context.Context, q db.Querier, id uuid.UUID, fee decimal.Decimal) (*domain.Order, error) {
	return collectOne[domain.Order](ctx, q, "order",
		`UPDATE orders SET shipping_fee = $2 WHERE id = $1 RETURNING `+orderColumns, id, fee)
}

// RecalculateTotals menghitung ulang subtotal dan total dari baris pesanan.
// Selalu dipanggil setelah item ditambah, diubah, atau dihapus, dan setelah
// barang datang dicocokkan, di dalam transaksi yang sama, sehingga angka order
// tidak pernah tertinggal dari isinya.
//
// Barang yang tidak berhasil dibeli tidak ikut ditagihkan. Begitu sebuah item
// ditandai tidak tersedia, kurang, atau direfund, yang dihitung adalah jumlah
// yang benar-benar diterima customer — bukan jumlah yang dulu dipesan. Kolom
// qty sengaja dibiarkan apa adanya supaya tetap terbaca apa yang dipesan
// semula; yang berubah hanya apa yang ditagih.
func (r *OrderRepo) RecalculateTotals(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Order, error) {
	// Kolom subquery sengaja dinamai items_subtotal, bukan subtotal: nama yang
	// sama dengan kolom orders akan membuat klausa RETURNING ambigu.
	return collectOne[domain.Order](ctx, q, "order", `
		UPDATE orders o
		SET subtotal = s.items_subtotal,
		    -- GREATEST menjaga total tidak negatif kalau diskon melebihi belanja;
		    -- validasi diskon yang wajar dilakukan di layer service.
		    total    = GREATEST(s.items_subtotal - o.discount + o.shipping_fee, 0)
		FROM (
			SELECT COALESCE(sum(
				CASE WHEN fulfillment_status IN ('unavailable', 'partial', 'refunded')
				     THEN qty_received * unit_price
				     ELSE subtotal
				END
			), 0) AS items_subtotal
			FROM order_items
			WHERE order_id = $1
		) s
		WHERE o.id = $1
		RETURNING `+orderColumnsPrefixed, id)
}

// RecalculatePaidAmount menjumlahkan seluruh pembayaran order. Refund dihitung
// negatif sehingga sisa tagihan otomatis bertambah kembali.
func (r *OrderRepo) RecalculatePaidAmount(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Order, error) {
	return collectOne[domain.Order](ctx, q, "order", `
		UPDATE orders o
		SET paid_amount = p.payments_total
		FROM (
			SELECT COALESCE(sum(CASE WHEN type = 'refund' THEN -amount ELSE amount END), 0) AS payments_total
			FROM payments
			WHERE order_id = $1
		) p
		WHERE o.id = $1
		RETURNING `+orderColumnsPrefixed, id)
}

// --- Baris pesanan ---------------------------------------------------------

const orderItemColumns = `id, order_id, product_id, trip_item_id, product_name, product_sku,
	                      qty, unit_price, unit_cost_est, subtotal, qty_purchased, qty_received,
	                      fulfillment_status, notes, created_at, updated_at`

type OrderItemParams struct {
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	TripItemID  *uuid.UUID
	ProductName string
	ProductSKU  string
	Qty         int
	UnitPrice   decimal.Decimal
	UnitCostEst decimal.Decimal
	Notes       *string
}

func (r *OrderRepo) AddItem(ctx context.Context, q db.Querier, p OrderItemParams) (*domain.OrderItem, error) {
	return collectOne[domain.OrderItem](ctx, q, "item order", `
		INSERT INTO order_items (order_id, product_id, trip_item_id, product_name, product_sku,
		                         qty, unit_price, unit_cost_est, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+orderItemColumns,
		p.OrderID, p.ProductID, p.TripItemID, p.ProductName, p.ProductSKU,
		p.Qty, p.UnitPrice, p.UnitCostEst, p.Notes)
}

func (r *OrderRepo) GetItem(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.OrderItem, error) {
	return collectOne[domain.OrderItem](ctx, q, "item order",
		`SELECT `+orderItemColumns+` FROM order_items WHERE id = $1`, id)
}

func (r *OrderRepo) ListItems(ctx context.Context, q db.Querier, orderID uuid.UUID) ([]domain.OrderItem, error) {
	return collectRows[domain.OrderItem](ctx, q,
		`SELECT `+orderItemColumns+` FROM order_items WHERE order_id = $1 ORDER BY created_at ASC`,
		orderID)
}

// FindItemByProduct dipakai agar menambahkan produk yang sudah ada di order
// menambah qty, bukan membuat baris kembar.
func (r *OrderRepo) FindItemByProduct(ctx context.Context, q db.Querier, orderID, productID uuid.UUID) (*domain.OrderItem, error) {
	return collectOne[domain.OrderItem](ctx, q, "item order",
		`SELECT `+orderItemColumns+` FROM order_items WHERE order_id = $1 AND product_id = $2`,
		orderID, productID)
}

func (r *OrderRepo) UpdateItem(ctx context.Context, q db.Querier, id uuid.UUID, qty int, unitPrice decimal.Decimal, notes *string) (*domain.OrderItem, error) {
	return collectOne[domain.OrderItem](ctx, q, "item order", `
		UPDATE order_items SET qty = $2, unit_price = $3, notes = $4
		WHERE id = $1
		RETURNING `+orderItemColumns, id, qty, unitPrice, notes)
}

func (r *OrderRepo) UpdateItemFulfillment(ctx context.Context, q db.Querier, id uuid.UUID, status string, qtyReceived int) (*domain.OrderItem, error) {
	return collectOne[domain.OrderItem](ctx, q, "item order", `
		UPDATE order_items SET fulfillment_status = $2, qty_received = $3
		WHERE id = $1
		RETURNING `+orderItemColumns, id, status, qtyReceived)
}

func (r *OrderRepo) DeleteItem(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "item order", `DELETE FROM order_items WHERE id = $1`, id)
}

// SyncItemPurchasedQty menyelaraskan kolom qty_purchased dengan alokasi
// pembelian yang benar-benar tercatat, lalu memperbarui status pemenuhannya.
func (r *OrderRepo) SyncItemPurchasedQty(ctx context.Context, q db.Querier, orderItemID uuid.UUID) error {
	_, err := exec(ctx, q, `
		UPDATE order_items oi
		SET qty_purchased = a.qty,
		    fulfillment_status = CASE
		        WHEN oi.fulfillment_status IN ('unavailable', 'refunded') THEN oi.fulfillment_status
		        WHEN a.qty = 0        THEN 'pending'
		        WHEN a.qty >= oi.qty  THEN 'purchased'
		        ELSE 'partial'
		    END
		FROM (
			SELECT COALESCE(sum(qty), 0)::int AS qty
			FROM purchase_allocations
			WHERE order_item_id = $1
		) a
		WHERE oi.id = $1`, orderItemID)
	return err
}

// --- Pembayaran ------------------------------------------------------------

const paymentColumns = `id, order_id, type, amount, method, reference, proof_url, paid_at,
	                    notes, recorded_by, created_at`

type PaymentParams struct {
	OrderID    uuid.UUID
	Type       string
	Amount     decimal.Decimal
	Method     string
	Reference  *string
	ProofURL   *string
	PaidAt     time.Time
	Notes      *string
	RecordedBy *uuid.UUID
}

func (r *OrderRepo) CreatePayment(ctx context.Context, q db.Querier, p PaymentParams) (*domain.Payment, error) {
	return collectOne[domain.Payment](ctx, q, "pembayaran", `
		INSERT INTO payments (order_id, type, amount, method, reference, proof_url, paid_at, notes, recorded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+paymentColumns,
		p.OrderID, p.Type, p.Amount, p.Method, p.Reference, p.ProofURL, p.PaidAt, p.Notes, p.RecordedBy)
}

func (r *OrderRepo) ListPayments(ctx context.Context, q db.Querier, orderID uuid.UUID) ([]domain.Payment, error) {
	return collectRows[domain.Payment](ctx, q,
		`SELECT `+paymentColumns+` FROM payments WHERE order_id = $1 ORDER BY paid_at ASC`, orderID)
}

func (r *OrderRepo) GetPayment(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Payment, error) {
	return collectOne[domain.Payment](ctx, q, "pembayaran",
		`SELECT `+paymentColumns+` FROM payments WHERE id = $1`, id)
}

func (r *OrderRepo) DeletePayment(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "pembayaran", `DELETE FROM payments WHERE id = $1`, id)
}

// TotalDPPaid menghitung uang muka yang sudah masuk untuk satu order.
func (r *OrderRepo) TotalDPPaid(ctx context.Context, q db.Querier, orderID uuid.UUID) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := q.QueryRow(ctx,
		`SELECT COALESCE(sum(amount), 0) FROM payments WHERE order_id = $1 AND type = 'dp'`,
		orderID).Scan(&total)
	if err != nil {
		return decimal.Zero, wrapPgError(err)
	}
	return total, nil
}

// ListByTripAndStatuses dipakai saat status seluruh order dalam satu trip perlu
// digeser bersamaan, misalnya ketika tripper mulai belanja.
func (r *OrderRepo) ListByTripAndStatuses(ctx context.Context, q db.Querier, tripID uuid.UUID, statuses []string) ([]domain.Order, error) {
	return collectRows[domain.Order](ctx, q,
		`SELECT `+orderColumns+` FROM orders WHERE trip_id = $1 AND status = ANY($2) ORDER BY created_at ASC`,
		tripID, statuses)
}

// BulkUpdateStatus menggeser status banyak order sekaligus dan mengembalikan
// jumlah yang benar-benar berubah.
func (r *OrderRepo) BulkUpdateStatus(ctx context.Context, q db.Querier, tripID uuid.UUID, from []string, to string) (int64, error) {
	return exec(ctx, q,
		`UPDATE orders SET status = $3 WHERE trip_id = $1 AND status = ANY($2)`, tripID, from, to)
}
