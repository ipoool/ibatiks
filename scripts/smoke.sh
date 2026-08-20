#!/usr/bin/env bash
#
# Smoke test end-to-end Ibatiks.
#
# Menelusuri satu siklus jastip lengkap lewat API: trip dibuka, order masuk,
# DP dibayar, tripper belanja (lebih banyak dari pesanan), qty diedit, barang
# diterima, dikemas, ditagih, dilunasi, dikirim, lalu profitnya dicocokkan
# dengan hitungan manual.
#
# Pakai:
#   ./scripts/smoke.sh [BASE_URL] [EMAIL] [PASSWORD]
#
# Contoh:
#   ./scripts/smoke.sh http://localhost:8080 owner@ibatiks.id rahasia123
set -euo pipefail

BASE_URL="${1:-${SMOKE_BASE_URL:-http://localhost:8080}}"
EMAIL="${2:-${SEED_OWNER_EMAIL:-owner@ibatiks.id}}"
PASSWORD="${3:-${SEED_OWNER_PASSWORD:-rahasia123}}"
API="$BASE_URL/api/v1"

command -v jq >/dev/null || { echo "butuh jq: brew install jq"; exit 1; }

# Penanda unik supaya skrip bisa dijalankan berkali-kali tanpa bentrok SKU/email.
RUN_ID="$(date +%s)"
STEP=0
TOKEN=""

bold()  { printf '\033[1m%s\033[0m\n' "$1"; }
ok()    { printf '  \033[32m✓\033[0m %s\n' "$1"; }
fail()  { printf '  \033[31m✗\033[0m %s\n' "$1"; exit 1; }
step()  { STEP=$((STEP + 1)); printf '\n\033[1m[%02d] %s\033[0m\n' "$STEP" "$1"; }

# api METHOD PATH [JSON_BODY] — memanggil API dan menggagalkan skrip pada error.
api() {
  local method="$1" path="$2" body="${3:-}" response status payload
  local args=(-s -w '\n%{http_code}' -X "$method" "$API$path" -H 'Content-Type: application/json')
  [[ -n "$TOKEN" ]] && args+=(-H "Authorization: Bearer $TOKEN")
  [[ -n "$body" ]] && args+=(-d "$body")

  response="$(curl "${args[@]}")"
  status="$(tail -n1 <<<"$response")"
  payload="$(sed '$d' <<<"$response")"

  if [[ "$status" -ge 400 ]]; then
    printf '  \033[31m✗\033[0m %s %s -> HTTP %s\n' "$method" "$path" "$status" >&2
    jq -C . <<<"$payload" >&2 2>/dev/null || echo "$payload" >&2
    exit 1
  fi
  echo "$payload"
}

# expect ACTUAL EXPECTED LABEL
expect() {
  if [[ "$1" == "$2" ]]; then
    ok "$3 = $1"
  else
    fail "$3: diharapkan '$2', dapat '$1'"
  fi
}

# money membulatkan nominal bergaya NUMERIC ("1250000.00") menjadi bilangan bulat.
money() { printf '%.0f' "$1"; }

# days_from_now mencetak tanggal N hari dari sekarang sebagai YYYY-MM-DD.
# Ditulis lewat epoch karena flag tanggal relatif berbeda-beda antara date
# bawaan macOS (BSD), GNU coreutils, dan busybox di container Alpine.
days_from_now() {
  local epoch=$(( $(date +%s) + $1 * 86400 ))
  date -r "$epoch" +%Y-%m-%d 2>/dev/null ||
    date -d "@$epoch" +%Y-%m-%d 2>/dev/null ||
    date -u -d "@$epoch" +%Y-%m-%d
}

bold "Smoke test Ibatiks -> $BASE_URL"

# ---------------------------------------------------------------------------
step "Cek kesehatan service"
curl -sf "$BASE_URL/health" >/dev/null || fail "endpoint /health tidak merespons"
ok "/health merespons"
curl -sf "$BASE_URL/health/ready" >/dev/null || fail "database tidak siap"
ok "/health/ready merespons (database terhubung)"

# ---------------------------------------------------------------------------
step "Login sebagai owner"
TOKEN="$(api POST /auth/login "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.data.access_token')"
[[ -n "$TOKEN" && "$TOKEN" != "null" ]] || fail "gagal mendapatkan access token"
ok "token diterima"
expect "$(api GET /auth/me | jq -r '.data.role')" "owner" "role pengguna"

# ---------------------------------------------------------------------------
step "Buat customer dan 2 produk"
CUSTOMER_A="$(api POST /customers "{
  \"name\":\"Smoke Rina $RUN_ID\",
  \"phone_wa\":\"081234500001\",
  \"address\":\"Jl. Melati No. 1\",
  \"city\":\"Jakarta Selatan\",
  \"province\":\"DKI Jakarta\",
  \"postal_code\":\"12140\"
}" | jq -r '.data.id')"
ok "customer A dibuat"

CUSTOMER_B="$(api POST /customers "{
  \"name\":\"Smoke Budi $RUN_ID\",
  \"phone_wa\":\"0812-3450-0002\",
  \"address\":\"Jl. Kenanga No. 2\",
  \"city\":\"Bandung\"
}" | jq -r '.data.id')"
ok "customer B dibuat"

# Nomor telepon harus ternormalisasi ke format internasional untuk link wa.me.
expect "$(api GET "/customers/$CUSTOMER_B" | jq -r '.data.phone_wa')" "6281234500002" "normalisasi nomor WA"

PRODUCT_A="$(api POST /products "{
  \"sku\":\"SMOKE-A-$RUN_ID\",
  \"name\":\"Smoke Lotion 170ml\",
  \"base_currency\":\"JPY\",
  \"base_price\":\"1000\",
  \"markup_type\":\"percent\",
  \"markup_value\":\"30\",
  \"weight_gram\":250
}" | jq -r '.data.id')"
ok "produk A dibuat"

PRODUCT_B="$(api POST /products "{
  \"sku\":\"SMOKE-B-$RUN_ID\",
  \"name\":\"Smoke Snack Box\",
  \"base_currency\":\"JPY\",
  \"base_price\":\"500\",
  \"markup_type\":\"nominal\",
  \"markup_value\":\"25000\",
  \"weight_gram\":400
}" | jq -r '.data.id')"
ok "produk B dibuat"

# ---------------------------------------------------------------------------
step "Buat trip Jepang dan susun katalognya"
DEPART="$(days_from_now 7)"
RETURN="$(days_from_now 14)"

TRIP="$(api POST /trips "{
  \"title\":\"Smoke Trip Tokyo $RUN_ID\",
  \"country\":\"Jepang\",
  \"city\":\"Tokyo\",
  \"depart_date\":\"$DEPART\",
  \"return_date\":\"$RETURN\",
  \"currency\":\"JPY\",
  \"exchange_rate\":\"100\"
}" | jq -r '.data.id')"
ok "trip dibuat (kurs JPY 1 = Rp100)"

api PATCH "/trips/$TRIP/status" '{"status":"open"}' >/dev/null
ok "trip dibuka untuk order"

# Markup persen: 1000 JPY x 100 = Rp100.000, +30% = Rp130.000
ITEM_A="$(api POST "/trips/$TRIP/items" "{
  \"product_id\":\"$PRODUCT_A\",
  \"cost_price\":\"1000\",
  \"markup_type\":\"percent\",
  \"markup_value\":\"30\"
}")"
expect "$(money "$(jq -r '.data.sell_price' <<<"$ITEM_A")")" "130000" "harga jual produk A (markup 30%)"

# Markup nominal: 500 JPY x 100 = Rp50.000, +Rp25.000 = Rp75.000
ITEM_B="$(api POST "/trips/$TRIP/items" "{
  \"product_id\":\"$PRODUCT_B\",
  \"cost_price\":\"500\",
  \"markup_type\":\"nominal\",
  \"markup_value\":\"25000\"
}")"
expect "$(money "$(jq -r '.data.sell_price' <<<"$ITEM_B")")" "75000" "harga jual produk B (markup nominal)"

# ---------------------------------------------------------------------------
step "Catat 2 order dan terima DP"
# Order A: 2 x produk A (Rp260.000) + 1 x produk B (Rp75.000) = Rp335.000
ORDER_A="$(api POST /orders "{
  \"trip_id\":\"$TRIP\",
  \"customer_id\":\"$CUSTOMER_A\",
  \"order_source\":\"instagram\",
  \"items\":[
    {\"product_id\":\"$PRODUCT_A\",\"qty\":2},
    {\"product_id\":\"$PRODUCT_B\",\"qty\":1}
  ]
}")"
ORDER_A_ID="$(jq -r '.data.id' <<<"$ORDER_A")"
expect "$(money "$(jq -r '.data.total' <<<"$ORDER_A")")" "335000" "total order A"
expect "$(money "$(jq -r '.data.dp_required' <<<"$ORDER_A")")" "167500" "DP order A (50% default)"
# Alamat customer harus tersalin otomatis ke order.
expect "$(jq -r '.data.shipping_city' <<<"$ORDER_A")" "Jakarta Selatan" "kota kirim tersalin dari customer"
expect "$(jq -r '.data.order_source' <<<"$ORDER_A")" "instagram" "asal order A tercatat"
# Order baru langsung menunggu DP tanpa perlu diubah statusnya manual.
expect "$(jq -r '.data.status' <<<"$ORDER_A")" "awaiting_dp" "status awal order A"

# Order B: 3 x produk A = Rp390.000
ORDER_B="$(api POST /orders "{
  \"trip_id\":\"$TRIP\",
  \"customer_id\":\"$CUSTOMER_B\",
  \"items\":[{\"product_id\":\"$PRODUCT_A\",\"qty\":3}]
}")"
ORDER_B_ID="$(jq -r '.data.id' <<<"$ORDER_B")"
expect "$(money "$(jq -r '.data.total' <<<"$ORDER_B")")" "390000" "total order B"
expect "$(jq -r '.data.order_source' <<<"$ORDER_B")" "whatsapp" "asal order B memakai default WhatsApp"

# Asal order yang tidak dikenal harus ditolak, bukan disimpan apa adanya.
BAD_SOURCE="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/orders" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"trip_id\":\"$TRIP\",\"customer_id\":\"$CUSTOMER_B\",\"order_source\":\"telepati\",\"items\":[{\"product_id\":\"$PRODUCT_A\",\"qty\":1}]}")"
expect "$BAD_SOURCE" "422" "asal order tak dikenal ditolak"

# Daftar order bisa disaring per channel.
BY_SOURCE="$(api GET "/orders?source=instagram")"
expect "$(jq -r --arg id "$ORDER_A_ID" '[.data[] | select(.id==$id)] | length' <<<"$BY_SOURCE")" "1" "filter order per channel"


# ---------------------------------------------------------------------------
step "Buka daftar belanja tripper"
# Sebelum DP masuk, permintaan hanya boleh muncul sebagai "menunggu DP" —
# tripper tidak boleh membelanjakan order yang uang mukanya belum ada.
BEFORE_DP="$(api GET "/trips/$TRIP/shopping-list")"
expect "$(jq -r --arg id "$PRODUCT_A" '.data[] | select(.product_id==$id) | .qty_ordered' <<<"$BEFORE_DP")" "0" "qty produk A sebelum DP masuk"
expect "$(jq -r --arg id "$PRODUCT_A" '.data[] | select(.product_id==$id) | .qty_awaiting_dp' <<<"$BEFORE_DP")" "5" "qty produk A yang masih menunggu DP"

api POST "/orders/$ORDER_A_ID/payments" '{"type":"dp","amount":"167500","method":"transfer"}' >/dev/null
api POST "/orders/$ORDER_B_ID/payments" '{"type":"dp","amount":"195000","method":"transfer"}' >/dev/null
ok "DP kedua order diverifikasi"

SHOPPING="$(api GET "/trips/$TRIP/shopping-list")"
# Produk A dipesan 2 (order A) + 3 (order B) = 5 unit.
expect "$(jq -r --arg id "$PRODUCT_A" '.data[] | select(.product_id==$id) | .qty_ordered' <<<"$SHOPPING")" "5" "total qty produk A di daftar belanja"
expect "$(jq -r --arg id "$PRODUCT_A" '.data[] | select(.product_id==$id) | .qty_awaiting_dp' <<<"$SHOPPING")" "0" "tidak ada lagi yang menunggu DP"
expect "$(jq -r --arg id "$PRODUCT_B" '.data[] | select(.product_id==$id) | .qty_ordered' <<<"$SHOPPING")" "1" "total qty produk B di daftar belanja"
expect "$(jq -r --arg id "$PRODUCT_A" '.data[] | select(.product_id==$id) | .order_count' <<<"$SHOPPING")" "2" "jumlah order yang memesan produk A"

# ---------------------------------------------------------------------------
step "Tripper belanja lebih banyak dari pesanan"
# Beli 8 unit produk A padahal yang dipesan 5 -> 3 unit sisanya jadi stok.
BUY_A="$(api POST "/trips/$TRIP/purchases" "{
  \"product_id\":\"$PRODUCT_A\",
  \"qty\":8,
  \"unit_cost_foreign\":\"1000\",
  \"store_name\":\"Don Quijote Shibuya\"
}")"
expect "$(jq -r '.data.qty_to_orders' <<<"$BUY_A")" "5" "unit produk A yang menutup pesanan"
expect "$(jq -r '.data.qty_to_stock' <<<"$BUY_A")" "3" "unit produk A yang masuk stok"

api POST "/trips/$TRIP/purchases" "{
  \"product_id\":\"$PRODUCT_B\",
  \"qty\":1,
  \"unit_cost_foreign\":\"500\"
}" >/dev/null
ok "produk B dibeli sesuai pesanan"

STOCK="$(api GET /stock)"
expect "$(jq -r --arg id "$PRODUCT_A" '.data[] | select(.product_id==$id) | .qty_on_hand' <<<"$STOCK")" "3" "stok produk A"
expect "$(money "$(jq -r --arg id "$PRODUCT_A" '.data[] | select(.product_id==$id) | .avg_cost_idr' <<<"$STOCK")")" "100000" "HPP rata-rata stok produk A"

# ---------------------------------------------------------------------------
step "Edit qty item order dan pastikan total ikut berubah"
ITEM_ID="$(api GET "/orders/$ORDER_B_ID" | jq -r '.data.items[0].id')"
# Turun dari 3 ke 2 unit: total Rp390.000 -> Rp260.000, 1 unit dilepas ke stok.
EDITED="$(api PUT "/orders/$ORDER_B_ID/items/$ITEM_ID" '{"qty":2}')"
expect "$(money "$(jq -r '.data.total' <<<"$EDITED")")" "260000" "total order B setelah qty dikurangi"
expect "$(money "$(jq -r '.data.balance_due' <<<"$EDITED")")" "65000" "sisa tagihan order B"

STOCK="$(api GET /stock)"
expect "$(jq -r --arg id "$PRODUCT_A" '.data[] | select(.product_id==$id) | .qty_on_hand' <<<"$STOCK")" "4" "stok produk A setelah 1 unit dilepas dari pesanan"

AUDIT="$(api GET "/audit-logs?entity=order&entity_id=$ORDER_B_ID")"
expect "$(jq -r '[.data[] | select(.action=="item_change")] | length > 0' <<<"$AUDIT")" "true" "perubahan qty tercatat di audit log"

# ---------------------------------------------------------------------------
step "Kemas dan terbitkan invoice"
api PATCH "/trips/$TRIP/status" '{"status":"closed"}' >/dev/null
ok "order untuk trip ini ditutup"

# Tidak ada lagi langkah mencocokkan barang datang. Yang ditagih dihitung dari
# qty yang benar-benar terbeli, dan itu sudah diselaraskan sendiri saat tripper
# mencatat belanjanya.
ORDER_A_STATE="$(api GET "/orders/$ORDER_A_ID")"
expect "$(jq -r '.data.status' <<<"$ORDER_A_STATE")" "dp_paid" "status order A tetap Diproses setelah belanja"

# Ongkir hanya punya satu sumber: kurir. Pada database yang baru direset, kurir
# belum bisa menjawab — entah karena API key-nya belum diisi, entah karena kota
# asal di Pengaturan masih kosong. Yang diuji di sini adalah permintaannya
# ditolak dengan alasan yang bisa ditindaklanjuti, bukan dijawab angka tebakan;
# dulu ada tabel tarif yang diam-diam menutupi keadaan itu.
#
# Pesannya tidak dipatok pada satu kalimat: mana dari dua sebab itu yang muncul
# bergantung pada isi .env mesin yang menjalankan, dan mematoknya membuat smoke
# gagal begitu seseorang mengisi API key-nya — kegagalan yang tidak menandakan
# apa pun soal aplikasinya.
EST_STATUS="$(curl -s -o /tmp/smoke-estimate.json -w '%{http_code}' -X POST \
  "$API/orders/$ORDER_A_ID/shipping-estimate" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"weight_gram":900,"length_cm":40,"width_cm":30,"height_cm":25}')"
expect "$EST_STATUS" "409" "estimasi ongkir ditolak saat kurir belum bisa menjawab"
expect "$(jq -r '.error.message | test("RAJAONGKIR_API_KEY|kota asal")' /tmp/smoke-estimate.json)" "true" \
  "penolakannya menyebutkan apa yang harus dibereskan"
rm -f /tmp/smoke-estimate.json

# Invoice pelunasan menagih seluruh sisa pesanan termasuk ongkir, jadi ia tidak
# boleh terbit sebelum paketnya ditimbang — customer akan menerima tagihan yang
# nilainya masih berubah.
TOO_EARLY="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/orders/$ORDER_A_ID/invoices" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"type":"final"}')"
expect "$TOO_EARLY" "409" "invoice pelunasan ditolak sebelum ongkir ditetapkan"

# Ongkir ditetapkan bersama data kemasan. Angkanya diketik admin sendiri —
# jaring pengaman yang menggantikan tabel tarif.
PACKED="$(api POST "/orders/$ORDER_A_ID/pack" '{"courier":"JNE","service":"REG","weight_gram":900,"length_cm":40,"width_cm":30,"height_cm":25,"shipping_fee":"60000"}')"
expect "$(jq -r '.data.status' <<<"$PACKED")" "ready" "status paket setelah dikemas"
expect "$(jq -r '.data.length_cm' <<<"$PACKED")" "40" "dimensi paket tersimpan"
# Estimasi pembanding kosong karena kurirnya tidak bisa dihubungi. Itu tidak
# menghentikan pengemasan: ongkir yang dipakai adalah yang diketik admin.
expect "$(money "$(jq -r '.data.estimated_cost' <<<"$PACKED")")" "0" "estimasi pembanding kosong saat kurir tak terhubung"

# Mengemas tidak menggeser status order, tapi ongkirnya masuk ke total.
AFTER_PACK="$(api GET "/orders/$ORDER_A_ID")"
expect "$(jq -r '.data.status' <<<"$AFTER_PACK")" "dp_paid" "status order A tetap Diproses setelah dikemas"
expect "$(money "$(jq -r '.data.shipping_fee' <<<"$AFTER_PACK")")" "60000" "ongkir tersimpan ke order"
expect "$(money "$(jq -r '.data.total' <<<"$AFTER_PACK")")" "395000" "total order naik sebesar ongkir"
expect "$(money "$(jq -r '.data.dp_required' <<<"$AFTER_PACK")")" "167500" "DP tidak ikut dihitung ulang saat ongkir masuk"

# Order kini muncul sebagai kandidat invoice pelunasan.
CANDIDATES="$(api GET "/invoices/candidates")"
expect "$(jq -r --arg id "$ORDER_A_ID" '[.data[] | select(.order_id==$id)] | length' <<<"$CANDIDATES")" "1" "order A siap ditagih pelunasan"

INVOICE="$(api POST "/orders/$ORDER_A_ID/invoices" '{"type":"final"}')"
INVOICE_ID="$(jq -r '.data.id' <<<"$INVOICE")"
expect "$(money "$(jq -r '.data.total' <<<"$INVOICE")")" "395000" "total invoice order A sudah termasuk ongkir"
expect "$(money "$(jq -r '.data.shipping_fee' <<<"$INVOICE")")" "60000" "ongkir tertulis di invoice"
expect "$(money "$(jq -r '.data.amount_due' <<<"$INVOICE")")" "227500" "sisa pelunasan pada invoice"

# Menerbitkan invoice tidak menggeser status order; yang menggesernya uang masuk.
expect "$(jq -r '.data.status' <<<"$(api GET "/orders/$ORDER_A_ID")")" "dp_paid" "status order A tetap Diproses setelah invoice terbit"

# Sudah punya invoice pelunasan yang berlaku, jadi hilang dari daftar kandidat.
expect "$(jq -r --arg id "$ORDER_A_ID" '[.data[] | select(.order_id==$id)] | length' <<<"$(api GET "/invoices/candidates")")" "0" "order A hilang dari kandidat setelah invoice terbit"

PDF_STATUS="$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$API/invoices/$INVOICE_ID/pdf")"
expect "$PDF_STATUS" "200" "unduh PDF invoice"

WA="$(api GET "/invoices/$INVOICE_ID/message")"
expect "$(jq -r '.data.whatsapp_url | startswith("https://wa.me/6281234500001")' <<<"$WA")" "true" "link wa.me terbentuk dengan nomor customer"
api POST "/invoices/$INVOICE_ID/mark-sent" '{"channel":"wa"}' >/dev/null
ok "invoice ditandai sudah dikirim via WA"

# ---------------------------------------------------------------------------
step "Catat pelunasan lalu kirim dengan resi JNE"
PAID="$(api POST "/orders/$ORDER_A_ID/payments" '{"type":"settlement","amount":"227500","method":"transfer"}')"
expect "$(jq -r '.data.status' <<<"$PAID")" "paid" "status order A setelah lunas"
expect "$(money "$(jq -r '.data.balance_due' <<<"$PAID")")" "0" "sisa tagihan order A"

SHIPPED="$(api POST "/orders/$ORDER_A_ID/ship" '{"tracking_number":"JNE0012345678","shipping_cost":"22000"}')"
expect "$(jq -r '.data.status' <<<"$SHIPPED")" "shipped" "status paket setelah diserahkan ke kurir"
expect "$(jq -r '.data.tracking_number' <<<"$SHIPPED")" "JNE0012345678" "nomor resi tersimpan"

# Antrean pengiriman menyaring menurut tahap kerja, bukan status tersimpan.
expect "$(jq -r --arg id "$ORDER_A_ID" '[.data[] | select(.order_id==$id)] | length' <<<"$(api GET "/shipments?stage=terkirim")")" "1" "order A masuk tahap sudah dikirim"
expect "$(jq -r --arg id "$ORDER_A_ID" '[.data[] | select(.order_id==$id)] | length' <<<"$(api GET "/shipments?stage=perlu_kemas")")" "0" "order A tidak lagi di tahap perlu dikemas"

# Label pengiriman 100 x 150 mm menggantikan surat jalan.
LABEL_STATUS="$(curl -s -o /tmp/smoke-label.pdf -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$API/orders/$ORDER_A_ID/label")"
expect "$LABEL_STATUS" "200" "unduh label pengiriman"
expect "$(head -c 4 /tmp/smoke-label.pdf)" "%PDF" "label berbentuk PDF"
# MediaBox 283.46 x 425.20 pt = 100 x 150 mm, ukuran kertas thermal kurir.
expect "$(grep -a -o '/MediaBox \[0 0 283.46 425.20\]' /tmp/smoke-label.pdf | head -1)" "/MediaBox [0 0 283.46 425.20]" "ukuran label 100 x 150 mm"
rm -f /tmp/smoke-label.pdf

SHIP_MSG="$(api GET "/orders/$ORDER_A_ID/shipment-message")"
expect "$(jq -r '.data.text | contains("JNE0012345678")' <<<"$SHIP_MSG")" "true" "pesan pengiriman memuat nomor resi"

# ---------------------------------------------------------------------------
step "Tambah biaya trip dan cocokkan laporan profit"
api POST "/trips/$TRIP/expenses" '{"category":"bagasi","description":"Extra baggage 10kg","amount":"850000"}' >/dev/null
api POST "/trips/$TRIP/expenses" '{"category":"transport","description":"Kereta bandara","amount":"150000"}' >/dev/null
ok "biaya trip dicatat (Rp1.000.000)"

REPORT="$(api GET "/reports/profit?trip_id=$TRIP")"

# Omzet  = order A (335.000) + order B (260.000)          = 595.000
# HPP    = 4 unit produk A (400.000) + 1 produk B (50.000) = 450.000
#          (order A: 2 unit A + 1 unit B, order B: 2 unit A)
# Kotor  = 595.000 - 450.000                               = 145.000
# Bersih = 145.000 - 1.000.000 biaya trip                  = -855.000
# Omzet ikut memuat ongkir yang ditagihkan ke customer (Rp60.000 pada order A),
# sementara ongkir yang benar-benar dibayar ke kurir dicatat terpisah pada
# pengiriman — selisih keduanya memang keuntungan atau tombokan toko.
expect "$(money "$(jq -r '.data.revenue' <<<"$REPORT")")" "655000" "omzet trip"
expect "$(money "$(jq -r '.data.cogs' <<<"$REPORT")")" "450000" "HPP riil trip"
expect "$(money "$(jq -r '.data.gross_profit' <<<"$REPORT")")" "205000" "laba kotor trip"
expect "$(money "$(jq -r '.data.trip_expenses' <<<"$REPORT")")" "1000000" "biaya perjalanan"
expect "$(money "$(jq -r '.data.net_profit' <<<"$REPORT")")" "-795000" "laba bersih trip"

# Surplus 4 unit produk A senilai Rp400.000 tidak boleh dibebankan sebagai HPP;
# nilainya tetap dipegang sebagai aset stok.
expect "$(jq -r '.data.surplus_stock_qty' <<<"$REPORT")" "4" "qty surplus yang masuk stok"
expect "$(money "$(jq -r '.data.surplus_stock_value' <<<"$REPORT")")" "400000" "nilai surplus stok"
# Uang keluar = seluruh belanja (8x100.000 + 1x50.000) + biaya trip 1.000.000
expect "$(money "$(jq -r '.data.total_capital_out' <<<"$REPORT")")" "1850000" "total modal keluar"

# ---------------------------------------------------------------------------
step "Cek dashboard dan laporan piutang"
DASH="$(api GET /reports/dashboard)"
# "Trip aktif" kini berarti trip yang masih menerima order. Trip pada skrip ini
# sudah ditutup, jadi yang diuji adalah angkanya ikut turun — bukan sekadar ada.
expect "$(jq -r '.data.active_trips' <<<"$DASH")" "0" "trip yang sudah ditutup tidak lagi dihitung aktif"

api PATCH "/trips/$TRIP/status" '{"status":"open"}' >/dev/null
DASH_OPEN="$(api GET /reports/dashboard)"
[[ "$(jq -r '.data.active_trips' <<<"$DASH_OPEN")" -ge 1 ]] || fail "dashboard tidak menghitung trip yang dibuka"
ok "dashboard menghitung trip yang masih menerima order"
api PATCH "/trips/$TRIP/status" '{"status":"closed"}' >/dev/null

RECEIVABLES="$(api GET /reports/receivables)"
expect "$(jq -r --arg id "$ORDER_B_ID" '[.data[] | select(.order_id==$id)] | length' <<<"$RECEIVABLES")" "1" "order B muncul di laporan piutang"

CSV_STATUS="$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$API/reports/receivables?format=csv")"
expect "$CSV_STATUS" "200" "ekspor CSV piutang"

# ---------------------------------------------------------------------------
step "Cek rekap penjualan per customer dan per channel"
CUSTOMERS="$(api GET "/reports/customers?trip_id=$TRIP")"
expect "$(jq -r --arg id "$CUSTOMER_A" '.data[] | select(.customer_id==$id) | .order_count' <<<"$CUSTOMERS")" "1" "jumlah order customer A"
expect "$(money "$(jq -r --arg id "$CUSTOMER_A" '.data[] | select(.customer_id==$id) | .revenue' <<<"$CUSTOMERS")")" "395000" "omzet customer A"
expect "$(money "$(jq -r --arg id "$CUSTOMER_B" '.data[] | select(.customer_id==$id) | .outstanding' <<<"$CUSTOMERS")")" "65000" "piutang customer B"

CHANNELS="$(api GET "/reports/channels?trip_id=$TRIP")"
expect "$(money "$(jq -r '.data[] | select(.order_source=="instagram") | .revenue' <<<"$CHANNELS")")" "395000" "omzet channel Instagram"
expect "$(money "$(jq -r '.data[] | select(.order_source=="whatsapp") | .revenue' <<<"$CHANNELS")")" "260000" "omzet channel WhatsApp"
# Porsi omzet seluruh channel harus menjumlah 100%.
expect "$(jq -r '[.data[].revenue_share | tonumber] | add | round' <<<"$CHANNELS")" "100" "porsi omzet seluruh channel"

for REPORT_PATH in customers channels; do
  RSTATUS="$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$API/reports/$REPORT_PATH?format=csv")"
  expect "$RSTATUS" "200" "ekspor CSV laporan $REPORT_PATH"
done

# ---------------------------------------------------------------------------
step "Cek riwayat harga produk antar trip"
HISTORY="$(api GET "/products/$PRODUCT_A/price-history")"
expect "$(jq -r --arg t "$TRIP" '[.data[] | select(.trip_id==$t)] | length' <<<"$HISTORY")" "1" "trip ini muncul di riwayat harga"
expect "$(jq -r --arg t "$TRIP" '.data[] | select(.trip_id==$t) | .actual_cost' <<<"$HISTORY" | cut -d. -f1)" "1000" "harga beli riil tercatat di riwayat"
expect "$(jq -r --arg t "$TRIP" '.data[] | select(.trip_id==$t) | .qty_purchased' <<<"$HISTORY")" "8" "qty pembelian tercatat di riwayat"

# ---------------------------------------------------------------------------
step "Cek kurs otomatis dan identitas customer"
RATE="$(api GET "/exchange-rate?from=IDR")"
expect "$(jq -r '.data.rate' <<<"$RATE")" "1" "kurs rupiah ke rupiah"

# Nomor WhatsApp adalah identitas customer; format berbeda untuk nomor yang
# sama harus ditolak supaya laporan per customer tidak terpecah.
DUP_STATUS="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/customers" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Andi Ganda","phone_wa":"0812 3450 0001"}')"
expect "$DUP_STATUS" "409" "customer dengan nomor WA yang sama ditolak"

# ---------------------------------------------------------------------------
step "Cek penjagaan aturan bisnis"
# Order yang sudah dikirim tidak boleh diedit lagi.
SHIPPED_ITEM="$(api GET "/orders/$ORDER_A_ID" | jq -r '.data.items[0].id')"
GUARD_STATUS="$(curl -s -o /dev/null -w '%{http_code}' -X PUT "$API/orders/$ORDER_A_ID/items/$SHIPPED_ITEM" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"qty":9}')"
expect "$GUARD_STATUS" "409" "edit order yang sudah dikirim ditolak"

# Order yang belum lunas tidak boleh dikirim.
UNPAID_STATUS="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/orders/$ORDER_B_ID/ship" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"tracking_number":"JNE0099999999","shipping_cost":"20000"}')"
expect "$UNPAID_STATUS" "409" "kirim order yang belum lunas ditolak"

# Ongkir yang ditetapkan setelah customer terlanjur melunasi harus menarik
# ordernya kembali ke Diproses. Tanpa itu ia tetap berlabel Pembayaran Lunas,
# ikut masuk antrean siap kirim, dan barangnya berangkat sementara ongkirnya
# tidak pernah tertagih.
LUNAS_DULU="$(api POST "/orders/$ORDER_B_ID/payments" "{\"type\":\"settlement\",\"amount\":\"$(money "$(jq -r '.data.balance_due' <<<"$(api GET "/orders/$ORDER_B_ID")")")\",\"method\":\"transfer\"}")"
expect "$(jq -r '.data.status' <<<"$LUNAS_DULU")" "paid" "order B lunas sebelum paketnya ditimbang"

api POST "/orders/$ORDER_B_ID/pack" '{"courier":"JNE","service":"REG","weight_gram":1000,"shipping_fee":"35000"}' >/dev/null
AFTER_FEE="$(api GET "/orders/$ORDER_B_ID")"
expect "$(jq -r '.data.status' <<<"$AFTER_FEE")" "dp_paid" "order B kembali ke Diproses saat ongkir menambah tagihan"
expect "$(money "$(jq -r '.data.balance_due' <<<"$AFTER_FEE")")" "35000" "sisa tagihan order B sebesar ongkirnya"

# --- Menghapus trip ----------------------------------------------------------
# Trip yang ordernya sudah diserahkan ke kurir tidak boleh dihapus: penjualannya
# sudah jadi, dan menghapus catatannya tidak membatalkan apa pun.
IMPACT="$(api GET "/trips/$TRIP/deletion-impact")"
expect "$(jq -r '.data.shipped_orders | length' <<<"$IMPACT")" "1" "order terkirim tercatat sebagai penghalang hapus trip"
expect "$(jq -r '.data.stock_on_hand | length > 0' <<<"$IMPACT")" "true" "surplus stok tercatat sebagai penghalang hapus trip"
# Order A lunas Rp395.000 dan order B lunas Rp260.000 — keduanya ikut hilang
# dari pembukuan kalau tripnya dihapus, dan itulah angka yang wajib ditampilkan
# dialog konfirmasi.
expect "$(money "$(jq -r '.data.payments_total' <<<"$IMPACT")")" "655000" "uang diterima yang akan hilang kalau trip dihapus"

DEL_BLOCKED="$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$API/trips/$TRIP" -H "Authorization: Bearer $TOKEN")"
expect "$DEL_BLOCKED" "409" "hapus trip berisi order terkirim ditolak"

# Trip kosong boleh dihapus tanpa syarat.
EMPTY_TRIP="$(api POST /trips '{"title":"Trip Kosong Uji","country":"Korea","depart_date":"2026-12-01","return_date":"2026-12-07","currency":"KRW","exchange_rate":"12"}' | jq -r '.data.id')"
DEL_EMPTY="$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$API/trips/$EMPTY_TRIP" -H "Authorization: Bearer $TOKEN")"
expect "$DEL_EMPTY" "204" "hapus trip kosong diterima"
expect "$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$API/trips/$EMPTY_TRIP")" "404" "trip yang dihapus tidak bisa dibuka lagi"

# Lima kali gagal login mengunci email itu selama lima menit. Dipakai email
# khusus supaya akun owner yang dipakai smoke sendiri tidak ikut terkunci.
UJI_EMAIL="uji-blokir-$$@ibatiks.id"
for _ in 1 2 3 4; do
  curl -s -o /dev/null -X POST "$API/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$UJI_EMAIL\",\"password\":\"salah\"}"
done
BLOKIR_STATUS="$(curl -s -o /tmp/smoke-blokir.json -w '%{http_code}' -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' -d "{\"email\":\"$UJI_EMAIL\",\"password\":\"salah\"}")"
expect "$BLOKIR_STATUS" "429" "login dikunci setelah 5 kali gagal"
expect "$(jq -r '.error.code' /tmp/smoke-blokir.json)" "TOO_MANY_REQUESTS" "kodenya membedakan dikunci dari password salah"
rm -f /tmp/smoke-blokir.json

# Email lain tidak ikut terkunci — penguncian dihitung per email.
LAIN_STATUS="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' -d '{"email":"orang-lain@ibatiks.id","password":"salah"}')"
expect "$LAIN_STATUS" "401" "email lain tidak ikut terkunci"

# Akun owner yang dipakai smoke tetap bisa login.
OWNER_STATUS="$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")"
expect "$OWNER_STATUS" "200" "akun yang tidak pernah gagal tetap bisa masuk"

# Endpoint tanpa token harus ditolak.
NOAUTH_STATUS="$(curl -s -o /dev/null -w '%{http_code}' "$API/orders")"
expect "$NOAUTH_STATUS" "401" "akses tanpa token ditolak"

printf '\n\033[1;32m✓ Smoke test lulus — %d langkah\033[0m\n' "$STEP"
