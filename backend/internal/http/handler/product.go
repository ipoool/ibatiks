package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type ProductHandler struct {
	products *service.ProductService
}

func NewProductHandler(products *service.ProductService) *ProductHandler {
	return &ProductHandler{products: products}
}

// --- Kategori --------------------------------------------------------------

type categoryRequest struct {
	Name        string  `json:"name"        validate:"required,min=2,max=80"`
	Description *string `json:"description"`
}

func (h *ProductHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	category, err := h.products.CreateCategory(r.Context(), req.Name, req.Description)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, category)
}

func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.products.ListCategories(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, categories)
}

func (h *ProductHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req categoryRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	category, err := h.products.UpdateCategory(r.Context(), id, req.Name, req.Description)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, category)
}

func (h *ProductHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.products.DeleteCategory(r.Context(), id); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}

// --- Produk ----------------------------------------------------------------

type productRequest struct {
	SKU          string          `json:"sku"`
	Name         string          `json:"name"          validate:"required,min=2,max=200"`
	CategoryID   *uuid.UUID      `json:"category_id"`
	Brand        *string         `json:"brand"`
	StoreName    *string         `json:"store_name"`
	BaseCurrency string          `json:"base_currency" validate:"omitempty,len=3"`
	BasePrice    decimal.Decimal `json:"base_price"`
	MarkupType   string          `json:"markup_type"   validate:"required,oneof=percent nominal"`
	MarkupValue  decimal.Decimal `json:"markup_value"`
	WeightGram   int             `json:"weight_gram"   validate:"gte=0"`
	ImageURL     *string         `json:"image_url"     validate:"omitempty,url"`
	Notes        *string         `json:"notes"`
	IsActive     *bool           `json:"is_active"`
}

func (req productRequest) toInput() service.ProductInput {
	// Produk baru dianggap aktif kecuali admin menyatakan sebaliknya.
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return service.ProductInput{
		SKU:          req.SKU,
		Name:         req.Name,
		CategoryID:   req.CategoryID,
		Brand:        req.Brand,
		StoreName:    req.StoreName,
		BaseCurrency: req.BaseCurrency,
		BasePrice:    req.BasePrice,
		MarkupType:   req.MarkupType,
		MarkupValue:  req.MarkupValue,
		WeightGram:   req.WeightGram,
		ImageURL:     req.ImageURL,
		Notes:        req.Notes,
		IsActive:     isActive,
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req productRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	product, err := h.products.Create(r.Context(), req.toInput())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, product)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	categoryID, err := request.UUIDQuery(r, "category_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	products, total, err := h.products.List(r.Context(), p, categoryID, request.BoolQuery(r, "active_only"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, products, p.Page, p.PerPage, total)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	product, err := h.products.Get(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, product)
}

// PriceHistory menampilkan harga produk ini dari trip ke trip.
func (h *ProductHandler) PriceHistory(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	history, err := h.products.PriceHistory(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, history)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req productRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	product, err := h.products.Update(r.Context(), id, req.toInput())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.products.Delete(r.Context(), id); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}

type pricePreviewRequest struct {
	CostPrice    decimal.Decimal `json:"cost_price"`
	ExchangeRate decimal.Decimal `json:"exchange_rate" validate:"required"`
	MarkupType   string          `json:"markup_type"   validate:"required,oneof=percent nominal"`
	MarkupValue  decimal.Decimal `json:"markup_value"`
}

// PreviewPrice memperlihatkan harga jual yang akan terbentuk sebelum admin
// menyimpan item katalog, supaya markup bisa dicoba-coba dulu.
func (h *ProductHandler) PreviewPrice(w http.ResponseWriter, r *http.Request) {
	var req pricePreviewRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	costIDR, sellPrice, err := h.products.PreviewPrice(
		req.CostPrice, req.ExchangeRate, req.MarkupType, req.MarkupValue)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, map[string]any{
		"cost_price_idr": costIDR,
		"sell_price":     sellPrice,
		"profit_per_pcs": sellPrice.Sub(costIDR),
	})
}
