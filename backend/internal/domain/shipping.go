package domain

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// VolumetricDivisor adalah pembagi berat volumetrik: (P x L x T dalam cm)
// dibagi angka ini menghasilkan kilogram.
//
// 6000 adalah angka yang dipakai JNE dan diikuti hampir semua kurir dalam
// negeri. Dulu ini bisa disetel per toko, tapi tidak ada yang pernah
// mengubahnya — dan salah menyetelnya membuat seluruh perkiraan ongkir meleset
// tanpa gejala apa pun. Lebih baik satu angka yang benar daripada satu kolom
// yang bisa salah diisi.
const VolumetricDivisor = 6000

// ShippingEstimate adalah hasil perhitungan ongkir beserta dasar hitungannya,
// supaya admin bisa melihat kenapa angkanya sekian.
type ShippingEstimate struct {
	Courier string `json:"courier"`
	Service string `json:"service"`
	City    string `json:"city"`

	ActualWeightGram     int `json:"actual_weight_gram"`
	VolumetricWeightGram int `json:"volumetric_weight_gram"`
	// ChargeableWeightGram adalah yang dipakai menagih: mana yang lebih besar
	// antara berat asli dan berat volumetrik, dibulatkan ke atas per kilogram.
	ChargeableWeightGram int `json:"chargeable_weight_gram"`

	PricePerKg decimal.Decimal `json:"price_per_kg"`
	Cost       decimal.Decimal `json:"cost"`
	ETD        string          `json:"etd"`
	// Destination adalah tujuan seperti dikenali layanan kurir, misalnya
	// "CILANDAK BARAT, CILANDAK, JAKARTA SELATAN, DKI JAKARTA, 12430". Admin
	// perlu melihatnya untuk memastikan yang dihitung memang alamat yang benar
	// — nama kota yang sama bisa menunjuk kecamatan yang berbeda tarifnya.
	Destination string `json:"destination,omitempty"`

	// Source menyebutkan dari mana angkanya diambil, yaitu nama layanan kurir
	// yang menjawabnya.
	Source string `json:"source"`
}

// NormalizeCity merapikan nama kota agar pencocokan tarif tidak bergantung pada
// cara admin mengetik. "Jakarta Selatan", "JAKARTA SELATAN", dan
// "Kota Jakarta Selatan" sama-sama menjadi "jakarta selatan".
func NormalizeCity(city string) string {
	normalized := strings.ToLower(strings.TrimSpace(city))
	for _, prefix := range []string{"kota ", "kab. ", "kabupaten ", "kotamadya "} {
		normalized = strings.TrimPrefix(normalized, prefix)
	}
	return strings.Join(strings.Fields(normalized), " ")
}

// VolumetricWeightGram menghitung berat volumetrik dari dimensi paket.
//
// Ekspedisi menagih berdasarkan ruang yang dimakan paket di dalam truk, bukan
// hanya beratnya. Kardus besar berisi tisu tetap mahal walau ringan. Rumusnya
// (P x L x T dalam cm) dibagi sebuah pembagi, hasilnya dalam kilogram; JNE
// memakai pembagi 6000.
func VolumetricWeightGram(lengthCM, widthCM, heightCM, divisor int) int {
	if lengthCM <= 0 || widthCM <= 0 || heightCM <= 0 || divisor <= 0 {
		return 0
	}
	kg := float64(lengthCM*widthCM*heightCM) / float64(divisor)
	return int(math.Round(kg * 1000))
}

// ChargeableWeightGram mengembalikan berat yang ditagihkan: yang lebih besar
// antara berat asli dan berat volumetrik, dibulatkan ke atas ke kilogram penuh
// dan tidak pernah kurang dari berat minimum tarif.
func ChargeableWeightGram(actualGram, volumetricGram, minGram int) int {
	weight := actualGram
	if volumetricGram > weight {
		weight = volumetricGram
	}
	if minGram > 0 && weight < minGram {
		weight = minGram
	}
	if weight <= 0 {
		return 0
	}

	// Ekspedisi selalu membulatkan ke atas: 1,2 kg ditagih 2 kg.
	kg := int(math.Ceil(float64(weight) / 1000.0))
	return kg * 1000
}

// ShippingDestination adalah pemetaan alamat ke ID tujuan milik RajaOngkir.
//
// Disimpan supaya alamat yang sama tidak memicu pencarian berulang: kuota
// langganan terbatas, dan pemetaan kota ke ID hampir tidak pernah berubah.
type ShippingDestination struct {
	ID            uuid.UUID `db:"id"             json:"id"`
	Query         string    `db:"query"          json:"query"`
	DestinationID int       `db:"destination_id" json:"destination_id"`
	Label         string    `db:"label"          json:"label"`
	CityName      *string   `db:"city_name"      json:"city_name"`
	ProvinceName  *string   `db:"province_name"  json:"province_name"`
	ZipCode       *string   `db:"zip_code"       json:"zip_code"`
	CreatedAt     time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updated_at"`
}

// ShippingProviderInfo menggambarkan keadaan layanan tarif yang sedang dipakai,
// supaya menu Pengaturan bisa menjelaskan apa adanya: apakah RajaOngkir aktif,
// kota asal mana yang dipakai, dan kurir apa saja yang ditanyakan.
type ShippingProviderInfo struct {
	// Name adalah sumber tarif yang sedang aktif, misalnya "RajaOngkir" atau
	// "tabel tarif".
	Name string `json:"name"`
	// Connected berarti API key terisi dan layanannya terpasang. Kalau false,
	// perhitungan ongkir jatuh ke tabel tarif yang dikelola sendiri.
	Connected bool `json:"connected"`
	// Ready berarti sudah bisa dipakai menghitung: terhubung dan kota asalnya
	// sudah dipilih.
	Ready         bool            `json:"ready"`
	OriginID      int             `json:"origin_id"`
	OriginLabel   string          `json:"origin_label"`
	Couriers      []string        `json:"couriers"`
	CourierOption []CourierOption `json:"courier_options"`
}

// CourierOption adalah satu kurir yang bisa dicentang di menu Pengaturan.
type CourierOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// KurirRajaOngkir adalah kurir domestik yang lazim dipakai toko jastip.
//
// Daftarnya ditulis di sini, bukan diambil dari API, karena RajaOngkir tidak
// menyediakan endpoint daftar kurir pada paket langganan mana pun — dan tim
// toko hanya perlu memilih dari yang benar-benar mereka pakai.
var KurirRajaOngkir = []CourierOption{
	{Code: "jne", Name: "JNE"},
	{Code: "jnt", Name: "J&T Express"},
	{Code: "sicepat", Name: "SiCepat"},
	{Code: "anteraja", Name: "AnterAja"},
	{Code: "ninja", Name: "Ninja Xpress"},
	{Code: "tiki", Name: "TIKI"},
	{Code: "pos", Name: "POS Indonesia"},
	{Code: "wahana", Name: "Wahana"},
	{Code: "lion", Name: "Lion Parcel"},
	{Code: "ide", Name: "ID Express"},
	{Code: "sap", Name: "SAP Express"},
	{Code: "ncs", Name: "NCS"},
	{Code: "rex", Name: "REX"},
	{Code: "sentral", Name: "Sentral Cargo"},
}
