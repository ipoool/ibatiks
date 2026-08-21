package domain

// Hak akses per menu.
//
// Nilainya sengaja mengikuti nama menu di aplikasi, bukan nama tabel, karena
// yang disusun tim toko adalah "menu apa yang boleh dipakai", bukan struktur
// data. Daftar menu sebuah role tinggal di tabel roles; centang per pengguna
// di users.permissions hanya boleh mempersempitnya.
const (
	PermTrips        = "trips"
	PermShoppingList = "shopping_list"
	PermPurchases    = "purchases"
	PermOrders       = "orders"
	PermInvoices     = "invoices"
	// PermShipments mencakup antrean kemas sekaligus daftar paket: keduanya
	// satu menu, dikerjakan orang yang sama di meja yang sama. Sebelumnya ada
	// PermPacking tersendiri, dan hak akses yang menyimpannya diterjemahkan
	// ke sini oleh migrasi 000017.
	PermShipments = "shipments"
	PermCustomers = "customers"
	PermProducts  = "products"
	PermStock     = "stock"
	PermReports   = "reports"
	// PermReportsFinance memisahkan angka laba-rugi dari laporan penjualan
	// biasa. Dulu penjagaannya menempel pada nama role ("khusus owner") dan
	// tidak punya centang sendiri; begitu role jadi data, penjagaan lewat nama
	// tidak punya pegangan lagi.
	//
	// Pemisahannya juga menutup lubang lama: admin punya PermReports, jadi tab
	// Profit / Loss ikut tampil untuknya, tapi endpoint-nya menolak — yang
	// terbaca adalah laba nol rupiah, bukan penolakan.
	PermReportsFinance = "reports_finance"
	PermSettings       = "settings"
	PermUsers          = "users"
)

// AllPermissions dipakai antarmuka pengaturan untuk menampilkan pilihannya,
// dan dipakai role root sebagai isinya.
var AllPermissions = []string{
	PermTrips, PermShoppingList, PermPurchases,
	PermOrders, PermInvoices, PermShipments,
	PermCustomers, PermProducts, PermStock,
	PermReports, PermReportsFinance,
	PermSettings, PermUsers,
}

func IsValidPermission(p string) bool {
	for _, valid := range AllPermissions {
		if valid == p {
			return true
		}
	}
	return false
}

// legacyRolePermissions adalah bawaan role sebelum role pindah ke database.
//
// Satu-satunya pemakainya adalah access token yang terbit sebelum perubahan
// ini dan belum kedaluwarsa. Token membawa hak aksesnya sendiri supaya
// middleware tidak perlu menyentuh database tiap request; token lama tidak
// membawa apa-apa, dan tanpa cadangan ini seluruh sesi yang sedang berjalan
// mendadak kehilangan seluruh menunya sampai tokennya diperbarui.
func legacyRolePermissions(role string) []string {
	switch role {
	case RoleRoot, RoleOwner:
		return append([]string(nil), AllPermissions...)
	case RoleAdmin:
		return []string{
			PermTrips, PermShoppingList, PermPurchases,
			PermOrders, PermInvoices, PermShipments,
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

// LegacyEffectivePermissions dipakai middleware untuk token lama yang belum
// membawa daftar hak aksesnya sendiri.
func LegacyEffectivePermissions(role string, custom []string) []string {
	return EffectivePermissions(role, legacyRolePermissions(role), custom)
}

// LegacyScope menebak scope dari nama role, untuk token yang terbit sebelum
// scope ada.
//
// Tebakannya harus condong ke yang paling sempit: role yang tidak dikenal
// diperlakukan sebagai petugas lapangan. Menebak ke arah sebaliknya berarti
// memberi wewenang penuh selama sisa umur token itu.
func LegacyScope(role string) string {
	switch role {
	case RoleRoot, RoleOwner, RoleAdmin:
		return ScopeFull
	default:
		return ScopeField
	}
}

// EffectivePermissions menggabungkan centang khusus pengguna dengan daftar menu
// milik rolenya.
//
// Daftar centang kosong berarti belum pernah disetel, jadi yang dipakai daftar
// rolenya utuh. Centang tidak boleh melampaui role: batas role adalah keputusan
// keamanan, sedangkan centang hanyalah penyempitan.
//
// Menu yang terkunci untuk sebuah role (lihat LockedPermissions) selalu ikut,
// supaya toko tidak pernah kehilangan pemiliknya dan root tetap jadi jalan
// pulih ketika hak akses siapa pun terlanjur salah disetel.
func EffectivePermissions(role string, rolePermissions, custom []string) []string {
	if len(custom) == 0 {
		return withLocked(role, rolePermissions)
	}

	allowed := make(map[string]struct{}, len(rolePermissions))
	for _, p := range rolePermissions {
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
	return withLocked(role, out)
}

func withLocked(role string, permissions []string) []string {
	locked := LockedPermissions(role)
	if len(locked) == 0 {
		return append([]string(nil), permissions...)
	}

	out := append([]string(nil), permissions...)
	seen := make(map[string]struct{}, len(out))
	for _, p := range out {
		seen[p] = struct{}{}
	}
	for _, p := range locked {
		if _, sudah := seen[p]; !sudah {
			seen[p] = struct{}{}
			out = append(out, p)
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
