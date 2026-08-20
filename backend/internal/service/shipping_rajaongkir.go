package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/rajaongkir"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

// CostProvider adalah sumber tarif yang mengembalikan ongkos utuh untuk sebuah
// berat, bukan tarif per kilogram.
//
// Dipisah dari RateProvider karena bentuk jawabannya memang berbeda. Kurir
// tidak menjual "harga per kg" — mereka menjual ongkos untuk satu paket dengan
// berat tertentu, dan tarifnya berjenjang. Memaksanya menjadi angka per kilo
// akan membuat hasilnya meleset pada berat yang bukan kelipatan bulat.
type CostProvider interface {
	Name() string
	// QuoteCost mengembalikan pilihan termurah untuk berat tersebut, atau nil
	// tanpa error kalau tujuannya tidak dikenali — pemanggil lalu jatuh ke
	// tabel tarif.
	QuoteCost(ctx context.Context, in CostQuoteInput) (*CostQuote, error)
	// QuoteOptions mengembalikan seluruh layanan yang tersedia, bukan satu.
	// Admin yang sedang mengemas perlu melihat pilihannya sendiri: yang
	// termurah tidak selalu yang dipilih, karena customer kadang minta layanan
	// yang lebih cepat.
	QuoteOptions(ctx context.Context, in CostQuoteInput) ([]CostQuote, error)
	// TrackWaybill melacak resi yang sudah ada. Kurir yang menerbitkan resi,
	// bukan API ini — yang bisa ditanyakan hanya posisi resi yang sudah dipegang.
	TrackWaybill(ctx context.Context, awb, courier string) (*domain.TrackingInfo, error)
	// SearchDestination mencari daftar tujuan yang cocok dengan kata kunci.
	// Dipakai menu Pengaturan untuk memilih kota asal pengiriman — asal yang
	// keliru membuat seluruh perhitungan ongkir meleset, jadi tim toko harus
	// memilihnya dari daftar resmi kurir, bukan mengetik sendiri.
	SearchDestination(ctx context.Context, q string, limit int) ([]domain.ShippingDestination, error)
}

type CostQuoteInput struct {
	Courier     string
	Service     string
	City        string
	District    string
	Subdistrict string
	PostalCode  string
	WeightGram  int
}

// CostQuote adalah satu penawaran ongkir dari kurir.
type CostQuote struct {
	Courier     string
	Service     string
	Cost        decimal.Decimal
	ETD         string
	Destination string // label tujuan seperti dikenali RajaOngkir
}

// rajaOngkirProvider mengambil ongkir langsung dari RajaOngkir.
type rajaOngkirProvider struct {
	pool     *pgxpool.Pool
	client   *rajaongkir.Client
	rates    *repository.ShippingRepo
	settings *repository.SettingsRepo
}

func NewRajaOngkirProvider(
	pool *pgxpool.Pool,
	client *rajaongkir.Client,
	rates *repository.ShippingRepo,
	settings *repository.SettingsRepo,
) CostProvider {
	return &rajaOngkirProvider{pool: pool, client: client, rates: rates, settings: settings}
}

func (p *rajaOngkirProvider) Name() string { return "RajaOngkir" }

// SearchDestination meneruskan pencarian ke RajaOngkir dan sekaligus menyimpan
// hasilnya, sehingga kota yang sudah pernah dicari tidak memakan kuota lagi
// saat dipakai menghitung ongkir.
func (p *rajaOngkirProvider) SearchDestination(ctx context.Context, q string, limit int) ([]domain.ShippingDestination, error) {
	q = strings.TrimSpace(q)
	if q == "" || !p.client.Enabled() {
		return nil, nil
	}
	if limit <= 0 || limit > 25 {
		limit = 10
	}

	hasil, err := p.client.SearchDestination(ctx, q, limit)
	if err != nil {
		return nil, err
	}

	tujuan := make([]domain.ShippingDestination, 0, len(hasil))
	for _, d := range hasil {
		tujuan = append(tujuan, domain.ShippingDestination{
			DestinationID: d.ID,
			Label:         d.Label,
			CityName:      teksOpsional(d.CityName),
			ProvinceName:  teksOpsional(d.ProvinceName),
			ZipCode:       teksOpsional(d.ZipCode),
		})
	}

	// Hasil teratas disimpan sebagai pemetaan kata kunci ini. Cukup yang
	// teratas: itu yang akan dipakai resolveDestination kalau alamat order
	// kebetulan sama persis dengan yang dicari admin.
	if len(tujuan) > 0 {
		if _, err := p.rates.SaveDestination(ctx, p.pool, domain.ShippingDestination{
			Query:         q,
			DestinationID: tujuan[0].DestinationID,
			Label:         tujuan[0].Label,
			CityName:      tujuan[0].CityName,
			ProvinceName:  tujuan[0].ProvinceName,
			ZipCode:       tujuan[0].ZipCode,
		}); err != nil {
			return nil, err
		}
	}
	return tujuan, nil
}

func (p *rajaOngkirProvider) QuoteOptions(ctx context.Context, in CostQuoteInput) ([]CostQuote, error) {
	if !p.client.Enabled() {
		return nil, nil
	}

	settings, err := p.settings.All(ctx, p.pool)
	if err != nil {
		return nil, err
	}
	originID := settingInt(settings, "shipping_origin_id", 0)
	if originID <= 0 {
		// Kota asal belum disetel. Bukan galat: admin memang belum sampai ke
		// menu Pengaturan, dan ongkir tetap bisa dihitung dari tabel tarif.
		return nil, nil
	}
	couriers := strings.TrimSpace(settingString(settings, "shipping_couriers", "jne"))
	if couriers == "" {
		couriers = "jne"
	}

	dest, err := p.resolveDestination(ctx, in)
	if err != nil || dest == nil {
		return nil, err
	}

	costs, err := p.client.DomesticCost(ctx, originID, dest.DestinationID, in.WeightGram, couriers)
	if err != nil {
		return nil, err
	}

	pilihan := make([]CostQuote, 0, len(costs))
	for _, c := range costs {
		// Layanan tanpa harga tidak berguna bagi admin yang sedang memilih.
		if c.Cost <= 0 {
			continue
		}
		pilihan = append(pilihan, CostQuote{
			Courier:     strings.ToUpper(c.Code),
			Service:     strings.ToUpper(c.Service),
			Cost:        decimal.NewFromInt(int64(c.Cost)),
			ETD:         strings.TrimSpace(c.ETD),
			Destination: dest.Label,
		})
	}
	sort.SliceStable(pilihan, func(i, j int) bool {
		return pilihan[i].Cost.LessThan(pilihan[j].Cost)
	})
	return pilihan, nil
}

func (p *rajaOngkirProvider) QuoteCost(ctx context.Context, in CostQuoteInput) (*CostQuote, error) {
	pilihan, err := p.QuoteOptions(ctx, in)
	if err != nil || len(pilihan) == 0 {
		return nil, err
	}
	return pilihOngkir(pilihan, in.Courier, in.Service), nil
}

// pilihOngkir memilih layanan yang paling mendekati permintaan admin.
//
// Kalau kurir dan layanannya cocok persis, itu yang dipakai. Kalau tidak, yang
// termurah — admin sedang memperkirakan biaya, dan angka termurah adalah dasar
// tawar-menawar yang paling masuk akal ketimbang layanan pertama yang kebetulan
// dikembalikan API.
func pilihOngkir(pilihan []CostQuote, courier, service string) *CostQuote {
	courier = strings.TrimSpace(courier)
	service = strings.TrimSpace(service)

	var cocokKurir, termurah *CostQuote
	for i := range pilihan {
		c := &pilihan[i]
		if termurah == nil || c.Cost.LessThan(termurah.Cost) {
			termurah = c
		}
		if courier != "" && strings.EqualFold(c.Courier, courier) {
			if service != "" && strings.EqualFold(c.Service, service) {
				return c
			}
			if cocokKurir == nil || c.Cost.LessThan(cocokKurir.Cost) {
				cocokKurir = c
			}
		}
	}
	if cocokKurir != nil {
		return cocokKurir
	}
	return termurah
}

// resolveDestination mencari ID tujuan RajaOngkir untuk sebuah alamat.
//
// Dicoba dari yang paling spesifik ke yang paling umum: kode pos menunjuk satu
// kelurahan, sementara nama kota saja bisa menunjuk puluhan kecamatan. Hasil
// yang ketemu disimpan supaya alamat yang sama tidak dicari dua kali.
/*
 * TrackWaybill menanyakan posisi paket ke kurir lewat RajaOngkir.
 *
 * Bentuk balasan pelacakan berbeda-beda antar kurir, jadi yang dibaca hanya
 * bagian yang benar-benar dipakai dan sisanya dibiarkan. Status "terkirim"
 * hanya diakui kalau kurir menyatakannya sendiri — lewat penanda delivered atau
 * kata "delivered" pada statusnya. Menebaknya dari riwayat perjalanan berarti
 * order bisa ditandai Selesai padahal paketnya masih di jalan.
 */
func (p *rajaOngkirProvider) TrackWaybill(ctx context.Context, awb, courier string) (*domain.TrackingInfo, error) {
	if !p.client.Enabled() {
		return nil, domain.InvalidState(
			"RajaOngkir belum terhubung — isi RAJAONGKIR_API_KEY di server")
	}

	hasil, err := p.client.TrackWaybill(ctx, awb, courier)
	if err != nil {
		// Penolakan dari kurir diteruskan apa adanya. "Invalid Awb" memberi tahu
		// tim toko bahwa resinya belum dikenali — entah salah ketik, entah belum
		// masuk sistem kurir — sementara "terjadi kesalahan pada server" tidak
		// memberi tahu apa pun.
		var ro *rajaongkir.Error
		if errors.As(err, &ro) {
			return nil, domain.InvalidState("%s menolak permintaan: %s", p.Name(), ro.Message)
		}
		return nil, err
	}

	status := strings.TrimSpace(hasil.DeliveryStatus.Status)
	if status == "" {
		status = strings.TrimSpace(hasil.Summary.Status)
	}

	info := &domain.TrackingInfo{
		WaybillNumber: awb,
		Courier:       courier,
		Status:        status,
		Delivered:     hasil.Delivered || strings.EqualFold(status, "delivered"),
		ReceivedBy:    strings.TrimSpace(hasil.DeliveryStatus.PODReceiver),
		History:       make([]domain.TrackingStep, 0, len(hasil.Manifest)),
	}
	for _, m := range hasil.Manifest {
		info.History = append(info.History, domain.TrackingStep{
			Description: strings.TrimSpace(m.Description),
			City:        strings.TrimSpace(m.City),
			At:          strings.TrimSpace(m.Date + " " + m.Time),
		})
	}
	return info, nil
}

func (p *rajaOngkirProvider) resolveDestination(ctx context.Context, in CostQuoteInput) (*domain.ShippingDestination, error) {
	kandidat := make([]string, 0, 4)
	tambah := func(v string) {
		v = strings.TrimSpace(v)
		if v != "" {
			kandidat = append(kandidat, v)
		}
	}
	tambah(in.PostalCode)
	if in.Subdistrict != "" && in.City != "" {
		tambah(in.Subdistrict + " " + in.City)
	}
	if in.District != "" && in.City != "" {
		tambah(in.District + " " + in.City)
	}
	tambah(in.City)

	for _, q := range kandidat {
		cached, err := p.rates.FindDestination(ctx, p.pool, q)
		if err != nil {
			return nil, err
		}
		if cached != nil {
			return cached, nil
		}
	}

	for _, q := range kandidat {
		hasil, err := p.client.SearchDestination(ctx, q, 5)
		if err != nil {
			return nil, err
		}
		if len(hasil) == 0 {
			continue
		}
		d := hasil[0]
		return p.rates.SaveDestination(ctx, p.pool, domain.ShippingDestination{
			Query:         q,
			DestinationID: d.ID,
			Label:         d.Label,
			CityName:      teksOpsional(d.CityName),
			ProvinceName:  teksOpsional(d.ProvinceName),
			ZipCode:       teksOpsional(d.ZipCode),
		})
	}
	return nil, nil
}

func teksOpsional(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func settingString(settings domain.Settings, key, fallback string) string {
	if v, ok := settings[key]; ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}
