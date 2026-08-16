package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

func dec(v string) decimal.Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		panic(err)
	}
	return d
}

func TestCalculateSellPrice(t *testing.T) {
	tests := []struct {
		name          string
		costForeign   string
		rate          string
		markupType    string
		markupValue   string
		wantCostIDR   string
		wantSellPrice string
	}{
		{
			name:        "markup persen dibulatkan ke atas ke ratusan",
			costForeign: "880", rate: "108.5",
			markupType: domain.MarkupPercent, markupValue: "35",
			// 880 x 108,5 = 95.480 ; +35% = 128.898 -> dibulatkan 128.900
			wantCostIDR: "95480", wantSellPrice: "128900",
		},
		{
			name:        "markup nominal ditambahkan setelah konversi",
			costForeign: "780", rate: "108.5",
			markupType: domain.MarkupNominal, markupValue: "40000",
			// 780 x 108,5 = 84.630 ; +40.000 = 124.630 -> dibulatkan 124.700
			wantCostIDR: "84630", wantSellPrice: "124700",
		},
		{
			name:        "hasil yang sudah kelipatan seratus tidak ikut naik",
			costForeign: "1000", rate: "100",
			markupType: domain.MarkupPercent, markupValue: "30",
			wantCostIDR: "100000", wantSellPrice: "130000",
		},
		{
			name:        "markup nol menghasilkan harga jual sama dengan modal",
			costForeign: "500", rate: "100",
			markupType: domain.MarkupPercent, markupValue: "0",
			wantCostIDR: "50000", wantSellPrice: "50000",
		},
		{
			name:        "harga modal nol tidak menghasilkan harga negatif",
			costForeign: "0", rate: "108.5",
			markupType: domain.MarkupPercent, markupValue: "35",
			wantCostIDR: "0", wantSellPrice: "0",
		},
		{
			name:        "kurs rupiah 1:1 untuk belanja domestik",
			costForeign: "125000", rate: "1",
			markupType: domain.MarkupNominal, markupValue: "15000",
			wantCostIDR: "125000", wantSellPrice: "140000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			costIDR, sellPrice := domain.CalculateSellPrice(
				dec(tt.costForeign), dec(tt.rate), tt.markupType, dec(tt.markupValue))

			if !costIDR.Equal(dec(tt.wantCostIDR)) {
				t.Errorf("harga modal IDR = %s, ingin %s", costIDR, tt.wantCostIDR)
			}
			if !sellPrice.Equal(dec(tt.wantSellPrice)) {
				t.Errorf("harga jual = %s, ingin %s", sellPrice, tt.wantSellPrice)
			}
			// Harga jual tidak boleh lebih rendah dari modal: itu berarti rugi
			// sejak sebelum barang dibeli.
			if sellPrice.LessThan(costIDR) {
				t.Errorf("harga jual %s lebih rendah dari modal %s", sellPrice, costIDR)
			}
		})
	}
}

func TestCalculateSellPriceTipeMarkupTidakDikenal(t *testing.T) {
	// Tipe markup asing diperlakukan sebagai tanpa markup, bukan panik, karena
	// validasi input sudah menolaknya lebih dulu di layer service.
	costIDR, sellPrice := domain.CalculateSellPrice(dec("1000"), dec("100"), "diskon", dec("50"))

	if !costIDR.Equal(dec("100000")) {
		t.Errorf("harga modal IDR = %s, ingin 100000", costIDR)
	}
	if !sellPrice.Equal(dec("100000")) {
		t.Errorf("harga jual = %s, ingin 100000 (tanpa markup)", sellPrice)
	}
}

func TestNormalizePhoneWA(t *testing.T) {
	tests := []struct{ input, want string }{
		{"081234567890", "6281234567890"},
		{"0812-3456-7890", "6281234567890"},
		{"+62 812 3456 7890", "6281234567890"},
		{"6281234567890", "6281234567890"},
		{"81234567890", "6281234567890"},
		{" 0812 3456 7890 ", "6281234567890"},
		{"(0812) 3456-7890", "6281234567890"},
		{"", ""},
		{"---", ""},
		// Nomor berkode negara eksplisit tidak boleh diberi awalan 62 lagi,
		// termasuk nomor Jepang yang kebetulan diawali angka 8.
		{"+81 90 1234 5678", "819012345678"},
		{"00819012345678", "819012345678"},
		{"+65 8123 4567", "6581234567"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := domain.NormalizePhoneWA(tt.input); got != tt.want {
				t.Errorf("NormalizePhoneWA(%q) = %q, ingin %q", tt.input, got, tt.want)
			}
		})
	}
}
