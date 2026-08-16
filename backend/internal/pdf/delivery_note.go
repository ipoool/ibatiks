package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

// DeliveryNoteData adalah bahan surat jalan: siapa mengirim, ke siapa, isinya
// apa, dan lewat kurir mana.
type DeliveryNoteData struct {
	Order    *domain.Order
	Customer *domain.Customer
	Trip     *domain.Trip
	Items    []domain.OrderItem
	Shipment *domain.Shipment
	Settings domain.Settings
}

// RenderDeliveryNote membuat surat jalan sebagai berkas PDF di memori.
//
// Berbeda dengan invoice, surat jalan tidak disimpan ke disk: dokumennya
// dicetak saat paket diserahkan ke kurir dan isinya selalu bisa dibentuk ulang
// dari data order. Menyimpannya hanya menambah berkas yang tidak pernah dibuka
// lagi.
func (r *Renderer) RenderDeliveryNote(data DeliveryNoteData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginLeft, 15, marginRight)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	tr := pdf.UnicodeTranslatorFromDescriptor("")

	deliveryHeader(pdf, tr, data)
	deliveryParties(pdf, tr, data)
	deliveryCourier(pdf, tr, data)
	deliveryItems(pdf, tr, data)
	deliverySignatures(pdf, tr)

	if pdf.Err() {
		return nil, fmt.Errorf("render surat jalan: %w", pdf.Error())
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("tulis surat jalan: %w", err)
	}
	return buf.Bytes(), nil
}

func deliveryHeader(pdf *fpdf.Fpdf, tr func(string) string, data DeliveryNoteData) {
	storeName := data.Settings.GetOr(domain.SettingStoreName, "Ibatiks")

	topY := pdf.GetY()
	textLeft := marginLeft
	if drawLogo(pdf, marginLeft, topY) {
		textLeft = marginLeft + logoWidth + 5
	}
	textW := pageWidth - marginRight - textLeft
	pdf.SetXY(textLeft, topY)

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(textW*0.5, 8, tr(storeName), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(textW*0.5, 8, tr("SURAT JALAN"), "", 1, "R", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(90, 90, 90)

	left := []string{
		data.Settings.Get(domain.SettingStoreAddress),
		joinNonEmpty(" · ", data.Settings.Get(domain.SettingStorePhone), data.Settings.Get(domain.SettingStoreEmail)),
	}
	right := []string{
		"Order " + data.Order.OrderNumber,
		"Tanggal cetak: " + time.Now().Format("02 Jan 2006"),
	}

	for i := 0; i < maxInt(len(left), len(right)); i++ {
		pdf.SetX(textLeft)
		pdf.CellFormat(textW*0.5, 5, tr(nthOrEmpty(left, i)), "", 0, "L", false, 0, "")
		pdf.CellFormat(textW*0.5, 5, tr(nthOrEmpty(right, i)), "", 1, "R", false, 0, "")
	}

	if bottom := topY + logoHeight; pdf.GetY() < bottom {
		pdf.SetY(bottom)
	}

	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(4)
	pdf.SetDrawColor(210, 210, 210)
	pdf.Line(marginLeft, pdf.GetY(), pageWidth-marginRight, pdf.GetY())
	pdf.Ln(6)
}

func deliveryParties(pdf *fpdf.Fpdf, tr func(string) string, data DeliveryNoteData) {
	startY := pdf.GetY()

	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(contentW/2, 5, tr("PENGIRIM"), "", 0, "L", false, 0, "")
	pdf.CellFormat(contentW/2, 5, tr("PENERIMA"), "", 1, "L", false, 0, "")

	sender := nonEmptyLines(
		data.Settings.GetOr(domain.SettingStoreName, "Ibatiks"),
		data.Settings.Get(domain.SettingStorePhone),
		data.Settings.Get(domain.SettingStoreAddress),
	)

	// Alamat penerima ditulis lengkap sampai kelurahan dan kecamatan: itu yang
	// dipakai kurir menentukan gudang sortir, dan paket dengan alamat setengah
	// jadi berakhir tertahan.
	// Baris kosong dibuang, bukan dicetak sebagai spasi: alamat yang berlubang
	// di tengah membuat kurir ragu apakah ada bagian yang terpotong.
	receiver := nonEmptyLines(
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
	)

	pdf.SetFont("Helvetica", "", 9)
	for i := 0; i < maxInt(len(sender), len(receiver)); i++ {
		y := pdf.GetY()
		pdf.SetXY(marginLeft, y)
		pdf.MultiCell(contentW/2-3, 5, tr(nthOrEmpty(sender, i)), "", "L", false)
		leftY := pdf.GetY()

		pdf.SetXY(marginLeft+contentW/2, y)
		pdf.MultiCell(contentW/2, 5, tr(nthOrEmpty(receiver, i)), "", "L", false)
		rightY := pdf.GetY()

		pdf.SetXY(marginLeft, maxFloat(leftY, rightY))
	}

	if pdf.GetY() < startY+22 {
		pdf.SetY(startY + 22)
	}
	pdf.Ln(4)
}

func deliveryCourier(pdf *fpdf.Fpdf, tr func(string) string, data DeliveryNoteData) {
	courier, service := "", ""
	tracking, dimension, weight := "belum ada", "-", "-"

	if s := data.Shipment; s != nil {
		if s.Courier != "" {
			courier = s.Courier
		}
		if s.Service != "" {
			service = s.Service
		}
		if s.TrackingNumber != nil && *s.TrackingNumber != "" {
			tracking = *s.TrackingNumber
		}
		if s.WeightGram > 0 {
			weight = fmt.Sprintf("%d gram", s.WeightGram)
		}
		if s.LengthCM > 0 && s.WidthCM > 0 && s.HeightCM > 0 {
			dimension = fmt.Sprintf("%d × %d × %d cm", s.LengthCM, s.WidthCM, s.HeightCM)
		}
	}

	// Surat jalan boleh dicetak sebelum paket dikemas, jadi kolom yang belum
	// terisi dinyatakan apa adanya alih-alih menampilkan tanda hubung kosong
	// yang membingungkan petugas konter.
	courierText := joinNonEmpty(" ", courier, service)
	if courierText == "" {
		courierText = "belum ditentukan"
	}

	rows := [][2]string{
		{"Kurir", courierText},
		{"Nomor resi", tracking},
		{"Berat paket", weight},
		{"Dimensi", dimension},
	}

	pdf.SetFillColor(248, 248, 248)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(contentW, 6, tr("DATA PENGIRIMAN"), "1", 1, "L", true, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	for _, row := range rows {
		pdf.CellFormat(contentW*0.3, 6, tr("  "+row[0]), "1", 0, "L", false, 0, "")
		pdf.CellFormat(contentW*0.7, 6, tr("  "+row[1]), "1", 1, "L", false, 0, "")
	}
	pdf.Ln(5)
}

func deliveryItems(pdf *fpdf.Fpdf, tr func(string) string, data DeliveryNoteData) {
	widths := []float64{12, contentW - 12 - 40 - 20, 40, 20}
	headers := []string{"No", "Produk", "SKU", "Qty"}
	aligns := []string{"C", "L", "L", "R"}

	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Helvetica", "B", 9)
	for i, header := range headers {
		pdf.CellFormat(widths[i], 7, tr(header), "1", 0, aligns[i], true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 9)
	totalQty := 0
	for i, item := range data.Items {
		// Barang yang batal atau tidak jadi dibeli tidak ikut dikirim, jadi
		// tidak boleh muncul di surat jalan — kurir dan customer memakai
		// dokumen ini untuk mencocokkan isi paket.
		if item.FulfillmentStatus == domain.FulfillmentUnavailable ||
			item.FulfillmentStatus == domain.FulfillmentRefunded {
			continue
		}

		qty := item.Qty
		if item.QtyReceived > 0 && item.QtyReceived < item.Qty {
			qty = item.QtyReceived
		}
		totalQty += qty

		pdf.CellFormat(widths[0], 6, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[1], 6, tr(truncate(item.ProductName, 60)), "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[2], 6, tr(item.ProductSKU), "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[3], 6, fmt.Sprintf("%d", qty), "1", 1, "R", false, 0, "")
	}

	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(widths[0]+widths[1]+widths[2], 7, tr("Total barang"), "1", 0, "R", false, 0, "")
	pdf.CellFormat(widths[3], 7, fmt.Sprintf("%d", totalQty), "1", 1, "R", false, 0, "")

	if data.Order.Notes != nil && strings.TrimSpace(*data.Order.Notes) != "" {
		pdf.Ln(4)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(contentW, 5, tr("CATATAN"), "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.MultiCell(contentW, 5, tr(*data.Order.Notes), "", "L", false)
	}

	pdf.Ln(6)
}

func deliverySignatures(pdf *fpdf.Fpdf, tr func(string) string) {
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(90, 90, 90)
	pdf.MultiCell(contentW, 5, tr(
		"Barang diperiksa saat serah terima. Tanda tangan penerima menyatakan paket diterima "+
			"dalam keadaan baik dan jumlahnya sesuai daftar di atas."), "", "L", false)
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(8)

	columns := []string{"Diserahkan oleh (toko)", "Kurir", "Diterima oleh (penerima)"}
	colW := contentW / 3

	pdf.SetFont("Helvetica", "", 9)
	for _, label := range columns {
		pdf.CellFormat(colW, 5, tr(label), "", 0, "C", false, 0, "")
	}
	pdf.Ln(24)

	// Garis tanda tangan digambar terpisah supaya jaraknya tetap sama walau
	// nama kolomnya berbeda panjang.
	y := pdf.GetY()
	for i := range columns {
		x := marginLeft + colW*float64(i)
		pdf.Line(x+10, y, x+colW-10, y)
	}
	pdf.Ln(2)

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(120, 120, 120)
	for range columns {
		pdf.CellFormat(colW, 5, tr("( nama jelas )"), "", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(0, 0, 0)
}

func nonEmptyLines(lines ...string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
