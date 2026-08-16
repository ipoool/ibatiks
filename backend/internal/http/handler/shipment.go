package handler

import (
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type ShipmentHandler struct {
	shipments *service.ShipmentService
}

func NewShipmentHandler(shipments *service.ShipmentService) *ShipmentHandler {
	return &ShipmentHandler{shipments: shipments}
}

// DeliveryNote mengirimkan surat jalan sebagai PDF siap cetak.
func (h *ShipmentHandler) DeliveryNote(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	content, name, err := h.shipments.DeliveryNote(r.Context(), orderID)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	// inline supaya langsung tampil di tab browser dan bisa dicetak dari sana.
	w.Header().Set("Content-Disposition", `inline; filename="`+name+`.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

type packRequest struct {
	Courier    string  `json:"courier"     validate:"omitempty,max=30"`
	Service    string  `json:"service"     validate:"omitempty,max=20"`
	WeightGram int     `json:"weight_gram" validate:"gte=0"`
	LengthCM   int     `json:"length_cm"   validate:"gte=0"`
	WidthCM    int     `json:"width_cm"    validate:"gte=0"`
	HeightCM   int     `json:"height_cm"   validate:"gte=0"`
	Notes      *string `json:"notes"`
}

// Pack menandai order sudah dikemas atas nama customer.
func (h *ShipmentHandler) Pack(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req packRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	shipment, err := h.shipments.Pack(r.Context(), orderID, service.PackInput{
		Courier:    req.Courier,
		Service:    req.Service,
		WeightGram: req.WeightGram,
		LengthCM:   req.LengthCM,
		WidthCM:    req.WidthCM,
		HeightCM:   req.HeightCM,
		Notes:      req.Notes,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, shipment)
}

type shipRequest struct {
	TrackingNumber string          `json:"tracking_number" validate:"required,min=6,max=40"`
	ShippingCost   decimal.Decimal `json:"shipping_cost"`
	ShippedAt      request.Date    `json:"shipped_at"`
	AllowUnpaid    bool            `json:"allow_unpaid"`
}

// Ship mencatat nomor resi JNE dan menandai paket sudah diserahkan ke kurir.
func (h *ShipmentHandler) Ship(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req shipRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	shipment, err := h.shipments.Ship(r.Context(), orderID, service.ShipInput{
		TrackingNumber: req.TrackingNumber,
		ShippingCost:   req.ShippingCost,
		ShippedAt:      req.ShippedAt.Ptr(),
		AllowUnpaid:    req.AllowUnpaid,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, shipment)
}

func (h *ShipmentHandler) MarkDelivered(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	shipment, err := h.shipments.MarkDelivered(r.Context(), orderID, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, shipment)
}

func (h *ShipmentHandler) GetByOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	shipment, err := h.shipments.GetByOrder(r.Context(), orderID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, shipment)
}

// Message menyiapkan pesan pemberitahuan pengiriman berisi nomor resi.
func (h *ShipmentHandler) Message(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	message, err := h.shipments.Message(r.Context(), orderID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, message)
}

func (h *ShipmentHandler) MarkNotified(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	shipment, err := h.shipments.MarkNotified(r.Context(), orderID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, shipment)
}

type updateShipmentRequest struct {
	Courier        string          `json:"courier"         validate:"omitempty,max=30"`
	Service        string          `json:"service"         validate:"omitempty,max=20"`
	WeightGram     int             `json:"weight_gram"     validate:"gte=0"`
	ShippingCost   decimal.Decimal `json:"shipping_cost"`
	TrackingNumber *string         `json:"tracking_number"`
	Notes          *string         `json:"notes"`
}

func (h *ShipmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req updateShipmentRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	shipment, err := h.shipments.Update(r.Context(), id, service.UpdateShipmentInput{
		Courier:        req.Courier,
		Service:        req.Service,
		WeightGram:     req.WeightGram,
		ShippingCost:   req.ShippingCost,
		TrackingNumber: req.TrackingNumber,
		Notes:          req.Notes,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, shipment)
}

func (h *ShipmentHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	tripID, err := request.UUIDQuery(r, "trip_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	shipments, total, err := h.shipments.List(r.Context(), p, r.URL.Query().Get("status"), tripID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, shipments, p.Page, p.PerPage, total)
}
