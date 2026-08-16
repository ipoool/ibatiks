// Package notify menyusun pesan untuk customer.
//
// Pengiriman WhatsApp dilakukan manual oleh admin: sistem menyiapkan teks yang
// sudah terisi lengkap plus tautan wa.me, admin tinggal menekannya dan menekan
// kirim. Pendekatan ini tidak butuh gateway berbayar, dan pesannya tetap
// terkirim dari nomor toko sendiri sehingga customer mengenali pengirimnya.
package notify

import (
	"net/url"
	"strings"
	"time"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
)

// Message adalah pesan siap kirim beserta tautan pembukanya.
type Message struct {
	Phone string `json:"phone"`
	Text  string `json:"text"`
	// WhatsAppURL membuka WhatsApp Web/aplikasi dengan pesan sudah terisi.
	WhatsAppURL string `json:"whatsapp_url"`
	// MailtoURL disediakan sebagai alternatif kalau customer lebih suka email.
	MailtoURL string `json:"mailto_url,omitempty"`
}

// Render mengganti seluruh placeholder {{kunci}} pada template dengan nilainya.
// Placeholder yang tidak dikenal dibiarkan apa adanya supaya salah ketik pada
// template langsung terlihat oleh admin, bukan menghilang diam-diam.
func Render(template string, values map[string]string) string {
	result := template
	for key, value := range values {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}

// WhatsAppURL membentuk tautan wa.me dengan teks pesan yang sudah di-encode.
func WhatsAppURL(phone, text string) string {
	normalized := domain.NormalizePhoneWA(phone)
	if normalized == "" {
		return ""
	}
	return "https://wa.me/" + normalized + "?text=" + url.QueryEscape(text)
}

func MailtoURL(email, subject, body string) string {
	if strings.TrimSpace(email) == "" {
		return ""
	}
	return "mailto:" + email +
		"?subject=" + url.QueryEscape(subject) +
		"&body=" + url.QueryEscape(body)
}

// DPRequestParams adalah data untuk pesan permintaan uang muka.
type DPRequestParams struct {
	Settings     domain.Settings
	Customer     *domain.Customer
	Order        *domain.Order
	TripTitle    string
	OrderNumber  string
	DPAmountText string
}

func BuildDPRequest(p DPRequestParams) Message {
	text := Render(p.Settings.GetOr(domain.SettingWATemplateDP, defaultDPTemplate), map[string]string{
		"customer_name": p.Customer.Name,
		"store_name":    p.Settings.GetOr(domain.SettingStoreName, "Ibatiks"),
		"trip_title":    p.TripTitle,
		"order_number":  p.OrderNumber,
		"total":         money.Format(p.Order.Total),
		"dp_amount":     p.DPAmountText,
		"bank_account":  p.Settings.Get(domain.SettingBankAccount),
	})

	return Message{
		Phone:       p.Customer.PhoneWA,
		Text:        text,
		WhatsAppURL: WhatsAppURL(p.Customer.PhoneWA, text),
		MailtoURL:   mailtoIfPresent(p.Customer.Email, "Permintaan DP "+p.OrderNumber, text),
	}
}

// InvoiceParams adalah data untuk pesan penagihan pelunasan.
type InvoiceParams struct {
	Settings domain.Settings
	Customer *domain.Customer
	Invoice  *domain.Invoice
	Order    *domain.Order
}

func BuildInvoiceMessage(p InvoiceParams) Message {
	text := Render(p.Settings.GetOr(domain.SettingWATemplateInv, defaultInvoiceTemplate), map[string]string{
		"customer_name":  p.Customer.Name,
		"store_name":     p.Settings.GetOr(domain.SettingStoreName, "Ibatiks"),
		"invoice_number": p.Invoice.InvoiceNumber,
		"order_number":   p.Order.OrderNumber,
		"total":          money.Format(p.Invoice.Total),
		"amount_paid":    money.Format(p.Invoice.AmountPaid),
		"amount_due":     money.Format(p.Invoice.AmountDue),
		"bank_account":   p.Settings.Get(domain.SettingBankAccount),
		"due_date":       formatDate(p.Invoice.DueDate),
	})

	return Message{
		Phone:       p.Customer.PhoneWA,
		Text:        text,
		WhatsAppURL: WhatsAppURL(p.Customer.PhoneWA, text),
		MailtoURL:   mailtoIfPresent(p.Customer.Email, "Invoice "+p.Invoice.InvoiceNumber, text),
	}
}

// ShipmentParams adalah data untuk pesan informasi pengiriman.
type ShipmentParams struct {
	Settings domain.Settings
	Customer *domain.Customer
	Order    *domain.Order
	Shipment *domain.Shipment
}

func BuildShipmentMessage(p ShipmentParams) Message {
	tracking := ""
	if p.Shipment.TrackingNumber != nil {
		tracking = *p.Shipment.TrackingNumber
	}

	text := Render(p.Settings.GetOr(domain.SettingWATemplateShip, defaultShipmentTemplate), map[string]string{
		"customer_name":   p.Customer.Name,
		"store_name":      p.Settings.GetOr(domain.SettingStoreName, "Ibatiks"),
		"order_number":    p.Order.OrderNumber,
		"recipient_name":  p.Order.RecipientName,
		"courier":         p.Shipment.Courier,
		"service":         p.Shipment.Service,
		"tracking_number": tracking,
	})

	return Message{
		Phone:       p.Customer.PhoneWA,
		Text:        text,
		WhatsAppURL: WhatsAppURL(p.Customer.PhoneWA, text),
		MailtoURL:   mailtoIfPresent(p.Customer.Email, "Pengiriman "+p.Order.OrderNumber, text),
	}
}

func mailtoIfPresent(email *string, subject, body string) string {
	if email == nil {
		return ""
	}
	return MailtoURL(*email, subject, body)
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("02 Jan 2006")
}

// Template cadangan kalau baris pengaturannya terhapus dari database.
const (
	defaultDPTemplate = "Halo {{customer_name}}, terima kasih sudah order.\n" +
		"Total pesanan: {{total}}\nDP: {{dp_amount}}\nTransfer ke: {{bank_account}}"

	defaultInvoiceTemplate = "Halo {{customer_name}}, invoice {{invoice_number}}.\n" +
		"Total: {{total}}\nSudah dibayar: {{amount_paid}}\nSisa: {{amount_due}}\n" +
		"Transfer ke: {{bank_account}}"

	defaultShipmentTemplate = "Halo {{customer_name}}, pesanan {{order_number}} sudah dikirim.\n" +
		"Kurir: {{courier}} {{service}}\nNo. resi: {{tracking_number}}"
)
