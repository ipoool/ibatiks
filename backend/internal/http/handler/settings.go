package handler

import (
	"net/http"

	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type SettingsHandler struct {
	settings *service.SettingsService
	audit    *service.AuditService
}

func NewSettingsHandler(settings *service.SettingsService, audit *service.AuditService) *SettingsHandler {
	return &SettingsHandler{settings: settings, audit: audit}
}

func (h *SettingsHandler) List(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settings.List(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, settings)
}

type updateSettingsRequest struct {
	Settings map[string]string `json:"settings" validate:"required,min=1"`
}

func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateSettingsRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	settings, err := h.settings.Update(r.Context(), req.Settings, middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, settings)
}

// AuditLogs menampilkan jejak perubahan, dipakai untuk menelusuri siapa
// mengubah qty atau nominal sebuah order.
func (h *SettingsHandler) AuditLogs(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	entityID, err := request.UUIDQuery(r, "entity_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	userID, err := request.UUIDQuery(r, "user_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	logs, total, err := h.audit.List(r.Context(), p, r.URL.Query().Get("entity"), entityID, userID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, logs, p.Page, p.PerPage, total)
}

// AuditActors mengisi penyaring "akun" pada jejak perubahan.
//
// Dibuat terpisah dari daftar pengguna karena keduanya menjawab pertanyaan yang
// berbeda — dan dijaga hak akses yang berbeda pula: Jejak Perubahan ada di menu
// Pengaturan, sedangkan daftar pengguna menuntut menu Pengguna.
func (h *SettingsHandler) AuditActors(w http.ResponseWriter, r *http.Request) {
	actors, err := h.audit.Actors(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, actors)
}
