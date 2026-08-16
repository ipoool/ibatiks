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

## Aturan yang tidak boleh dilanggar

**Uang.** `NUMERIC(18,2)` di database, `decimal.Decimal` di Go, **string** di JSON. Tidak ada `float64` untuk uang di mana pun. Di frontend, nominal dari API adalah string — lewatkan `toNumber()` sebelum berhitung.

**Data layer.** pgx v5 langsung, tanpa ORM. Query ditulis di `internal/repository`, dipetakan dengan `pgx.CollectRows` + `RowToStructByName`/`Lax` dan tag `db`. Operasi multi-tabel dibungkus `db.WithTx`.

**Aturan bisnis tinggal di service**, bukan di handler maupun di UI. Peta transisi status order ada di satu tempat: `internal/domain`.

**Snapshot historis.** `order_items` menyimpan salinan nama dan harga produk; `orders` menyimpan salinan alamat kirim. Mengedit master data tidak boleh mengubah dokumen lama.

**Migrasi bersifat tambahan.** Jangan menyunting file migrasi yang sudah pernah dijalankan — buat migrasi baru. Setiap `.up.sql` wajib punya `.down.sql` yang benar-benar membalik; uji dengan `make migrate-down && make migrate-up`.

## Frontend

**shadcn/ui gaya `new-york` (berbasis Radix).** Jangan memasang komponen gaya `base-nova` — mencampur dua pustaka primitif akan mengacaukan token dan perilaku fokus. Komponen dasar ada di `src/components/ui/`.

**Jangan menyetel state di dalam efek** — aturan lint `react-hooks/set-state-in-effect` menolaknya. State turunan dihitung langsung saat render (lihat `nav-menu.tsx`: pilihan pengguna disimpan bersama `pathname` supaya kedaluwarsa sendiri saat pindah halaman).

**Select tidak boleh punya item bernilai string kosong** (batasan Radix). Pakai `FilterSelect` yang sudah menangani sentinel `__all__`, atau `OptionSelect` bila memang tanpa opsi "semua".

**Input uang memakai `step="any"`.** `step="1000"` membuat nominal seperti Rp348.400 ditolak browser — termasuk DP 50% yang dihitung sistem sendiri dan transfer berkode unik.

**Tabel harus menyusut, bukan menggeser halaman.** Kartu dan kolom grid yang memuat tabel wajib `min-w-0` — anak grid/flex bawaannya tidak boleh lebih sempit dari isinya, sehingga tabel lebar mendorong seluruh halaman ke samping dan `overflow-x-auto` di dalamnya tidak pernah terpakai. Kolom sekunder disembunyikan bertahap (`hidden sm:table-cell` lalu `hidden lg:table-cell`/`xl:table-cell`), dan isian yang ikut hilang dilipat ke kolom utama sebagai baris kecil `sm:hidden`. Sel bawaan shadcn memakai `whitespace-nowrap`; kolom berisi nama panjang perlu `whitespace-normal` supaya boleh turun baris.

**Dropdown yang isinya bisa panjang** (customer, katalog produk) memakai `Combobox`, dengan `keywords` untuk hal yang dicari orang tapi bukan judul: nomor WA, SKU.

**`src/lib/utils.ts` adalah pusat helper format** (`cn`, `formatIDR`, `formatDate`, `toNumber`, …) dan dipakai puluhan berkas. `shadcn init` pernah menimpanya — jangan jalankan ulang init; tambahkan komponen satu per satu dengan `shadcn add`.

**Middleware harus berada di `src/middleware.ts`.** Proyek ini memakai direktori `src/`; berkas `middleware.ts` di akar `web/` diabaikan Next tanpa peringatan apa pun. Pastikan `ƒ Proxy (Middleware)` muncul di keluaran `next build`.

**Konstanta yang dibaca komponen server tidak boleh diekspor dari modul `"use client"`.** Nilai yang diimpor dari modul klien berubah jadi rujukan modul, dan pemakaiannya gagal diam-diam (`cookies().get(konstanta)` mengembalikan undefined). Simpan di modul biasa seperti `src/lib/sidebar.ts`.

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
