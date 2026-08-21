package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

// RoleService mengelola role beserta daftar menunya.
//
// Aturan yang dijaga di sini bukan soal kerapian data, melainkan soal supaya
// toko tidak pernah kehilangan jalan masuk ke pengaturannya sendiri: role
// bawaan tidak bisa dihapus, root tidak bisa dipersempit, dan role yang masih
// dipakai orang tidak bisa dibuang begitu saja.
type RoleService struct {
	pool  *pgxpool.Pool
	roles *repository.RoleRepo
}

func NewRoleService(pool *pgxpool.Pool, roles *repository.RoleRepo) *RoleService {
	return &RoleService{pool: pool, roles: roles}
}

// List menyembunyikan role root dari siapa pun selain root sendiri.
func (s *RoleService) List(ctx context.Context, pemintaRole string) ([]domain.Role, error) {
	return s.roles.List(ctx, s.pool, domain.IsRootRole(pemintaRole))
}

func (s *RoleService) Get(ctx context.Context, name, pemintaRole string) (*domain.Role, error) {
	if domain.IsRootRole(name) && !domain.IsRootRole(pemintaRole) {
		return nil, domain.NotFound("role")
	}
	return s.roles.Get(ctx, s.pool, name)
}

type SaveRoleInput struct {
	Label       string
	Description string
	Scope       string
	Permissions []string
}

func (s *RoleService) Create(ctx context.Context, in SaveRoleInput) (*domain.Role, error) {
	label := strings.TrimSpace(in.Label)
	if len([]rune(label)) < 2 {
		return nil, domain.Validation("nama role terlalu pendek", map[string]string{
			"label": "isi minimal dua huruf",
		})
	}

	// Nama pengenalnya diturunkan dari label, jadi tim toko cukup mengetik
	// "Kasir" dan tidak perlu tahu soal slug. Yang tersimpan di users.role
	// adalah nama ini.
	name := domain.SlugRoleName(label)
	if !domain.IsValidRoleName(name) {
		return nil, domain.Validation("nama role tidak bisa dipakai", map[string]string{
			"label": "pakai huruf dan angka, misalnya \"Kasir\" atau \"Admin Gudang\"",
		})
	}
	// Root tidak boleh dilahirkan dari sini. Ia jalan pulih yang jumlahnya
	// harus tetap satu dan lahir dari migrasi; membiarkan siapa pun yang
	// memegang menu Pengguna membuat role bernama root berarti membuat jalan
	// pulih kedua yang tidak diketahui pemilik toko.
	if domain.IsSystemRole(name) {
		return nil, domain.Conflict("nama role %s sudah dipakai, pilih nama lain", label)
	}

	scope, err := validScope(in.Scope)
	if err != nil {
		return nil, err
	}
	permissions, err := validRolePermissions(scope, in.Permissions)
	if err != nil {
		return nil, err
	}

	return s.roles.Create(ctx, s.pool, repository.SaveRoleParams{
		Name:        name,
		Label:       label,
		Description: strings.TrimSpace(in.Description),
		Scope:       scope,
		Permissions: permissions,
	})
}

func (s *RoleService) Update(ctx context.Context, name, pemintaRole string, in SaveRoleInput) (*domain.Role, error) {
	if domain.IsRootRole(name) && !domain.IsRootRole(pemintaRole) {
		return nil, domain.NotFound("role")
	}

	current, err := s.roles.Get(ctx, s.pool, name)
	if err != nil {
		return nil, err
	}

	label := strings.TrimSpace(in.Label)
	if len([]rune(label)) < 2 {
		return nil, domain.Validation("nama role terlalu pendek", map[string]string{
			"label": "isi minimal dua huruf",
		})
	}

	scope, err := validScope(in.Scope)
	if err != nil {
		return nil, err
	}
	permissions, err := validRolePermissions(scope, in.Permissions)
	if err != nil {
		return nil, err
	}

	// Root adalah jalan pulih terakhir ketika hak akses siapa pun terlanjur
	// salah disetel. Jalan pulih yang bisa dipersempit bukan jalan pulih, jadi
	// daftar menunya selalu utuh dan wewenangnya selalu penuh.
	if current.Name == domain.RoleRoot {
		permissions = append([]string(nil), domain.AllPermissions...)
		scope = domain.ScopeFull
	}

	// Owner wajib tetap memegang Pengaturan dan Pengguna: keduanya satu-satunya
	// pintu untuk mengembalikan hak akses siapa pun. Ditambahkan diam-diam,
	// bukan ditolak — yang dimaksud orang saat melepas centangnya adalah
	// mempersempit, bukan mengunci dirinya sendiri di luar.
	permissions = domain.EffectivePermissions(current.Name, permissions, nil)
	if current.Name == domain.RoleOwner {
		scope = domain.ScopeFull
	}

	updated, err := s.roles.Update(ctx, s.pool, name, repository.SaveRoleParams{
		Label:       label,
		Description: strings.TrimSpace(in.Description),
		Scope:       scope,
		Permissions: permissions,
	})
	if err != nil {
		return nil, err
	}

	// Hak akses ikut dibawa di dalam access token, jadi perubahannya baru
	// terasa saat token berikutnya terbit. Sesi seluruh pemakai role ini
	// dicabut supaya pembatasannya berlaku saat itu juga.
	if !samePermissions(current.Permissions, updated.Permissions) || current.Scope != updated.Scope {
		if err := s.roles.RevokeSessionsByRole(ctx, s.pool, name); err != nil {
			return nil, err
		}
	}
	return updated, nil
}

func (s *RoleService) Delete(ctx context.Context, name, pemintaRole string) error {
	if domain.IsRootRole(name) && !domain.IsRootRole(pemintaRole) {
		return domain.NotFound("role")
	}

	role, err := s.roles.Get(ctx, s.pool, name)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return domain.Conflict("role bawaan tidak bisa dihapus")
	}

	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		// users.role menunjuk ke sini dengan ON DELETE RESTRICT, jadi database
		// akan menolaknya sendiri. Hitungannya diambil lebih dulu supaya
		// pesannya menyebutkan berapa akun yang menghalangi, bukan sekadar
		// "melanggar batasan".
		used, err := s.roles.CountUsers(ctx, tx, name)
		if err != nil {
			return err
		}
		if used > 0 {
			return domain.Conflict(
				"role ini masih dipakai %d akun, pindahkan dulu ke role lain", used)
		}
		return s.roles.Delete(ctx, tx, name)
	})
}

func validScope(scope string) (string, error) {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return domain.ScopeFull, nil
	}
	if !domain.IsValidScope(scope) {
		return "", domain.Validation("tingkat akses tidak dikenal", map[string]string{
			"scope": "pilih staf toko atau petugas lapangan",
		})
	}
	return scope, nil
}

func validRolePermissions(scope string, requested []string) ([]string, error) {
	out := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	var takBisaLapangan []string

	for _, p := range requested {
		if !domain.IsValidPermission(p) {
			return nil, domain.Validation("hak akses tidak dikenal", map[string]string{
				"permissions": p + " bukan menu yang ada di aplikasi",
			})
		}
		// Menu di luar daftar lapangan menuntut wewenang staf di tingkat rute.
		// Membiarkannya tercentang berarti menu yang muncul di sidebar tapi
		// halamannya ditolak — dan yang terbaca "belum ada data", seolah
		// tokonya kosong.
		if scope == domain.ScopeField && !domain.IsFieldPermission(p) {
			takBisaLapangan = append(takBisaLapangan, p)
			continue
		}
		if _, sudah := seen[p]; sudah {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	if len(takBisaLapangan) > 0 {
		return nil, domain.Validation("menu itu tidak bisa dipakai petugas lapangan", map[string]string{
			"permissions": strings.Join(takBisaLapangan, ", ") +
				" menuntut wewenang staf toko — ganti tingkat aksesnya atau lepas centangnya",
		})
	}
	if len(out) == 0 {
		return nil, domain.Validation("role tanpa menu tidak bisa dipakai", map[string]string{
			"permissions": "centang minimal satu menu",
		})
	}
	return out, nil
}

func samePermissions(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]struct{}, len(a))
	for _, p := range a {
		seen[p] = struct{}{}
	}
	for _, p := range b {
		if _, ok := seen[p]; !ok {
			return false
		}
	}
	return true
}
