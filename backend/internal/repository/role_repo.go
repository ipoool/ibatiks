package repository

import (
	"context"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
)

type RoleRepo struct{}

func NewRoleRepo() *RoleRepo { return &RoleRepo{} }

const roleColumns = `name, label, description, scope, permissions, is_system,
	created_at, updated_at`

// listQuery ikut menghitung jumlah pengguna tiap role, supaya antarmuka bisa
// memberi tahu berapa akun yang terdampak sebelum sebuah role diubah atau
// dihapus.
const roleListQuery = `
	SELECT r.name, r.label, r.description, r.scope, r.permissions, r.is_system,
	       r.created_at, r.updated_at,
	       (SELECT count(*) FROM users u WHERE u.role = r.name)::int AS user_count
	FROM roles r`

// List mengurutkan role bawaan lebih dulu, lalu role bikinan toko menurut
// namanya. Urutan ini yang dilihat orang saat memilih role sebuah akun, dan
// role bawaan adalah yang paling sering dipakai.
//
// sertakanRoot hanya benar untuk pemakai root sendiri; lihat domain.IsRootRole.
func (r *RoleRepo) List(ctx context.Context, q db.Querier, sertakanRoot bool) ([]domain.Role, error) {
	return collectRows[domain.Role](ctx, q,
		roleListQuery+` WHERE ($1 OR r.name <> $2) ORDER BY r.is_system DESC, r.label ASC`,
		sertakanRoot, domain.RoleRoot)
}

func (r *RoleRepo) Get(ctx context.Context, q db.Querier, name string) (*domain.Role, error) {
	return collectOne[domain.Role](ctx, q, "role",
		roleListQuery+` WHERE r.name = $1`, name)
}

// Map mengembalikan seluruh role dikunci namanya.
//
// Dipakai saat menghitung hak akses efektif sekumpulan pengguna sekaligus:
// tanpa ini, daftar pengguna akan menembak satu query per baris hanya untuk
// mengambil daftar menu role yang itu-itu juga.
func (r *RoleRepo) Map(ctx context.Context, q db.Querier) (map[string]domain.Role, error) {
	rows, err := collectRowsLax[domain.Role](ctx, q,
		`SELECT `+roleColumns+` FROM roles`)
	if err != nil {
		return nil, err
	}

	out := make(map[string]domain.Role, len(rows))
	for _, role := range rows {
		out[role.Name] = role
	}
	return out, nil
}

type SaveRoleParams struct {
	Name        string
	Label       string
	Description string
	Scope       string
	Permissions []string
}

func (r *RoleRepo) Create(ctx context.Context, q db.Querier, p SaveRoleParams) (*domain.Role, error) {
	return collectOneLax[domain.Role](ctx, q, "role", `
		INSERT INTO roles (name, label, description, scope, permissions)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+roleColumns,
		p.Name, p.Label, p.Description, p.Scope, p.Permissions)
}

// Update sengaja tidak menyentuh kolom name: namanya tersimpan di users.role
// sebagai pengenal, dan menggantinya berarti memindahkan seluruh akun yang
// memakainya. Yang bisa diganti adalah labelnya — itu yang dibaca orang.
func (r *RoleRepo) Update(ctx context.Context, q db.Querier, name string, p SaveRoleParams) (*domain.Role, error) {
	return collectOneLax[domain.Role](ctx, q, "role", `
		UPDATE roles
		SET label = $2, description = $3, scope = $4, permissions = $5
		WHERE name = $1
		RETURNING `+roleColumns,
		name, p.Label, p.Description, p.Scope, p.Permissions)
}

func (r *RoleRepo) Delete(ctx context.Context, q db.Querier, name string) error {
	return execExpectOne(ctx, q, "role", `DELETE FROM roles WHERE name = $1`, name)
}

func (r *RoleRepo) CountUsers(ctx context.Context, q db.Querier, name string) (int, error) {
	var total int
	if err := q.QueryRow(ctx,
		`SELECT count(*) FROM users WHERE role = $1`, name).Scan(&total); err != nil {
		return 0, wrapPgError(err)
	}
	return total, nil
}

// RevokeSessionsByRole mencabut sesi seluruh pengguna yang memakai sebuah role.
//
// Hak akses ikut dibawa di dalam access token, jadi mengubah daftar menu sebuah
// role tidak terasa apa-apa sampai token berikutnya terbit. Untuk pencabutan
// hak yang mendesak, jeda itu persis yang tidak boleh ada.
func (r *RoleRepo) RevokeSessionsByRole(ctx context.Context, q db.Querier, name string) error {
	_, err := exec(ctx, q, `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE revoked_at IS NULL
		  AND user_id IN (SELECT id FROM users WHERE role = $1)`, name)
	return err
}
