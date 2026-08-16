package domain

import (
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/pkg/money"
)

// Jenis markup yang didukung saat menyusun harga jual.
const (
	MarkupPercent = "percent" // harga modal ditambah sekian persen
	MarkupNominal = "nominal" // harga modal ditambah sekian rupiah
)

func IsValidMarkupType(t string) bool {
	return t == MarkupPercent || t == MarkupNominal
}

// CalculateSellPrice menghitung harga jual satu produk pada sebuah trip.
//
// Urutannya: harga modal dalam mata uang asing dikonversi ke rupiah memakai
// kurs trip, lalu markup diterapkan, lalu hasilnya dibulatkan ke atas ke
// kelipatan seratus rupiah agar enak dipublikasikan ke customer.
//
// Mengembalikan harga modal dalam rupiah dan harga jualnya.
func CalculateSellPrice(
	costForeign decimal.Decimal,
	exchangeRate decimal.Decimal,
	markupType string,
	markupValue decimal.Decimal,
) (costIDR, sellPrice decimal.Decimal) {
	costIDR = money.Convert(costForeign, exchangeRate)

	switch markupType {
	case MarkupPercent:
		sellPrice = costIDR.Add(costIDR.Mul(markupValue).Div(decimal.NewFromInt(100)))
	case MarkupNominal:
		sellPrice = costIDR.Add(markupValue)
	default:
		// Tipe markup tak dikenal diperlakukan sebagai tanpa markup; validasi
		// input yang menolaknya ada di layer service.
		sellPrice = costIDR
	}

	return costIDR, money.RoundUpToHundred(sellPrice)
}
