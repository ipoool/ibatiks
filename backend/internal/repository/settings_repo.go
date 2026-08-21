package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
)

type SettingsRepo struct{}

func NewSettingsRepo() *SettingsRepo { return &SettingsRepo{} }

func (r *SettingsRepo) All(ctx context.Context, q db.Querier) (domain.Settings, error) {
	rows, err := collectRows[domain.Setting](ctx, q,
		`SELECT key, value, description, updated_at FROM app_settings ORDER BY key ASC`)
	if err != nil {
		return nil, err
	}

	settings := make(domain.Settings, len(rows))
	for _, row := range rows {
		settings[row.Key] = row.Value
	}
	return settings, nil
}

func (r *SettingsRepo) List(ctx context.Context, q db.Querier) ([]domain.Setting, error) {
	return collectRows[domain.Setting](ctx, q,
		`SELECT key, value, description, updated_at FROM app_settings ORDER BY key ASC`)
}

func (r *SettingsRepo) Get(ctx context.Context, q db.Querier, key string) (string, error) {
	var value string
	err := q.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = $1`, key).Scan(&value)
	if err != nil {
		return "", wrapPgError(err)
	}
	return value, nil
}

// Upsert menyimpan satu pengaturan. Kunci baru diizinkan supaya pemilik toko
// bisa menambah template pesan tambahan tanpa perlu migrasi.
func (r *SettingsRepo) Upsert(ctx context.Context, q db.Querier, key, value string) error {
	_, err := exec(ctx, q, `
		INSERT INTO app_settings (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, key, value)
	return err
}

// --- Audit log -------------------------------------------------------------

type AuditRepo struct{}

func NewAuditRepo() *AuditRepo { return &AuditRepo{} }

type AuditParams struct {
	UserID    *uuid.UUID
	Entity    string
	EntityID  *uuid.UUID
	Action    string
	Changes   any
	IPAddress *string
}

// Record menulis satu jejak perubahan. Kegagalan menulis audit tidak boleh
// membatalkan operasi bisnis, jadi pemanggil cukup mencatat errornya ke log.
func (r *AuditRepo) Record(ctx context.Context, q db.Querier, p AuditParams) error {
	var payload []byte
	if p.Changes != nil {
		encoded, err := json.Marshal(p.Changes)
		if err != nil {
			return fmt.Errorf("encode perubahan audit: %w", err)
		}
		payload = encoded
	}

	_, err := exec(ctx, q, `
		INSERT INTO audit_logs (user_id, entity, entity_id, action, changes, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		p.UserID, p.Entity, p.EntityID, p.Action, payload, p.IPAddress)
	return err
}

func (r *AuditRepo) List(ctx context.Context, q db.Querier, p pagination.Params, entity string, entityID, userID *uuid.UUID) ([]domain.AuditLogDetail, int64, error) {
	conditions := []string{}
	args := []any{}

	if userID != nil {
		args = append(args, *userID)
		conditions = append(conditions, fmt.Sprintf("a.user_id = $%d", len(args)))
	}
	if entity != "" {
		args = append(args, entity)
		conditions = append(conditions, fmt.Sprintf("a.entity = $%d", len(args)))
	}
	if entityID != nil {
		args = append(args, *entityID)
		conditions = append(conditions, fmt.Sprintf("a.entity_id = $%d", len(args)))
	}
	where := buildWhere(conditions)

	var total int64
	if err := q.QueryRow(ctx, `SELECT count(*) FROM audit_logs a`+where, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT a.id, a.user_id, a.entity, a.entity_id, a.action, a.changes, a.ip_address, a.created_at,
		       u.name AS user_name
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.user_id%s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))

	logs, err := collectRows[domain.AuditLogDetail](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// Actors mengembalikan akun yang benar-benar pernah tercatat di jejak
// perubahan, untuk mengisi penyaringnya.
//
// Bukan daftar pengguna: akun yang sudah dihapus tetap punya jejak, dan tanpa
// namanya di daftar penyaring barisnya jadi tidak bisa dicari sama sekali.
// Sebaliknya, akun yang belum pernah mengubah apa pun hanya jadi pilihan yang
// selalu berujung daftar kosong.
func (r *AuditRepo) Actors(ctx context.Context, q db.Querier) ([]domain.AuditActor, error) {
	return collectRows[domain.AuditActor](ctx, q, `
		SELECT a.user_id AS id, COALESCE(max(u.name), 'Akun terhapus') AS name,
		       count(*)::int AS log_count
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.user_id
		WHERE a.user_id IS NOT NULL
		GROUP BY a.user_id
		ORDER BY 2 ASC`)
}
