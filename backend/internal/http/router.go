// Package http merangkai seluruh rute HTTP aplikasi.
package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ipoool/jastipin/backend/internal/config"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/http/handler"
	"github.com/ipoool/jastipin/backend/internal/http/middleware"
	"github.com/ipoool/jastipin/backend/internal/pkg/token"
)

// Handlers mengumpulkan seluruh handler yang dibutuhkan router.
type Handlers struct {
	Auth      *handler.AuthHandler
	Users     *handler.UserHandler
	Customers *handler.CustomerHandler
	Products  *handler.ProductHandler
	Trips     *handler.TripHandler
	Orders    *handler.OrderHandler
	Purchases *handler.PurchaseHandler
	Stock     *handler.StockHandler
	Invoices  *handler.InvoiceHandler
	Shipments *handler.ShipmentHandler
	Reports   *handler.ReportHandler
	Settings  *handler.SettingsHandler
	Shipping  *handler.ShippingHandler
	FX        *handler.FXHandler
	Uploads   *handler.UploadHandler
}

type RouterDeps struct {
	Config   *config.Config
	Logger   *slog.Logger
	Pool     *pgxpool.Pool
	Tokens   *token.Manager
	Handlers Handlers
}

// NewRouter menyusun seluruh rute beserta middleware-nya.
//
// Pembagian hak akses:
//   - owner   : manajemen pengguna, pengaturan toko, dan laporan profit
//   - admin   : seluruh operasional harian (trip, order, invoice, kirim, stok)
//   - tripper : trip dan katalog (baca), daftar belanja, serta input pembelian
func NewRouter(d RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	// chimw.RealIP sengaja tidak dipakai: middleware itu menimpa r.RemoteAddr
	// dengan nilai header yang gampang dipalsukan. IP klien dibaca secara
	// eksplisit di handler login, dan hanya untuk keperluan catatan sesi.
	r.Use(middleware.Recoverer(d.Logger))
	r.Use(middleware.RequestLogger(d.Logger))
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   d.Config.App.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", healthHandler)
	r.Get("/health/ready", readinessHandler(d.Pool))

	// Berkas unggahan dilayani langsung oleh backend supaya deployment tetap
	// satu container tanpa perlu object storage terpisah.
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir(d.Config.Storage.UploadDir)))
	r.Handle("/uploads/*", fileServer)

	authenticated := middleware.Authenticate(d.Tokens)
	staffOnly := middleware.RequireRole(domain.RoleOwner, domain.RoleAdmin)
	ownerOnly := middleware.RequireRole(domain.RoleOwner)

	// Hak akses per menu memperhalus batas role: owner boleh mempersempit menu
	// mana saja yang dibuka untuk seorang pengguna. Dipasang berdampingan
	// dengan penjaga role, bukan menggantikannya — role tetap batas kasarnya.
	canAccess := middleware.RequirePermission

	r.Route("/api/v1", func(api chi.Router) {
		// --- Publik ---------------------------------------------------------
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/login", d.Handlers.Auth.Login)
			auth.Post("/refresh", d.Handlers.Auth.Refresh)
			auth.Post("/logout", d.Handlers.Auth.Logout)

			auth.Group(func(private chi.Router) {
				private.Use(authenticated)
				private.Get("/me", d.Handlers.Auth.Me)
				private.Post("/change-password", d.Handlers.Auth.ChangePassword)
			})
		})

		// --- Butuh login ----------------------------------------------------
		api.Group(func(private chi.Router) {
			private.Use(authenticated)

			private.Post("/uploads", d.Handlers.Uploads.Upload)

			// Manajemen pengguna khusus owner.
			private.Route("/users", func(users chi.Router) {
				users.Use(ownerOnly, canAccess(domain.PermUsers))
				users.Post("/", d.Handlers.Users.Create)
				users.Get("/", d.Handlers.Users.List)
				users.Get("/{id}", d.Handlers.Users.Get)
				users.Put("/{id}", d.Handlers.Users.Update)
				users.Post("/{id}/reset-password", d.Handlers.Users.ResetPassword)
				users.Delete("/{id}", d.Handlers.Users.Delete)
			})

			private.Route("/customers", func(customers chi.Router) {
				customers.Use(staffOnly, canAccess(domain.PermCustomers))
				customers.Post("/", d.Handlers.Customers.Create)
				customers.Get("/", d.Handlers.Customers.List)
				customers.Get("/{id}", d.Handlers.Customers.Get)
				customers.Get("/{id}/stats", d.Handlers.Customers.Stats)
				customers.Put("/{id}", d.Handlers.Customers.Update)
				customers.Delete("/{id}", d.Handlers.Customers.Delete)
			})

			private.Route("/product-categories", func(categories chi.Router) {
				// Tripper boleh membaca kategori karena dipakai memfilter
				// daftar belanja di lapangan.
				categories.Get("/", d.Handlers.Products.ListCategories)

				categories.Group(func(write chi.Router) {
					write.Use(staffOnly)
					write.Post("/", d.Handlers.Products.CreateCategory)
					write.Put("/{id}", d.Handlers.Products.UpdateCategory)
					write.Delete("/{id}", d.Handlers.Products.DeleteCategory)
				})
			})

			private.Route("/products", func(products chi.Router) {
				products.Use(canAccess(domain.PermProducts))
				products.Get("/", d.Handlers.Products.List)
				products.Get("/{id}", d.Handlers.Products.Get)
				products.Get("/{id}/price-history", d.Handlers.Products.PriceHistory)

				products.Group(func(write chi.Router) {
					write.Use(staffOnly)
					write.Post("/", d.Handlers.Products.Create)
					write.Post("/price-preview", d.Handlers.Products.PreviewPrice)
					write.Put("/{id}", d.Handlers.Products.Update)
					write.Delete("/{id}", d.Handlers.Products.Delete)
				})
			})

			private.Route("/trips", func(trips chi.Router) {
				trips.With(canAccess(domain.PermTrips)).Get("/", d.Handlers.Trips.List)
				trips.With(canAccess(domain.PermTrips)).Get("/{id}", d.Handlers.Trips.Get)
				trips.With(canAccess(domain.PermTrips)).Get("/{id}/items", d.Handlers.Trips.ListItems)
				trips.With(canAccess(domain.PermTrips)).Get("/{id}/expenses", d.Handlers.Trips.ListExpenses)

				// Daftar belanja dan input pembelian adalah pekerjaan tripper,
				// jadi sengaja tidak dibatasi staffOnly.
				trips.With(canAccess(domain.PermShoppingList)).
					Get("/{id}/shopping-list", d.Handlers.Purchases.ShoppingList)
				trips.With(canAccess(domain.PermPurchases)).
					Post("/{id}/purchases", d.Handlers.Purchases.Record)

				trips.Group(func(write chi.Router) {
					write.Use(staffOnly, canAccess(domain.PermTrips))
					write.Post("/", d.Handlers.Trips.Create)
					write.Put("/{id}", d.Handlers.Trips.Update)
					write.Patch("/{id}/status", d.Handlers.Trips.ChangeStatus)
					write.Delete("/{id}", d.Handlers.Trips.Delete)

					write.Post("/{id}/items", d.Handlers.Trips.AddItem)
					write.Put("/{id}/items/{itemId}", d.Handlers.Trips.UpdateItem)
					write.Delete("/{id}/items/{itemId}", d.Handlers.Trips.DeleteItem)
					write.Post("/{id}/recalculate-prices", d.Handlers.Trips.RecalculatePrices)
					write.Post("/{id}/sync-exchange-rate", d.Handlers.Trips.SyncExchangeRate)

					write.Post("/{id}/expenses", d.Handlers.Trips.AddExpense)
					write.Put("/{id}/expenses/{expenseId}", d.Handlers.Trips.UpdateExpense)
					write.Delete("/{id}/expenses/{expenseId}", d.Handlers.Trips.DeleteExpense)
				})
			})

			private.Route("/orders", func(orders chi.Router) {
				orders.Use(staffOnly, canAccess(domain.PermOrders))
				orders.Post("/", d.Handlers.Orders.Create)
				orders.Get("/", d.Handlers.Orders.List)
				orders.Get("/{id}", d.Handlers.Orders.Get)
				orders.Put("/{id}", d.Handlers.Orders.Update)
				orders.Patch("/{id}/status", d.Handlers.Orders.ChangeStatus)
				orders.Post("/{id}/cancel", d.Handlers.Orders.Cancel)

				orders.Post("/{id}/items", d.Handlers.Orders.AddItem)
				orders.Put("/{id}/items/{itemId}", d.Handlers.Orders.UpdateItem)
				orders.Delete("/{id}/items/{itemId}", d.Handlers.Orders.DeleteItem)

				orders.Post("/{id}/payments", d.Handlers.Orders.RecordPayment)
				orders.Delete("/{id}/payments/{paymentId}", d.Handlers.Orders.DeletePayment)

				orders.Post("/{id}/receive", d.Handlers.Orders.Receive)

				orders.Get("/{id}/label", d.Handlers.Shipments.Label)
				orders.Get("/{id}/invoices", d.Handlers.Invoices.ListByOrder)
				orders.Post("/{id}/invoices", d.Handlers.Invoices.Create)
				orders.Get("/{id}/dp-message", d.Handlers.Invoices.DPMessage)

				orders.Get("/{id}/shipment", d.Handlers.Shipments.GetByOrder)
				orders.Post("/{id}/pack", d.Handlers.Shipments.Pack)
				orders.Post("/{id}/ship", d.Handlers.Shipments.Ship)
				orders.Post("/{id}/delivered", d.Handlers.Shipments.MarkDelivered)
				orders.Get("/{id}/shipment-message", d.Handlers.Shipments.Message)
				orders.Post("/{id}/shipment-notified", d.Handlers.Shipments.MarkNotified)
				orders.Post("/{id}/shipping-estimate", d.Handlers.Shipping.EstimateForOrder)
				orders.Post("/{id}/shipping-options", d.Handlers.Shipping.OptionsForOrder)
			})

			private.Route("/purchases", func(purchases chi.Router) {
				purchases.Use(canAccess(domain.PermPurchases))
				purchases.Get("/", d.Handlers.Purchases.List)
				purchases.Get("/{id}", d.Handlers.Purchases.Get)
				purchases.Get("/{id}/allocations", d.Handlers.Purchases.ListAllocations)

				purchases.Group(func(write chi.Router) {
					write.Use(staffOnly)
					write.Delete("/{id}", d.Handlers.Purchases.Delete)
				})
			})

			private.Route("/stock", func(stock chi.Router) {
				stock.Use(staffOnly, canAccess(domain.PermStock))
				stock.Get("/", d.Handlers.Stock.List)
				stock.Get("/movements", d.Handlers.Stock.ListMovements)
				stock.Post("/sell", d.Handlers.Stock.Sell)
				stock.Post("/adjust", d.Handlers.Stock.Adjust)
			})

			private.Route("/invoices", func(invoices chi.Router) {
				invoices.Use(staffOnly, canAccess(domain.PermInvoices))
				invoices.Get("/", d.Handlers.Invoices.List)
				invoices.Get("/candidates", d.Handlers.Invoices.Candidates)
				invoices.Get("/{id}", d.Handlers.Invoices.Get)
				invoices.Get("/{id}/pdf", d.Handlers.Invoices.PDF)
				invoices.Get("/{id}/message", d.Handlers.Invoices.Message)
				invoices.Post("/{id}/mark-sent", d.Handlers.Invoices.MarkSent)
				invoices.Post("/{id}/void", d.Handlers.Invoices.Void)
			})

			private.Route("/shipments", func(shipments chi.Router) {
				shipments.Use(staffOnly, canAccess(domain.PermShipments))
				shipments.Get("/", d.Handlers.Shipments.List)
				shipments.Put("/{id}", d.Handlers.Shipments.Update)
			})

			// Estimasi ongkir dan tabel tarifnya. Estimasi boleh dipakai
			// seluruh staf; mengubah tarif hanya owner, karena angka ini
			// memengaruhi harga yang ditagihkan ke customer.
			// Kurs terkini dipakai mengisi kolom kurs saat trip dibuat. Boleh
			// diakses seluruh staf karena hanya membaca angka publik.
			private.With(staffOnly).Get("/exchange-rate", d.Handlers.FX.Rate)

			private.Route("/shipping", func(shipping chi.Router) {
				shipping.Use(staffOnly)
				shipping.Post("/estimate", d.Handlers.Shipping.Estimate)
				shipping.Get("/rates", d.Handlers.Shipping.ListRates)
				shipping.Get("/provider", d.Handlers.Shipping.Provider)

				shipping.Group(func(write chi.Router) {
					write.Use(ownerOnly, canAccess(domain.PermSettings))
					write.Post("/rates", d.Handlers.Shipping.SaveRate)
					write.Delete("/rates/{id}", d.Handlers.Shipping.DeleteRate)
					// Pencarian tujuan memakan kuota langganan, jadi hanya
					// dibuka untuk yang memang sedang menyetel pengiriman.
					write.Get("/destinations", d.Handlers.Shipping.SearchDestinations)
				})
			})

			private.Route("/reports", func(reports chi.Router) {
				reports.Use(staffOnly, canAccess(domain.PermReports))
				reports.Get("/dashboard", d.Handlers.Reports.Dashboard)
				reports.Get("/receivables", d.Handlers.Reports.Receivables)
				reports.Get("/products", d.Handlers.Reports.ProductSales)
				reports.Get("/customers", d.Handlers.Reports.CustomerSales)
				reports.Get("/channels", d.Handlers.Reports.ChannelSales)

				// Laporan laba-rugi hanya untuk owner.
				reports.Group(func(financial chi.Router) {
					financial.Use(ownerOnly)
					financial.Get("/trips/{id}/profit", d.Handlers.Reports.TripProfit)
					financial.Get("/orders", d.Handlers.Reports.OrderProfits)
				})
			})

			private.Route("/settings", func(settings chi.Router) {
				settings.Get("/", d.Handlers.Settings.List)

				settings.Group(func(write chi.Router) {
					write.Use(ownerOnly, canAccess(domain.PermSettings))
					write.Put("/", d.Handlers.Settings.Update)
				})
			})

			private.With(ownerOnly, canAccess(domain.PermSettings)).
				Get("/audit-logs", d.Handlers.Settings.AuditLogs)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": map[string]string{"code": "NOT_FOUND", "message": "endpoint tidak ditemukan"},
		})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": map[string]string{"code": "METHOD_NOT_ALLOWED", "message": "metode HTTP tidak didukung"},
		})
	})

	return r
}

// healthHandler menjawab tanpa menyentuh database: dipakai container untuk tahu
// apakah proses masih hidup.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readinessHandler ikut mengecek database, dipakai untuk memutuskan apakah
// instance siap menerima trafik.
func readinessHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := contextWithTimeout(r, 3*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "unavailable",
				"reason": "database tidak bisa dihubungi",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func contextWithTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}
