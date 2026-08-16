# =============================================================================
# Image production backend Ibatiks.
#
# Hasil akhirnya adalah image distroless berisi tiga binary statis tanpa shell,
# package manager, maupun pustaka sistem — permukaan serangnya sangat kecil dan
# ukurannya hanya belasan megabyte.
# =============================================================================

# --- Tahap build -------------------------------------------------------------
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Manifest disalin lebih dulu supaya layer unduhan dependensi bisa dipakai ulang
# selama go.mod/go.sum tidak berubah.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO dimatikan agar binary benar-benar statis dan bisa jalan di image distroless.
# -s -w membuang tabel simbol dan info debug sehingga ukurannya jauh lebih kecil.
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/api     ./cmd/api     && \
    go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate && \
    go build -trimpath -ldflags="-s -w" -o /out/seed    ./cmd/seed

# Direktori data dibuat di sini lengkap dengan kepemilikannya. Volume Docker
# yang di-mount ke /data akan mewarisi kepemilikan ini, sehingga proses non-root
# tetap bisa menulis berkas unggahan dan PDF invoice.
RUN mkdir -p /out/data/uploads /out/data/invoices && \
    chown -R 65532:65532 /out/data

# --- Tahap runtime -----------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/api     /app/api
COPY --from=builder /out/migrate /app/migrate
COPY --from=builder /out/seed    /app/seed
COPY --from=builder --chown=65532:65532 /out/data /data

# 65532 adalah uid "nonroot" pada image distroless.
USER 65532:65532

ENV APP_PORT=8080 \
    UPLOAD_DIR=/data/uploads \
    INVOICE_DIR=/data/invoices

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/api", "--health"]

ENTRYPOINT ["/app/api"]
