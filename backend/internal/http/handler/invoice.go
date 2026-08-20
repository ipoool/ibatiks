package handler

import (
	"net/http"
	"path/filepath"

	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type InvoiceHandler struct {
	invoices *service.InvoiceService
}

func NewInvoiceHandler(invoices *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoices: invoices}
}

type createInvoiceRequest struct {
	Type    string       `json:"type"     validate:"required,oneof=dp final"`
	DueDate request.Date `json:"due_date"`
	Notes   *string      `json:"notes"`
}

func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req createInvoiceRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	invoice, err := h.invoices.Create(r.Context(), orderID, service.CreateInvoiceInput{
		Type:    req.Type,
		DueDate: req.DueDate.Ptr(),
		Notes:   req.Notes,
	}, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, invoice)
}

func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	from, err := request.DateQuery(r, "from")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	to, err := request.DateQuery(r, "to")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	invoices, total, err := h.invoices.List(r.Context(), p,
		r.URL.Query().Get("status"), r.URL.Query().Get("type"), from, to)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, invoices, p.Page, p.PerPage, total)
}

// Candidates mendaftar order yang siap ditagih pelunasannya, untuk dialog Buat
// Invoice.
func (h *InvoiceHandler) Candidates(w http.ResponseWriter, r *http.Request) {
	candidates, err := h.invoices.Candidates(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, candidates)
}

func (h *InvoiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	invoice, err := h.invoices.Get(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, invoice)
}

func (h *InvoiceHandler) ListByOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	invoices, err := h.invoices.ListByOrder(r.Context(), orderID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, invoices)
}

// PDF mengirim berkas invoice untuk diunduh atau dilihat admin.
func (h *InvoiceHandler) PDF(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	path, number, err := h.invoices.PDFPath(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	// inline supaya PDF langsung tampil di tab browser; admin tetap bisa
	// menyimpannya dari viewer bawaan.
	w.Header().Set("Content-Disposition", `inline; filename="`+number+`.pdf"`)
	http.ServeFile(w, r, filepath.Clean(path))
}

// Message menyiapkan teks penagihan beserta tautan wa.me siap klik.
func (h *InvoiceHandler) Message(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	message, err := h.invoices.Message(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, message)
}

type markSentRequest struct {
	Channel string `json:"channel" validate:"required,oneof=wa email manual"`
}

func (h *InvoiceHandler) MarkSent(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req markSentRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	invoice, err := h.invoices.MarkSent(r.Context(), id, req.Channel, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, invoice)
}

func (h *InvoiceHandler) Void(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	invoice, err := h.invoices.Void(r.Context(), id, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, invoice)
}

// DPMessage menyiapkan pesan permintaan uang muka untuk sebuah order.
func (h *InvoiceHandler) DPMessage(w http.ResponseWriter, r *http.Request) {
	orderID, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	message, err := h.invoices.DPMessage(r.Context(), orderID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, message)
}
