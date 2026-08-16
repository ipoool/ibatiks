package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/docnum"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type TripService struct {
	pool     *pgxpool.Pool
	trips    *repository.TripRepo
	products *repository.ProductRepo
	orders   *repository.OrderRepo
	audit    *repository.AuditRepo
}

func NewTripService(
	pool *pgxpool.Pool,
	trips *repository.TripRepo,
	products *repository.ProductRepo,
	orders *repository.OrderRepo,
	audit *repository.AuditRepo,
) *TripService {
	return &TripService{pool: pool, trips: trips, products: products, orders: orders, audit: audit}
}

type TripInput struct {
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
}

func (s *TripService) Create(ctx context.Context, in TripInput, actorID uuid.UUID) (*domain.Trip, error) {
	if err := validateTripInput(in); err != nil {
		return nil, err
	}

	var trip *domain.Trip
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		code, err := docnum.Next(ctx, tx, docnum.Trip, in.DepartDate.Year())
		if err != nil {
			return err
		}

		trip, err = s.trips.Create(ctx, tx, repository.TripParams{
			Code:          code,
			Title:         strings.TrimSpace(in.Title),
			Country:       strings.TrimSpace(in.Country),
			City:          trimPtr(in.City),
			TripperUserID: in.TripperUserID,
			DepartDate:    in.DepartDate,
			ReturnDate:    in.ReturnDate,
			OrderDeadline: in.OrderDeadline,
			Currency:      strings.ToUpper(strings.TrimSpace(in.Currency)),
			ExchangeRate:  in.ExchangeRate,
			Notes:         trimPtr(in.Notes),
			CreatedBy:     nullableUUID(actorID),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return trip, nil
}

func (s *TripService) List(ctx context.Context, p pagination.Params, status string) ([]domain.TripDetail, int64, error) {
	return s.trips.List(ctx, s.pool, p, status)
}

func (s *TripService) Get(ctx context.Context, id uuid.UUID) (*domain.TripDetail, error) {
	return s.trips.GetDetail(ctx, s.pool, id)
}

// Update mengubah data trip. Kalau kurs berubah, harga di katalog tidak ikut
// berubah otomatis: harga yang sudah dipublikasikan ke customer tidak boleh
// bergeser diam-diam. Admin memicunya secara sadar lewat RecalculatePrices.
func (s *TripService) Update(ctx context.Context, id uuid.UUID, in TripInput) (*domain.Trip, error) {
	if err := validateTripInput(in); err != nil {
		return nil, err
	}

	trip, err := s.trips.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	if trip.Status == domain.TripSettled || trip.Status == domain.TripCancelled {
		return nil, domain.InvalidState("trip berstatus %s tidak bisa diubah lagi", trip.Status)
	}

	return s.trips.Update(ctx, s.pool, id, repository.TripParams{
		Title:         strings.TrimSpace(in.Title),
		Country:       strings.TrimSpace(in.Country),
		City:          trimPtr(in.City),
		TripperUserID: in.TripperUserID,
		DepartDate:    in.DepartDate,
		ReturnDate:    in.ReturnDate,
		OrderDeadline: in.OrderDeadline,
		Currency:      strings.ToUpper(strings.TrimSpace(in.Currency)),
		ExchangeRate:  in.ExchangeRate,
		Notes:         trimPtr(in.Notes),
	})
}

// ChangeStatus menggeser status trip sekaligus menyeret status order-order di
// dalamnya, supaya admin tidak perlu mengubah puluhan order satu per satu.
func (s *TripService) ChangeStatus(ctx context.Context, id uuid.UUID, newStatus string, actorID uuid.UUID) (*domain.Trip, error) {
	if !domain.IsValidTripStatus(newStatus) {
		return nil, domain.Validationf("status trip %q tidak dikenal", newStatus)
	}

	trip, err := s.trips.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	if trip.Status == newStatus {
		return trip, nil
	}
	if !domain.CanTransitionTrip(trip.Status, newStatus) {
		return nil, domain.InvalidState(
			"trip tidak bisa langsung dari status %s ke %s (pilihan yang tersedia: %s)",
			trip.Status, newStatus, strings.Join(domain.NextTripStatuses(trip.Status), ", "))
	}

	var updated *domain.Trip
	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		var err error
		updated, err = s.trips.UpdateStatus(ctx, tx, id, newStatus)
		if err != nil {
			return err
		}

		// Efek samping yang wajar untuk tiap perpindahan status trip.
		switch newStatus {
		case domain.TripShopping:
			// Order yang DP-nya sudah masuk otomatis masuk tahap dibelikan.
			if _, err := s.orders.BulkUpdateStatus(ctx, tx, id,
				[]string{domain.OrderDPPaid}, domain.OrderPurchasing); err != nil {
				return err
			}
		case domain.TripArrived:
			// Barang sudah di Indonesia: order yang sedang dibelikan siap dicocokkan.
			if _, err := s.orders.BulkUpdateStatus(ctx, tx, id,
				[]string{domain.OrderPurchasing}, domain.OrderArrived); err != nil {
				return err
			}
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "trip",
			EntityID: &id,
			Action:   domain.AuditStatusChange,
			Changes:  map[string]any{"from": trip.Status, "to": newStatus},
		})
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete hanya diizinkan untuk trip yang belum punya order sama sekali.
func (s *TripService) Delete(ctx context.Context, id uuid.UUID) error {
	count, err := s.trips.CountOrders(ctx, s.pool, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.Conflict("trip sudah punya %d order, batalkan trip alih-alih menghapusnya", count)
	}
	return s.trips.Delete(ctx, s.pool, id)
}

// --- Katalog trip ----------------------------------------------------------

type TripItemInput struct {
	ProductID   uuid.UUID
	CostPrice   decimal.Decimal
	MarkupType  string
	MarkupValue decimal.Decimal
	MaxQty      *int
	IsActive    bool
	Notes       *string
}

// AddItem memasukkan produk ke katalog trip. Harga jual dihitung sekali di sini
// dari kurs trip dan markup, lalu disimpan.
func (s *TripService) AddItem(ctx context.Context, tripID uuid.UUID, in TripItemInput) (*domain.TripItem, error) {
	trip, err := s.trips.GetByID(ctx, s.pool, tripID)
	if err != nil {
		return nil, err
	}
	if trip.Status == domain.TripSettled || trip.Status == domain.TripCancelled {
		return nil, domain.InvalidState("katalog trip berstatus %s tidak bisa diubah", trip.Status)
	}

	product, err := s.products.GetByID(ctx, s.pool, in.ProductID)
	if err != nil {
		return nil, err
	}

	markupType, markupValue := resolveMarkup(in.MarkupType, in.MarkupValue, product)
	if !domain.IsValidMarkupType(markupType) {
		return nil, domain.Validation("tipe markup tidak dikenal", map[string]string{
			"markup_type": "pilih percent atau nominal",
		})
	}

	costIDR, sellPrice := domain.CalculateSellPrice(in.CostPrice, trip.ExchangeRate, markupType, markupValue)

	return s.trips.CreateItem(ctx, s.pool, repository.TripItemParams{
		TripID:       tripID,
		ProductID:    in.ProductID,
		CostPrice:    in.CostPrice,
		CostPriceIDR: costIDR,
		MarkupType:   markupType,
		MarkupValue:  markupValue,
		SellPrice:    sellPrice,
		MaxQty:       in.MaxQty,
		IsActive:     in.IsActive,
		Notes:        trimPtr(in.Notes),
	})
}

func (s *TripService) ListItems(ctx context.Context, tripID uuid.UUID) ([]domain.TripItemDetail, error) {
	return s.trips.ListItems(ctx, s.pool, tripID)
}

func (s *TripService) UpdateItem(ctx context.Context, tripID, itemID uuid.UUID, in TripItemInput) (*domain.TripItem, error) {
	trip, err := s.trips.GetByID(ctx, s.pool, tripID)
	if err != nil {
		return nil, err
	}

	item, err := s.trips.GetItem(ctx, s.pool, itemID)
	if err != nil {
		return nil, err
	}
	if item.TripID != tripID {
		return nil, domain.NotFound("item katalog")
	}

	markupType := in.MarkupType
	if !domain.IsValidMarkupType(markupType) {
		return nil, domain.Validation("tipe markup tidak dikenal", map[string]string{
			"markup_type": "pilih percent atau nominal",
		})
	}

	costIDR, sellPrice := domain.CalculateSellPrice(in.CostPrice, trip.ExchangeRate, markupType, in.MarkupValue)

	return s.trips.UpdateItem(ctx, s.pool, itemID, repository.TripItemParams{
		CostPrice:    in.CostPrice,
		CostPriceIDR: costIDR,
		MarkupType:   markupType,
		MarkupValue:  in.MarkupValue,
		SellPrice:    sellPrice,
		MaxQty:       in.MaxQty,
		IsActive:     in.IsActive,
		Notes:        trimPtr(in.Notes),
	})
}

// DeleteItem menolak menghapus produk yang sudah terlanjur dipesan customer,
// karena order yang mereferensikannya akan kehilangan konteks harga.
func (s *TripService) DeleteItem(ctx context.Context, tripID, itemID uuid.UUID) error {
	item, err := s.trips.GetItem(ctx, s.pool, itemID)
	if err != nil {
		return err
	}
	if item.TripID != tripID {
		return domain.NotFound("item katalog")
	}

	ordered, err := s.trips.CountItemOrders(ctx, s.pool, itemID)
	if err != nil {
		return err
	}
	if ordered > 0 {
		return domain.Conflict(
			"produk ini sudah dipesan sebanyak %d unit, nonaktifkan saja alih-alih menghapus", ordered)
	}

	return s.trips.DeleteItem(ctx, s.pool, itemID)
}

// RecalculatePrices menghitung ulang seluruh harga katalog memakai kurs trip
// terkini. Sengaja jadi aksi eksplisit karena mengubah harga yang sudah
// diumumkan ke customer adalah keputusan bisnis, bukan efek samping edit data.
func (s *TripService) RecalculatePrices(ctx context.Context, tripID uuid.UUID) ([]domain.TripItemDetail, error) {
	trip, err := s.trips.GetByID(ctx, s.pool, tripID)
	if err != nil {
		return nil, err
	}

	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		items, err := s.trips.ListItems(ctx, tx, tripID)
		if err != nil {
			return err
		}

		for _, item := range items {
			costIDR, sellPrice := domain.CalculateSellPrice(
				item.CostPrice, trip.ExchangeRate, item.MarkupType, item.MarkupValue)

			if _, err := s.trips.UpdateItem(ctx, tx, item.ID, repository.TripItemParams{
				CostPrice:    item.CostPrice,
				CostPriceIDR: costIDR,
				MarkupType:   item.MarkupType,
				MarkupValue:  item.MarkupValue,
				SellPrice:    sellPrice,
				MaxQty:       item.MaxQty,
				IsActive:     item.IsActive,
				Notes:        item.Notes,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.trips.ListItems(ctx, s.pool, tripID)
}

// --- Biaya perjalanan ------------------------------------------------------

type TripExpenseInput struct {
	Category    string
	Description string
	Amount      decimal.Decimal
	SpentAt     time.Time
	ReceiptURL  *string
}

func (s *TripService) AddExpense(ctx context.Context, tripID uuid.UUID, in TripExpenseInput, actorID uuid.UUID) (*domain.TripExpense, error) {
	if !domain.IsValidExpenseCategory(in.Category) {
		return nil, domain.Validation("kategori biaya tidak dikenal", map[string]string{
			"category": "pilih: tiket, bagasi, akomodasi, transport, visa, lainnya",
		})
	}
	if in.Amount.IsNegative() {
		return nil, domain.Validation("nominal biaya tidak valid", map[string]string{
			"amount": "harus 0 atau lebih",
		})
	}
	if _, err := s.trips.GetByID(ctx, s.pool, tripID); err != nil {
		return nil, err
	}

	return s.trips.CreateExpense(ctx, s.pool, repository.TripExpenseParams{
		TripID:      tripID,
		Category:    in.Category,
		Description: strings.TrimSpace(in.Description),
		Amount:      in.Amount,
		SpentAt:     in.SpentAt,
		ReceiptURL:  trimPtr(in.ReceiptURL),
		CreatedBy:   nullableUUID(actorID),
	})
}

func (s *TripService) ListExpenses(ctx context.Context, tripID uuid.UUID) ([]domain.TripExpense, error) {
	return s.trips.ListExpenses(ctx, s.pool, tripID)
}

func (s *TripService) UpdateExpense(ctx context.Context, id uuid.UUID, in TripExpenseInput) (*domain.TripExpense, error) {
	if !domain.IsValidExpenseCategory(in.Category) {
		return nil, domain.Validation("kategori biaya tidak dikenal", map[string]string{
			"category": "pilih: tiket, bagasi, akomodasi, transport, visa, lainnya",
		})
	}
	return s.trips.UpdateExpense(ctx, s.pool, id, repository.TripExpenseParams{
		Category:    in.Category,
		Description: strings.TrimSpace(in.Description),
		Amount:      in.Amount,
		SpentAt:     in.SpentAt,
		ReceiptURL:  trimPtr(in.ReceiptURL),
	})
}

func (s *TripService) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	return s.trips.DeleteExpense(ctx, s.pool, id)
}

func validateTripInput(in TripInput) error {
	fields := map[string]string{}

	if in.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		fields["exchange_rate"] = "kurs harus lebih besar dari 0"
	}
	if in.ReturnDate.Before(in.DepartDate) {
		fields["return_date"] = "tanggal pulang tidak boleh lebih awal dari tanggal berangkat"
	}
	if in.OrderDeadline != nil && in.OrderDeadline.After(in.ReturnDate) {
		fields["order_deadline"] = "batas order tidak boleh melewati tanggal pulang"
	}

	if len(fields) > 0 {
		return domain.Validation("data trip tidak valid", fields)
	}
	return nil
}

// resolveMarkup memakai markup bawaan produk kalau admin tidak menentukan
// markup khusus untuk trip ini.
func resolveMarkup(markupType string, markupValue decimal.Decimal, product *domain.Product) (string, decimal.Decimal) {
	if markupType == "" {
		return product.MarkupType, product.MarkupValue
	}
	return markupType, markupValue
}

func nullableUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
