// Command seed mengisi database dengan data awal.
//
//	seed          -> hanya membuat akun owner pertama kalau belum ada
//	seed --demo   -> tambah data contoh (customer, produk, trip, katalog)
//
// Perintah ini aman dijalankan berulang: data yang sudah ada tidak digandakan.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/config"
	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/docnum"
	"github.com/ipoool/jastipin/backend/internal/pkg/logger"
	"github.com/ipoool/jastipin/backend/internal/repository"
	"github.com/ipoool/jastipin/backend/internal/service"
)

func main() {
	demo := flag.Bool("demo", false, "ikut membuat data contoh untuk mencoba aplikasi")
	flag.Parse()

	if err := run(*demo); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
}

func run(withDemo bool) error {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.App.Env, cfg.App.LogLevel)
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer pool.Close()

	owner, err := seedOwner(ctx, pool, log)
	if err != nil {
		return err
	}

	if err := seedRoot(ctx, pool, log); err != nil {
		return err
	}

	if withDemo {
		if err := seedDemo(ctx, pool, log, owner.ID); err != nil {
			return err
		}
	}

	log.Info("seed selesai")
	return nil
}

// seedRoot membuat akun root kalau SEED_ROOT_EMAIL diisi.
//
// Role root memegang seluruh menu tanpa kecuali dan daftarnya tidak bisa
// dipersempit. Gunanya bukan pekerjaan sehari-hari, melainkan jalan pulih:
// begitu hak akses owner terlanjur salah disetel, menu Pengaturan dan Pengguna
// bisa saja tidak bisa dibuka siapa pun, dan pemulihannya cuma lewat database.
//
// Dibuat di seed, bukan di migrasi, supaya passwordnya tidak ikut tersimpan di
// riwayat repositori — dan supaya akunnya lahir lagi setiap kali database
// direset.
func seedRoot(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	email := strings.ToLower(strings.TrimSpace(envOr("SEED_ROOT_EMAIL", "")))
	if email == "" {
		return nil
	}

	users := repository.NewUserRepo()
	if _, err := users.GetByEmail(ctx, pool, email); err == nil {
		log.Info("akun root sudah ada, dilewati", "email", email)
		return nil
	} else if !isNotFound(err) {
		return err
	}

	password := envOr("SEED_ROOT_PASSWORD", "")
	if password == "" {
		return errors.New("SEED_ROOT_PASSWORD wajib diisi kalau SEED_ROOT_EMAIL diisi")
	}

	hashed, err := service.HashPassword(password)
	if err != nil {
		return err
	}

	if _, err := users.Create(ctx, pool, repository.CreateUserParams{
		Name:         envOr("SEED_ROOT_NAME", "Root"),
		Email:        email,
		PasswordHash: hashed,
		Role:         domain.RoleRoot,
	}); err != nil {
		return err
	}

	log.Info("akun root dibuat", "email", email)
	return nil
}

// seedOwner membuat akun owner pertama dari environment. Kalau sudah ada
// pengguna dengan email tersebut, akunnya dibiarkan apa adanya.
func seedOwner(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) (*domain.User, error) {
	users := repository.NewUserRepo()

	email := strings.ToLower(strings.TrimSpace(envOr("SEED_OWNER_EMAIL", "owner@ibatiks.id")))
	name := envOr("SEED_OWNER_NAME", "Owner")
	password := envOr("SEED_OWNER_PASSWORD", "")

	if existing, err := users.GetByEmail(ctx, pool, email); err == nil {
		log.Info("akun owner sudah ada, dilewati", "email", email)
		return existing, nil
	} else if !isNotFound(err) {
		return nil, err
	}

	if password == "" {
		return nil, errors.New("SEED_OWNER_PASSWORD wajib diisi untuk membuat akun owner pertama")
	}

	hashed, err := service.HashPassword(password)
	if err != nil {
		return nil, err
	}

	owner, err := users.Create(ctx, pool, repository.CreateUserParams{
		Name:         name,
		Email:        email,
		PasswordHash: hashed,
		Role:         domain.RoleOwner,
	})
	if err != nil {
		return nil, err
	}

	log.Info("akun owner dibuat", "email", email)
	return owner, nil
}

// seedDemo membuat satu skenario lengkap: kategori, produk, customer, dan trip
// beserta katalognya, supaya aplikasi bisa langsung dicoba tanpa input manual.
func seedDemo(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger, ownerID uuid.UUID) error {
	products := repository.NewProductRepo()
	customers := repository.NewCustomerRepo()
	trips := repository.NewTripRepo()

	// Kategori dibuat idempoten: kalau slug-nya sudah ada, dipakai yang lama.
	categories := map[string]uuid.UUID{}
	existing, err := products.ListCategories(ctx, pool)
	if err != nil {
		return err
	}
	for _, c := range existing {
		categories[c.Slug] = c.ID
	}

	for _, name := range []string{"Skincare", "Snack", "Fashion"} {
		slug := strings.ToLower(name)
		if _, ok := categories[slug]; ok {
			continue
		}
		category, err := products.CreateCategory(ctx, pool, name, slug, nil)
		if err != nil {
			return err
		}
		categories[slug] = category.ID
	}

	demoProducts := []struct {
		SKU        string
		Name       string
		Category   string
		Brand      string
		Currency   string
		BasePrice  string
		MarkupType string
		Markup     string
		WeightGram int
	}{
		{"SKN-0001", "Hada Labo Gokujyun Lotion 170ml", "skincare", "Hada Labo", "JPY", "880", domain.MarkupPercent, "35", 220},
		{"SKN-0002", "Melano CC Vitamin C Serum 20ml", "skincare", "Rohto", "JPY", "1180", domain.MarkupPercent, "35", 80},
		{"SNK-0001", "Tokyo Banana 8pcs", "snack", "Tokyo Banana", "JPY", "1180", domain.MarkupPercent, "40", 350},
		{"SNK-0002", "Kit Kat Matcha 12pcs", "snack", "Nestle", "JPY", "780", domain.MarkupNominal, "40000", 180},
		{"FSH-0001", "Uniqlo Airism Tee", "fashion", "Uniqlo", "JPY", "1500", domain.MarkupPercent, "30", 200},
	}

	productIDs := map[string]uuid.UUID{}
	for _, p := range demoProducts {
		basePrice, _ := decimal.NewFromString(p.BasePrice)
		markup, _ := decimal.NewFromString(p.Markup)
		categoryID := categories[p.Category]
		brand := p.Brand

		created, err := products.Create(ctx, pool, repository.ProductParams{
			SKU:          p.SKU,
			Name:         p.Name,
			CategoryID:   &categoryID,
			Brand:        &brand,
			BaseCurrency: p.Currency,
			BasePrice:    basePrice,
			MarkupType:   p.MarkupType,
			MarkupValue:  markup,
			WeightGram:   p.WeightGram,
			IsActive:     true,
		})
		if err != nil {
			if isConflict(err) {
				log.Info("produk contoh sudah ada, dilewati", "sku", p.SKU)
				continue
			}
			return err
		}
		productIDs[p.SKU] = created.ID
	}

	// Kalau produk contoh sudah pernah dibuat, sisa seed tidak perlu diulang.
	if len(productIDs) == 0 {
		log.Info("data contoh sudah ada, dilewati")
		return nil
	}

	demoCustomers := []struct {
		Name    string
		Phone   string
		City    string
		Address string
	}{
		{"Rina Kartika", "081234567890", "Jakarta Selatan", "Jl. Kemang Raya No. 12, RT 03/RW 05"},
		{"Budi Santoso", "081298765432", "Bandung", "Jl. Dago Atas No. 88"},
		{"Sari Dewi", "085712345678", "Surabaya", "Jl. Darmo Permai III No. 21"},
	}

	for _, c := range demoCustomers {
		city, address := c.City, c.Address
		err := db.WithTx(ctx, pool, func(tx pgx.Tx) error {
			code, err := docnum.Next(ctx, tx, docnum.Customer, time.Now().Year())
			if err != nil {
				return err
			}
			_, err = customers.Create(ctx, tx, repository.CustomerParams{
				Code:     code,
				Name:     c.Name,
				PhoneWA:  domain.NormalizePhoneWA(c.Phone),
				City:     &city,
				Address:  &address,
				Province: nil,
			})
			return err
		})
		if err != nil && !isConflict(err) {
			return err
		}
	}

	// Trip contoh: berangkat minggu depan, kurs JPY yang wajar.
	departDate := time.Now().AddDate(0, 0, 7)
	returnDate := departDate.AddDate(0, 0, 7)
	deadline := departDate.AddDate(0, 0, -2)
	rate := decimal.NewFromFloat(108.5)

	var tripID uuid.UUID
	err = db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		code, err := docnum.Next(ctx, tx, docnum.Trip, departDate.Year())
		if err != nil {
			return err
		}

		city := "Tokyo"
		trip, err := trips.Create(ctx, tx, repository.TripParams{
			Code:          code,
			Title:         "Jastip Tokyo " + departDate.Format("Jan 2006"),
			Country:       "Jepang",
			City:          &city,
			DepartDate:    departDate,
			ReturnDate:    returnDate,
			OrderDeadline: &deadline,
			Currency:      "JPY",
			ExchangeRate:  rate,
			CreatedBy:     &ownerID,
		})
		if err != nil {
			return err
		}
		tripID = trip.ID

		// Katalog trip memakai harga modal dan markup dari master produk.
		for _, p := range demoProducts {
			productID, ok := productIDs[p.SKU]
			if !ok {
				continue
			}
			costPrice, _ := decimal.NewFromString(p.BasePrice)
			markup, _ := decimal.NewFromString(p.Markup)

			costIDR, sellPrice := domain.CalculateSellPrice(costPrice, rate, p.MarkupType, markup)
			if _, err := trips.CreateItem(ctx, tx, repository.TripItemParams{
				TripID:       tripID,
				ProductID:    productID,
				CostPrice:    costPrice,
				CostPriceIDR: costIDR,
				MarkupType:   p.MarkupType,
				MarkupValue:  markup,
				SellPrice:    sellPrice,
				IsActive:     true,
			}); err != nil {
				return err
			}
		}

		// Trip langsung dibuka supaya order contoh bisa langsung dicatat.
		_, err = trips.UpdateStatus(ctx, tx, tripID, domain.TripOpen)
		return err
	})
	if err != nil {
		return err
	}

	log.Info("data contoh dibuat",
		"produk", len(productIDs), "customer", len(demoCustomers), "trip", tripID)
	return nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func isNotFound(err error) bool {
	domainErr, ok := domain.AsError(err)
	return ok && domainErr.Code == domain.CodeNotFound
}

func isConflict(err error) bool {
	domainErr, ok := domain.AsError(err)
	return ok && domainErr.Code == domain.CodeConflict
}
