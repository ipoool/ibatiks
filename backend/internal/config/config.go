// Package config memuat seluruh konfigurasi aplikasi dari environment variable.
// Konfigurasi divalidasi sekali saat boot supaya aplikasi gagal cepat (fail fast)
// ketimbang meledak di tengah request karena env yang salah.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      App
	DB       DB
	JWT      JWT
	Storage  Storage
	Business Business
}

type App struct {
	Env             string        // development | production
	Port            int           // port HTTP server
	BaseURL         string        // URL publik backend, dipakai untuk link file
	LogLevel        string        // debug | info | warn | error
	AllowedOrigins  []string      // CORS
	ShutdownTimeout time.Duration // grace period saat shutdown
}

func (a App) IsProduction() bool { return a.Env == "production" }

type DB struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type JWT struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

type Storage struct {
	UploadDir  string // bukti transfer, foto struk
	InvoiceDir string // hasil render PDF invoice
}

// Business menampung default operasional yang sewaktu-waktu bisa berbeda per
// pemilik toko. Nilai yang lebih spesifik disimpan di tabel app_settings.
type Business struct {
	DefaultDPPercent int // persentase DP default saat order dibuat
}

// Load membaca environment dan mengembalikan konfigurasi yang sudah tervalidasi.
func Load() (*Config, error) {
	cfg := &Config{
		App: App{
			Env:             envString("APP_ENV", "development"),
			Port:            envInt("APP_PORT", 8080),
			BaseURL:         envString("APP_BASE_URL", "http://localhost:8080"),
			LogLevel:        envString("LOG_LEVEL", "info"),
			AllowedOrigins:  envStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			ShutdownTimeout: envDuration("APP_SHUTDOWN_TIMEOUT", 15*time.Second),
		},
		DB: DB{
			URL:             envString("DATABASE_URL", ""),
			MaxConns:        int32(envInt("DB_MAX_CONNS", 20)),
			MinConns:        int32(envInt("DB_MIN_CONNS", 2)),
			MaxConnLifetime: envDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: envDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
		},
		JWT: JWT{
			Secret:     envString("JWT_SECRET", ""),
			AccessTTL:  envDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: envDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
			Issuer:     envString("JWT_ISSUER", "jastipin"),
		},
		Storage: Storage{
			UploadDir:  envString("UPLOAD_DIR", "./data/uploads"),
			InvoiceDir: envString("INVOICE_DIR", "./data/invoices"),
		},
		Business: Business{
			DefaultDPPercent: envInt("DEFAULT_DP_PERCENT", 50),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	var problems []string

	if c.DB.URL == "" {
		problems = append(problems, "DATABASE_URL wajib diisi")
	}
	if c.JWT.Secret == "" {
		problems = append(problems, "JWT_SECRET wajib diisi")
	} else if c.App.IsProduction() && len(c.JWT.Secret) < 32 {
		// Di production secret pendek terlalu mudah ditebak; di development
		// kita biarkan supaya tidak menghambat percobaan lokal.
		problems = append(problems, "JWT_SECRET minimal 32 karakter di environment production")
	}
	if c.App.Port < 1 || c.App.Port > 65535 {
		problems = append(problems, "APP_PORT harus antara 1 dan 65535")
	}
	if c.Business.DefaultDPPercent < 0 || c.Business.DefaultDPPercent > 100 {
		problems = append(problems, "DEFAULT_DP_PERCENT harus antara 0 dan 100")
	}
	if c.DB.MinConns > c.DB.MaxConns {
		problems = append(problems, "DB_MIN_CONNS tidak boleh lebih besar dari DB_MAX_CONNS")
	}

	if len(problems) > 0 {
		return fmt.Errorf("konfigurasi tidak valid:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

func envString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func envStringSlice(key string, fallback []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
