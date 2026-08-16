// Package handler berisi handler HTTP. Setiap handler hanya melakukan tiga hal:
// membaca request, memanggil service, dan menulis response. Aturan bisnis
// sepenuhnya ada di layer service.
package handler

import (
	"net/http"
	"strings"

	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	session, err := h.auth.Login(r.Context(), req.Email, req.Password, r.UserAgent(), clientIP(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	withEffectivePermissions(session.User)
	response.OK(w, session)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	session, err := h.auth.Refresh(r.Context(), req.RefreshToken, r.UserAgent(), clientIP(r))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	withEffectivePermissions(session.User)
	response.OK(w, session)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, map[string]string{"message": "berhasil keluar"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.auth.Me(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, withEffectivePermissions(user))
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8,max=72"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := request.DecodeJSON(w, r, &req); err != nil {
		response.Error(w, r, err)
		return
	}

	if err := h.auth.ChangePassword(r.Context(), middleware.UserIDFrom(r.Context()),
		req.CurrentPassword, req.NewPassword); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, map[string]string{"message": "password berhasil diganti, silakan login ulang"})
}

// clientIP mengambil IP asli klien. Di belakang reverse proxy, IP sebenarnya
// ada di header X-Forwarded-For, bukan di RemoteAddr.
func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Entri pertama adalah klien asli; sisanya rantai proxy.
		if first, _, found := strings.Cut(forwarded, ","); found {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(forwarded)
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	host, _, found := strings.Cut(r.RemoteAddr, ":")
	if !found {
		return r.RemoteAddr
	}
	return host
}
