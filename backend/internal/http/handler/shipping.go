package handler

import (
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type ShippingHandler struct {
	shipping *service.ShippingService
}

func NewShippingHandler(shipping *service.ShippingService) *ShippingHandler {
	return &ShippingHandler{shipping: shipping}
}

type estimateRequest struct {
	Courier string `json:"courier"     validate:"omitempty,max=30"`
	Service string `json:"service"     validate:"omitempty,max=20"`
	City    string `json:"city"`
	// Bagian alamat yang lebih rinci membuat layanan kurir menunjuk kecamatan
	// yang tepat, bukan sekadar kotanya.
	District    string `json:"district"    validate:"omitempty,max=100"`
	Subdistrict string `json:"subdistrict" validate:"omitempty,max=100"`
	PostalCode  string `json:"postal_code" validate:"omitempty,max=10"`
	WeightGram  int    `json:"weight_gram" validate:"gte=0"`
	LengthCM    int    `json:"length_cm"   validate:"gte=0"`
	WidthCM     int    `json:"width_cm"    validate:"gte=0"`
	HeightCM    int    `json:"height_cm"   validate:"gte=0"`
}

// Estimate menghitung perkiraan ongkir dari berat dan dimensi paket.
func (h *ShippingHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	var req estimateRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	estimate, err := h.shipping.Estimate(r.Context(), service.EstimateInput{
		Courier:     req.Courier,
		Service:     req.Service,
		City:        req.City,
		District:    req.District,
		Subdistrict: req.Subdistrict,
		PostalCode:  req.PostalCode,
		WeightGram:  req.WeightGram,
		LengthCM:    req.LengthCM,
		WidthCM:     req.WidthCM,
		HeightCM:    req.HeightCM,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, estimate)
}

// Provider menjelaskan layanan tarif yang sedang aktif kepada menu Pengaturan:
// terhubung atau tidak, kota asalnya apa, dan kurir mana saja yang ditanyakan.
func (h *ShippingHandler) Provider(w http.ResponseWriter, r *http.Request) {
	info, err := h.shipping.ProviderInfo(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, info)
}

// SearchDestinations mencari kota asal di daftar tujuan milik layanan kurir.
func (h *ShippingHandler) SearchDestinations(w http.ResponseWriter, r *http.Request) {
	tujuan, err := h.shipping.SearchDestinations(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, tujuan)
}

// EstimateForOrder memakai kota tujuan dari alamat order, sehingga admin tidak
// perlu mengetik ulang kotanya saat mengemas.
func (h *ShippingHandler) EstimateForOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req estimateRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	estimate, err := h.shipping.EstimateForOrder(r.Context(), orderID, service.EstimateInput{
		Courier:     req.Courier,
		Service:     req.Service,
		City:        req.City,
		District:    req.District,
		Subdistrict: req.Subdistrict,
		PostalCode:  req.PostalCode,
		WeightGram:  req.WeightGram,
		LengthCM:    req.LengthCM,
		WidthCM:     req.WidthCM,
		HeightCM:    req.HeightCM,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, estimate)
}

func (h *ShippingHandler) ListRates(w http.ResponseWriter, r *http.Request) {
	rates, err := h.shipping.ListRates(r.Context(),
		r.URL.Query().Get("courier"), r.URL.Query().Get("q"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, rates)
}

type rateRequest struct {
	Courier         string          `json:"courier"          validate:"omitempty,max=30"`
	Service         string          `json:"service"          validate:"omitempty,max=20"`
	DestinationCity string          `json:"destination_city" validate:"required,max=80"`
	Province        *string         `json:"province"`
	PricePerKg      decimal.Decimal `json:"price_per_kg"     validate:"required"`
	MinWeightGram   int             `json:"min_weight_gram"  validate:"gte=0"`
	ETD             *string         `json:"etd"`
}

func (h *ShippingHandler) SaveRate(w http.ResponseWriter, r *http.Request) {
	var req rateRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	rate, err := h.shipping.SaveRate(r.Context(), service.RateInput{
		Courier:         req.Courier,
		Service:         req.Service,
		DestinationCity: req.DestinationCity,
		Province:        req.Province,
		PricePerKg:      req.PricePerKg,
		MinWeightGram:   req.MinWeightGram,
		ETD:             req.ETD,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, rate)
}

func (h *ShippingHandler) DeleteRate(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.shipping.DeleteRate(r.Context(), id); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}
