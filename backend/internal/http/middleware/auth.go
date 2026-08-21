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
	ID          uuid.UUID
	Email       string
	Role        string
	Scope       string
	Permissions []string
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

			// Token membawa hak akses dan wewenangnya sendiri supaya
			// pengecekan di sini tidak perlu menyentuh database tiap request.
			//
			// Token yang terbit sebelum role pindah ke database tidak membawa
			// keduanya. Selama sisa umurnya — paling lama satu putaran access
			// token — isinya dijatuhkan ke bawaan role lama, supaya sesi yang
			// sedang berjalan tidak mendadak kehilangan seluruh menunya.
			permissions := claims.Permissions
			if len(permissions) == 0 {
				permissions = domain.LegacyEffectivePermissions(claims.Role, nil)
			}
			scope := claims.Scope
			if scope == "" {
				scope = domain.LegacyScope(claims.Role)
			}

			user := AuthUser{
				ID:          claims.UserID,
				Email:       claims.Email,
				Role:        claims.Role,
				Scope:       scope,
				Permissions: permissions,
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userCtxKey, user)))
		})
	}
}

// RequireScope membatasi akses ke wewenang tertentu. Dipasang setelah
// Authenticate.
//
// Menggantikan penjaga lama yang mencocokkan nama role. Sejak role jadi data,
// nama role tidak lagi bisa jadi pegangan: role bikinan toko sendiri bukan
// owner, admin, maupun tripper, dan penjaga berbasis nama akan menolaknya di
// seluruh endpoint operasional walaupun menunya sudah dicentang.
func RequireScope(scopes ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(scopes))
	for _, s := range scopes {
		allowed[s] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFrom(r.Context())
			if !ok {
				response.Error(w, r, domain.Unauthorized("token akses tidak ditemukan"))
				return
			}
			if _, permitted := allowed[user.Scope]; !permitted {
				response.Error(w, r, domain.Forbidden("role kamu tidak punya akses ke menu ini"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission membatasi akses ke satu menu. Dipasang setelah Authenticate,
// biasanya bersama RequireScope yang menjaga batas kasarnya.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFrom(r.Context())
			if !ok {
				response.Error(w, r, domain.Unauthorized("token akses tidak ditemukan"))
				return
			}
			if !domain.HasPermission(user.Permissions, permission) {
				response.Error(w, r, domain.Forbidden("akunmu tidak diberi akses ke menu ini"))
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
