package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
	"github.com/ipoool/jastipin/backend/internal/pkg/rajaongkir"
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
	// costs diisi kalau ada layanan yang bisa menghitung ongkos utuh, misalnya
	// RajaOngkir. Dicoba lebih dulu; tabel tarif jadi cadangannya.
	costs CostProvider
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

// UseProvider mengganti sumber tarif per kilogram.
func (s *ShippingService) UseProvider(provider RateProvider) {
	s.provider = provider
}

// UseCostProvider memasang layanan yang menghitung ongkos utuh, seperti
// RajaOngkir. Tabel tarif tetap dipertahankan sebagai cadangan supaya toko
// tidak lumpuh saat langganan habis atau layanannya sedang tidak bisa dihubungi.
func (s *ShippingService) UseCostProvider(provider CostProvider) {
	s.costs = provider
}

type EstimateInput struct {
	Courier string
	Service string
	// City, District, Subdistrict, dan PostalCode dipakai mencari ID tujuan di
	// layanan kurir. Makin lengkap alamatnya, makin tepat tujuan yang ketemu:
	// kode pos menunjuk satu kelurahan, sementara nama kota saja bisa menunjuk
	// puluhan kecamatan dengan tarif berbeda.
	City        string
	District    string
	Subdistrict string
	PostalCode  string
	WeightGram  int
	LengthCM    int
	WidthCM     int
	HeightCM    int
}

// hargaPerKg menurunkan tarif per kilogram dari ongkos utuh, hanya untuk
// ditampilkan sebagai dasar hitungan. Kurir tidak menjual per kilogram, jadi
// angka ini perkiraan — yang ditagih tetap Cost.
func hargaPerKg(cost decimal.Decimal, chargeableGram int) decimal.Decimal {
	if chargeableGram <= 0 {
		return cost
	}
	kg := decimal.NewFromInt(int64(chargeableGram)).Div(decimal.NewFromInt(1000))
	if kg.IsZero() {
		return cost
	}
	return cost.Div(kg).Round(2)
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

	volumetricAwal := domain.VolumetricWeightGram(in.LengthCM, in.WidthCM, in.HeightCM, divisor)

	// Layanan kurir dicoba lebih dulu kalau terpasang. Kegagalannya sengaja
	// tidak menghentikan perhitungan: admin sedang menimbang paket di depan
	// customer, dan lebih baik menerima angka dari tabel tarif dengan penanda
	// asalnya daripada layar galat.
	if s.costs != nil {
		beratKirim := domain.ChargeableWeightGram(in.WeightGram, volumetricAwal, 1000)
		quote, errQuote := s.costs.QuoteCost(ctx, CostQuoteInput{
			Courier:     courier,
			Service:     service,
			City:        in.City,
			District:    in.District,
			Subdistrict: in.Subdistrict,
			PostalCode:  in.PostalCode,
			WeightGram:  beratKirim,
		})
		if errQuote == nil && quote != nil {
			return &domain.ShippingEstimate{
				Courier:              quote.Courier,
				Service:              quote.Service,
				City:                 domain.NormalizeCity(in.City),
				ActualWeightGram:     in.WeightGram,
				VolumetricWeightGram: volumetricAwal,
				ChargeableWeightGram: beratKirim,
				PricePerKg:           hargaPerKg(quote.Cost, beratKirim),
				Cost:                 money.RoundRupiah(quote.Cost),
				ETD:                  quote.ETD,
				Destination:          quote.Destination,
				Source:               s.costs.Name(),
				RateFound:            true,
			}, nil
		}
	}

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

	volumetric := volumetricAwal
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
	// Bagian alamat yang lebih rinci ikut dibawa supaya layanan kurir bisa
	// menunjuk kecamatan yang tepat, bukan sekadar kotanya.
	if strings.TrimSpace(in.District) == "" {
		in.District = derefString(order.ShippingDistrict)
	}
	if strings.TrimSpace(in.Subdistrict) == "" {
		in.Subdistrict = derefString(order.ShippingSubdistrict)
	}
	if strings.TrimSpace(in.PostalCode) == "" {
		in.PostalCode = derefString(order.ShippingPostalCode)
	}
	return s.Estimate(ctx, in)
}

// --- Layanan tarif kurir ----------------------------------------------------

// SearchDestinations mencari kota/kecamatan tujuan di layanan kurir.
//
// Mengembalikan daftar kosong kalau layanannya tidak terpasang, bukan galat:
// menu Pengaturan tetap harus bisa dibuka toko yang belum berlangganan.
func (s *ShippingService) SearchDestinations(ctx context.Context, q string) ([]domain.ShippingDestination, error) {
	if s.costs == nil {
		return []domain.ShippingDestination{}, nil
	}
	if len(strings.TrimSpace(q)) < 3 {
		return nil, domain.Validation("kata kunci terlalu pendek", map[string]string{
			"q": "ketik minimal 3 huruf nama kota atau kecamatan",
		})
	}

	tujuan, err := s.costs.SearchDestination(ctx, q, 10)
	if err != nil {
		// Pesan penolakan dari kurir diteruskan apa adanya. API key keliru atau
		// langganan habis adalah hal yang hanya bisa dibereskan tim toko, dan
		// "terjadi kesalahan pada server" tidak memberi tahu apa pun.
		var ro *rajaongkir.Error
		if errors.As(err, &ro) {
			return nil, domain.InvalidState("%s menolak permintaan: %s", s.costs.Name(), ro.Message)
		}
		return nil, err
	}
	if tujuan == nil {
		tujuan = []domain.ShippingDestination{}
	}
	return tujuan, nil
}

// ProviderInfo menjelaskan keadaan layanan tarif kepada menu Pengaturan.
func (s *ShippingService) ProviderInfo(ctx context.Context) (*domain.ShippingProviderInfo, error) {
	settings, err := s.settings.All(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	info := domain.ShippingProviderInfo{
		Name:          s.provider.Name(),
		OriginID:      settingInt(settings, "shipping_origin_id", 0),
		OriginLabel:   settingString(settings, "shipping_origin_label", ""),
		Couriers:      daftarKurir(settingString(settings, "shipping_couriers", "jne")),
		CourierOption: domain.KurirRajaOngkir,
	}
	if s.costs != nil {
		info.Name = s.costs.Name()
		info.Connected = true
		info.Ready = info.OriginID > 0 && len(info.Couriers) > 0
	}
	return &info, nil
}

// daftarKurir memecah nilai setting "jne:jnt:sicepat" menjadi daftar kode.
// Pemisahnya titik dua karena itu yang diminta RajaOngkir, jadi nilainya bisa
// dikirim apa adanya tanpa diterjemahkan lagi.
func daftarKurir(nilai string) []string {
	kurir := make([]string, 0, 4)
	for _, k := range strings.Split(nilai, ":") {
		k = strings.ToLower(strings.TrimSpace(k))
		if k != "" {
			kurir = append(kurir, k)
		}
	}
	return kurir
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
