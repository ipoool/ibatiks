package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

/*
 * Ukuran label kurir yang lazim di Indonesia: 100 × 150 mm, seukuran kertas
 * thermal yang dipakai JNE, J&T, dan SiCepat. Marginnya sempit karena
 * printer thermal mencetak nyaris sampai tepi, dan setiap milimeter yang
 * disisakan mengurangi ruang untuk alamat.
 */
const (
	labelWidth   = 100.0
	labelHeight  = 150.0
	labelMargin  = 5.0
	labelContent = labelWidth - labelMargin*2
)

// LabelData adalah bahan label pengiriman: siapa mengirim, ke siapa, dan lewat
// kurir mana.
//
// Tanpa daftar barang dan tanpa nominal. Label ditempel di luar kardus dan
// terbaca siapa pun yang memegang paket di jalan — isi belanjaan dan harganya
// bukan urusan mereka.
type LabelData struct {
	Order    *domain.Order
	Shipment *domain.Shipment
	Settings domain.Settings
}

// RenderLabel membuat label pengiriman sebagai berkas PDF di memori.
//
// Tidak disimpan ke disk seperti invoice: isinya seluruhnya bisa dibentuk ulang
// dari order, jadi mencetak ulang selalu menghasilkan label yang sesuai keadaan
// terkini — termasuk saat nomor resinya baru terisi setelah label pertama
// dicetak.
func (r *Renderer) RenderLabel(data LabelData) ([]byte, error) {
	pdf := fpdf.NewCustom(&fpdf.InitType{
		UnitStr: "mm",
		Size:    fpdf.SizeType{Wd: labelWidth, Ht: labelHeight},
	})
	pdf.SetMargins(labelMargin, labelMargin, labelMargin)
	pdf.SetAutoPageBreak(false, labelMargin)
	pdf.AddPage()

	tr := pdf.UnicodeTranslatorFromDescriptor("")

	labelHeader(pdf, tr, data)
	labelPengirim(pdf, tr, pengirimDariSettings(data.Settings))
	labelPenerima(pdf, tr, penerimaDariOrder(data.Order))
	labelFooter(pdf, tr, data)

	if pdf.Err() {
		return nil, fmt.Errorf("render label: %w", pdf.Error())
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("tulis label: %w", err)
	}
	return buf.Bytes(), nil
}

// labelHeader memuat kurir dan nomor resi — dua hal yang dicari petugas kurir
// lebih dulu, jadi ditaruh paling atas dan paling besar.
func labelHeader(pdf *fpdf.Fpdf, tr func(string) string, data LabelData) {
	kurir := "—"
	resi := ""
	if data.Shipment != nil {
		kurir = strings.TrimSpace(data.Shipment.Courier + " " + data.Shipment.Service)
		if data.Shipment.TrackingNumber != nil {
			resi = strings.TrimSpace(*data.Shipment.TrackingNumber)
		}
	}

	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(labelContent, 10, tr(kurir), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 15)
	if resi == "" {
		// Label sering dicetak sebelum resi didapat, lalu ditempel dan
		// nomornya ditulis tangan di konter. Ruangnya disediakan, bukan
		// dihilangkan.
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(labelContent, 8, tr("No. resi: ______________________"), "", 1, "L", false, 0, "")
	} else {
		pdf.CellFormat(labelContent, 8, tr(resi), "", 1, "L", false, 0, "")
	}

	pdf.Ln(1)
	labelGaris(pdf)
}

type pihakLabel struct {
	Nama   string
	Telp   string
	Alamat []string
}

func pengirimDariSettings(settings domain.Settings) pihakLabel {
	return pihakLabel{
		Nama: settings.GetOr(domain.SettingStoreName, "Ibatiks"),
		Telp: settings.Get(domain.SettingStorePhone),
		Alamat: []string{
			settings.Get(domain.SettingStoreAddress),
		},
	}
}

func penerimaDariOrder(order *domain.Order) pihakLabel {
	// Urutannya mengikuti cara alamat Indonesia ditulis: jalan, lalu kelurahan,
	// kecamatan, kota, provinsi, dan kode pos.
	wilayah := joinNonEmpty(", ",
		derefStr(order.ShippingSubdistrict),
		derefStr(order.ShippingDistrict),
		order.ShippingCity,
		derefStr(order.ShippingProvince),
	)
	if kode := derefStr(order.ShippingPostalCode); kode != "" {
		wilayah = joinNonEmpty(" ", wilayah, kode)
	}

	return pihakLabel{
		Nama:   order.RecipientName,
		Telp:   order.RecipientPhone,
		Alamat: []string{order.ShippingAddress, wilayah},
	}
}

// labelPengirim ditulis kecil dan rapat. Alamat toko tidak dibaca kurir saat
// mengantar — gunanya hanya kalau paket harus dikembalikan.
func labelPengirim(pdf *fpdf.Fpdf, tr func(string) string, pihak pihakLabel) {
	pdf.Ln(3)
	labelJudulBagian(pdf, tr, "PENGIRIM")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(labelContent, 5, tr(joinNonEmpty("  ·  ", pihak.Nama, pihak.Telp)), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	for _, baris := range pihak.Alamat {
		if strings.TrimSpace(baris) == "" {
			continue
		}
		pdf.MultiCell(labelContent, 4, tr(baris), "", "L", false)
	}

	pdf.Ln(2)
	labelGaris(pdf)
}

// labelPenerima mengambil porsi terbesar label.
//
// Inilah satu-satunya bagian yang benar-benar dibaca orang di jalan, dan
// dibacanya sambil berdiri memegang tumpukan paket. Ukurannya dibuat jauh lebih
// besar dari bagian lain supaya terbaca sekali lihat, dan diberi bingkai supaya
// mata langsung jatuh ke sana.
func labelPenerima(pdf *fpdf.Fpdf, tr func(string) string, pihak pihakLabel) {
	pdf.Ln(3)
	labelJudulBagian(pdf, tr, "PENERIMA")

	atas := pdf.GetY()
	pdf.SetY(atas + 2)

	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetX(labelMargin + 3)
	pdf.MultiCell(labelContent-6, 7, tr(pihak.Nama), "", "L", false)

	if pihak.Telp != "" {
		pdf.SetFont("Helvetica", "B", 13)
		pdf.SetX(labelMargin + 3)
		pdf.CellFormat(labelContent-6, 6, tr(pihak.Telp), "", 1, "L", false, 0, "")
	}

	pdf.Ln(1)
	pdf.SetFont("Helvetica", "", 11)
	for _, baris := range pihak.Alamat {
		if strings.TrimSpace(baris) == "" {
			continue
		}
		pdf.SetX(labelMargin + 3)
		pdf.MultiCell(labelContent-6, 5, tr(baris), "", "L", false)
	}

	bawah := pdf.GetY() + 2
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.4)
	pdf.Rect(labelMargin, atas, labelContent, bawah-atas, "D")
	pdf.SetLineWidth(0.2)
	pdf.SetY(bawah)
}

func labelJudulBagian(pdf *fpdf.Fpdf, tr func(string) string, judul string) {
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(110, 110, 110)
	pdf.CellFormat(labelContent, 4, tr(judul), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

// labelFooter menyebut nomor order dan berat paket. Nomor order adalah cara
// tim toko mencocokkan label yang tercetak dengan kardus yang tepat sebelum
// ditempel — begitu tertukar, dua customer menerima barang orang lain.
func labelFooter(pdf *fpdf.Fpdf, tr func(string) string, data LabelData) {
	// Ditambatkan ke dasar label, bukan mengalir setelah alamat: panjang alamat
	// berbeda-beda, dan footer yang ikut naik-turun membuat tiap label tercetak
	// dengan tata letak yang sedikit berlainan.
	pdf.SetY(labelHeight - 22)
	labelGaris(pdf)
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(labelContent, 5, tr(data.Order.OrderNumber), "", 1, "L", false, 0, "")

	if data.Shipment != nil && data.Shipment.WeightGram > 0 {
		pdf.SetFont("Helvetica", "", 10)
		berat := fmt.Sprintf("%.1f kg", float64(data.Shipment.WeightGram)/1000)
		if data.Shipment.LengthCM > 0 && data.Shipment.WidthCM > 0 && data.Shipment.HeightCM > 0 {
			berat += fmt.Sprintf("  ·  %d × %d × %d cm",
				data.Shipment.LengthCM, data.Shipment.WidthCM, data.Shipment.HeightCM)
		}
		pdf.CellFormat(labelContent, 5, tr(berat), "", 1, "L", false, 0, "")
	}

	if data.Shipment != nil && data.Shipment.Notes != nil {
		catatan := strings.TrimSpace(*data.Shipment.Notes)
		if catatan != "" {
			pdf.SetFont("Helvetica", "B", 10)
			pdf.MultiCell(labelContent, 4.5, tr(catatan), "", "L", false)
		}
	}
}

func labelGaris(pdf *fpdf.Fpdf) {
	y := pdf.GetY()
	pdf.SetDrawColor(160, 160, 160)
	pdf.Line(labelMargin, y, labelWidth-labelMargin, y)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetY(y)
}
