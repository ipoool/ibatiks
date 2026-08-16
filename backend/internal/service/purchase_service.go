package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type PurchaseService struct {
	pool      *pgxpool.Pool
	purchases *repository.PurchaseRepo
	trips     *repository.TripRepo
	orders    *repository.OrderRepo
	products  *repository.ProductRepo
	audit     *repository.AuditRepo
}

func NewPurchaseService(
	pool *pgxpool.Pool,
	purchases *repository.PurchaseRepo,
	trips *repository.TripRepo,
	orders *repository.OrderRepo,
	products *repository.ProductRepo,
	audit *repository.AuditRepo,
) *PurchaseService {
	return &PurchaseService{
		pool: pool, purchases: purchases, trips: trips,
		orders: orders, products: products, audit: audit,
	}
}

// ShoppingList adalah daftar belanja yang dibawa tripper: apa saja yang harus
// dibeli pada trip ini dan berapa yang sudah terbeli.
func (s *PurchaseService) ShoppingList(ctx context.Context, tripID uuid.UUID) ([]domain.ShoppingListEntry, error) {
	if _, err := s.trips.GetByID(ctx, s.pool, tripID); err != nil {
		return nil, err
	}
	return s.purchases.ShoppingList(ctx, s.pool, tripID)
}

type PurchaseInput struct {
	ProductID       uuid.UUID
	PurchaseDate    time.Time
	Qty             int
	UnitCostForeign decimal.Decimal
	// ExchangeRate opsional: kalau nil, dipakai kurs trip.
	ExchangeRate *decimal.Decimal
	StoreName    *string
	ReceiptURL   *string
	Notes        *string
}

// PurchaseResult melaporkan ke mana barang hasil belanja dialokasikan, supaya
// tripper langsung tahu berapa yang menutup pesanan dan berapa yang jadi stok.
type PurchaseResult struct {
	Purchase    *domain.Purchase                  `json:"purchase"`
	Allocations []domain.PurchaseAllocationDetail `json:"allocations"`
	QtyToOrders int                               `json:"qty_to_orders"`
	QtyToStock  int                               `json:"qty_to_stock"`
}

// Record mencatat belanja tripper lalu langsung mengalokasikannya.
//
// Alokasi berjalan FIFO menurut tanggal order: customer yang memesan lebih dulu
// dipenuhi lebih dulu. Sisa yang tidak dipesan siapa pun otomatis masuk stok
// untuk dijual di marketplace — inilah yang membedakan barang titipan dengan
// barang kulakan, dan keduanya diperlakukan berbeda di laporan profit.
func (s *PurchaseService) Record(ctx context.Context, tripID uuid.UUID, in PurchaseInput, actorID uuid.UUID) (*PurchaseResult, error) {
	if in.Qty < 1 {
		return nil, domain.Validation("jumlah pembelian tidak valid", map[string]string{
			"qty": "minimal 1",
		})
	}
	if in.UnitCostForeign.IsNegative() {
		return nil, domain.Validation("harga beli tidak valid", map[string]string{
			"unit_cost_foreign": "harus 0 atau lebih",
		})
	}

	trip, err := s.trips.GetByID(ctx, s.pool, tripID)
	if err != nil {
		return nil, err
	}
	if trip.Status == domain.TripCancelled {
		return nil, domain.InvalidState("trip sudah dibatalkan")
	}
	if _, err := s.products.GetByID(ctx, s.pool, in.ProductID); err != nil {
		return nil, err
	}

	rate := trip.ExchangeRate
	if in.ExchangeRate != nil {
		if in.ExchangeRate.LessThanOrEqual(decimal.Zero) {
			return nil, domain.Validation("kurs tidak valid", map[string]string{
				"exchange_rate": "harus lebih besar dari 0",
			})
		}
		rate = *in.ExchangeRate
	}

	unitCostIDR := money.Convert(in.UnitCostForeign, rate)
	totalCostIDR := money.RoundRupiah(unitCostIDR.Mul(decimal.NewFromInt(int64(in.Qty))))

	var result PurchaseResult
	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		purchase, err := s.purchases.Create(ctx, tx, repository.PurchaseParams{
			TripID:          tripID,
			ProductID:       in.ProductID,
			PurchaseDate:    in.PurchaseDate,
			Qty:             in.Qty,
			UnitCostForeign: in.UnitCostForeign,
			Currency:        trip.Currency,
			ExchangeRate:    rate,
			UnitCostIDR:     unitCostIDR,
			TotalCostIDR:    totalCostIDR,
			StoreName:       trimPtr(in.StoreName),
			ReceiptURL:      trimPtr(in.ReceiptURL),
			Notes:           trimPtr(in.Notes),
			PurchasedBy:     nullableUUID(actorID),
		})
		if err != nil {
			return err
		}

		toOrders, toStock, err := s.allocate(ctx, tx, purchase, tripID, actorID)
		if err != nil {
			return err
		}

		allocations, err := s.purchases.ListAllocations(ctx, tx, purchase.ID)
		if err != nil {
			return err
		}

		result = PurchaseResult{
			Purchase:    purchase,
			Allocations: allocations,
			QtyToOrders: toOrders,
			QtyToStock:  toStock,
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "purchase",
			EntityID: &purchase.ID,
			Action:   domain.AuditCreate,
			Changes: map[string]any{
				"qty": in.Qty, "unit_cost_idr": unitCostIDR.String(),
				"to_orders": toOrders, "to_stock": toStock,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// allocate membagi hasil satu pembelian ke pesanan yang menunggu, lalu sisanya
// ke stok. Mengembalikan jumlah unit yang masuk ke pesanan dan ke stok.
func (s *PurchaseService) allocate(
	ctx context.Context, tx pgx.Tx, purchase *domain.Purchase, tripID, actorID uuid.UUID,
) (toOrders, toStock int, err error) {
	pending, err := s.purchases.ListPendingOrderItems(ctx, tx, tripID, purchase.ProductID)
	if err != nil {
		return 0, 0, err
	}

	remaining := purchase.Qty
	refType := "purchase"

	for _, item := range pending {
		if remaining <= 0 {
			break
		}

		give := min(remaining, item.QtyNeeded)
		if give <= 0 {
			continue
		}

		if _, err := s.purchases.CreateAllocation(
			ctx, tx, purchase.ID, &item.OrderItemID, give, purchase.UnitCostIDR); err != nil {
			return 0, 0, err
		}
		if err := s.orders.SyncItemPurchasedQty(ctx, tx, item.OrderItemID); err != nil {
			return 0, 0, err
		}

		remaining -= give
		toOrders += give
	}

	// Sisanya menjadi stok: uangnya sudah keluar, tapi nilainya masih dipegang
	// sebagai barang, jadi tidak dibebankan sebagai HPP trip ini.
	if remaining > 0 {
		if _, err := s.purchases.CreateAllocation(ctx, tx, purchase.ID, nil, remaining, purchase.UnitCostIDR); err != nil {
			return 0, 0, err
		}
		if _, err := s.purchases.StockIn(ctx, tx, purchase.ProductID, remaining, purchase.UnitCostIDR); err != nil {
			return 0, 0, err
		}
		if _, err := s.purchases.CreateMovement(ctx, tx, repository.StockMovementParams{
			ProductID:   purchase.ProductID,
			Type:        domain.StockInPurchase,
			Qty:         remaining,
			UnitCostIDR: purchase.UnitCostIDR,
			TripID:      &tripID,
			RefType:     &refType,
			RefID:       &purchase.ID,
			Note:        strPtr("kelebihan belanja masuk stok"),
			CreatedBy:   nullableUUID(actorID),
		}); err != nil {
			return 0, 0, err
		}
		toStock = remaining
	}

	return toOrders, toStock, nil
}

func (s *PurchaseService) List(ctx context.Context, p pagination.Params, tripID, productID *uuid.UUID) ([]domain.PurchaseDetail, int64, error) {
	return s.purchases.List(ctx, s.pool, p, tripID, productID)
}

func (s *PurchaseService) Get(ctx context.Context, id uuid.UUID) (*domain.Purchase, error) {
	return s.purchases.GetByID(ctx, s.pool, id)
}

func (s *PurchaseService) ListAllocations(ctx context.Context, purchaseID uuid.UUID) ([]domain.PurchaseAllocationDetail, error) {
	return s.purchases.ListAllocations(ctx, s.pool, purchaseID)
}

// Delete membatalkan pembelian sekaligus seluruh dampaknya: alokasi dilepas,
// stok yang sempat bertambah ditarik kembali, dan qty terbeli pada pesanan
// dihitung ulang.
func (s *PurchaseService) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		purchase, err := s.purchases.GetByID(ctx, tx, id)
		if err != nil {
			return err
		}

		allocations, err := s.purchases.ListAllocations(ctx, tx, id)
		if err != nil {
			return err
		}

		var stockQty int
		affectedItems := make([]uuid.UUID, 0, len(allocations))
		for _, alloc := range allocations {
			if alloc.OrderItemID == nil {
				stockQty += alloc.Qty
			} else {
				affectedItems = append(affectedItems, *alloc.OrderItemID)
			}
		}

		if stockQty > 0 {
			if _, err := s.purchases.StockOut(ctx, tx, purchase.ProductID, stockQty); err != nil {
				return err
			}
			refType := "purchase"
			if _, err := s.purchases.CreateMovement(ctx, tx, repository.StockMovementParams{
				ProductID:   purchase.ProductID,
				Type:        domain.StockAdjustment,
				Qty:         -stockQty,
				UnitCostIDR: purchase.UnitCostIDR,
				TripID:      &purchase.TripID,
				RefType:     &refType,
				RefID:       &purchase.ID,
				Note:        strPtr("pembelian dihapus, stok ditarik kembali"),
				CreatedBy:   nullableUUID(actorID),
			}); err != nil {
				return err
			}
		}

		if err := s.purchases.Delete(ctx, tx, id); err != nil {
			return err
		}
		// Alokasi ikut terhapus lewat ON DELETE CASCADE, jadi qty terbeli pada
		// baris pesanan tinggal diselaraskan ulang.
		for _, itemID := range affectedItems {
			if err := s.orders.SyncItemPurchasedQty(ctx, tx, itemID); err != nil {
				return err
			}
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "purchase",
			EntityID: &id,
			Action:   domain.AuditDelete,
			Changes:  map[string]any{"qty": purchase.Qty, "stock_reverted": stockQty},
		})
	})
}

// --- Stok ------------------------------------------------------------------

func (s *PurchaseService) ListStock(ctx context.Context, p pagination.Params, inStockOnly bool) ([]domain.StockItemDetail, int64, error) {
	return s.purchases.ListStock(ctx, s.pool, p, inStockOnly)
}

func (s *PurchaseService) ListMovements(ctx context.Context, p pagination.Params, productID *uuid.UUID) ([]domain.StockMovementDetail, int64, error) {
	return s.purchases.ListMovements(ctx, s.pool, p, productID)
}

type StockSaleInput struct {
	ProductID uuid.UUID
	Qty       int
	SalePrice decimal.Decimal
	Channel   string
	Note      *string
}

// SellFromStock mencatat penjualan stok di marketplace. HPP-nya diambil dari
// harga rata-rata stok, sehingga margin penjualan marketplace terpisah rapi
// dari profit trip.
func (s *PurchaseService) SellFromStock(ctx context.Context, in StockSaleInput, actorID uuid.UUID) (*domain.StockMovement, error) {
	if in.Qty < 1 {
		return nil, domain.Validation("jumlah tidak valid", map[string]string{"qty": "minimal 1"})
	}
	if in.SalePrice.IsNegative() {
		return nil, domain.Validation("harga jual tidak valid", map[string]string{
			"sale_price": "harus 0 atau lebih",
		})
	}

	var movement *domain.StockMovement
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		stock, err := s.purchases.GetStock(ctx, tx, in.ProductID)
		if err != nil {
			return err
		}
		avgCost := stock.AvgCostIDR

		if _, err := s.purchases.StockOut(ctx, tx, in.ProductID, in.Qty); err != nil {
			return err
		}

		note := in.Note
		if in.Channel != "" {
			combined := "terjual di " + in.Channel
			if note != nil {
				combined += " — " + *note
			}
			note = &combined
		}

		salePrice := in.SalePrice
		movement, err = s.purchases.CreateMovement(ctx, tx, repository.StockMovementParams{
			ProductID:    in.ProductID,
			Type:         domain.StockOutMarketplace,
			Qty:          -in.Qty,
			UnitCostIDR:  avgCost,
			SalePriceIDR: &salePrice,
			Note:         note,
			CreatedBy:    nullableUUID(actorID),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return movement, nil
}

type StockAdjustInput struct {
	ProductID uuid.UUID
	NewQty    int
	Note      *string
}

// AdjustStock menyetel jumlah stok mengikuti hasil stock opname.
func (s *PurchaseService) AdjustStock(ctx context.Context, in StockAdjustInput, actorID uuid.UUID) (*domain.StockItem, error) {
	if in.NewQty < 0 {
		return nil, domain.Validation("jumlah stok tidak valid", map[string]string{
			"new_qty": "harus 0 atau lebih",
		})
	}

	var item *domain.StockItem
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := s.products.GetByID(ctx, tx, in.ProductID); err != nil {
			return err
		}

		previous := 0
		avgCost := decimal.Zero
		if current, err := s.purchases.GetStock(ctx, tx, in.ProductID); err == nil {
			previous = current.QtyOnHand
			avgCost = current.AvgCostIDR
		} else if domainErr, ok := domain.AsError(err); !ok || domainErr.Code != domain.CodeNotFound {
			return err
		}

		delta := in.NewQty - previous
		if delta == 0 {
			var err error
			item, err = s.purchases.GetStock(ctx, tx, in.ProductID)
			return err
		}

		var err error
		item, err = s.purchases.AdjustStock(ctx, tx, in.ProductID, in.NewQty)
		if err != nil {
			return err
		}

		_, err = s.purchases.CreateMovement(ctx, tx, repository.StockMovementParams{
			ProductID:   in.ProductID,
			Type:        domain.StockAdjustment,
			Qty:         delta,
			UnitCostIDR: avgCost,
			Note:        in.Note,
			CreatedBy:   nullableUUID(actorID),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}
