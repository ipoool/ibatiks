package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/docnum"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type CustomerService struct {
	pool      *pgxpool.Pool
	customers *repository.CustomerRepo
}

func NewCustomerService(pool *pgxpool.Pool, customers *repository.CustomerRepo) *CustomerService {
	return &CustomerService{pool: pool, customers: customers}
}

type CustomerInput struct {
	Name       string
	PhoneWA    string
	Email      *string
	Instagram  *string
	Address    *string
	City       *string
	Province   *string
	PostalCode *string
	Notes      *string
}

func (s *CustomerService) Create(ctx context.Context, in CustomerInput) (*domain.Customer, error) {
	phone := domain.NormalizePhoneWA(in.PhoneWA)
	if phone == "" {
		return nil, domain.Validation("nomor WhatsApp tidak valid", map[string]string{
			"phone_wa": "isi nomor WhatsApp yang benar, contoh 081234567890",
		})
	}

	var customer *domain.Customer
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		// Nomor diambil di dalam transaksi yang sama dengan insert-nya, supaya
		// dua admin yang menyimpan bersamaan tidak mendapat kode kembar.
		code, err := docnum.Next(ctx, tx, docnum.Customer, time.Now().Year())
		if err != nil {
			return err
		}

		customer, err = s.customers.Create(ctx, tx, repository.CustomerParams{
			Code:       code,
			Name:       strings.TrimSpace(in.Name),
			PhoneWA:    phone,
			Email:      trimPtr(in.Email),
			Instagram:  trimPtr(in.Instagram),
			Address:    trimPtr(in.Address),
			City:       trimPtr(in.City),
			Province:   trimPtr(in.Province),
			PostalCode: trimPtr(in.PostalCode),
			Notes:      trimPtr(in.Notes),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *CustomerService) List(ctx context.Context, p pagination.Params) ([]domain.Customer, int64, error) {
	return s.customers.List(ctx, s.pool, p)
}

func (s *CustomerService) Get(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	return s.customers.GetByID(ctx, s.pool, id)
}

func (s *CustomerService) Stats(ctx context.Context, id uuid.UUID) (*repository.CustomerStats, error) {
	if _, err := s.customers.GetByID(ctx, s.pool, id); err != nil {
		return nil, err
	}
	return s.customers.Stats(ctx, s.pool, id)
}

func (s *CustomerService) Update(ctx context.Context, id uuid.UUID, in CustomerInput) (*domain.Customer, error) {
	phone := domain.NormalizePhoneWA(in.PhoneWA)
	if phone == "" {
		return nil, domain.Validation("nomor WhatsApp tidak valid", map[string]string{
			"phone_wa": "isi nomor WhatsApp yang benar, contoh 081234567890",
		})
	}

	return s.customers.Update(ctx, s.pool, id, repository.CustomerParams{
		Name:       strings.TrimSpace(in.Name),
		PhoneWA:    phone,
		Email:      trimPtr(in.Email),
		Instagram:  trimPtr(in.Instagram),
		Address:    trimPtr(in.Address),
		City:       trimPtr(in.City),
		Province:   trimPtr(in.Province),
		PostalCode: trimPtr(in.PostalCode),
		Notes:      trimPtr(in.Notes),
	})
}

// Delete melakukan soft delete. Customer yang punya order tetap dipertahankan
// datanya karena order lama masih mereferensikannya.
func (s *CustomerService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.customers.GetByID(ctx, s.pool, id); err != nil {
		return err
	}
	return s.customers.SoftDelete(ctx, s.pool, id)
}

// trimPtr merapikan string opsional: spasi dibuang, dan string kosong menjadi
// NULL supaya kolom opsional tidak terisi string kosong yang membingungkan.
func trimPtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
