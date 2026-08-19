package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type OrderHandler struct {
	orders *service.OrderService
}

func NewOrderHandler(orders *service.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

type orderItemRequest struct {
	ProductID uuid.UUID        `json:"product_id" validate:"required"`
	Qty       int              `json:"qty"        validate:"required,gt=0"`
	UnitPrice *decimal.Decimal `json:"unit_price"`
	Notes     *string          `json:"notes"`
}

type createOrderRequest struct {
	TripID      uuid.UUID          `json:"trip_id"      validate:"required"`
	CustomerID  uuid.UUID          `json:"customer_id"  validate:"required"`
	OrderDate   request.Date       `json:"order_date"`
	OrderSource string             `json:"order_source" validate:"omitempty,oneof=whatsapp instagram tiktok marketplace lainnya"`
	Items       []orderItemRequest `json:"items"        validate:"required,min=1,dive"`
	Discount    decimal.Decimal    `json:"discount"`
	DPRequired  *decimal.Decimal   `json:"dp_required"`

	RecipientName       *string `json:"recipient_name"`
	RecipientPhone      *string `json:"recipient_phone"`
	ShippingAddress     *string `json:"shipping_address"`
	ShippingCity        *string `json:"shipping_city"`
	ShippingDistrict    *string `json:"shipping_district"`
	ShippingSubdistrict *string `json:"shipping_subdistrict"`
	ShippingProvince    *string `json:"shipping_province"`
	ShippingPostalCode  *string `json:"shipping_postal_code"`
	Notes               *string `json:"notes"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	items := make([]service.OrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.OrderItemInput{
			ProductID: item.ProductID,
			Qty:       item.Qty,
			UnitPrice: item.UnitPrice,
			Notes:     item.Notes,
		})
	}

	order, err := h.orders.Create(r.Context(), service.CreateOrderInput{
		TripID:              req.TripID,
		CustomerID:          req.CustomerID,
		OrderDate:           req.OrderDate.OrNow(),
		OrderSource:         req.OrderSource,
		Items:               items,
		Discount:            req.Discount,
		DPRequired:          req.DPRequired,
		RecipientName:       req.RecipientName,
		RecipientPhone:      req.RecipientPhone,
		ShippingAddress:     req.ShippingAddress,
		ShippingCity:        req.ShippingCity,
		ShippingDistrict:    req.ShippingDistrict,
		ShippingSubdistrict: req.ShippingSubdistrict,
		ShippingProvince:    req.ShippingProvince,
		ShippingPostalCode:  req.ShippingPostalCode,
		Notes:               req.Notes,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, order)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	tripID, err := request.UUIDQuery(r, "trip_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	customerID, err := request.UUIDQuery(r, "customer_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	orders, total, err := h.orders.List(r.Context(), p, repository.OrderFilter{
		TripID:      tripID,
		CustomerID:  customerID,
		Status:      r.URL.Query().Get("status"),
		Source:      r.URL.Query().Get("source"),
		UnpaidOnly:  request.BoolQuery(r, "unpaid_only"),
		ReadyToShip: request.BoolQuery(r, "ready_to_ship"),
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, orders, p.Page, p.PerPage, total)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	order, err := h.orders.Get(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order)
}

type updateOrderRequest struct {
	OrderDate   request.Date     `json:"order_date"`
	OrderSource string           `json:"order_source" validate:"omitempty,oneof=whatsapp instagram tiktok marketplace lainnya"`
	Discount    decimal.Decimal  `json:"discount"`
	ShippingFee *decimal.Decimal `json:"shipping_fee"`
	DPRequired  *decimal.Decimal `json:"dp_required"`

	RecipientName       string  `json:"recipient_name"   validate:"required"`
	RecipientPhone      string  `json:"recipient_phone"  validate:"required"`
	ShippingAddress     string  `json:"shipping_address" validate:"required"`
	ShippingCity        string  `json:"shipping_city"    validate:"required"`
	ShippingDistrict    *string `json:"shipping_district"`
	ShippingSubdistrict *string `json:"shipping_subdistrict"`
	ShippingProvince    *string `json:"shipping_province"`
	ShippingPostalCode  *string `json:"shipping_postal_code"`
	Notes               *string `json:"notes"`
}

func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req updateOrderRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	order, err := h.orders.Update(r.Context(), id, service.UpdateOrderInput{
		OrderDate:           req.OrderDate.OrNow(),
		OrderSource:         req.OrderSource,
		Discount:            req.Discount,
		DPRequired:          req.DPRequired,
		RecipientName:       req.RecipientName,
		RecipientPhone:      req.RecipientPhone,
		ShippingAddress:     req.ShippingAddress,
		ShippingCity:        req.ShippingCity,
		ShippingDistrict:    req.ShippingDistrict,
		ShippingSubdistrict: req.ShippingSubdistrict,
		ShippingProvince:    req.ShippingProvince,
		ShippingPostalCode:  req.ShippingPostalCode,
		Notes:               req.Notes,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order)
}

func (h *OrderHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req orderItemRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	order, err := h.orders.AddItem(r.Context(), orderID, service.OrderItemInput{
		ProductID: req.ProductID,
		Qty:       req.Qty,
		UnitPrice: req.UnitPrice,
		Notes:     req.Notes,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order)
}

type updateOrderItemRequest struct {
	Qty       int              `json:"qty"        validate:"required,gt=0"`
	UnitPrice *decimal.Decimal `json:"unit_price"`
	Notes     *string          `json:"notes"`
}

// UpdateItem adalah endpoint yang dipakai admin saat customer mengubah jumlah
// pesanan — kasus paling sering pada operasional harian.
func (h *OrderHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	itemID, err := request.UUIDParam(r, "itemId")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req updateOrderItemRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	order, err := h.orders.UpdateItem(r.Context(), orderID, itemID,
		req.Qty, req.UnitPrice, req.Notes, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order)
}

func (h *OrderHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	itemID, err := request.UUIDParam(r, "itemId")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	order, err := h.orders.DeleteItem(r.Context(), orderID, itemID, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order)
}

func (h *OrderHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
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

	order, err := h.orders.ChangeStatus(r.Context(), id, req.Status, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order)
}

type cancelOrderRequest struct {
	Reason *string `json:"reason"`
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req cancelOrderRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	order, err := h.orders.Cancel(r.Context(), id, req.Reason, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order)
}

type paymentRequest struct {
	Type      string          `json:"type"      validate:"required,oneof=dp settlement refund adjustment"`
	Amount    decimal.Decimal `json:"amount"    validate:"required"`
	Method    string          `json:"method"    validate:"required,oneof=transfer cash qris ewallet lainnya"`
	Reference *string         `json:"reference"`
	ProofURL  *string         `json:"proof_url" validate:"omitempty,url"`
	PaidAt    request.Date    `json:"paid_at"`
	Notes     *string         `json:"notes"`
}

func (h *OrderHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req paymentRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	order, err := h.orders.RecordPayment(r.Context(), orderID, service.PaymentInput{
		Type:      req.Type,
		Amount:    req.Amount,
		Method:    req.Method,
		Reference: req.Reference,
		ProofURL:  req.ProofURL,
		PaidAt:    req.PaidAt.OrNow(),
		Notes:     req.Notes,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, order)
}

func (h *OrderHandler) DeletePayment(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	paymentID, err := request.UUIDParam(r, "paymentId")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	order, err := h.orders.DeletePayment(r.Context(), orderID, paymentID, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, order)
}
