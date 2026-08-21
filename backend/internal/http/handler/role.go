package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type RoleHandler struct {
	roles *service.RoleService
}

// pemintaRole adalah role pengguna yang mengirim permintaan. Dipakai untuk
// menyembunyikan role dan akun root dari siapa pun selain root sendiri.
func pemintaRole(r *http.Request) string {
	user, _ := middleware.UserFrom(r.Context())
	return user.Role
}

func NewRoleHandler(roles *service.RoleService) *RoleHandler {
	return &RoleHandler{roles: roles}
}

// roleOptions ikut dikirim bersama daftar role supaya antarmuka tidak perlu
// menyalin daftar menu aplikasi sendiri lalu ikut ketinggalan saat ada menu
// baru.
type roleOptions struct {
	Permissions []string `json:"permissions"`
	Scopes      []string `json:"scopes"`
	// FieldPermissions adalah bagian daftar di atas yang masih masuk akal untuk
	// petugas lapangan. Dikirim dari sini supaya antarmuka bisa meredupkan
	// centang yang tidak akan berlaku, tanpa menyalin aturannya sendiri.
	FieldPermissions []string `json:"field_permissions"`
}

func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roles.List(r.Context(), pemintaRole(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, map[string]any{
		"roles": roles,
		"options": roleOptions{
			Permissions:      domain.AllPermissions,
			Scopes:           domain.AllScopes,
			FieldPermissions: domain.FieldPermissions,
		},
	})
}

func (h *RoleHandler) Get(w http.ResponseWriter, r *http.Request) {
	role, err := h.roles.Get(r.Context(), chi.URLParam(r, "name"), pemintaRole(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, role)
}

type saveRoleRequest struct {
	Label       string   `json:"label"       validate:"required,min=2,max=40"`
	Description string   `json:"description" validate:"max=200"`
	Scope       string   `json:"scope"       validate:"omitempty,oneof=full field"`
	Permissions []string `json:"permissions" validate:"required,min=1"`
}

func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req saveRoleRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	role, err := h.roles.Create(r.Context(), service.SaveRoleInput{
		Label:       req.Label,
		Description: req.Description,
		Scope:       req.Scope,
		Permissions: req.Permissions,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, role)
}

func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req saveRoleRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	role, err := h.roles.Update(r.Context(), chi.URLParam(r, "name"), pemintaRole(r), service.SaveRoleInput{
		Label:       req.Label,
		Description: req.Description,
		Scope:       req.Scope,
		Permissions: req.Permissions,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, role)
}

func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.roles.Delete(r.Context(), chi.URLParam(r, "name"), pemintaRole(r)); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}
