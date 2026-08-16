// Package token menangani pembuatan dan verifikasi JWT access token serta
// refresh token opaque.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ipoool/jastipin/backend/internal/config"
)

var ErrInvalidToken = errors.New("token tidak valid atau sudah kedaluwarsa")

// Claims adalah isi access token. Role dan daftar hak akses ikut dibawa supaya
// pengecekan di middleware tidak perlu query database tiap request.
//
// Konsekuensinya, perubahan hak akses baru berlaku saat token diperbarui.
// Karena itu service pengguna mencabut sesi seseorang begitu haknya diubah —
// lihat UserService.Update.
type Claims struct {
	UserID      uuid.UUID `json:"uid"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"perms,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(cfg config.JWT) *Manager {
	return &Manager{
		secret:     []byte(cfg.Secret),
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

// IssueAccessToken membuat JWT bertanda tangan HS256.
func (m *Manager) IssueAccessToken(userID uuid.UUID, email, role string, permissions []string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTTL)

	claims := Claims{
		UserID:      userID,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.NewString(),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("tanda tangani token: %w", err)
	}
	return signed, expiresAt, nil
}

// ParseAccessToken memverifikasi tanda tangan dan masa berlaku token.
func (m *Manager) ParseAccessToken(raw string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		// Tolak token yang mengklaim algoritma lain (mis. "none" atau RS256
		// dengan public key kita sebagai HMAC secret).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritma tanda tangan tidak didukung: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GenerateRefreshToken menghasilkan token acak beserta hash-nya. Yang dikirim
// ke client adalah token mentahnya, yang disimpan di database hash-nya, supaya
// isi tabel refresh_tokens tidak bisa langsung dipakai login kalau bocor.
func GenerateRefreshToken() (raw, hashed string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("buat refresh token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, HashRefreshToken(raw), nil
}

// HashRefreshToken memakai SHA-256, bukan bcrypt: nilainya sudah acak 256-bit
// sehingga tidak rentan brute force, dan lookup-nya harus cepat karena dipakai
// setiap kali access token diperbarui.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
