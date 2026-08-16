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
		{domain.OrderDPPaid, domain.OrderPurchasing, true},
		{domain.OrderPurchasing, domain.OrderArrived, true},
		{domain.OrderArrived, domain.OrderPacked, true},
		{domain.OrderPacked, domain.OrderInvoiced, true},
		{domain.OrderInvoiced, domain.OrderPaid, true},
		{domain.OrderPaid, domain.OrderShipped, true},
		{domain.OrderShipped, domain.OrderCompleted, true},

		// Lompatan yang melewati tahap penting harus ditolak.
		{domain.OrderDraft, domain.OrderPaid, false},
		{domain.OrderDraft, domain.OrderShipped, false},
		{domain.OrderAwaitingDP, domain.OrderShipped, false},
		{domain.OrderDPPaid, domain.OrderPaid, false},
		{domain.OrderArrived, domain.OrderShipped, false},

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
		domain.OrderPurchasing, domain.OrderArrived, domain.OrderPacked,
		domain.OrderInvoiced, domain.OrderPaid,
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
		{domain.TripDraft, domain.TripOpen, true},
		{domain.TripOpen, domain.TripShopping, true},
		{domain.TripOpen, domain.TripClosed, true},
		{domain.TripClosed, domain.TripOpen, true}, // order dibuka kembali
		{domain.TripShopping, domain.TripInTransit, true},
		{domain.TripInTransit, domain.TripArrived, true},
		{domain.TripArrived, domain.TripSettled, true},

		{domain.TripDraft, domain.TripArrived, false},
		{domain.TripSettled, domain.TripOpen, false},
		{domain.TripCancelled, domain.TripOpen, false},
		{domain.TripInTransit, domain.TripCancelled, false},
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
	// Order masih boleh masuk saat tripper sudah di lokasi, karena di lapangan
	// sering ada customer yang menyusul titip.
	if !domain.TripAcceptsOrder(domain.TripOpen) {
		t.Error("trip open seharusnya menerima order")
	}
	if !domain.TripAcceptsOrder(domain.TripShopping) {
		t.Error("trip shopping seharusnya masih menerima order")
	}

	for _, status := range []string{
		domain.TripDraft, domain.TripClosed, domain.TripInTransit,
		domain.TripArrived, domain.TripSettled, domain.TripCancelled,
	} {
		if domain.TripAcceptsOrder(status) {
			t.Errorf("trip berstatus %s seharusnya tidak menerima order baru", status)
		}
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
