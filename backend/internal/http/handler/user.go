package handler

import (
	"net/http"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

type createUserRequest struct {
	Name        string   `json:"name"     validate:"required,min=2,max=100"`
	Email       string   `json:"email"    validate:"required,email"`
	Password    string   `json:"password" validate:"required,min=8,max=72"`
	Role        string   `json:"role"        validate:"required,oneof=owner admin tripper"`
	Phone       *string  `json:"phone"`
	Permissions []string `json:"permissions"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	user, err := h.users.Create(r.Context(), service.CreateUserInput{
		Name:        req.Name,
		Email:       req.Email,
		Password:    req.Password,
		Role:        req.Role,
		Phone:       req.Phone,
		Permissions: req.Permissions,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, withEffectivePermissions(user))
}

// withEffectivePermissions mengisi hak akses hasil gabungan sebelum data
// dikirim ke antarmuka.
//
// Aturan penggabungannya tinggal di domain, dan antarmuka cukup membaca
// hasilnya — kalau tidak, aturan yang sama harus ditulis ulang di frontend dan
// cepat atau lambat keduanya berbeda.
func withEffectivePermissions(user *domain.User) *domain.User {
	if user == nil {
		return nil
	}
	user.EffectivePermissions = domain.EffectivePermissions(user.Role, user.Permissions)
	return user
}

func withEffectivePermissionsAll(users []domain.User) []domain.User {
	for i := range users {
		users[i].EffectivePermissions = domain.EffectivePermissions(users[i].Role, users[i].Permissions)
	}
	return users
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	users, total, err := h.users.List(r.Context(), p)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Paginated(w, withEffectivePermissionsAll(users), p.Page, p.PerPage, total)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	user, err := h.users.Get(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, withEffectivePermissions(user))
}

type updateUserRequest struct {
	Name        string   `json:"name"      validate:"required,min=2,max=100"`
	Role        string   `json:"role"        validate:"required,oneof=owner admin tripper"`
	Phone       *string  `json:"phone"`
	IsActive    bool     `json:"is_active"`
	Permissions []string `json:"permissions"`
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req updateUserRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	user, err := h.users.Update(r.Context(), id, service.UpdateUserInput{
		Name:        req.Name,
		Role:        req.Role,
		Phone:       req.Phone,
		IsActive:    req.IsActive,
		Permissions: req.Permissions,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, withEffectivePermissions(user))
}

type resetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req resetPasswordRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.users.ResetPassword(r.Context(), id, req.NewPassword); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, map[string]string{"message": "password berhasil direset"})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := request.UUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.users.Delete(r.Context(), id, middleware.UserIDFrom(r.Context())); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}
