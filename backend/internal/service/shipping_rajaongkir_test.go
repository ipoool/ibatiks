package service

import (
	"testing"

	"github.com/shopspring/decimal"
)

func rp(n int64) decimal.Decimal { return decimal.NewFromInt(n) }

func TestPilihOngkir(t *testing.T) {
	// Layanan tanpa harga sudah disingkirkan QuoteOptions sebelum sampai sini.
	costs := []CostQuote{
		{Courier: "JNE", Service: "REG", Cost: rp(25000)},
		{Courier: "JNE", Service: "YES", Cost: rp(40000)},
		{Courier: "SICEPAT", Service: "BEST", Cost: rp(18000)},
	}

	t.Run("kurir dan layanan cocok persis", func(t *testing.T) {
		got := pilihOngkir(costs, "JNE", "YES")
		if got == nil || !got.Cost.Equal(rp(40000)) {
			t.Fatalf("mau JNE YES 40000, dapat %+v", got)
		}
	})

	t.Run("kurir cocok tapi layanannya tidak, ambil termurah kurir itu", func(t *testing.T) {
		got := pilihOngkir(costs, "JNE", "OKE")
		if got == nil || !got.Cost.Equal(rp(25000)) {
			t.Fatalf("mau JNE REG 25000, dapat %+v", got)
		}
	})

	t.Run("kurirnya tidak ada, ambil termurah keseluruhan", func(t *testing.T) {
		got := pilihOngkir(costs, "wahana", "REG")
		if got == nil || !got.Cost.Equal(rp(18000)) {
			t.Fatalf("mau sicepat 18000, dapat %+v", got)
		}
	})

	t.Run("daftar kosong tidak memilih apa pun", func(t *testing.T) {
		if got := pilihOngkir(nil, "jnt", "EZ"); got != nil {
			t.Fatalf("mau nil, dapat %+v", got)
		}
	})
}

func TestDaftarKurir(t *testing.T) {
	cases := []struct {
		nilai string
		mau   []string
	}{
		{"jne:jnt:sicepat", []string{"jne", "jnt", "sicepat"}},
		{" JNE : Sicepat ", []string{"jne", "sicepat"}},
		{"jne::", []string{"jne"}},
		{"", []string{}},
	}
	for _, c := range cases {
		got := daftarKurir(c.nilai)
		if len(got) != len(c.mau) {
			t.Fatalf("daftarKurir(%q) = %v, mau %v", c.nilai, got, c.mau)
		}
		for i := range got {
			if got[i] != c.mau[i] {
				t.Fatalf("daftarKurir(%q) = %v, mau %v", c.nilai, got, c.mau)
			}
		}
	}
}
