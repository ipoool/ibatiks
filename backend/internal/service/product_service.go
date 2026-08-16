package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type ProductService struct {
	pool     *pgxpool.Pool
	products *repository.ProductRepo
}

func NewProductService(pool *pgxpool.Pool, products *repository.ProductRepo) *ProductService {
	return &ProductService{pool: pool, products: products}
}

// --- Kategori --------------------------------------------------------------

func (s *ProductService) CreateCategory(ctx context.Context, name string, description *string) (*domain.ProductCategory, error) {
	name = strings.TrimSpace(name)
	return s.products.CreateCategory(ctx, s.pool, name, slugify(name), trimPtr(description))
}

func (s *ProductService) ListCategories(ctx context.Context) ([]domain.ProductCategory, error) {
	return s.products.ListCategories(ctx, s.pool)
}

func (s *ProductService) UpdateCategory(ctx context.Context, id uuid.UUID, name string, description *string) (*domain.ProductCategory, error) {
	name = strings.TrimSpace(name)
	return s.products.UpdateCategory(ctx, s.pool, id, name, slugify(name), trimPtr(description))
}

func (s *ProductService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.products.DeleteCategory(ctx, s.pool, id)
}

// --- Produk ----------------------------------------------------------------

type ProductInput struct {
	SKU          string
	Name         string
	CategoryID   *uuid.UUID
	Brand        *string
	StoreName    *string
	BaseCurrency string
	BasePrice    decimal.Decimal
	MarkupType   string
	MarkupValue  decimal.Decimal
	WeightGram   int
	ImageURL     *string
	Notes        *string
	IsActive     bool
}

func (s *ProductService) Create(ctx context.Context, in ProductInput) (*domain.Product, error) {
	params, err := s.toParams(in)
	if err != nil {
		return nil, err
	}
	return s.products.Create(ctx, s.pool, *params)
}

func (s *ProductService) List(ctx context.Context, p pagination.Params, categoryID *uuid.UUID, activeOnly bool) ([]domain.ProductWithCategory, int64, error) {
	return s.products.List(ctx, s.pool, p, categoryID, activeOnly)
}

func (s *ProductService) Get(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return s.products.GetByID(ctx, s.pool, id)
}

// PriceHistory mengembalikan riwayat harga produk dari trip ke trip. Dipakai
// admin saat menyusun katalog trip baru supaya harga modal tidak perlu digali
// ulang dari catatan lama.
func (s *ProductService) PriceHistory(ctx context.Context, id uuid.UUID) ([]domain.ProductPriceHistory, error) {
	// Produk diambil lebih dulu supaya id yang tidak ada menghasilkan 404, bukan
	// daftar kosong yang terlihat seperti produk tanpa riwayat.
	if _, err := s.products.GetByID(ctx, s.pool, id); err != nil {
		return nil, err
	}
	return s.products.PriceHistory(ctx, s.pool, id)
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, in ProductInput) (*domain.Product, error) {
	params, err := s.toParams(in)
	if err != nil {
		return nil, err
	}
	return s.products.Update(ctx, s.pool, id, *params)
}

// Delete menonaktifkan produk. Produk yang pernah dipakai transaksi tidak
// pernah benar-benar dihapus supaya riwayat order dan laporan tetap utuh.
func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.products.GetByID(ctx, s.pool, id); err != nil {
		return err
	}
	return s.products.SoftDelete(ctx, s.pool, id)
}

// PreviewPrice memperlihatkan harga jual yang akan terbentuk dari sebuah kurs
// dan markup, dipakai form katalog trip sebelum admin menyimpan.
func (s *ProductService) PreviewPrice(costForeign, exchangeRate decimal.Decimal, markupType string, markupValue decimal.Decimal) (costIDR, sellPrice decimal.Decimal, err error) {
	if !domain.IsValidMarkupType(markupType) {
		return decimal.Zero, decimal.Zero, domain.Validation("tipe markup tidak dikenal", map[string]string{
			"markup_type": "pilih percent atau nominal",
		})
	}
	if exchangeRate.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, decimal.Zero, domain.Validation("kurs tidak valid", map[string]string{
			"exchange_rate": "harus lebih besar dari 0",
		})
	}

	costIDR, sellPrice = domain.CalculateSellPrice(costForeign, exchangeRate, markupType, markupValue)
	return costIDR, sellPrice, nil
}

func (s *ProductService) toParams(in ProductInput) (*repository.ProductParams, error) {
	if !domain.IsValidMarkupType(in.MarkupType) {
		return nil, domain.Validation("tipe markup tidak dikenal", map[string]string{
			"markup_type": "pilih percent atau nominal",
		})
	}
	if in.BasePrice.IsNegative() || in.MarkupValue.IsNegative() {
		return nil, domain.Validation("nominal tidak boleh negatif", map[string]string{
			"base_price": "harus 0 atau lebih",
		})
	}

	currency := strings.ToUpper(strings.TrimSpace(in.BaseCurrency))
	if currency == "" {
		currency = "IDR"
	}

	sku := strings.ToUpper(strings.TrimSpace(in.SKU))
	if sku == "" {
		sku = generateSKU(in.Name)
	}

	return &repository.ProductParams{
		SKU:          sku,
		Name:         strings.TrimSpace(in.Name),
		CategoryID:   in.CategoryID,
		Brand:        trimPtr(in.Brand),
		StoreName:    trimPtr(in.StoreName),
		BaseCurrency: currency,
		BasePrice:    in.BasePrice,
		MarkupType:   in.MarkupType,
		MarkupValue:  in.MarkupValue,
		WeightGram:   in.WeightGram,
		ImageURL:     trimPtr(in.ImageURL),
		Notes:        trimPtr(in.Notes),
		IsActive:     in.IsActive,
	}, nil
}

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(v string) string {
	slug := nonSlugChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(v)), "-")
	return strings.Trim(slug, "-")
}

// generateSKU membuat SKU dari nama produk saat admin mengosongkannya.
// Sengaja pendek dan mudah dibaca supaya enak ditulis tangan di label paket.
func generateSKU(name string) string {
	slug := strings.ToUpper(slugify(name))
	slug = strings.ReplaceAll(slug, "-", "")
	if len(slug) > 8 {
		slug = slug[:8]
	}
	if slug == "" {
		slug = "PRD"
	}
	return slug + "-" + strings.ToUpper(uuid.NewString()[:4])
}
