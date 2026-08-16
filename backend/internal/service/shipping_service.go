package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

// RateProvider adalah sumber tarif kirim.
//
// Antarmuka ini sengaja dipisah supaya integrasi API kurir (RajaOngkir, Komerce,
// atau API JNE langsung) bisa dipasang tanpa mengubah satu pun pemanggilnya:
// cukup buat tipe baru yang memenuhi antarmuka ini lalu daftarkan di
// ShippingService. Implementasi bawaan membaca tabel shipping_rates.
type RateProvider interface {
	// Name dipakai untuk menandai asal angka pada hasil estimasi.
	Name() string
	// Quote mengembalikan tarif per kilogram untuk satu tujuan. Mengembalikan
	// nil tanpa error kalau tujuannya tidak dikenali, supaya pemanggil bisa
	// jatuh ke tarif default.
	Quote(ctx context.Context, courier, service, city string) (*domain.ShippingRate, error)
}

// tableRateProvider membaca tarif dari tabel shipping_rates yang dikelola
// sendiri oleh pemilik toko lewat menu Pengaturan.
type tableRateProvider struct {
	pool  *pgxpool.Pool
	rates *repository.ShippingRepo
}

func (p *tableRateProvider) Name() string { return "tabel tarif" }

func (p *tableRateProvider) Quote(ctx context.Context, courier, service, city string) (*domain.ShippingRate, error) {
	return p.rates.FindRate(ctx, p.pool, courier, service, city)
}

type ShippingService struct {
	pool     *pgxpool.Pool
	rates    *repository.ShippingRepo
	orders   *repository.OrderRepo
	settings *repository.SettingsRepo
	provider RateProvider
}

func NewShippingService(
	pool *pgxpool.Pool,
	rates *repository.ShippingRepo,
	orders *repository.OrderRepo,
	settings *repository.SettingsRepo,
) *ShippingService {
	return &ShippingService{
		pool:     pool,
		rates:    rates,
		orders:   orders,
		settings: settings,
		provider: &tableRateProvider{pool: pool, rates: rates},
	}
}

// UseProvider mengganti sumber tarif, misalnya dengan klien API kurir.
func (s *ShippingService) UseProvider(provider RateProvider) {
	s.provider = provider
}

type EstimateInput struct {
	Courier    string
	Service    string
	City       string
	WeightGram int
	LengthCM   int
	WidthCM    int
	HeightCM   int
}

// Estimate menghitung perkiraan ongkir.
//
// Ekspedisi menagih berdasarkan berat yang lebih besar antara berat asli dan
// berat volumetrik, dibulatkan ke atas per kilogram. Hasilnya dikembalikan
// lengkap dengan dasar hitungannya supaya admin bisa memeriksa kenapa angkanya
// sekian, bukan sekadar menerima satu angka.
func (s *ShippingService) Estimate(ctx context.Context, in EstimateInput) (*domain.ShippingEstimate, error) {
	courier := strings.ToUpper(strings.TrimSpace(in.Courier))
	if courier == "" {
		courier = "JNE"
	}
	service := strings.ToUpper(strings.TrimSpace(in.Service))
	if service == "" {
		service = "REG"
	}
	if strings.TrimSpace(in.City) == "" {
		return nil, domain.Validation("kota tujuan belum diisi", map[string]string{
			"city": "wajib diisi untuk menghitung ongkir",
		})
	}

	settings, err := s.settings.All(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	divisor := settingInt(settings, "shipping_volumetric_divisor", 6000)
	defaultPerKg := settingDecimal(settings, "shipping_default_price_per_kg", decimal.NewFromInt(25000))

	rate, err := s.provider.Quote(ctx, courier, service, in.City)
	if err != nil {
		return nil, err
	}

	pricePerKg := defaultPerKg
	minWeight := 1000
	etd := ""
	source := "tarif default"
	found := false

	if rate != nil {
		pricePerKg = rate.PricePerKg
		minWeight = rate.MinWeightGram
		if rate.ETD != nil {
			etd = *rate.ETD
		}
		source = s.provider.Name()
		found = true
	}

	volumetric := domain.VolumetricWeightGram(in.LengthCM, in.WidthCM, in.HeightCM, divisor)
	chargeable := domain.ChargeableWeightGram(in.WeightGram, volumetric, minWeight)

	kg := decimal.NewFromInt(int64(chargeable)).Div(decimal.NewFromInt(1000))
	cost := money.RoundRupiah(pricePerKg.Mul(kg))

	return &domain.ShippingEstimate{
		Courier:              courier,
		Service:              service,
		City:                 domain.NormalizeCity(in.City),
		ActualWeightGram:     in.WeightGram,
		VolumetricWeightGram: volumetric,
		ChargeableWeightGram: chargeable,
		PricePerKg:           pricePerKg,
		Cost:                 cost,
		ETD:                  etd,
		Source:               source,
		RateFound:            found,
	}, nil
}

// EstimateForOrder menghitung ongkir memakai kota tujuan dari alamat order,
// supaya admin tidak perlu mengetik ulang kotanya.
func (s *ShippingService) EstimateForOrder(ctx context.Context, orderID uuid.UUID, in EstimateInput) (*domain.ShippingEstimate, error) {
	order, err := s.orders.GetByID(ctx, s.pool, orderID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.City) == "" {
		in.City = order.ShippingCity
	}
	return s.Estimate(ctx, in)
}

// --- Pengelolaan tarif ------------------------------------------------------

func (s *ShippingService) ListRates(ctx context.Context, courier, search string) ([]domain.ShippingRate, error) {
	return s.rates.List(ctx, s.pool, courier, search)
}

type RateInput struct {
	Courier         string
	Service         string
	DestinationCity string
	Province        *string
	PricePerKg      decimal.Decimal
	MinWeightGram   int
	ETD             *string
}

func (s *ShippingService) SaveRate(ctx context.Context, in RateInput) (*domain.ShippingRate, error) {
	if strings.TrimSpace(in.DestinationCity) == "" {
		return nil, domain.Validation("kota tujuan wajib diisi", map[string]string{
			"destination_city": "wajib diisi",
		})
	}
	if in.PricePerKg.IsNegative() {
		return nil, domain.Validation("tarif tidak valid", map[string]string{
			"price_per_kg": "harus 0 atau lebih",
		})
	}

	courier := strings.ToUpper(strings.TrimSpace(in.Courier))
	if courier == "" {
		courier = "JNE"
	}
	service := strings.ToUpper(strings.TrimSpace(in.Service))
	if service == "" {
		service = "REG"
	}
	minWeight := in.MinWeightGram
	if minWeight <= 0 {
		minWeight = 1000
	}

	return s.rates.Upsert(ctx, s.pool, repository.ShippingRateParams{
		Courier:         courier,
		Service:         service,
		DestinationCity: in.DestinationCity,
		Province:        trimPtr(in.Province),
		PricePerKg:      in.PricePerKg,
		MinWeightGram:   minWeight,
		ETD:             trimPtr(in.ETD),
	})
}

func (s *ShippingService) DeleteRate(ctx context.Context, id uuid.UUID) error {
	return s.rates.Delete(ctx, s.pool, id)
}

func settingInt(settings domain.Settings, key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(settings.Get(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func settingDecimal(settings domain.Settings, key string, fallback decimal.Decimal) decimal.Decimal {
	value, err := decimal.NewFromString(strings.TrimSpace(settings.Get(key)))
	if err != nil || value.IsNegative() {
		return fallback
	}
	return value
}
