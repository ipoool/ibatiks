package domain

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ShippingRate adalah tarif kirim per kilogram untuk satu kota tujuan.
type ShippingRate struct {
	ID              uuid.UUID       `db:"id"               json:"id"`
	Courier         string          `db:"courier"          json:"courier"`
	Service         string          `db:"service"          json:"service"`
	DestinationCity string          `db:"destination_city" json:"destination_city"`
	Province        *string         `db:"province"         json:"province"`
	PricePerKg      decimal.Decimal `db:"price_per_kg"     json:"price_per_kg"`
	MinWeightGram   int             `db:"min_weight_gram"  json:"min_weight_gram"`
	ETD             *string         `db:"etd"              json:"etd"`
	CreatedAt       time.Time       `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"       json:"updated_at"`
}

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

	// Source menjelaskan asal tarifnya: "tabel tarif", "tarif default", atau
	// nama vendor kalau nanti dipasang integrasi API.
	Source string `json:"source"`
	// RateFound bernilai false kalau kota tujuan belum ada di tabel tarif dan
	// perhitungannya memakai tarif default.
	RateFound bool `json:"rate_found"`
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
