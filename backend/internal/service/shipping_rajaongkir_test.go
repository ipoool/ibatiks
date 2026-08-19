package service

import (
	"testing"

	"github.com/ipoool/jastipin/backend/internal/pkg/rajaongkir"
)

func TestPilihOngkir(t *testing.T) {
	costs := []rajaongkir.Cost{
		{Code: "jne", Service: "REG", Cost: 25000},
		{Code: "jne", Service: "YES", Cost: 40000},
		{Code: "sicepat", Service: "BEST", Cost: 18000},
		{Code: "jnt", Service: "EZ", Cost: 0}, // tanpa harga, harus diabaikan
	}

	t.Run("kurir dan layanan cocok persis", func(t *testing.T) {
		got := pilihOngkir(costs, "JNE", "YES")
		if got == nil || got.Cost != 40000 {
			t.Fatalf("mau JNE YES 40000, dapat %+v", got)
		}
	})

	t.Run("kurir cocok tapi layanannya tidak, ambil termurah kurir itu", func(t *testing.T) {
		got := pilihOngkir(costs, "JNE", "OKE")
		if got == nil || got.Cost != 25000 {
			t.Fatalf("mau JNE REG 25000, dapat %+v", got)
		}
	})

	t.Run("kurirnya tidak ada, ambil termurah keseluruhan", func(t *testing.T) {
		got := pilihOngkir(costs, "wahana", "REG")
		if got == nil || got.Cost != 18000 {
			t.Fatalf("mau sicepat 18000, dapat %+v", got)
		}
	})

	t.Run("harga nol tidak pernah dipilih", func(t *testing.T) {
		got := pilihOngkir([]rajaongkir.Cost{{Code: "jnt", Service: "EZ", Cost: 0}}, "jnt", "EZ")
		if got != nil {
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
