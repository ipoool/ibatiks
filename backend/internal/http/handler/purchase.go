package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type PurchaseHandler struct {
	purchases *service.PurchaseService
}

func NewPurchaseHandler(purchases *service.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{purchases: purchases}
}

// ShoppingList adalah daftar belanja yang dibuka tripper di negara tujuan.
func (h *PurchaseHandler) ShoppingList(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	list, err := h.purchases.ShoppingList(r.Context(), tripID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, list)
}

type purchaseRequest struct {
	ProductID       uuid.UUID        `json:"product_id"        validate:"required"`
	PurchaseDate    request.Date     `json:"purchase_date"`
	Qty             int              `json:"qty"               validate:"required,gt=0"`
	UnitCostForeign decimal.Decimal  `json:"unit_cost_foreign"`
	ExchangeRate    *decimal.Decimal `json:"exchange_rate"`
	StoreName       *string          `json:"store_name"`
	ReceiptURL      *string          `json:"receipt_url"       validate:"omitempty,url"`
	Notes           *string          `json:"notes"`
}

// Record mencatat belanja tripper dan langsung mengalokasikannya ke pesanan
// serta stok.
func (h *PurchaseHandler) Record(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req purchaseRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	result, err := h.purchases.Record(r.Context(), tripID, service.PurchaseInput{
		ProductID:       req.ProductID,
		PurchaseDate:    req.PurchaseDate.OrNow(),
		Qty:             req.Qty,
		UnitCostForeign: req.UnitCostForeign,
		ExchangeRate:    req.ExchangeRate,
		StoreName:       req.StoreName,
		ReceiptURL:      req.ReceiptURL,
		Notes:           req.Notes,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, result)
}

func (h *PurchaseHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	tripID, err := request.UUIDQuery(r, "trip_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	productID, err := request.UUIDQuery(r, "product_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	purchases, total, err := h.purchases.List(r.Context(), p, tripID, productID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, purchases, p.Page, p.PerPage, total)
}

func (h *PurchaseHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	purchase, err := h.purchases.Get(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, purchase)
}

// ListAllocations memperlihatkan ke mana tiap unit hasil belanja pergi.
func (h *PurchaseHandler) ListAllocations(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	allocations, err := h.purchases.ListAllocations(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, allocations)
}

func (h *PurchaseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.purchases.Delete(r.Context(), id, middleware.UserIDFrom(r.Context())); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}

// --- Stok ------------------------------------------------------------------

type StockHandler struct {
	purchases *service.PurchaseService
}

func NewStockHandler(purchases *service.PurchaseService) *StockHandler {
	return &StockHandler{purchases: purchases}
}

func (h *StockHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	items, total, err := h.purchases.ListStock(r.Context(), p, request.BoolQuery(r, "in_stock_only"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, items, p.Page, p.PerPage, total)
}

func (h *StockHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	productID, err := request.UUIDQuery(r, "product_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	movements, total, err := h.purchases.ListMovements(r.Context(), p, productID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, movements, p.Page, p.PerPage, total)
}

type stockSaleRequest struct {
	ProductID uuid.UUID       `json:"product_id" validate:"required"`
	Qty       int             `json:"qty"        validate:"required,gt=0"`
	SalePrice decimal.Decimal `json:"sale_price"`
	Channel   string          `json:"channel"    validate:"omitempty,max=60"`
	Note      *string         `json:"note"`
}

// Sell mencatat penjualan stok di marketplace.
func (h *StockHandler) Sell(w http.ResponseWriter, r *http.Request) {
	var req stockSaleRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	movement, err := h.purchases.SellFromStock(r.Context(), service.StockSaleInput{
		ProductID: req.ProductID,
		Qty:       req.Qty,
		SalePrice: req.SalePrice,
		Channel:   req.Channel,
		Note:      req.Note,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, movement)
}

type stockAdjustRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	NewQty    int       `json:"new_qty"    validate:"gte=0"`
	Note      *string   `json:"note"`
}

// Adjust menyetel stok mengikuti hasil stock opname.
func (h *StockHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	var req stockAdjustRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	item, err := h.purchases.AdjustStock(r.Context(), service.StockAdjustInput{
		ProductID: req.ProductID,
		NewQty:    req.NewQty,
		Note:      req.Note,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, item)
}
