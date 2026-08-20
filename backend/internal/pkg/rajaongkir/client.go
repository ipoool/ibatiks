// Package rajaongkir membungkus API tarif kirim RajaOngkir (Komerce).
//
// Dua endpoint yang dipakai aplikasi ini:
//
//	GET  destination/domestic-destination?search=…  mencari ID tujuan
//	POST calculate/domestic-cost                    menghitung ongkir
//	POST track/waybill                              melacak resi yang sudah ada
//
// Keduanya memakai header "key" berisi API key, dan membungkus jawabannya dalam
// amplop {meta, data}. Kode status HTTP tidak selalu bisa dipercaya sendirian —
// meta.code ikut diperiksa.
package rajaongkir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client memanggil API RajaOngkir. Aman dipakai bersamaan dari banyak goroutine.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func New(apiKey, baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

// Enabled benar bila kliennya punya kredensial untuk dipakai.
func (c *Client) Enabled() bool { return c != nil && c.apiKey != "" && c.baseURL != "" }

// Destination adalah satu tujuan pengiriman beserta ID yang dipakai RajaOngkir.
type Destination struct {
	ID              int    `json:"id"`
	Label           string `json:"label"`
	SubdistrictName string `json:"subdistrict_name"`
	DistrictName    string `json:"district_name"`
	CityName        string `json:"city_name"`
	ProvinceName    string `json:"province_name"`
	ZipCode         string `json:"zip_code"`
}

// Cost adalah satu pilihan layanan kirim beserta ongkos dan estimasi tibanya.
type Cost struct {
	Name        string `json:"name"`        // nama kurir, misal "JNE"
	Code        string `json:"code"`        // kode kurir, misal "jne"
	Service     string `json:"service"`     // nama layanan, misal "REG"
	Description string `json:"description"` // keterangan layanan
	Cost        int    `json:"cost"`        // ongkos dalam rupiah
	ETD         string `json:"etd"`         // estimasi tiba, misal "2-3 day"
}

type amplop[T any] struct {
	Meta struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Status  string `json:"status"`
	} `json:"meta"`
	Data T `json:"data"`
}

// Error adalah kegagalan yang datang dari RajaOngkir sendiri, bukan dari
// jaringan. Dipisah supaya pemanggil bisa membedakan "keynya salah" dari
// "internetnya mati" dan memilih pesan yang tepat untuk admin.
type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("rajaongkir menolak permintaan (%d): %s", e.StatusCode, e.Message)
}

// SearchDestination mencari tujuan berdasarkan nama kota, kecamatan, kelurahan,
// atau kode pos. Hasilnya diurutkan oleh RajaOngkir dari yang paling cocok.
func (c *Client) SearchDestination(ctx context.Context, q string, limit int) ([]Destination, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	params := url.Values{}
	params.Set("search", q)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", "0")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/destination/domestic-destination?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var out amplop[[]Destination]
	if err := c.kirim(req, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// DomesticCost menghitung ongkir dari satu ID tujuan ke ID tujuan lain.
//
// Berat dikirim dalam gram; RajaOngkir sendiri yang membulatkan ke kilogram
// sesuai aturan tiap kurir. couriers diisi kode kurir dipisah koma, misalnya
// "jne:sicepat:jnt" — RajaOngkir memakai titik dua sebagai pemisah.
func (c *Client) DomesticCost(ctx context.Context, origin, destination, weightGram int, couriers string) ([]Cost, error) {
	if origin <= 0 || destination <= 0 {
		return nil, fmt.Errorf("rajaongkir: id asal dan tujuan wajib diisi")
	}
	if weightGram <= 0 {
		weightGram = 1000
	}
	couriers = strings.TrimSpace(couriers)
	if couriers == "" {
		couriers = "jne"
	}

	form := url.Values{}
	form.Set("origin", strconv.Itoa(origin))
	form.Set("destination", strconv.Itoa(destination))
	form.Set("weight", strconv.Itoa(weightGram))
	form.Set("courier", couriers)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/calculate/domestic-cost", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var out amplop[[]Cost]
	if err := c.kirim(req, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// Waybill adalah hasil pelacakan sebuah resi.
//
// Bentuk balasan pelacakan tidak seragam antar kurir, jadi yang dibaca hanya
// bagian yang benar-benar dibutuhkan dan sisanya dibiarkan. Field yang tidak
// ada akan bernilai kosong, bukan membuat seluruh pembacaan gagal — laporan
// posisi paket yang sebagian terbaca masih berguna, sedangkan galat parsing
// tidak berguna sama sekali.
type Waybill struct {
	Delivered bool `json:"delivered"`
	Summary   struct {
		Status       string `json:"status"`
		CourierName  string `json:"courier_name"`
		WaybillNo    string `json:"waybill_number"`
		ServiceCode  string `json:"service_code"`
		ShipperName  string `json:"shipper_name"`
		ReceiverName string `json:"receiver_name"`
	} `json:"summary"`
	DeliveryStatus struct {
		Status      string `json:"status"`
		PODReceiver string `json:"pod_receiver"`
		PODDate     string `json:"pod_date"`
		PODTime     string `json:"pod_time"`
	} `json:"delivery_status"`
	Manifest []struct {
		Description string `json:"manifest_description"`
		Date        string `json:"manifest_date"`
		Time        string `json:"manifest_time"`
		City        string `json:"city_name"`
	} `json:"manifest"`
}

// TrackWaybill menanyakan posisi sebuah resi ke kurir lewat RajaOngkir.
//
// Resi yang baru saja diserahkan sering belum dikenali sistem kurir; dalam
// keadaan itu RajaOngkir menjawab "Invalid Awb". Pemanggil yang membedakan
// "salah ketik" dari "belum masuk sistem" harus melihat pesannya, bukan sekadar
// menganggap resinya palsu.
func (c *Client) TrackWaybill(ctx context.Context, awb, courier string) (*Waybill, error) {
	awb = strings.TrimSpace(awb)
	courier = strings.ToLower(strings.TrimSpace(courier))
	if awb == "" || courier == "" {
		return nil, fmt.Errorf("rajaongkir: nomor resi dan kurir wajib diisi")
	}

	form := url.Values{}
	form.Set("awb", awb)
	form.Set("courier", courier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/track/waybill", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var out amplop[Waybill]
	if err := c.kirim(req, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

func (c *Client) kirim(req *http.Request, out any) error {
	if !c.Enabled() {
		return &Error{StatusCode: http.StatusUnauthorized, Message: "API key RajaOngkir belum diisi"}
	}
	req.Header.Set("key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("rajaongkir tidak bisa dihubungi: %w", err)
	}
	defer resp.Body.Close()

	// Badan dibaca lebih dulu supaya pesan galat dari RajaOngkir tetap terbaca
	// walau kode HTTP-nya sudah menandakan gagal.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("rajaongkir: jawaban tidak terbaca: %w", err)
	}

	var meta amplop[json.RawMessage]
	_ = json.Unmarshal(body, &meta)

	if resp.StatusCode >= 400 || (meta.Meta.Code != 0 && meta.Meta.Code >= 400) {
		pesan := strings.TrimSpace(meta.Meta.Message)
		if pesan == "" {
			pesan = strings.TrimSpace(string(body))
		}
		kode := meta.Meta.Code
		if kode == 0 {
			kode = resp.StatusCode
		}
		return &Error{StatusCode: kode, Message: pesan}
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("rajaongkir: jawaban tidak dikenali: %w", err)
	}
	return nil
}
