package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	Code       string     `db:"code"        json:"code"`
	Name       string     `db:"name"        json:"name"`
	PhoneWA    string     `db:"phone_wa"    json:"phone_wa"`
	Email      *string    `db:"email"       json:"email"`
	Instagram  *string    `db:"instagram"   json:"instagram"`
	Address    *string    `db:"address"     json:"address"`
	City       *string    `db:"city"        json:"city"`
	Province   *string    `db:"province"    json:"province"`
	PostalCode *string    `db:"postal_code" json:"postal_code"`
	Notes      *string    `db:"notes"       json:"notes"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"  json:"-"`
}

var nonDigit = regexp.MustCompile(`\D`)

// NormalizePhoneWA merapikan nomor telepon ke format internasional tanpa tanda
// baca (62812xxxx), yang bisa langsung dipakai pada link wa.me.
//
// Menerima berbagai gaya penulisan yang biasa diketik admin:
// "0812-3456-7890", "+62 812 3456 7890", "62812345678", "(0812) 3456-7890".
func NormalizePhoneWA(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	// Awalan "+" atau "00" harus diperiksa sebelum tanda baca dibuang: itulah
	// satu-satunya penanda bahwa nomor sudah ditulis lengkap dengan kode
	// negara. Tanpa pemeriksaan ini, nomor Jepang "+81 90..." akan disangka
	// nomor Indonesia tanpa nol dan salah diberi awalan 62.
	international := strings.HasPrefix(trimmed, "+")

	digits := nonDigit.ReplaceAllString(trimmed, "")
	if digits == "" {
		return ""
	}
	if strings.HasPrefix(digits, "00") {
		international = true
		digits = strings.TrimPrefix(digits, "00")
	}
	if international {
		return digits
	}

	switch {
	case strings.HasPrefix(digits, "62"):
		return digits
	case strings.HasPrefix(digits, "0"):
		return "62" + strings.TrimPrefix(digits, "0")
	case strings.HasPrefix(digits, "8"):
		// Admin kadang mengetik nomor Indonesia tanpa nol di depan.
		return "62" + digits
	default:
		return digits
	}
}
