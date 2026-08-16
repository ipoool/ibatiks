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
		// Alur normal satu pesanan dari awal sampai selesai.
		{domain.OrderDraft, domain.OrderAwaitingDP, true},
		{domain.OrderAwaitingDP, domain.OrderDPPaid, true},
		// Belanja dan penerimaan barang terjadi selagi order Diproses, jadi
		// tahap berikutnya langsung Sedang Dikemas.
		{domain.OrderDPPaid, domain.OrderPacked, true},
		{domain.OrderPacked, domain.OrderInvoiced, true},
		{domain.OrderInvoiced, domain.OrderPaid, true},

		// Pelanggan lama sering melunasi begitu diberi tahu barangnya sudah
		// sampai, sebelum invoice resmi diterbitkan.
		{domain.OrderPacked, domain.OrderPaid, true},
		{domain.OrderPaid, domain.OrderShipped, true},
		{domain.OrderShipped, domain.OrderCompleted, true},

		// Lompatan yang melewati tahap penting harus ditolak.
		{domain.OrderDraft, domain.OrderPaid, false},
		{domain.OrderDraft, domain.OrderShipped, false},
		{domain.OrderAwaitingDP, domain.OrderShipped, false},
		{domain.OrderDPPaid, domain.OrderPaid, false},
		{domain.OrderDPPaid, domain.OrderShipped, false},

		// Status akhir tidak bisa berpindah ke mana pun.
		{domain.OrderCompleted, domain.OrderShipped, false},
		{domain.OrderCompleted, domain.OrderCancelled, false},
		{domain.OrderCancelled, domain.OrderDraft, false},

		// Order yang sudah dikirim tidak bisa dibatalkan.
		{domain.OrderShipped, domain.OrderCancelled, false},

		// Pembatalan diizinkan selama barang belum diserahkan ke kurir.
		{domain.OrderDraft, domain.OrderCancelled, true},
		{domain.OrderDPPaid, domain.OrderCancelled, true},
		{domain.OrderPacked, domain.OrderCancelled, true},

		// Status tak dikenal tidak boleh membuka jalur apa pun.
		{"entah", domain.OrderPaid, false},
		{domain.OrderDraft, "entah", false},
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
		domain.OrderDraft, domain.OrderAwaitingDP, domain.OrderDPPaid,
		domain.OrderPacked, domain.OrderInvoiced, domain.OrderPaid,
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
	// Draft belum disepakati customer dan batal tidak jadi uang; keduanya
	// tidak boleh mengembungkan angka omzet pada laporan.
	if domain.OrderCountsAsRevenue(domain.OrderDraft) {
		t.Error("order draft tidak boleh dihitung sebagai omzet")
	}
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
	first := domain.NextOrderStatuses(domain.OrderDraft)
	if len(first) == 0 {
		t.Fatal("order draft seharusnya punya status lanjutan")
	}
	first[0] = "dirusak"

	if domain.NextOrderStatuses(domain.OrderDraft)[0] == "dirusak" {
		t.Error("peta transisi internal ikut berubah saat hasilnya dimodifikasi")
	}
}
