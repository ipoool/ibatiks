# Ibatiks

Back office pengelolaan jasa titip (jastip) luar negeri — dari trip dibuka sampai laba dihitung.

Aplikasi ini dipakai **internal oleh tim toko**, bukan customer. Customer tetap memesan lewat WhatsApp atau sosial media seperti biasa; admin yang mencatatnya di sini.

```
Trip dibuat (tanggal, negara, kurs)
  └─ Katalog produk + markup harga
      └─ Order masuk dicatat admin (+ channel asalnya) → minta DP
          → DP diverifikasi lewat bukti transfer → order jadi "Diproses"
          └─ Daftar belanja otomatis untuk tripper (hanya order ber-DP)
              └─ Tripper belanja (boleh lebih → sisanya jadi stok)
                  └─ Barang tiba → dicocokkan → dikemas + hitung ongkir
                      └─ Invoice pelunasan → WhatsApp
                          └─ Lunas → kirim JNE + resi → customer dikabari
                              └─ Laporan: omzet, HPP riil, biaya trip, laba,
                                 rekap per customer dan per channel
```

## Isi

- [Teknologi](#teknologi)
- [Menjalankan di komputer sendiri](#menjalankan-di-komputer-sendiri)
- [Struktur project](#struktur-project)
- [Cara kerja beberapa bagian penting](#cara-kerja-beberapa-bagian-penting)
- [Hak akses pengguna](#hak-akses-pengguna)
- [Deploy ke server production](#deploy-ke-server-production)
- [Perawatan](#perawatan)
- [Perintah yang sering dipakai](#perintah-yang-sering-dipakai)

---

## Teknologi

| Bagian | Teknologi |
|---|---|
| Frontend | Next.js 16 (App Router), TypeScript, Tailwind CSS v4, **shadcn/ui**, TanStack Query |
| Backend | Go 1.26, chi, **pgx v5 langsung** (tanpa ORM) |
| Database | PostgreSQL 17 |
| Deployment | Docker Compose + Caddy (TLS otomatis) |

Nominal uang memakai `NUMERIC(18,2)` di database dan `decimal.Decimal` di Go — tidak ada `float64` untuk uang di mana pun, karena selisih satu rupiah pada laporan keuangan akan terlihat. Pada JSON, nominal dikirim sebagai **string** (mis. `"1250000.00"`) karena presisi `NUMERIC` bisa melampaui `number` JavaScript.

---

## Menjalankan di komputer sendiri

Yang dibutuhkan hanya **Docker** — Go dan Node.js tidak perlu dipasang.

```bash
git clone <repo> ibatiks && cd ibatiks
make setup        # menyalin .env.example → .env lalu menyalakan semuanya
make seed-demo    # akun owner + data contoh untuk dicoba
```

Buka:

| Alamat | Isi |
|---|---|
| http://localhost:3000 | Aplikasi |
| http://localhost:8080 | API |
| http://localhost:8081 | Adminer (GUI database) |

Login dengan `SEED_OWNER_EMAIL` dan `SEED_OWNER_PASSWORD` dari `.env` (bawaannya `owner@ibatiks.id` / `rahasia123`).

Backend dan frontend berjalan dengan **hot reload**: ubah kode di editor, hasilnya langsung terlihat tanpa build ulang container.

### Tanpa Docker

```bash
# Backend  (butuh PostgreSQL berjalan)
cd backend && go run ./cmd/migrate up && go run ./cmd/seed --demo && go run ./cmd/api

# Frontend
cd web && npm install && npm run dev
```

---

## Struktur project

```
ibatiks/
├── backend/                     API Go
│   ├── cmd/
│   │   ├── api/                 HTTP server (juga menyediakan `--health`)
│   │   ├── migrate/             up · down · reset · version
│   │   └── seed/                akun owner pertama + data contoh
│   ├── internal/
│   │   ├── config/              pemuat env, divalidasi saat boot
│   │   ├── db/                  pgxpool, helper transaksi, migrasi (embedded)
│   │   ├── domain/              entitas, enum, peta transisi status, aturan harga
│   │   ├── repository/          SQL murni dengan pgx
│   │   ├── service/             aturan bisnis & transaksi
│   │   ├── http/                router, handler, middleware, DTO
│   │   ├── pdf/                 render invoice
│   │   ├── notify/              penyusun pesan WhatsApp
│   │   └── pkg/                 money, docnum, token, validate, pagination
│   ├── Dockerfile               produksi (distroless, ±37 MB)
│   └── Dockerfile.dev           pengembangan (air, hot reload)
│
├── web/                         Frontend Next.js
│   ├── src/app/
│   │   ├── (auth)/login/        halaman masuk
│   │   ├── (dashboard)/         seluruh halaman aplikasi
│   │   └── api/                 BFF: auth cookie + proxy ke backend
│   ├── src/components/          komponen UI bersama
│   ├── src/hooks/               hook TanStack Query per modul
│   ├── src/lib/                 klien API, sesi, format rupiah/tanggal
│   ├── src/middleware.ts        penjaga rute + perbarui token sebelum render
│   ├── Dockerfile               produksi (standalone)
│   └── Dockerfile.dev           pengembangan
│
├── deploy/Caddyfile             reverse proxy + TLS otomatis
├── scripts/smoke.sh             uji alur bisnis end-to-end
├── docker-compose.yml           pengembangan
├── docker-compose.prod.yml      produksi
├── Makefile                     perintah pintas (`make` untuk daftar)
└── .env.example                 seluruh konfigurasi, terdokumentasi
```

---

## Cara kerja beberapa bagian penting

### Harga jual dan markup

Harga modal dalam mata uang asing dikonversi memakai **kurs yang dikunci di level trip**, lalu ditambah markup, lalu dibulatkan ke atas ke kelipatan seratus rupiah.

```
880 JPY × kurs 108,5 = Rp95.480  →  +35%  =  Rp128.898  →  Rp128.900
```

Hasilnya **disimpan**, bukan dihitung ulang saat ditampilkan. Kalau kurs trip diubah setelah katalog diumumkan, harga tidak ikut bergeser diam-diam — admin harus menekan "Hitung Ulang Harga" secara sadar.

### Daftar belanja

Bukan tabel tersendiri, melainkan **agregasi langsung dari pesanan**. Konsekuensinya, begitu admin mengubah qty sebuah order, daftar yang dilihat tripper di lapangan ikut berubah saat itu juga.

### Surplus belanja → stok

Tripper sering membeli lebih banyak dari yang dipesan. Kelebihannya otomatis masuk stok, dan **tidak dibebankan sebagai HPP trip**:

> Uangnya memang sudah keluar (terhitung di "modal keluar"), tapi nilainya masih dipegang sebagai barang. Barang itu baru menjadi HPP ketika terjual di marketplace.

Tanpa aturan ini, trip yang kebetulan borong stok akan terlihat rugi padahal tidak.

### HPP dan laba trip

```
laba kotor  = omzet order − HPP riil barang pesanan
laba bersih = laba kotor − biaya perjalanan (tiket, bagasi, akomodasi, …)
```

HPP diambil dari **biaya belanja yang benar-benar terjadi** (tabel `purchase_allocations`), bukan estimasi saat order dibuat. Jadi kalau harga di toko ternyata naik, laba yang dilaporkan ikut mencerminkannya.

### Mengubah qty order

Operasi paling sering dipakai, sekaligus paling rawan. Penjagaannya:

- Order yang sudah diserahkan ke kurir tidak bisa diubah lagi
- Qty tidak boleh turun di bawah jumlah yang sudah **diterima**
- Kalau qty turun di bawah jumlah yang sudah **dibeli**, kelebihannya dilepas menjadi stok — barangnya nyata, jadi tidak boleh hilang dari pembukuan
- Total, sisa tagihan, dan status order dihitung ulang; kalau order tadinya lunas lalu totalnya naik, statusnya turun kembali agar sisa tagihan baru tidak tersembunyi
- Semua perubahan tercatat di jejak audit

### Invoice DP dan invoice pelunasan

Keduanya memuat nilai pesanan yang sebenarnya — subtotal barang, diskon, ongkir, dan totalnya —
dan hanya berbeda pada apa yang ditagih:

```
Invoice DP                     Invoice pelunasan
  Subtotal      Rp 666.800       Subtotal        Rp 666.800
  Ongkir        Rp  30.000       Ongkir          Rp  30.000
  Total tagihan Rp 696.800       Total tagihan   Rp 696.800
  Down payment  Rp 348.400       Down payment   -Rp 348.400
  Ditagihkan    Rp 348.400       Sisa ditagihkan Rp 348.400
```

Dulu invoice DP menuliskan nilai uang muka sebagai subtotal dan totalnya, sehingga customer
menerima dokumen yang terbaca seolah-olah harga pesanannya hanya sebesar DP.

### Surat jalan

Order punya surat jalan siap cetak (tombol **Surat Jalan** di detail order dan di kedua tab menu
**Pengiriman**). Isinya pengirim, penerima dengan alamat lengkap sampai kelurahan dan
kecamatan, data kurir dan resi, daftar barang, serta kolom tanda tangan serah terima.

Dokumennya tidak disimpan ke disk seperti invoice: isinya seluruhnya berasal dari order, jadi
mencetak ulang selalu menghasilkan lembar yang sesuai keadaan terkini — termasuk ketika dicetak
lebih dulu sebagai pendamping saat mengemas, dengan kolom resi yang masih kosong.

### Status trip dan status order

Status trip cukup dua, dan keduanya menjawab satu pertanyaan yang benar-benar dipakai sehari-hari:

| Status | Artinya |
|---|---|
| **Open** | Order baru masih boleh dicatat untuk trip ini |
| **Closed** | Pendaftaran order ditutup; order yang sudah masuk tetap diproses seperti biasa |

Posisi barangnya — sedang dibelanjakan, dalam perjalanan pulang, sudah sampai — dibaca dari
pembelian dan penerimaan yang tercatat, bukan dari status trip, supaya satu kejadian tidak perlu
dicatat di dua tempat.

Status order mengikuti perjalanan satu pesanan:

| Status | Artinya |
|---|---|
| **Draft** | Masih disusun admin, belum ditagihkan |
| **Menunggu DP** | Sudah dikonfirmasi, menunggu uang muka masuk |
| **Diproses** | DP diterima. Belanja, penerimaan, dan pencocokan barang terjadi di tahap ini |
| **Sedang Dikemas** | Sudah dikemas atas nama customer |
| **Penagihan** | Invoice pelunasan sudah dikirim |
| **Pembayaran Lunas** | Lunas, siap diserahkan ke kurir |
| **Dikirim** | Sudah diserahkan ke kurir dan resinya terisi |
| **Selesai** | Diterima customer |
| **Batal** | Dibatalkan |

Order yang sudah lunas tapi barangnya belum siap **tidak** dilompatkan ke Pembayaran Lunas —
kalau dilompatkan, ia muncul di antrean siap kirim padahal barangnya belum ada. Yang muncul
adalah penanda "Lunas" di sebelah statusnya.

### Daftar belanja hanya menghitung order ber-DP

Yang muncul sebagai "harus dibeli" hanyalah order berstatus `dp_paid` ke atas.
Permintaan yang DP-nya belum masuk dihitung terpisah pada kolom `qty_awaiting_dp`.

Alasannya soal arus kas, bukan kerapian: membelanjakan order yang uang mukanya
belum ada berarti menalangi pembelian dengan uang toko, dan kalau customer itu
batal, barangnya mengendap di stok tanpa ada yang membayar. Kolomnya tetap
ditampilkan supaya tripper tahu ada permintaan yang tertahan dan bisa meminta
admin menagih DP-nya lebih dulu.

### Riwayat harga produk antar trip

`GET /products/{id}/price-history` menggabungkan harga katalog (`trip_items`) dan
harga beli riil (`purchases`) per trip, sehingga produk yang sama bisa dipakai
ulang tanpa menggali harga dari catatan trip lama.

Harga tidak pernah disalin otomatis antar trip dengan mata uang berbeda: produk
yang sama dibeli di Korea, bukan Jepang, bukan angka yang sebanding. Saat mata
uangnya berbeda, kolom harga modal justru dibiarkan kosong agar admin sadar harus
mengisinya.

### Kurs otomatis, tapi tetap dikunci

`GET /exchange-rate?from=JPY` mengambil kurs harian dari layanan publik (tanpa API
key) dan menyimpannya di cache satu jam. Nilainya hanya dipakai mengisi kolom kurs
saat trip dibuat; setelah tersimpan, kurs trip tidak pernah bergerak lagi supaya
laba trip yang sudah selesai tidak berubah karena kurs hari ini.

Kalau layanannya tidak bisa dihubungi, permintaan gagal dengan pesan yang
mengarahkan admin mengisi kursnya manual — form-nya tidak pernah terkunci karena
layanan luar.

### Komponen antarmuka

Seluruh komponen memakai **shadcn/ui** gaya `new-york` (berbasis Radix). Konfigurasinya
ada di `web/components.json`; menambah komponen baru cukup `npx shadcn@latest add <nama>`.

Berkas di `web/src/components/ui/` adalah salinan milik proyek, bukan dependensi —
memang dirancang untuk diedit. Yang sudah diperluas dari bawaannya:

| Komponen | Perluasan | Alasan |
|---|---|---|
| `button` | prop `loading`, varian `success` | Hampir setiap tombol memicu panggilan jaringan; tanpa penanda bawaan, tiap pemanggil menyusun spinner sendiri |
| `badge` | varian `neutral`/`info`/`progress`/`success`/`warning`/`danger` | Lencana status perlu latar pucat, bukan blok warna penuh yang mengalahkan angka di sebelahnya |

Token warnanya tetap palet Ibatiks (navy), bukan abu-abu bawaan shadcn, ditambah
`success` dan `warning` yang tidak ada di palet standar.

Komponen aplikasi yang dibangun di atas primitif itu berada di luar `ui/`:
`data-table.tsx` (tabel dengan keadaan memuat dan kosong), `filter-select.tsx`
(`FilterSelect` untuk saringan berpilihan "semua", `OptionSelect` untuk isian form),
dan `status-badge.tsx` (kosakata status jastip).

`FilterSelect` menjembatani satu ketidakcocokan yang mudah terlewat: Radix melarang
`SelectItem` bernilai string kosong, sementara API memakai string kosong untuk
"tanpa filter". Sentinel-nya ditangani di satu tempat supaya tidak ada halaman yang
diam-diam mengirim nilai sentinel ke server.

### Perkiraan ongkir

Ekspedisi menagih mana yang lebih besar antara berat asli dan **berat volume**, yaitu `(P × L × T dalam cm) ÷ pembagi` — pembagi 6000 untuk JNE. Hasilnya dibulatkan **ke atas** ke kilogram penuh, karena kurir tidak menjual pecahan kilo. Kardus 40 × 30 × 25 cm berisi 800 gram tetap ditagih 5 kg.

Angkanya diambil dari dua sumber, berurutan.

**RajaOngkir** dipakai lebih dulu kalau `RAJAONGKIR_API_KEY` terisi. Layanannya mengembalikan ongkos **utuh** untuk berat paket, bukan tarif per kilogram — kurir menagih berjenjang, jadi memaksanya jadi angka per kilo akan meleset pada berat yang bukan kelipatan bulat. Kota asal dan daftar kurir dipilih di menu Pengaturan → Ongkir dan tersimpan di `app_settings` (`shipping_origin_id`, `shipping_origin_label`, `shipping_couriers`; pemisah kurirnya titik dua, `jne:jnt:sicepat`, karena itu bentuk yang diminta API-nya). Alamat order diterjemahkan ke ID tujuan RajaOngkir dari yang paling spesifik ke yang paling umum — kode pos, lalu kelurahan + kota, kecamatan + kota, baru kotanya saja — dan pemetaan yang ketemu disimpan di `shipping_destinations` supaya alamat yang sama tidak memakan kuota dua kali.

**Tabel tarif sendiri** (`shipping_rates`, juga dikelola di menu yang sama) adalah cadangannya: dipakai kalau key-nya kosong, kota asalnya belum dipilih, atau layanannya sedang tidak bisa dihubungi. Kegagalan RajaOngkir sengaja tidak menghentikan perhitungan — admin sedang menimbang paket di depan customer, dan lebih baik menerima angka dari tabel dengan penanda asalnya daripada layar galat. Kota tujuan dicocokkan setelah dinormalisasi (`Kota Bandung`, `BANDUNG`, dan `bandung` sama), dan kalau tidak ketemu dipakai tarif cadangan dari `app_settings`.

Sumber yang benar-benar dipakai selalu disebutkan pada hasil hitungnya (`source`), berikut tujuan seperti dikenali kurir (`destination`) — nama kota yang sama bisa menunjuk kecamatan berbeda dengan tarif berbeda, jadi admin perlu bisa memeriksanya.

Keduanya berada di balik interface: `service.RateProvider` untuk tarif per kilogram dan `service.CostProvider` untuk ongkos utuh. Mengganti vendor berarti menambah satu tipe yang memenuhi salah satunya, tanpa menyentuh handler maupun frontend.

### Invoice dan WhatsApp

Invoice dirender menjadi PDF dengan seluruh nominal **disalin saat diterbitkan**, sehingga dokumen yang sudah dilihat customer tidak berubah kalau order diedit setelahnya.

Pesan WhatsApp disiapkan sistem (teks lengkap + tautan `wa.me`), tapi **dikirim sendiri oleh admin** dari nomor toko. Tidak ada gateway berbayar, dan customer menerima pesan dari kontak yang sudah dikenalnya. Template pesannya bisa diubah di menu Pengaturan.

### Autentikasi

Token JWT **tidak pernah menyentuh `localStorage`**. Route handler Next.js menyimpannya di cookie `httpOnly`, dan seluruh panggilan browser lewat proxy BFF (`/api/proxy/*`) yang menyisipkan header `Authorization` serta memperbarui token kedaluwarsa secara otomatis.

Refresh token dirotasi setiap kali dipakai dan disimpan sebagai hash — kalau isi tabelnya bocor, token mentahnya tetap tidak bisa dipakai login.

---

## Hak akses pengguna

| Role | Bisa apa |
|---|---|
| **owner** | Semuanya, termasuk laporan laba, pengaturan toko, dan manajemen pengguna |
| **admin** | Seluruh operasional harian: trip, order, invoice, pengiriman, stok, piutang, rekap customer dan channel |
| **tripper** | Trip dan katalog (baca), daftar belanja, dan input pembelian di lapangan |

Beri tripper akun sendiri agar bisa mencatat belanja langsung dari ponsel saat di toko.

### Mempersempit per pengguna

Role menentukan batas kasarnya. Di dalam batas itu, owner bisa mencentang menu apa saja yang boleh
dibuka seorang pengguna lewat **Pengaturan → Pengguna** — misalnya admin yang mengurus order tapi
tidak perlu melihat laporan laba.

Yang perlu diketahui:

- Tidak mencentang apa pun berarti mengikuti bawaan role, bukan mengunci semuanya.
- Centang hanya bisa mempersempit. Tripper tetap tidak bisa diberi menu pengaturan, karena batas
  role adalah keputusan keamanan sedangkan centang cuma penyempitan.
- Menu yang dimatikan benar-benar tertutup: backend menolak endpointnya, bukan sekadar
  menyembunyikan menunya.
- Begitu hak akses seseorang diubah, sesinya dicabut dan ia diminta login ulang supaya
  pembatasannya berlaku saat itu juga.

---

## Deploy ke server production

Yang dibutuhkan: server Linux dengan Docker, dan domain yang sudah diarahkan ke IP server.

### 1. Siapkan konfigurasi

```bash
git clone <repo> ibatiks && cd ibatiks
cp .env.example .env
```

Isi `.env` — **wajib diganti semua**:

```bash
APP_ENV=production
APP_DOMAIN=jastip.tokokamu.com        # domain yang sudah diarahkan ke server ini

POSTGRES_USER=jastipin
POSTGRES_PASSWORD=<password kuat>     # openssl rand -base64 24
POSTGRES_DB=jastipin

JWT_SECRET=<minimal 32 karakter>      # openssl rand -base64 48
                                      # backend menolak start kalau terlalu pendek

SEED_OWNER_NAME=Nama Kamu
SEED_OWNER_EMAIL=kamu@tokokamu.com
SEED_OWNER_PASSWORD=<password kuat>
```

### 2. Nyalakan

```bash
make prod-build     # atau: docker compose -f docker-compose.prod.yml build
make prod-up
make prod-seed      # buat akun owner pertama
```

Migrasi database berjalan otomatis sebagai job tersendiri sebelum backend menerima trafik. Caddy mengurus sertifikat TLS Let's Encrypt sendiri — tidak ada langkah manual.

Buka `https://<APP_DOMAIN>` dan login.

### 3. Memperbarui versi

```bash
git pull
make prod-build && make prod-up
```

Migrasi baru diterapkan otomatis saat container start.

### Catatan keamanan production

- PostgreSQL **tidak** mem-publish port ke host; hanya service di jaringan internal compose yang bisa menghubunginya
- Backend berjalan sebagai pengguna non-root di image distroless tanpa shell
- Frontend berjalan sebagai pengguna non-root
- Caddy memasang HSTS, `X-Frame-Options: DENY`, dan `nosniff`
- Aplikasi ditandai `noindex` dan `robots.txt` melarang seluruh crawler

### Memakai registry image

Kalau ingin membangun image di CI lalu menariknya di server:

```bash
# di mesin build
IMAGE_REGISTRY=ghcr.io/username IMAGE_TAG=v1.0.0 make prod-build push

# di server
IMAGE_REGISTRY=ghcr.io/username IMAGE_TAG=v1.0.0 docker compose -f docker-compose.prod.yml up -d
```

---

## Perawatan

### Backup database

```bash
make backup      # tersimpan ke ./backups/jastipin-<tanggal>.sql.gz
```

Jadwalkan lewat cron agar berjalan otomatis:

```cron
0 2 * * * cd /srv/ibatiks && make backup
```

### Restore

```bash
gunzip -c backups/jastipin-20260815-020000.sql.gz | \
  docker compose -f docker-compose.prod.yml exec -T db psql -U jastipin -d jastipin
```

### Berkas unggahan

Bukti transfer, foto struk, dan PDF invoice tersimpan di volume `uploads`. Ikut sertakan dalam rencana backup:

```bash
docker run --rm -v jastipin-prod_uploads:/data -v $(pwd)/backups:/backup \
  alpine tar czf /backup/uploads-$(date +%Y%m%d).tar.gz -C /data .
```

---

## Perintah yang sering dipakai

```bash
make              # daftar semua perintah
make up           # nyalakan lingkungan pengembangan
make logs-be      # ikuti log backend
make shell-db     # masuk ke psql
make test         # unit test backend
make lint         # gofmt + go vet + tsc + eslint
make smoke        # uji alur bisnis end-to-end lewat API
make reset        # hapus semuanya termasuk data
```

### Uji end-to-end

`scripts/smoke.sh` menelusuri satu siklus jastip lengkap lewat API — trip dibuka, order masuk, DP dibayar, tripper belanja lebih banyak dari pesanan, qty diedit, barang diterima, dikemas, ditagih, dilunasi, dikirim — lalu **mencocokkan angka laba trip dengan hitungan manual**, termasuk memastikan surplus stok tidak salah dibebankan sebagai HPP.

```bash
make smoke                                  # terhadap lingkungan pengembangan
./scripts/smoke.sh https://jastip.domain.com  # terhadap server lain
```

---

## Lisensi

Proyek privat. Seluruh hak dimiliki pemilik repositori.
