package domain_test

import (
	"testing"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

func TestCanTransitionOrder(t *testing.T) {
	tests := []struct {
		from, to string
		want     bool
	}{
		// Alur normal satu pesanan dari awal sampai selesai. Hanya empat
		// perpindahan: DP masuk, pelunasan masuk, diserahkan ke kurir, tiba.
		{domain.OrderAwaitingDP, domain.OrderDPPaid, true},
		{domain.OrderDPPaid, domain.OrderPaid, true},
		{domain.OrderPaid, domain.OrderShipped, true},
		{domain.OrderShipped, domain.OrderCompleted, true},

		// Lompatan yang melewati tahap penting harus ditolak.
		{domain.OrderAwaitingDP, domain.OrderPaid, false},
		{domain.OrderAwaitingDP, domain.OrderShipped, false},
		{domain.OrderDPPaid, domain.OrderShipped, false},
		{domain.OrderDPPaid, domain.OrderCompleted, false},

		// Status akhir tidak bisa berpindah ke mana pun.
		{domain.OrderCompleted, domain.OrderShipped, false},
		{domain.OrderCompleted, domain.OrderCancelled, false},
		{domain.OrderCancelled, domain.OrderAwaitingDP, false},

		// Order yang sudah dikirim tidak bisa dibatalkan.
		{domain.OrderShipped, domain.OrderCancelled, false},
		// Begitu juga yang sudah lunas: uangnya sudah diterima penuh, jadi
		// pembatalannya urusan pengembalian dana, bukan ganti status.
		{domain.OrderPaid, domain.OrderCancelled, false},

		// Pembatalan diizinkan selama customer belum melunasi.
		{domain.OrderAwaitingDP, domain.OrderCancelled, true},
		{domain.OrderDPPaid, domain.OrderCancelled, true},

		// Status yang sudah dihapus tidak boleh membuka jalur apa pun.
		{"draft", domain.OrderAwaitingDP, false},
		{"packed", domain.OrderPaid, false},
		{domain.OrderDPPaid, "invoiced", false},

		// Status tak dikenal juga.
		{"entah", domain.OrderPaid, false},
		{domain.OrderAwaitingDP, "entah", false},
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			if got := domain.CanTransitionOrder(tt.from, tt.to); got != tt.want {
				t.Errorf("CanTransitionOrder(%q, %q) = %v, ingin %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestOrderIsEditable(t *testing.T) {
	editable := []string{
		domain.OrderAwaitingDP, domain.OrderDPPaid, domain.OrderPaid,
	}
	for _, status := range editable {
		if !domain.OrderIsEditable(status) {
			t.Errorf("order berstatus %s seharusnya masih bisa diedit", status)
		}
	}

	// Setelah barang diserahkan ke kurir, isi order dibekukan.
	frozen := []string{domain.OrderShipped, domain.OrderCompleted, domain.OrderCancelled}
	for _, status := range frozen {
		if domain.OrderIsEditable(status) {
			t.Errorf("order berstatus %s seharusnya tidak bisa diedit lagi", status)
		}
	}
}

func TestOrderCountsAsRevenue(t *testing.T) {
	// Order batal tidak jadi uang, jadi tidak boleh mengembungkan angka omzet
	// pada laporan. Sisanya sudah punya kesepakatan harga dengan customer.
	if domain.OrderCountsAsRevenue(domain.OrderCancelled) {
		t.Error("order batal tidak boleh dihitung sebagai omzet")
	}
	for _, status := range []string{
		domain.OrderAwaitingDP, domain.OrderDPPaid, domain.OrderPaid, domain.OrderCompleted,
	} {
		if !domain.OrderCountsAsRevenue(status) {
			t.Errorf("order berstatus %s seharusnya dihitung sebagai omzet", status)
		}
	}
}

func TestCanTransitionTrip(t *testing.T) {
	tests := []struct {
		from, to string
		want     bool
	}{
		{domain.TripOpen, domain.TripClosed, true},
		{domain.TripClosed, domain.TripOpen, true}, // order dibuka kembali

		// Status yang sudah tidak ada tidak boleh membuka jalur apa pun.
		{"shopping", domain.TripOpen, false},
		{domain.TripOpen, "settled", false},
		{"draft", domain.TripOpen, false},
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			if got := domain.CanTransitionTrip(tt.from, tt.to); got != tt.want {
				t.Errorf("CanTransitionTrip(%q, %q) = %v, ingin %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestTripAcceptsOrder(t *testing.T) {
	if !domain.TripAcceptsOrder(domain.TripOpen) {
		t.Error("trip open seharusnya menerima order")
	}
	// Menutup trip adalah cara admin menghentikan order yang masuk, jadi
	// harus benar-benar menutupnya.
	if domain.TripAcceptsOrder(domain.TripClosed) {
		t.Error("trip closed seharusnya tidak menerima order baru")
	}
}

func TestNextOrderStatusesTidakBisaDimutasi(t *testing.T) {
	// Pemanggil menerima salinan, sehingga peta transisi internal tidak bisa
	// dirusak dari luar package.
	first := domain.NextOrderStatuses(domain.OrderAwaitingDP)
	if len(first) == 0 {
		t.Fatal("order yang menunggu DP seharusnya punya status lanjutan")
	}
	first[0] = "dirusak"

	if domain.NextOrderStatuses(domain.OrderAwaitingDP)[0] == "dirusak" {
		t.Error("peta transisi internal ikut berubah saat hasilnya dimodifikasi")
	}
}
