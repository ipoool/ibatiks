package handler

import (
	"net/http"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type CustomerHandler struct {
	customers *service.CustomerService
}

func NewCustomerHandler(customers *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customers: customers}
}

type socialRequest struct {
	Platform string `json:"platform" validate:"required,oneof=instagram tiktok facebook lainnya"`
	Handle   string `json:"handle"`
}

// socials memindahkan bentuk permintaan ke bentuk domain. Barisnya tidak
// disaring di sini — service yang membuang yang kosong, supaya aturannya satu
// tempat dan berlaku dari mana pun customer disimpan.
func socials(list []socialRequest) []domain.Social {
	hasil := make([]domain.Social, 0, len(list))
	for _, akun := range list {
		hasil = append(hasil, domain.Social{Platform: akun.Platform, Handle: akun.Handle})
	}
	return hasil
}

type customerRequest struct {
	Name        string          `json:"name"        validate:"required,min=2,max=120"`
	PhoneWA     string          `json:"phone_wa"    validate:"required,min=8,max=20"`
	Email       *string         `json:"email"       validate:"omitempty,email"`
	Socials     []socialRequest `json:"socials" validate:"omitempty,dive"`
	Address     *string         `json:"address"`
	City        *string         `json:"city"`
	District    *string         `json:"district"`
	Subdistrict *string         `json:"subdistrict"`
	Province    *string         `json:"province"`
	PostalCode  *string         `json:"postal_code"`
	Notes       *string         `json:"notes"`
}

func (req customerRequest) toInput() service.CustomerInput {
	return service.CustomerInput{
		Name:        req.Name,
		PhoneWA:     req.PhoneWA,
		Email:       req.Email,
		Socials:     socials(req.Socials),
		Address:     req.Address,
		City:        req.City,
		District:    req.District,
		Subdistrict: req.Subdistrict,
		Province:    req.Province,
		PostalCode:  req.PostalCode,
		Notes:       req.Notes,
	}
}

func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req customerRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	customer, err := h.customers.Create(r.Context(), req.toInput())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, customer)
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	customers, total, err := h.customers.List(r.Context(), p)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, customers, p.Page, p.PerPage, total)
}

func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	customer, err := h.customers.Get(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, customer)
}

// Stats menampilkan ringkasan riwayat belanja customer pada halaman detailnya.
func (h *CustomerHandler) Stats(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	stats, err := h.customers.Stats(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, stats)
}

func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req customerRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	customer, err := h.customers.Update(r.Context(), id, req.toInput())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, customer)
}

func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.customers.Delete(r.Context(), id); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}
