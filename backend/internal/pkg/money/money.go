// Package money berisi helper perhitungan dan format nominal rupiah.
//
// Seluruh nominal di aplikasi ini memakai decimal.Decimal, bukan float64.
// Alasannya sederhana: float64 tidak bisa merepresentasikan pecahan desimal
// secara persis, dan selisih satu rupiah pada laporan keuangan akan terlihat.
package money

import (
	"strings"

	"github.com/shopspring/decimal"
)

var (
	Zero    = decimal.Zero
	hundred = decimal.NewFromInt(100)
)

// FromInt membuat Decimal dari bilangan bulat rupiah.
func FromInt(v int64) decimal.Decimal { return decimal.NewFromInt(v) }

// Parse membaca nominal dari string. String kosong dianggap nol.
func Parse(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Zero, nil
	}
	return decimal.NewFromString(s)
}

// RoundRupiah membulatkan ke rupiah penuh (2 desimal dibuang) dan dipakai untuk
// nominal yang benar-benar ditagihkan ke customer.
func RoundRupiah(d decimal.Decimal) decimal.Decimal {
	return d.Round(0)
}

// RoundUpToHundred membulatkan ke atas ke kelipatan seratus rupiah.
// Harga jual jastip lazimnya tidak berakhiran angka ganjil seperti 137.412,
// jadi hasil konversi kurs selalu dirapikan dulu sebelum dipublikasikan.
func RoundUpToHundred(d decimal.Decimal) decimal.Decimal {
	if d.IsZero() {
		return Zero
	}
	return d.Div(hundred).Ceil().Mul(hundred)
}

// Convert mengubah nominal mata uang asing ke rupiah memakai kurs trip.
func Convert(amount, rate decimal.Decimal) decimal.Decimal {
	return RoundRupiah(amount.Mul(rate))
}

// Percent menghitung persentase dari sebuah nominal, misalnya DP 50%.
func Percent(amount decimal.Decimal, percent int) decimal.Decimal {
	if percent <= 0 {
		return Zero
	}
	return RoundRupiah(amount.Mul(decimal.NewFromInt(int64(percent))).Div(hundred))
}

// Format merender nominal sebagai rupiah dengan pemisah ribuan titik,
// contoh: 1250000 -> "Rp1.250.000".
func Format(d decimal.Decimal) string {
	rounded := RoundRupiah(d)
	negative := rounded.IsNegative()
	digits := rounded.Abs().String()

	// Buang bagian desimal kalau ada; nominal sudah dibulatkan ke rupiah.
	if idx := strings.IndexByte(digits, '.'); idx >= 0 {
		digits = digits[:idx]
	}

	var sb strings.Builder
	if negative {
		sb.WriteString("-")
	}
	sb.WriteString("Rp")

	// Sisipkan titik tiap tiga digit dari kanan.
	lead := len(digits) % 3
	if lead > 0 {
		sb.WriteString(digits[:lead])
	}
	for i := lead; i < len(digits); i += 3 {
		if i > 0 {
			sb.WriteString(".")
		}
		sb.WriteString(digits[i : i+3])
	}

	return sb.String()
}

// Max mengembalikan nilai terbesar; berguna untuk menjaga sisa tagihan tidak
// pernah negatif saat customer melebihkan transfer.
func Max(a, b decimal.Decimal) decimal.Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}
