package money_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/pkg/money"
)

func dec(v string) decimal.Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		panic(err)
	}
	return d
}

func TestFormat(t *testing.T) {
	tests := []struct{ input, want string }{
		{"0", "Rp0"},
		{"100", "Rp100"},
		{"1000", "Rp1.000"},
		{"12500", "Rp12.500"},
		{"128900", "Rp128.900"},
		{"1250000", "Rp1.250.000"},
		{"12500000", "Rp12.500.000"},
		{"1234567890", "Rp1.234.567.890"},
		// Nominal dari NUMERIC(18,2) datang dengan desimal; rupiah tidak
		// memakai sen, jadi dibulatkan lebih dulu.
		{"1250000.00", "Rp1.250.000"},
		{"1250000.49", "Rp1.250.000"},
		{"1250000.50", "Rp1.250.001"},
		// Saldo negatif muncul saat customer melebihkan transfer.
		{"-50000", "-Rp50.000"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := money.Format(dec(tt.input)); got != tt.want {
				t.Errorf("Format(%s) = %q, ingin %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRoundUpToHundred(t *testing.T) {
	tests := []struct{ input, want string }{
		{"0", "0"},
		{"1", "100"},
		{"99", "100"},
		{"100", "100"},
		{"101", "200"},
		{"128898", "128900"},
		{"128900", "128900"},
		{"128901", "129000"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := money.RoundUpToHundred(dec(tt.input))
			if !got.Equal(dec(tt.want)) {
				t.Errorf("RoundUpToHundred(%s) = %s, ingin %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	tests := []struct {
		amount, rate, want string
	}{
		{"880", "108.5", "95480"},
		{"1000", "100", "100000"},
		{"1180", "108.5", "128030"},
		{"0", "108.5", "0"},
		// Kurs 1 dipakai untuk belanja domestik.
		{"125000", "1", "125000"},
		// Hasil pecahan dibulatkan ke rupiah penuh.
		{"3", "108.53", "326"},
	}

	for _, tt := range tests {
		t.Run(tt.amount+"x"+tt.rate, func(t *testing.T) {
			got := money.Convert(dec(tt.amount), dec(tt.rate))
			if !got.Equal(dec(tt.want)) {
				t.Errorf("Convert(%s, %s) = %s, ingin %s", tt.amount, tt.rate, got, tt.want)
			}
		})
	}
}

func TestPercent(t *testing.T) {
	tests := []struct {
		amount string
		pct    int
		want   string
	}{
		{"335000", 50, "167500"},
		{"390000", 50, "195000"},
		{"100000", 30, "30000"},
		{"100000", 100, "100000"},
		{"100000", 0, "0"},
		{"100000", -10, "0"},
		// Hasil pecahan dibulatkan ke rupiah penuh.
		{"335001", 50, "167501"},
	}

	for _, tt := range tests {
		t.Run(tt.amount, func(t *testing.T) {
			got := money.Percent(dec(tt.amount), tt.pct)
			if !got.Equal(dec(tt.want)) {
				t.Errorf("Percent(%s, %d) = %s, ingin %s", tt.amount, tt.pct, got, tt.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	// Dipakai menjaga sisa tagihan tidak pernah tampil negatif.
	if got := money.Max(dec("-5000"), money.Zero); !got.Equal(money.Zero) {
		t.Errorf("Max(-5000, 0) = %s, ingin 0", got)
	}
	if got := money.Max(dec("5000"), money.Zero); !got.Equal(dec("5000")) {
		t.Errorf("Max(5000, 0) = %s, ingin 5000", got)
	}
}

func TestParse(t *testing.T) {
	// String kosong dari form yang tidak diisi dianggap nol, bukan error.
	got, err := money.Parse("")
	if err != nil {
		t.Fatalf("Parse(\"\") error: %v", err)
	}
	if !got.Equal(money.Zero) {
		t.Errorf("Parse(\"\") = %s, ingin 0", got)
	}

	if _, err := money.Parse("seratus ribu"); err == nil {
		t.Error("Parse teks non-angka seharusnya menghasilkan error")
	}
}
