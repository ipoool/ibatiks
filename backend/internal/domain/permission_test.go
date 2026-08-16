package domain_test

import (
	"testing"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

func TestEffectivePermissions(t *testing.T) {
	t.Run("daftar kosong memakai bawaan role", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleAdmin, nil)
		if len(got) != len(domain.DefaultPermissions(domain.RoleAdmin)) {
			t.Fatalf("ingin bawaan role admin, dapat %v", got)
		}
	})

	t.Run("centang hanya bisa mempersempit", func(t *testing.T) {
		// Tripper tidak punya hak laporan, jadi memintanya lewat daftar khusus
		// pun tidak boleh membukanya.
		got := domain.EffectivePermissions(domain.RoleTripper,
			[]string{domain.PermShoppingList, domain.PermReports, domain.PermSettings})

		if domain.HasPermission(got, domain.PermReports) || domain.HasPermission(got, domain.PermSettings) {
			t.Fatalf("hak di luar role ikut terbuka: %v", got)
		}
		if !domain.HasPermission(got, domain.PermShoppingList) {
			t.Fatalf("hak yang sah malah hilang: %v", got)
		}
	})

	t.Run("owner memegang seluruh hak", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleOwner, nil)
		for _, p := range domain.AllPermissions {
			if !domain.HasPermission(got, p) {
				t.Fatalf("owner kehilangan hak %q", p)
			}
		}
	})

	// Regresi: AKS-07 — owner mencabut centang Pengaturan dan Pengguna dari
	// dirinya sendiri, lalu terkunci keluar. Pengaturan dan Pengguna adalah
	// satu-satunya jalan mengembalikan hak akses; sekali hilang, satu-satunya
	// pemulihan adalah menyunting database langsung.
	// Ditemukan /qa 16 Agustus 2026.
	t.Run("owner tidak bisa mencabut menu pengaturan dari dirinya sendiri", func(t *testing.T) {
		// Owner mencentang semuanya kecuali Pengaturan dan Pengguna.
		var dipersempit []string
		for _, p := range domain.AllPermissions {
			if p == domain.PermSettings || p == domain.PermUsers {
				continue
			}
			dipersempit = append(dipersempit, p)
		}

		got := domain.EffectivePermissions(domain.RoleOwner, dipersempit)

		for _, p := range domain.OwnerLockedPermissions {
			if !domain.HasPermission(got, p) {
				t.Fatalf("owner terkunci keluar: hak %q ikut tercabut, hasil %v", p, got)
			}
		}
	})

	t.Run("owner tetap bisa mempersempit menu selain yang dikunci", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleOwner,
			[]string{domain.PermOrders, domain.PermSettings, domain.PermUsers})

		if domain.HasPermission(got, domain.PermStock) {
			t.Fatalf("penyempitan tidak berlaku, menu stok masih terbuka: %v", got)
		}
		if !domain.HasPermission(got, domain.PermOrders) {
			t.Fatalf("hak yang dicentang malah hilang: %v", got)
		}
	})

	t.Run("hak yang dicentang dua kali tidak jadi ganda", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleAdmin,
			[]string{domain.PermOrders, domain.PermOrders})

		var n int
		for _, p := range got {
			if p == domain.PermOrders {
				n++
			}
		}
		if n != 1 {
			t.Fatalf("hak orders muncul %d kali, ingin 1: %v", n, got)
		}
	})
}
