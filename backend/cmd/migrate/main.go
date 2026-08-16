// Command migrate menjalankan migrasi skema database.
//
//	migrate up          -> terapkan semua migrasi yang tertunda
//	migrate down [n]    -> batalkan n migrasi terakhir (default 1)
//	migrate reset       -> batalkan semua migrasi
//	migrate version     -> tampilkan versi skema saat ini
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/ipoool/jastipin/backend/internal/config"
	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// .env hanya untuk kenyamanan development; di container env var diinjeksi
	// langsung sehingga file ini tidak ada dan errornya sengaja diabaikan.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.App.Env, cfg.App.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer pool.Close()

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		return db.MigrateUp(ctx, pool, log)

	case "down":
		steps := 1
		if len(os.Args) > 2 {
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil || steps < 1 {
				return fmt.Errorf("jumlah langkah harus angka positif, dapat %q", os.Args[2])
			}
		}
		return db.MigrateDown(ctx, pool, log, steps)

	case "reset":
		return db.MigrateDown(ctx, pool, log, 0)

	case "version":
		version, err := db.CurrentVersion(ctx, pool)
		if err != nil {
			return err
		}
		fmt.Printf("versi skema saat ini: %d\n", version)
		return nil

	default:
		return fmt.Errorf("perintah tidak dikenal %q (pilihan: up, down, reset, version)", command)
	}
}
