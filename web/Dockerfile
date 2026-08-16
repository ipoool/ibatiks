# =============================================================================
# Image production frontend Ibatiks.
#
# Memakai keluaran "standalone" Next.js: hanya berkas dan dependensi yang
# benar-benar dipakai server yang ikut disalin, bukan seluruh node_modules.
# =============================================================================

# --- Tahap dependensi --------------------------------------------------------
FROM node:22-alpine AS deps

WORKDIR /app

# Hanya manifest yang disalin supaya layer install dipakai ulang selama
# package-lock.json tidak berubah.
COPY package.json package-lock.json ./
RUN npm ci

# --- Tahap build -------------------------------------------------------------
FROM node:22-alpine AS builder

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED=1 \
    NODE_ENV=production

RUN npm run build

# --- Tahap runtime -----------------------------------------------------------
FROM node:22-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME=0.0.0.0

# Aplikasi dijalankan sebagai pengguna biasa, bukan root.
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
# Keluaran standalone sudah memuat server.js beserta dependensi minimalnya.
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD node -e "fetch('http://127.0.0.1:'+(process.env.PORT||3000)+'/login').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"

CMD ["node", "server.js"]
