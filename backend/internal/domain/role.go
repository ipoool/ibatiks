package domain

import (
	"regexp"
	"strings"
	"time"
)

// Role adalah sekumpulan menu yang boleh dibuka, disimpan sebagai data supaya
// toko bisa menyusun pembagian kerjanya sendiri.
//
// Sebelumnya role adalah daftar tertutup di dalam kode dan namanya dipakai
// langsung oleh penjaga rute. Akibatnya role bikinan sendiri akan ditolak
// seluruh endpoint operasional walaupun menunya sudah dicentang. Yang dibaca
// penjaga sekarang adalah Scope, yang ikut dimiliki role baru.
type Role struct {
	Name        string   `db:"name"        json:"name"`
	Label       string   `db:"label"       json:"label"`
	Description string   `db:"description" json:"description"`
	Scope       string   `db:"scope"       json:"scope"`
	Permissions []string `db:"permissions" json:"permissions"`
	IsSystem    bool     `db:"is_system"   json:"is_system"`
	// UserCount diisi lapisan repository supaya antarmuka bisa memberi tahu
	// berapa akun yang ikut terdampak sebelum sebuah role diubah atau dihapus.
	UserCount int       `db:"user_count" json:"user_count"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Nama role bawaan. Keempatnya tidak bisa dihapus maupun diganti namanya —
// kode merujuk nama-nama ini untuk penjagaan yang tidak bisa dinyatakan lewat
// daftar menu: owner terakhir tidak boleh hilang, dan root tidak boleh
// dipersempit.
const (
	RoleRoot    = "root"    // seluruh menu tanpa kecuali, termasuk yang ditambahkan kemudian
	RoleOwner   = "owner"   // pemilik toko: seluruh menu termasuk laba-rugi & pengguna
	RoleAdmin   = "admin"   // seluruh operasional harian
	RoleTripper = "tripper" // hanya daftar belanja dan input realisasi belanja
)

// Scope adalah batas kasar sebuah role, dipakai penjaga rute yang tidak bisa
// dinyatakan lewat daftar menu.
//
// Daftar menu menjawab "menu apa yang boleh dibuka", bukan "boleh mengubah
// isinya atau cuma melihat". Petugas lapangan perlu membuka menu Produk untuk
// membaca daftar belanjanya, tapi tidak boleh menyunting master produk — satu
// hal yang tidak bisa diwakili centang menu.
const (
	ScopeFull  = "full"  // staf toko: boleh mengubah data
	ScopeField = "field" // petugas lapangan: produk dan pembelian hanya baca
)

var AllScopes = []string{ScopeFull, ScopeField}

// FieldPermissions adalah menu yang benar-benar bisa dipakai petugas lapangan.
//
// Sisanya menuntut wewenang staf di tingkat rute, jadi mencentangnya untuk role
// lapangan hanya menghasilkan menu yang muncul di sidebar tapi halamannya
// ditolak — yang terbaca "belum ada data", seolah tokonya kosong.
var FieldPermissions = []string{
	PermTrips, PermShoppingList, PermPurchases, PermProducts,
}

func IsFieldPermission(p string) bool {
	for _, valid := range FieldPermissions {
		if valid == p {
			return true
		}
	}
	return false
}

func IsValidScope(scope string) bool {
	return scope == ScopeFull || scope == ScopeField
}

// SystemRoles adalah role bawaan yang selalu ada.
var SystemRoles = []string{RoleRoot, RoleOwner, RoleAdmin, RoleTripper}

// IsRootRole menandai role jalan pulih.
//
// Root sengaja tidak tampil di daftar role dan akunnya tidak bisa disentuh
// siapa pun selain root sendiri. Kalau ia terlihat dan bisa disunting, ia
// berhenti jadi jalan pulih: siapa pun yang memegang menu Pengguna bisa
// menurunkan rolenya atau menghapus akunnya, dan setelah itu hak akses yang
// terlanjur salah disetel tidak punya jalan kembali selain menyunting database.
func IsRootRole(name string) bool {
	return name == RoleRoot
}

func IsSystemRole(name string) bool {
	for _, r := range SystemRoles {
		if r == name {
			return true
		}
	}
	return false
}

// roleNamePattern membatasi nama role ke huruf kecil, angka, dan garis bawah.
//
// Namanya ikut tersimpan di users.role dan dibaca kode sebagai pengenal, bukan
// sebagai tulisan yang dibaca orang — itu tugas Label. Membiarkan spasi dan
// huruf besar berarti "Kasir" dan "kasir " jadi dua role berbeda yang terlihat
// sama di layar.
var roleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,30}$`)

func IsValidRoleName(name string) bool {
	return roleNamePattern.MatchString(name)
}

// SlugRoleName menurunkan nama pengenal dari label yang diketik pengguna,
// supaya tim toko cukup mengetik "Kasir" dan tidak perlu tahu soal slug.
func SlugRoleName(label string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(label)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '-', r == '_':
			b.WriteRune('_')
		}
	}
	name := strings.Trim(b.String(), "_")
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	// Nama wajib diawali huruf; label yang diawali angka tetap harus jadi nama
	// yang sah, bukan ditolak dengan alasan yang tidak dimengerti pengguna.
	if name != "" && (name[0] < 'a' || name[0] > 'z') {
		name = "role_" + name
	}
	if len(name) > 31 {
		name = strings.Trim(name[:31], "_")
	}
	return name
}

// LockedPermissions adalah menu yang tidak bisa dicabut dari sebuah role.
//
// Root memegang seluruhnya: ia adalah jalan pulih terakhir ketika hak akses
// siapa pun terlanjur salah disetel, dan jalan pulih yang bisa dipersempit
// bukan jalan pulih. Owner memegang Pengaturan dan Pengguna karena keduanya
// satu-satunya pintu untuk mengembalikan hak akses — sekali hilang, satu-
// satunya pemulihan adalah menyunting database langsung.
func LockedPermissions(role string) []string {
	switch role {
	case RoleRoot:
		return append([]string(nil), AllPermissions...)
	case RoleOwner:
		return []string{PermSettings, PermUsers}
	default:
		return nil
	}
}
