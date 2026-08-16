# Catatan Kesesuaian PRD

Hasil penelusuran PRD *Aplikasi Manajemen Jastip* (15 Agustus 2026) terhadap
aplikasi yang sudah dibangun. Dokumen ini mencatat apa yang sudah sesuai, apa
yang baru disesuaikan, dan apa yang sengaja belum dikerjakan beserta alasannya.

---

## 1. Proses bisnis (PRD bab 5)

| # | Langkah PRD | Status |
|---|---|---|
| 1–2 | Order via WA/IG dicatat admin: nama, no HP, alamat, **order source**, multi-item | Sesuai |
| 3 | Status otomatis **Menunggu DP** | Sesuai |
| 4–5 | Buyer transfer DP + kirim bukti → admin verifikasi → **Diproses** | **Disesuaikan** |
| 6 | Dashboard owner: agregat barang yang harus dibeli per trip | **Disesuaikan** |
| 7 | Owner beli sesuai jumlah order (tanpa partial fulfillment) | Melebihi PRD, lihat bagian 3 |
| 8–9 | Rekap pembelian → update master data product, histori harga tersimpan | **Disesuaikan** |
| 10 | Trip ditutup setelah pembelian selesai | Sesuai |
| 11–12 | Packing per order, input ongkir berdasarkan berat/volume | Sesuai |
| 13 | Estimasi ongkir otomatis via integrasi JNE | **Belum**, lihat bagian 4 |
| 14–17 | Pelunasan manual + bukti transfer → input resi | Sesuai |
| 18 | Notifikasi WA satu arah saat resi terbit | **Sebagian**, lihat bagian 4 |
| 19–21 | Rekap per trip, per customer, per channel | Sesuai |
| 22 | List order menampilkan Rupiah hasil konversi otomatis | Sesuai |

---

## 2. Yang baru disesuaikan agar cocok PRD

### Daftar belanja hanya menghitung order "Diproses"

PRD 5.2 poin 6: agregat barang yang harus dibeli diambil dari order **berstatus
Diproses**. Sebelumnya sistem menghitung semua order kecuali batal dan draft —
termasuk yang DP-nya belum masuk.

Ini bukan sekadar beda angka. Membelanjakan order yang uang mukanya belum ada
berarti menalangi pembelian dengan uang toko; kalau customer itu batal, barangnya
mengendap di stok tanpa ada yang membayar, sementara uangnya sudah keluar di luar
negeri.

Sekarang daftar belanja punya dua kolom terpisah: **Dipesan** (yang boleh
dibelanjakan) dan **Menunggu DP** (yang belum). Kolom kedua tetap ditampilkan
supaya tripper melihat permintaan yang tertahan dan bisa meminta admin menagih
DP-nya lebih dulu, bukan supaya angkanya disembunyikan.

### Istilah status "Diproses"

PRD memakai istilah *Diproses* untuk order yang DP-nya sudah diverifikasi (5.1
poin 5 dan 5.2 poin 6). Label yang sebelumnya tertulis "DP Masuk" diubah menjadi
**Diproses**. Nilai internal di basis data tetap `dp_paid`.

### Nomor WhatsApp sebagai identitas customer

PRD menetapkan no HP buyer sebagai *unique identifier* untuk laporan per
customer. Kolomnya sebelumnya tidak dijamin unik, sehingga satu orang yang
terlanjur dicatat dua kali membuat belanjanya terpecah pada laporan.

Ditambahkan indeks unik parsial pada `customers.phone_wa` untuk baris yang belum
dihapus. Nomor dinormalisasi lebih dulu, jadi `0812 3456 7890` dan
`6281234567890` dikenali sebagai nomor yang sama.

### Upload bukti transfer

PRD memasukkan *upload/verifikasi bukti transfer* ke dalam scope. Endpoint
`POST /uploads` sudah ada di backend sejak awal, tapi antarmukanya masih meminta
URL diketik manual — praktis tidak terpakai.

Sekarang form pembayaran dan form biaya perjalanan memakai pengunggah berkas
(JPG, PNG, WEBP, PDF). Gambar menampilkan pratinjau sebelum disimpan, dan
buktinya tetap bisa dibuka dari daftar pembayaran untuk dicocokkan dengan mutasi
rekening.

### Histori harga produk antar trip

PRD menjadikan ini *secondary goal*: master data product reusable antar trip,
harga tidak perlu dicatat ulang dari nol. Datanya sudah tersimpan (`trip_items`
dan `purchases`) tapi belum pernah bisa dilihat.

Ditambahkan `GET /products/{id}/price-history` yang menggabungkan harga katalog
dan harga beli riil per trip, ditampilkan lewat ikon jam di halaman Produk, dan
sebagai keterangan satu baris saat menyusun katalog trip baru.

PRD juga mencatat bahwa *meski merek barang sama, harga modal bisa berbeda antar
trip karena beda negara asal*. Karena itu harga tidak pernah disalin otomatis
antar trip yang mata uangnya berbeda — tautan "Pakai harga ini" hanya muncul
untuk mata uang yang sama, dan memilih produk yang mata uang masternya berbeda
justru mengosongkan kolom harga modal supaya admin sadar harus mengisinya.

### Kurs diambil otomatis

PRD mengasumsikan kurs diambil otomatis dari third-party API, bukan diketik
manual. Ditambahkan `GET /exchange-rate?from=JPY` yang mengambil kurs harian dari
layanan publik tanpa API key, dengan cache satu jam dan timeout enam detik.

**Kursnya tetap dikunci per trip.** Yang diotomasi hanya pengisian nilainya saat
trip dibuat. Ini penyimpangan yang disengaja dari pembacaan paling harfiah PRD
("kurs real-time"): kalau konversi pada daftar order memakai kurs hari ini, laba
sebuah trip yang sudah selesai akan berubah setiap hari dan laporannya tidak bisa
dipakai untuk evaluasi — padahal evaluasi profitabilitas justru alasan PRD ini
ditulis.

Kalau layanan kursnya tidak bisa dihubungi, form tetap bisa diisi manual.

---

## 3. Fitur di luar scope PRD yang sudah terlanjur dibangun

Dibiarkan apa adanya. Semuanya melebihi PRD, bukan bertentangan.

| Fitur | Catatan PRD |
|---|---|
| Partial fulfillment (`Sebagian`, `Tidak Ada`) | *Out of scope v1* — tapi kejadian toko kehabisan barang nyata di lapangan |
| Surplus belanja → stok + penjualan marketplace | Tidak disebut PRD |
| Laporan profit per order dan per trip | PRD hanya minta rekap penjualan |
| Manajemen stok dan harga pokok rata-rata | Konsekuensi dari surplus di atas |
| Toggle mata uang trip pada daftar order | PRD hanya minta tampilan Rupiah |
| Estimasi ongkir dengan berat volumetrik | PRD hanya menyebut "berat atau volume" |
| Peran **tripper** terpisah dari owner | PRD hanya mengenal Admin Toko dan Owner |

---

## 4. Yang belum dikerjakan

Keduanya butuh keputusan komersial, bukan pekerjaan teknis yang bisa diselesaikan
sendiri.

### Integrasi API JNE untuk estimasi ongkir

PRD: *"Sistem menghitung estimasi ongkir otomatis melalui integrasi dengan JNE"*.

Saat ini estimasi memakai **tabel tarif internal** yang sudah berisi 28 kota
tujuan dan bisa dikelola sendiri lewat Pengaturan → Ongkir. Perhitungan berat
volumetriknya sudah sesuai cara ekspedisi menagih.

Yang dibutuhkan untuk integrasi sungguhan: akun dan kredensial API dari JNE atau
agregator seperti RajaOngkir/Biteship, yang umumnya berbayar dan perlu
pendaftaran bisnis.

**Titik integrasinya sudah disiapkan.** Interface `service.RateProvider`
(`backend/internal/service/shipping_service.go`) tinggal diisi implementasi baru
lalu dipasang lewat `UseProvider()`; handler dan seluruh antarmuka tidak perlu
diubah.

### Notifikasi WhatsApp otomatis

PRD: *"sistem otomatis kirim notifikasi satu arah via WhatsApp ke buyer berupa
invoice pengiriman"*.

Saat ini sistem menyiapkan pesan lengkap beserta tautan `wa.me`, tapi **admin
yang menekan kirim** dari nomor toko sendiri. Ini juga pilihan yang sudah
dikonfirmasi sebelum pembangunan dimulai.

Untuk benar-benar otomatis dibutuhkan WhatsApp Business API resmi (lewat penyedia
seperti Twilio atau 360dialog): berbayar per pesan, perlu verifikasi bisnis, dan
template pesannya harus disetujui Meta lebih dulu. Alternatif tidak resmi berisiko
nomor toko diblokir.

Kalau nanti diputuskan berlangganan, yang perlu ditambah hanyalah pengiriman di
sisi server pada `internal/notify` — teks pesan dan pemicunya (saat resi
tersimpan) sudah ada.

---

## 5. Catatan kecil

- **"Invoice pengiriman"** pada PRD 5.4 poin 18 saat ini berbentuk pesan WhatsApp
  berisi kurir, layanan, nomor resi, dan tautan pelacakan — bukan dokumen PDF
  terpisah. Kalau yang dimaksud PRD adalah dokumen tersendiri, ini perlu
  ditambahkan.
- **Evidence dan Why now** pada PRD masih berupa placeholder `[Isi dengan …]`.
  Tidak memengaruhi pembangunan, tapi perlu diisi kalau dokumen ini dipakai untuk
  meyakinkan pihak lain.
