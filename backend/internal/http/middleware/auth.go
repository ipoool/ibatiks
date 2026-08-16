// Package middleware berisi middleware HTTP: autentikasi, otorisasi, logging,
// dan pemulihan dari panic.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/token"
)

type ctxKey int

const userCtxKey ctxKey = iota

// AuthUser adalah identitas pengguna yang sudah terverifikasi dari access token.
type AuthUser struct {
	ID    uuid.UUID
	Email string
	Role  string
}

// Authenticate memvalidasi header Authorization: Bearer <token> dan menaruh
// identitas pengguna di context request.
func Authenticate(tm *token.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := bearerToken(r)
			if raw == "" {
				response.Error(w, r, domain.Unauthorized("token akses tidak ditemukan"))
				return
			}

			claims, err := tm.ParseAccessToken(raw)
			if err != nil {
				response.Error(w, r, domain.Unauthorized("sesi tidak valid, silakan login ulang"))
				return
			}

			user := AuthUser{ID: claims.UserID, Email: claims.Email, Role: claims.Role}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userCtxKey, user)))
		})
	}
}

// RequireRole membatasi akses hanya untuk role tertentu. Dipasang setelah
// Authenticate.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFrom(r.Context())
			if !ok {
				response.Error(w, r, domain.Unauthorized("token akses tidak ditemukan"))
				return
			}
			if _, permitted := allowed[user.Role]; !permitted {
				response.Error(w, r, domain.Forbidden("role kamu tidak punya akses ke menu ini"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserFrom(ctx context.Context) (AuthUser, bool) {
	user, ok := ctx.Value(userCtxKey).(AuthUser)
	return user, ok
}

// UserIDFrom mengembalikan ID pengguna, atau uuid.Nil kalau request tidak
// terautentikasi. Dipakai untuk mengisi kolom created_by/recorded_by.
func UserIDFrom(ctx context.Context) uuid.UUID {
	if user, ok := UserFrom(ctx); ok {
		return user.ID
	}
	return uuid.Nil
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	scheme, value, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(value)
}
