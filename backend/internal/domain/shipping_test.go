package domain

import "testing"

func TestNormalizeCity(t *testing.T) {
	// Admin mengetik kota dengan gaya yang berbeda-beda; semuanya harus jatuh ke
	// satu kunci yang sama supaya tarifnya ketemu.
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"apa adanya", "bandung", "bandung"},
		{"huruf besar", "BANDUNG", "bandung"},
		{"campuran", "Jakarta Selatan", "jakarta selatan"},
		{"awalan kota", "Kota Bandung", "bandung"},
		{"awalan kabupaten disingkat", "Kab. Bogor", "bogor"},
		{"awalan kabupaten lengkap", "Kabupaten Bekasi", "bekasi"},
		{"awalan kotamadya", "Kotamadya Surabaya", "surabaya"},
		{"spasi berlebih", "  Jakarta   Pusat  ", "jakarta pusat"},
		{"kosong", "", ""},
		{"hanya spasi", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeCity(tt.in); got != tt.want {
				t.Errorf("NormalizeCity(%q) = %q, mau %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestVolumetricWeightGram(t *testing.T) {
	tests := []struct {
		name                  string
		length, width, height int
		divisor               int
		want                  int
	}{
		// Kardus mi instan 40x30x25: 30000/6000 = 5 kg. Ini contoh yang paling
		// sering bikin ongkir meleset kalau hanya menimbang.
		{"kardus besar isi ringan", 40, 30, 25, 6000, 5000},
		{"kardus kecil", 10, 10, 10, 6000, 167},
		{"pembagi 5000 lebih mahal", 40, 30, 25, 5000, 6000},
		{"satu dimensi nol", 40, 0, 25, 6000, 0},
		{"dimensi negatif", -40, 30, 25, 6000, 0},
		{"pembagi nol tidak bikin panik", 40, 30, 25, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VolumetricWeightGram(tt.length, tt.width, tt.height, tt.divisor)
			if got != tt.want {
				t.Errorf("VolumetricWeightGram(%d,%d,%d,%d) = %d, mau %d",
					tt.length, tt.width, tt.height, tt.divisor, got, tt.want)
			}
		})
	}
}

func TestChargeableWeightGram(t *testing.T) {
	tests := []struct {
		name                          string
		actual, volumetric, min, want int
	}{
		// Yang terbesar menang, lalu dibulatkan ke atas ke kilogram penuh.
		{"volume menang", 800, 5000, 1000, 5000},
		{"berat asli menang", 2300, 500, 1000, 3000},
		{"tepat 1 kg tidak naik", 1000, 0, 1000, 1000},
		{"1 gram di atas kg tetap naik", 1001, 0, 1000, 2000},
		{"di bawah minimum diangkat", 300, 0, 1000, 1000},
		{"minimum 0 tetap dibulatkan", 1200, 0, 0, 2000},
		{"tanpa dimensi", 4200, 0, 1000, 5000},
		{"semuanya nol", 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChargeableWeightGram(tt.actual, tt.volumetric, tt.min)
			if got != tt.want {
				t.Errorf("ChargeableWeightGram(%d,%d,%d) = %d, mau %d",
					tt.actual, tt.volumetric, tt.min, got, tt.want)
			}
		})
	}
}

// Menirukan perhitungan lengkap yang dilihat admin di dialog pengemasan, supaya
// contoh yang ditulis di manual benar-benar cocok dengan hasil kodenya.
func TestEstimasiOngkirContohManual(t *testing.T) {
	const divisor = 6000

	t.Run("kardus 40x30x25 berisi 800 gram ditagih 5 kg", func(t *testing.T) {
		volumetric := VolumetricWeightGram(40, 30, 25, divisor)
		chargeable := ChargeableWeightGram(800, volumetric, 1000)

		if volumetric != 5000 {
			t.Fatalf("berat volume = %d gram, mau 5000", volumetric)
		}
		if chargeable != 5000 {
			t.Fatalf("berat ditagih = %d gram, mau 5000", chargeable)
		}

		// 5 kg x Rp28.000 = Rp140.000 seperti contoh di manual.
		if cost := chargeable / 1000 * 28000; cost != 140000 {
			t.Fatalf("ongkir = %d, mau 140000", cost)
		}
	})

	t.Run("skincare 2,3 kg dalam kardus kecil dibulatkan ke 3 kg", func(t *testing.T) {
		volumetric := VolumetricWeightGram(10, 10, 10, divisor)
		chargeable := ChargeableWeightGram(2300, volumetric, 1000)

		if chargeable != 3000 {
			t.Fatalf("berat ditagih = %d gram, mau 3000", chargeable)
		}
	})
}

func TestIsValidOrderSource(t *testing.T) {
	for _, source := range OrderSources {
		if !IsValidOrderSource(source) {
			t.Errorf("IsValidOrderSource(%q) = false, mau true", source)
		}
	}

	// Nilai di luar daftar harus ditolak, termasuk yang hanya beda huruf besar —
	// kalau lolos, kolomnya jadi berisi varian ejaan dan rekap channel pecah.
	for _, source := range []string{"", "telepati", "WHATSAPP", "wa"} {
		if IsValidOrderSource(source) {
			t.Errorf("IsValidOrderSource(%q) = true, mau false", source)
		}
	}
}
