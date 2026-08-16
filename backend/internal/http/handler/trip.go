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

type TripHandler struct {
	trips *service.TripService
}

func NewTripHandler(trips *service.TripService) *TripHandler {
	return &TripHandler{trips: trips}
}

type tripRequest struct {
	Title         string          `json:"title"           validate:"required,min=3,max=150"`
	Country       string          `json:"country"         validate:"required,min=2,max=80"`
	City          *string         `json:"city"`
	TripperUserID *uuid.UUID      `json:"tripper_user_id"`
	DepartDate    request.Date    `json:"depart_date"     validate:"required"`
	ReturnDate    request.Date    `json:"return_date"     validate:"required"`
	OrderDeadline request.Date    `json:"order_deadline"`
	Currency      string          `json:"currency"        validate:"required,len=3"`
	ExchangeRate  decimal.Decimal `json:"exchange_rate"   validate:"required"`
	Notes         *string         `json:"notes"`
}

func (req tripRequest) toInput() service.TripInput {
	return service.TripInput{
		Title:         req.Title,
		Country:       req.Country,
		City:          req.City,
		TripperUserID: req.TripperUserID,
		DepartDate:    req.DepartDate.Time,
		ReturnDate:    req.ReturnDate.Time,
		OrderDeadline: req.OrderDeadline.Ptr(),
		Currency:      req.Currency,
		ExchangeRate:  req.ExchangeRate,
		Notes:         req.Notes,
	}
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req tripRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	trip, err := h.trips.Create(r.Context(), req.toInput(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, trip)
}

func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	trips, total, err := h.trips.List(r.Context(), p, r.URL.Query().Get("status"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, trips, p.Page, p.PerPage, total)
}

func (h *TripHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	trip, err := h.trips.Get(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, trip)
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req tripRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	trip, err := h.trips.Update(r.Context(), id, req.toInput())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, trip)
}

type statusRequest struct {
	Status string `json:"status" validate:"required"`
}

func (h *TripHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req statusRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	trip, err := h.trips.ChangeStatus(r.Context(), id, req.Status, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, trip)
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.trips.Delete(r.Context(), id); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}

// --- Katalog trip ----------------------------------------------------------

type tripItemRequest struct {
	ProductID   uuid.UUID       `json:"product_id"   validate:"required"`
	CostPrice   decimal.Decimal `json:"cost_price"`
	MarkupType  string          `json:"markup_type"  validate:"omitempty,oneof=percent nominal"`
	MarkupValue decimal.Decimal `json:"markup_value"`
	MaxQty      *int            `json:"max_qty"      validate:"omitempty,gt=0"`
	IsActive    *bool           `json:"is_active"`
	Notes       *string         `json:"notes"`
}

func (req tripItemRequest) toInput() service.TripItemInput {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return service.TripItemInput{
		ProductID:   req.ProductID,
		CostPrice:   req.CostPrice,
		MarkupType:  req.MarkupType,
		MarkupValue: req.MarkupValue,
		MaxQty:      req.MaxQty,
		IsActive:    isActive,
		Notes:       req.Notes,
	}
}

func (h *TripHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req tripItemRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	item, err := h.trips.AddItem(r.Context(), tripID, req.toInput())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, item)
}

func (h *TripHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	items, err := h.trips.ListItems(r.Context(), tripID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, items)
}

func (h *TripHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	itemID, err := request.UUIDParam(r, "itemId")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req tripItemRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	item, err := h.trips.UpdateItem(r.Context(), tripID, itemID, req.toInput())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, item)
}

func (h *TripHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	itemID, err := request.UUIDParam(r, "itemId")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.trips.DeleteItem(r.Context(), tripID, itemID); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}

// RecalculatePrices menghitung ulang seluruh harga katalog dengan kurs terkini.
func (h *TripHandler) RecalculatePrices(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	items, err := h.trips.RecalculatePrices(r.Context(), tripID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, items)
}

// --- Biaya perjalanan ------------------------------------------------------

type tripExpenseRequest struct {
	Category    string          `json:"category"    validate:"required,oneof=tiket bagasi akomodasi transport visa lainnya"`
	Description string          `json:"description" validate:"required,min=2,max=200"`
	Amount      decimal.Decimal `json:"amount"      validate:"required"`
	SpentAt     request.Date    `json:"spent_at"`
	ReceiptURL  *string         `json:"receipt_url" validate:"omitempty,url"`
}

func (req tripExpenseRequest) toInput() service.TripExpenseInput {
	return service.TripExpenseInput{
		Category:    req.Category,
		Description: req.Description,
		Amount:      req.Amount,
		SpentAt:     req.SpentAt.OrNow(),
		ReceiptURL:  req.ReceiptURL,
	}
}

func (h *TripHandler) AddExpense(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req tripExpenseRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	expense, err := h.trips.AddExpense(r.Context(), tripID, req.toInput(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, expense)
}

func (h *TripHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	expenses, err := h.trips.ListExpenses(r.Context(), tripID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, expenses)
}

func (h *TripHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	expenseID, err := request.UUIDParam(r, "expenseId")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req tripExpenseRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	expense, err := h.trips.UpdateExpense(r.Context(), expenseID, req.toInput())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, expense)
}

func (h *TripHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	expenseID, err := request.UUIDParam(r, "expenseId")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.trips.DeleteExpense(r.Context(), expenseID); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}
