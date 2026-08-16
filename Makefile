# =============================================================================
# Ibatiks — perintah pintas untuk pengembangan dan deployment.
# Jalankan `make` tanpa argumen untuk melihat daftar perintah.
# =============================================================================

COMPOSE      := docker compose
COMPOSE_PROD := docker compose -f docker-compose.prod.yml

.DEFAULT_GOAL := help
.PHONY: help setup dev up down restart logs logs-be logs-web ps shell-be shell-db \
        migrate-up migrate-down migrate-reset migrate-version seed seed-demo \
        test test-be lint build build-be build-web push smoke clean reset \
        prod-build prod-up prod-down prod-logs prod-migrate prod-seed backup

help: ## Tampilkan daftar perintah
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# --- Pengembangan ------------------------------------------------------------

setup: ## Siapkan .env dari contoh lalu jalankan semuanya
	@test -f .env || (cp .env.example .env && echo "✓ .env dibuat dari .env.example — sesuaikan isinya")
	@$(MAKE) up
	@echo ""
	@echo "  Web      : http://localhost:3000"
	@echo "  API      : http://localhost:8080"
	@echo "  Adminer  : http://localhost:8081"
	@echo ""
	@echo "  Jalankan 'make seed-demo' untuk mengisi data contoh."

dev up: ## Nyalakan seluruh service (dengan hot reload)
	$(COMPOSE) up -d --build

down: ## Matikan seluruh service
	$(COMPOSE) down

restart: ## Nyalakan ulang seluruh service
	$(COMPOSE) restart

ps: ## Lihat status service
	$(COMPOSE) ps

logs: ## Ikuti log semua service
	$(COMPOSE) logs -f

logs-be: ## Ikuti log backend saja
	$(COMPOSE) logs -f backend

logs-web: ## Ikuti log frontend saja
	$(COMPOSE) logs -f web

shell-be: ## Masuk ke shell container backend
	$(COMPOSE) exec backend sh

shell-db: ## Masuk ke psql
	$(COMPOSE) exec db psql -U $${POSTGRES_USER:-jastipin} -d $${POSTGRES_DB:-jastipin}

# --- Database ----------------------------------------------------------------

migrate-up: ## Terapkan migrasi yang tertunda
	$(COMPOSE) exec backend go run ./cmd/migrate up

migrate-down: ## Batalkan satu migrasi terakhir
	$(COMPOSE) exec backend go run ./cmd/migrate down 1

migrate-reset: ## Batalkan seluruh migrasi (menghapus semua tabel)
	$(COMPOSE) exec backend go run ./cmd/migrate reset

migrate-version: ## Tampilkan versi skema saat ini
	$(COMPOSE) exec backend go run ./cmd/migrate version

seed: ## Buat akun owner pertama
	$(COMPOSE) exec backend go run ./cmd/seed

seed-demo: ## Buat akun owner + data contoh untuk mencoba aplikasi
	$(COMPOSE) exec backend go run ./cmd/seed --demo

# --- Kualitas kode -----------------------------------------------------------

test test-be: ## Jalankan unit test backend
	cd backend && go test ./...

lint: ## Periksa format, vet, dan tipe
	cd backend && gofmt -l . && go vet ./...
	cd web && npx tsc --noEmit && npm run lint

build build-be: ## Build binary backend
	cd backend && go build ./...

build-web: ## Build frontend
	cd web && npm run build

smoke: ## Uji alur bisnis end-to-end lewat API
	./scripts/smoke.sh $${SMOKE_BASE_URL:-http://localhost:8080}

# --- Production --------------------------------------------------------------

prod-build: ## Build image production
	$(COMPOSE_PROD) build

prod-up: ## Nyalakan stack production
	$(COMPOSE_PROD) up -d

prod-down: ## Matikan stack production
	$(COMPOSE_PROD) down

prod-logs: ## Ikuti log production
	$(COMPOSE_PROD) logs -f

prod-migrate: ## Jalankan migrasi di production
	$(COMPOSE_PROD) run --rm migrate up

prod-seed: ## Buat akun owner pertama di production
	$(COMPOSE_PROD) run --rm --entrypoint /app/seed migrate

push: ## Kirim image ke registry (butuh IMAGE_REGISTRY dan IMAGE_TAG)
	$(COMPOSE_PROD) push backend web

backup: ## Cadangkan database production ke ./backups
	@mkdir -p backups
	$(COMPOSE_PROD) exec -T db pg_dump -U $${POSTGRES_USER} -d $${POSTGRES_DB} \
		| gzip > backups/jastipin-$$(date +%Y%m%d-%H%M%S).sql.gz
	@echo "✓ Backup tersimpan di ./backups"

# --- Pembersihan -------------------------------------------------------------

clean: ## Hapus container dan artefak build (data tetap aman)
	$(COMPOSE) down --remove-orphans
	rm -rf backend/tmp web/.next

reset: ## Hapus SEMUANYA termasuk data database
	$(COMPOSE) down -v --remove-orphans
	@echo "✓ Semua container dan volume dihapus"
