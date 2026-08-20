// Package pdf merender invoice menjadi berkas PDF yang siap dikirim ke customer.
package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
)

// Logo ikut ditanam ke dalam binary supaya invoice tetap berkop walau image
// production hanya berisi satu berkas biner tanpa direktori aset.
//
//go:embed assets/logo-ibatiks.png
var logoPNG []byte

// Ukuran dan posisi dalam milimeter pada kertas A4 (210 x 297).
const (
	pageWidth   = 210.0
	marginLeft  = 15.0
	marginRight = 15.0
	contentW    = pageWidth - marginLeft - marginRight

	// Lebar logo pada kop invoice; tingginya mengikuti rasio berkas aslinya.
	logoWidth  = 20.0
	logoHeight = logoWidth * 1027.0 / 900.0

	// Margin atas halaman, dipakai juga sebagai batas atas saat menghitung
	// titik tengah isi untuk watermark.
	marginAtas = 15.0
)

type InvoiceData struct {
	Invoice  *domain.Invoice
	Order    *domain.Order
	Items    []domain.OrderItem
	Customer *domain.Customer
	Trip     *domain.Trip
	Payments []domain.Payment
	Settings domain.Settings
}

type Renderer struct {
	outputDir string
}

func NewRenderer(outputDir string) *Renderer {
	return &Renderer{outputDir: outputDir}
}

// Render menulis PDF ke disk dan mengembalikan path berkasnya.
func (r *Renderer) Render(data InvoiceData) (string, error) {
	doc, err := r.build(data)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(r.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("buat direktori invoice: %w", err)
	}

	filename := sanitizeFilename(data.Invoice.InvoiceNumber) + ".pdf"
	path := filepath.Join(r.outputDir, filename)

	if err := doc.OutputFileAndClose(path); err != nil {
		return "", fmt.Errorf("tulis PDF invoice: %w", err)
	}
	return path, nil
}

/*
 * build merender invoice dua kali.
 *
 * Watermark harus berada di tengah isi invoice, bukan di tengah kertas — isi
 * satu halaman A4 biasanya berhenti jauh di atas batas bawahnya, dan watermark
 * yang dipatok ke tengah kertas mendarat di area kosong di bawah tabel.
 *
 * Tapi ia juga harus digambar lebih dulu supaya angka invoice berada di atasnya;
 * digambar belakangan, ia menutupi yang justru harus terbaca. Dua tuntutan itu
 * berlawanan, jadi jalan pintasnya: render sekali tanpa watermark hanya untuk
 * mengukur di mana isinya berhenti, lalu render ulang dengan watermark yang
 * sudah tahu titik tengahnya. Satu halaman A4 hitungannya milidetik.
 */
func (r *Renderer) build(data InvoiceData) (*fpdf.Fpdf, error) {
	ukur, err := r.render(data, 0)
	if err != nil {
		return nil, err
	}

	_, tinggiHalaman := ukur.GetPageSize()
	tengah := tinggiHalaman / 2
	// Isi yang tumpah ke halaman kedua sudah memenuhi halaman pertama, jadi
	// tengah kertas justru tepat.
	if ukur.PageNo() == 1 {
		if bawah := ukur.GetY(); bawah > marginAtas {
			tengah = (marginAtas + bawah) / 2
		}
	}

	return r.render(data, tengah)
}

func (r *Renderer) render(data InvoiceData, watermarkY float64) (*fpdf.Fpdf, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginLeft, marginAtas, marginRight)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// fpdf memakai encoding CP1252; teks Indonesia aman karena tidak memakai
	// karakter di luar Latin-1, tapi tetap ditranslasikan agar aman.
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	if watermarkY > 0 {
		watermark(pdf, tr, data.Invoice.Status, watermarkY)
	}

	r.header(pdf, tr, data)
	r.parties(pdf, tr, data)
	r.itemTable(pdf, tr, data)
	r.totals(pdf, tr, data)
	r.paymentInfo(pdf, tr, data)
	r.footer(pdf, tr, data)

	if pdf.Err() {
		return nil, fmt.Errorf("render PDF: %w", pdf.Error())
	}
	return pdf, nil
}

/*
 * watermark menuliskan keadaan invoice melintang di tengah halaman.
 *
 * Dua keadaan yang perlu terbaca sekilas dari kertasnya sendiri:
 *
 *   LUNAS  — supaya lembar yang sudah dibayar tidak ikut ditagihkan lagi.
 *            Nomor dan nominalnya sama persis dengan invoice yang dulu dikirim,
 *            jadi tanpa penanda ini keduanya tidak bisa dibedakan.
 *   DRAFT  — supaya lembar yang belum pernah dikirim ke customer tidak dikira
 *            tagihan resmi kalau terlanjur tercetak atau diteruskan.
 *   BATAL  — supaya tagihan yang sudah dicabut tidak terbayar oleh customer
 *            yang terlanjur memegang cetakannya.
 *
 * Hanya "terkirim" yang tanpa watermark: itu keadaan normal sebuah tagihan yang
 * memang sedang menunggu dibayar.
 */
func watermark(pdf *fpdf.Fpdf, tr func(string) string, status string, tengahY float64) {
	var teks string
	switch status {
	case domain.InvoicePaid:
		teks = "LUNAS"
	case domain.InvoiceDraft:
		teks = "DRAFT"
	case domain.InvoiceVoid:
		teks = "BATAL"
	default:
		return
	}

	lebar, _ := pdf.GetPageSize()

	// Abu-abu muda, bukan transparansi: SetAlpha menuntut PDF versi 1.4 ke atas
	// dan tidak semua pembaca struk menampilkannya. Warna terang menghasilkan
	// tampilan yang sama di mana pun tanpa syarat itu.
	pdf.SetTextColor(232, 232, 232)
	pdf.SetFont("Helvetica", "B", 90)

	lebarTeks := pdf.GetStringWidth(teks)

	// Diputar 45 derajat pada titik tengah halaman. TransformBegin/End dipakai
	// berpasangan; tanpa End, seluruh gambar sesudahnya ikut miring.
	pdf.TransformBegin()
	pdf.TransformRotate(45, lebar/2, tengahY)
	pdf.SetXY(lebar/2-lebarTeks/2, tengahY-16)
	pdf.CellFormat(lebarTeks, 32, tr(teks), "", 0, "C", false, 0, "")
	pdf.TransformEnd()

	// Warna dikembalikan supaya isi invoice tidak ikut pucat.
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(marginLeft, marginAtas)
}

func (r *Renderer) header(pdf *fpdf.Fpdf, tr func(string) string, data InvoiceData) {
	storeName := data.Settings.GetOr(domain.SettingStoreName, "Ibatiks")

	topY := pdf.GetY()
	textLeft := marginLeft
	if drawLogo(pdf, marginLeft, topY) {
		textLeft = marginLeft + logoWidth + 5
	}

	// Seluruh teks kop digeser ke kanan logo, jadi lebarnya dihitung ulang —
	// memakai contentW akan membuat kolom kanan meleset keluar margin.
	textW := pageWidth - marginRight - textLeft
	pdf.SetXY(textLeft, topY)

	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(textW*0.55, 9, tr(storeName), "", 0, "L", false, 0, "")

	title := "INVOICE"
	if data.Invoice.Type == domain.InvoiceDP {
		title = "INVOICE DP"
	}
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(textW*0.45, 9, tr(title), "", 1, "R", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(90, 90, 90)

	storeLines := []string{
		data.Settings.Get(domain.SettingStoreAddress),
		joinNonEmpty(" · ", data.Settings.Get(domain.SettingStorePhone), data.Settings.Get(domain.SettingStoreEmail)),
	}
	invoiceLines := []string{
		"No. " + data.Invoice.InvoiceNumber,
		"Tanggal: " + data.Invoice.IssueDate.Format("02 Jan 2006"),
	}
	if data.Invoice.DueDate != nil {
		invoiceLines = append(invoiceLines, "Jatuh tempo: "+data.Invoice.DueDate.Format("02 Jan 2006"))
	}

	for i := 0; i < maxInt(len(storeLines), len(invoiceLines)); i++ {
		pdf.SetX(textLeft)
		pdf.CellFormat(textW*0.55, 5, tr(nthOrEmpty(storeLines, i)), "", 0, "L", false, 0, "")
		pdf.CellFormat(textW*0.45, 5, tr(nthOrEmpty(invoiceLines, i)), "", 1, "R", false, 0, "")
	}

	// Garis pemisah tidak boleh naik memotong logo ketika teks kopnya pendek.
	if bottom := topY + logoHeight; pdf.GetY() < bottom {
		pdf.SetY(bottom)
	}

	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(4)
	pdf.SetDrawColor(210, 210, 210)
	pdf.Line(marginLeft, pdf.GetY(), pageWidth-marginRight, pdf.GetY())
	pdf.Ln(6)
}

// drawLogo menempatkan logo di pojok kiri kop dan melaporkan berhasil-tidaknya.
//
// Kegagalan sengaja tidak dijadikan error: invoice tanpa logo masih sah dan
// masih bisa ditagihkan, sedangkan menolak menerbitkan invoice hanya karena
// gambarnya bermasalah akan menghentikan penagihan tanpa alasan yang sepadan.
func drawLogo(pdf *fpdf.Fpdf, x, y float64) bool {
	const name = "logo-ibatiks"

	if pdf.GetImageInfo(name) == nil {
		pdf.RegisterImageOptionsReader(
			name,
			fpdf.ImageOptions{ImageType: "PNG"},
			bytes.NewReader(logoPNG),
		)
		if pdf.Err() {
			pdf.ClearError()
			return false
		}
	}

	pdf.ImageOptions(name, x, y, logoWidth, 0, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	if pdf.Err() {
		pdf.ClearError()
		return false
	}
	return true
}

func (r *Renderer) parties(pdf *fpdf.Fpdf, tr func(string) string, data InvoiceData) {
	startY := pdf.GetY()

	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(contentW/2, 5, tr("DITAGIHKAN KEPADA"), "", 0, "L", false, 0, "")
	pdf.CellFormat(contentW/2, 5, tr("DIKIRIM KE"), "", 1, "L", false, 0, "")

	billTo := []string{
		data.Customer.Name,
		data.Customer.PhoneWA,
	}
	if data.Customer.Email != nil {
		billTo = append(billTo, *data.Customer.Email)
	}

	shipTo := []string{
		data.Order.RecipientName,
		data.Order.RecipientPhone,
		data.Order.ShippingAddress,
		joinNonEmpty(", ",
			derefStr(data.Order.ShippingSubdistrict),
			derefStr(data.Order.ShippingDistrict)),
		joinNonEmpty(", ",
			data.Order.ShippingCity,
			derefStr(data.Order.ShippingProvince),
			derefStr(data.Order.ShippingPostalCode)),
	}

	pdf.SetFont("Helvetica", "", 9)
	for i := 0; i < maxInt(len(billTo), len(shipTo)); i++ {
		y := pdf.GetY()
		pdf.SetXY(marginLeft, y)
		pdf.MultiCell(contentW/2-3, 5, tr(nthOrEmpty(billTo, i)), "", "L", false)
		leftY := pdf.GetY()

		pdf.SetXY(marginLeft+contentW/2, y)
		pdf.MultiCell(contentW/2, 5, tr(nthOrEmpty(shipTo, i)), "", "L", false)
		rightY := pdf.GetY()

		pdf.SetXY(marginLeft, maxFloat(leftY, rightY))
	}

	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(90, 90, 90)
	pdf.CellFormat(contentW, 5, tr(fmt.Sprintf("Order %s · Trip %s (%s)",
		data.Order.OrderNumber, data.Trip.Title, data.Trip.Code)), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	if pdf.GetY() < startY+20 {
		pdf.SetY(startY + 20)
	}
	pdf.Ln(4)
}

func (r *Renderer) itemTable(pdf *fpdf.Fpdf, tr func(string) string, data InvoiceData) {
	// Lebar kolom: nama produk, qty, harga satuan, subtotal.
	widths := []float64{contentW - 20 - 32 - 32, 20, 32, 32}
	headers := []string{"Produk", "Qty", "Harga", "Subtotal"}
	aligns := []string{"L", "C", "R", "R"}

	pdf.SetFillColor(242, 242, 242)
	pdf.SetFont("Helvetica", "B", 9)
	for i, header := range headers {
		pdf.CellFormat(widths[i], 8, tr(header), "1", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 9)
	for _, item := range data.Items {
		// Baris produk yang tidak tersedia tetap dicetak agar customer paham
		// mengapa totalnya berbeda dari pesanan awal.
		name := item.ProductName
		if item.FulfillmentStatus == domain.FulfillmentUnavailable {
			name += " (tidak tersedia)"
		}

		lineHeight := 7.0
		x, y := pdf.GetX(), pdf.GetY()

		pdf.MultiCell(widths[0], lineHeight, tr(name), "1", "L", false)
		rowHeight := maxFloat(pdf.GetY()-y, lineHeight)

		pdf.SetXY(x+widths[0], y)
		pdf.CellFormat(widths[1], rowHeight, fmt.Sprintf("%d", item.Qty), "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[2], rowHeight, tr(money.Format(item.UnitPrice)), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[3], rowHeight, tr(money.Format(item.Subtotal)), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(3)
}

func (r *Renderer) totals(pdf *fpdf.Fpdf, tr func(string) string, data InvoiceData) {
	labelW := 40.0
	valueW := 40.0
	offset := contentW - labelW - valueW

	row := func(label, value string, bold bool) {
		style := ""
		if bold {
			style = "B"
		}
		pdf.SetX(marginLeft + offset)
		pdf.SetFont("Helvetica", style, 10)
		pdf.CellFormat(labelW, 6, tr(label), "", 0, "R", false, 0, "")
		pdf.CellFormat(valueW, 6, tr(value), "", 1, "R", false, 0, "")
	}

	row("Subtotal", money.Format(data.Invoice.Subtotal), false)
	if data.Invoice.Discount.IsPositive() {
		row("Diskon", "-"+money.Format(data.Invoice.Discount), false)
	}
	if data.Invoice.ShippingFee.IsPositive() {
		row("Ongkir", money.Format(data.Invoice.ShippingFee), false)
	}

	line := func() {
		pdf.SetX(marginLeft + offset)
		pdf.SetDrawColor(180, 180, 180)
		pdf.Line(marginLeft+offset, pdf.GetY(), pageWidth-marginRight, pdf.GetY())
		pdf.Ln(1)
	}

	line()
	// Nilai pesanan seutuhnya, sama pada invoice DP maupun pelunasan: itulah
	// harga yang disepakati, dan customer harus melihatnya di dokumen mana pun.
	row("Total tagihan", money.Format(data.Invoice.Total), true)

	if data.Invoice.Type == domain.InvoiceDP {
		row("Down payment", money.Format(data.Invoice.DPAmount), false)
		if data.Invoice.AmountPaid.IsPositive() {
			row("Sudah dibayar", "-"+money.Format(data.Invoice.AmountPaid), false)
		}
		line()
		row("Ditagihkan sekarang", money.Format(data.Invoice.AmountDue), true)
		pdf.Ln(4)
		return
	}

	if data.Invoice.DPAmount.IsPositive() {
		row("Down payment", "-"+money.Format(data.Invoice.DPAmount), false)
	}
	// Pembayaran di luar uang muka dipisah barisnya supaya angka DP tetap
	// terbaca sebagai DP, bukan bercampur dengan cicilan lain.
	if others := data.Invoice.AmountPaid.Sub(data.Invoice.DPAmount); others.IsPositive() {
		row("Pembayaran lain", "-"+money.Format(others), false)
	}
	line()
	row("Sisa ditagihkan", money.Format(data.Invoice.AmountDue), true)

	pdf.Ln(4)
}

func (r *Renderer) paymentInfo(pdf *fpdf.Fpdf, tr func(string) string, data InvoiceData) {
	if len(data.Payments) > 0 {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(contentW, 6, tr("RIWAYAT PEMBAYARAN"), "", 1, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		for _, payment := range data.Payments {
			label := paymentLabel(payment.Type)
			amount := money.Format(payment.Amount)
			if payment.Type == domain.PaymentRefund {
				amount = "-" + amount
			}
			pdf.CellFormat(contentW*0.7, 5, tr(fmt.Sprintf("%s · %s · %s",
				payment.PaidAt.Format("02 Jan 2006"), label, payment.Method)), "", 0, "L", false, 0, "")
			pdf.CellFormat(contentW*0.3, 5, tr(amount), "", 1, "R", false, 0, "")
		}
		pdf.Ln(3)
	}

	if bank := data.Settings.Get(domain.SettingBankAccount); bank != "" && data.Invoice.AmountDue.IsPositive() {
		pdf.SetFillColor(248, 248, 248)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(contentW, 6, tr("TRANSFER PEMBAYARAN KE"), "1", 1, "L", true, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(contentW, 8, tr(bank), "1", 1, "L", false, 0, "")
		pdf.Ln(3)
	}
}

func (r *Renderer) footer(pdf *fpdf.Fpdf, tr func(string) string, data InvoiceData) {
	if note := data.Settings.Get(domain.SettingInvoiceFooter); note != "" {
		pdf.SetFont("Helvetica", "I", 9)
		pdf.SetTextColor(90, 90, 90)
		pdf.MultiCell(contentW, 5, tr(note), "", "L", false)
		pdf.SetTextColor(0, 0, 0)
	}
}

func paymentLabel(paymentType string) string {
	switch paymentType {
	case domain.PaymentDP:
		return "DP"
	case domain.PaymentSettlement:
		return "Pelunasan"
	case domain.PaymentRefund:
		return "Refund"
	default:
		return "Penyesuaian"
	}
}

func sanitizeFilename(v string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "_", ":", "-")
	return replacer.Replace(v)
}

func joinNonEmpty(sep string, parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			filtered = append(filtered, p)
		}
	}
	return strings.Join(filtered, sep)
}

func nthOrEmpty(values []string, i int) string {
	if i < len(values) {
		return values[i]
	}
	return ""
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
