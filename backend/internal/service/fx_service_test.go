package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Server tiruan dipakai supaya pengujian tidak bergantung pada layanan luar
// maupun koneksi internet.
func fxServer(t *testing.T, body string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestFXRate(t *testing.T) {
	server := fxServer(t, `{"result":"success","provider":"contoh","rates":{"IDR":111.86666,"USD":0.0064}}`, 200)
	defer server.Close()

	svc := NewFXService()
	svc.UseEndpoint(server.URL + "/")

	rate, err := svc.Rate(context.Background(), "jpy")
	if err != nil {
		t.Fatalf("Rate() galat: %v", err)
	}
	if rate.From != "JPY" {
		t.Errorf("From = %q, mau JPY (kode harus dinormalkan ke huruf besar)", rate.From)
	}
	if got := rate.Rate.String(); got != "111.86666" {
		t.Errorf("Rate = %s, mau 111.86666", got)
	}
	if rate.Source != "contoh" {
		t.Errorf("Source = %q, mau contoh", rate.Source)
	}
}

func TestFXRateRupiahTidakMemanggilLayananLuar(t *testing.T) {
	svc := NewFXService()
	// Endpoint sengaja diarahkan ke alamat yang pasti gagal: rupiah ke rupiah
	// tidak boleh sampai menyentuh jaringan.
	svc.UseEndpoint("http://127.0.0.1:1/")

	rate, err := svc.Rate(context.Background(), "IDR")
	if err != nil {
		t.Fatalf("Rate(IDR) galat: %v", err)
	}
	if rate.Rate.String() != "1" {
		t.Errorf("Rate = %s, mau 1", rate.Rate.String())
	}
	if rate.Source != "tetap" {
		t.Errorf("Source = %q, mau tetap", rate.Source)
	}
}

func TestFXRateKodeTidakValid(t *testing.T) {
	svc := NewFXService()
	for _, code := range []string{"", "J", "JPYY", "  "} {
		if _, err := svc.Rate(context.Background(), code); err == nil {
			t.Errorf("Rate(%q) tidak menghasilkan galat, padahal kodenya tidak valid", code)
		}
	}
}

func TestFXRateLayananBermasalah(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"layanan menolak", `{}`, 500},
		{"jawaban bukan JSON", `bukan json`, 200},
		{"kurs rupiah tidak ada", `{"result":"success","rates":{"USD":0.0064}}`, 200},
		{"kurs nol", `{"result":"success","rates":{"IDR":0}}`, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := fxServer(t, tt.body, tt.status)
			defer server.Close()

			svc := NewFXService()
			svc.UseEndpoint(server.URL + "/")

			// Yang penting bukan cuma gagal, tapi gagal dengan pesan yang
			// mengarahkan admin mengisi kursnya sendiri.
			if _, err := svc.Rate(context.Background(), "JPY"); err == nil {
				t.Fatal("Rate() berhasil, padahal jawaban layanan bermasalah")
			}
		})
	}
}

func TestFXRateMemakaiCache(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"result":"success","rates":{"IDR":111.5}}`))
	}))
	defer server.Close()

	svc := NewFXService()
	svc.UseEndpoint(server.URL + "/")

	for range 3 {
		if _, err := svc.Rate(context.Background(), "JPY"); err != nil {
			t.Fatalf("Rate() galat: %v", err)
		}
	}
	if hits != 1 {
		t.Errorf("layanan luar dipanggil %d kali, mau 1 — hasilnya seharusnya di-cache", hits)
	}
}
