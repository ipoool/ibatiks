# Ibatiks — panduan kerja di repo ini

Back office jasa titip (jastip): Next.js 16 + Go 1.26 + PostgreSQL 17, dijalankan lewat Docker Compose.
Pengguna aplikasi ini adalah **tim toko**, bukan customer. Bahasa antarmuka, komentar kode, dan pesan commit: **Bahasa Indonesia**.

Alur bisnis dan penjelasan arsitektur ada di `README.md` — baca itu dulu sebelum mengubah aturan bisnis.

## Perintah

```bash
make up | down | restart | logs        # stack pengembangan
make lint                              # gofmt + go vet + tsc + eslint
make test                              # go test ./...
make build-web                         # next build
make smoke                             # uji alur bisnis end-to-end (16 langkah)
make migrate-up | migrate-down | migrate-reset | migrate-version
make seed-demo && ./scripts/demo-data.sh   # data contoh yang bisa dilihat di UI
```

Login lokal: `owner@ibatiks.id` / `rahasia123` (dari `SEED_OWNER_*` di `.env`).

`scripts/smoke.sh` **merusak data demo** karena membuat trip dan order sendiri. Setelah menjalankannya, pulihkan dengan `make migrate-reset && make migrate-up && make seed-demo && ./scripts/demo-data.sh`.

**Backend wajib direstart setelah `make migrate-reset`.** Skema dibuat ulang, tapi cache tipe pgx di kolam koneksi masih memegang OID lama — gejalanya seluruh permintaan membalas 500 dengan `cache lookup failed for type NNNNN`. `docker compose restart backend` lalu tunggu sekitar sepuluh detik.

**Smoke test menuntut database yang benar-benar bersih.** Ia menghitung trip aktif dan mengandaikan miliknya satu-satunya, dan customernya memakai nomor WA tetap yang bentrok dengan sisa jalannya sendiri. Jalankan siklus penuh — reset, migrate, restart, seed, demo-data — baru `make smoke`. `demo-data.sh` menutup trip demo, dan itu memang yang diandaikan smoke.

## Aturan yang tidak boleh dilanggar

**Uang.** `NUMERIC(18,2)` di database, `decimal.Decimal` di Go, **string** di JSON. Tidak ada `float64` untuk uang di mana pun. Di frontend, nominal dari API adalah string — lewatkan `toNumber()` sebelum berhitung.

**Data layer.** pgx v5 langsung, tanpa ORM. Query ditulis di `internal/repository`, dipetakan dengan `pgx.CollectRows` + `RowToStructByName`/`Lax` dan tag `db`. Operasi multi-tabel dibungkus `db.WithTx`.

**Aturan bisnis tinggal di service**, bukan di handler maupun di UI. Peta transisi status order dan trip ada di satu tempat: `internal/domain`.

**Status sengaja sedikit.** Trip hanya `open`/`closed`; order hanya sembilan: `draft`, `awaiting_dp`, `dp_paid`, `packed`, `invoiced`, `paid`, `shipped`, `completed`, `cancelled`. Belanja dan penerimaan barang terjadi *di dalam* tahap `dp_paid` — kemajuannya dibaca dari data pembelian dan penerimaan per item, bukan dari status order. Jangan menambah status baru untuk menandai kejadian yang datanya sudah tercatat di tempat lain.

**Invoice memuat nilai order seutuhnya.** Baik invoice DP maupun pelunasan menuliskan subtotal, diskon, ongkir, dan total order yang sebenarnya; yang membedakan hanya apa yang ditagih (`dp_amount` untuk DP, sisa tagihan untuk pelunasan). Jangan kembali menyimpan nilai DP sebagai `total` — customer akan menerima dokumen yang seolah-olah menyatakan pesanannya cuma seharga uang muka.

**Barang yang tidak dapat tidak ikut ditagih.** Begitu sebuah item dicocokkan sebagai `unavailable`, `partial`, atau `refunded`, yang ditagihkan adalah `qty_received × unit_price` — bukan jumlah yang dulu dipesan. Aturannya ada di `RecalculateTotals`, dan `ReceiveItems` wajib memanggilnya. Kolom `qty` sengaja dibiarkan apa adanya supaya tetap terbaca apa yang dipesan semula; yang berubah hanya apa yang ditagih. Tanpa ini invoice pelunasan memuat barang yang tidak akan pernah dikirim.

**Snapshot historis.** `order_items` menyimpan salinan nama dan harga produk; `orders` menyimpan salinan alamat kirim. Mengedit master data tidak boleh mengubah dokumen lama.

**Dokumen historis tetap bisa dibuka walau master datanya dihapus.** Detail order memakai `GetByIDIncludingDeleted` untuk customernya: menghapus customer tidak menghapus ordernya, jadi menyaring `deleted_at IS NULL` di jalur itu membuat seluruh order lamanya mati dengan "customer tidak ditemukan" — padahal ordernya masih ada dan bisa jadi masih punya sisa tagihan. Daftar dan penyuntingan tetap menyaring yang sudah dihapus.

**Pencarian customer menormalkan nomor.** Nomor disimpan sebagai `62812…`, sementara admin mengetiknya seperti yang tertera di WhatsApp customer (`0812…`). Kata kunci yang berbentuk nomor ikut dinormalkan sebelum dicocokkan; `domain.LooksLikePhone` sengaja ketat supaya kata kunci berhuruf tetap diperlakukan sebagai pencarian nama dan hasilnya tidak melebar.

**Migrasi bersifat tambahan.** Jangan menyunting file migrasi yang sudah pernah dijalankan — buat migrasi baru. Setiap `.up.sql` wajib punya `.down.sql` yang benar-benar membalik; uji dengan `make migrate-down && make migrate-up`.

## Ongkir

Ada dua sumber tarif dan urutannya tidak boleh dibalik: **RajaOngkir dulu, tabel tarif sendiri sebagai cadangan.** Keduanya di balik interface — `service.RateProvider` (tarif per kg) dan `service.CostProvider` (ongkos utuh). Menambah vendor berarti menambah satu tipe, bukan menyentuh handler atau UI.

- **Kegagalan RajaOngkir tidak boleh menghentikan perhitungan.** `Estimate` menelan galatnya dan tetap jatuh ke tabel tarif. Admin sedang menimbang paket di depan customer; angka dari tabel dengan penanda asalnya jauh lebih berguna daripada layar galat. Yang wajib ada: `source` selalu ikut di hasil, supaya orang tahu angkanya datang dari mana.
- **Kurir menjual ongkos utuh, bukan harga per kilo.** Jangan mengubah balasan RajaOngkir jadi tarif per kg lalu dikalikan — tarifnya berjenjang dan hasilnya meleset pada berat yang bukan kelipatan bulat. `price_per_kg` pada hasil RajaOngkir cuma turunan untuk ditampilkan.
- **`shipping_couriers` dipisah titik dua** (`jne:jnt:sicepat`) karena itu bentuk yang diminta API-nya. Simpan apa adanya, jangan diterjemahkan bolak-balik.
- **Pemetaan alamat ke ID tujuan disimpan** di `shipping_destinations`. Kuota langganan terbatas dan pemetaan kota ke ID hampir tidak pernah berubah; mencari ulang alamat yang sama membuang kuota. Urutan percobaannya dari yang paling spesifik: kode pos → kelurahan + kota → kecamatan + kota → kota.
- **API key tidak pernah menyentuh browser.** `RAJAONGKIR_API_KEY` tinggal di server; menu Pengaturan hanya menyimpan ID kota asal, labelnya, dan daftar kurir.
- **Pesan penolakan dari kurir diteruskan apa adanya** ke pencarian kota asal. "API key tidak valid" hanya bisa dibereskan tim toko, dan "terjadi kesalahan pada server" tidak memberi tahu apa pun.

## Frontend

**shadcn/ui gaya `new-york` (berbasis Radix).** Jangan memasang komponen gaya `base-nova` — mencampur dua pustaka primitif akan mengacaukan token dan perilaku fokus. Komponen dasar ada di `src/components/ui/`.

**Jangan menyetel state di dalam efek** — aturan lint `react-hooks/set-state-in-effect` menolaknya. State turunan dihitung langsung saat render (lihat `nav-menu.tsx`: pilihan pengguna disimpan bersama `pathname` supaya kedaluwarsa sendiri saat pindah halaman).

**Select tidak boleh punya item bernilai string kosong** (batasan Radix). Pakai `FilterSelect` yang sudah menangani sentinel `__all__`, atau `OptionSelect` bila memang tanpa opsi "semua".

**Input uang memakai `step="any"`.** `step="1000"` membuat nominal seperti Rp348.400 ditolak browser — termasuk DP 50% yang dihitung sistem sendiri dan transfer berkode unik.

**Tabel harus menyusut, bukan menggeser halaman.** Kartu dan kolom grid yang memuat tabel wajib `min-w-0` — anak grid/flex bawaannya tidak boleh lebih sempit dari isinya, sehingga tabel lebar mendorong seluruh halaman ke samping dan `overflow-x-auto` di dalamnya tidak pernah terpakai. Kolom sekunder disembunyikan bertahap (`hidden sm:table-cell` lalu `hidden lg:table-cell`/`xl:table-cell`), dan isian yang ikut hilang dilipat ke kolom utama sebagai baris kecil `sm:hidden`. Sel bawaan shadcn memakai `whitespace-nowrap`; kolom berisi nama panjang perlu `whitespace-normal` supaya boleh turun baris.

**Dropdown yang isinya bisa panjang** (customer, katalog produk) memakai `Combobox`, dengan `keywords` untuk hal yang dicari orang tapi bukan judul: nomor WA, SKU.

**`src/lib/utils.ts` adalah pusat helper format** (`cn`, `formatIDR`, `formatDate`, `toNumber`, …) dan dipakai puluhan berkas. `shadcn init` pernah menimpanya — jangan jalankan ulang init; tambahkan komponen satu per satu dengan `shadcn add`.

**Middleware harus berada di `src/middleware.ts`.** Proyek ini memakai direktori `src/`; berkas `middleware.ts` di akar `web/` diabaikan Next tanpa peringatan apa pun. Pastikan `ƒ Proxy (Middleware)` muncul di keluaran `next build`.

**Konstanta yang dibaca komponen server tidak boleh diekspor dari modul `"use client"`.** Nilai yang diimpor dari modul klien berubah jadi rujukan modul, dan pemakaiannya gagal diam-diam (`cookies().get(konstanta)` mengembalikan undefined). Simpan di modul biasa seperti `src/lib/sidebar.ts` dan `src/lib/route-permissions.ts`.

**Penjagaan kiriman ganda memakai `useRef`, bukan `isPending`.** Baik atribut `disabled` maupun `mutation.isPending` baru berubah setelah React merender ulang, jadi dua kiriman di tick yang sama — klik ganda, tombol Enter yang ditahan — sama-sama membaca nilai lama dan lolos berdua. Di form catat order itu berarti dua order kembar untuk customer yang sama. Ref berubah saat itu juga; jangan lupa melepasnya kembali di `onError` supaya admin tetap bisa mencoba ulang.

**Dialog tidak boleh menjalankan submit form halaman di baliknya.** Radix memindahkan isi dialog ke ujung body, tapi React merambatkan event lewat pohon komponen, bukan pohon DOM. `FormDialog` sudah menghentikan perambatan submit-nya; jangan menghapus baris itu. Tanpa itu, menekan Simpan di Tambah Customer pada halaman catat order ikut membuat ordernya.

**Tombol simpan pada dialog berisi `Combobox` wajib memakai `submitDisabled`.** Radix Combobox bukan `<select>` bawaan, jadi validasi bawaan browser tidak melihatnya sama sekali: tombolnya tetap bisa ditekan selagi kosong, dan gelembung yang muncul justru menunjuk kolom lain yang kebetulan berupa input biasa.

**Pesan validasi bawaan browser disetel sendiri.** Bahasanya mengikuti bahasa antarmuka browser, bukan `lang` halaman, jadi `<html lang="id">` tidak menolong. `Input` dan `Textarea` memasang pesan Indonesia lewat `src/lib/validasi-bawaan.ts` dan membuangnya lagi begitu isinya berubah — pembuangan itu wajib, sebab `customValidity` yang tertinggal membuat kolomnya dianggap tidak sah selamanya dan formulirnya tidak akan pernah mau dikirim.

**Bar tab tidak menggulir.** `overflow-x-auto` memaksa `overflow-y` ikut jadi `auto`, dan tinggi yang terkunci membuat isi yang satu piksel lebih tinggi memunculkan scrollbar tegak kecil di ujung kanan. Tab membungkus ke baris berikutnya kalau tidak muat, dan tingginya mengikuti isi.

## Hak akses

Role (`owner`/`admin`/`tripper`) menentukan batas kasar; di dalamnya owner bisa mencentang menu per pengguna lewat **Pengaturan → Pengguna**. Aturannya hanya ditulis sekali, di `internal/domain/permission.go`:

- Daftar kosong di kolom `users.permissions` berarti "ikut bawaan role" — bukan "tanpa akses".
- Centang hanya bisa **mempersempit**; backend menyaring ulang permintaan supaya tripper tidak bisa diberi menu pengaturan.
- Hak akses ikut dibawa di dalam access token, jadi mengubahnya mencabut sesi pengguna itu supaya pembatasannya berlaku saat itu juga.
- Frontend memakai `effective_permissions` yang dihitung backend; jangan menyalin tabel bawaan role ke UI selain untuk menampilkan pilihan centang.
- **Owner tidak bisa mencabut Pengaturan dan Pengguna dari dirinya sendiri** (`OwnerLockedPermissions`). Dua menu itu satu-satunya jalan mengembalikan hak akses siapa pun; sekali hilang, satu-satunya pemulihan adalah `UPDATE users SET permissions = NULL` langsung ke database. Karena dihitung dan bukan disimpan, baris yang terlanjur rusak ikut pulih sendiri.

Menu yang tidak dimiliki pengguna juga **dijaga di tingkat rute**, bukan cuma disembunyikan dari sidebar. Petanya di `src/lib/route-permissions.ts` — modul biasa supaya bisa dibaca middleware tanpa menyeret ikon menu ke bundel edge — dan dipakai bersama oleh sidebar dan `src/middleware.ts`, jadi menu yang disembunyikan dan halaman yang ditolak tidak pernah berbeda pendapat. Ini bukan lapisan keamanan; backend tetap menolak endpoint-nya sendiri. Yang dijaga adalah supaya orang tidak mendarat di halaman yang datanya gagal dimuat lalu membaca "Belum ada customer" seolah tokonya kosong.

## Sesi dan autentikasi

Token tidak pernah menyentuh `localStorage`. Alurnya:

- `src/app/api/auth/*` menaruh access + refresh token di cookie httpOnly.
- `src/app/api/proxy/[...path]` menyisipkan `Authorization` untuk panggilan dari browser dan memperbarui token saat 401.
- `src/middleware.ts` memperbarui token **sebelum halaman dirender di server** — render halaman tidak lewat proxy, jadi tanpa ini sesi putus tiap 15 menit.
- Layout dashboard mengarahkan ke `/login?expired=1` bila backend menolak sesi. Penanda itu wajib: tanpanya middleware melihat cookie yang tampak sah dan melempar balik, dan halaman terjebak putaran pengalihan.

Kalau menyentuh salah satu bagian ini, uji tiga keadaan: token kedaluwarsa (masih bisa lanjut), refresh token ditolak (mendarat di form login, cookie terhapus), dan tanpa cookie sama sekali (dialihkan dengan `?next=`).

## Merek

Nama aplikasi **Ibatiks**. Logo: `web/public/logo-ibatiks.png` (penuh, untuk login), `web/public/logo-ibatiks-mark.png` (guratan "iba." saja, untuk sidebar dan ukuran kecil), `web/src/app/icon.png` (favicon), `backend/internal/pdf/assets/logo-ibatiks.png` (kop invoice, ditanam lewat `go:embed`). Pakai komponen `Logo`, jangan menaruh `<img>` sendiri.

Identitas toko yang tampil di invoice dan pesan WhatsApp diambil dari tabel `app_settings`, bukan dari konstanta — nilai di kode hanya cadangan bila settingnya kosong. Identifier infrastruktur (nama database, pengguna Postgres, path modul Go, nama cookie) sengaja masih memakai `jastipin`; mengubahnya berarti migrasi volume dan tidak terlihat pengguna.

## Sebelum menyatakan selesai

1. `make lint` dan `make test` bersih
2. `make build-web` sukses
3. `make smoke` lulus, lalu data demo dipulihkan
4. Perubahan UI dilihat sendiri di browser — termasuk keadaan kosong, keadaan memuat, dan layar sempit
5. Migrasi baru diuji turun-naik

Laporkan apa adanya: kalau ada langkah yang dilewati, sebutkan.
