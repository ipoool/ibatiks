package domain_test

import (
	"testing"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

// Daftar menu role, sekarang datang dari tabel roles. Di dalam uji ini ditulis
// tangan supaya yang diuji tetap aturan penggabungannya, bukan isi database.
var (
	menuAdmin = []string{
		domain.PermTrips, domain.PermShoppingList, domain.PermPurchases,
		domain.PermOrders, domain.PermInvoices, domain.PermShipments,
		domain.PermCustomers, domain.PermProducts, domain.PermStock,
		domain.PermReports,
	}
	menuTripper = []string{
		domain.PermTrips, domain.PermShoppingList,
		domain.PermPurchases, domain.PermProducts,
	}
	menuPenuh = domain.AllPermissions
)

func TestEffectivePermissions(t *testing.T) {
	t.Run("daftar kosong memakai seluruh menu role", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleAdmin, menuAdmin, nil)
		if len(got) != len(menuAdmin) {
			t.Fatalf("ingin seluruh menu admin, dapat %v", got)
		}
	})

	t.Run("centang hanya bisa mempersempit", func(t *testing.T) {
		// Tripper tidak punya menu laporan, jadi memintanya lewat daftar khusus
		// pun tidak boleh membukanya.
		got := domain.EffectivePermissions(domain.RoleTripper, menuTripper,
			[]string{domain.PermShoppingList, domain.PermReports, domain.PermSettings})

		if domain.HasPermission(got, domain.PermReports) || domain.HasPermission(got, domain.PermSettings) {
			t.Fatalf("hak di luar role ikut terbuka: %v", got)
		}
		if !domain.HasPermission(got, domain.PermShoppingList) {
			t.Fatalf("hak yang sah malah hilang: %v", got)
		}
	})

	t.Run("owner memegang seluruh hak", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleOwner, menuPenuh, nil)
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

		got := domain.EffectivePermissions(domain.RoleOwner, menuPenuh, dipersempit)

		for _, p := range domain.LockedPermissions(domain.RoleOwner) {
			if !domain.HasPermission(got, p) {
				t.Fatalf("owner terkunci keluar: hak %q ikut tercabut, hasil %v", p, got)
			}
		}
	})

	t.Run("owner tetap bisa mempersempit menu selain yang dikunci", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleOwner, menuPenuh,
			[]string{domain.PermOrders, domain.PermSettings, domain.PermUsers})

		if domain.HasPermission(got, domain.PermStock) {
			t.Fatalf("penyempitan tidak berlaku, menu stok masih terbuka: %v", got)
		}
		if !domain.HasPermission(got, domain.PermOrders) {
			t.Fatalf("hak yang dicentang malah hilang: %v", got)
		}
	})

	// Root adalah jalan pulih terakhir ketika hak akses siapa pun terlanjur
	// salah disetel. Jalan pulih yang bisa dipersempit bukan jalan pulih.
	t.Run("root tidak bisa dipersempit sama sekali", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleRoot, menuPenuh,
			[]string{domain.PermOrders})

		for _, p := range domain.AllPermissions {
			if !domain.HasPermission(got, p) {
				t.Fatalf("root kehilangan hak %q, hasil %v", p, got)
			}
		}
	})

	// Role bikinan toko tidak punya menu yang dikunci: kalau seluruh centangnya
	// dilepas, memang tidak ada menu yang tersisa.
	t.Run("role bikinan sendiri tidak punya menu terkunci", func(t *testing.T) {
		got := domain.EffectivePermissions("kasir",
			[]string{domain.PermOrders, domain.PermInvoices},
			[]string{domain.PermOrders})

		if len(got) != 1 || got[0] != domain.PermOrders {
			t.Fatalf("ingin hanya menu order, dapat %v", got)
		}
	})

	t.Run("hak yang dicentang dua kali tidak jadi ganda", func(t *testing.T) {
		got := domain.EffectivePermissions(domain.RoleAdmin, menuAdmin,
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

func TestSlugRoleName(t *testing.T) {
	cases := map[string]string{
		"Kasir":            "kasir",
		"Admin Gudang":     "admin_gudang",
		"  Kepala  Toko  ": "kepala_toko",
		"Tim CS 2":         "tim_cs_2",
		// Nama wajib diawali huruf; label yang diawali angka tetap harus jadi
		// nama yang sah, bukan ditolak dengan alasan yang tidak dimengerti.
		"2nd Shift": "role_2nd_shift",
	}

	for label, want := range cases {
		got := domain.SlugRoleName(label)
		if got != want {
			t.Fatalf("SlugRoleName(%q) = %q, ingin %q", label, got, want)
		}
		if !domain.IsValidRoleName(got) {
			t.Fatalf("SlugRoleName(%q) = %q, bukan nama role yang sah", label, got)
		}
	}
}

// LegacyScope menebak wewenang token yang terbit sebelum scope ada. Tebakannya
// harus condong ke yang paling sempit: role tak dikenal diperlakukan sebagai
// petugas lapangan, bukan diberi wewenang penuh selama sisa umur token.
func TestLegacyScopeCondongKeYangPalingSempit(t *testing.T) {
	if got := domain.LegacyScope("kasir"); got != domain.ScopeField {
		t.Fatalf("role tak dikenal dapat scope %q, ingin %q", got, domain.ScopeField)
	}
	if got := domain.LegacyScope(domain.RoleTripper); got != domain.ScopeField {
		t.Fatalf("tripper dapat scope %q, ingin %q", got, domain.ScopeField)
	}
	if got := domain.LegacyScope(domain.RoleOwner); got != domain.ScopeFull {
		t.Fatalf("owner dapat scope %q, ingin %q", got, domain.ScopeFull)
	}
}
