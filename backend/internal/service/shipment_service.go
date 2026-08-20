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
// Label membentuk label pengiriman siap tempel untuk sebuah order.
//
// Paket belum tentu sudah dibuat: label boleh dicetak lebih dulu sebagai
// penanda kardus saat mengemas, dan nomor resinya menyusul — bagiannya
// disediakan kosong untuk ditulis tangan di konter kurir.
func (s *ShipmentService) Label(ctx context.Context, orderID uuid.UUID) ([]byte, string, error) {
	order, err := s.orders.GetByID(ctx, s.pool, orderID)
	if err != nil {
		return nil, "", err
	}
	settings, err := s.settings.All(ctx, s.pool)
	if err != nil {
		return nil, "", err
	}

	shipment, err := s.shipments.GetByOrder(ctx, s.pool, orderID)
	if err != nil {
		if domainErr, ok := domain.AsError(err); !ok || domainErr.Code != domain.CodeNotFound {
			return nil, "", err
		}
		shipment = nil
	}

	content, err := s.renderer.RenderLabel(pdf.LabelData{
		Order:    order,
		Shipment: shipment,
		Settings: settings,
	})
	if err != nil {
		return nil, "", domain.Internal(err)
	}
	return content, "LABEL-" + order.OrderNumber, nil
}

type PackInput struct {
	Courier    string
	Service    string
	WeightGram int
	// ShippingFee adalah ongkir yang ditagihkan ke customer, diisi saat admin
	// memilih layanan kurir. Nil berarti admin baru menyimpan ukuran paketnya
	// dan belum menetapkan ongkir — nilai di order tidak disentuh.
	ShippingFee *decimal.Decimal
	// InsuranceFee adalah premi asuransi kiriman, diketik admin dari struk
	// kurir. Kurir tidak mengembalikannya lewat API — balasan RajaOngkir hanya
	// berisi nama kurir, layanan, ongkos, dan estimasi tiba — jadi tidak ada
	// angka yang bisa dihitung sistem tanpa menebak.
	InsuranceFee decimal.Decimal
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

		if in.ShippingFee != nil && in.ShippingFee.IsNegative() {
			return domain.Validation("ongkir tidak valid", map[string]string{
				"shipping_fee": "harus 0 atau lebih",
			})
		}
		if in.InsuranceFee.IsNegative() {
			return domain.Validation("premi asuransi tidak valid", map[string]string{
				"insurance_fee": "harus 0 atau lebih",
			})
		}

		// Estimasi ongkir disimpan bersama data kemasan supaya admin punya
		// pembanding saat memasukkan ongkir yang sebenarnya dibayar nanti.
		// Kegagalan menghitung tidak boleh membatalkan proses packing.
		estimated := decimal.Zero
		if estimate, estErr := s.shipping.Estimate(ctx, EstimateInput{
			Courier:     courier,
			Service:     serviceName,
			City:        order.ShippingCity,
			District:    derefString(order.ShippingDistrict),
			Subdistrict: derefString(order.ShippingSubdistrict),
			PostalCode:  derefString(order.ShippingPostalCode),
			WeightGram:  in.WeightGram,
			LengthCM:    in.LengthCM,
			WidthCM:     in.WidthCM,
			HeightCM:    in.HeightCM,
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
			InsuranceFee:  in.InsuranceFee,
			Notes:         trimPtr(in.Notes),
			PackedBy:      nullableUUID(actorID),
		})
		if err != nil {
			return err
		}

		// Mengemas tidak mengubah status order. Kemajuannya sudah terbaca dari
		// data kemasan ini sendiri — berat, dimensi, dan waktu dikemasnya —
		// jadi status tersendiri hanya akan jadi salinan yang bisa berbeda.

		// Ongkir baru diketahui di sini, setelah paketnya ditimbang dan
		// layanan kurirnya dipilih. Menuliskannya ke order membuat totalnya
		// naik, dan itulah angka yang nanti ditagihkan invoice pelunasan.
		//
		// dp_required sengaja tidak ikut dihitung ulang: DP sudah disepakati
		// dan besar kemungkinan sudah dibayar, jadi menaikkannya belakangan
		// berarti customer tiba-tiba dianggap kurang bayar.
		// Premi asuransi menyatu dengan ongkir saat ditagihkan: invoice dan label
		// membaca satu angka dari orders.shipping_fee, dan memecahnya jadi dua
		// baris tagihan berarti mengubah bentuk dokumen yang sudah dipegang
		// customer. Rinciannya tetap tersimpan di baris kemasan.
		if in.ShippingFee != nil {
			ditagihkan := in.ShippingFee.Add(in.InsuranceFee)
			if !ditagihkan.Equal(order.ShippingFee) {
				if _, err := s.orders.SetShippingFee(ctx, tx, orderID, ditagihkan); err != nil {
					return err
				}
				diperbarui, err := s.orders.RecalculateTotals(ctx, tx, orderID)
				if err != nil {
					return err
				}
				// Customer yang sudah membayar lunas sebelum paketnya ditimbang
				// kini kembali punya sisa tagihan sebesar ongkirnya. Tanpa langkah
				// ini ordernya tetap berlabel Pembayaran Lunas, ikut masuk antrean
				// siap kirim, dan barangnya berangkat sementara ongkirnya tidak
				// pernah tertagih.
				if err := reconcileOrderStatus(ctx, tx, s.orders, diperbarui); err != nil {
					return err
				}
			}
		}

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

/*
 * Track menanyakan posisi paket ke kurir, dan menandai ordernya Selesai bila
 * kurir menyatakan paketnya sudah diterima.
 *
 * Perpindahan status dilakukan lewat MarkDelivered yang sudah ada, bukan jalur
 * sendiri: aturan transisinya, penanda waktu, dan catatan auditnya jadi sama
 * persis dengan yang dipakai saat admin menandainya manual.
 *
 * Kurir yang menjawab "resi tidak dikenal" bukan berarti admin salah ketik.
 * Resi yang baru diserahkan sering belum masuk sistem kurir sampai beberapa jam
 * kemudian, jadi galatnya diteruskan apa adanya dan tidak ada yang dibatalkan.
 */
func (s *ShipmentService) Track(ctx context.Context, orderID uuid.UUID, actorID uuid.UUID) (*domain.TrackingInfo, error) {
	shipment, err := s.shipments.GetByOrder(ctx, s.pool, orderID)
	if err != nil {
		return nil, err
	}
	if shipment.TrackingNumber == nil || strings.TrimSpace(*shipment.TrackingNumber) == "" {
		return nil, domain.InvalidState("paket belum punya nomor resi untuk dilacak")
	}

	info, err := s.shipping.TrackWaybill(ctx, *shipment.TrackingNumber, shipment.Courier)
	if err != nil {
		return nil, err
	}

	// Hanya menandai Selesai kalau kurir benar-benar menyatakannya, dan hanya
	// kalau ordernya memang masih dalam perjalanan.
	if info.Delivered && shipment.Status != domain.ShipmentDelivered {
		if _, err := s.MarkDelivered(ctx, orderID, actorID); err != nil {
			return nil, err
		}
		info.OrderCompleted = true
	}
	return info, nil
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

// Queue mendaftar pekerjaan pengiriman: order yang DP-nya sudah masuk, beserta
// data kemasannya bila sudah ada.
func (s *ShipmentService) Queue(
	ctx context.Context, p pagination.Params, stage string, tripID *uuid.UUID,
) ([]domain.ShippingQueueItem, int64, error) {
	if stage != "" && !domain.IsValidShippingStage(stage) {
		return nil, 0, domain.Validation("tahap pengiriman tidak dikenal", map[string]string{
			"stage": "pilih perlu_kemas, siap_kirim, atau terkirim",
		})
	}
	return s.orders.ListForShipping(ctx, s.pool, p, stage, tripID)
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
