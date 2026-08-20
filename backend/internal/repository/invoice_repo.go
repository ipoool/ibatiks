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

const invoiceColumns = `id, invoice_number, order_id, type, issue_date, due_date,
	                    subtotal, discount, shipping_fee, total, dp_amount, amount_paid, amount_due,
	                    status, pdf_path, sent_channel, sent_at, paid_at, notes,
	                    created_by, created_at, updated_at`

const invoiceColumnsPrefixed = `i.id, i.invoice_number, i.order_id, i.type, i.issue_date, i.due_date,
	                            i.subtotal, i.discount, i.shipping_fee, i.total, i.dp_amount, i.amount_paid,
	                            i.amount_due, i.status, i.pdf_path, i.sent_channel, i.sent_at,
	                            i.paid_at, i.notes, i.created_by, i.created_at, i.updated_at`

type InvoiceRepo struct{}

func NewInvoiceRepo() *InvoiceRepo { return &InvoiceRepo{} }

type InvoiceParams struct {
	InvoiceNumber string
	OrderID       uuid.UUID
	Type          string
	IssueDate     time.Time
	DueDate       *time.Time
	Subtotal      decimal.Decimal
	Discount      decimal.Decimal
	ShippingFee   decimal.Decimal
	Total         decimal.Decimal
	DPAmount      decimal.Decimal
	AmountPaid    decimal.Decimal
	AmountDue     decimal.Decimal
	Notes         *string
	CreatedBy     *uuid.UUID
}

func (r *InvoiceRepo) Create(ctx context.Context, q db.Querier, p InvoiceParams) (*domain.Invoice, error) {
	return collectOne[domain.Invoice](ctx, q, "invoice", `
		INSERT INTO invoices (invoice_number, order_id, type, issue_date, due_date, subtotal,
		                      discount, shipping_fee, total, dp_amount, amount_paid, amount_due,
		                      notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING `+invoiceColumns,
		p.InvoiceNumber, p.OrderID, p.Type, p.IssueDate, p.DueDate, p.Subtotal,
		p.Discount, p.ShippingFee, p.Total, p.DPAmount, p.AmountPaid, p.AmountDue,
		p.Notes, p.CreatedBy)
}

func (r *InvoiceRepo) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Invoice, error) {
	return collectOne[domain.Invoice](ctx, q, "invoice",
		`SELECT `+invoiceColumns+` FROM invoices WHERE id = $1`, id)
}

func (r *InvoiceRepo) ListByOrder(ctx context.Context, q db.Querier, orderID uuid.UUID) ([]domain.Invoice, error) {
	return collectRows[domain.Invoice](ctx, q,
		`SELECT `+invoiceColumns+` FROM invoices WHERE order_id = $1 ORDER BY created_at DESC`, orderID)
}

func (r *InvoiceRepo) List(ctx context.Context, q db.Querier, p pagination.Params, status, invoiceType string, from, to *time.Time) ([]domain.InvoiceListItem, int64, error) {
	conditions := []string{}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf(
			"(i.invoice_number ILIKE $%d OR o.order_number ILIKE $%d OR c.name ILIKE $%d)", n, n, n))
	}
	if status != "" {
		args = append(args, status)
		conditions = append(conditions, fmt.Sprintf("i.status = $%d", len(args)))
	}
	if invoiceType != "" {
		args = append(args, invoiceType)
		conditions = append(conditions, fmt.Sprintf("i.type = $%d", len(args)))
	}
	// Disaring menurut tanggal terbit — itu tanggal milik invoice sendiri, dan
	// itu pula yang tampil di kolom Terbit.
	if from != nil {
		args = append(args, *from)
		conditions = append(conditions, fmt.Sprintf("i.issue_date >= $%d", len(args)))
	}
	if to != nil {
		args = append(args, *to)
		conditions = append(conditions, fmt.Sprintf("i.issue_date <= $%d", len(args)))
	}
	where := buildWhere(conditions)

	var total int64
	countQuery := `
		SELECT count(*) FROM invoices i
		JOIN orders o    ON o.id = i.order_id
		JOIN customers c ON c.id = o.customer_id` + where
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"issue_date":     "i.issue_date",
		"due_date":       "i.due_date",
		"total":          "i.total",
		"invoice_number": "i.invoice_number",
	}, "i.created_at")

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT %s,
			o.order_number AS order_number,
			c.name         AS customer_name,
			c.phone_wa     AS customer_phone,
			t.code         AS trip_code
		FROM invoices i
		JOIN orders o    ON o.id = i.order_id
		JOIN customers c ON c.id = o.customer_id
		JOIN trips t     ON t.id = o.trip_id%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		invoiceColumnsPrefixed, where, sortCol, p.Order, len(args)-1, len(args))

	invoices, err := collectRows[domain.InvoiceListItem](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return invoices, total, nil
}

func (r *InvoiceRepo) MarkSent(ctx context.Context, q db.Querier, id uuid.UUID, channel string) (*domain.Invoice, error) {
	return collectOne[domain.Invoice](ctx, q, "invoice", `
		UPDATE invoices
		SET status = CASE WHEN status = 'draft' THEN 'sent' ELSE status END,
		    sent_channel = $2,
		    sent_at = now()
		WHERE id = $1 AND status <> 'void'
		RETURNING `+invoiceColumns, id, channel)
}

func (r *InvoiceRepo) MarkPaid(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Invoice, error) {
	return collectOne[domain.Invoice](ctx, q, "invoice", `
		UPDATE invoices
		SET status = 'paid',
		    amount_paid = CASE WHEN type = 'dp' THEN dp_amount ELSE total END,
		    amount_due = 0,
		    paid_at = now()
		WHERE id = $1 AND status <> 'void'
		RETURNING `+invoiceColumns, id)
}

func (r *InvoiceRepo) Void(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Invoice, error) {
	return collectOne[domain.Invoice](ctx, q, "invoice", `
		UPDATE invoices SET status = 'void' WHERE id = $1 AND status <> 'paid'
		RETURNING `+invoiceColumns, id)
}

func (r *InvoiceRepo) SetPDFPath(ctx context.Context, q db.Querier, id uuid.UUID, path string) error {
	return execExpectOne(ctx, q, "invoice",
		`UPDATE invoices SET pdf_path = $2 WHERE id = $1`, id, path)
}

// SyncAmountsFromOrder menyelaraskan nominal terbayar pada invoice yang belum
// lunas dengan kondisi order terkini, dipanggil setelah pembayaran dicatat.
func (r *InvoiceRepo) SyncAmountsFromOrder(ctx context.Context, q db.Querier, orderID uuid.UUID) error {
	_, err := exec(ctx, q, `
		UPDATE invoices i
		SET amount_paid = LEAST(o.paid_amount, CASE WHEN i.type = 'dp' THEN i.dp_amount ELSE i.total END),
		    amount_due  = GREATEST((CASE WHEN i.type = 'dp' THEN i.dp_amount ELSE i.total END) - o.paid_amount, 0),
		    status      = CASE WHEN o.paid_amount >= (CASE WHEN i.type = 'dp' THEN i.dp_amount ELSE i.total END)
		                       THEN 'paid' ELSE i.status END,
		    paid_at     = CASE WHEN o.paid_amount >= (CASE WHEN i.type = 'dp' THEN i.dp_amount ELSE i.total END)
		                            AND i.paid_at IS NULL THEN now() ELSE i.paid_at END
		FROM orders o
		WHERE i.order_id = o.id AND o.id = $1 AND i.status NOT IN ('void', 'paid')`, orderID)
	return err
}

// --- Pengiriman ------------------------------------------------------------

const shipmentColumns = `id, order_id, courier, service, tracking_number, weight_gram,
	                     length_cm, width_cm, height_cm, estimated_cost, insurance_fee, shipping_cost,
	                     status, packed_at, packed_by, shipped_at, delivered_at,
	                     customer_notified_at, notes, created_at, updated_at`

type ShipmentRepo struct{}

func NewShipmentRepo() *ShipmentRepo { return &ShipmentRepo{} }

type PackParams struct {
	OrderID       uuid.UUID
	Courier       string
	Service       string
	WeightGram    int
	LengthCM      int
	WidthCM       int
	HeightCM      int
	EstimatedCost decimal.Decimal
	InsuranceFee  decimal.Decimal
	Notes         *string
	PackedBy      *uuid.UUID
}

// Pack membuat atau memperbarui data paket saat order dikemas. Memakai upsert
// supaya admin bisa memperbaiki data packing tanpa harus menghapus dulu.
func (r *ShipmentRepo) Pack(ctx context.Context, q db.Querier, p PackParams) (*domain.Shipment, error) {
	return collectOne[domain.Shipment](ctx, q, "pengiriman", `
		INSERT INTO shipments (order_id, courier, service, weight_gram, length_cm, width_cm,
		                       height_cm, estimated_cost, insurance_fee, notes, status, packed_at, packed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'ready', now(), $11)
		ON CONFLICT (order_id) DO UPDATE
		SET courier        = EXCLUDED.courier,
		    service        = EXCLUDED.service,
		    weight_gram    = EXCLUDED.weight_gram,
		    length_cm      = EXCLUDED.length_cm,
		    width_cm       = EXCLUDED.width_cm,
		    height_cm      = EXCLUDED.height_cm,
		    estimated_cost = EXCLUDED.estimated_cost,
		    insurance_fee  = EXCLUDED.insurance_fee,
		    notes          = EXCLUDED.notes,
		    status         = CASE WHEN shipments.status = 'packing' THEN 'ready' ELSE shipments.status END,
		    packed_at      = COALESCE(shipments.packed_at, now()),
		    packed_by      = COALESCE(shipments.packed_by, EXCLUDED.packed_by)
		RETURNING `+shipmentColumns,
		p.OrderID, p.Courier, p.Service, p.WeightGram, p.LengthCM, p.WidthCM,
		p.HeightCM, p.EstimatedCost, p.InsuranceFee, p.Notes, p.PackedBy)
}

func (r *ShipmentRepo) GetByOrder(ctx context.Context, q db.Querier, orderID uuid.UUID) (*domain.Shipment, error) {
	return collectOne[domain.Shipment](ctx, q, "pengiriman",
		`SELECT `+shipmentColumns+` FROM shipments WHERE order_id = $1`, orderID)
}

func (r *ShipmentRepo) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Shipment, error) {
	return collectOne[domain.Shipment](ctx, q, "pengiriman",
		`SELECT `+shipmentColumns+` FROM shipments WHERE id = $1`, id)
}

// Ship mencatat nomor resi dan menandai paket sudah diserahkan ke kurir.
func (r *ShipmentRepo) Ship(ctx context.Context, q db.Querier, id uuid.UUID, trackingNumber string, shippingCost decimal.Decimal, shippedAt time.Time) (*domain.Shipment, error) {
	return collectOne[domain.Shipment](ctx, q, "pengiriman", `
		UPDATE shipments
		SET tracking_number = $2, shipping_cost = $3, shipped_at = $4, status = 'shipped'
		WHERE id = $1
		RETURNING `+shipmentColumns, id, trackingNumber, shippingCost, shippedAt)
}

func (r *ShipmentRepo) UpdateStatus(ctx context.Context, q db.Querier, id uuid.UUID, status string) (*domain.Shipment, error) {
	return collectOne[domain.Shipment](ctx, q, "pengiriman", `
		UPDATE shipments
		SET status = $2,
		    delivered_at = CASE WHEN $2 = 'delivered' THEN COALESCE(delivered_at, now()) ELSE delivered_at END
		WHERE id = $1
		RETURNING `+shipmentColumns, id, status)
}

func (r *ShipmentRepo) MarkNotified(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Shipment, error) {
	return collectOne[domain.Shipment](ctx, q, "pengiriman",
		`UPDATE shipments SET customer_notified_at = now() WHERE id = $1 RETURNING `+shipmentColumns, id)
}

func (r *ShipmentRepo) Update(ctx context.Context, q db.Querier, id uuid.UUID, courier, service string, weightGram int, shippingCost decimal.Decimal, trackingNumber, notes *string) (*domain.Shipment, error) {
	return collectOne[domain.Shipment](ctx, q, "pengiriman", `
		UPDATE shipments
		SET courier = $2, service = $3, weight_gram = $4, shipping_cost = $5,
		    tracking_number = $6, notes = $7
		WHERE id = $1
		RETURNING `+shipmentColumns,
		id, courier, service, weightGram, shippingCost, trackingNumber, notes)
}

// ListCandidates mendaftar order yang siap ditagih pelunasannya: DP-nya sudah
// masuk, ongkirnya sudah ditetapkan, masih ada sisa tagihan, dan belum punya
// invoice pelunasan yang berlaku.
//
// Invoice yang dibatalkan sengaja tidak menghalangi: itulah cara admin
// menerbitkan pengganti setelah salah terbit.
func (r *InvoiceRepo) ListCandidates(ctx context.Context, q db.Querier, search string) ([]domain.InvoiceCandidate, error) {
	args := []any{}
	extra := ""
	if search != "" {
		args = append(args, "%"+search+"%")
		extra = fmt.Sprintf(" AND (o.order_number ILIKE $%d OR c.name ILIKE $%d)", len(args), len(args))
	}

	return collectRows[domain.InvoiceCandidate](ctx, q, `
		SELECT
			o.id            AS order_id,
			o.order_number  AS order_number,
			o.order_date    AS order_date,
			c.name          AS customer_name,
			t.code          AS trip_code,
			o.total         AS total,
			o.shipping_fee  AS shipping_fee,
			o.paid_amount   AS paid_amount,
			o.balance_due   AS balance_due
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN trips t     ON t.id = o.trip_id
		WHERE o.status = 'dp_paid'
		  AND o.shipping_fee > 0
		  AND o.balance_due > 0
		  AND NOT EXISTS (
			SELECT 1 FROM invoices i
			WHERE i.order_id = o.id AND i.type = 'final' AND i.status <> 'void'
		  )`+extra+`
		ORDER BY o.order_date ASC, o.order_number ASC
		LIMIT 100`, args...)
}

// TotalShippingCostByTrip dipakai laporan profit untuk memisahkan ongkir yang
// benar-benar dibayar ke kurir dari ongkir yang ditagihkan ke customer.
func (r *ShipmentRepo) TotalShippingCostByTrip(ctx context.Context, q db.Querier, tripID uuid.UUID) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := q.QueryRow(ctx, `
		SELECT COALESCE(sum(s.shipping_cost), 0)
		FROM shipments s
		JOIN orders o ON o.id = s.order_id
		WHERE o.trip_id = $1 AND o.status <> 'cancelled'`, tripID).Scan(&total)
	if err != nil {
		return decimal.Zero, wrapPgError(err)
	}
	return total, nil
}
