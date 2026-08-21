package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Kunci pengaturan yang dikenal aplikasi. Nilai defaultnya di-seed oleh migrasi
// 000008_support.
const (
	SettingStoreName      = "store_name"
	SettingStorePhone     = "store_phone"
	SettingStoreEmail     = "store_email"
	SettingStoreAddress   = "store_address"
	SettingBankAccount    = "bank_account"
	SettingInvoiceFooter  = "invoice_footer"
	SettingInvoiceDueDays = "invoice_due_days"
	SettingWATemplateDP   = "wa_template_dp"
	SettingWATemplateInv  = "wa_template_invoice"
	SettingWATemplateShip = "wa_template_shipped"
)

type Setting struct {
	Key         string    `db:"key"         json:"key"`
	Value       string    `db:"value"       json:"value"`
	Description *string   `db:"description" json:"description"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

// Settings adalah kumpulan pengaturan dalam bentuk peta agar gampang dipakai
// saat merender invoice dan menyusun pesan WA.
type Settings map[string]string

func (s Settings) Get(key string) string { return s[key] }

func (s Settings) GetOr(key, fallback string) string {
	if v, ok := s[key]; ok && v != "" {
		return v
	}
	return fallback
}

// Aksi yang dicatat pada audit log.
const (
	AuditCreate        = "create"
	AuditUpdate        = "update"
	AuditDelete        = "delete"
	AuditStatusChange  = "status_change"
	AuditPaymentRecord = "payment_record"
	AuditItemChange    = "item_change"
)

type AuditLog struct {
	ID       int64      `db:"id"         json:"id"`
	UserID   *uuid.UUID `db:"user_id"    json:"user_id"`
	Entity   string     `db:"entity"     json:"entity"`
	EntityID *uuid.UUID `db:"entity_id"  json:"entity_id"`
	Action   string     `db:"action"     json:"action"`
	// Changes wajib json.RawMessage, bukan []byte. encoding/json meng-encode
	// []byte jadi base64, jadi seluruh detail jejak perubahan sampai ke layar
	// sebagai deretan huruf yang tidak berarti apa-apa bagi siapa pun.
	Changes   json.RawMessage `db:"changes"    json:"changes"`
	IPAddress *string         `db:"ip_address" json:"ip_address"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

type AuditLogDetail struct {
	AuditLog
	UserName *string `db:"user_name" json:"user_name"`
}

// AuditActor adalah satu akun yang pernah muncul di jejak perubahan.
type AuditActor struct {
	ID       uuid.UUID `db:"id"        json:"id"`
	Name     string    `db:"name"      json:"name"`
	LogCount int       `db:"log_count" json:"log_count"`
}
