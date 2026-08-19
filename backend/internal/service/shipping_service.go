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

// ShippingService menjawab pertanyaan ongkir memakai tarif dari kurir.
//
// Hanya ada satu sumber angka: layanan kurir. Toko pernah menyimpan tabel tarif
// per kota sendiri sebagai cadangan, dan itu dilepas — tabel yang diisi tangan
// selalu tertinggal dari tarif yang sebenarnya berlaku, dan angka yang salah
// tapi terlihat resmi lebih berbahaya daripada tidak ada angka sama sekali.
//
// Kalau kurirnya tidak bisa dihubungi, jawabannya adalah galat yang menyebutkan
// sebabnya, bukan tebakan. Admin lalu mengetik ongkirnya sendiri dari struk atau
// aplikasi kurir saat mengemas.
type ShippingService struct {
	pool     *pgxpool.Pool
	rates    *repository.ShippingRepo
	orders   *repository.OrderRepo
	settings *repository.SettingsRepo
	costs    CostProvider
}

func NewShippingService(
	pool *pgxpool.Pool,
	rates *repository.ShippingRepo,
	orders *repository.OrderRepo,
	settings *repository.SettingsRepo,
) *ShippingService {
	return &ShippingService{pool: pool, rates: rates, orders: orders, settings: settings}
}

// UseCostProvider memasang layanan yang menghitung ongkos utuh, yaitu RajaOngkir.
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

// beratDitagih menghitung berat yang dipakai menagih: yang lebih besar antara
// berat timbangan dan berat volumetrik, dibulatkan ke atas per kilogram.
func beratDitagih(in EstimateInput) (volumetrik, ditagih int) {
	volumetrik = domain.VolumetricWeightGram(
		in.LengthCM, in.WidthCM, in.HeightCM, domain.VolumetricDivisor)
	return volumetrik, domain.ChargeableWeightGram(in.WeightGram, volumetrik, 1000)
}

// Estimate mengembalikan satu perkiraan ongkir, yang termurah kalau kurir dan
// layanannya tidak ditentukan.
func (s *ShippingService) Estimate(ctx context.Context, in EstimateInput) (*domain.ShippingEstimate, error) {
	pilihan, volumetrik, ditagih, err := s.hitung(ctx, in)
	if err != nil {
		return nil, err
	}

	terpilih := pilihOngkir(pilihan, in.Courier, in.Service)
	if terpilih == nil {
		return nil, domain.InvalidState(
			"kurir tidak punya layanan untuk paket seberat ini — isi ongkirnya sendiri saat mengemas")
	}

	return &domain.ShippingEstimate{
		Courier:              terpilih.Courier,
		Service:              terpilih.Service,
		City:                 domain.NormalizeCity(in.City),
		ActualWeightGram:     in.WeightGram,
		VolumetricWeightGram: volumetrik,
		ChargeableWeightGram: ditagih,
		PricePerKg:           hargaPerKg(terpilih.Cost, ditagih),
		Cost:                 money.RoundRupiah(terpilih.Cost),
		ETD:                  terpilih.ETD,
		Destination:          terpilih.Destination,
		Source:               s.costs.Name(),
	}, nil
}

// ShippingOption adalah satu layanan kurir yang bisa dipilih admin saat
// mengemas, lengkap dengan harganya.
type ShippingOption struct {
	Courier     string          `json:"courier"`
	Service     string          `json:"service"`
	Cost        decimal.Decimal `json:"cost"`
	ETD         string          `json:"etd"`
	Destination string          `json:"destination,omitempty"`
	Source      string          `json:"source"`
}

// Options mendaftar seluruh layanan kurir untuk sebuah paket.
//
// Berbeda dari Estimate yang menjawab "berapa kira-kira", ini menjawab "apa
// saja pilihannya" — yang termurah tidak selalu yang dipakai, karena customer
// kadang minta layanan yang lebih cepat.
func (s *ShippingService) Options(ctx context.Context, in EstimateInput) ([]ShippingOption, error) {
	pilihan, _, ditagih, err := s.hitung(ctx, in)
	if err != nil {
		return nil, err
	}

	out := make([]ShippingOption, 0, len(pilihan))
	for _, p := range pilihan {
		out = append(out, ShippingOption{
			Courier:     p.Courier,
			Service:     p.Service,
			Cost:        money.RoundRupiah(p.Cost),
			ETD:         p.ETD,
			Destination: p.Destination,
			Source:      s.costs.Name(),
		})
	}
	_ = ditagih
	return out, nil
}

// hitung menanyakan seluruh pilihan ongkir ke layanan kurir.
//
// Galatnya sengaja tidak ditelan seperti dulu. Waktu masih ada tabel tarif,
// menelan galat berarti tetap memberi angka; sekarang tidak ada lagi angka
// pengganti, jadi menelannya hanya akan menghasilkan daftar kosong tanpa
// penjelasan. Admin berhak tahu kenapa daftarnya tidak keluar.
func (s *ShippingService) hitung(ctx context.Context, in EstimateInput) ([]CostQuote, int, int, error) {
	if strings.TrimSpace(in.City) == "" {
		return nil, 0, 0, domain.Validation("kota tujuan belum diisi", map[string]string{
			"city": "wajib diisi untuk menghitung ongkir",
		})
	}
	if s.costs == nil {
		return nil, 0, 0, domain.InvalidState(
			"RajaOngkir belum terhubung — isi RAJAONGKIR_API_KEY di server, atau ketik ongkirnya sendiri saat mengemas")
	}

	volumetrik, ditagih := beratDitagih(in)

	pilihan, err := s.costs.QuoteOptions(ctx, CostQuoteInput{
		City:        in.City,
		District:    in.District,
		Subdistrict: in.Subdistrict,
		PostalCode:  in.PostalCode,
		WeightGram:  ditagih,
	})
	if err != nil {
		// Pesan penolakan dari kurir diteruskan apa adanya: "API key tidak
		// valid" hanya bisa dibereskan tim toko, dan "terjadi kesalahan pada
		// server" tidak memberi tahu apa pun.
		var ro *rajaongkir.Error
		if errors.As(err, &ro) {
			return nil, 0, 0, domain.InvalidState("%s menolak permintaan: %s", s.costs.Name(), ro.Message)
		}
		return nil, 0, 0, err
	}
	if len(pilihan) == 0 {
		return nil, 0, 0, domain.InvalidState(
			"kota asal atau tujuannya belum dikenali kurir — periksa kota asal di Pengaturan, atau ketik ongkirnya sendiri saat mengemas")
	}

	return pilihan, volumetrik, ditagih, nil
}

// EstimateForOrder dan OptionsForOrder memakai alamat tujuan dari order, supaya
// admin tidak perlu mengetik ulang kotanya saat mengemas.
func (s *ShippingService) EstimateForOrder(ctx context.Context, orderID uuid.UUID, in EstimateInput) (*domain.ShippingEstimate, error) {
	in, err := s.lengkapiAlamat(ctx, orderID, in)
	if err != nil {
		return nil, err
	}
	return s.Estimate(ctx, in)
}

func (s *ShippingService) OptionsForOrder(ctx context.Context, orderID uuid.UUID, in EstimateInput) ([]ShippingOption, error) {
	in, err := s.lengkapiAlamat(ctx, orderID, in)
	if err != nil {
		return nil, err
	}
	return s.Options(ctx, in)
}

func (s *ShippingService) lengkapiAlamat(ctx context.Context, orderID uuid.UUID, in EstimateInput) (EstimateInput, error) {
	order, err := s.orders.GetByID(ctx, s.pool, orderID)
	if err != nil {
		return in, err
	}
	if strings.TrimSpace(in.City) == "" {
		in.City = order.ShippingCity
	}
	if strings.TrimSpace(in.District) == "" {
		in.District = derefString(order.ShippingDistrict)
	}
	if strings.TrimSpace(in.Subdistrict) == "" {
		in.Subdistrict = derefString(order.ShippingSubdistrict)
	}
	if strings.TrimSpace(in.PostalCode) == "" {
		in.PostalCode = derefString(order.ShippingPostalCode)
	}
	return in, nil
}

// --- Layanan tarif kurir ----------------------------------------------------

// SearchDestinations mencari kota/kecamatan tujuan di layanan kurir.
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
		Name:          "RajaOngkir",
		OriginID:      settingInt(settings, "shipping_origin_id", 0),
		OriginLabel:   settingString(settings, "shipping_origin_label", ""),
		Couriers:      daftarKurir(settingString(settings, "shipping_couriers", "jne")),
		CourierOption: domain.KurirRajaOngkir,
	}
	if s.costs != nil {
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

func settingInt(settings domain.Settings, key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(settings.Get(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
