package handler

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/ipoool/jastipin/backend/internal/http/request"
	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type ReportHandler struct {
	reports *service.ReportService
}

func NewReportHandler(reports *service.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports}
}

func (h *ReportHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	summary, err := h.reports.Dashboard(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, summary)
}

// TripProfit adalah laporan inti: berapa untung sebuah perjalanan.
//
// Tanpa trip_id, seluruh trip dijumlahkan jadi satu laporan. Bentuk jawabannya
// sama persis untuk keduanya supaya satu tampilan bisa melayani dua-duanya —
// yang berbeda hanya identitas tripnya, yang memang tidak ada saat semuanya
// dijumlahkan.
func (h *ReportHandler) TripProfit(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDQuery(r, "trip_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	from, err := request.DateQuery(r, "from")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	to, err := request.DateQuery(r, "to")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	/*
	 * Ekspornya satu baris per trip, bukan satu baris berisi angka gabungan
	 * yang sudah terbaca di layar. Yang dibawa ke spreadsheet biasanya untuk
	 * membandingkan trip — dan angka gabungannya tinggal dijumlah sendiri di
	 * sana, sementara pemecahannya tidak bisa dibalik.
	 */
	if r.URL.Query().Get("format") == "csv" {
		rows, err := h.reports.TripProfitRows(r.Context(), tripID, from, to)
		if err != nil {
			response.Error(w, r, err)
			return
		}

		writeCSV(w, r, "laba-rugi", []string{
			"kode_trip", "trip", "negara", "status",
			"jumlah_order", "customer", "qty",
			"omzet", "hpp", "laba_kotor", "biaya_perjalanan", "laba_bersih", "margin_persen",
			"modal_keluar", "uang_masuk", "sisa_tagihan",
			"ongkir_ditagihkan", "ongkir_dibayar", "diskon",
			"surplus_qty", "surplus_nilai",
		}, func(write func([]string) error) error {
			for _, t := range rows {
				if err := write([]string{
					t.TripCode, t.Title, t.Country, t.Status,
					strconv.Itoa(t.OrderCount), strconv.Itoa(t.CustomerCount), strconv.Itoa(t.ItemQty),
					t.Revenue.String(), t.COGS.String(), t.GrossProfit.String(),
					t.TripExpenses.String(), t.NetProfit.String(), t.MarginPct.String(),
					t.TotalCapitalOut.String(), t.PaymentReceived.String(), t.Outstanding.String(),
					t.ShippingFeeCollected.String(), t.ShippingCostPaid.String(), t.DiscountGiven.String(),
					strconv.Itoa(t.SurplusStockQty), t.SurplusStockValue.String(),
				}); err != nil {
					return err
				}
			}
			return nil
		})
		return
	}

	report, err := h.reports.TripProfit(r.Context(), tripID, from, to)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, report)
}

func (h *ReportHandler) OrderProfits(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)
	// Ekspor mengambil seluruh baris, bukan sehalaman. Tombolnya bernama
	// "Ekspor CSV", bukan "Ekspor halaman ini".
	if r.URL.Query().Get("format") == "csv" {
		p = p.ForExport()
	}

	tripID, err := request.UUIDQuery(r, "trip_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	from, err := request.DateQuery(r, "from")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	to, err := request.DateQuery(r, "to")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	profits, total, err := h.reports.OrderProfits(r.Context(), p, tripID, from, to)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if r.URL.Query().Get("format") == "csv" {
		writeCSV(w, r, "profit-order", []string{
			"nomor_order", "customer", "trip", "status", "tanggal", "omzet", "hpp", "profit", "margin_persen",
		}, func(write func([]string) error) error {
			for _, item := range profits {
				if err := write([]string{
					item.OrderNumber, item.CustomerName, item.TripCode, item.Status,
					item.OrderDate.Format("2006-01-02"),
					item.Revenue.String(), item.COGS.String(), item.Profit.String(), item.MarginPct.String(),
				}); err != nil {
					return err
				}
			}
			return nil
		})
		return
	}

	response.Paginated(w, profits, p.Page, p.PerPage, total)
}

// Receivables mendaftar tagihan yang belum lunas, urut dari yang paling lama.
func (h *ReportHandler) Receivables(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)
	// Ekspor mengambil seluruh baris, bukan sehalaman. Tombolnya bernama
	// "Ekspor CSV", bukan "Ekspor halaman ini".
	if r.URL.Query().Get("format") == "csv" {
		p = p.ForExport()
	}

	from, err := request.DateQuery(r, "from")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	to, err := request.DateQuery(r, "to")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	receivables, total, err := h.reports.Receivables(r.Context(), p, from, to)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if r.URL.Query().Get("format") == "csv" {
		writeCSV(w, r, "piutang", []string{
			"nomor_order", "customer", "telepon", "trip", "status", "tanggal",
			"total", "sudah_bayar", "sisa", "umur_hari",
		}, func(write func([]string) error) error {
			for _, item := range receivables {
				if err := write([]string{
					item.OrderNumber, item.CustomerName, item.CustomerPhone, item.TripCode, item.Status,
					item.OrderDate.Format("2006-01-02"),
					item.Total.String(), item.PaidAmount.String(), item.BalanceDue.String(),
					strconv.Itoa(item.DaysOutstanding),
				}); err != nil {
					return err
				}
			}
			return nil
		})
		return
	}

	response.Paginated(w, receivables, p.Page, p.PerPage, total)
}

// CustomerSales menampilkan rekap penjualan per customer.
func (h *ReportHandler) CustomerSales(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)

	tripID, err := request.UUIDQuery(r, "trip_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	from, err := request.DateQuery(r, "from")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	to, err := request.DateQuery(r, "to")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	/*
	 * Ekspor dipecah per kanal asal order, jadi satu customer bisa muncul lebih
	 * dari sekali. Di layar satu baris per customer sudah menjawab siapa
	 * pembelanja terbesar; pemecahannya berguna justru di spreadsheet — untuk
	 * menilai kanal mana yang mendatangkan pelanggan bernilai, bukan sekadar
	 * kanal mana yang ramai.
	 *
	 * Kuerinya sendiri, bukan hasil layar yang dibumbui: yang tampil di layar
	 * cuma satu halaman, dan ekspor yang berhenti di halaman pertama kehilangan
	 * sebagian besar datanya tanpa tanda apa pun di berkasnya.
	 */
	if r.URL.Query().Get("format") == "csv" {
		rows, err := h.reports.CustomerSalesByChannel(r.Context(), tripID, from, to)
		if err != nil {
			response.Error(w, r, err)
			return
		}

		writeCSV(w, r, "penjualan-customer", []string{
			"kode", "customer", "telepon", "kota", "channel", "jumlah_order", "qty",
			"omzet", "hpp", "profit", "piutang", "order_pertama", "order_terakhir",
		}, func(write func([]string) error) error {
			for _, item := range rows {
				city := ""
				if item.City != nil {
					city = *item.City
				}
				if err := write([]string{
					item.CustomerCode, item.CustomerName, item.CustomerPhone, city,
					item.Source,
					strconv.Itoa(item.OrderCount), strconv.Itoa(item.ItemQty),
					item.Revenue.String(), item.COGS.String(), item.Profit.String(),
					item.Outstanding.String(),
					formatDatePtr(item.FirstOrderAt), formatDatePtr(item.LastOrderAt),
				}); err != nil {
					return err
				}
			}
			return nil
		})
		return
	}

	rows, total, err := h.reports.CustomerSales(r.Context(), p, tripID, from, to)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Paginated(w, rows, p.Page, p.PerPage, total)
}

// ChannelSales menampilkan rekap penjualan per asal order.
func (h *ReportHandler) ChannelSales(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDQuery(r, "trip_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	from, err := request.DateQuery(r, "from")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	to, err := request.DateQuery(r, "to")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	rows, err := h.reports.ChannelSales(r.Context(), tripID, from, to)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if r.URL.Query().Get("format") == "csv" {
		writeCSV(w, r, "penjualan-channel", []string{
			"channel", "jumlah_order", "jumlah_customer", "qty",
			"omzet", "hpp", "profit", "rata_rata_order", "porsi_omzet_persen",
		}, func(write func([]string) error) error {
			for _, item := range rows {
				if err := write([]string{
					item.Source, strconv.Itoa(item.OrderCount), strconv.Itoa(item.CustomerCount),
					strconv.Itoa(item.ItemQty), item.Revenue.String(), item.COGS.String(),
					item.Profit.String(), item.AvgOrderValue.String(), item.RevenueShare.String(),
				}); err != nil {
					return err
				}
			}
			return nil
		})
		return
	}

	response.OK(w, rows)
}

func formatDatePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func (h *ReportHandler) ProductSales(w http.ResponseWriter, r *http.Request) {
	tripID, err := request.UUIDQuery(r, "trip_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	from, err := request.DateQuery(r, "from")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	to, err := request.DateQuery(r, "to")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	sales, err := h.reports.ProductSales(r.Context(), request.IntQuery(r, "limit", 20), tripID, from, to)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if r.URL.Query().Get("format") == "csv" {
		writeCSV(w, r, "penjualan-produk", []string{
			"sku", "produk", "kategori", "qty_terjual", "jumlah_order", "omzet", "hpp", "profit",
		}, func(write func([]string) error) error {
			for _, item := range sales {
				category := ""
				if item.CategoryName != nil {
					category = *item.CategoryName
				}
				if err := write([]string{
					item.ProductSKU, item.ProductName, category,
					strconv.Itoa(item.QtySold), strconv.Itoa(item.OrderCount),
					item.Revenue.String(), item.COGS.String(), item.Profit.String(),
				}); err != nil {
					return err
				}
			}
			return nil
		})
		return
	}

	response.OK(w, sales)
}

// writeCSV mengirim data sebagai berkas CSV yang bisa dibuka di Excel.
// BOM UTF-8 ditulis di depan supaya Excel di Windows tidak merusak karakter
// beraksen pada nama customer.
func writeCSV(w http.ResponseWriter, r *http.Request, name string, header []string, rows func(write func([]string) error) error) {
	filename := name + "-" + time.Now().Format("20060102") + ".csv"

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		return
	}
	if err := rows(writer.Write); err != nil {
		// Header sudah terkirim, jadi error hanya bisa dicatat di log server.
		response.Error(w, r, err)
	}
}
