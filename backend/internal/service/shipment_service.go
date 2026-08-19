package service

import (
	"context"
	"os"
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
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type ShipmentService struct {
	pool      *pgxpool.Pool
	shipments *repository.ShipmentRepo
	orders    *repository.OrderRepo
	customers *repository.CustomerRepo
	settings  *repository.SettingsRepo
	shipping  *ShippingService
	audit     *repository.AuditRepo
	trips     *repository.TripRepo
	renderer  *pdf.Renderer
}

func NewShipmentService(
	pool *pgxpool.Pool,
	shipments *repository.ShipmentRepo,
	orders *repository.OrderRepo,
	customers *repository.CustomerRepo,
	settings *repository.SettingsRepo,
	shipping *ShippingService,
	audit *repository.AuditRepo,
	trips *repository.TripRepo,
	renderer *pdf.Renderer,
) *ShipmentService {
	return &ShipmentService{
		pool: pool, shipments: shipments, orders: orders,
		customers: customers, settings: settings, shipping: shipping, audit: audit,
		trips: trips, renderer: renderer,
	}
}

// DeliveryNote menyusun surat jalan sebuah order sebagai PDF.
//
// Dokumennya dibentuk saat diminta, bukan disimpan: isinya seluruhnya berasal
// dari order dan paketnya, jadi mencetak ulang selalu menghasilkan dokumen yang
// sama dengan keadaan terkini.
func (s *ShipmentService) DeliveryNote(ctx context.Context, orderID uuid.UUID) ([]byte, string, error) {
	order, err := s.orders.GetByID(ctx, s.pool, orderID)
	if err != nil {
		return nil, "", err
	}
	items, err := s.orders.ListItems(ctx, s.pool, orderID)
	if err != nil {
		return nil, "", err
	}
	customer, err := s.customers.GetByID(ctx, s.pool, order.CustomerID)
	if err != nil {
		return nil, "", err
	}
	trip, err := s.trips.GetByID(ctx, s.pool, order.TripID)
	if err != nil {
		return nil, "", err
	}
	settings, err := s.settings.All(ctx, s.pool)
	if err != nil {
		return nil, "", err
	}

	// Paket belum tentu sudah dibuat: surat jalan boleh dicetak lebih dulu
	// sebagai lembar pendamping saat mengemas, dan kolom resinya menyusul.
	shipment, err := s.shipments.GetByOrder(ctx, s.pool, orderID)
	if err != nil {
		if domainErr, ok := domain.AsError(err); !ok || domainErr.Code != domain.CodeNotFound {
			return nil, "", err
		}
		shipment = nil
	}

	content, err := s.renderer.RenderDeliveryNote(pdf.DeliveryNoteData{
		Order:    order,
		Customer: customer,
		Trip:     trip,
		Items:    items,
		Shipment: shipment,
		Settings: settings,
	})
	if err != nil {
		return nil, "", domain.Internal(err)
	}
	return content, "SJ-" + order.OrderNumber, nil
}

type PackInput struct {
	Courier    string
	Service    string
	WeightGram int
	// Dimensi kardus dalam sentimeter. Boleh dikosongkan; kalau diisi, ongkir
	// dihitung memakai berat volumetrik bila hasilnya lebih besar dari berat asli.
	LengthCM int
	WidthCM  int
	HeightCM int
	Notes    *string
}

// Pack menandai order sudah dikemas atas nama customer dan siap dikirim.
func (s *ShipmentService) Pack(ctx context.Context, orderID uuid.UUID, in PackInput, actorID uuid.UUID) (*domain.Shipment, error) {
	courier := strings.ToUpper(strings.TrimSpace(in.Courier))
	if courier == "" {
		courier = "JNE"
	}
	serviceName := strings.ToUpper(strings.TrimSpace(in.Service))
	if serviceName == "" {
		serviceName = "REG"
	}
	if in.WeightGram < 0 || in.LengthCM < 0 || in.WidthCM < 0 || in.HeightCM < 0 {
		return nil, domain.Validation("ukuran paket tidak valid", map[string]string{
			"weight_gram": "berat dan dimensi harus 0 atau lebih",
		})
	}

	var shipment *domain.Shipment
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if order.Status == domain.OrderCancelled {
			return domain.InvalidState("order sudah dibatalkan")
		}

		// Estimasi ongkir disimpan bersama data kemasan supaya admin punya
		// pembanding saat memasukkan ongkir yang sebenarnya dibayar nanti.
		// Kegagalan menghitung tidak boleh membatalkan proses packing.
		estimated := decimal.Zero
		if estimate, estErr := s.shipping.Estimate(ctx, EstimateInput{
			Courier:    courier,
			Service:    serviceName,
			City:       order.ShippingCity,
			WeightGram: in.WeightGram,
			LengthCM:   in.LengthCM,
			WidthCM:    in.WidthCM,
			HeightCM:   in.HeightCM,
		}); estErr == nil {
			estimated = estimate.Cost
		}

		shipment, err = s.shipments.Pack(ctx, tx, repository.PackParams{
			OrderID:       orderID,
			Courier:       courier,
			Service:       serviceName,
			WeightGram:    in.WeightGram,
			LengthCM:      in.LengthCM,
			WidthCM:       in.WidthCM,
			HeightCM:      in.HeightCM,
			EstimatedCost: estimated,
			Notes:         trimPtr(in.Notes),
			PackedBy:      nullableUUID(actorID),
		})
		if err != nil {
			return err
		}

		// Mengemas tidak mengubah status order. Kemajuannya sudah terbaca dari
		// data kemasan ini sendiri — berat, dimensi, dan waktu dikemasnya —
		// jadi status tersendiri hanya akan jadi salinan yang bisa berbeda.

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "shipment",
			EntityID: &shipment.ID,
			Action:   "pack",
			Changes:  map[string]any{"courier": courier, "service": serviceName, "weight_gram": in.WeightGram},
		})
	})
	if err != nil {
		return nil, err
	}
	return shipment, nil
}

type ShipInput struct {
	TrackingNumber string
	ShippingCost   decimal.Decimal
	ShippedAt      *time.Time
	// AllowUnpaid mengizinkan pengiriman meski belum lunas, untuk kasus khusus
	// seperti pelanggan lama yang dipercaya membayar setelah barang diterima.
	AllowUnpaid bool
}

// Ship mencatat nomor resi JNE dan menandai paket sudah diserahkan ke kurir.
//
// Secara default order yang belum lunas ditolak: mengirim barang yang belum
// dibayar adalah keputusan sadar, bukan sesuatu yang boleh terjadi karena
// admin lupa mengecek.
func (s *ShipmentService) Ship(ctx context.Context, orderID uuid.UUID, in ShipInput, actorID uuid.UUID) (*domain.Shipment, error) {
	tracking := strings.ToUpper(strings.TrimSpace(in.TrackingNumber))
	if tracking == "" {
		return nil, domain.Validation("nomor resi wajib diisi", map[string]string{
			"tracking_number": "isi nomor resi dari JNE",
		})
	}
	if in.ShippingCost.IsNegative() {
		return nil, domain.Validation("ongkir tidak valid", map[string]string{
			"shipping_cost": "harus 0 atau lebih",
		})
	}

	shippedAt := time.Now()
	if in.ShippedAt != nil {
		shippedAt = *in.ShippedAt
	}

	var shipment *domain.Shipment
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if order.Status == domain.OrderCancelled {
			return domain.InvalidState("order sudah dibatalkan")
		}
		// Paket yang sudah diserahkan ke kurir tidak boleh dikirim ulang.
		//
		// Tanpa penjagaan ini nomor resinya tertimpa diam-diam sementara
		// balasannya tetap 200, jadi admin mengira tidak terjadi apa-apa —
		// padahal customer sudah terlanjur dikabari resi yang lama lewat
		// WhatsApp, dan sejak saat itu aplikasi dan customer melacak nomor yang
		// berbeda. Order yang sudah Selesai bahkan bisa terdorong balik ke
		// Dikirim. Untuk resi yang keliru, perbaikannya lewat data pengiriman.
		if order.Status == domain.OrderShipped || order.Status == domain.OrderCompleted {
			return domain.InvalidState(
				"order ini sudah diserahkan ke kurir — kalau nomor resinya keliru, perbaiki lewat data pengiriman, jangan kirim ulang")
		}
		if !in.AllowUnpaid && order.BalanceDue.GreaterThan(decimal.Zero) {
			return domain.InvalidState(
				"order belum lunas, sisa tagihan %s — catat pelunasan dulu atau kirim dengan penanda khusus",
				money.Format(order.BalanceDue))
		}

		existing, err := s.shipments.GetByOrder(ctx, tx, orderID)
		if err != nil {
			if domainErr, ok := domain.AsError(err); !ok || domainErr.Code != domain.CodeNotFound {
				return err
			}
			// Order yang langsung dikirim tanpa langkah packing terpisah tetap
			// perlu baris shipment, jadi dibuatkan di sini.
			existing, err = s.shipments.Pack(ctx, tx, repository.PackParams{
				OrderID:  orderID,
				Courier:  "JNE",
				Service:  "REG",
				PackedBy: nullableUUID(actorID),
			})
			if err != nil {
				return err
			}
		}

		shipment, err = s.shipments.Ship(ctx, tx, existing.ID, tracking, in.ShippingCost, shippedAt)
		if err != nil {
			return err
		}

		if domain.CanTransitionOrder(order.Status, domain.OrderShipped) {
			if _, err := s.orders.UpdateStatus(ctx, tx, orderID, domain.OrderShipped); err != nil {
				return err
			}
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "shipment",
			EntityID: &shipment.ID,
			Action:   "ship",
			Changes: map[string]any{
				"tracking_number": tracking,
				"shipping_cost":   in.ShippingCost.String(),
				"allow_unpaid":    in.AllowUnpaid,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return shipment, nil
}

// MarkDelivered menutup siklus order setelah customer mengonfirmasi barang tiba.
func (s *ShipmentService) MarkDelivered(ctx context.Context, orderID uuid.UUID, actorID uuid.UUID) (*domain.Shipment, error) {
	var shipment *domain.Shipment
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		order, err := s.orders.GetForUpdate(ctx, tx, orderID)
		if err != nil {
			return err
		}

		existing, err := s.shipments.GetByOrder(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if existing.TrackingNumber == nil {
			return domain.InvalidState("paket belum punya nomor resi")
		}

		shipment, err = s.shipments.UpdateStatus(ctx, tx, existing.ID, domain.ShipmentDelivered)
		if err != nil {
			return err
		}

		if domain.CanTransitionOrder(order.Status, domain.OrderCompleted) {
			if _, err := s.orders.UpdateStatus(ctx, tx, orderID, domain.OrderCompleted); err != nil {
				return err
			}
		}

		return s.audit.Record(ctx, tx, repository.AuditParams{
			UserID:   nullableUUID(actorID),
			Entity:   "shipment",
			EntityID: &existing.ID,
			Action:   "delivered",
		})
	})
	if err != nil {
		return nil, err
	}
	return shipment, nil
}

type UpdateShipmentInput struct {
	Courier        string
	Service        string
	WeightGram     int
	ShippingCost   decimal.Decimal
	TrackingNumber *string
	Notes          *string
}

func (s *ShipmentService) Update(ctx context.Context, id uuid.UUID, in UpdateShipmentInput) (*domain.Shipment, error) {
	courier := strings.ToUpper(strings.TrimSpace(in.Courier))
	if courier == "" {
		courier = "JNE"
	}
	serviceName := strings.ToUpper(strings.TrimSpace(in.Service))
	if serviceName == "" {
		serviceName = "REG"
	}

	tracking := trimPtr(in.TrackingNumber)
	if tracking != nil {
		upper := strings.ToUpper(*tracking)
		tracking = &upper
	}

	return s.shipments.Update(ctx, s.pool, id, courier, serviceName, in.WeightGram, in.ShippingCost, tracking, trimPtr(in.Notes))
}

func (s *ShipmentService) List(ctx context.Context, p pagination.Params, status string, tripID *uuid.UUID) ([]domain.ShipmentListItem, int64, error) {
	return s.shipments.List(ctx, s.pool, p, status, tripID)
}

func (s *ShipmentService) GetByOrder(ctx context.Context, orderID uuid.UUID) (*domain.Shipment, error) {
	return s.shipments.GetByOrder(ctx, s.pool, orderID)
}

// Message menyiapkan pesan pemberitahuan pengiriman berisi nomor resi.
func (s *ShipmentService) Message(ctx context.Context, orderID uuid.UUID) (*notify.Message, error) {
	order, err := s.orders.GetByID(ctx, s.pool, orderID)
	if err != nil {
		return nil, err
	}
	shipment, err := s.shipments.GetByOrder(ctx, s.pool, orderID)
	if err != nil {
		return nil, err
	}
	if shipment.TrackingNumber == nil {
		return nil, domain.InvalidState("paket belum punya nomor resi, isi resi terlebih dahulu")
	}
	customer, err := s.customers.GetByID(ctx, s.pool, order.CustomerID)
	if err != nil {
		return nil, err
	}
	settings, err := s.settings.All(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	message := notify.BuildShipmentMessage(notify.ShipmentParams{
		Settings: settings,
		Customer: customer,
		Order:    order,
		Shipment: shipment,
	})
	return &message, nil
}

// MarkNotified mencatat bahwa customer sudah dikabari nomor resinya.
func (s *ShipmentService) MarkNotified(ctx context.Context, orderID uuid.UUID) (*domain.Shipment, error) {
	shipment, err := s.shipments.GetByOrder(ctx, s.pool, orderID)
	if err != nil {
		return nil, err
	}
	return s.shipments.MarkNotified(ctx, s.pool, shipment.ID)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
