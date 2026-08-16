# Panduan Pengguna Ibatiks

Ibatiks adalah sistem back office untuk menjalankan bisnis jasa titip: membeli
barang di luar negeri atas nama customer, lalu mengirimkannya setelah kamu
kembali ke Indonesia.

Aplikasi ini dipakai **hanya oleh tim kamu sendiri**. Customer tidak pernah
login. Mereka tetap memesan seperti biasa lewat WhatsApp atau media sosial, dan
admin yang mencatatnya di sini.

Sistem ini ada untuk menjawab empat pertanyaan yang jadi mahal kalau dikerjakan
manual begitu satu trip sudah melewati beberapa puluh order:

1. Apa saja yang harus dibeli di trip ini, dan berapa banyak masing-masing?
2. Siapa sudah bayar berapa, dan siapa yang masih punya utang?
3. Barang mana milik customer mana, dan mana yang jadi stok?
4. Trip ini benar-benar untung atau tidak setelah tiket, bagasi, dan ongkir?

Panduan ini disusun mengikuti urutan menu. Baca dua bab pertama sekali saja,
lalu pakai sisanya sebagai rujukan sesuai halaman yang sedang kamu buka.

## Siapa yang memakainya

| Role | Biasanya siapa | Yang dikerjakan di sini |
|---|---|---|
| **Owner** | Pemilik bisnis | Semuanya, termasuk laporan laba, pengaturan toko, dan manajemen pengguna |
| **Admin** | Operator harian | Trip, order, invoice, pengiriman, stok, piutang |
| **Tripper** | Yang berangkat | Membuka daftar belanja dan mencatat pembelian dari toko |

Beri tripper akun sendiri. Mereka akan memakai aplikasi ini dari ponsel sambil
berdiri di toko di Tokyo, dan hanya butuh dua layar.

## Siklus lengkapnya

Setiap fitur di aplikasi ini duduk di salah satu titik garis berikut. Baca sekali
dan sisa panduan ini akan jauh lebih mudah dipahami.

```
  1  BUAT TRIP                   Tanggal, negara tujuan, kurs
       │
       ▼
  2  SUSUN KATALOG               Pilih produk, atur markup, harga dikunci
       │
       ▼
  3  POSTING KE SOSMED           (di luar aplikasi)
       │
       ▼
  4  CATAT ORDER                 Admin mengetik pesanan yang masuk,
                                 sekalian menandai channel asalnya
       │
       ▼
  5  TERIMA DP                   Order terkunci setelah uang masuk
       │
       ▼
  6  DAFTAR BELANJA              Tersusun otomatis dari semua order trip
       │
       ▼
  7  TRIPPER BELANJA             Boleh beli lebih → kelebihannya jadi stok
       │
       ▼
  8  BARANG DATANG               Cocokkan yang dibawa dengan yang dipesan
       │
       ▼
  9  KEMAS PER CUSTOMER          Satu paket untuk satu order; timbang, ukur,
                                 lalu hitung perkiraan ongkirnya
       │
       ▼
 10  TERBITKAN INVOICE           PDF + pesan WhatsApp siap kirim
       │
       ▼
 11  TERIMA PELUNASAN            Order menjadi lunas
       │
       ▼
 12  KIRIM LEWAT JNE             Input nomor resi, kabari customer
       │
       ▼
 13  LAPORAN                     Omzet − HPP riil − biaya trip = laba,
                                 dirinci per trip, customer, dan channel
```

## Lima konsep kunci

Lima gagasan ini mendasari hampir semua aturan di sistem. Pahami dulu sebelum
memakai aplikasinya, karena inilah yang menjelaskan kenapa sebagian tindakan
diizinkan dan sebagian lagi ditolak.

**Kurs dikunci per trip.** Semua harga pada satu trip dikonversi memakai satu
kurs yang ditetapkan saat trip dibuat. Tombol **Ambil kurs terkini** di form trip
menarik kurs hari ini dari layanan kurs publik supaya tidak perlu disalin dari
aplikasi lain — tapi begitu trip tersimpan, angka itu terkunci. Kalau kurs pasar
bergerak setelah trip ditutup, laba yang dilaporkan tidak ikut bergerak.

**Harga adalah salinan, bukan tautan.** Saat produk dimasukkan ke katalog trip,
harga jualnya dihitung sekali lalu disimpan. Saat customer memesannya, harga itu
disalin lagi ke dalam order. Mengedit master produk nanti tidak pernah mengubah
order yang sudah terlanjur ada.

**HPP berasal dari pembelian nyata.** Laba tidak dihitung dari harga modal
perkiraan yang kamu isi saat membuat order, melainkan dari yang benar-benar
dibayar tripper di kasir.

**Kelebihan belanja jadi aset, bukan beban.** Kalau tripper membeli 8 unit
padahal yang dipesan hanya 5, sisa 3 masuk stok. Biayanya **tidak** dibebankan
ke laba trip. Uangnya memang keluar dari kantong, tapi nilainya masih ada dalam
bentuk barang di gudang. Baru menjadi biaya ketika stok itu terjual.

**Ekspedisi menagih ruang, bukan cuma berat.** Kardus besar yang isinya ringan
lebih mahal daripada yang ditunjukkan timbangan, karena kurir menagih mana yang
lebih besar antara berat asli dan berat volume hasil ukuran kardusnya. Aplikasi
menghitung keduanya, jadi catat dimensinya saat mengemas.

---

# Memulai

## Masuk ke aplikasi

Buka alamat aplikasi lalu isi email dan password yang diberikan owner.

Tidak ada pendaftaran mandiri dan tidak ada reset password lewat email. Kalau
lupa password, owner yang mereset dari **Pengaturan → Pengguna**.

Sesi bertahan tujuh hari, jadi kamu tetap login walau browser ditutup. Menekan
**Keluar** di kiri bawah akan meminta konfirmasi dulu sebelum benar-benar
mengeluarkanmu.

## Peta navigasi

Sidebar dikelompokkan berdasarkan *kapan* layar itu dipakai selama satu trip,
bukan berdasarkan abjad. Tiap kelompok hanya satu baris yang terbentang di
tempat saat ditekan; membuka satu kelompok akan melipat yang sebelumnya,
sehingga daftarnya tidak pernah kembali sepanjang semula. Hanya **Dashboard**
yang jadi tautan langsung, karena melipat satu pilihan cuma menambah klik tanpa
menyembunyikan apa pun.

Kelompok yang memuat halaman terbuka **terbentang otomatis**, jadi sesampainya
di sebuah halaman kamu langsung melihat posisimu. Saat terlipat, baris kelompok
menuliskan nama halaman aktif di bawah judulnya.

```
  ┌──────────────────────┐
  │  Dashboard           │  ← buka tiap pagi
  │  Perjalanan        › │
  │  Penjualan         ⌄ │  ← terbuka: memuat halaman aktif
  │      Order           │
  │      Invoice         │
  │      Pengiriman      │
  │      Siap Kemas      │
  │  Data Master       › │
  │  Lainnya           › │
  ├──────────────────────┤
  │  Owner Ibatiks      │
  │  [ Keluar ]          │  ← minta konfirmasi dulu
  └──────────────────────┘
```

| Kelompok | Menu | Dipakai saat |
|---|---|---|
| Ringkasan | Dashboard | Tiap pagi, melihat apa yang butuh perhatian |
| Perjalanan | Trip | Merencanakan trip, menyusun katalog |
| Perjalanan | Daftar Belanja | Sedang berdiri di toko, belanja |
| Perjalanan | Pembelian | Meninjau apa yang dibeli dan ke mana perginya |
| Penjualan | Order | Mencatat dan mengelola pesanan customer |
| Penjualan | Invoice | Menagih customer |
| Penjualan | Pengiriman | Menginput nomor resi |
| Penjualan | Siap Kemas | Mengerjakan antrean gudang |
| Data Master | Customer | Menambah pembeli baru |
| Data Master | Produk | Merawat katalog produk |
| Data Master | Stok | Mengelola barang sisa untuk marketplace |
| Lainnya | Laporan | Menagih utang, meninjau margin, channel, dan customer |
| Lainnya | Pengaturan | Identitas toko, template pesan, tarif ongkir, pengguna |

## Siapa bisa melihat apa

Menu yang tidak boleh kamu akses disembunyikan, bukan diabu-abukan.

| Layar | Owner | Admin | Tripper |
|---|:---:|:---:|:---:|
| Dashboard | ✓ | ✓ | ✓ |
| Trip (lihat) | ✓ | ✓ | ✓ |
| Trip (buat, ubah, katalog) | ✓ | ✓ | — |
| Daftar Belanja | ✓ | ✓ | ✓ |
| Catat pembelian | ✓ | ✓ | ✓ |
| Hapus pembelian | ✓ | ✓ | — |
| Order | ✓ | ✓ | — |
| Invoice | ✓ | ✓ | — |
| Pengiriman dan packing | ✓ | ✓ | — |
| Customer | ✓ | ✓ | — |
| Produk (lihat) | ✓ | ✓ | ✓ |
| Produk (ubah) | ✓ | ✓ | — |
| Stok | ✓ | ✓ | — |
| Laporan: piutang, performa produk, per customer, per channel | ✓ | ✓ | — |
| Laporan: profit trip, profit order | ✓ | — | — |
| Tarif ongkir (lihat dan pakai perkiraan) | ✓ | ✓ | — |
| Tarif ongkir (tambah, hapus) | ✓ | — | — |
| Pengaturan (semua tab) | ✓ | — | — |

Laporan laba sengaja hanya untuk owner. Admin bisa melihat siapa yang masih
berutang, karena itu memang dibutuhkan untuk menagih, tapi tidak bisa melihat
margin kamu. Tab **Profit per Order** di dalam Laporan disembunyikan sama sekali
dari admin, bukan ditampilkan lalu ditolak.

---

# Dashboard

Halaman pertama setelah login. Menjawab satu pertanyaan: apa yang butuh
perhatianku hari ini.

## Baris atas

Empat penghitung yang normalnya berangka kecil.

| Kartu | Yang dihitung |
|---|---|
| Trip aktif | Trip berstatus buka order, ditutup, belanja, perjalanan pulang, atau tiba |
| Order berjalan | Order yang belum selesai dan belum dibatalkan |
| Siap dikirim | Order sudah lunas tapi belum punya nomor resi |
| Piutang berjalan | Total uang yang masih ditunggu dari customer |

**Siap dikirim** berubah kuning kalau lebih dari nol. Itulah antrean yang paling
merugikan hubungan dengan customer: mereka sudah membayar dan sedang menunggu.

## Baris uang

| Kartu | Isinya |
|---|---|
| Omzet bulan ini | Total semua order bulan ini, kecuali yang batal dan draft |
| Laba kotor bulan ini | Omzet dikurangi HPP barang yang sudah dibeli |
| Nilai stok | Qty stok × harga pokok rata-rata, seluruh produk |

Laba kotor di sini belum dikurangi biaya perjalanan. Untuk gambaran utuh satu
trip, buka tab **Profit** pada trip tersebut.

## Panel bawah

**Order terbaru** menampilkan delapan order terakhir beserta statusnya. Klik
nomor ordernya untuk membuka.

**Trip mendatang** menampilkan trip yang sedang membuka order, diurutkan dari
yang paling dekat berangkat.

**Produk terlaris** mengurutkan produk berdasarkan jumlah terjual, lengkap
dengan omzet, HPP, dan profitnya. Profit negatif di sini biasanya berarti harga
di toko naik setelah harga jual terlanjur diumumkan.

---

# Trip

Satu trip adalah satu perjalanan ke luar negeri. Trip memiliki katalog, kumpulan
order, daftar belanja, biaya, dan laporan laba sendiri.

## Membuat trip

**Trip → Buat Trip**

| Kolom | Wajib | Catatan |
|---|:---:|---|
| Judul trip | Ya | Teks bebas, mis. "Jastip Tokyo Maret 2026" |
| Negara | Ya | |
| Kota | Tidak | |
| Tanggal berangkat | Ya | |
| Tanggal pulang | Ya | Tidak boleh lebih awal dari tanggal berangkat |
| Batas terima order | Tidak | Kosongkan kalau tidak dibatasi |
| Mata uang | Ya | Kode 3 huruf, mis. JPY, KRW, SGD |
| Kurs ke Rupiah | Ya | Berapa rupiah per 1 satuan mata uang tersebut |
| Catatan | Tidak | Rencana toko, batas bagasi |

Kurs adalah isian paling penting. Salah di sini berarti semua harga di trip ini
salah. Pakai kurs sedikit lebih tinggi dari kurs pasar sebagai penyangga.

Trip baru berstatus **Draft** dan belum menerima order sampai kamu membukanya.

## Alur status trip

```
     draft ────────▶ open ◀──────▶ closed
       │              │               │
       │              └──────┬────────┘
       │                     ▼
       │                 shopping ────▶ in_transit
       │                     │               │
       │                     └───────┬───────┘
       │                             ▼
       │                          arrived
       │                             │
       │                             ▼
       │                          settled
       ▼
   cancelled  ◀── (dari draft, open, closed, atau shopping)
```

| Status | Label di aplikasi | Artinya | Terima order? |
|---|---|---|:---:|
| `draft` | Draft | Masih disusun | Tidak |
| `open` | Buka Order | Sudah diumumkan, menerima order | Ya |
| `closed` | Order Ditutup | Batas waktu terlewat | Tidak |
| `shopping` | Sedang Belanja | Tripper sedang belanja di luar negeri | Ya |
| `in_transit` | Perjalanan Pulang | Dalam penerbangan pulang | Tidak |
| `arrived` | Tiba di Indonesia | Barang sudah mendarat | Tidak |
| `settled` | Selesai Dibukukan | Trip ditutup dan dibukukan | Tidak |
| `cancelled` | Batal | Dibatalkan | Tidak |

Order masih diterima saat trip berstatus `shopping`. Ini disengaja. Di lapangan
selalu ada customer yang menyusul titip satu barang lagi padahal tripper sudah
berdiri di toko, dan menolak order itu berarti kehilangan penjualan.

## Perubahan status punya efek samping

Setiap tombol status trip memunculkan kotak konfirmasi yang menyebutkan
dampaknya. Dua di antaranya memindahkan banyak order sekaligus, dan justru
itulah alasan perpindahan trip pantas dijeda sebentar. **Selesai Dibukukan**
diwarnai merah karena tidak ada status setelahnya.

Menggeser status trip ikut menyeret order di dalamnya, jadi kamu tidak perlu
mengubah puluhan order satu per satu.

| Trip berpindah ke | Efek pada order |
|---|---|
| `shopping` | Order berstatus `dp_paid` menjadi `purchasing` |
| `arrived` | Order berstatus `purchasing` menjadi `arrived` |

## Detail trip: lima tab

Membuka trip memberimu lima tab. Masing-masing untuk pekerjaan yang berbeda.

```
  ┌──────────┬────────┬──────────┬────────┬────────┐
  │ Katalog  │ Order  │ Belanja  │ Biaya  │ Profit │
  └──────────┴────────┴──────────┴────────┴────────┘
       │         │          │         │        │
     atur     lihat siapa  beli di  catat    untung
     harga    yang pesan   toko     modal    atau tidak
```

---

## Tab 1 — Katalog

Daftar produk yang ditawarkan pada trip ini, masing-masing dengan harga modal
dan markup sendiri.

### Cara harga jual dihitung

```
   harga modal dalam mata uang asing
             │
             │  × kurs trip
             ▼
   harga modal dalam rupiah  ─────────────────┐
             │                                │
             │  + markup                      │ ini yang jadi
             ▼                                │ modal perkiraan
   harga jual mentah                          │ sebelum
             │                                │ pembelian nyata
             │  bulatkan KE ATAS ke Rp100     │ dicatat
             ▼                                │
   HARGA JUAL YANG DIUMUMKAN  ◀───────────────┘
```

Tersedia dua jenis markup.

| Jenis | Rumus | Dipakai saat |
|---|---|---|
| Persen (%) | `modal × (1 + nilai / 100)` | Barang umum, margin ikut naik bersama harga |
| Nominal (Rp) | `modal + nilai` | Barang murah, persentase jadi terlalu kecil |

### Contoh perhitungan

| Modal | Kurs | Markup | Modal (Rp) | Harga mentah | Diumumkan |
|---|---|---|---|---|---|
| ¥880 | 108,5 | 35% | Rp95.480 | Rp128.898 | **Rp128.900** |
| ¥780 | 108,5 | +Rp40.000 | Rp84.630 | Rp124.630 | **Rp124.700** |
| ¥1.000 | 100 | 30% | Rp100.000 | Rp130.000 | **Rp130.000** |

Pembulatan selalu ke atas ke kelipatan seratus rupiah. Harga Rp128.898 terlihat
asal-asalan di postingan media sosial; Rp128.900 tidak.

### Isian katalog

| Kolom | Catatan |
|---|---|
| Produk | Hanya produk yang belum ada di katalog ini yang muncul |
| Harga modal | Dalam mata uang trip, bukan rupiah |
| Jenis markup | Persen atau nominal |
| Markup | Terisi otomatis dari master produk, bisa diubah |
| Batas kuota | Opsional, batas total unit dari semua order |
| Aktif | Hilangkan centang untuk menyembunyikan tanpa menghapus |

Di bawah form ada pratinjau langsung yang menampilkan modal dalam rupiah, harga
jual, dan margin per unit sebelum kamu menyimpan.

### Hitung ulang harga

**Hitung Ulang Harga** menghitung ulang seluruh harga katalog memakai kurs trip
saat ini.

Ini tindakan manual yang disengaja. Mengubah kurs trip **tidak** diam-diam
mengubah harga katalog yang sudah terlanjur diumumkan ke customer. Mengubah
harga yang sudah dilihat orang adalah keputusan bisnis, bukan efek samping dari
membetulkan salah ketik.

### Mengeluarkan produk dari katalog

Produk yang sudah dipesan customer tidak bisa dikeluarkan. Hilangkan centang
**Aktif** saja. Dengan begitu riwayat order tetap utuh.

---

## Tab 2 — Order

Semua order pada trip ini, lengkap dengan jumlah, total, sisa tagihan, dan
status. **Catat Order** membuka form order dengan trip ini sudah terpilih.

Penjelasan lengkap ada di bab [Order](#order).

---

## Tab 3 — Belanja

Layar kerja tripper. Dibahas lengkap di bab
[Daftar Belanja](#daftar-belanja), karena punya menu sendiri di sidebar.

---

## Tab 4 — Biaya

Biaya trip di luar harga barang. Nilainya dikurangkan dari laba kotor untuk
mendapat laba bersih.

| Kategori | Contoh |
|---|---|
| Tiket | Tiket pesawat, pilih kursi |
| Bagasi | Bagasi tambahan, kelebihan berat |
| Akomodasi | Hotel, penginapan |
| Transport | Kereta, taksi, antar-jemput bandara |
| Visa | Biaya visa, asuransi perjalanan |
| Lainnya | Kartu SIM, bahan packing |

Tiap entri butuh tanggal, keterangan, nominal, dan URL bukti (opsional). Total
berjalan tampil di atas tabel.

Catat biaya sambil jalan. Biaya yang lupa dimasukkan membuat trip terlihat lebih
untung dari kenyataannya, dan itu diam-diam membuat trip berikutnya dihargai
terlalu murah.

---

## Tab 5 — Profit

Laporan keuangan trip.

### Perhitungannya

```
     Omzet              total semua order, kecuali batal dan draft
        −
     HPP                biaya belanja riil yang dialokasikan ke order
   ─────────────────
   = Laba kotor
        −
     Biaya perjalanan   tiket, bagasi, akomodasi, transport
   ─────────────────
   = LABA BERSIH
```

### Arti tiap angka

| Angka | Definisi |
|---|---|
| Omzet | Total nilai order, tidak termasuk batal dan draft |
| HPP barang pesanan | Biaya belanja riil yang dialokasikan ke barang pesanan |
| Laba kotor | Omzet − HPP |
| Biaya perjalanan | Total dari tab Biaya |
| Laba bersih | Laba kotor − biaya perjalanan |
| Margin | Laba bersih ÷ omzet × 100 |

### Angka arus kas

| Angka | Definisi |
|---|---|
| Total modal keluar | **Seluruh** belanja (termasuk kelebihan) + biaya perjalanan |
| Uang masuk dari customer | Yang benar-benar sudah diterima |
| Sisa tagihan belum masuk | Masih ditunggu dari customer |
| Ongkir ditagihkan | Ongkir yang dibebankan ke customer |
| Ongkir dibayar ke kurir | Ongkir yang benar-benar dibayar ke JNE |

Perhatikan bedanya **HPP** dan **total modal keluar**. Modal keluar mencakup
kelebihan belanja yang jadi stok; HPP tidak.

### Kenapa kelebihan stok tidak dihitung sebagai HPP

Kalau panelnya menampilkan surplus stok, baca dengan teliti. Ini angka yang
paling sering disalahpahami di seluruh sistem.

Misalkan kamu membeli 8 unit seharga Rp100.000 per unit padahal yang dipesan
hanya 5:

```
   8 unit dibeli  ×  Rp100.000  =  Rp800.000 keluar dari kantong
        │
        ├── 5 unit → pesanan customer  →  Rp500.000 dihitung sebagai HPP
        │
        └── 3 unit → stok gudang       →  Rp300.000 dihitung sebagai ASET
                                           (belum jadi biaya)
```

Rp300.000 itu muncul di **total modal keluar**, karena uangnya memang benar
keluar. Tapi tidak muncul di **HPP**, karena barangnya masih milikmu. Baru
menjadi biaya pada hari stok itu terjual di marketplace.

Tanpa aturan ini, trip yang kebetulan borong stok akan terlihat rugi padahal
bisnisnya baik-baik saja.

---

# Daftar Belanja

Layar yang dipakai tripper sambil belanja.

Pilih trip dari dropdown di kanan atas. Sistem otomatis memilih trip yang sedang
berjalan, jadi tripper biasanya tidak perlu memilih apa pun.

## Dari mana daftarnya berasal

Daftar belanja **bukan** tabel yang diisi seseorang. Angkanya dihitung langsung
dari seluruh order pada trip itu, setiap kali layar dibuka.

```
   Order A (Diproses):    2 × Lotion, 1 × Snack Box
   Order B (Diproses):    3 × Lotion            ┐
   Order C (Menunggu DP): 4 × Lotion            │ dikelompokkan per produk
                    │                           ┘
                    ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Lotion      dipesan 5   menunggu DP 4   dibeli 0   sisa 5    │
   │ Snack Box   dipesan 2   menunggu DP 0   dibeli 0   sisa 2    │
   └──────────────────────────────────────────────────────────────┘
```

Akibatnya terasa nyata di lapangan: kalau admin di Jakarta mengubah sebuah order
saat tripper sedang di toko di Tokyo, daftar tripper ikut berubah pada
penyegaran berikutnya. Tidak ada daftar terpisah yang harus disamakan.

### Hanya order ber-DP yang dihitung sebagai belanja

Kolom **Dipesan** hanya menghitung order yang DP-nya sudah diverifikasi, yaitu
status `Diproses` ke atas. Yang masih menunggu DP dihitung terpisah pada kolom
**Menunggu DP** dan sengaja **tidak** ikut dibelanjakan.

Alasannya soal uang, bukan kerapian. Membelanjakan order yang DP-nya belum masuk
berarti menalangi pembelian dengan uang toko; kalau customer itu batal, barangnya
mengendap di stok tanpa ada yang membayar. Kolom menunggu DP tetap ditampilkan
supaya tripper melihat permintaan yang menumpuk dan bisa memutuskan untuk menagih
DP-nya dulu sebelum meninggalkan toko.

Selama masih ada yang menunggu, muncul keterangan di atas daftar — supaya angka
yang mengecil tidak dikira pesanan yang berkurang.

Order batal dan draft tidak dihitung sama sekali.

## Empat kartu ringkasan

| Kartu | Artinya |
|---|---|
| Total dipesan | Unit yang diminta customer |
| Sudah dibeli | Unit yang sudah tercatat dibeli |
| Sisa belanja | Yang masih harus dibeli |
| Perkiraan modal | Perkiraan total biaya dalam rupiah |

Baris yang sudah tidak menyisakan apa-apa diredupkan dan ditandai **Lengkap**.

## Mencatat pembelian

Tekan **Catat Beli** pada barisnya.

| Kolom | Catatan |
|---|---|
| Jumlah dibeli | Otomatis terisi sejumlah sisa yang belum dibeli |
| Harga satuan | Dalam mata uang trip, sesuai yang benar-benar dibayar |
| Tanggal beli | Otomatis hari ini |
| Kurs khusus | Opsional, kosongkan untuk memakai kurs trip |
| Toko | Tempat kamu membelinya |
| Catatan | Varian, promo, apa pun yang perlu diingat |

Isi harga yang benar-benar dibayar, termasuk diskon yang kamu dapat. Inilah yang
membuat laporan laba jujur.

Kolom **kurs khusus** ada karena trip panjang bisa melewati beberapa hari dengan
kurs yang berubah. Tiap pembelian menyimpan kurs yang berlaku saat itu.

## Yang terjadi setelah disimpan

Sistem mengalokasikan unitnya secara otomatis. Order dipenuhi dari yang paling
awal memesan, sehingga customer yang pesan duluan dilayani duluan saat barang
di toko terbatas.

```
   Dibeli 8 unit Lotion
        │
        ▼
   ┌────────────────────────────────────────────────────────┐
   │  Order A (pesan 15 Agu, butuh 2)   ──▶  2 unit         │
   │  Order B (pesan 16 Agu, butuh 3)   ──▶  3 unit         │
   │                                         ─────────      │
   │                        dialokasikan ke pesanan: 5      │
   │                                                        │
   │  Tidak ada yang memesan sisanya    ──▶  3 unit         │
   │                                 masuk stok:        3   │
   └────────────────────────────────────────────────────────┘
```

Pesan konfirmasi memberi tahu pembagiannya, misalnya
"5 unit untuk pesanan, 3 unit masuk stok".

Kalau kamu membeli lebih sedikit dari yang dipesan, order paling awal dipenuhi
lebih dulu dan sisanya tetap menunggu. Status pemenuhannya menjadi **Sebagian**.

---

# Pembelian

Jejak seluruh belanja, lintas semua trip.

Saring per trip atau cari berdasarkan nama produk atau toko. Tiap baris
menampilkan jumlah, harga satuan dalam dua mata uang, dan total dalam rupiah.

Dua label menunjukkan ke mana unitnya pergi:

| Label | Artinya |
|---|---|
| *N* pesanan | *N* unit dialokasikan ke pesanan customer |
| *N* stok | *N* unit menjadi stok gudang |

Buka barisnya dengan panah di kiri untuk melihat persisnya order dan customer
mana yang menerima tiap unit.

## Menghapus pembelian

Menghapus akan membatalkan seluruh dampaknya:

- Alokasi ke pesanan dilepas
- Stok yang sempat bertambah ditarik kembali
- Jumlah terbeli pada baris pesanan dihitung ulang

Kalau stok kelebihannya sudah terlanjur terjual, pembatalan akan gagal karena
barangnya tidak ada lagi untuk ditarik. Perbaiki lewat penyesuaian stok saja.

---

# Order

Jantung sistem ini.

## Daftar order

Cari berdasarkan nomor order, nama customer, penerima, atau nomor telepon.
Saring per trip, status, channel penjualan, atau **Belum lunas**.

Sisa tagihan ditampilkan berwarna kuning supaya order yang belum lunas menonjol.

Kolom **Channel** menunjukkan order itu masuk dari mana, dengan warna berbeda
per channel supaya sekali lihat ketahuan channel mana yang paling ramai.

### Melihat nominal dalam mata uang trip

Centang **Mata uang trip**, maka seluruh kolom uang berubah dari rupiah ke mata
uang trip order tersebut, memakai kurs yang dikunci saat trip dibuat.

Order dari trip berbeda dikonversi dengan kursnya masing-masing, jadi daftar
yang mencampur trip Jepang dan Korea akan menampilkan JPY di sebagian baris dan
KRW di baris lain. Tidak ada yang dihitung ulang di database; ini hanya cara
menampilkan, dan order pada trip berkurs IDR tetap tampil dalam rupiah.

Pakai saat mencocokkan angka dengan nota dari toko di luar negeri, yang seluruh
nominalnya memakai mata uang setempat.

## Membuat order

**Catat Order** membuka form yang terbagi tiga bagian.

### Bagian 1 — Trip dan customer

Pilih trip lebih dulu. Pemilih produk masih kosong sebelum itu, karena harganya
diambil dari katalog trip tersebut.

Mengganti trip akan mengosongkan item yang sudah dipilih, karena harganya tidak
lagi berlaku.

**Asal order** mencatat order itu masuk lewat channel mana. Nilai bawaannya
WhatsApp karena dari situlah mayoritas order jastip datang, dan isian ini yang
mengisi laporan Per Channel.

| Nilai | Label | Dipakai untuk |
|---|---|---|
| `whatsapp` | WhatsApp | Chat langsung, broadcast, grup |
| `instagram` | Instagram | Balasan story, DM, komentar |
| `tiktok` | TikTok | Komentar dan DM dari video atau live |
| `marketplace` | Marketplace | Shopee, Tokopedia, dan sejenisnya |
| `lainnya` | Lainnya | Selain itu: datang langsung, referral, telepon |

Pilih sambil mengetik ordernya. Mengingat-ingat asal order seminggu kemudian
hanya menghasilkan tebakan, dan tebakan membuat laporan channel tidak ada
gunanya.

Kalau salah pilih, masih bisa diperbaiki lewat **Ubah Order**.

### Bagian 2 — Produk

Pilih dari katalog trip. Tiap penambahan membawa harga katalognya, yang boleh
kamu ubah per order untuk memberi harga khusus.

Memilih produk yang sama dua kali akan menambah jumlahnya, bukan membuat baris
kembar.

Kalau produk punya kuota dan jumlahmu melebihi sisanya, peringatan muncul di
bawah baris dan server akan menolak saat disimpan.

### Bagian 3 — Alamat pengiriman

Secara bawaan order dikirim ke alamat tersimpan milik customer, ditampilkan
untuk dikonfirmasi.

Centang **Kirim ke alamat lain** untuk mengirimkannya ke tempat lain: hadiah,
kantor, rumah teman. Alamat customer disalin sebagai titik awal, jadi kamu hanya
mengubah bagian yang berbeda.

Alamat itu **disalin ke dalam order**, bukan ditautkan. Kalau tahun depan
customer pindah rumah, order ini tetap menunjukkan ke mana barang benar-benar
dikirim.

### Panel ringkasan

| Kolom | Catatan |
|---|---|
| Diskon | Tidak boleh melebihi subtotal |
| Ongkir ditagihkan | Yang kamu bebankan ke customer |
| DP diminta | Kosongkan untuk memakai 50% otomatis |
| Catatan | Permintaan khusus |

```
   Subtotal            jumlah × harga satuan, semua baris
      − Diskon
      + Ongkir
   ─────────────────
   = TOTAL
      − Sudah dibayar
   ─────────────────
   = SISA TAGIHAN
```

## Alur status order

Order yang baru dicatat langsung berstatus **awaiting_dp**, bukan `draft` —
mencatatnya di back office itu sendiri sudah berarti konfirmasi, dan langkah
berikutnya pasti menagih DP. Status `draft` tetap ada untuk order yang
pengisiannya belum tuntas.

```
              draft
                │
                ▼
           awaiting_dp ◀────┐  ← order baru mulai di sini
                │           │  (mundur kalau
                ▼           │   DP gagal)
             dp_paid ───────┘
                │
                ▼
           purchasing
                │
                ▼
            arrived
                │
                ▼
             packed ◀───────┐
                │           │  (kemas ulang kalau
                ▼           │   invoice dibatalkan)
            invoiced ───────┘
                │
                ▼
              paid
                │
                ▼
            shipped
                │
                ▼
           completed

   cancelled  ◀── dari status mana pun sampai dengan invoiced
```

| Status | Label di aplikasi | Artinya |
|---|---|---|
| `draft` | Draft | Pengisian belum tuntas, belum dikonfirmasi |
| `awaiting_dp` | Menunggu DP | Sudah dikonfirmasi, menunggu uang muka — status awal order baru |
| `dp_paid` | Diproses | DP diterima dan diverifikasi; order masuk hitungan daftar belanja |
| `purchasing` | Dibelikan | Tripper sedang membelikan barangnya |
| `arrived` | Barang Tiba | Barang diterima dan dicocokkan |
| `packed` | Sudah Dikemas | Sudah dikemas atas nama customer |
| `invoiced` | Ditagihkan | Invoice pelunasan sudah dikirim |
| `paid` | Lunas | Lunas, siap dikirim |
| `shipped` | Dikirim | Sudah diserahkan ke kurir, resi tercatat |
| `completed` | Selesai | Dikonfirmasi diterima customer |
| `cancelled` | Batal | Dibatalkan |

Setiap tombol status memunculkan kotak konfirmasi lebih dulu, yang menyebutkan
akibat perpindahannya, bukan sekadar bertanya yakin atau tidak. Perpindahan
status tidak punya tombol urung dan sebagian di antaranya mengunci order, jadi
kotak itu memang untuk dibaca — perpindahan ke `shipped` sengaja diwarnai merah
karena itu.

Dua perpindahan dijaga oleh uang, bukan oleh tombol:

- Tidak bisa pindah ke `dp_paid` sebelum DP yang tercatat benar-benar mencapai
  nominal yang diminta.
- Tidak bisa pindah ke `paid` selama masih ada sisa tagihan.

## Mengubah item order

Ini tindakan yang paling sering dipakai sehari-hari. Customer mengirim pesan
"boleh jadi dua saja tidak?" dan kamu mengubahnya di sini.

Tiap baris bisa diedit langsung lewat ikon pensil.

### Aturannya

| Aturan | Alasannya |
|---|---|
| Tidak bisa diedit sama sekali setelah order `shipped`, `completed`, atau `cancelled` | Paketnya sudah pergi; dokumen harus sesuai kenyataan |
| Jumlah tidak boleh turun di bawah jumlah yang sudah diterima | Barang yang sudah ada di gudang tidak bisa "dibatalkan terima" |
| Menurunkan di bawah jumlah yang sudah dibeli akan melepas kelebihannya ke stok | Barangnya nyata; tidak boleh hilang dari pembukuan |
| Order harus punya minimal satu item | Batalkan ordernya saja |
| Diskon tidak boleh melebihi subtotal | Mencegah total menjadi negatif |

### Yang otomatis terjadi

```
   Kamu mengubah jumlah dari 3 menjadi 2
        │
        ├─▶ Subtotal baris dihitung ulang
        ├─▶ Subtotal dan total order dihitung ulang
        ├─▶ Sisa tagihan dihitung ulang
        ├─▶ 1 unit kelebihan dilepas dari order ini ke stok
        ├─▶ Status mundur kalau order tadinya lunas tapi kini berutang
        └─▶ Catatan ditulis ke jejak perubahan
```

Langkah kelima itu penting. Kalau order sudah lunas lalu kamu *menambah* item,
statusnya turun dari `paid` kembali ke `invoiced` supaya sisa tagihan yang baru
terlihat, bukan tersembunyi di balik label hijau "Lunas".

## Pembayaran

Panel **Pembayaran** menampilkan DP diminta, sudah dibayar, dan sisa tagihan,
diikuti seluruh pembayaran yang tercatat.

| Jenis | Efek ke sisa tagihan |
|---|---|
| DP | Menambah yang sudah dibayar |
| Pelunasan | Menambah yang sudah dibayar |
| Refund | Mengurangi yang sudah dibayar |
| Penyesuaian | Menambah yang sudah dibayar |

Metode yang tersedia: transfer bank, tunai, QRIS, e-wallet, lainnya.

### Mencatat pembayaran

**Catat Bayar** mengisi nominalnya otomatis dengan yang masih kurang: sisa DP
kalau uang mukanya belum lunas, atau seluruh sisa tagihan kalau sudah.

Kolom opsional: referensi (nomor transaksi atau nama pengirim), bukti transfer,
tanggal bayar, dan catatan.

**Bukti transfer** mengunggah struk transfer dari customer — foto atau PDF — ke
server toko sendiri; yang diterima JPG, PNG, WEBP, dan PDF. Berkas gambar
menampilkan pratinjau kecil supaya bisa dipastikan tangkapan layarnya benar
sebelum disimpan, dan setelah tersimpan tautannya tetap bisa dibuka dari daftar
pembayaran untuk dicocokkan dengan mutasi rekening.

Jenis berkas dikenali dari isinya, bukan dari namanya, lalu disimpan dengan nama
acak yang baru — jadi berkas bernama menyesatkan tidak bisa menentukan sendiri di
mana ia mendarat.

### Perubahan status otomatis

| Kondisi setelah pembayaran | Order menjadi |
|---|---|
| DP mencapai nominal yang diminta, order tadinya `awaiting_dp` | `dp_paid` |
| Sisa tagihan mencapai nol | `paid` |

Refund yang lebih besar dari uang yang sudah diterima akan ditolak.

### Menagih DP

Selama DP masih kurang, tombol **Tagih DP** muncul. Tombol itu membuka pesan
WhatsApp siap kirim yang dibentuk dari templatemu, sudah terisi nama customer,
nama trip, total, nominal DP, dan nomor rekening.

## Menerima barang

**Cocokkan Barang** membuka layar pencocokan. Tersedia saat order berstatus
`dp_paid`, `purchasing`, atau `arrived`.

Semua baris otomatis dianggap diterima penuh, karena itulah yang paling sering
terjadi. Ubah hanya baris yang bermasalah.

| Jumlah diterima | Status pemenuhan menjadi |
|---|---|
| Sama dengan yang dipesan | Sudah Dibeli |
| Antara 1 dan jumlah pesanan | Sebagian |
| Nol | Tidak Ada |

Menyimpannya membuat order berpindah ke `arrived`.

Barang yang ditandai tidak tersedia tetap muncul di invoice dengan keterangan
tersebut, supaya customer paham kenapa totalnya berbeda dari pesanan awal.

## Membatalkan order

**Batalkan** meminta alasan (opsional) dan memberi tahu apa yang akan terjadi:

- Barang yang sudah terlanjur dibeli untuk order ini dilepas ke stok
- Uang yang sudah diterima perlu dikembalikan dengan mencatat **refund**

Membatalkan tidak otomatis melakukan refund. Itu tindakan terpisah yang
disengaja, karena uangnya juga harus benar-benar keluar dari rekeningmu.

---

# Invoice

Dokumen tagihan yang dikirim ke customer.

## Dua jenis

| Jenis | Menagih | Diterbitkan saat |
|---|---|---|
| DP | Hanya nominal uang muka | Order dikonfirmasi |
| Pelunasan | Seluruh nilai order dikurangi yang sudah dibayar | Barang sudah tiba |

Kebanyakan bisnis hanya menerbitkan invoice pelunasan dan meminta DP lewat pesan
WhatsApp. Keduanya didukung.

## Menerbitkan invoice

Dari halaman detail order, **Invoice → Terbitkan**.

Pilih jenisnya, atur jatuh tempo (kalau kosong, memakai bawaan dari Pengaturan),
dan tambahkan catatan.

Seluruh nominal **disalin ke dalam invoice** pada saat diterbitkan. Mengedit
order setelah itu tidak mengubah invoice yang sudah dilihat customer. Kalau
ordernya memang berubah, batalkan invoice lama lalu terbitkan yang baru.

Menerbitkan invoice pelunasan membuat order berpindah ke `invoiced`.

## PDF-nya

Tiap invoice dirender menjadi PDF berisi:

- Nama, alamat, telepon, dan email toko dari Pengaturan
- Nomor invoice, tanggal terbit, jatuh tempo
- Data penagihan dan alamat pengiriman
- Rincian barang beserta jumlah, harga satuan, dan subtotal
- Total, sudah dibayar, dan sisa tagihan
- Riwayat pembayaran
- Nomor rekening, kalau masih ada sisa tagihan
- Catatan penutup dari kamu

Buka dengan tombol **PDF**. Berkasnya dirender saat pertama diminta lalu
disimpan.

## Mengirim ke customer

**Kirim** membuka dialog berisi teks pesan lengkap dan dua tombol.

```
   ┌──────────────────────────────────────────────┐
   │  Pesan dibentuk dari templatemu              │
   │  ──────────────────────────────────────────  │
   │  Halo Rina, barang pesananmu sudah sampai…   │
   │  Invoice INV-2026-0001                       │
   │  Total: Rp335.000                            │
   │  Sisa pelunasan: Rp167.500                   │
   │  Transfer ke: BCA 1234567890                 │
   └──────────────────────────────────────────────┘
        │                    │
        ▼                    ▼
   [Salin teks]      [Buka WhatsApp]
    ke clipboard      membuka wa.me dengan
                      pesan sudah terisi
```

Kamu sendiri yang menekan kirim, dari nomormu sendiri. Tidak ada gateway
berbayar dan tidak ada nomor pengirim asing, jadi customer melihat pesan dari
kontak yang mereka kenal.

Begitu kamu membuka WhatsApp, invoice ditandai sudah dikirim beserta kanalnya.

## Status invoice

| Status | Label | Artinya |
|---|---|---|
| `draft` | Draft | Dibuat, belum dikirim |
| `sent` | Terkirim | Sudah disampaikan ke customer |
| `paid` | Lunas | Sudah dibayar penuh |
| `void` | Dibatalkan | Dibatalkan, digantikan invoice lain |

Invoice yang sudah lunas tidak bisa dibatalkan. Uangnya sudah tercatat masuk.

## Membatalkan invoice

**Batalkan** — ikon larangan merah di daftar invoice, atau tombol **Batalkan** di
panel Invoice pada halaman order. Tombolnya hanya muncul untuk invoice yang
belum lunas dan belum dibatalkan.

Dipakai kalau invoice terlanjur diterbitkan dengan angka yang salah: customer
menambah barang setelah ditagih, atau jumlahnya dikoreksi. Jangan diam-diam
diganti — customer bisa jadi sudah memegang PDF yang lama, dan jejaknya perlu
menunjukkan bahwa tagihan itu ditarik.

Dampaknya sengaja dibatasi:

| Yang berubah | Yang tidak berubah |
|---|---|
| Status menjadi `void` | Pembayaran yang sudah tercatat tetap utuh |
| Invoice berhenti dihitung sebagai tagihan berjalan | Status order tetap di tempatnya |
| PDF-nya masih bisa diunduh sebagai arsip | Tidak ada data yang dihapus |

Karena status order tidak ikut bergerak, invoice pengganti bisa langsung
diterbitkan dari panel yang sama.

Kotak konfirmasi menyebutkan semua ini sebelum aksinya dijalankan.

## Daftar invoice

Cari berdasarkan nomor invoice, nomor order, atau customer. Saring per status
atau jenis. Invoice yang lewat jatuh tempo dan belum lunas ditandai **lewat
tempo**.

---

# Pengiriman

Paket dan nomor resi JNE-nya.

## Langkah 1 — Kemas

Dari order, **Tandai Dikemas**.

| Kolom | Catatan |
|---|---|
| Kurir | Otomatis JNE |
| Layanan | REG, YES, OKE, atau JTR |
| Berat | Dalam gram, hasil timbangan |
| Dimensi paket | Panjang × lebar × tinggi dalam cm, boleh dikosongkan |
| Catatan kemasan | Bubble wrap, mudah pecah, jangan ditumpuk |

Order berpindah ke `packed`.

### Menghitung perkiraan ongkir

**Hitung Ongkir** memperkirakan tagihan kurir memakai kota tujuan yang sudah ada
pada order — alamatnya tidak perlu diketik ulang.

Ekspedisi menagih mana yang lebih besar antara berat asli dan **berat volume**,
yaitu perkiraan ruang yang dimakan kardus di dalam truk.

```
                          panjang × lebar × tinggi (cm)
   berat volume (kg)  =  ──────────────────────────────
                                   pembagi

   berat ditagih      =  yang terbesar antara berat asli dan berat volume,
                         dibulatkan KE ATAS ke kilogram penuh
   ongkir             =  berat ditagih × tarif per kg kota tujuan
```

Pembagi bawaannya 6000, sama seperti yang dipakai JNE. Ubah di Pengaturan →
Ongkir kalau ekspedisimu memakai 5000.

**Contoh nyata.** Sekardus mi instan berukuran 40 × 30 × 25 cm, berat 800 gram.

```
   berat volume  =  (40 × 30 × 25) ÷ 6000  =  30000 ÷ 6000  =  5 kg
   berat asli    =  0,8 kg
   ditagih       =  yang terbesar → 5 kg
   ongkir        =  5 × Rp28.000 (Bandung, YES)  =  Rp140.000
```

Kalau hanya menimbang, angkanya terbaca 1 kg alias Rp28.000 — selisih Rp112.000
yang akhirnya ditanggung sendiri dari margin. Inilah penyebab ongkir jastip
meleset yang paling sering terjadi.

Kebalikannya: skincare seberat 2,3 kg dalam kardus kecil dibulatkan **naik** ke
3 kg, karena ekspedisi tidak menjual pecahan kilogram.

Panel hasilnya menampilkan kedua berat, mana yang dipakai, dan tarif per kg-nya.
Kalau kota tujuan belum punya tarif, sistem memberi tahu dan memakai tarif
cadangan — angkanya tetap bisa dipakai, hanya kurang presisi.

Menekan **Simpan** ikut menyimpan dimensi dan hasil perkiraannya pada data
pengiriman, jadi angkanya masih ada saat mengisi ongkir sebenarnya nanti.

## Langkah 2 — Input nomor resi

**Input Resi & Kirim**.

| Kolom | Catatan |
|---|---|
| Nomor resi | Wajib, disimpan huruf besar |
| Ongkir dibayar | Yang **kamu** bayar ke JNE, bukan yang ditagihkan |
| Tanggal kirim | Otomatis hari ini |

Kalau ongkirnya sudah diperkirakan saat pengemasan, ada tautan **Pakai estimasi**
di bawah kolom ongkir untuk mengisinya. Ganti dengan angka sebenarnya dari resi
kalau berbeda — perkiraan itu titik awal, bukan catatan resmi.

Pisahkan dua angka ongkir itu. Ongkir yang ditagihkan ke customer ada di order;
yang kamu bayar ke kurir ada di sini. Selisihnya adalah margin nyata, dan tab
Profit melaporkan keduanya.

### Penjagaan belum lunas

Order yang masih punya sisa tagihan akan ditolak, lengkap dengan nominalnya.

Untuk menerobos, centang **Kirim walau belum lunas**. Pakai hanya untuk
pelanggan yang memang kamu percaya membayar setelah barang diterima. Penerobosan
ini tercatat di jejak perubahan.

Mengirim akan membekukan order: itemnya tidak bisa diubah lagi.

## Langkah 3 — Kabari customer

**Kabari Customer** menyusun pesan WhatsApp berisi kurir, layanan, dan nomor
resi, plus tautan pelacakan JNE. Alurnya sama dengan invoice: kamu yang menekan
kirim.

Sistem mencatat kapan customer dikabari. Daftar pengiriman menandai paket yang
sudah dikirim tapi belum diberitahukan, supaya tidak ada customer yang
bertanya-tanya di mana paketnya.

### Kolom ongkir pada daftar pengiriman

Daftar **Pengiriman** menampilkan ongkir yang benar-benar dibayar. Paket yang
belum diserahkan ke kurir menampilkan hasil perkiraan dengan keterangan
*estimasi*, supaya tidak tertukar dengan angka final saat merekap biaya.

## Langkah 4 — Tutup

**Tandai Diterima** mengubah pengiriman menjadi diterima dan order menjadi
`completed`.

## Status pengiriman

| Status | Label | Artinya |
|---|---|---|
| `packing` | Dikemas | Sedang dikemas |
| `ready` | Siap Kirim | Sudah dikemas, menunggu nomor resi |
| `shipped` | Dikirim | Sudah diserahkan ke kurir |
| `delivered` | Diterima | Dikonfirmasi diterima |
| `returned` | Retur | Kembali karena tidak terkirim |

Pengiriman tidak bisa ditandai terkirim tanpa nomor resi. Database sendiri yang
menjaganya, jadi customer tidak akan pernah dikabari nomor resi kosong.

---

# Siap Kemas

Daftar kerja gudang.

Menjawab satu pertanyaan: hari ini aku harus mengerjakan apa? Pindah antar empat
tahap lewat dropdown.

```
   arrived  ──▶  packed  ──▶  invoiced  ──▶  paid  ──▶  (shipped)
      │            │             │             │
   kemas       terbitkan      tagih         input
               invoice        pelunasan     nomor resi
```

| Tahap | Label | Yang harus dikerjakan |
|---|---|---|
| `arrived` | Siap dikemas | Barang sudah dicocokkan, kemas per customer |
| `packed` | Sudah dikemas | Terbitkan invoice pelunasan |
| `invoiced` | Menunggu pelunasan | Tunggu atau tagih pembayarannya |
| `paid` | Siap dikirim | Input nomor resi JNE |

Tiap baris menampilkan penerima, kota tujuan, jumlah item, dan sisa tagihan.
**Proses** melompat ke halaman order.

Saring per trip ketika satu kiriman baru mendarat dan kamu ingin membereskan
antrean trip itu sekaligus.

---

# Customer

Orang-orang yang memesan darimu.

## Isian

| Kolom | Wajib | Catatan |
|---|:---:|---|
| Nama | Ya | |
| Nomor WhatsApp | Ya | Format apa pun diterima |
| Email | Tidak | Mengaktifkan pilihan kirim lewat email |
| Instagram | Tidak | |
| Alamat | Tidak | Jalan, nomor, RT/RW, kelurahan |
| Kota | Tidak | |
| Provinsi | Tidak | |
| Kode Pos | Tidak | |
| Catatan | Tidak | Preferensi packing, patokan alamat |

Kode customer (CUS-0001) diberikan otomatis.

## Nomor WhatsApp adalah identitasnya

Satu nomor hanya boleh dipakai satu customer. Nomor yang sama, ditulis dengan
format apa pun, akan ditolak saat disimpan.

Ini disengaja: laporan penjualan per customer dikumpulkan berdasarkan nomor
WhatsApp. Kalau satu orang terlanjur tercatat dua kali karena beda ejaan nama,
belanjanya terpecah jadi dua baris dan tidak ada yang terlihat sebagai pelanggan
besar.

## Penanganan nomor telepon

Nomor dirapikan ke format internasional supaya tautan WhatsApp selalu berfungsi.

| Kamu mengetik | Disimpan sebagai |
|---|---|
| `081234567890` | `6281234567890` |
| `0812-3456-7890` | `6281234567890` |
| `+62 812 3456 7890` | `6281234567890` |
| `(0812) 3456-7890` | `6281234567890` |
| `81234567890` | `6281234567890` |
| `+81 90 1234 5678` | `819012345678` |

Awalan `+` atau `00` berarti nomornya sudah membawa kode negara, jadi dibiarkan
apa adanya. Tanpa penanda itu, nomor yang diawali `8` dianggap nomor Indonesia.

Isi alamatnya. Order tidak bisa dibuat tanpa alamat pengiriman, dan menyimpannya
di data customer menghemat pengetikan berulang.

## Menghapus

Menghapus menyembunyikan customer dari daftar tapi menyimpan datanya, karena
order lama masih menunjuk ke sana. Riwayat order tidak pernah hilang.

---

# Produk

Katalog master.

Ini daftar rujukan, bukan daftar harga. Harga jual sesungguhnya ada di katalog
tiap trip, karena kurs dan harga toko berubah tiap perjalanan.

## Isian

| Kolom | Catatan |
|---|---|
| Nama produk | Sertakan ukuran atau varian |
| SKU | Dibuatkan otomatis kalau dikosongkan |
| Kategori | Untuk penyaringan dan laporan |
| Brand | |
| Toko langganan | Membantu tripper menemukannya |
| Mata uang | Mata uang dari harga rujukan |
| Harga modal | Harga beli yang biasanya di luar negeri |
| Jenis markup | Bawaan saat dimasukkan ke katalog trip |
| Markup | Markup bawaan |
| Berat | Gram, untuk perkiraan ongkir |
| URL gambar | |
| Catatan | |
| Produk aktif | Produk nonaktif tidak bisa dimasukkan ke katalog |

Harga modal dan markup di sini adalah **nilai bawaan** saat produk dimasukkan ke
sebuah trip. Keduanya bisa diubah per trip.

Harga modalnya hanya ikut terisi kalau mata uang trip **sama** dengan mata uang
produk ini. Kalau berbeda, kolomnya dibiarkan kosong: merek yang sama dibeli di
Korea, bukan Jepang, bukan angka yang sebanding, dan salah salin di situ meleset
ratusan kali lipat, bukan beberapa persen.

## Riwayat harga antar trip

Ikon jam pada tiap baris produk membuka riwayat harganya: berapa modal produk itu
pada setiap trip yang pernah memuatnya, dalam mata uang trip tersebut dan pada
kurs yang dikunci saat itu.

| Kolom | Artinya |
|---|---|
| Trip | Kode trip, negara, tanggal berangkat, dan kurs yang dikunci |
| Katalog | Harga modal yang diisi saat menyusun katalog trip itu |
| Beli riil | Rata-rata harga yang benar-benar dibayar di kasir |
| Harga jual | Harga jual yang diumumkan ke customer |
| Dibeli / Terjual | Jumlah unit yang dibeli dan yang laku pada trip itu |

Baca kedua kolom modal itu berdampingan. Kalau **Beli riil** terus-menerus lebih
tinggi daripada **Katalog**, berarti harga katalognya dipasang terlalu optimistis
dan selisihnya diam-diam memakan markup.

Angka yang sama muncul sebagai keterangan satu baris saat produk dimasukkan ke
katalog trip baru, lengkap dengan tautan **Pakai harga ini**. Tautan itu hanya
muncul kalau trip sebelumnya memakai mata uang yang sama.

## Kategori

**Kategori** membuka pengelola kecil untuk menambah, mengubah nama, dan
menghapus kategori. Kategori yang masih dipakai produk tidak bisa dihapus.

## Menghapus

Menghapus akan menonaktifkan produk, bukan melenyapkannya. Produk itu hilang
dari pemilih katalog, tapi seluruh order, pembelian, dan laporan lama tetap
utuh.

---

# Stok

Barang milikmu yang tidak dipesan siapa pun.

## Dari mana stok berasal

```
   Tripper beli 8   ──▶   5 ke pesanan customer
                          3 tidak bermilik
                               │
                               ▼
                          ┌──────────┐
                          │   STOK   │
                          └──────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
        terjual di       dilepas dari      dikoreksi lewat
        marketplace      order yang        stock opname
                         dikurangi
```

Tiga hal menciptakan stok: kelebihan belanja, pengurangan jumlah pada order yang
barangnya sudah dibeli, dan penyesuaian manual.

## Harga pokok rata-rata bergerak

Tiap produk membawa satu harga pokok rata-rata yang dihitung ulang setiap kali
stok bertambah.

```
   Stok lama :  3 unit @ Rp100.000  =  Rp300.000
   Stok masuk:  2 unit @ Rp110.000  =  Rp220.000
                                       ──────────
   Sekarang  :  5 unit                  Rp520.000
                                       ──────────
   HPP rata-rata baru = Rp520.000 ÷ 5 = Rp104.000 per unit
```

Dipakai rata-rata, bukan pelacakan per batch, karena barang jastip adalah unit
yang identik dan tidak ada yang menempeli label trip pada tiap botol. Cara ini
menjaga nilai stok tetap masuk akal saat harga beli berbeda antar trip.

## Mencatat penjualan marketplace

**Jual** pada baris stok.

| Kolom | Catatan |
|---|---|
| Jumlah terjual | Tidak boleh melebihi stok tersedia |
| Harga jual per pcs | |
| Kanal penjualan | Shopee, Tokopedia, Instagram |
| Catatan | |

Margin per unit ditampilkan langsung sambil kamu mengetik, memakai harga pokok
rata-rata.

Penjualan marketplace dicatat terpisah dari laba trip. Laporan trip mencakup
barang yang dipesan customer; penjualan stok adalah lini bisnis berbeda dengan
margin berbeda.

## Penyesuaian stok

**Sesuaikan** menyetel jumlah stok ke angka hasil hitung fisik. Selisihnya
ditulis ke riwayat pergerakan beserta alasanmu, sehingga stok yang menyusut
selalu punya penjelasan.

## Riwayat pergerakan

Tab **Riwayat Pergerakan** mencatat setiap perubahan.

| Jenis | Label | Arah |
|---|---|---|
| `in_purchase` | Masuk dari belanja | + |
| `out_order` | Dipakai pesanan | − |
| `out_marketplace` | Terjual marketplace | − |
| `adjustment` | Penyesuaian | + atau − |

---

# Laporan

Sampai lima tab, masing-masing bisa diekspor ke CSV.

Owner melihat kelimanya. Admin melihat semuanya kecuali Profit per Order; tab
margin itu disembunyikan dari mereka.

## Piutang

Semua order yang masih punya sisa tagihan, dari yang paling lama menunggu.

| Kolom | Catatan |
|---|---|
| Order | Klik untuk membuka |
| Customer | Ada tautan "Tagih via WA" |
| Total | Nilai order |
| Sudah bayar | Yang sudah diterima |
| Sisa | Yang masih ditunggu |
| Umur | Jumlah hari sejak tanggal order |

Yang lewat 14 hari ditampilkan merah. Kerjakan daftar ini dari atas: utang
paling lama adalah yang paling sulit ditagih.

## Profit per Order

Margin per order. Khusus owner; tabnya tidak muncul untuk admin.

| Kolom | Artinya |
|---|---|
| Omzet | Total order |
| HPP | Biaya belanja riil yang dialokasikan ke order ini |
| Profit | Omzet − HPP |
| Margin | Profit ÷ omzet × 100 |

Order yang HPP-nya nol berarti pembeliannya belum diinput. Itu data yang belum
lengkap, bukan margin 100%.

Profit negatif biasanya berarti harga di toko naik setelah harga jual terlanjur
diumumkan. Bandingkan dengan katalog trip untuk memastikan.

## Performa Produk

Jumlah terjual, jumlah order, omzet, HPP, dan profit per produk. Bisa disaring
per trip.

Pakai untuk memutuskan apa yang layak dibawa trip berikutnya dan di mana
markup-mu terlalu tipis.

## Per Customer

Satu baris per customer, dari yang paling besar belanjanya. Saring per trip
untuk melihat siapa saja yang membeli pada satu trip, atau biarkan semua trip
untuk angka sepanjang masa.

| Kolom | Artinya |
|---|---|
| Customer | Nama, kode, kota, dan tanggal order terakhirnya |
| Order | Jumlah order yang tidak dibatalkan |
| Pcs | Total jumlah barang dari order-order itu |
| Omzet | Jumlah total order |
| Rata-rata | Omzet ÷ jumlah order |
| Profit | Omzet − HPP riil |
| Piutang | Yang masih ditunggu dari seluruh ordernya |

Ada dua cara memakainya. Baca dari atas untuk menentukan siapa yang didahulukan
saat slot trip terbatas atau produk punya kuota. Lalu baca kolom Piutang:
customer yang sering belanja *sekaligus* sering nunggak adalah persoalan yang
berbeda dengan yang sekadar belanja banyak.

Customer dengan omzet tinggi tapi profit rendah berarti yang dibelinya
produk-produk bermargin tipis. Itu perlu diketahui sebelum memberi mereka diskon.

## Per Channel

Order datang dari mana, berdasarkan isian **Asal order**.

| Kolom | Artinya |
|---|---|
| Channel | WhatsApp, Instagram, TikTok, Marketplace, atau Lainnya |
| Order | Order dari channel itu yang tidak dibatalkan |
| Customer | Jumlah customer berbeda yang order lewat channel itu |
| Omzet | Jumlah total order |
| Rata-rata | Omzet ÷ jumlah order |
| Profit | Omzet − HPP riil |
| Porsi omzet | Persentase channel itu terhadap total omzet, digambar sebagai bar |

Porsinya selalu berjumlah 100%, jadi bar antar channel bisa langsung
dibandingkan.

Baca sambil mempertimbangkan tenaga yang dihabiskan tiap channel. Channel dengan
sedikit order tapi nilai rata-rata tinggi bisa jadi lebih berharga daripada
channel ramai yang ordernya kecil-kecil — kolom Rata-rata yang membedakannya.

Order yang tercatat sebelum isian ini ada dianggap berasal dari WhatsApp, jadi
angka-angka awal akan condong ke sana sampai cukup banyak order baru terkumpul.

## Ekspor CSV

Tiap tab punya tombol **Ekspor CSV**. Berkasnya terbuka rapi di Excel, termasuk
karakter beraksen pada nama customer.

---

# Pengaturan

Hanya owner yang bisa mengubah. Lima tab.

## Toko

Muncul di PDF invoice dan pesan ke customer.

| Pengaturan | Dipakai di |
|---|---|
| Nama toko | Kepala invoice, template pesan |
| Nomor WA toko | Invoice |
| Email toko | Invoice |
| Alamat toko | Invoice |
| Rekening pembayaran | Invoice dan setiap permintaan pembayaran |
| Catatan penutup invoice | Kaki invoice |
| Jatuh tempo invoice | Jatuh tempo bawaan pada invoice baru |

Pastikan nomor rekeningnya benar. Nomor itu muncul di setiap invoice dan setiap
permintaan pembayaran, dan salah ketik di sini berarti uang tidak sampai ke mana
pun.

## Template Pesan

Tiga template yang bisa diubah. Placeholder dalam `{{kurung ganda}}` diganti
dengan data sesungguhnya saat pesan dibentuk.

| Template | Dipakai saat |
|---|---|
| Pesan permintaan DP | Menekan **Tagih DP** |
| Pesan penagihan pelunasan | Mengirim invoice |
| Pesan informasi pengiriman | Mengabari nomor resi |

### Placeholder yang tersedia

| Placeholder | Diganti dengan |
|---|---|
| `{{customer_name}}` | Nama customer |
| `{{store_name}}` | Nama tokomu |
| `{{trip_title}}` | Judul trip |
| `{{order_number}}` | Nomor order |
| `{{invoice_number}}` | Nomor invoice |
| `{{total}}` | Total terformat, mis. Rp335.000 |
| `{{dp_amount}}` | Nominal DP yang diminta |
| `{{amount_paid}}` | Yang sudah dibayar |
| `{{amount_due}}` | Sisa tagihan |
| `{{due_date}}` | Jatuh tempo invoice |
| `{{bank_account}}` | Rekening dari pengaturan toko |
| `{{courier}}` | Nama kurir |
| `{{service}}` | Jenis layanan |
| `{{tracking_number}}` | Nomor resi |
| `{{recipient_name}}` | Nama penerima pada alamat kirim |

Placeholder yang tidak dikenal dibiarkan apa adanya di dalam pesan, sehingga
salah ketik langsung terlihat alih-alih menghilang diam-diam.

## Ongkir

Dua bagian: cara ongkir dihitung, dan tabel tarif yang dipakai menghitungnya.

### Pengaturan perhitungan

| Pengaturan | Bawaan | Artinya |
|---|---|---|
| Pembagi berat volume | 6000 | Membagi panjang × lebar × tinggi dalam cm menjadi kilogram volume |
| Tarif cadangan per kg | Rp25.000 | Dipakai kalau kota tujuan belum ada di tabel tarif |

JNE, SiCepat, dan J&T sama-sama memakai 6000 untuk paket dalam negeri. Sebagian
ekspedisi dan hampir semua kargo udara internasional memakai 5000, yang membuat
setiap paket besar jadi lebih mahal. Ubah hanya kalau ketentuan ekspedisimu
memang begitu.

### Tabel tarif

Satu baris per kombinasi kurir, layanan, dan kota tujuan. Saat menghitung
perkiraan, kota kirim pada order dicocokkan ke sini lebih dulu; baru kalau tidak
ketemu, tarif cadangan yang dipakai.

| Kolom | Catatan |
|---|---|
| Kota tujuan | Pencocokannya tidak membedakan huruf besar-kecil; "Kota Bandung" dan "bandung" dianggap sama |
| Provinsi | Hanya keterangan, untuk membedakan kota yang namanya mirip |
| Kurir | Bawaannya JNE |
| Layanan | REG, YES, OKE, atau JTR |
| Tarif per kg | Yang ditagih ekspedisi per kilogram |
| Berat minimum | Biasanya 1000 g — ekspedisi menagih minimal 1 kg |
| Estimasi tiba | Teks bebas yang ikut ditampilkan, misalnya "2-3 hari" |

Menyimpan kota yang sudah ada untuk kurir dan layanan yang sama akan
**memperbarui** baris itu, bukan menggandakannya. Jadi memperbarui tarif cukup
dengan memasukkan ulang kotanya dengan harga baru.

Sistem sudah berisi tarif untuk kota-kota yang paling sering dikirimi jastip.
Tambahkan kotamu sendiri begitu ketemu: perkiraan yang memakai tarif cadangan
itu tebakan, sedangkan yang memakai tarif asli adalah angka yang berani kamu
sebutkan ke customer.

Mengubah tarif hanya boleh owner; admin bisa melihatnya dan memakai fitur
perkiraan.

## Pengguna

Menambah, mengubah, menonaktifkan, dan menghapus akun, serta mereset password.

| Tindakan | Catatan |
|---|---|
| Tambah pengguna | Nama, email, password, role |
| Ubah pengguna | Email tidak bisa diubah; nonaktifkan lalu buat baru |
| Reset password | Mengeluarkan pengguna itu dari semua perangkat |
| Nonaktifkan | Memblokir login tanpa menghapus riwayat |
| Hapus | Permanen; catatan jejak perubahan tetap tersimpan |

Ada dua penjagaan: kamu tidak bisa menghapus akunmu sendiri, dan tidak bisa
menghapus atau menurunkan owner aktif terakhir. Kalau bisa, tidak akan ada lagi
yang mampu mengelola pengguna.

## Jejak Perubahan

Siapa mengubah apa, dan kapan. Bisa disaring per jenis entitas.

| Aksi | Label | Dicatat untuk |
|---|---|---|
| `create` | Dibuat | Order, pembelian, invoice baru |
| `update` | Diubah | Perubahan order, perubahan pengaturan |
| `item_change` | Ubah item | Perubahan jumlah dan harga pada order |
| `status_change` | Ubah status | Perpindahan status order dan trip |
| `payment_record` | Catat pembayaran | Pembayaran yang dicatat |
| `delete` | Dihapus | Pembayaran dan pembelian yang dihapus |
| `ship` | Kirim | Nomor resi yang diinput |

Inilah layar yang dibuka saat ada angka yang terasa janggal dan kamu perlu tahu
siapa yang mengubah jumlahnya, dan kapan.

---

# Lampiran A — Uang dan pembulatan

Semua nominal disimpan dengan dua angka desimal dan dihitung memakai aritmetika
desimal yang persis, bukan bilangan pecahan mengambang. Selisih satu rupiah pada
seratus order akan terlihat di laporan laba, jadi hal itu dirancang supaya tidak
terjadi.

| Situasi | Aturan |
|---|---|
| Konversi mata uang | Dikalikan kurs, dibulatkan ke rupiah penuh |
| Harga jual yang diumumkan | Dibulatkan **ke atas** ke kelipatan Rp100 |
| Persentase DP | Dibulatkan ke rupiah penuh |
| Total laporan | Dibulatkan ke rupiah penuh untuk ditampilkan |

Ditampilkan sebagai `Rp1.250.000`, memakai pemisah ribuan gaya Indonesia.

---

# Lampiran B — Contoh lengkap dari awal sampai akhir

Satu trip utuh dengan angka nyata, supaya kamu bisa mencocokkan sistem dengan
hitunganmu sendiri.

**Persiapan.** Trip ke Jepang, kurs ¥1 = Rp100.

| Produk | Modal | Markup | Harga jual |
|---|---|---|---|
| Lotion | ¥1.000 | 30% | Rp130.000 |
| Snack Box | ¥500 | +Rp25.000 | Rp75.000 |

**Order.**

| Order | Customer | Item | Total |
|---|---|---|---|
| ORD-0001 | Rina | 2 × Lotion, 1 × Snack Box | Rp335.000 |
| ORD-0002 | Budi | 3 × Lotion | Rp390.000 |

**DP.** Rina membayar Rp167.500 (50%). Budi membayar Rp195.000 (50%).

**Pembelian.** Tripper membeli 8 Lotion @ ¥1.000 dan 1 Snack Box @ ¥500.

```
   8 Lotion  ──▶  5 ke pesanan (2 Rina + 3 Budi),  3 ke stok
   1 Snack   ──▶  1 ke pesanan (Rina),             0 ke stok
```

**Ada perubahan.** Budi mengurangi pesanannya dari 3 menjadi 2.

| Efek | Hasil |
|---|---|
| Total order | Rp390.000 → Rp260.000 |
| Sisa tagihan Budi | Rp65.000 |
| Stok | 3 → 4 unit Lotion |

**Penyelesaian.** Order Rina diterima, dikemas, ditagih, dilunasi, dan dikirim.
Order Budi masih menyisakan tagihan.

**Biaya trip.** Bagasi tambahan Rp850.000, kereta bandara Rp150.000.

**Laporannya.**

| Baris | Perhitungan | Nominal |
|---|---|---|
| Omzet | 335.000 + 260.000 | **Rp595.000** |
| HPP | 4 Lotion × 100.000 + 1 Snack × 50.000 | **Rp450.000** |
| Laba kotor | 595.000 − 450.000 | **Rp145.000** |
| Biaya perjalanan | 850.000 + 150.000 | **Rp1.000.000** |
| **Laba bersih** | 145.000 − 1.000.000 | **−Rp855.000** |

| Baris kas | Perhitungan | Nominal |
|---|---|---|
| Total modal keluar | (8 × 100.000) + (1 × 50.000) + 1.000.000 | Rp1.850.000 |
| Kelebihan stok | 4 unit × Rp100.000 | Rp400.000 |
| Sisa tagihan | Sisa utang Budi | Rp65.000 |

Perhatikan bahwa HPP menghitung 4 unit Lotion (2 untuk Rina, 2 untuk Budi),
bukan 8 yang dibeli. Sisa 4 unit ada di stok senilai Rp400.000 dan tidak
dibebankan ke trip ini.

Trip ini rugi karena biaya Rp1.000.000 hanya ditanggung dua order kecil. Itu
kesimpulan yang benar: pada trip sungguhan, biaya tetap seperti itu ditanggung
jauh lebih banyak order.

---

# Lampiran C — Tugas yang sering dikerjakan

| Kalau ingin… | Buka |
|---|---|
| Memulai trip baru | Trip → Buat Trip |
| Mengatur harga untuk sebuah trip | Trip → buka → Katalog |
| Mencatat order dari WhatsApp | Order → Catat Order |
| Menagih DP ke customer | Detail order → Tagih DP |
| Melihat apa yang harus dibeli di luar negeri | Daftar Belanja |
| Mencatat barang yang baru dibeli | Daftar Belanja → Catat Beli |
| Mencocokkan barang datang dengan pesanan | Detail order → Cocokkan Barang |
| Melihat pekerjaan gudang hari ini | Siap Kemas |
| Menagih customer | Detail order → Invoice → Terbitkan |
| Menginput nomor resi JNE | Detail order → Input Resi & Kirim |
| Mengetahui siapa yang berutang | Laporan → Piutang |
| Melihat trip untung atau tidak | Trip → buka → Profit |
| Menjual barang sisa | Stok → Jual |
| Mengganti rekening di invoice | Pengaturan → Toko |
| Memberi tripper akun sendiri | Pengaturan → Pengguna |
| Mengetahui siapa mengubah sebuah order | Pengaturan → Jejak Perubahan |

---

# Lampiran D — Daftar istilah

| Istilah | Padanan | Artinya di sini |
|---|---|---|
| Jastip / jasa titip | Personal shopping | Membelikan barang di luar negeri atas permintaan |
| Tripper | Traveller | Orang yang berangkat dan berbelanja |
| DP (uang muka) | Deposit | Pembayaran sebagian yang mengunci order |
| Pelunasan | Settlement | Pembayaran terakhir |
| Resi | Tracking number | Nomor pelacakan kurir |
| Ongkir | Shipping fee | Biaya pengiriman |
| HPP | COGS | Harga pokok penjualan |
| Omzet | Revenue | Total penjualan kotor |
| Laba kotor | Gross profit | Omzet dikurangi HPP |
| Laba bersih | Net profit | Laba kotor dikurangi biaya |
| Piutang | Receivables | Uang yang masih ditunggu dari customer |
| Kurs | Exchange rate | Nilai tukar mata uang asing ke rupiah |
| Markup | Markup | Tambahan atas modal untuk menetapkan harga jual |
| Stok | Stock | Barang di gudang yang belum bermilik |
| Kuota | Quota | Batas maksimal unit yang ditawarkan pada satu trip |
