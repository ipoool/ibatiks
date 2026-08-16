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
}
