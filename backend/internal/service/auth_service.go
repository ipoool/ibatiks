// Package service berisi aturan bisnis aplikasi.
//
// Handler HTTP hanya menerjemahkan request menjadi pemanggilan service, dan
// repository hanya bicara SQL. Semua keputusan — boleh atau tidaknya sebuah
// operasi, bagaimana angka dihitung, kapan status berpindah — ada di sini.
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/pkg/token"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

// bcryptCost 12 adalah kompromi wajar untuk aplikasi back office: cukup mahal
// untuk menahan brute force, tapi login tetap terasa instan.
const bcryptCost = 12

type AuthService struct {
	pool   *pgxpool.Pool
	users  *repository.UserRepo
	tokens *token.Manager
}

func NewAuthService(pool *pgxpool.Pool, users *repository.UserRepo, tokens *token.Manager) *AuthService {
	return &AuthService{pool: pool, users: users, tokens: tokens}
}

type Session struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         *domain.User `json:"user"`
}

// Login memverifikasi kredensial lalu menerbitkan sepasang token.
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*Session, error) {
	user, err := s.users.GetByEmail(ctx, s.pool, strings.TrimSpace(email))
	if err != nil {
		if domainErr, ok := domain.AsError(err); ok && domainErr.Code == domain.CodeNotFound {
			// Jangan bocorkan apakah email terdaftar: pesan yang sama dipakai
			// untuk email tak dikenal maupun password salah.
			return nil, domain.Unauthorized("email atau password salah")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.Unauthorized("email atau password salah")
	}
	if !user.IsActive {
		return nil, domain.Forbidden("akun ini sudah dinonaktifkan, hubungi owner")
	}

	session, err := s.issueSession(ctx, user, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	// Kegagalan mencatat waktu login tidak boleh menggagalkan login itu sendiri.
	_ = s.users.TouchLastLogin(ctx, s.pool, user.ID)
	return session, nil
}

// Refresh menukar refresh token dengan sepasang token baru.
//
// Token lama langsung dicabut (rotasi): kalau refresh token bocor dan dipakai
// penyerang, pemakaian berikutnya oleh pemilik asli akan gagal, sehingga
// pembobolan cepat ketahuan.
func (s *AuthService) Refresh(ctx context.Context, rawToken, userAgent, ipAddress string) (*Session, error) {
	hashed := token.HashRefreshToken(rawToken)

	stored, err := s.users.GetRefreshToken(ctx, s.pool, hashed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.Unauthorized("sesi tidak ditemukan, silakan login ulang")
		}
		if domainErr, ok := domain.AsError(err); ok && domainErr.Code == domain.CodeNotFound {
			return nil, domain.Unauthorized("sesi tidak ditemukan, silakan login ulang")
		}
		return nil, err
	}
	if !stored.IsUsable() {
		return nil, domain.Unauthorized("sesi sudah berakhir, silakan login ulang")
	}

	user, err := s.users.GetByID(ctx, s.pool, stored.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, domain.Forbidden("akun ini sudah dinonaktifkan, hubungi owner")
	}

	var session *Session
	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.users.RevokeRefreshToken(ctx, tx, hashed); err != nil {
			return err
		}
		session, err = s.issueSessionTx(ctx, tx, user, userAgent, ipAddress)
		return err
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Logout mencabut satu sesi.
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	return s.users.RevokeRefreshToken(ctx, s.pool, token.HashRefreshToken(rawToken))
}

// LogoutAll mencabut seluruh sesi pengguna, dipakai saat password diganti.
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.users.RevokeAllRefreshTokens(ctx, s.pool, userID)
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, s.pool, userID)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, s.pool, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return domain.Validation("password saat ini salah", map[string]string{
			"current_password": "tidak cocok dengan password akun",
		})
	}

	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.users.UpdatePassword(ctx, tx, userID, hashed); err != nil {
			return err
		}
		// Semua perangkat lain dipaksa login ulang dengan password baru.
		return s.users.RevokeAllRefreshTokens(ctx, tx, userID)
	})
}

func (s *AuthService) issueSession(ctx context.Context, user *domain.User, userAgent, ipAddress string) (*Session, error) {
	return s.issueSessionTx(ctx, s.pool, user, userAgent, ipAddress)
}

func (s *AuthService) issueSessionTx(ctx context.Context, q db.Querier, user *domain.User, userAgent, ipAddress string) (*Session, error) {
	accessToken, expiresAt, err := s.tokens.IssueAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, domain.Internal(err)
	}

	rawRefresh, hashedRefresh, err := token.GenerateRefreshToken()
	if err != nil {
		return nil, domain.Internal(err)
	}

	if err := s.users.CreateRefreshToken(ctx, q, user.ID, hashedRefresh,
		time.Now().Add(s.tokens.RefreshTTL()), optionalString(userAgent), optionalString(ipAddress)); err != nil {
		return nil, err
	}

	return &Session{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}

// --- Manajemen pengguna ----------------------------------------------------

type UserService struct {
	pool  *pgxpool.Pool
	users *repository.UserRepo
}

func NewUserService(pool *pgxpool.Pool, users *repository.UserRepo) *UserService {
	return &UserService{pool: pool, users: users}
}

type CreateUserInput struct {
	Name     string
	Email    string
	Password string
	Role     string
	Phone    *string
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	if !domain.IsValidRole(in.Role) {
		return nil, domain.Validation("role tidak dikenal", map[string]string{
			"role": "pilih salah satu: owner, admin, tripper",
		})
	}

	hashed, err := HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	return s.users.Create(ctx, s.pool, repository.CreateUserParams{
		Name:         strings.TrimSpace(in.Name),
		Email:        strings.ToLower(strings.TrimSpace(in.Email)),
		PasswordHash: hashed,
		Role:         in.Role,
		Phone:        in.Phone,
	})
}

func (s *UserService) List(ctx context.Context, p pagination.Params) ([]domain.User, int64, error) {
	return s.users.List(ctx, s.pool, p)
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, s.pool, id)
}

type UpdateUserInput struct {
	Name     string
	Role     string
	Phone    *string
	IsActive bool
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, in UpdateUserInput) (*domain.User, error) {
	if !domain.IsValidRole(in.Role) {
		return nil, domain.Validation("role tidak dikenal", map[string]string{
			"role": "pilih salah satu: owner, admin, tripper",
		})
	}

	current, err := s.users.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}

	// Owner terakhir tidak boleh diturunkan perannya atau dinonaktifkan, karena
	// setelah itu tidak ada lagi yang bisa mengelola pengguna.
	if current.Role == domain.RoleOwner && (in.Role != domain.RoleOwner || !in.IsActive) {
		remaining, err := s.users.CountActiveOwners(ctx, s.pool, id)
		if err != nil {
			return nil, err
		}
		if remaining == 0 {
			return nil, domain.Conflict("tidak bisa mengubah owner terakhir, angkat owner lain terlebih dahulu")
		}
	}

	return s.users.Update(ctx, s.pool, id, repository.UpdateUserParams{
		Name:     strings.TrimSpace(in.Name),
		Role:     in.Role,
		Phone:    in.Phone,
		IsActive: in.IsActive,
	})
}

// ResetPassword dipakai owner untuk mengganti password pengguna lain yang lupa.
func (s *UserService) ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.users.UpdatePassword(ctx, tx, id, hashed); err != nil {
			return err
		}
		return s.users.RevokeAllRefreshTokens(ctx, tx, id)
	})
}

func (s *UserService) Delete(ctx context.Context, id, actorID uuid.UUID) error {
	if id == actorID {
		return domain.Conflict("tidak bisa menghapus akun yang sedang kamu pakai")
	}

	user, err := s.users.GetByID(ctx, s.pool, id)
	if err != nil {
		return err
	}
	if user.Role == domain.RoleOwner {
		remaining, err := s.users.CountActiveOwners(ctx, s.pool, id)
		if err != nil {
			return err
		}
		if remaining == 0 {
			return domain.Conflict("tidak bisa menghapus owner terakhir")
		}
	}

	return s.users.Delete(ctx, s.pool, id)
}

// HashPassword membuat hash bcrypt. Dipakai bersama oleh service dan perintah seed.
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", domain.Validation("password terlalu pendek", map[string]string{
			"password": "minimal 8 karakter",
		})
	}
	// bcrypt memotong input di 72 byte; menolaknya lebih jujur daripada
	// diam-diam mengabaikan sisa karakter yang diketik pengguna.
	if len(password) > 72 {
		return "", domain.Validation("password terlalu panjang", map[string]string{
			"password": "maksimal 72 karakter",
		})
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", domain.Internal(err)
	}
	return string(hashed), nil
}

func optionalString(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}
