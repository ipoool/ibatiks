package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/config"
	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/docnum"
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type OrderService struct {
	pool      *pgxpool.Pool
	orders    *repository.OrderRepo
	trips     *repository.TripRepo
	customers *repository.CustomerRepo
	products  *repository.ProductRepo
	purchases *repository.PurchaseRepo
	invoices  *repository.InvoiceRepo
	audit     *repository.AuditRepo
	business  config.Business
}

func NewOrderService(
	pool *pgxpool.Pool,
	orders *repository.OrderRepo,
	trips *repository.TripRepo,
	customers *repository.CustomerRepo,
	products *repository.ProductRepo,
	purchases *repository.PurchaseRepo,
	invoices *repository.InvoiceRepo,
	audit *repository.AuditRepo,
	business config.Business,
) *OrderService {
	return &OrderService{
		pool: pool, orders: orders, trips: trips, customers: customers,
		products: products, purchases: purchases, invoices: invoices,
		audit: audit, business: business,
	}
}

type OrderItemInput struct {
	ProductID uuid.UUID
	Qty       int
	// UnitPrice opsional: kalau nil, harga diambil dari katalog trip.
	UnitPrice *decimal.Decimal
	Notes     *string
}

type CreateOrderInput struct {
	TripID     uuid.UUID
	CustomerID uuid.UUID
	OrderDate  time.Time
	// OrderSource adalah asal pesanan: whatsapp, instagram, dan seterusnya.
	// Kosong berarti WhatsApp, kanal yang paling lazim.
	OrderSource string
	Items       []OrderItemInput
	Discount    decimal.Decimal
	ShippingFee decimal.Decimal
	// DPRequired opsional: kalau nil, dipakai persentase DP default.
	DPRequired *decimal.Decimal
	// Alamat pengiriman opsional: kalau kosong, disalin dari data customer.
	RecipientName       *string
	RecipientPhone      *string
	ShippingAddress     *string
	ShippingCity        *string
	ShippingDistrict    *string
	ShippingSubdistrict *string
	ShippingProvince    *string
	ShippingPostalCode  *string
	Notes               *string
}

// Create mencatat order baru beserta seluruh itemnya dalam satu transaksi.
// Harga tiap item diambil dari katalog trip dan disalin ke order_items, supaya
// perubahan harga katalog nanti tidak mengubah pesanan yang sudah masuk.
func (s *OrderService) Create(ctx context.Context, in CreateOrderInput, actorID uuid.UUID) (*domain.OrderDetail, error) {
	if len(in.Items) == 0 {
		return nil, domain.Validation("order harus punya minimal satu item", map[string]string{
			"items": "tambahkan minimal satu produk",
		})
	}

	source := in.OrderSource
	if source == "" {
		source = domain.SourceWhatsApp
	}
	if !domain.IsValidOrderSource(source) {
		return nil, domain.Validation("asal order tidak dikenal", map[string]string{
			"order_source": "pilih: whatsapp, instagram, tiktok, marketplace, lainnya",
		})
	}

	trip, err := s.trips.GetByID(ctx, s.pool, in.TripID)
	if err != nil {
		return nil, err
	}
	if !domain.TripAcceptsOrder(trip.Status) {
		return nil, domain.InvalidState(
			"trip %s berstatus %s sehingga belum/tidak lagi menerima order baru", trip.Code, trip.Status)
	}

	customer, err := s.customers.GetByID(ctx, s.pool, in.CustomerID)
	if err != nil {
		return nil, err
	}

	shipping := resolveShipping(in, customer)
	if err := validateShipping(shipping); err != nil {
		return nil, err
	}

	var orderID uuid.UUID
	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		orderNumber, err := docnum.Next(ctx, tx, docnum.Order, in.OrderDate.Year())
		if err != nil {
			return err
		}

		order, err := s.orders.Create(ctx, tx, repository.CreateOrderParams{
			OrderNumber: orderNumber,
			TripID:      in.TripID,
			CustomerID:  in.CustomerID,
			OrderDate:   in.OrderDate,
			OrderSource: source,
			// Order yang baru dicatat langsung menunggu DP. Admin mencatatnya
			// setelah customer benar-benar memesan, jadi tahap draft hanya
			// menambah satu klik tanpa memberi manfaat.
			Status:              domain.OrderAwaitingDP,
			Discount:            in.Discount,
			ShippingFee:         in.ShippingFee,
			DPRequired:          decimal.Zero, // diisi setelah total diketahui
			RecipientName:       shipping.Name,
			RecipientPhone:      shipping.Phone,
			ShippingAddress:     shipping.Address,
			ShippingCity:        shipping.City,
			ShippingDistrict:    shipping.District,
			ShippingSubdistrict: shipping.Subdistrict,
			ShippingProvince:    shipping.Province,
			ShippingPostalCode:  shipping.PostalCode,
			Notes:               trimPtr(in.Notes),
			CreatedBy:           nullableUUID(actorID),
		})
		if err != nil {
			return err
		}
		orderID = order.ID

		for _, item := range in.Items {
			if _, err := s.addItemTx(ctx, tx, order, trip, item); err != nil {
				return err
			}
		}

		order, err = s.orders.RecalculateTotals(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if err := s.validateDiscount(order); err != nil {
			return err
		}
		if err := s.applyDPRequired(ctx, tx, order, in.DPRequired); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &orderID,
			Action:   domain.AuditCreate,
			Changes:  map[string]any{"order_number": orderNumber, "item_count": len(in.Items)},
		})
	})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, orderID)
}

// Get menyusun tampilan lengkap sebuah order: item, pembayaran, invoice, dan
// data pengiriman sekaligus, supaya halaman detail cukup satu panggilan.
func (s *OrderService) Get(ctx context.Context, id uuid.UUID) (*domain.OrderDetail, error) {
	order, err := s.orders.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}

	items, err := s.orders.ListItems(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	payments, err := s.orders.ListPayments(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	invoices, err := s.invoices.ListByOrder(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	// Sengaja termasuk customer yang sudah dihapus. Menghapus customer tidak
	// menghapus ordernya, jadi kalau baris ini menolak yang sudah dihapus,
	// seluruh halaman order lamanya mati dengan pesan "customer tidak
	// ditemukan" — padahal ordernya masih ada dan masih perlu ditagih.
	customer, err := s.customers.GetByIDIncludingDeleted(ctx, s.pool, order.CustomerID)
	if err != nil {
		return nil, err
	}
	trip, err := s.trips.GetByID(ctx, s.pool, order.TripID)
	if err != nil {
		return nil, err
	}

	// Paket belum tentu ada: order yang belum dikemas tidak punya baris shipment.
	shipment, err := repository.NewShipmentRepo().GetByOrder(ctx, s.pool, id)
	if err != nil {
		if domainErr, ok := domain.AsError(err); !ok || domainErr.Code != domain.CodeNotFound {
			return nil, err
		}
		shipment = nil
	}

	return &domain.OrderDetail{
		Order:        *order,
		Customer:     customer,
		Trip:         trip,
		Items:        items,
		Payments:     payments,
		Invoices:     invoices,
		Shipment:     shipment,
		NextStatuses: domain.NextOrderStatuses(order.Status),
		Editable:     domain.OrderIsEditable(order.Status),
	}, nil
}

func (s *OrderService) List(ctx context.Context, p pagination.Params, f repository.OrderFilter) ([]domain.OrderListItem, int64, error) {
	return s.orders.List(ctx, s.pool, p, f)
}

type UpdateOrderInput struct {
	OrderDate           time.Time
	OrderSource         string
	Discount            decimal.Decimal
	ShippingFee         decimal.Decimal
	DPRequired          *decimal.Decimal
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

func (s *OrderService) Update(ctx context.Context, id uuid.UUID, in UpdateOrderInput, actorID uuid.UUID) (*domain.OrderDetail, error) {
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if !domain.OrderIsEditable(order.Status) {
			return domain.InvalidState("order berstatus %s tidak bisa diubah lagi", order.Status)
		}

		shipping := shippingInfo{
			Name:       strings.TrimSpace(in.RecipientName),
			Phone:      domain.NormalizePhoneWA(in.RecipientPhone),
			Address:    strings.TrimSpace(in.ShippingAddress),
			City:       strings.TrimSpace(in.ShippingCity),
			Province:   trimPtr(in.ShippingProvince),
			PostalCode: trimPtr(in.ShippingPostalCode),
		}
		if err := validateShipping(shipping); err != nil {
			return err
		}

		source := in.OrderSource
		if source == "" {
			source = order.OrderSource
		}
		if !domain.IsValidOrderSource(source) {
			return domain.Validation("asal order tidak dikenal", map[string]string{
				"order_source": "pilih: whatsapp, instagram, tiktok, marketplace, lainnya",
			})
		}

		updated, err := s.orders.Update(ctx, tx, id, repository.UpdateOrderParams{
			OrderDate:           in.OrderDate,
			OrderSource:         source,
			Discount:            in.Discount,
			ShippingFee:         in.ShippingFee,
			DPRequired:          order.DPRequired,
			RecipientName:       shipping.Name,
			RecipientPhone:      shipping.Phone,
			ShippingAddress:     shipping.Address,
			ShippingCity:        shipping.City,
			ShippingDistrict:    shipping.District,
			ShippingSubdistrict: shipping.Subdistrict,
			ShippingProvince:    shipping.Province,
			ShippingPostalCode:  shipping.PostalCode,
			Notes:               trimPtr(in.Notes),
		})
		if err != nil {
			return err
		}

		updated, err = s.orders.RecalculateTotals(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.validateDiscount(updated); err != nil {
			return err
		}
		if in.DPRequired != nil {
			if err := s.applyDPRequired(ctx, tx, updated, in.DPRequired); err != nil {
				return err
			}
		}
		if err := s.reconcileStatusAfterAmountChange(ctx, tx, updated); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &id,
			Action:   domain.AuditUpdate,
			Changes: map[string]any{
				"discount_from": order.Discount.String(), "discount_to": in.Discount.String(),
				"shipping_from": order.ShippingFee.String(), "shipping_to": in.ShippingFee.String(),
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// --- Item order ------------------------------------------------------------

// AddItem menambahkan produk ke order yang sudah ada. Kalau produknya sudah
// ada di order, qty-nya ditambahkan alih-alih membuat baris kembar.
func (s *OrderService) AddItem(ctx context.Context, orderID uuid.UUID, in OrderItemInput, actorID uuid.UUID) (*domain.OrderDetail, error) {
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if !domain.OrderIsEditable(order.Status) {
			return domain.InvalidState("order berstatus %s tidak bisa diubah lagi", order.Status)
		}

		trip, err := s.trips.GetByID(ctx, tx, order.TripID)
		if err != nil {
			return err
		}

		item, err := s.addItemTx(ctx, tx, order, trip, in)
		if err != nil {
			return err
		}

		updated, err := s.orders.RecalculateTotals(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if err := s.validateDiscount(updated); err != nil {
			return err
		}
		if err := s.reconcileStatusAfterAmountChange(ctx, tx, updated); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &orderID,
			Action:   domain.AuditItemChange,
			Changes: map[string]any{
				"action": "add", "product": item.ProductName, "qty": item.Qty,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, orderID)
}

// UpdateItem mengubah qty atau harga satu baris pesanan.
//
// Ini operasi yang paling sering dipakai admin ("customer minta tambah 1 lagi")
// sekaligus yang paling rawan: qty tidak boleh turun di bawah jumlah yang sudah
// diterima, dan kalau turun di bawah jumlah yang sudah dibeli, kelebihannya
// dilepas jadi stok, bukan hilang begitu saja.
func (s *OrderService) UpdateItem(ctx context.Context, orderID, itemID uuid.UUID, qty int, unitPrice *decimal.Decimal, notes *string, actorID uuid.UUID) (*domain.OrderDetail, error) {
	if qty < 1 {
		return nil, domain.Validation("jumlah tidak valid", map[string]string{
			"qty": "minimal 1, gunakan hapus item kalau ingin membatalkan",
		})
	}

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if !domain.OrderIsEditable(order.Status) {
			return domain.InvalidState("order berstatus %s tidak bisa diubah lagi", order.Status)
		}

		item, err := s.orders.GetItem(ctx, tx, itemID)
		if err != nil {
			return err
		}
		if item.OrderID != orderID {
			return domain.NotFound("item order")
		}
		if qty < item.QtyReceived {
			return domain.Conflict(
				"jumlah tidak bisa dikurangi menjadi %d karena %d unit sudah diterima",
				qty, item.QtyReceived)
		}

		price := item.UnitPrice
		if unitPrice != nil {
			if unitPrice.IsNegative() {
				return domain.Validation("harga tidak valid", map[string]string{
					"unit_price": "harus 0 atau lebih",
				})
			}
			price = *unitPrice
		}

		if _, err := s.orders.UpdateItem(ctx, tx, itemID, qty, price, trimPtr(notes)); err != nil {
			return err
		}

		// Barang yang sudah terlanjur dibeli untuk unit yang dibatalkan tetap
		// ada wujudnya, jadi dipindahkan ke stok, bukan dihapus catatannya.
		if err := s.releaseExcessAllocations(ctx, tx, item, qty, order.TripID, actorID); err != nil {
			return err
		}
		if err := s.orders.SyncItemPurchasedQty(ctx, tx, itemID); err != nil {
			return err
		}

		updated, err := s.orders.RecalculateTotals(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if err := s.validateDiscount(updated); err != nil {
			return err
		}
		if err := s.reconcileStatusAfterAmountChange(ctx, tx, updated); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &orderID,
			Action:   domain.AuditItemChange,
			Changes: map[string]any{
				"action": "update", "product": item.ProductName,
				"qty_from": item.Qty, "qty_to": qty,
				"price_from": item.UnitPrice.String(), "price_to": price.String(),
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, orderID)
}

func (s *OrderService) DeleteItem(ctx context.Context, orderID, itemID uuid.UUID, actorID uuid.UUID) (*domain.OrderDetail, error) {
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if !domain.OrderIsEditable(order.Status) {
			return domain.InvalidState("order berstatus %s tidak bisa diubah lagi", order.Status)
		}

		item, err := s.orders.GetItem(ctx, tx, itemID)
		if err != nil {
			return err
		}
		if item.OrderID != orderID {
			return domain.NotFound("item order")
		}
		if item.QtyReceived > 0 {
			return domain.Conflict(
				"item ini tidak bisa dihapus karena %d unit sudah diterima; kurangi jumlahnya saja",
				item.QtyReceived)
		}

		items, err := s.orders.ListItems(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if len(items) <= 1 {
			return domain.Conflict("order harus punya minimal satu item; batalkan order kalau memang tidak jadi")
		}

		// Seluruh alokasi dilepas ke stok sebelum barisnya dihapus.
		if err := s.releaseExcessAllocations(ctx, tx, item, 0, order.TripID, actorID); err != nil {
			return err
		}
		if err := s.orders.DeleteItem(ctx, tx, itemID); err != nil {
			return err
		}

		updated, err := s.orders.RecalculateTotals(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if err := s.validateDiscount(updated); err != nil {
			return err
		}
		if err := s.reconcileStatusAfterAmountChange(ctx, tx, updated); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &orderID,
			Action:   domain.AuditItemChange,
			Changes:  map[string]any{"action": "delete", "product": item.ProductName, "qty": item.Qty},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, orderID)
}

// --- Status & pembayaran ---------------------------------------------------

func (s *OrderService) ChangeStatus(ctx context.Context, id uuid.UUID, newStatus string, actorID uuid.UUID) (*domain.OrderDetail, error) {
	if !domain.IsValidOrderStatus(newStatus) {
		return nil, domain.Validationf("status order %q tidak dikenal", newStatus)
	}

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if order.Status == newStatus {
			return nil
		}
		if !domain.CanTransitionOrder(order.Status, newStatus) {
			return domain.InvalidState(
				"order tidak bisa langsung dari status %s ke %s (pilihan yang tersedia: %s)",
				order.Status, newStatus, strings.Join(domain.NextOrderStatuses(order.Status), ", "))
		}

		// Dua penjaga penting supaya status tidak berbohong tentang uang.
		switch newStatus {
		case domain.OrderDPPaid:
			paidDP, err := s.orders.TotalDPPaid(ctx, tx, id)
			if err != nil {
				return err
			}
			if paidDP.LessThan(order.DPRequired) {
				return domain.InvalidState(
					"DP yang masuk baru %s dari %s yang diminta",
					money.Format(paidDP), money.Format(order.DPRequired))
			}
		case domain.OrderPaid:
			if order.BalanceDue.GreaterThan(decimal.Zero) {
				return domain.InvalidState(
					"order belum lunas, masih ada sisa %s", money.Format(order.BalanceDue))
			}
		}

		if _, err := s.orders.UpdateStatus(ctx, tx, id, newStatus); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &id,
			Action:   domain.AuditStatusChange,
			Changes:  map[string]any{"from": order.Status, "to": newStatus},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *OrderService) Cancel(ctx context.Context, id uuid.UUID, reason *string, actorID uuid.UUID) (*domain.OrderDetail, error) {
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if order.Status == domain.OrderCancelled {
			return domain.Conflict("order ini sudah dibatalkan sebelumnya")
		}
		if !domain.CanTransitionOrder(order.Status, domain.OrderCancelled) {
			return domain.InvalidState("order berstatus %s tidak bisa dibatalkan", order.Status)
		}

		// Barang yang sudah dibeli untuk order ini dilepas menjadi stok.
		items, err := s.orders.ListItems(ctx, tx, id)
		if err != nil {
			return err
		}
		for i := range items {
			if err := s.releaseExcessAllocations(ctx, tx, &items[i], 0, order.TripID, actorID); err != nil {
				return err
			}
			if err := s.orders.SyncItemPurchasedQty(ctx, tx, items[i].ID); err != nil {
				return err
			}
		}

		if _, err := s.orders.Cancel(ctx, tx, id, trimPtr(reason)); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &id,
			Action:   domain.AuditStatusChange,
			Changes: map[string]any{
				"from": order.Status, "to": domain.OrderCancelled,
				"reason":      derefString(reason),
				"paid_amount": order.PaidAmount.String(),
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

type PaymentInput struct {
	Type      string
	Amount    decimal.Decimal
	Method    string
	Reference *string
	ProofURL  *string
	PaidAt    time.Time
	Notes     *string
}

// RecordPayment mencatat uang masuk dan menyesuaikan status order secara
// otomatis: DP lunas menaikkan status ke dp_paid, pelunasan penuh ke paid.
func (s *OrderService) RecordPayment(ctx context.Context, orderID uuid.UUID, in PaymentInput, actorID uuid.UUID) (*domain.OrderDetail, error) {
	if !domain.IsValidPaymentType(in.Type) {
		return nil, domain.Validation("jenis pembayaran tidak dikenal", map[string]string{
			"type": "pilih: dp, settlement, refund, adjustment",
		})
	}
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, domain.Validation("nominal pembayaran tidak valid", map[string]string{
			"amount": "harus lebih besar dari 0",
		})
	}

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if order.Status == domain.OrderCancelled && in.Type != domain.PaymentRefund {
			return domain.InvalidState("order sudah dibatalkan, hanya refund yang bisa dicatat")
		}
		if in.Type == domain.PaymentRefund && in.Amount.GreaterThan(order.PaidAmount) {
			return domain.Conflict(
				"refund %s melebihi uang yang sudah diterima (%s)",
				money.Format(in.Amount), money.Format(order.PaidAmount))
		}

		if _, err := s.orders.CreatePayment(ctx, tx, repository.PaymentParams{
			OrderID:    orderID,
			Type:       in.Type,
			Amount:     in.Amount,
			Method:     in.Method,
			Reference:  trimPtr(in.Reference),
			ProofURL:   trimPtr(in.ProofURL),
			PaidAt:     in.PaidAt,
			Notes:      trimPtr(in.Notes),
			RecordedBy: nullableUUID(actorID),
		}); err != nil {
			return err
		}

		updated, err := s.orders.RecalculatePaidAmount(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if err := s.invoices.SyncAmountsFromOrder(ctx, tx, orderID); err != nil {
			return err
		}
		if err := s.advanceStatusAfterPayment(ctx, tx, updated); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &orderID,
			Action:   domain.AuditPaymentRecord,
			Changes: map[string]any{
				"type": in.Type, "amount": in.Amount.String(), "method": in.Method,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, orderID)
}

func (s *OrderService) DeletePayment(ctx context.Context, orderID, paymentID uuid.UUID, actorID uuid.UUID) (*domain.OrderDetail, error) {
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := s.orders.GetForUpdate(ctx, tx, orderID); err != nil {
			return err
		}

		payment, err := s.orders.GetPayment(ctx, tx, paymentID)
		if err != nil {
			return err
		}
		if payment.OrderID != orderID {
			return domain.NotFound("pembayaran")
		}

		if err := s.orders.DeletePayment(ctx, tx, paymentID); err != nil {
			return err
		}

		updated, err := s.orders.RecalculatePaidAmount(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if err := s.invoices.SyncAmountsFromOrder(ctx, tx, orderID); err != nil {
			return err
		}
		if err := s.reconcileStatusAfterAmountChange(ctx, tx, updated); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &orderID,
			Action:   domain.AuditDelete,
			Changes: map[string]any{
				"entity": "payment", "type": payment.Type, "amount": payment.Amount.String(),
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, orderID)
}

// --- Penerimaan barang -----------------------------------------------------

type ItemReceipt struct {
	ItemID      uuid.UUID
	QtyReceived int
	// Status opsional: dipakai untuk menandai item yang ternyata tidak tersedia.
	Status string
}

// ReceiveItems mencatat pencocokan barang yang datang dengan pesanan.
// Setelah semua item terhitung, status order otomatis naik ke arrived.
func (s *OrderService) ReceiveItems(ctx context.Context, orderID uuid.UUID, receipts []ItemReceipt, actorID uuid.UUID) (*domain.OrderDetail, error) {
	if len(receipts) == 0 {
		return nil, domain.Validation("tidak ada item yang dicatat", map[string]string{
			"items": "isi minimal satu item",
		})
	}

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if order.Status == domain.OrderCancelled {
			return domain.InvalidState("order sudah dibatalkan")
		}

		for _, receipt := range receipts {
			item, err := s.orders.GetItem(ctx, tx, receipt.ItemID)
			if err != nil {
				return err
			}
			if item.OrderID != orderID {
				return domain.NotFound("item order")
			}
			if receipt.QtyReceived < 0 || receipt.QtyReceived > item.Qty {
				return domain.Validationf(
					"jumlah diterima untuk %s harus antara 0 dan %d", item.ProductName, item.Qty)
			}

			status := receipt.Status
			if status == "" {
				switch {
				case receipt.QtyReceived == 0:
					status = domain.FulfillmentUnavailable
				case receipt.QtyReceived >= item.Qty:
					status = domain.FulfillmentPurchased
				default:
					status = domain.FulfillmentPartial
				}
			}

			if _, err := s.orders.UpdateItemFulfillment(ctx, tx, receipt.ItemID, status, receipt.QtyReceived); err != nil {
				return err
			}
		}

		// Status order sengaja tidak digeser di sini. Mencocokkan barang datang
		// adalah bagian dari tahap "Diproses"; yang naik justru catatan
		// penerimaan tiap item, dan order berpindah ke "Sedang Dikemas" ketika
		// benar-benar dikemas.

		// Yang berubah adalah uangnya. Barang yang tidak berhasil dibeli tidak
		// boleh ikut ditagihkan: tanpa hitung ulang di sini, invoice pelunasan
		// memuat barang yang tidak akan pernah dikirim ke customer.
		if _, err := s.orders.RecalculateTotals(ctx, tx, orderID); err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "order",
			EntityID: &orderID,
			Action:   "receive",
			Changes:  map[string]any{"item_count": len(receipts)},
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, orderID)
}

// --- Helper internal -------------------------------------------------------

func (s *OrderService) addItemTx(ctx context.Context, tx pgx.Tx, order *domain.Order, trip *domain.Trip, in OrderItemInput) (*domain.OrderItem, error) {
	if in.Qty < 1 {
		return nil, domain.Validation("jumlah tidak valid", map[string]string{
			"qty": "minimal 1",
		})
	}

	product, err := s.products.GetByID(ctx, tx, in.ProductID)
	if err != nil {
		return nil, err
	}

	// Harga diambil dari katalog trip. Produk yang belum masuk katalog hanya
	// bisa dipesan kalau admin menyebutkan harganya sendiri.
	var (
		tripItemID  *uuid.UUID
		unitPrice   decimal.Decimal
		unitCostEst decimal.Decimal
	)

	tripItem, err := s.trips.GetItemByProduct(ctx, tx, trip.ID, in.ProductID)
	switch {
	case err == nil:
		tripItemID = &tripItem.ID
		unitPrice = tripItem.SellPrice
		unitCostEst = tripItem.CostPriceIDR

		if tripItem.MaxQty != nil {
			ordered, err := s.trips.CountItemOrders(ctx, tx, tripItem.ID)
			if err != nil {
				return nil, err
			}
			if ordered+in.Qty > *tripItem.MaxQty {
				return nil, domain.Conflict(
					"kuota %s pada trip ini tinggal %d unit",
					product.Name, max(*tripItem.MaxQty-ordered, 0))
			}
		}
	default:
		if domainErr, ok := domain.AsError(err); !ok || domainErr.Code != domain.CodeNotFound {
			return nil, err
		}
		if in.UnitPrice == nil {
			return nil, domain.Validationf(
				"produk %s belum ada di katalog trip ini, tambahkan ke katalog atau isi harga jualnya",
				product.Name)
		}
		unitCostEst = money.Convert(product.BasePrice, trip.ExchangeRate)
	}

	if in.UnitPrice != nil {
		if in.UnitPrice.IsNegative() {
			return nil, domain.Validation("harga tidak valid", map[string]string{
				"unit_price": "harus 0 atau lebih",
			})
		}
		unitPrice = *in.UnitPrice
	}

	// Produk yang sudah ada di order ditambah qty-nya, bukan digandakan.
	existing, err := s.orders.FindItemByProduct(ctx, tx, order.ID, in.ProductID)
	if err == nil {
		return s.orders.UpdateItem(ctx, tx, existing.ID, existing.Qty+in.Qty, unitPrice, existing.Notes)
	}
	if domainErr, ok := domain.AsError(err); !ok || domainErr.Code != domain.CodeNotFound {
		return nil, err
	}

	return s.orders.AddItem(ctx, tx, repository.OrderItemParams{
		OrderID:     order.ID,
		ProductID:   in.ProductID,
		TripItemID:  tripItemID,
		ProductName: product.Name,
		ProductSKU:  product.SKU,
		Qty:         in.Qty,
		UnitPrice:   unitPrice,
		UnitCostEst: unitCostEst,
		Notes:       trimPtr(in.Notes),
	})
}

// releaseExcessAllocations melepas alokasi pembelian yang melebihi qty pesanan
// terbaru, lalu memindahkan unit-unit itu ke stok. Alokasi termuda dilepas
// lebih dulu supaya pemenuhan pesanan yang lebih awal tidak terganggu.
func (s *OrderService) releaseExcessAllocations(
	ctx context.Context, tx pgx.Tx, item *domain.OrderItem, newQty int, tripID, actorID uuid.UUID,
) error {
	allocated, err := s.purchases.AllocatedQtyByOrderItem(ctx, tx, item.ID)
	if err != nil {
		return err
	}
	excess := allocated - newQty
	if excess <= 0 {
		return nil
	}

	allocations, err := s.purchases.ListAllocationsByOrderItem(ctx, tx, item.ID)
	if err != nil {
		return err
	}

	refType := "order_item"
	for _, alloc := range allocations {
		if excess <= 0 {
			break
		}

		release := min(excess, alloc.Qty)
		if release == alloc.Qty {
			if err := s.purchases.DeleteAllocation(ctx, tx, alloc.ID); err != nil {
				return err
			}
		} else {
			if err := s.purchases.UpdateAllocationQty(ctx, tx, alloc.ID, alloc.Qty-release); err != nil {
				return err
			}
		}

		// Unit yang dilepas tetap tercatat sebagai hasil pembelian, hanya
		// pemiliknya berubah dari order menjadi stok.
		if _, err := s.purchases.CreateAllocation(ctx, tx, alloc.PurchaseID, nil, release, alloc.UnitCostIDR); err != nil {
			return err
		}
		if _, err := s.purchases.StockIn(ctx, tx, item.ProductID, release, alloc.UnitCostIDR); err != nil {
			return err
		}
		if _, err := s.purchases.CreateMovement(ctx, tx, repository.StockMovementParams{
			ProductID:   item.ProductID,
			Type:        domain.StockInPurchase,
			Qty:         release,
			UnitCostIDR: alloc.UnitCostIDR,
			TripID:      nullableUUID(tripID),
			RefType:     &refType,
			RefID:       &item.ID,
			Note:        strPtr("dilepas dari pesanan yang dikurangi/dibatalkan"),
			CreatedBy:   nullableUUID(actorID),
		}); err != nil {
			return err
		}

		excess -= release
	}

	return nil
}

// applyDPRequired menetapkan nominal DP. Kalau admin tidak menentukan, dipakai
// persentase default dari konfigurasi.
func (s *OrderService) applyDPRequired(ctx context.Context, tx pgx.Tx, order *domain.Order, explicit *decimal.Decimal) error {
	dp := money.Percent(order.Total, s.business.DefaultDPPercent)
	if explicit != nil {
		if explicit.IsNegative() {
			return domain.Validation("nominal DP tidak valid", map[string]string{
				"dp_required": "harus 0 atau lebih",
			})
		}
		if explicit.GreaterThan(order.Total) {
			return domain.Validationf(
				"DP %s melebihi total order %s", money.Format(*explicit), money.Format(order.Total))
		}
		dp = *explicit
	}

	_, err := s.orders.Update(ctx, tx, order.ID, repository.UpdateOrderParams{
		OrderDate:           order.OrderDate,
		OrderSource:         order.OrderSource,
		Discount:            order.Discount,
		ShippingFee:         order.ShippingFee,
		DPRequired:          dp,
		RecipientName:       order.RecipientName,
		RecipientPhone:      order.RecipientPhone,
		ShippingAddress:     order.ShippingAddress,
		ShippingCity:        order.ShippingCity,
		ShippingDistrict:    order.ShippingDistrict,
		ShippingSubdistrict: order.ShippingSubdistrict,
		ShippingProvince:    order.ShippingProvince,
		ShippingPostalCode:  order.ShippingPostalCode,
		Notes:               order.Notes,
	})
	return err
}

func (s *OrderService) validateDiscount(order *domain.Order) error {
	if order.Discount.GreaterThan(order.Subtotal) {
		return domain.Validationf(
			"diskon %s melebihi subtotal %s", money.Format(order.Discount), money.Format(order.Subtotal))
	}
	return nil
}

// advanceStatusAfterPayment menaikkan status order mengikuti uang yang masuk.
func (s *OrderService) advanceStatusAfterPayment(ctx context.Context, tx pgx.Tx, order *domain.Order) error {
	switch {
	case order.IsFullyPaid() && domain.CanTransitionOrder(order.Status, domain.OrderPaid):
		_, err := s.orders.UpdateStatus(ctx, tx, order.ID, domain.OrderPaid)
		return err

	case order.Status == domain.OrderAwaitingDP && order.PaidAmount.GreaterThanOrEqual(order.DPRequired) &&
		order.DPRequired.GreaterThan(decimal.Zero):
		_, err := s.orders.UpdateStatus(ctx, tx, order.ID, domain.OrderDPPaid)
		return err
	}
	return nil
}

// reconcileStatusAfterAmountChange menurunkan status order yang tadinya lunas
// tapi totalnya bertambah setelah diedit, supaya sisa tagihan yang baru tidak
// tersembunyi di balik status "paid".
func (s *OrderService) reconcileStatusAfterAmountChange(ctx context.Context, tx pgx.Tx, order *domain.Order) error {
	switch {
	// Sudah ditandai lunas, tapi ternyata masih ada sisa: kembali ke penagihan.
	case order.Status == domain.OrderPaid && order.BalanceDue.GreaterThan(decimal.Zero):
		_, err := s.orders.UpdateStatus(ctx, tx, order.ID, domain.OrderInvoiced)
		return err

	// DP-nya ternyata tidak lagi tertutup — biasanya karena pembayaran yang
	// salah catat dihapus. Tanpa langkah mundur ini ordernya tetap berlabel
	// Diproses, ikut masuk daftar belanja tripper, dan bisa terus dikemas
	// sampai dikirim padahal customer belum menyetor sepeser pun.
	case order.Status == domain.OrderDPPaid &&
		order.DPRequired.GreaterThan(decimal.Zero) &&
		order.PaidAmount.LessThan(order.DPRequired):
		_, err := s.orders.UpdateStatus(ctx, tx, order.ID, domain.OrderAwaitingDP)
		return err
	}
	return nil
}

type shippingInfo struct {
	Name        string
	Phone       string
	Address     string
	City        string
	District    *string
	Subdistrict *string
	Province    *string
	PostalCode  *string
}

// resolveShipping menyalin alamat customer sebagai default, tapi tetap
// mengizinkan pengiriman ke alamat lain (hadiah, kantor, titip teman).
func resolveShipping(in CreateOrderInput, customer *domain.Customer) shippingInfo {
	info := shippingInfo{
		Name:        strings.TrimSpace(customer.Name),
		Phone:       customer.PhoneWA,
		City:        derefString(customer.City),
		Address:     derefString(customer.Address),
		District:    customer.District,
		Subdistrict: customer.Subdistrict,
		Province:    customer.Province,
		PostalCode:  customer.PostalCode,
	}

	if v := trimPtr(in.RecipientName); v != nil {
		info.Name = *v
	}
	if v := trimPtr(in.RecipientPhone); v != nil {
		info.Phone = domain.NormalizePhoneWA(*v)
	}
	if v := trimPtr(in.ShippingAddress); v != nil {
		info.Address = *v
	}
	if v := trimPtr(in.ShippingCity); v != nil {
		info.City = *v
	}
	if v := trimPtr(in.ShippingDistrict); v != nil {
		info.District = v
	}
	if v := trimPtr(in.ShippingSubdistrict); v != nil {
		info.Subdistrict = v
	}
	if v := trimPtr(in.ShippingProvince); v != nil {
		info.Province = v
	}
	if v := trimPtr(in.ShippingPostalCode); v != nil {
		info.PostalCode = v
	}

	return info
}

func validateShipping(info shippingInfo) error {
	fields := map[string]string{}

	if info.Name == "" {
		fields["recipient_name"] = "wajib diisi"
	}
	if info.Phone == "" {
		fields["recipient_phone"] = "wajib diisi"
	}
	if info.Address == "" {
		fields["shipping_address"] = "wajib diisi, lengkapi alamat customer atau isi di sini"
	}
	if info.City == "" {
		fields["shipping_city"] = "wajib diisi"
	}

	if len(fields) > 0 {
		return domain.Validation("alamat pengiriman belum lengkap", fields)
	}
	return nil
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func strPtr(v string) *string { return &v }
