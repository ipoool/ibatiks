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
	Name        string
	PhoneWA     string
	Email       *string
	Socials     []domain.Social
	Address     *string
	City        *string
	District    *string
	Subdistrict *string
	Province    *string
	PostalCode  *string
	Notes       *string
}

func (s *CustomerService) Create(ctx context.Context, in CustomerInput) (*domain.Customer, error) {
	phone := domain.NormalizePhoneWA(in.PhoneWA)
	if phone == "" {
		return nil, domain.Validation("nomor WhatsApp tidak valid", map[string]string{
			"phone_wa": "isi nomor WhatsApp yang benar, contoh 081234567890",
		})
	}

	socials, err := bersihkanSocials(in.Socials)
	if err != nil {
		return nil, err
	}

	var customer *domain.Customer
	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		// Nomor diambil di dalam transaksi yang sama dengan insert-nya, supaya
		// dua admin yang menyimpan bersamaan tidak mendapat kode kembar.
		code, err := docnum.Next(ctx, tx, docnum.Customer, time.Now().Year())
		if err != nil {
			return err
		}

		customer, err = s.customers.Create(ctx, tx, repository.CustomerParams{
			Code:        code,
			Name:        strings.TrimSpace(in.Name),
			PhoneWA:     phone,
			Email:       trimPtr(in.Email),
			Socials:     socials,
			Address:     trimPtr(in.Address),
			City:        trimPtr(in.City),
			District:    trimPtr(in.District),
			Subdistrict: trimPtr(in.Subdistrict),
			Province:    trimPtr(in.Province),
			PostalCode:  trimPtr(in.PostalCode),
			Notes:       trimPtr(in.Notes),
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

	socials, err := bersihkanSocials(in.Socials)
	if err != nil {
		return nil, err
	}

	return s.customers.Update(ctx, s.pool, id, repository.CustomerParams{
		Name:        strings.TrimSpace(in.Name),
		PhoneWA:     phone,
		Email:       trimPtr(in.Email),
		Socials:     socials,
		Address:     trimPtr(in.Address),
		City:        trimPtr(in.City),
		District:    trimPtr(in.District),
		Subdistrict: trimPtr(in.Subdistrict),
		Province:    trimPtr(in.Province),
		PostalCode:  trimPtr(in.PostalCode),
		Notes:       trimPtr(in.Notes),
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

/*
 * bersihkanSocials membuang baris kosong dan menolak platform yang tidak
 * dikenal.
 *
 * Baris kosong muncul dari formulir: admin menekan "Tambah akun" lalu berpindah
 * pikiran, dan barisnya tertinggal tanpa isi. Menyimpannya berarti daftar akun
 * yang penuh baris hampa, dan tidak ada yang tahu itu bekas apa.
 */
func bersihkanSocials(list []domain.Social) ([]domain.Social, error) {
	hasil := make([]domain.Social, 0, len(list))
	for _, akun := range list {
		platform := strings.ToLower(strings.TrimSpace(akun.Platform))
		handle := strings.TrimSpace(akun.Handle)
		if handle == "" {
			continue
		}
		if !domain.IsValidSocialPlatform(platform) {
			return nil, domain.Validation("platform media sosial tidak dikenal", map[string]string{
				"socials": "pilih: " + strings.Join(domain.SocialPlatforms, ", "),
			})
		}
		hasil = append(hasil, domain.Social{Platform: platform, Handle: handle})
	}
	return hasil, nil
}
