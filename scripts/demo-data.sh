#!/usr/bin/env bash
#
# Mengisi database dengan satu trip yang sudah berjalan penuh, supaya setiap
# layar aplikasi punya isi yang wajar. Dipakai untuk menyiapkan screenshot
# manual pengguna dan untuk mencoba aplikasi tanpa mengetik data manual.
#
# Jalankan setelah `make seed-demo` (yang membuat produk, customer, dan trip).
#
#   ./scripts/demo-data.sh [BASE_URL] [EMAIL] [PASSWORD]
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
EMAIL="${2:-owner@ibatiks.id}"
PASSWORD="${3:-rahasia123}"
API="$BASE_URL/api/v1"

command -v jq >/dev/null || { echo "butuh jq"; exit 1; }

TOKEN=""
say() { printf '  %s\n' "$1"; }
step() { printf '\n\033[1m%s\033[0m\n' "$1"; }

api() {
  local method="$1" path="$2" body="${3:-}" response status payload
  local args=(-s -w '\n%{http_code}' -X "$method" "$API$path" -H 'Content-Type: application/json')
  [[ -n "$TOKEN" ]] && args+=(-H "Authorization: Bearer $TOKEN")
  [[ -n "$body" ]] && args+=(-d "$body")

  response="$(curl "${args[@]}")"
  status="$(tail -n1 <<<"$response")"
  payload="$(sed '$d' <<<"$response")"

  if [[ "$status" -ge 400 ]]; then
    echo "GAGAL $method $path -> HTTP $status" >&2
    echo "$payload" >&2
    exit 1
  fi
  echo "$payload"
}

TOKEN="$(api POST /auth/login "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.data.access_token')"

# --- Ambil trip dan katalog yang dibuat oleh `seed --demo` -------------------
step "Membaca trip dan katalog"
TRIP="$(api GET '/trips?per_page=1' | jq -r '.data[0].id')"
TRIP_CODE="$(api GET "/trips/$TRIP" | jq -r '.data.code')"
say "trip $TRIP_CODE"

# Endpoint mengembalikan amplop {"data": [...]}, jadi isinya diambil dulu.
CATALOG="$(api GET "/trips/$TRIP/items" | jq -c '.data')"
sku_id() { jq -r --arg s "$1" '.[] | select(.product_sku==$s) | .product_id' <<<"$CATALOG"; }

SKN1="$(sku_id SKN-0001)"   # Hada Labo Gokujyun Lotion
SKN2="$(sku_id SKN-0002)"   # Melano CC Serum
SNK1="$(sku_id SNK-0001)"   # Tokyo Banana
SNK2="$(sku_id SNK-0002)"   # Kit Kat Matcha
FSH1="$(sku_id FSH-0001)"   # Uniqlo Airism

CUSTOMERS="$(api GET '/customers?per_page=50')"
cust_id() { jq -r --arg n "$1" '.data[] | select(.name==$n) | .id' <<<"$CUSTOMERS"; }

RINA="$(cust_id 'Rina Kartika')"
BUDI="$(cust_id 'Budi Santoso')"
SARI="$(cust_id 'Sari Dewi')"

# --- Order ------------------------------------------------------------------
step "Membuat order"

ORDER_RINA="$(api POST /orders "{
  \"trip_id\":\"$TRIP\", \"customer_id\":\"$RINA\",
  \"items\":[
    {\"product_id\":\"$SKN1\",\"qty\":2},
    {\"product_id\":\"$SNK1\",\"qty\":1},
    {\"product_id\":\"$SNK2\",\"qty\":2}
  ],
  \"order_source\":\"whatsapp\",
  \"shipping_fee\":\"22000\",
  \"notes\":\"Tolong dibubble wrap dobel ya, botolnya kaca\"
}")"
ORDER_RINA_ID="$(jq -r '.data.id' <<<"$ORDER_RINA")"
say "$(jq -r '.data.order_number' <<<"$ORDER_RINA") — Rina Kartika"

ORDER_BUDI="$(api POST /orders "{
  \"trip_id\":\"$TRIP\", \"customer_id\":\"$BUDI\",
  \"items\":[
    {\"product_id\":\"$SKN2\",\"qty\":3},
    {\"product_id\":\"$FSH1\",\"qty\":2}
  ],
  \"order_source\":\"instagram\",
  \"shipping_fee\":\"25000\"
}")"
ORDER_BUDI_ID="$(jq -r '.data.id' <<<"$ORDER_BUDI")"
say "$(jq -r '.data.order_number' <<<"$ORDER_BUDI") — Budi Santoso"

ORDER_SARI="$(api POST /orders "{
  \"trip_id\":\"$TRIP\", \"customer_id\":\"$SARI\",
  \"items\":[
    {\"product_id\":\"$SNK1\",\"qty\":3},
    {\"product_id\":\"$SKN1\",\"qty\":1}
  ],
  \"order_source\":\"tiktok\",
  \"shipping_fee\":\"30000\",
  \"notes\":\"Untuk oleh-oleh kantor\"
}")"
ORDER_SARI_ID="$(jq -r '.data.id' <<<"$ORDER_SARI")"
say "$(jq -r '.data.order_number' <<<"$ORDER_SARI") — Sari Dewi"

# --- DP ---------------------------------------------------------------------
step "Mencatat DP"
# Order yang baru dicatat sudah otomatis berstatus "menunggu DP".

dp_of() { api GET "/orders/$1" | jq -r '.data.dp_required'; }

api POST "/orders/$ORDER_RINA_ID/payments" \
  "{\"type\":\"dp\",\"amount\":\"$(dp_of "$ORDER_RINA_ID")\",\"method\":\"transfer\",\"reference\":\"BCA a/n Rina Kartika\"}" >/dev/null
say "DP Rina diterima"

api POST "/orders/$ORDER_BUDI_ID/payments" \
  "{\"type\":\"dp\",\"amount\":\"$(dp_of "$ORDER_BUDI_ID")\",\"method\":\"qris\",\"reference\":\"QRIS 2409\"}" >/dev/null
say "DP Budi diterima"

# Order Sari sengaja dibiarkan menunggu DP, supaya layar antrean ada isinya.
say "Order Sari sengaja dibiarkan menunggu DP"

# --- Belanja ----------------------------------------------------------------
step "Mencatat pembelian tripper"
api PATCH "/trips/$TRIP/status" '{"status":"shopping"}' >/dev/null

buy() {
  local product="$1" qty="$2" cost="$3" store="$4"
  api POST "/trips/$TRIP/purchases" \
    "{\"product_id\":\"$product\",\"qty\":$qty,\"unit_cost_foreign\":\"$cost\",\"store_name\":\"$store\"}" \
    | jq -r '"  \(.data.qty_to_orders) ke pesanan, \(.data.qty_to_stock) ke stok"'
}

# Lotion dan Kit Kat dibeli lebih banyak dari pesanan supaya stok terisi.
buy "$SKN1" 6 880  "Don Quijote Shibuya"
buy "$SNK1" 4 1180 "Tokyo Station Gransta"
buy "$SNK2" 5 780  "Don Quijote Shibuya"
buy "$FSH1" 2 1500 "Uniqlo Ginza"

# Melano CC sengaja dibeli kurang dari pesanan: menirukan kejadian yang lazim
# di lapangan, yaitu stok toko habis. Ini membuat layar daftar belanja punya
# baris yang masih menyisakan pekerjaan, dan order Budi punya item berstatus
# "Sebagian".
api POST "/trips/$TRIP/purchases" \
  "{\"product_id\":\"$SKN2\",\"qty\":2,\"unit_cost_foreign\":\"1180\",\"store_name\":\"Matsumoto Kiyoshi\",\"notes\":\"Stok toko tinggal 2\"}" >/dev/null
say "Melano CC hanya dapat 2 dari 3 (stok toko habis)"

# --- Barang tiba ------------------------------------------------------------
step "Barang tiba di Indonesia"
api PATCH "/trips/$TRIP/status" '{"status":"arrived"}' >/dev/null

# Diterima sejumlah yang benar-benar dibeli, bukan sejumlah yang dipesan.
receive_as_purchased() {
  local order="$1"
  local items
  items="$(api GET "/orders/$order" | jq -c '[.data.items[] |
    {item_id: .id,
     qty_received: (if .qty_purchased < .qty then .qty_purchased else .qty end),
     status: (if .qty_purchased == 0 then "unavailable"
              elif .qty_purchased < .qty then "partial"
              else "purchased" end)}]')"
  api POST "/orders/$order/receive" "{\"items\":$items}" >/dev/null
}

receive_as_purchased "$ORDER_RINA_ID"
say "Barang order Rina dicocokkan, lengkap"
receive_as_purchased "$ORDER_BUDI_ID"
say "Barang order Budi dicocokkan, satu item sebagian"

# --- Order Rina diselesaikan sampai dikirim ---------------------------------
step "Menyelesaikan order Rina sampai dikirim"
api POST "/orders/$ORDER_RINA_ID/pack" \
  '{"courier":"JNE","service":"REG","weight_gram":1250,"length_cm":30,"width_cm":22,"height_cm":15}' >/dev/null
say "dikemas"

api POST "/orders/$ORDER_RINA_ID/invoices" '{"type":"final"}' >/dev/null
say "invoice pelunasan diterbitkan"

BALANCE="$(api GET "/orders/$ORDER_RINA_ID" | jq -r '.data.balance_due')"
api POST "/orders/$ORDER_RINA_ID/payments" \
  "{\"type\":\"settlement\",\"amount\":\"$BALANCE\",\"method\":\"transfer\",\"reference\":\"BCA a/n Rina Kartika\"}" >/dev/null
say "pelunasan diterima"

api POST "/orders/$ORDER_RINA_ID/ship" \
  '{"tracking_number":"JNE0081234567","shipping_cost":"19000"}' >/dev/null
say "resi JNE0081234567 tersimpan"

# --- Order Budi dikemas dan ditagih, belum lunas -----------------------------
step "Order Budi dikemas dan ditagih"
# Kardus besar tapi isinya ringan: berat volume (40x30x25)/6000 = 5 kg yang
# ditagih, bukan berat asli 800 g. Ini contoh kasus yang paling sering bikin
# ongkir meleset kalau hanya menimbang.
api POST "/orders/$ORDER_BUDI_ID/pack" \
  '{"courier":"JNE","service":"YES","weight_gram":800,"length_cm":40,"width_cm":30,"height_cm":25}' >/dev/null
api POST "/orders/$ORDER_BUDI_ID/invoices" '{"type":"final"}' >/dev/null
say "menunggu pelunasan"

# --- Biaya perjalanan -------------------------------------------------------
step "Mencatat biaya perjalanan"
add_expense() {
  api POST "/trips/$TRIP/expenses" \
    "{\"category\":\"$1\",\"description\":\"$2\",\"amount\":\"$3\"}" >/dev/null
  say "$2"
}
# Nominalnya sengaja disesuaikan dengan skala tiga order contoh di atas. Pada
# trip sungguhan biaya tetap seperti tiket dan hotel ikut dibebankan, tapi juga
# ditanggung puluhan order, bukan tiga.
add_expense bagasi    "Bagasi tambahan 10kg"       "350000"
add_expense transport "Kereta bandara dan JR line" "150000"

# --- Penjualan stok ---------------------------------------------------------
step "Menjual sebagian stok di marketplace"
STOCK="$(api GET '/stock?per_page=50')"
STOCK_PRODUCT="$(jq -r '.data[] | select(.qty_on_hand > 1) | .product_id' <<<"$STOCK" | head -1)"
if [[ -n "$STOCK_PRODUCT" ]]; then
  api POST /stock/sell \
    "{\"product_id\":\"$STOCK_PRODUCT\",\"qty\":1,\"sale_price\":\"145000\",\"channel\":\"Shopee\"}" >/dev/null
  say "1 unit terjual di Shopee"
fi

printf '\n\033[1;32m✓ Data demo siap\033[0m\n'
printf '  Trip     : %s\n' "$TRIP_CODE"
printf '  Order    : 3 (dikirim, menunggu pelunasan, menunggu DP)\n'
printf '  Channel  : WhatsApp, Instagram, TikTok\n'
printf '  Buka     : %s\n' "${BASE_URL/8080/3000}"
