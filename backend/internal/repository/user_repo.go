package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
)

const userColumns = `id, name, email, password_hash, role, phone, is_active,
	                 COALESCE(permissions, '{}') AS permissions,
	                 last_login_at, created_at, updated_at`

type UserRepo struct{}

func NewUserRepo() *UserRepo { return &UserRepo{} }

type CreateUserParams struct {
	Name         string
	Email        string
	PasswordHash string
	Role         string
	Phone        *string
	Permissions  []string
}

func (r *UserRepo) Create(ctx context.Context, q db.Querier, p CreateUserParams) (*domain.User, error) {
	return collectOne[domain.User](ctx, q, "pengguna", `
		INSERT INTO users (name, email, password_hash, role, phone, permissions)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+userColumns,
		p.Name, p.Email, p.PasswordHash, p.Role, p.Phone, nullableStrings(p.Permissions))
}

func (r *UserRepo) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.User, error) {
	return collectOne[domain.User](ctx, q, "pengguna",
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id)
}

func (r *UserRepo) GetByEmail(ctx context.Context, q db.Querier, email string) (*domain.User, error) {
	return collectOne[domain.User](ctx, q, "pengguna",
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email)
}

func (r *UserRepo) List(ctx context.Context, q db.Querier, p pagination.Params) ([]domain.User, int64, error) {
	conditions := []string{}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", len(args), len(args)))
	}
	where := buildWhere(conditions)

	var total int64
	if err := q.QueryRow(ctx, `SELECT count(*) FROM users`+where, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"name":       "name",
		"email":      "email",
		"role":       "role",
		"created_at": "created_at",
	}, "created_at")

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(
		`SELECT %s FROM users%s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		userColumns, where, sortCol, p.Order, len(args)-1, len(args))

	users, err := collectRows[domain.User](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

type UpdateUserParams struct {
	Name        string
	Role        string
	Phone       *string
	IsActive    bool
	Permissions []string
}

func (r *UserRepo) Update(ctx context.Context, q db.Querier, id uuid.UUID, p UpdateUserParams) (*domain.User, error) {
	return collectOne[domain.User](ctx, q, "pengguna", `
		UPDATE users
		SET name = $2, role = $3, phone = $4, is_active = $5, permissions = $6
		WHERE id = $1
		RETURNING `+userColumns,
		id, p.Name, p.Role, p.Phone, p.IsActive, nullableStrings(p.Permissions))
}

// nullableStrings menyimpan daftar kosong sebagai NULL.
//
// Bedanya penting: NULL berarti "ikut bawaan role", sedangkan array kosong
// berarti "tidak boleh membuka menu apa pun" — dan itu bukan yang dimaksud
// ketika owner belum pernah menyetel apa-apa.
func nullableStrings(values []string) any {
	if len(values) == 0 {
		return nil
	}
	return values
}

func (r *UserRepo) UpdatePassword(ctx context.Context, q db.Querier, id uuid.UUID, passwordHash string) error {
	return execExpectOne(ctx, q, "pengguna",
		`UPDATE users SET password_hash = $2 WHERE id = $1`, id, passwordHash)
}

func (r *UserRepo) TouchLastLogin(ctx context.Context, q db.Querier, id uuid.UUID) error {
	_, err := exec(ctx, q, `UPDATE users SET last_login_at = now() WHERE id = $1`, id)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "pengguna", `DELETE FROM users WHERE id = $1`, id)
}

// CountOwners dipakai untuk mencegah owner terakhir dihapus atau diturunkan
// perannya, yang akan mengunci semua orang dari menu manajemen pengguna.
func (r *UserRepo) CountActiveOwners(ctx context.Context, q db.Querier, excludeID uuid.UUID) (int, error) {
	var count int
	err := q.QueryRow(ctx, `
		SELECT count(*) FROM users
		WHERE role = $1 AND is_active = TRUE AND id <> $2`,
		domain.RoleOwner, excludeID).Scan(&count)
	if err != nil {
		return 0, wrapPgError(err)
	}
	return count, nil
}

// --- Refresh token ---------------------------------------------------------

func (r *UserRepo) CreateRefreshToken(
	ctx context.Context, q db.Querier,
	userID uuid.UUID, tokenHash string, expiresAt time.Time,
	userAgent, ipAddress *string,
) error {
	_, err := exec(ctx, q, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, tokenHash, expiresAt, userAgent, ipAddress)
	return err
}

func (r *UserRepo) GetRefreshToken(ctx context.Context, q db.Querier, tokenHash string) (*domain.RefreshToken, error) {
	return collectOne[domain.RefreshToken](ctx, q, "sesi", `
		SELECT id, user_id, token_hash, user_agent, ip_address, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`, tokenHash)
}

func (r *UserRepo) RevokeRefreshToken(ctx context.Context, q db.Querier, tokenHash string) error {
	_, err := exec(ctx, q,
		`UPDATE refresh_tokens SET revoked_at = now()
		 WHERE token_hash = $1 AND revoked_at IS NULL`, tokenHash)
	return err
}

func (r *UserRepo) RevokeAllRefreshTokens(ctx context.Context, q db.Querier, userID uuid.UUID) error {
	_, err := exec(ctx, q,
		`UPDATE refresh_tokens SET revoked_at = now()
		 WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

// PurgeExpiredRefreshTokens membersihkan sesi lama supaya tabel tidak tumbuh
// tanpa batas. Dipanggil berkala oleh background job di cmd/api.
func (r *UserRepo) PurgeExpiredRefreshTokens(ctx context.Context, q db.Querier) (int64, error) {
	return exec(ctx, q,
		`DELETE FROM refresh_tokens
		 WHERE expires_at < now() - INTERVAL '30 days'
		    OR (revoked_at IS NOT NULL AND revoked_at < now() - INTERVAL '30 days')`)
}
