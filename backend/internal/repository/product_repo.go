package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
)

const productColumns = `id, sku, name, category_id, brand, store_name, base_currency, base_price,
	                    markup_type, markup_value, weight_gram, image_url, notes, is_active,
	                    created_at, updated_at, deleted_at`

// Alias p dipakai supaya kolom produk tidak bentrok dengan kolom kategori
// saat di-join.
const productColumnsPrefixed = `p.id, p.sku, p.name, p.category_id, p.brand, p.store_name,
	                            p.base_currency, p.base_price, p.markup_type, p.markup_value,
	                            p.weight_gram, p.image_url, p.notes, p.is_active,
	                            p.created_at, p.updated_at, p.deleted_at`

type ProductRepo struct{}

func NewProductRepo() *ProductRepo { return &ProductRepo{} }

// --- Kategori --------------------------------------------------------------

const categoryColumns = `id, name, slug, description, created_at, updated_at`

func (r *ProductRepo) CreateCategory(ctx context.Context, q db.Querier, name, slug string, description *string) (*domain.ProductCategory, error) {
	return collectOne[domain.ProductCategory](ctx, q, "kategori", `
		INSERT INTO product_categories (name, slug, description)
		VALUES ($1, $2, $3)
		RETURNING `+categoryColumns, name, slug, description)
}

func (r *ProductRepo) GetCategory(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.ProductCategory, error) {
	return collectOne[domain.ProductCategory](ctx, q, "kategori",
		`SELECT `+categoryColumns+` FROM product_categories WHERE id = $1`, id)
}

func (r *ProductRepo) ListCategories(ctx context.Context, q db.Querier) ([]domain.ProductCategory, error) {
	return collectRows[domain.ProductCategory](ctx, q,
		`SELECT `+categoryColumns+` FROM product_categories ORDER BY name ASC`)
}

func (r *ProductRepo) UpdateCategory(ctx context.Context, q db.Querier, id uuid.UUID, name, slug string, description *string) (*domain.ProductCategory, error) {
	return collectOne[domain.ProductCategory](ctx, q, "kategori", `
		UPDATE product_categories SET name = $2, slug = $3, description = $4
		WHERE id = $1
		RETURNING `+categoryColumns, id, name, slug, description)
}

func (r *ProductRepo) DeleteCategory(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "kategori", `DELETE FROM product_categories WHERE id = $1`, id)
}

// --- Produk ----------------------------------------------------------------

type ProductParams struct {
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

func (r *ProductRepo) Create(ctx context.Context, q db.Querier, p ProductParams) (*domain.Product, error) {
	return collectOne[domain.Product](ctx, q, "produk", `
		INSERT INTO products (sku, name, category_id, brand, store_name, base_currency, base_price,
		                      markup_type, markup_value, weight_gram, image_url, notes, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING `+productColumns,
		p.SKU, p.Name, p.CategoryID, p.Brand, p.StoreName, p.BaseCurrency, p.BasePrice,
		p.MarkupType, p.MarkupValue, p.WeightGram, p.ImageURL, p.Notes, p.IsActive)
}

func (r *ProductRepo) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Product, error) {
	return collectOne[domain.Product](ctx, q, "produk",
		`SELECT `+productColumns+` FROM products WHERE id = $1 AND deleted_at IS NULL`, id)
}

func (r *ProductRepo) List(ctx context.Context, q db.Querier, p pagination.Params, categoryID *uuid.UUID, activeOnly bool) ([]domain.ProductWithCategory, int64, error) {
	conditions := []string{"p.deleted_at IS NULL"}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf(
			"(p.name ILIKE $%d OR p.sku ILIKE $%d OR COALESCE(p.brand, '') ILIKE $%d)", n, n, n))
	}
	if categoryID != nil {
		args = append(args, *categoryID)
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", len(args)))
	}
	if activeOnly {
		conditions = append(conditions, "p.is_active = TRUE")
	}
	where := buildWhere(conditions)

	var total int64
	if err := q.QueryRow(ctx, `SELECT count(*) FROM products p`+where, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"name":       "p.name",
		"sku":        "p.sku",
		"base_price": "p.base_price",
		"created_at": "p.created_at",
	}, "p.created_at")

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`
		SELECT %s, c.name AS category_name
		FROM products p
		LEFT JOIN product_categories c ON c.id = p.category_id%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		productColumnsPrefixed, where, sortCol, p.Order, len(args)-1, len(args))

	products, err := collectRows[domain.ProductWithCategory](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *ProductRepo) Update(ctx context.Context, q db.Querier, id uuid.UUID, p ProductParams) (*domain.Product, error) {
	return collectOne[domain.Product](ctx, q, "produk", `
		UPDATE products
		SET sku = $2, name = $3, category_id = $4, brand = $5, store_name = $6,
		    base_currency = $7, base_price = $8, markup_type = $9, markup_value = $10,
		    weight_gram = $11, image_url = $12, notes = $13, is_active = $14
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+productColumns,
		id, p.SKU, p.Name, p.CategoryID, p.Brand, p.StoreName, p.BaseCurrency, p.BasePrice,
		p.MarkupType, p.MarkupValue, p.WeightGram, p.ImageURL, p.Notes, p.IsActive)
}

func (r *ProductRepo) SoftDelete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "produk",
		`UPDATE products SET deleted_at = now(), is_active = FALSE
		 WHERE id = $1 AND deleted_at IS NULL`, id)
}

// CountUsage menghitung berapa transaksi yang masih memakai produk ini, dipakai
// untuk memperingatkan admin sebelum menghapus.
func (r *ProductRepo) CountUsage(ctx context.Context, q db.Querier, id uuid.UUID) (int, error) {
	var count int
	err := q.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM order_items WHERE product_id = $1)
		     + (SELECT count(*) FROM trip_items  WHERE product_id = $1)
		     + (SELECT count(*) FROM purchases   WHERE product_id = $1)`, id).Scan(&count)
	if err != nil {
		return 0, wrapPgError(err)
	}
	return count, nil
}

// PriceHistory mengumpulkan riwayat harga sebuah produk dari trip ke trip.
//
// Diambil dari dua sisi sekaligus: harga yang dipasang di katalog trip dan
// harga yang benar-benar dibayar tripper. Keduanya perlu, karena selisihnya
// justru yang paling berguna saat menyusun harga trip berikutnya.
func (r *ProductRepo) PriceHistory(ctx context.Context, q db.Querier, productID uuid.UUID) ([]domain.ProductPriceHistory, error) {
	return collectRows[domain.ProductPriceHistory](ctx, q, `
		SELECT
			t.id            AS trip_id,
			t.code          AS trip_code,
			t.title         AS trip_title,
			t.country       AS country,
			t.depart_date   AS depart_date,
			t.currency      AS currency,
			t.exchange_rate AS exchange_rate,
			COALESCE(ti.cost_price, 0)     AS catalog_cost,
			COALESCE(ti.cost_price_idr, 0) AS catalog_cost_idr,
			COALESCE(ti.sell_price, 0)     AS sell_price,
			COALESCE(pu.avg_cost, 0)       AS actual_cost,
			COALESCE(pu.avg_cost_idr, 0)   AS actual_cost_idr,
			COALESCE(pu.qty, 0)            AS qty_purchased,
			COALESCE(so.qty_sold, 0)       AS qty_sold
		FROM trips t
		LEFT JOIN trip_items ti ON ti.trip_id = t.id AND ti.product_id = $1
		LEFT JOIN LATERAL (
			SELECT
				sum(p.qty)::int AS qty,
				sum(p.unit_cost_foreign * p.qty) / NULLIF(sum(p.qty), 0) AS avg_cost,
				sum(p.total_cost_idr)            / NULLIF(sum(p.qty), 0) AS avg_cost_idr
			FROM purchases p
			WHERE p.trip_id = t.id AND p.product_id = $1
		) pu ON TRUE
		LEFT JOIN LATERAL (
			SELECT sum(oi.qty)::int AS qty_sold
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE o.trip_id = t.id
			  AND oi.product_id = $1
			  AND o.status <> 'cancelled'
		) so ON TRUE
		-- Trip yang sama sekali tidak menyentuh produk ini tidak perlu muncul.
		WHERE ti.id IS NOT NULL OR COALESCE(pu.qty, 0) > 0
		ORDER BY t.depart_date DESC, t.code DESC`, productID)
}
