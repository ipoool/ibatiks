package service

import (
	"context"
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

func (p *rajaOngkirProvider) QuoteCost(ctx context.Context, in CostQuoteInput) (*CostQuote, error) {
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
	if len(costs) == 0 {
		return nil, nil
	}

	pilihan := pilihOngkir(costs, in.Courier, in.Service)
	if pilihan == nil {
		return nil, nil
	}

	return &CostQuote{
		Courier:     strings.ToUpper(pilihan.Code),
		Service:     strings.ToUpper(pilihan.Service),
		Cost:        decimal.NewFromInt(int64(pilihan.Cost)),
		ETD:         strings.TrimSpace(pilihan.ETD),
		Destination: dest.Label,
	}, nil
}

// pilihOngkir memilih layanan yang paling mendekati permintaan admin.
//
// Kalau kurir dan layanannya cocok persis, itu yang dipakai. Kalau tidak, yang
// termurah — admin sedang memperkirakan biaya, dan angka termurah adalah dasar
// tawar-menawar yang paling masuk akal ketimbang layanan pertama yang kebetulan
// dikembalikan API.
func pilihOngkir(costs []rajaongkir.Cost, courier, service string) *rajaongkir.Cost {
	courier = strings.ToLower(strings.TrimSpace(courier))
	service = strings.ToLower(strings.TrimSpace(service))

	var cocokKurir, termurah *rajaongkir.Cost
	for i := range costs {
		c := &costs[i]
		if c.Cost <= 0 {
			continue
		}
		if termurah == nil || c.Cost < termurah.Cost {
			termurah = c
		}
		if courier != "" && strings.EqualFold(c.Code, courier) {
			if service != "" && strings.EqualFold(c.Service, service) {
				return c
			}
			if cocokKurir == nil || c.Cost < cocokKurir.Cost {
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
