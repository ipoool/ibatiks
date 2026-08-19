package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/notify"
	"github.com/ipoool/jastipin/backend/internal/pdf"
	"github.com/ipoool/jastipin/backend/internal/pkg/docnum"
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type InvoiceService struct {
	pool      *pgxpool.Pool
	invoices  *repository.InvoiceRepo
	orders    *repository.OrderRepo
	customers *repository.CustomerRepo
	trips     *repository.TripRepo
	settings  *repository.SettingsRepo
	audit     *repository.AuditRepo
	renderer  *pdf.Renderer
}

func NewInvoiceService(
	pool *pgxpool.Pool,
	invoices *repository.InvoiceRepo,
	orders *repository.OrderRepo,
	customers *repository.CustomerRepo,
	trips *repository.TripRepo,
	settings *repository.SettingsRepo,
	audit *repository.AuditRepo,
	renderer *pdf.Renderer,
) *InvoiceService {
	return &InvoiceService{
		pool: pool, invoices: invoices, orders: orders, customers: customers,
		trips: trips, settings: settings, audit: audit, renderer: renderer,
	}
}

type CreateInvoiceInput struct {
	Type    string
	DueDate *time.Time
	Notes   *string
}

// Create menerbitkan invoice untuk sebuah order.
//
// Seluruh nominal disalin ke baris invoice, bukan direferensikan dari order.
// Dengan begitu, PDF yang sudah dikirim ke customer tetap cocok dengan angka
// yang tersimpan meskipun order-nya diedit setelah itu.
func (s *InvoiceService) Create(ctx context.Context, orderID uuid.UUID, in CreateInvoiceInput, actorID uuid.UUID) (*domain.Invoice, error) {
	if !domain.IsValidInvoiceType(in.Type) {
		return nil, domain.Validation("jenis invoice tidak dikenal", map[string]string{
			"type": "pilih dp atau final",
		})
	}

	var invoice *domain.Invoice
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if order.Status == domain.OrderCancelled {
			return domain.InvalidState("order sudah dibatalkan, invoice tidak bisa diterbitkan")
		}
		if order.Total.LessThanOrEqual(decimal.Zero) {
			return domain.InvalidState("order belum punya nilai, tambahkan item terlebih dahulu")
		}

		// Invoice pelunasan menagih seluruh sisa pesanan, termasuk ongkir.
		// Menerbitkannya sebelum ongkir diketahui berarti mengirim tagihan yang
		// nilainya masih akan berubah — customer membayar lalu ditagih lagi,
		// dan dokumen yang sudah ia terima jadi tidak cocok dengan yang
		// tercatat di sistem.
		if in.Type == domain.InvoiceFinal {
			if order.Status == domain.OrderAwaitingDP {
				return domain.InvalidState(
					"DP order ini belum masuk — tagih uang mukanya dulu lewat invoice DP")
			}
			if order.ShippingFee.LessThanOrEqual(decimal.Zero) {
				return domain.InvalidState(
					"ongkir belum ditetapkan — timbang paketnya dulu di menu Pengiriman")
			}
		}

		settings, err := s.settings.All(ctx, tx)
		if err != nil {
			return err
		}

		amounts := s.calculateAmounts(order, in.Type)
		dueDate := in.DueDate
		if dueDate == nil {
			dueDate = defaultDueDate(settings, order.OrderDate)
		}

		number, err := docnum.Next(ctx, tx, docnum.Invoice, time.Now().Year())
		if err != nil {
			return err
		}

		invoice, err = s.invoices.Create(ctx, tx, repository.InvoiceParams{
			InvoiceNumber: number,
			OrderID:       orderID,
			Type:          in.Type,
			IssueDate:     time.Now(),
			DueDate:       dueDate,
			Subtotal:      amounts.Subtotal,
			Discount:      amounts.Discount,
			ShippingFee:   amounts.ShippingFee,
			Total:         amounts.Total,
			DPAmount:      amounts.DPAmount,
			AmountPaid:    amounts.AmountPaid,
			AmountDue:     amounts.AmountDue,
			Notes:         trimPtr(in.Notes),
			CreatedBy:     nullableUUID(actorID),
		})
		if err != nil {
			return err
		}

		// Menerbitkan invoice tidak mengubah status order. Yang memindahkan
		// order ke Pembayaran Lunas adalah uang yang benar-benar masuk, bukan
		// dokumen yang dikirimkan — invoice terkirim yang tak kunjung dibayar
		// tidak boleh terlihat seperti kemajuan.

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "invoice",
			EntityID: &invoice.ID,
			Action:   domain.AuditCreate,
			Changes: map[string]any{
				"invoice_number": number, "type": in.Type, "total": amounts.Total.String(),
			},
		})
	})
	if err != nil {
		return nil, err
	}

	// PDF dirender setelah transaksi commit: kegagalan menulis berkas tidak
	// boleh membatalkan invoice yang secara data sudah sah.
	if path, err := s.renderPDF(ctx, invoice); err == nil {
		invoice.PDFPath = &path
	}

	return invoice, nil
}

type invoiceAmounts struct {
	Subtotal    decimal.Decimal
	Discount    decimal.Decimal
	ShippingFee decimal.Decimal
	Total       decimal.Decimal
	DPAmount    decimal.Decimal
	AmountPaid  decimal.Decimal
	AmountDue   decimal.Decimal
}

// calculateAmounts membedakan invoice DP dari invoice pelunasan.
//
// Nilai order ditulis apa adanya pada kedua jenis invoice — subtotal barang,
// diskon, ongkir, dan totalnya. Yang membedakan hanyalah apa yang ditagih:
// invoice DP menagih uang mukanya, invoice pelunasan menagih sisanya setelah
// uang muka dikurangkan. Dengan begitu customer selalu melihat harga pesanan
// yang sebenarnya, bukan dokumen yang seolah-olah menyatakan pesanannya cuma
// seharga uang muka.
func (s *InvoiceService) calculateAmounts(order *domain.Order, invoiceType string) invoiceAmounts {
	amounts := invoiceAmounts{
		Subtotal:    order.Subtotal,
		Discount:    order.Discount,
		ShippingFee: order.ShippingFee,
		Total:       order.Total,
		DPAmount:    order.DPRequired,
	}

	if invoiceType == domain.InvoiceDP {
		paid := decimal.Min(order.PaidAmount, order.DPRequired)
		amounts.AmountPaid = paid
		amounts.AmountDue = money.Max(order.DPRequired.Sub(paid), decimal.Zero)
		return amounts
	}

	// Pada invoice pelunasan yang ditulis sebagai uang muka adalah yang benar-
	// benar sudah diterima, bukan yang diminta: kalau customer membayar kurang
	// dari DP yang disepakati, sisanya tetap ikut ditagih di sini.
	amounts.DPAmount = decimal.Min(order.PaidAmount, order.DPRequired)
	amounts.AmountPaid = order.PaidAmount
	amounts.AmountDue = money.Max(order.BalanceDue, decimal.Zero)
	return amounts
}

// Candidates mendaftar order yang siap ditagih pelunasannya.
func (s *InvoiceService) Candidates(ctx context.Context, search string) ([]domain.InvoiceCandidate, error) {
	return s.invoices.ListCandidates(ctx, s.pool, strings.TrimSpace(search))
}

func (s *InvoiceService) Get(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	return s.invoices.GetByID(ctx, s.pool, id)
}

func (s *InvoiceService) List(ctx context.Context, p pagination.Params, status, invoiceType string) ([]domain.InvoiceListItem, int64, error) {
	return s.invoices.List(ctx, s.pool, p, status, invoiceType)
}

func (s *InvoiceService) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]domain.Invoice, error) {
	return s.invoices.ListByOrder(ctx, s.pool, orderID)
}

// PDFPath mengembalikan path berkas PDF, merender ulang kalau berkasnya belum
// pernah dibuat atau sudah hilang dari disk.
func (s *InvoiceService) PDFPath(ctx context.Context, id uuid.UUID) (string, string, error) {
	invoice, err := s.invoices.GetByID(ctx, s.pool, id)
	if err != nil {
		return "", "", err
	}

	if invoice.PDFPath != nil && fileExists(*invoice.PDFPath) {
		return *invoice.PDFPath, invoice.InvoiceNumber, nil
	}

	path, err := s.renderPDF(ctx, invoice)
	if err != nil {
		return "", "", err
	}
	return path, invoice.InvoiceNumber, nil
}

func (s *InvoiceService) renderPDF(ctx context.Context, invoice *domain.Invoice) (string, error) {
	order, err := s.orders.GetByID(ctx, s.pool, invoice.OrderID)
	if err != nil {
		return "", err
	}
	items, err := s.orders.ListItems(ctx, s.pool, order.ID)
	if err != nil {
		return "", err
	}
	payments, err := s.orders.ListPayments(ctx, s.pool, order.ID)
	if err != nil {
		return "", err
	}
	customer, err := s.customers.GetByID(ctx, s.pool, order.CustomerID)
	if err != nil {
		return "", err
	}
	trip, err := s.trips.GetByID(ctx, s.pool, order.TripID)
	if err != nil {
		return "", err
	}
	settings, err := s.settings.All(ctx, s.pool)
	if err != nil {
		return "", err
	}

	path, err := s.renderer.Render(pdf.InvoiceData{
		Invoice:  invoice,
		Order:    order,
		Items:    items,
		Customer: customer,
		Trip:     trip,
		Payments: payments,
		Settings: settings,
	})
	if err != nil {
		return "", domain.Internal(err)
	}

	if err := s.invoices.SetPDFPath(ctx, s.pool, invoice.ID, path); err != nil {
		return "", err
	}
	return path, nil
}

// Message menyusun teks penagihan siap kirim beserta tautan wa.me.
func (s *InvoiceService) Message(ctx context.Context, id uuid.UUID) (*notify.Message, error) {
	invoice, err := s.invoices.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	order, err := s.orders.GetByID(ctx, s.pool, invoice.OrderID)
	if err != nil {
		return nil, err
	}
	customer, err := s.customers.GetByID(ctx, s.pool, order.CustomerID)
	if err != nil {
		return nil, err
	}
	settings, err := s.settings.All(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	trip, err := s.trips.GetByID(ctx, s.pool, order.TripID)
	if err != nil {
		return nil, err
	}

	// Invoice DP dan invoice pelunasan memakai template yang berbeda karena
	// bicara pada saat yang berbeda; pemilihannya ada di notify supaya bisa
	// diuji tanpa database.
	message := notify.BuildInvoiceMessageFor(notify.DPInvoiceParams{
		Settings:  settings,
		Customer:  customer,
		Invoice:   invoice,
		Order:     order,
		TripTitle: trip.Title,
	})
	return &message, nil
}

// MarkSent mencatat bahwa invoice sudah dikirim lewat kanal tertentu.
// Dipanggil setelah admin menekan tombol kirim WA, supaya sistem tahu tagihan
// mana yang belum pernah disampaikan ke customer.
func (s *InvoiceService) MarkSent(ctx context.Context, id uuid.UUID, channel string, actorID uuid.UUID) (*domain.Invoice, error) {
	if !domain.IsValidSentChannel(channel) {
		return nil, domain.Validation("kanal pengiriman tidak dikenal", map[string]string{
			"channel": "pilih wa, email, atau manual",
		})
	}

	var invoice *domain.Invoice
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		var err error
		invoice, err = s.invoices.MarkSent(ctx, tx, id, channel)
		if err != nil {
			return err
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "invoice",
			EntityID: &id,
			Action:   "sent",
			Changes:  map[string]any{"channel": channel},
		})
	})
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

// Void membatalkan invoice yang salah terbit. Invoice yang sudah lunas tidak
// bisa dibatalkan karena uangnya sudah tercatat masuk.
func (s *InvoiceService) Void(ctx context.Context, id uuid.UUID, actorID uuid.UUID) (*domain.Invoice, error) {
	invoice, err := s.invoices.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	if invoice.Status == domain.InvoicePaid {
		return nil, domain.InvalidState("invoice sudah lunas dan tidak bisa dibatalkan")
	}

	var voided *domain.Invoice
	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		var err error
		voided, err = s.invoices.Void(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "invoice",
			EntityID: &id,
			Action:   "void",
		})
	})
	if err != nil {
		return nil, err
	}
	return voided, nil
}

// DPMessage menyiapkan pesan permintaan uang muka untuk sebuah order.
func (s *InvoiceService) DPMessage(ctx context.Context, orderID uuid.UUID) (*notify.Message, error) {
	order, err := s.orders.GetByID(ctx, s.pool, orderID)
	if err != nil {
		return nil, err
	}
	customer, err := s.customers.GetByID(ctx, s.pool, order.CustomerID)
	if err != nil {
		return nil, err
	}
	trip, err := s.trips.GetByID(ctx, s.pool, order.TripID)
	if err != nil {
		return nil, err
	}
	settings, err := s.settings.All(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	message := notify.BuildDPRequest(notify.DPRequestParams{
		Settings:     settings,
		Customer:     customer,
		Order:        order,
		TripTitle:    trip.Title,
		OrderNumber:  order.OrderNumber,
		DPAmountText: money.Format(order.DPRequired),
	})
	return &message, nil
}

// defaultDueDate menghitung jatuh tempo dari pengaturan invoice_due_days.
func defaultDueDate(settings domain.Settings, from time.Time) *time.Time {
	days, err := strconv.Atoi(settings.GetOr(domain.SettingInvoiceDueDays, "3"))
	if err != nil || days < 0 {
		days = 3
	}
	due := time.Now().AddDate(0, 0, days)
	return &due
}
