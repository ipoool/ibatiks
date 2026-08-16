package domain

// Hak akses per menu.
//
// Role tetap jadi bawaan, tapi owner bisa menyetel daftar ini per pengguna
// ketika pembagian kerja di tokonya tidak persis mengikuti tiga role bawaan.
// Nilainya sengaja mengikuti nama menu di aplikasi, bukan nama tabel, karena
// yang diatur owner adalah "menu apa yang boleh dipakai", bukan struktur data.
const (
	PermTrips        = "trips"
	PermShoppingList = "shopping_list"
	PermPurchases    = "purchases"
	PermOrders       = "orders"
	PermInvoices     = "invoices"
	PermPacking      = "packing"
	PermShipments    = "shipments"
	PermCustomers    = "customers"
	PermProducts     = "products"
	PermStock        = "stock"
	PermReports      = "reports"
	PermSettings     = "settings"
	PermUsers        = "users"
)

// AllPermissions dipakai antarmuka pengaturan untuk menampilkan pilihannya.
var AllPermissions = []string{
	PermTrips, PermShoppingList, PermPurchases,
	PermOrders, PermInvoices, PermPacking, PermShipments,
	PermCustomers, PermProducts, PermStock,
	PermReports, PermSettings, PermUsers,
}

func IsValidPermission(p string) bool {
	for _, valid := range AllPermissions {
		if valid == p {
			return true
		}
	}
	return false
}

// DefaultPermissions adalah hak akses bawaan sebuah role.
func DefaultPermissions(role string) []string {
	switch role {
	case RoleOwner:
		return append([]string(nil), AllPermissions...)
	case RoleAdmin:
		return []string{
			PermTrips, PermShoppingList, PermPurchases,
			PermOrders, PermInvoices, PermPacking, PermShipments,
			PermCustomers, PermProducts, PermStock, PermReports,
		}
	case RoleTripper:
		// Tripper bekerja di lapangan: melihat trip, daftar belanjaan, dan
		// mencatat pembeliannya. Produk ikut dibuka karena daftar belanja
		// merujuk ke sana.
		return []string{PermTrips, PermShoppingList, PermPurchases, PermProducts}
	default:
		return nil
	}
}

// OwnerLockedPermissions adalah menu yang tidak bisa dicabut dari owner.
//
// Pengaturan dan Pengguna adalah satu-satunya jalan untuk mengembalikan hak
// akses siapa pun, termasuk hak owner itu sendiri. Tanpa penjagaan ini, satu
// centang yang terlepas mengunci owner keluar dari tokonya sendiri dan satu-
// satunya jalan pulih adalah menyunting database langsung.
var OwnerLockedPermissions = []string{PermSettings, PermUsers}

// EffectivePermissions menggabungkan daftar khusus pengguna dengan bawaan role.
//
// Daftar kosong berarti belum pernah disetel, jadi yang dipakai bawaan role.
// Pengaturan khusus tidak boleh melampaui role: seorang tripper tetap tidak
// bisa diberi menu pengaturan lewat centang, sebab batas role adalah keputusan
// keamanan sedangkan centang hanyalah penyempitan.
//
// Owner adalah pengecualiannya ke arah sebaliknya: penyempitan tidak boleh
// sampai mencabut OwnerLockedPermissions, supaya tokonya tidak pernah kehilangan
// pemiliknya.
func EffectivePermissions(role string, custom []string) []string {
	defaults := DefaultPermissions(role)
	if len(custom) == 0 {
		return defaults
	}

	allowed := make(map[string]struct{}, len(defaults))
	for _, p := range defaults {
		allowed[p] = struct{}{}
	}

	out := make([]string, 0, len(custom))
	seen := make(map[string]struct{}, len(custom))
	for _, p := range custom {
		if _, ok := allowed[p]; !ok {
			continue
		}
		if _, sudah := seen[p]; sudah {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	if role == RoleOwner {
		for _, p := range OwnerLockedPermissions {
			if _, sudah := seen[p]; !sudah {
				seen[p] = struct{}{}
				out = append(out, p)
			}
		}
	}
	return out
}

// HasPermission memeriksa satu hak akses pada daftar yang sudah efektif.
func HasPermission(permissions []string, want string) bool {
	for _, p := range permissions {
		if p == want {
			return true
		}
	}
	return false
}
