package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

// defaultFXEndpoint adalah sumber kurs bawaan: gratis, tanpa API key, dan
// memberi kurs harian untuk seluruh mata uang utama. Diletakkan di sini sebagai
// konstanta, bukan pengaturan, karena mengganti penyedia berarti mengganti
// bentuk responsnya juga — bukan sekadar menukar URL.
const defaultFXEndpoint = "https://open.er-api.com/v6/latest/"

// fxCacheTTL menahan hasil selama satu jam. Kurs harian tidak berubah lebih
// cepat dari itu, dan tanpa cache setiap admin yang membuka form trip akan
// memanggil layanan luar.
const fxCacheTTL = time.Hour

// FXRate adalah kurs satu mata uang terhadap rupiah beserta asal-usulnya.
// Waktu pengambilan ikut dikembalikan supaya admin bisa menilai sendiri apakah
// angkanya masih layak dipakai.
type FXRate struct {
	From      string          `json:"from"`
	To        string          `json:"to"`
	Rate      decimal.Decimal `json:"rate"`
	Source    string          `json:"source"`
	FetchedAt time.Time       `json:"fetched_at"`
}

type fxCacheEntry struct {
	rate     FXRate
	expireAt time.Time
}

// FXService mengambil kurs mata uang asing terhadap rupiah dari layanan luar.
//
// Kurs ini hanya dipakai untuk mengisi nilai awal saat trip dibuat. Setelah
// tersimpan, kurs trip tidak pernah ikut bergerak lagi: laporan laba sebuah
// trip yang sudah selesai tidak boleh berubah hanya karena kurs pasar hari ini
// berbeda.
type FXService struct {
	client   *http.Client
	endpoint string

	mu    sync.Mutex
	cache map[string]fxCacheEntry
}

func NewFXService() *FXService {
	return &FXService{
		// Timeout dibuat pendek: form trip tidak boleh menggantung menunggu
		// layanan luar. Kalau lambat, admin cukup mengetik kursnya sendiri.
		client:   &http.Client{Timeout: 6 * time.Second},
		endpoint: defaultFXEndpoint,
		cache:    map[string]fxCacheEntry{},
	}
}

// UseEndpoint menukar sumber kurs. Dipakai pengujian untuk mengarahkan ke
// server tiruan tanpa menyentuh jaringan.
func (s *FXService) UseEndpoint(endpoint string) { s.endpoint = endpoint }

type fxResponse struct {
	Result   string             `json:"result"`
	Provider string             `json:"provider"`
	Rates    map[string]float64 `json:"rates"`
}

// Rate mengembalikan berapa rupiah nilai satu satuan mata uang `from`.
func (s *FXService) Rate(ctx context.Context, from string) (*FXRate, error) {
	code := strings.ToUpper(strings.TrimSpace(from))
	if len(code) != 3 {
		return nil, domain.Validation("kode mata uang tidak valid", map[string]string{
			"from": "isi kode tiga huruf, contoh JPY",
		})
	}

	// Rupiah tidak perlu ditanyakan ke mana pun.
	if code == "IDR" {
		return &FXRate{
			From: code, To: "IDR", Rate: decimal.NewFromInt(1),
			Source: "tetap", FetchedAt: time.Now(),
		}, nil
	}

	if cached, ok := s.fromCache(code); ok {
		return &cached, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint+code, nil)
	if err != nil {
		return nil, domain.Internal(err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, domain.InvalidState(
			"kurs otomatis sedang tidak bisa diambil, isi kursnya manual dulu")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.InvalidState(
			"layanan kurs menolak permintaan (HTTP %d), isi kursnya manual dulu", resp.StatusCode)
	}

	var payload fxResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, domain.InvalidState("jawaban layanan kurs tidak terbaca, isi kursnya manual dulu")
	}

	value, ok := payload.Rates["IDR"]
	if !ok || value <= 0 {
		return nil, domain.InvalidState("kurs %s ke rupiah tidak tersedia di layanan kurs", code)
	}

	rate := FXRate{
		From: code,
		To:   "IDR",
		// Dibulatkan ke enam angka di belakang koma, mengikuti presisi kolom
		// exchange_rate di database.
		Rate:      decimal.NewFromFloat(value).Round(6),
		Source:    provider(payload.Provider),
		FetchedAt: time.Now(),
	}
	s.store(code, rate)
	return &rate, nil
}

func (s *FXService) fromCache(code string) (FXRate, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.cache[code]
	if !ok || time.Now().After(entry.expireAt) {
		return FXRate{}, false
	}
	return entry.rate, true
}

func (s *FXService) store(code string, rate FXRate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[code] = fxCacheEntry{rate: rate, expireAt: time.Now().Add(fxCacheTTL)}
}

func provider(name string) string {
	if strings.TrimSpace(name) == "" {
		return "layanan kurs"
	}
	return name
}
