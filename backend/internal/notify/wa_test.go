package notify

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

// Regresi: ISSUE-001 — invoice DP dikirim dengan template pelunasan.
// Ditemukan lewat /qa pada 19 Agustus 2026.
// Laporan: .gstack/qa-reports/qa-report-ibatiks-2026-08-19.md
//
// Gejalanya: customer yang baru menyetujui pesanannya menerima pesan "barang
// pesananmu sudah sampai di Indonesia" pada hari ia memesan, lengkap dengan
// label "Sisa pelunasan" untuk uang muka yang belum ia bayar.
func TestBuildDPInvoiceMessageMemakaiTemplateDP(t *testing.T) {
	settings := domain.Settings{
		domain.SettingWATemplateDP: "Halo {{customer_name}}, terima kasih sudah order di trip {{trip_title}}.\n" +
			"Invoice {{invoice_number}}\nTotal pesanan: {{total}}\nDP: {{dp_amount}}",
		domain.SettingWATemplateInv: "Halo {{customer_name}}, barang pesananmu sudah sampai di Indonesia.\n" +
			"Sisa pelunasan: {{amount_due}}",
		domain.SettingBankAccount: "BCA 123 a/n Ibatiks",
	}
	customer := &domain.Customer{Name: "Sari Dewi", PhoneWA: "6285712345678"}
	order := &domain.Order{OrderNumber: "ORD-2026-0004", Total: decimal.NewFromInt(252000)}
	invoice := &domain.Invoice{
		InvoiceNumber: "INV-2026-0003",
		Type:          domain.InvoiceDP,
		Total:         decimal.NewFromInt(210000),
		AmountDue:     decimal.NewFromInt(105000),
	}

	got := BuildInvoiceMessageFor(DPInvoiceParams{
		Settings:  settings,
		Customer:  customer,
		Invoice:   invoice,
		Order:     order,
		TripTitle: "Trip Bangkok",
	}).Text

	if strings.Contains(got, "sudah sampai di Indonesia") {
		t.Errorf("invoice DP memakai template pelunasan:\n%s", got)
	}
	for _, mau := range []string{"Sari Dewi", "Trip Bangkok", "INV-2026-0003", "Rp210.000", "Rp105.000"} {
		if !strings.Contains(got, mau) {
			t.Errorf("pesan invoice DP tidak memuat %q:\n%s", mau, got)
		}
	}
}

// Invoice pelunasan tetap memakai template pelunasan — perbaikan di atas tidak
// boleh menukar keduanya ke arah sebaliknya.
func TestBuildInvoiceMessageTetapTemplatePelunasan(t *testing.T) {
	settings := domain.Settings{
		domain.SettingWATemplateInv: "Halo {{customer_name}}, barang pesananmu sudah sampai.\n" +
			"Invoice {{invoice_number}}\nSisa pelunasan: {{amount_due}}",
	}
	got := BuildInvoiceMessageFor(DPInvoiceParams{
		Settings: settings,
		Customer: &domain.Customer{Name: "Sari Dewi", PhoneWA: "628571234"},
		Order:    &domain.Order{OrderNumber: "ORD-2026-0004"},
		Invoice: &domain.Invoice{
			InvoiceNumber: "INV-2026-0004",
			Type:          domain.InvoiceFinal,
			AmountDue:     decimal.NewFromInt(212000),
		},
		TripTitle: "Trip Bangkok",
	}).Text

	if !strings.Contains(got, "sudah sampai") {
		t.Errorf("invoice pelunasan tidak memakai template pelunasan:\n%s", got)
	}
	if !strings.Contains(got, "Rp212.000") {
		t.Errorf("sisa pelunasan tidak muncul di pesan:\n%s", got)
	}
}
