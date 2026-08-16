package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role pengguna back office.
const (
	RoleOwner   = "owner"   // akses penuh termasuk laporan keuangan & manajemen user
	RoleAdmin   = "admin"   // seluruh operasional harian
	RoleTripper = "tripper" // hanya shopping list dan input realisasi belanja
)

// AllRoles dipakai middleware untuk endpoint yang boleh diakses siapa saja
// yang sudah login.
var AllRoles = []string{RoleOwner, RoleAdmin, RoleTripper}

// StaffRoles adalah role yang boleh mengubah data operasional.
var StaffRoles = []string{RoleOwner, RoleAdmin}

func IsValidRole(role string) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleTripper:
		return true
	default:
		return false
	}
}

type User struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	Name         string     `db:"name"          json:"name"`
	Email        string     `db:"email"         json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Role         string     `db:"role"          json:"role"`
	Phone        *string    `db:"phone"         json:"phone"`
	IsActive     bool       `db:"is_active"     json:"is_active"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"    json:"updated_at"`
}

// RefreshToken adalah sesi login yang masih berlaku.
type RefreshToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	UserAgent *string    `db:"user_agent"`
	IPAddress *string    `db:"ip_address"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
}

func (t RefreshToken) IsUsable() bool {
	return t.RevokedAt == nil && t.ExpiresAt.After(time.Now())
}
