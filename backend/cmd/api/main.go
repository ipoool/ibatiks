// Command api menjalankan HTTP server Ibatiks.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/ipoool/jastipin/backend/internal/config"
	"github.com/ipoool/jastipin/backend/internal/db"
	apihttp "github.com/ipoool/jastipin/backend/internal/http"
	"github.com/ipoool/jastipin/backend/internal/http/handler"
	"github.com/ipoool/jastipin/backend/internal/pdf"
	"github.com/ipoool/jastipin/backend/internal/pkg/logger"
	"github.com/ipoool/jastipin/backend/internal/pkg/rajaongkir"
	"github.com/ipoool/jastipin/backend/internal/pkg/token"
	"github.com/ipoool/jastipin/backend/internal/repository"
	"github.com/ipoool/jastipin/backend/internal/service"
)

func main() {
	// Image production memakai distroless yang tidak punya shell maupun curl,
	// jadi binary ini sendiri yang menyediakan perintah health check untuk
	// dipakai HEALTHCHECK pada Dockerfile.
	if len(os.Args) > 1 && os.Args[1] == "--health" {
		if err := healthCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "health: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "api: %v\n", err)
		os.Exit(1)
	}
}

func healthCheck() error {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%s/health", port))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func run() error {
	// .env hanya untuk development; di container env var diinjeksi langsung.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.App.Env, cfg.App.LogLevel)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer pool.Close()
	log.Info("terhubung ke database")

	// Migrasi dijalankan saat boot supaya deployment cukup satu langkah dan
	// skema tidak pernah tertinggal dari kode yang sedang berjalan.
	if err := db.MigrateUp(ctx, pool, log); err != nil {
		return fmt.Errorf("jalankan migrasi: %w", err)
	}

	if err := ensureDirs(cfg.Storage.UploadDir, cfg.Storage.InvoiceDir); err != nil {
		return err
	}

	router := apihttp.NewRouter(apihttp.RouterDeps{
		Config:   cfg,
		Logger:   log,
		Pool:     pool,
		Tokens:   token.NewManager(cfg.JWT),
		Handlers: buildHandlers(cfg, pool),
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: router,
		// Timeout eksplisit mencegah koneksi menggantung menghabiskan slot
		// server. ReadHeaderTimeout juga menutup celah serangan Slowloris.
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go startTokenCleanup(ctx, pool, log)

	serverErr := make(chan error, 1)
	go func() {
		log.Info("server berjalan", "port", cfg.App.Port, "env", cfg.App.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server berhenti: %w", err)

	case <-ctx.Done():
		log.Info("sinyal shutdown diterima, menyelesaikan request yang sedang berjalan")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown tidak selesai tepat waktu: %w", err)
		}
		log.Info("server berhenti dengan rapi")
		return nil
	}
}

// buildHandlers merangkai repository, service, dan handler. Semua dependensi
// disuntikkan dari satu tempat ini supaya alurnya gampang ditelusuri.
func buildHandlers(cfg *config.Config, pool *pgxpool.Pool) apihttp.Handlers {
	userRepo := repository.NewUserRepo()
	customerRepo := repository.NewCustomerRepo()
	productRepo := repository.NewProductRepo()
	tripRepo := repository.NewTripRepo()
	orderRepo := repository.NewOrderRepo()
	purchaseRepo := repository.NewPurchaseRepo()
	invoiceRepo := repository.NewInvoiceRepo()
	shipmentRepo := repository.NewShipmentRepo()
	shippingRepo := repository.NewShippingRepo()
	reportRepo := repository.NewReportRepo()
	settingsRepo := repository.NewSettingsRepo()
	auditRepo := repository.NewAuditRepo()

	tokens := token.NewManager(cfg.JWT)
	renderer := pdf.NewRenderer(cfg.Storage.InvoiceDir)

	authService := service.NewAuthService(pool, userRepo, tokens)
	userService := service.NewUserService(pool, userRepo)
	customerService := service.NewCustomerService(pool, customerRepo)
	productService := service.NewProductService(pool, productRepo)
	fxService := service.NewFXService()
	tripService := service.NewTripService(pool, tripRepo, productRepo, orderRepo, auditRepo, fxService)
	orderService := service.NewOrderService(pool, orderRepo, tripRepo, customerRepo, productRepo,
		purchaseRepo, invoiceRepo, auditRepo, cfg.Business)
	purchaseService := service.NewPurchaseService(pool, purchaseRepo, tripRepo, orderRepo, productRepo, auditRepo)
	invoiceService := service.NewInvoiceService(pool, invoiceRepo, orderRepo, customerRepo,
		tripRepo, settingsRepo, auditRepo, renderer)
	shippingService := service.NewShippingService(pool, shippingRepo, orderRepo, settingsRepo)

	// Tarif diambil dari RajaOngkir kalau API key-nya diisi. Tanpa key,
	// aplikasinya tetap jalan memakai tabel tarif yang dikelola sendiri —
	// toko yang belum berlangganan tidak boleh terhalang.
	rajaOngkir := rajaongkir.New(cfg.RajaOngkir.APIKey, cfg.RajaOngkir.BaseURL, cfg.RajaOngkir.Timeout)
	if rajaOngkir.Enabled() {
		shippingService.UseCostProvider(
			service.NewRajaOngkirProvider(pool, rajaOngkir, shippingRepo, settingsRepo))
		slog.Info("tarif kirim memakai RajaOngkir")
	} else {
		slog.Info("RAJAONGKIR_API_KEY kosong, tarif kirim memakai tabel tarif sendiri")
	}
	shipmentService := service.NewShipmentService(pool, shipmentRepo, orderRepo, customerRepo,
		settingsRepo, shippingService, auditRepo, tripRepo, renderer)
	reportService := service.NewReportService(pool, reportRepo, tripRepo, orderRepo)
	settingsService := service.NewSettingsService(pool, settingsRepo, auditRepo)
	auditService := service.NewAuditService(pool, auditRepo)

	return apihttp.Handlers{
		Auth:      handler.NewAuthHandler(authService),
		Users:     handler.NewUserHandler(userService),
		Customers: handler.NewCustomerHandler(customerService),
		Products:  handler.NewProductHandler(productService),
		Trips:     handler.NewTripHandler(tripService),
		Orders:    handler.NewOrderHandler(orderService),
		Purchases: handler.NewPurchaseHandler(purchaseService),
		Stock:     handler.NewStockHandler(purchaseService),
		Invoices:  handler.NewInvoiceHandler(invoiceService),
		Shipments: handler.NewShipmentHandler(shipmentService),
		Reports:   handler.NewReportHandler(reportService),
		Settings:  handler.NewSettingsHandler(settingsService, auditService),
		Shipping:  handler.NewShippingHandler(shippingService),
		FX:        handler.NewFXHandler(fxService),
		Uploads:   handler.NewUploadHandler(cfg.Storage.UploadDir, cfg.App.BaseURL),
	}
}

// startTokenCleanup membuang sesi kedaluwarsa secara berkala supaya tabel
// refresh_tokens tidak tumbuh tanpa batas.
func startTokenCleanup(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) {
	repo := repository.NewUserRepo()
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleted, err := repo.PurgeExpiredRefreshTokens(ctx, pool)
			if err != nil {
				log.Warn("gagal membersihkan sesi kedaluwarsa", "error", err)
				continue
			}
			if deleted > 0 {
				log.Info("sesi kedaluwarsa dibersihkan", "jumlah", deleted)
			}
		}
	}
}

func ensureDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("buat direktori %s: %w", dir, err)
		}
	}
	return nil
}
