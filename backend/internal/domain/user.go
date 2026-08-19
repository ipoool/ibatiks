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
	ID           uuid.UUID `db:"id"            json:"id"`
	Name         string    `db:"name"          json:"name"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         string    `db:"role"          json:"role"`
	Phone        *string   `db:"phone"         json:"phone"`
	IsActive     bool      `db:"is_active"     json:"is_active"`
	// Permissions kosong berarti mengikuti bawaan role; lihat permission.go.
	Permissions []string `db:"permissions" json:"permissions"`
	// EffectivePermissions adalah hasil gabungan Permissions dengan bawaan
	// role. Tidak disimpan di database — diisi lapisan HTTP sebelum dikirim,
	// supaya antarmuka tidak perlu menyalin aturan hak akses sendiri lalu ikut
	// melenceng ketika aturannya berubah.
	EffectivePermissions []string   `db:"-" json:"effective_permissions"`
	LastLoginAt          *time.Time `db:"last_login_at" json:"last_login_at"`
	CreatedAt            time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"    json:"updated_at"`
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

// Penjagaan percobaan login.
const (
	// LoginMaxAttempts adalah jumlah kegagalan berturut-turut sebelum sebuah
	// email dikunci.
	LoginMaxAttempts = 5
	// LoginBlockDuration adalah lama penguncian, sekaligus lebar jendela
	// penghitungan: lima kegagalan harus terjadi dalam rentang ini untuk
	// mengunci. Tanpa jendela, satu salah ketik hari ini dan empat lagi bulan
	// depan akan mengunci akun tanpa ada yang menyerang apa pun.
	LoginBlockDuration = 5 * time.Minute
)

// LoginAttempt adalah rekaman kegagalan login untuk satu alamat email.
//
// Dihitung per email, bukan per alamat IP. Itu keputusan sadar: penebak
// password yang berpindah-pindah IP tetap tertahan, dengan konsekuensi yang
// harus diketahui — siapa pun yang tahu alamat email seorang pengguna bisa
// membuatnya terkunci dengan sengaja salah lima kali.
//
// Emailnya dicatat apa adanya termasuk yang tidak terdaftar. Kalau hanya email
// terdaftar yang dihitung, pola penguncian justru membocorkan email mana yang
// ada di sistem.
type LoginAttempt struct {
	Email        string     `db:"email"          json:"email"`
	FailedCount  int        `db:"failed_count"   json:"failed_count"`
	LastFailedAt *time.Time `db:"last_failed_at" json:"last_failed_at"`
	BlockedUntil *time.Time `db:"blocked_until"  json:"blocked_until"`
	LastIP       *string    `db:"last_ip"        json:"last_ip"`
	UpdatedAt    time.Time  `db:"updated_at"     json:"updated_at"`
}

// BlockedFor mengembalikan sisa waktu penguncian, atau nol kalau tidak dikunci.
func (a *LoginAttempt) BlockedFor(now time.Time) time.Duration {
	if a == nil || a.BlockedUntil == nil {
		return 0
	}
	if sisa := a.BlockedUntil.Sub(now); sisa > 0 {
		return sisa
	}
	return 0
}

// AttemptsLeft mengembalikan sisa percobaan sebelum terkunci.
//
// Kegagalan yang sudah lewat jendela penghitungan tidak ikut dihitung — hitungan
// dimulai lagi dari nol.
func (a *LoginAttempt) AttemptsLeft(now time.Time) int {
	if a == nil || a.LastFailedAt == nil || now.Sub(*a.LastFailedAt) > LoginBlockDuration {
		return LoginMaxAttempts
	}
	sisa := LoginMaxAttempts - a.FailedCount
	if sisa < 0 {
		return 0
	}
	return sisa
}
