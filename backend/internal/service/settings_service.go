package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type SettingsService struct {
	pool     *pgxpool.Pool
	settings *repository.SettingsRepo
	audit    *repository.AuditRepo
}

func NewSettingsService(pool *pgxpool.Pool, settings *repository.SettingsRepo, audit *repository.AuditRepo) *SettingsService {
	return &SettingsService{pool: pool, settings: settings, audit: audit}
}

func (s *SettingsService) List(ctx context.Context) ([]domain.Setting, error) {
	return s.settings.List(ctx, s.pool)
}

func (s *SettingsService) All(ctx context.Context) (domain.Settings, error) {
	return s.settings.All(ctx, s.pool)
}

// Update menyimpan beberapa pengaturan sekaligus dalam satu transaksi, supaya
// form pengaturan tidak pernah tersimpan separuh.
func (s *SettingsService) Update(ctx context.Context, values map[string]string, actorID uuid.UUID) (domain.Settings, error) {
	if len(values) == 0 {
		return nil, domain.Validation("tidak ada pengaturan yang dikirim", map[string]string{
			"settings": "isi minimal satu pengaturan",
		})
	}

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		for key, value := range values {
			key = strings.TrimSpace(key)
			if key == "" {
				return domain.Validation("kunci pengaturan kosong", map[string]string{
					"settings": "nama pengaturan tidak boleh kosong",
				})
			}
			if err := s.settings.Upsert(ctx, tx, key, value); err != nil {
				return err
			}
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:  nullableUUID(actorID),
			Entity:  "settings",
			Action:  domain.AuditUpdate,
			Changes: map[string]any{"keys": mapKeys(values)},
		})
	})
	if err != nil {
		return nil, err
	}

	return s.settings.All(ctx, s.pool)
}

// --- Audit log -------------------------------------------------------------

type AuditService struct {
	pool  *pgxpool.Pool
	audit *repository.AuditRepo
}

func NewAuditService(pool *pgxpool.Pool, audit *repository.AuditRepo) *AuditService {
	return &AuditService{pool: pool, audit: audit}
}

func (s *AuditService) List(ctx context.Context, p pagination.Params, entity string, entityID *uuid.UUID) ([]domain.AuditLogDetail, int64, error) {
	return s.audit.List(ctx, s.pool, p, entity, entityID)
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
