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

const customerColumns = `id, code, name, phone_wa, email, socials, address, city,
	                     district, subdistrict, province, postal_code, notes,
	                     created_at, updated_at, deleted_at`

type CustomerRepo struct{}

func NewCustomerRepo() *CustomerRepo { return &CustomerRepo{} }

type CustomerParams struct {
	Code        string
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

func (r *CustomerRepo) Create(ctx context.Context, q db.Querier, p CustomerParams) (*domain.Customer, error) {
	return collectOne[domain.Customer](ctx, q, "customer", `
		INSERT INTO customers (code, name, phone_wa, email, socials, address, city,
		                       district, subdistrict, province, postal_code, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING `+customerColumns,
		p.Code, p.Name, p.PhoneWA, p.Email, socialsJSON(p.Socials), p.Address, p.City,
		p.District, p.Subdistrict, p.Province, p.PostalCode, p.Notes)
}

func (r *CustomerRepo) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Customer, error) {
	return collectOne[domain.Customer](ctx, q, "customer",
		`SELECT `+customerColumns+` FROM customers WHERE id = $1 AND deleted_at IS NULL`, id)
}

// GetByIDIncludingDeleted mengambil customer tanpa peduli sudah dihapus atau
// belum. Dipakai dokumen historis seperti detail order: ordernya sendiri tidak
// ikut hilang saat customernya dihapus, jadi halamannya tetap harus bisa
// dibuka. Untuk daftar dan penyuntingan tetap pakai GetByID.
func (r *CustomerRepo) GetByIDIncludingDeleted(ctx context.Context, q db.Querier, id uuid.UUID) (*domain.Customer, error) {
	return collectOne[domain.Customer](ctx, q, "customer",
		`SELECT `+customerColumns+` FROM customers WHERE id = $1`, id)
}

func (r *CustomerRepo) List(ctx context.Context, q db.Querier, p pagination.Params) ([]domain.Customer, int64, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []any{}

	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		n := len(args)
		cocok := fmt.Sprintf(
			"name ILIKE $%d OR phone_wa ILIKE $%d OR code ILIKE $%d OR COALESCE(email, '') ILIKE $%d",
			n, n, n, n)

		// Nomor disimpan sudah ternormalkan (62812…), tapi admin mengetiknya
		// seperti yang tertera di WhatsApp customer (0812…). Tanpa baris ini,
		// mencari orang lewat nomornya sendiri tidak menemukan apa-apa.
		if domain.LooksLikePhone(p.Search) {
			if normal := domain.NormalizePhoneWA(p.Search); normal != "" {
				args = append(args, "%"+normal+"%")
				cocok += fmt.Sprintf(" OR phone_wa ILIKE $%d", len(args))
			}
		}

		conditions = append(conditions, "("+cocok+")")
	}
	where := buildWhere(conditions)

	var total int64
	if err := q.QueryRow(ctx, `SELECT count(*) FROM customers`+where, args...).Scan(&total); err != nil {
		return nil, 0, wrapPgError(err)
	}

	sortCol := pagination.SortColumn(p.Sort, map[string]string{
		"name":       "name",
		"code":       "code",
		"city":       "city",
		"created_at": "created_at",
	}, "created_at")

	args = append(args, p.Limit(), p.Offset())
	query := fmt.Sprintf(`SELECT %s FROM customers%s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		customerColumns, where, sortCol, p.Order, len(args)-1, len(args))

	customers, err := collectRows[domain.Customer](ctx, q, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}

func (r *CustomerRepo) Update(ctx context.Context, q db.Querier, id uuid.UUID, p CustomerParams) (*domain.Customer, error) {
	return collectOne[domain.Customer](ctx, q, "customer", `
		UPDATE customers
		SET name = $2, phone_wa = $3, email = $4, socials = $5, address = $6,
		    city = $7, district = $8, subdistrict = $9, province = $10,
		    postal_code = $11, notes = $12
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+customerColumns,
		id, p.Name, p.PhoneWA, p.Email, socialsJSON(p.Socials), p.Address, p.City,
		p.District, p.Subdistrict, p.Province, p.PostalCode, p.Notes)
}

// SoftDelete menandai customer terhapus tanpa menghilangkan datanya, karena
// order lama masih mereferensikan baris ini.
func (r *CustomerRepo) SoftDelete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	return execExpectOne(ctx, q, "customer",
		`UPDATE customers SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
}

// CountOrders dipakai sebelum menghapus, untuk memberi tahu admin bahwa
// customer ini masih punya riwayat pesanan.
func (r *CustomerRepo) CountOrders(ctx context.Context, q db.Querier, id uuid.UUID) (int, error) {
	var count int
	err := q.QueryRow(ctx, `SELECT count(*) FROM orders WHERE customer_id = $1`, id).Scan(&count)
	if err != nil {
		return 0, wrapPgError(err)
	}
	return count, nil
}

// CustomerStats adalah ringkasan riwayat belanja seorang customer.
type CustomerStats struct {
	TotalOrders int     `db:"total_orders" json:"total_orders"`
	TotalSpent  string  `db:"total_spent"  json:"total_spent"`
	LastOrderAt *string `db:"last_order_at" json:"last_order_at"`
}

func (r *CustomerRepo) Stats(ctx context.Context, q db.Querier, id uuid.UUID) (*CustomerStats, error) {
	return collectOne[CustomerStats](ctx, q, "customer", `
		SELECT
			count(*)::int                                      AS total_orders,
			COALESCE(sum(total) FILTER (WHERE status <> 'cancelled'), 0)::text AS total_spent,
			to_char(max(order_date), 'YYYY-MM-DD')             AS last_order_at
		FROM orders
		WHERE customer_id = $1`, id)
}

// socialsJSON menyiapkan daftar akun untuk kolom JSONB.
//
// Slice nil ditulis sebagai array kosong, bukan null: kolomnya NOT NULL, dan
// membiarkan null lolos berarti setiap pembacaan berikutnya harus mengurus dua
// bentuk kosong yang berbeda.
func socialsJSON(list []domain.Social) []byte {
	if list == nil {
		list = []domain.Social{}
	}
	data, err := json.Marshal(list)
	if err != nil {
		return []byte("[]")
	}
	return data
}
