# Manual Pengguna Ibatiks

Panduan langkah demi langkah memakai aplikasi Ibatiks, lengkap dengan tampilan
layar sebenarnya.

Manual ini disusun mengikuti urutan menu di sisi kiri aplikasi. Kalau kamu baru
pertama kali memakainya, baca berurutan dari awal. Kalau sudah terbiasa,
langsung lompat ke bab menu yang sedang kamu buka.

Semua contoh di manual ini memakai satu trip yang sama, yaitu **Jastip Tokyo**
dengan tiga customer: Rina, Budi, dan Sari. Mengikuti mereka dari awal sampai
barangnya terkirim akan memberi gambaran utuh cara kerja aplikasinya.

> **Catatan.** Angka dan nama pada tangkapan layar adalah data contoh. Yang kamu
> lihat di aplikasimu sendiri tentu berbeda.

---

# 1. Masuk ke aplikasi

Buka alamat aplikasi di browser. Kamu akan disambut halaman masuk.

{{img:01-login}}

**Langkahnya:**

1. Isi **Email** dengan alamat email yang diberikan owner.
2. Isi **Password**.
3. Tekan **Masuk**.

Kalau email atau passwordnya salah, muncul pesan merah di bawah kolom isian.
Aplikasi sengaja tidak memberi tahu mana yang salah, email atau passwordnya.

**Kalau lupa password:** tidak ada tombol reset lewat email. Hubungi owner, dan
dia bisa membuatkan password baru dari menu Pengaturan.

Sekali masuk, kamu akan tetap login selama tujuh hari, walau browser ditutup.

## Cara memakai menu di kiri

Menu dikelompokkan jadi lima baris saja. Menekan satu baris membentangkan isinya
ke bawah; menekan baris lain akan melipat yang sebelumnya, jadi daftarnya tidak
pernah memanjang ke mana-mana.

{{img:04-menu-accordion}}

| Baris | Isinya |
|---|---|
| **Dashboard** | Langsung membuka dashboard, tanpa dilipat |
| **Perjalanan** | Trip, Daftar Belanja, Pembelian |
| **Penjualan** | Order, Invoice, Pengiriman, Siap Kemas |
| **Data Master** | Customer, Produk, Stok |
| **Lainnya** | Laporan, Pengaturan |

Kelompok yang memuat halaman sedang kamu buka **terbentang otomatis**, jadi
begitu sampai di sebuah halaman kamu langsung melihat posisimu — pada gambar di
atas, "Penjualan" terbuka dengan "Order" tersorot.

Kalau kelompoknya sedang terlipat, nama halaman yang aktif tetap dituliskan
kecil di bawah judul kelompoknya.

> Menu yang tidak boleh kamu akses tidak ditampilkan sama sekali, bukan
> ditampilkan lalu ditolak. Jadi isi menu admin dan tripper berbeda dengan
> owner.

## Tombol keluar

Tombol **Keluar** ada di paling bawah sidebar, di bawah nama akunmu, bergaris
merah supaya mudah dicari. Posisinya menempel di dasar layar dan tidak ikut
menggulir, jadi tetap terlihat walau isi halamannya panjang.

Aplikasi meminta konfirmasi dulu sebelum benar-benar mengeluarkanmu.

{{img:03-keluar}}

---

# 2. Dashboard

Halaman pertama setelah masuk. Gunanya menjawab satu pertanyaan: *apa yang perlu
saya kerjakan hari ini?*

{{img:02-dashboard}}

## Cara membacanya

**Baris atas** berisi empat angka yang normalnya kecil:

| Kartu | Artinya | Kapan perlu ditindaklanjuti |
|---|---|---|
| Trip aktif | Trip yang sedang berjalan | — |
| Order berjalan | Order yang belum selesai | — |
| Siap dikirim | Sudah lunas tapi belum ada resi | **Segera**, customer sudah bayar dan menunggu |
| Piutang berjalan | Total uang yang belum masuk | Kalau angkanya membesar terus |

Kartu **Siap dikirim** berubah kuning kalau lebih dari nol. Itu antrean paling
mendesak: uangnya sudah diterima, barangnya belum jalan.

**Baris kedua** berisi angka uang bulan berjalan. Laba kotor di sini belum
dikurangi biaya perjalanan. Untuk melihat untung-rugi satu trip secara utuh,
buka tab **Profit** di halaman trip.

**Panel bawah** menampilkan order terbaru, trip yang akan berangkat, dan produk
terlaris. Klik nomor order atau nama trip untuk membukanya langsung.

---

# 3. Trip

Semua pekerjaan di aplikasi ini berputar di sekitar trip. Buat tripnya dulu,
baru order bisa dicatat.

## 3.1 Melihat daftar trip

Buka menu **Trip**.

{{img:10-trip-list}}

Tabelnya menampilkan tanggal berangkat dan pulang, kurs, jumlah produk di
katalog, jumlah order, dan status trip. Gunakan kotak pencarian atau saringan
status kalau tripmu sudah banyak.

## 3.2 Membuat trip baru

Tekan **Buat Trip** di kanan atas.

{{img:11-trip-form}}

**Langkahnya:**

1. Isi **Judul trip**, misalnya "Jastip Tokyo Oktober 2026".
2. Isi **Negara** dan **Kota** tujuan.
3. Pilih **Tanggal berangkat** dan **Tanggal pulang**.
4. Isi **Batas terima order** kalau kamu ingin menutup pendaftaran sebelum
   berangkat. Kosongkan kalau tidak dibatasi.
5. Isi **Mata uang** dengan kode tiga huruf: `JPY` untuk yen, `KRW` untuk won,
   `SGD` untuk dolar Singapura.
6. Isi **Kurs ke Rupiah**, yaitu berapa rupiah untuk 1 satuan mata uang itu.
7. Tekan **Buat Trip**.

> **Kurs adalah isian paling penting di form ini.** Semua harga jual pada trip
> ini dihitung dari kurs tersebut. Pakai angka sedikit di atas kurs pasar
> sebagai penyangga, supaya kamu tidak rugi kalau kurs bergerak sebelum
> berangkat.

Setelah dibuat, trip berstatus **Draft** dan belum menerima order. Susun
katalognya dulu, baru ubah statusnya menjadi **Buka Order**.

## 3.2.1 Mengisi kurs tanpa mengetik

Saat membuat trip, isi dulu **Mata uang** (tiga huruf, misalnya `KRW`), lalu
tekan **Ambil kurs terkini**.

{{img:12-trip-kurs}}

Kolom kurs terisi sendiri dari layanan kurs internet, lengkap dengan keterangan
kapan diambil dan dari mana. Kalau layanannya sedang tidak bisa dihubungi, muncul
pesan merah dan kamu tinggal mengetik kursnya sendiri seperti biasa.

> **Kurs ini dikunci begitu trip disimpan.** Menekan tombol itu lagi besok tidak
> mengubah harga yang sudah telanjur diumumkan. Itu memang disengaja: laba trip
> yang sudah selesai tidak boleh berubah hanya karena kurs hari ini bergerak.

## 3.3 Menyusun katalog

Katalog adalah daftar barang yang kamu tawarkan pada trip ini, lengkap dengan
harganya. Buka trip lalu masuk ke tab **Katalog**.

{{img:12-trip-katalog}}

Tekan **Tambah Produk** untuk memasukkan barang.

{{img:13-trip-katalog-form}}

**Langkahnya:**

1. Pilih **Produk** dari daftar. Yang muncul hanya produk yang belum ada di
   katalog trip ini.
2. Isi **Harga modal** dalam mata uang trip, bukan rupiah. Kalau harganya ¥880,
   tulis `880`.
3. Pilih **Jenis markup**:
   - **Persen (%)** untuk barang umum, keuntungan ikut naik bersama harga.
   - **Nominal (Rp)** untuk barang murah, karena persentase jadi terlalu kecil.
4. Isi **Markup**.
5. Isi **Batas kuota** kalau stok toko terbatas dan kamu tidak mau kebanjiran
   pesanan. Kosongkan kalau bebas.
6. Tekan **Simpan**.

Perhatikan kotak **Perkiraan harga** di tengah form. Kotak itu menghitung
langsung sambil kamu mengetik, jadi kamu bisa coba-coba markup sampai harganya
terasa pas sebelum menyimpan.

### Cara harga jualnya dihitung

```
   harga modal (mata uang asing)
        │  × kurs trip
        ▼
   harga modal rupiah
        │  + markup
        ▼
   harga mentah
        │  dibulatkan KE ATAS ke kelipatan Rp100
        ▼
   HARGA JUAL
```

Contoh: ¥880 × kurs 108,5 = Rp95.480. Ditambah markup 35% menjadi Rp128.898.
Dibulatkan ke atas menjadi **Rp128.900**.

Pembulatan ke ratusan membuat harga enak dilihat saat diposting ke media sosial.

> **Harga dikunci saat disimpan.** Kalau nanti kamu mengubah kurs trip, harga
> yang sudah ada di katalog **tidak** ikut berubah. Ini disengaja: harga yang
> sudah terlanjur diumumkan ke customer tidak boleh bergeser diam-diam. Kalau
> memang ingin mengubahnya, tekan **Hitung Ulang Harga**.

### Mengeluarkan produk dari katalog

Produk yang sudah dipesan customer tidak bisa dihapus. Tekan ikon pensil lalu
hilangkan centang **Aktif** — produknya berhenti ditawarkan tapi riwayat order
tetap utuh.

## 3.4 Membuka trip untuk order

Setelah katalog siap, tekan tombol status di kanan atas halaman trip. Urutannya:

```
  Draft ──▶ Buka Order ──▶ Sedang Belanja ──▶ Tiba di Indonesia
                                                      │
                                                      ▼
                                             Selesai Dibukukan
```

Tombol yang muncul hanya status berikutnya yang masuk akal, jadi kamu tidak bisa
salah melompat.

Setiap tombol status memunculkan **kotak konfirmasi** yang menjelaskan apa yang
akan terjadi. Baca dulu isinya, terutama untuk dua perpindahan di bawah ini yang
mengubah banyak order sekaligus. Kalau salah tekan, tekan **Batal**.

Dua perpindahan punya efek otomatis yang menghemat banyak pekerjaan:

| Kamu tekan | Yang terjadi otomatis |
|---|---|
| **Sedang Belanja** | Semua order yang DP-nya sudah masuk berpindah ke status "Dibelikan" |
| **Tiba di Indonesia** | Semua order yang sedang dibelikan berpindah ke "Barang Tiba" |

> **Selesai Dibukukan** tombolnya berwarna merah karena tidak ada jalan kembali.
> Pastikan seluruh belanja dan biaya perjalanan sudah tercatat sebelum
> menekannya.

## 3.5 Melihat order pada satu trip

Tab **Order** menampilkan semua pesanan pada trip ini.

{{img:14-trip-order}}

Tombol **Catat Order** di sini langsung membuka form dengan trip yang tepat
sudah terpilih.

## 3.6 Mencatat biaya perjalanan

Tab **Biaya** untuk mencatat pengeluaran di luar harga barang: tiket, bagasi,
hotel, transport lokal.

{{img:16-trip-biaya}}

**Langkahnya:**

1. Tekan **Catat Biaya**.
2. Pilih **Kategori**.
3. Isi **Tanggal**, **Keterangan**, dan **Nominal**.
4. Isi **URL bukti** kalau kamu menyimpan foto struknya di suatu tempat.
5. Tekan **Simpan**.

> **Catat sambil jalan, jangan ditunda.** Biaya yang lupa dimasukkan membuat
> trip terlihat lebih untung dari kenyataannya. Akibatnya kamu bisa memasang
> harga terlalu murah di trip berikutnya.

## 3.7 Melihat untung-rugi trip

Tab **Profit**. Hanya owner yang bisa membukanya.

{{img:17-trip-profit}}

**Cara membacanya, dari atas ke bawah:**

```
   Omzet             total semua order (kecuali batal dan draft)
     − HPP           biaya belanja yang benar-benar terjadi
   ─────────────
   = Laba kotor
     − Biaya trip    tiket, bagasi, akomodasi, transport
   ─────────────
   = LABA BERSIH
```

Panel **Arus kas & posisi** di sebelah kanan menjawab pertanyaan berbeda: berapa
uang yang sudah keluar, berapa yang sudah masuk, dan berapa yang masih ditunggu.

### Kotak biru tentang stok

Kalau kamu membeli lebih banyak dari yang dipesan, akan muncul kotak biru
seperti pada gambar di atas.

Maksudnya begini. Misalkan kamu membeli 8 unit tapi yang dipesan hanya 5:

```
   8 unit × Rp100.000 = Rp800.000 keluar dari kantong
        │
        ├── 5 unit → pesanan customer → Rp500.000 dihitung sebagai HPP
        │
        └── 3 unit → stok gudang      → Rp300.000 dihitung sebagai ASET
```

Rp300.000 itu **tidak** dipotong dari laba trip, karena barangnya masih milikmu
dan bisa dijual kapan saja. Uangnya tetap tercatat di baris "Total modal keluar".
Nanti kalau stok itu terjual, barulah menjadi biaya.

Tanpa aturan ini, trip yang kebetulan borong stok akan terlihat rugi padahal
tidak.

---

# 4. Daftar Belanja

Layar yang dipakai tripper sambil berdiri di toko. Buka menu **Daftar Belanja**.

{{img:20-belanja}}

## Cara membacanya

Daftar ini **tidak diketik siapa pun**. Isinya dijumlahkan otomatis dari semua
pesanan pada trip tersebut. Kalau admin di Jakarta mengubah jumlah pesanan saat
kamu sedang di toko, daftar ini ikut berubah begitu halaman disegarkan.

| Kolom | Artinya |
|---|---|
| Produk | Nama barang, jumlah customer yang memesan, dan harga modal perkiraan |
| Toko | Toko langganan, kalau sudah diisi di data produk |
| **Dipesan** | Unit dari order yang **DP-nya sudah masuk** — inilah yang harus dibeli |
| **Menunggu DP** | Unit dari order yang DP-nya belum diverifikasi — **jangan dibeli dulu** |
| Dibeli | Yang sudah kamu catat pembeliannya |
| Sisa | Yang masih harus dibeli |

Baris yang sudah selesai diberi label hijau **Lengkap** dan diredupkan, supaya
matamu langsung tertuju ke yang belum beres.

### Kenapa ada dua kolom jumlah

Yang dibelanjakan hanya kolom **Dipesan**. Order yang DP-nya belum masuk sengaja
dipisah ke kolom **Menunggu DP** dan tidak ikut dihitung.

Alasannya soal uang. Membeli barang untuk order yang DP-nya belum ada berarti
menalangi pembelian dengan uang toko. Kalau customer itu akhirnya batal, barang
tersebut mengendap di stok tanpa ada yang membayar — sementara uangmu sudah
keluar di Tokyo.

Kolom **Menunggu DP** tetap ditampilkan supaya kamu bisa menilai sendiri: kalau
angkanya besar, hubungi admin agar DP-nya ditagih dulu selagi kamu masih di dekat
toko itu.

Selama masih ada yang menunggu, muncul kotak kuning di atas daftar. Itu penanda
supaya angka **Dipesan** yang terlihat kecil tidak dikira pesanannya memang
sedikit.

> Pada contoh di atas: Tokyo Banana **dipesan 1**, tapi ada **3 lagi menunggu
> DP**. Yang wajib dibeli sekarang cuma 1.

Pilih trip lewat dropdown di kanan atas. Aplikasi otomatis memilih trip yang
sedang berjalan, jadi biasanya kamu tidak perlu memilih apa pun.

## Mencatat pembelian

Tekan **Catat Beli** pada baris produknya.

{{img:21-belanja-catat}}

**Langkahnya:**

1. **Jumlah dibeli** sudah terisi otomatis sebanyak sisa yang belum dibeli.
   Ubah kalau kamu membeli lebih atau kurang.
2. **Harga satuan** isi dengan harga yang **benar-benar kamu bayar** di kasir,
   dalam mata uang setempat.
3. **Tanggal beli** otomatis hari ini.
4. **Kurs khusus** boleh dikosongkan. Isi hanya kalau kurs hari itu berbeda jauh
   dari kurs trip.
5. Isi **Toko** dan **Catatan** kalau perlu.
6. Tekan **Catat Pembelian**.

> **Isi harga yang sesungguhnya, termasuk diskon yang kamu dapat.** Inilah angka
> yang dipakai menghitung laba. Kalau kamu mengisi harga daftar padahal dapat
> diskon, laporan labamu akan lebih kecil dari kenyataan.

## Yang terjadi setelah disimpan

Aplikasi langsung membagi barangnya. Pesanan yang masuk lebih dulu dilayani
lebih dulu, dan sisanya masuk stok.

```
   Beli 6 unit Lotion
        │
        ├── Order Rina (pesan 15 Agu, butuh 2)  ──▶  2 unit
        ├── Order Sari (pesan 15 Agu, butuh 1)  ──▶  1 unit
        │                                            ───────
        │                       masuk ke pesanan:      3
        │
        └── Tidak ada yang memesan sisanya       ──▶  3 unit
                                    masuk stok:       3
```

Notifikasi hijau akan memberi tahu pembagiannya, misalnya *"3 unit untuk
pesanan, 3 unit masuk stok"*.

Kalau kamu membeli lebih sedikit dari yang dipesan (toko kehabisan), pesanan
paling awal yang dipenuhi dan sisanya tetap menunggu.

---

# 5. Pembelian

Menu **Pembelian** adalah catatan semua belanja, dari semua trip.

{{img:22-pembelian}}

Gunakan halaman ini untuk memeriksa ke mana perginya setiap unit yang kamu beli.

Dua label di kolom Qty menunjukkan pembagiannya:

| Label | Artinya |
|---|---|
| **N pesanan** | N unit dipakai memenuhi order customer |
| **N stok** | N unit tidak dipesan siapa pun dan masuk gudang |

Tekan tanda panah di kolom paling kiri untuk membuka rinciannya. Kamu akan
melihat order dan nama customer yang menerima tiap unit.

## Menghapus pembelian

Kalau ada salah catat, tekan ikon tempat sampah. Menghapus akan membatalkan
seluruh dampaknya sekaligus: alokasi ke pesanan dilepas, dan stok yang sempat
bertambah ditarik kembali.

Kalau stoknya sudah terlanjur terjual, penghapusan akan gagal. Perbaiki lewat
penyesuaian stok saja (lihat bab Stok).

---

# 6. Order

Bab paling panjang, karena inilah pekerjaan harian.

## 6.1 Melihat daftar order

Buka menu **Order**.

{{img:30-order-list}}

Kolom **Sisa Bayar** berwarna kuning kalau masih ada tagihan. Centang **Belum
lunas** untuk menyaring hanya order yang perlu ditagih.

Kolom **Channel** menunjukkan order itu masuk dari mana. Dropdown **Semua
channel** menyaringnya, misalnya untuk melihat berapa banyak order yang datang
dari Instagram bulan ini.

### Melihat total dalam mata uang trip

Centang **Mata uang trip**, maka kolom Total dan Sisa Bayar berubah dari rupiah
ke mata uang trip masing-masing order.

{{img:36-order-mata-uang}}

Order dari trip Jepang tampil dalam JPY, order dari trip Korea dalam KRW,
memakai kurs yang dikunci saat trip dibuat. Berguna saat kamu memegang nota dari
toko di sana dan ingin mencocokkan angkanya tanpa menghitung manual.

Hilangkan centangnya untuk kembali ke rupiah. Data di database tidak berubah
sama sekali; ini hanya cara menampilkan.

## 6.2 Mencatat order baru

Tekan **Catat Order**.

{{img:31-order-baru}}

Formnya terbagi tiga bagian.

### Bagian 1 — Trip & customer

1. Pilih **Trip** lebih dulu. Daftar produk masih kosong sebelum ini diisi,
   karena harganya diambil dari katalog trip tersebut.
2. Pilih **Customer**.
3. **Tanggal order** otomatis hari ini.
4. Pilih **Asal order** — dari channel mana pesanan ini masuk. Pilihannya
   WhatsApp, Instagram, TikTok, Marketplace, dan Lainnya. Bawaannya WhatsApp.

> Isi asal order sambil mengetik pesanannya, selagi masih ingat. Kalau diisi
> asal-asalan, laporan Per Channel jadi tidak bisa dipercaya dan kamu kehilangan
> satu-satunya cara mengetahui promosi di channel mana yang benar-benar
> menghasilkan.

Kalau ternyata salah pilih, bisa diperbaiki lewat tombol **Ubah Order** di
halaman detail.

> Kalau kamu mengganti trip di tengah pengisian, produk yang sudah dipilih akan
> dikosongkan. Itu wajar, karena harga di trip lain berbeda.

### Bagian 2 — Produk yang dipesan

5. Pilih produk dari dropdown **Tambah dari katalog trip**. Produk langsung
   masuk ke tabel dengan harga katalognya.
6. Ubah **Qty** sesuai pesanan customer.
7. **Harga** boleh kamu ubah kalau memberi harga khusus.
8. Untuk menghapus baris, tekan ikon tempat sampah di kanan.

Memilih produk yang sama dua kali akan menambah jumlahnya, bukan membuat baris
kembar.

### Bagian 3 — Alamat pengiriman

Secara bawaan, barang dikirim ke alamat customer yang tersimpan. Alamat itu
ditampilkan supaya kamu bisa memastikan.

9. Centang **Kirim ke alamat lain** kalau barangnya dikirim ke tempat lain,
   misalnya hadiah atau alamat kantor. Alamat customer disalin sebagai titik
   awal, jadi kamu tinggal mengubah bagian yang berbeda.

### Panel ringkasan di kanan

10. Isi **Diskon** dan **Ongkir ditagihkan** kalau ada.
11. **DP diminta** boleh dikosongkan; aplikasi otomatis memakai 50% dari total.
12. Isi **Catatan** untuk permintaan khusus customer.
13. Tekan **Buat Order**.

Order yang baru dibuat langsung berstatus **Menunggu DP**, jadi kamu bisa
langsung menagih uang mukanya tanpa mengubah status dulu.

Angka **Total** di panel kanan ikut berubah setiap kali kamu mengubah isi
pesanan, jadi kamu bisa langsung menyebutkan totalnya ke customer.

## 6.3 Halaman detail order

Setelah dibuat, kamu masuk ke halaman detail order. Di sinilah hampir semua
pekerjaan berikutnya dilakukan.

{{img:32-order-detail}}

Halamannya terbagi dua kolom:

| Kolom kiri | Kolom kanan |
|---|---|
| Item pesanan | Ringkasan biaya |
| Pembayaran | Data customer dan alamat kirim |
| Invoice | Pengiriman |

Tombol di kanan atas berubah mengikuti status order, dan hanya menampilkan
langkah berikutnya yang masuk akal.

Setiap tombol status memunculkan **kotak konfirmasi** berisi penjelasan apa yang
akan terjadi. Biasakan membacanya sebentar: perpindahan status tidak punya
tombol "urungkan", dan sebagian di antaranya mengunci order.

> Tombol **Dikirim** berwarna merah karena setelah ditekan, isi order tidak bisa
> diubah lagi selamanya. Pastikan jumlah dan harganya sudah benar.

## 6.4 Meminta DP

Order baru langsung berstatus **Menunggu DP**, jadi bisa ditagih saat itu juga:

1. Di panel Pembayaran, tekan **Tagih DP**.
2. Pesan WhatsApp yang sudah terisi lengkap akan muncul. Tekan **Buka WhatsApp**
   lalu kirim dari nomormu sendiri.

Kalau ternyata pesanannya belum pasti, tekan tombol **Draft** di kanan atas untuk
mengeluarkannya dari antrean penagihan.

## 6.5 Mencatat pembayaran

Setelah customer transfer, tekan **Catat Bayar** di panel Pembayaran.

{{img:33-order-bayar}}

**Langkahnya:**

1. Pilih **Jenis**: `DP` untuk uang muka, `Pelunasan` untuk sisanya, `Refund`
   kalau uang dikembalikan.
2. **Nominal** sudah terisi otomatis dengan sisa yang kurang. Ubah kalau
   customer membayar sebagian.
3. Pilih **Metode** pembayaran.
4. Isi **Referensi** dengan nama pengirim atau nomor transaksi, supaya mudah
   dicocokkan dengan mutasi rekening.
5. Tekan **Unggah bukti**, lalu pilih foto atau PDF struk transfer dari customer.
   Berkas gambar langsung menampilkan pratinjau kecil — pastikan itu memang
   tangkapan layar yang benar sebelum lanjut.
6. Tekan **Simpan**.

Bukti yang sudah terunggah tetap bisa dibuka dari daftar pembayaran lewat tautan
**Lihat bukti transfer**, jadi masih bisa dicocokkan ulang dengan mutasi rekening
kalau nanti ada selisih.

**Status order berubah sendiri** mengikuti uang yang masuk:

| Kondisi | Status order menjadi |
|---|---|
| DP mencapai nominal yang diminta | **Diproses** |
| Sisa tagihan menjadi nol | **Lunas** |

## 6.6 Mengubah jumlah pesanan

Ini yang paling sering terjadi: customer mengirim pesan *"boleh tambah satu
lagi tidak?"*.

Tekan ikon pensil pada baris item yang ingin diubah. Kolom Qty dan Harga berubah
menjadi kotak isian.

{{img:35-order-edit-item}}

1. Ubah angkanya.
2. Tekan ikon centang untuk menyimpan, atau silang untuk membatalkan.

Total order, sisa tagihan, dan daftar belanja tripper langsung ikut menyesuaikan.

**Aturan yang perlu diketahui:**

| Aturan | Kenapa begitu |
|---|---|
| Order yang sudah **Dikirim** tidak bisa diubah lagi | Paketnya sudah jalan, dokumen harus sesuai kenyataan |
| Jumlah tidak boleh kurang dari yang sudah diterima | Barang yang sudah ada di gudang tidak bisa dianggap tidak ada |
| Kalau jumlah dikurangi padahal barangnya sudah dibeli, kelebihannya masuk stok | Barangnya nyata, tidak boleh hilang dari pembukuan |
| Order harus punya minimal satu item | Kalau memang batal, batalkan ordernya |

Kalau order sudah lunas lalu kamu menambah item, statusnya otomatis turun
kembali ke **Ditagihkan** supaya sisa tagihan yang baru terlihat jelas.

Semua perubahan tercatat di Jejak Perubahan, lengkap dengan siapa yang mengubah.

## 6.7 Mencocokkan barang yang datang

Setelah tripper pulang, cocokkan barangnya dengan pesanan. Tekan **Cocokkan
Barang** di kanan atas.

{{img:34-order-cocokkan}}

1. Semua baris otomatis dianggap diterima penuh, karena itu yang paling sering
   terjadi.
2. Ubah hanya baris yang bermasalah. Isi **Diterima** sesuai jumlah yang benar
   ada.
3. Kolom **Status** ikut menyesuaikan sendiri:
   - jumlah penuh → **Lengkap**
   - sebagian → **Sebagian**
   - nol → **Tidak tersedia**
4. Tekan **Simpan Penerimaan**.

Order otomatis berpindah ke status **Barang Tiba**.

Barang yang ditandai tidak tersedia tetap muncul di invoice dengan keterangan
tersebut, supaya customer paham kenapa totalnya berbeda dari pesanan awal.

## 6.8 Membatalkan order

Tekan **Batalkan**, isi alasannya, lalu konfirmasi.

Barang yang sudah terlanjur dibeli untuk order itu otomatis dipindahkan ke stok.
**Uang yang sudah diterima tidak otomatis kembali** — kamu perlu mencatatnya
sebagai pembayaran jenis **Refund**, karena uangnya juga harus benar-benar keluar
dari rekeningmu.

---

# 7. Invoice

## 7.1 Menerbitkan invoice

Dari halaman detail order, di panel **Invoice**, tekan **Terbitkan**.

1. Pilih jenis:
   - **Pelunasan** untuk menagih seluruh nilai order dikurangi yang sudah
     dibayar. Ini yang paling sering dipakai.
   - **DP** kalau kamu ingin membuat dokumen resmi untuk uang muka.
2. **Jatuh tempo** boleh dikosongkan; aplikasi memakai bawaan dari Pengaturan.
3. Tekan **Terbitkan**.

> **Nominalnya disalin saat invoice dibuat.** Kalau ordernya kamu edit setelah
> ini, invoice yang sudah dilihat customer tidak ikut berubah. Kalau perubahannya
> penting, batalkan invoice lama lalu terbitkan yang baru.

## 7.2 Daftar invoice

Menu **Invoice** menampilkan semua tagihan yang pernah diterbitkan.

{{img:40-invoice-list}}

Invoice yang melewati jatuh tempo dan belum lunas ditandai **lewat tempo**
berwarna merah.

Tiga tombol di kolom Aksi:

- **Ikon dokumen** membuka PDF invoice di tab baru.
- **Ikon chat hijau** membuka pesan siap kirim.
- **Ikon larangan merah** membatalkan invoice.

Dua tombol terakhir tidak muncul untuk invoice yang sudah lunas atau sudah
dibatalkan.

### Membatalkan invoice

Dipakai kalau invoice terlanjur diterbitkan dengan angka yang salah — misalnya
customer menambah barang setelah ditagih, atau jumlahnya dikoreksi.

1. Tekan **ikon larangan merah** di baris invoicenya. Di halaman detail order,
   tombolnya bertulisan **Batalkan**.
2. Baca kotak konfirmasinya, lalu tekan **Ya, batalkan invoice**.

Statusnya berubah menjadi **Dibatalkan** dan invoice itu berhenti dihitung
sebagai tagihan berjalan.

Yang **tidak** berubah: pembayaran yang sudah tercatat tetap utuh, status order
tetap di tempatnya, dan PDF lamanya masih bisa dibuka sebagai arsip. Karena
status order tidak bergeser, kamu bisa langsung menekan **Terbitkan** untuk
membuat invoice penggantinya.

> Invoice yang sudah **lunas** tidak bisa dibatalkan, jadi tombolnya memang tidak
> muncul. Uangnya sudah tercatat masuk; kalau perlu dikembalikan, catat sebagai
> **Refund** di panel Pembayaran.

## 7.3 Mengirim invoice ke customer

Tekan ikon chat hijau.

{{img:41-invoice-wa}}

Pesannya sudah terisi lengkap: nama customer, nomor invoice, total, yang sudah
dibayar, sisa tagihan, dan nomor rekening.

**Dua pilihan:**

- **Salin teks** — menyalin ke clipboard, kalau kamu mau menempelkannya sendiri.
- **Buka WhatsApp** — membuka WhatsApp dengan pesan sudah terketik. Kamu tinggal
  menekan kirim.

> Pesan dikirim dari nomormu sendiri, bukan dari nomor gateway. Customer melihat
> pesan dari kontak yang sudah mereka kenal, jadi tidak dikira spam.

Begitu kamu menekan **Buka WhatsApp**, invoice otomatis ditandai sudah dikirim.

Teks pesannya bisa kamu ubah di **Pengaturan → Template Pesan**.

---

# 8. Pengiriman

Tiga langkah: kemas, input resi, kabari customer.

## 8.1 Menandai sudah dikemas

Dari halaman detail order, di panel **Pengiriman**, tekan **Tandai Dikemas**.

{{img:43-kemas-ongkir}}

1. **Kurir** otomatis JNE.
2. Pilih **Layanan**: REG, YES, OKE, atau JTR.
3. Isi **Berat** dalam gram, sesuai timbangan.
4. Isi **Dimensi paket**: panjang, lebar, dan tinggi kardus dalam sentimeter.
   Boleh dikosongkan kalau paketnya kecil dan padat.
5. Isi **Catatan kemasan** kalau ada perlakuan khusus.
6. Tekan **Simpan**.

## 8.1.1 Menghitung perkiraan ongkir

Setelah berat dan dimensi terisi, tekan **Hitung Ongkir ke [nama kota]**. Kota
tujuannya diambil otomatis dari alamat kirim order, jadi tidak perlu diketik.

Hasilnya muncul di bawah tombol, seperti pada gambar di atas.

### Kenapa perlu ukuran kardus

Ekspedisi tidak hanya menagih berdasarkan berat. Kardus besar memakan ruang di
truk, jadi kurir menghitung juga **berat volume**, lalu menagih mana yang lebih
besar di antara keduanya.

```
                          panjang × lebar × tinggi (cm)
   berat volume (kg)  =  ──────────────────────────────
                                    6000

   yang ditagih       =  berat asli atau berat volume, ambil yang terbesar,
                         lalu dibulatkan KE ATAS ke kilogram penuh
```

**Contoh pada gambar di atas.** Kardus 40 × 30 × 25 cm berisi baju, beratnya
cuma 800 gram.

```
   berat volume  =  (40 × 30 × 25) ÷ 6000  =  5 kg
   berat asli    =  0,8 kg
   ditagih       =  5 kg  (yang terbesar)
   ongkir        =  5 × Rp28.000  =  Rp140.000
```

Kalau hanya melihat timbangan, kamu akan mengira ongkirnya Rp28.000 saja.
Selisih Rp112.000 itu yang akhirnya kamu tanggung sendiri kalau ongkir ke
customer terlanjur ditagih kekecilan.

Sebaliknya, skincare seberat 2,3 kg dalam kardus kecil dibulatkan naik jadi
3 kg, karena ekspedisi tidak menjual setengah kilo.

> Kalau muncul peringatan kuning "tarif kota ini belum ada", artinya kota tujuan
> belum terdaftar dan aplikasi memakai tarif umum. Angkanya tetap bisa dipakai,
> tapi tambahkan tarif kota itu di **Pengaturan → Ongkir** supaya lain kali
> lebih tepat.

Hasil perkiraannya ikut tersimpan saat kamu menekan **Simpan**, dan muncul di
panel Pengiriman sebagai **Estimasi ongkir**.

## 8.2 Input nomor resi

Setelah paket diserahkan ke JNE, tekan **Input Resi & Kirim**.

{{img:42-kirim-resi}}

1. Isi **Nomor resi** dari struk JNE.
2. Isi **Ongkir dibayar** — ini yang **kamu bayar ke JNE**, bukan yang kamu
   tagihkan ke customer. Keduanya dicatat terpisah, dan selisihnya masuk ke
   laporan laba.
   Kalau tadi sudah dihitung perkiraannya, ada tautan **Pakai estimasi
   Rp…** di bawah kolom ini untuk mengisinya sekali tekan. Ganti dengan angka
   di struk kalau ternyata berbeda.
3. **Tanggal kirim** otomatis hari ini.
4. Tekan **Simpan Resi**.

### Kalau order belum lunas

Aplikasi akan menolak, dan menampilkan sisa tagihannya seperti pada gambar di
atas. Ini penjagaan yang disengaja supaya tidak ada paket terkirim karena admin
lupa mengecek.

Kalau memang ingin tetap mengirim, centang **Kirim walau belum lunas**. Pakai
hanya untuk pelanggan lama yang kamu percaya. Tindakan ini tercatat di Jejak
Perubahan.

> Setelah resi tersimpan, isi order **tidak bisa diubah lagi**. Pastikan
> jumlahnya sudah benar sebelum menekan Simpan Resi.

## 8.3 Mengabari customer

Tekan **Kabari Customer**. Pesan berisi kurir, layanan, nomor resi, dan tautan
pelacakan JNE akan muncul. Kirim seperti biasa lewat WhatsApp.

## 8.4 Daftar pengiriman

Menu **Pengiriman** menampilkan semua paket.

{{img:50-pengiriman}}

Perhatikan tulisan kuning kecil **belum dikabari** di kolom Dikirim. Itu penanda
paket yang sudah jalan tapi customernya belum diberi tahu nomor resinya.

Kolom **Ongkir** menampilkan ongkir yang benar-benar dibayar. Untuk paket yang
belum diserahkan ke kurir, yang tampil adalah hasil perkiraan dengan keterangan
kecil **estimasi** di bawahnya — jangan dipakai sebagai angka final saat merekap
biaya.

## 8.5 Menutup order

Setelah customer mengonfirmasi barang diterima, tekan **Tandai Diterima** di
halaman order, lalu setujui kotak konfirmasinya. Statusnya berubah menjadi
**Selesai**.

> Tekan ini setelah customer benar-benar mengabari, bukan sekadar karena resi
> JNE sudah menunjukkan terkirim. Paket bisa saja tercatat terkirim padahal
> diterima tetangga.

---

# 9. Siap Kemas

Menu **Siap Kemas** adalah daftar kerja gudang. Gunanya menjawab: *hari ini saya
harus mengerjakan apa?*

{{img:51-siap-kemas}}

Pilih tahapnya lewat dropdown di kiri atas:

```
   Siap dikemas ──▶ Sudah dikemas ──▶ Menunggu ──▶ Siap dikirim
        │                 │            pelunasan        │
     kemas            terbitkan          │          input resi
                       invoice        tagih
```

| Tahap | Yang harus dikerjakan |
|---|---|
| **Siap dikemas** | Barang sudah dicocokkan, tinggal dikemas per customer |
| **Sudah dikemas** | Terbitkan invoice pelunasan |
| **Menunggu pelunasan** | Tunggu atau tagih pembayarannya |
| **Siap dikirim** | Input nomor resi JNE |

Tombol **Proses** di kanan tiap baris langsung membuka halaman ordernya.

Saring per trip kalau satu kiriman baru datang dan kamu ingin membereskan
antrean trip itu sekaligus.

---

# 10. Customer

## 10.1 Daftar customer

Buka menu **Customer**.

{{img:60-customer}}

Cari berdasarkan nama, nomor WA, atau kode customer. Nomor teleponnya bisa
diklik untuk langsung membuka WhatsApp.

## 10.2 Menambah customer

Tekan **Tambah Customer**.

{{img:61-customer-form}}

**Langkahnya:**

1. Isi **Nama** dan **Nomor WhatsApp** (wajib).
2. Isi **Email**, **Instagram** kalau ada.
3. Isi **Alamat**, **Kota**, **Provinsi**, dan **Kode Pos**.
4. Isi **Catatan** untuk hal seperti patokan alamat atau preferensi packing.
5. Tekan **Simpan**.

Kode customer (CUS-0001) dibuatkan otomatis.

### Satu nomor untuk satu customer

Nomor WhatsApp tidak boleh dipakai dua customer. Kalau kamu menyimpan nomor yang
sudah terdaftar — dalam format apa pun — aplikasi menolaknya.

Ini penting untuk laporan: rekap penjualan per customer dikelompokkan berdasarkan
nomor WhatsApp. Kalau satu orang tercatat dua kali karena beda ejaan nama,
belanjanya terbelah jadi dua baris dan dia tidak akan terlihat sebagai pelanggan
besar padahal sebenarnya iya.

Kalau nomornya ditolak, cari customer yang sudah ada lewat kolom pencarian, lalu
pakai data itu.

### Nomor telepon boleh ditulis bebas

Aplikasi merapikannya sendiri ke format internasional:

| Kamu ketik | Disimpan jadi |
|---|---|
| `081234567890` | `6281234567890` |
| `0812-3456-7890` | `6281234567890` |
| `+62 812 3456 7890` | `6281234567890` |
| `(0812) 3456-7890` | `6281234567890` |

Nomor luar negeri yang diawali `+` dibiarkan apa adanya.

> **Isi alamatnya sejak awal.** Order tidak bisa dibuat tanpa alamat kirim.
> Kalau alamatnya sudah tersimpan di data customer, kamu tidak perlu
> mengetiknya ulang tiap kali membuat order.

## 10.3 Menghapus customer

Menghapus hanya menyembunyikan customer dari daftar. Riwayat ordernya tetap
tersimpan, karena order lama masih menunjuk ke data tersebut.

---

# 11. Produk

Menu **Produk** adalah katalog induk. Ini daftar rujukan, **bukan** daftar
harga jual — harga jual ditentukan per trip, karena kurs dan harga toko berubah
tiap perjalanan.

{{img:62-produk}}

## Menambah produk

Tekan **Tambah Produk**.

{{img:63-produk-form}}

**Langkahnya:**

1. Isi **Nama produk** lengkap dengan ukuran atau varian, misalnya
   "Hada Labo Gokujyun Lotion 170ml".
2. **SKU** boleh dikosongkan, nanti dibuatkan otomatis.
3. Pilih **Kategori**.
4. Isi **Brand** dan **Toko langganan**. Toko langganan sangat membantu tripper
   mencari barangnya di lapangan.
5. Isi **Mata uang** dan **Harga modal** sebagai acuan.
6. Pilih **Jenis markup** dan **Markup** bawaan.
7. Isi **Berat** dalam gram untuk memperkirakan ongkir.
8. Tekan **Simpan**.

> Harga modal dan markup di sini hanya **nilai bawaan**. Saat produk dimasukkan
> ke katalog sebuah trip, kedua angka itu terisi otomatis tapi tetap bisa kamu
> ubah sesuai harga toko saat itu.

## Mengelola kategori

Tekan **Kategori** di kanan atas. Dialog kecil akan muncul untuk menambah,
mengubah nama, dan menghapus kategori. Kategori yang masih dipakai produk tidak
bisa dihapus.

## Melihat riwayat harga antar trip

Tekan **ikon jam** di baris produk untuk melihat harga produk itu pada setiap
trip yang pernah memuatnya.

{{img:66-riwayat-harga}}

| Kolom | Artinya |
|---|---|
| Trip | Kode trip, negara, tanggal berangkat, dan kurs yang berlaku saat itu |
| Katalog | Harga modal yang kamu isi saat menyusun katalog trip itu |
| Beli riil | Rata-rata harga yang benar-benar dibayar di kasir |
| Harga jual | Harga yang diumumkan ke customer |
| Dibeli / Terjual | Unit yang dibeli dan yang laku pada trip itu |

**Cara membacanya:** bandingkan kolom **Katalog** dengan **Beli riil**. Kalau
harga di kasir terus-menerus lebih tinggi daripada yang kamu pasang di katalog,
berarti markup-mu diam-diam terpotong tiap trip — naikkan harga katalognya, bukan
markup-nya.

Riwayat yang sama juga muncul otomatis sebagai keterangan satu baris saat kamu
menambahkan produk ke katalog trip baru, lengkap dengan tautan **Pakai harga
ini**.

> Tautan itu hanya muncul kalau trip sebelumnya memakai **mata uang yang sama**.
> Produk yang sama dibeli di Korea tidak bisa disalin harganya dari trip Jepang —
> 880 JPY dan 880 KRW jauh berbeda nilainya.

## Menghapus produk

Menghapus akan **menonaktifkan** produk, bukan melenyapkannya. Produk itu tidak
muncul lagi saat menyusun katalog, tapi semua order dan laporan lama tetap utuh.

---

# 12. Stok

Menu **Stok** berisi barang milikmu yang tidak dipesan siapa pun.

{{img:64-stok}}

## Dari mana stok berasal

```
   Beli 6 unit  ──▶  3 unit ke pesanan customer
                     3 unit tidak bermilik
                          │
                          ▼
                       S T O K
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
   dijual di        dilepas dari       dikoreksi lewat
   marketplace      order yang         stock opname
                    dikurangi
```

## Harga pokok rata-rata

Tiap produk punya satu harga pokok rata-rata yang dihitung ulang setiap stok
bertambah:

```
   Stok lama : 3 unit @ Rp100.000 = Rp300.000
   Stok masuk: 2 unit @ Rp110.000 = Rp220.000
                                    ──────────
   Sekarang  : 5 unit                Rp520.000

   Rata-rata baru = Rp520.000 ÷ 5 = Rp104.000 per unit
```

Dipakai rata-rata karena barang jastip adalah unit yang identik, dan tidak ada
yang menempeli label trip pada tiap botol.

## Menjual stok di marketplace

Tekan **Jual** pada baris produknya.

{{img:65-stok-jual}}

1. Isi **Jumlah terjual**.
2. Isi **Harga jual per pcs**. Margin per unit langsung dihitung dan ditampilkan
   di bawah kolom isian.
3. Isi **Kanal penjualan**, misalnya Shopee atau Tokopedia.
4. Tekan **Catat Penjualan**.

Penjualan stok dicatat terpisah dari laba trip. Laporan trip hanya mencakup
barang yang dipesan customer.

## Menyesuaikan stok

Tekan ikon penggeser di kanan baris. Isi **Jumlah sebenarnya** sesuai hasil
hitung fisik, dan tulis alasannya. Selisihnya tercatat di Riwayat Pergerakan.

## Riwayat pergerakan

Tab **Riwayat Pergerakan** mencatat setiap perubahan stok: masuk dari belanja,
dipakai pesanan, terjual marketplace, dan penyesuaian.

---

# 13. Laporan

Menu **Laporan** punya lima tab. Tab **Profit per Order** hanya muncul untuk
owner; sisanya bisa dibuka admin juga.

## 13.1 Piutang

Daftar order yang masih punya sisa tagihan, diurutkan dari yang paling lama
menunggu.

{{img:70-laporan-piutang}}

Kolom **Umur** menunjukkan berapa hari sejak order dibuat. Yang lewat 14 hari
ditandai merah.

Tautan **Tagih via WA** di bawah nama customer langsung membuka WhatsApp ke
nomor tersebut.

> **Kerjakan daftar ini dari atas.** Utang yang paling lama adalah yang paling
> sulit ditagih.

## 13.2 Profit per Order

Margin tiap order. Hanya owner yang bisa melihatnya.

{{img:71-laporan-profit}}

| Kolom | Artinya |
|---|---|
| Omzet | Total order |
| HPP | Biaya belanja yang benar-benar dialokasikan ke order ini |
| Profit | Omzet dikurangi HPP |
| Margin | Persentase profit terhadap omzet |

Order yang HPP-nya masih Rp0 berarti pembeliannya belum diinput — itu data yang
belum lengkap, bukan berarti untungnya 100%.

Profit negatif biasanya berarti harga di toko naik setelah harga jual terlanjur
diumumkan.

## 13.3 Performa Produk

Produk mana yang paling laku dan paling menguntungkan.

{{img:72-laporan-produk}}

Pakai laporan ini untuk memutuskan barang apa yang layak dibawa trip berikutnya,
dan di mana markup-mu terlalu tipis.

## 13.4 Per Customer

Siapa yang paling banyak belanja, diurutkan dari yang terbesar.

{{img:73-laporan-customer}}

| Kolom | Artinya |
|---|---|
| Customer | Nama, kode, kota, dan tanggal order terakhirnya |
| Order | Jumlah order yang tidak dibatalkan |
| Pcs | Total barang dari order-order itu |
| Omzet | Jumlah seluruh total ordernya |
| Rata-rata | Omzet dibagi jumlah order |
| Profit | Omzet dikurangi HPP riil |
| Piutang | Yang masih ditunggu dari seluruh ordernya |

Dropdown di kiri atas menyaring per trip. Kosongkan untuk melihat angka
sepanjang masa.

**Dua cara memakainya.** Baca dari atas saat harus menentukan siapa yang
didahulukan — misalnya slot trip terbatas atau produk berkuota. Lalu lihat kolom
**Piutang**: customer yang sering belanja tapi juga sering nunggak adalah
persoalan yang berbeda dengan yang murni belanja banyak.

> Customer beromzet tinggi tapi profitnya kecil berarti yang dia beli
> produk-produk bermargin tipis. Tahu ini sebelum memberi dia diskon.

## 13.5 Per Channel

Order datang dari mana, berdasarkan isian **Asal order** saat pencatatan.

{{img:74-laporan-channel}}

| Kolom | Artinya |
|---|---|
| Channel | WhatsApp, Instagram, TikTok, Marketplace, atau Lainnya |
| Order | Jumlah order dari channel itu |
| Customer | Jumlah customer berbeda yang order lewat channel itu |
| Omzet | Jumlah total ordernya |
| Rata-rata | Omzet dibagi jumlah order |
| Profit | Omzet dikurangi HPP riil |
| Porsi omzet | Persentase terhadap total omzet, digambar sebagai bar |

Porsi seluruh channel selalu berjumlah 100%, jadi panjang bar-nya bisa langsung
dibandingkan.

Bacalah sambil mengingat berapa banyak tenaga yang dihabiskan tiap channel.
Channel yang ordernya sedikit tapi nilai **Rata-rata**-nya tinggi bisa jadi
lebih menguntungkan daripada channel ramai yang ordernya kecil-kecil.

> Order yang tercatat sebelum kolom asal order ada dihitung sebagai WhatsApp,
> jadi angka awalnya condong ke sana sampai cukup banyak order baru masuk.

## Mengunduh ke Excel

Tiap tab punya tombol **Ekspor CSV** di kanan atas. Berkasnya terbuka rapi di
Excel, termasuk huruf beraksen pada nama customer.

---

# 14. Pengaturan

Menu **Pengaturan** hanya bisa dibuka owner. Ada lima tab.

## 14.1 Toko

Data yang muncul di invoice PDF dan pesan ke customer.

{{img:80-pengaturan-toko}}

| Isian | Muncul di |
|---|---|
| Nama toko | Kepala invoice dan pesan |
| Nomor WA toko | Invoice |
| Email toko | Invoice |
| Alamat toko | Invoice |
| **Rekening pembayaran** | Invoice dan **setiap** permintaan pembayaran |
| Catatan penutup invoice | Bagian bawah invoice |
| Jatuh tempo invoice | Bawaan untuk invoice baru, dalam hari |

> **Periksa dua kali nomor rekeningnya.** Nomor itu muncul di setiap invoice dan
> setiap pesan penagihan. Salah satu angka saja, uang customer tidak akan sampai.

Tekan **Simpan** setelah mengubah.

## 14.2 Template Pesan

Teks pesan WhatsApp yang dipakai aplikasi.

{{img:81-pengaturan-template}}

Ada tiga template: permintaan DP, penagihan pelunasan, dan informasi pengiriman.

Tulisan dalam kurung ganda seperti `{{customer_name}}` disebut *placeholder*.
Aplikasi menggantinya dengan data sungguhan saat pesan dibentuk.

**Placeholder yang bisa dipakai:**

| Placeholder | Diganti jadi |
|---|---|
| `{{customer_name}}` | Nama customer |
| `{{store_name}}` | Nama tokomu |
| `{{trip_title}}` | Judul trip |
| `{{order_number}}` | Nomor order |
| `{{invoice_number}}` | Nomor invoice |
| `{{total}}` | Total, misalnya Rp335.000 |
| `{{dp_amount}}` | Nominal DP |
| `{{amount_paid}}` | Yang sudah dibayar |
| `{{amount_due}}` | Sisa tagihan |
| `{{due_date}}` | Jatuh tempo |
| `{{bank_account}}` | Nomor rekening dari tab Toko |
| `{{courier}}` | Nama kurir |
| `{{service}}` | Jenis layanan |
| `{{tracking_number}}` | Nomor resi |

Kalau kamu salah mengetik nama placeholder, tulisannya akan muncul apa adanya di
pesan. Jadi salah ketik langsung ketahuan saat kamu membaca pratinjaunya, bukan
hilang diam-diam.

## 14.3 Ongkir

Dua hal diatur di sini: cara ongkir dihitung, dan tabel tarif per kota.

{{img:84-pengaturan-ongkir}}

### Perhitungan ongkir

| Isian | Bawaan | Artinya |
|---|---|---|
| Pembagi berat volume | 6000 | Angka pembagi untuk mengubah ukuran kardus jadi kilogram |
| Tarif cadangan per kg | Rp25.000 | Dipakai kalau kota tujuan belum ada di tabel |

JNE, SiCepat, dan J&T sama-sama memakai 6000 untuk kiriman dalam negeri. Jangan
diubah kecuali ketentuan ekspedisimu memang berbeda.

### Tabel tarif per kota

Setiap baris berisi satu kombinasi kota, kurir, dan layanan. Saat menghitung
perkiraan, kota kirim pada order dicari di tabel ini dulu; kalau tidak ketemu,
barulah tarif cadangan dipakai.

**Menambah tarif:**

1. Tekan **Tambah Tarif**.
2. Isi **Kota tujuan**, misalnya `Bandung`. Huruf besar-kecil tidak berpengaruh.
3. Isi **Provinsi** kalau perlu, untuk membedakan kota bernama mirip.
4. Pilih **Layanan** dan isi **Tarif per kg** sesuai daftar harga ekspedisi.
5. **Berat minimum** biasanya 1000 gram, karena ekspedisi menagih minimal 1 kg.
6. **Estimasi tiba** ditulis lengkap dengan satuannya, misalnya `2-3 hari`.
7. Tekan **Simpan**.

> Menyimpan kota yang sudah ada untuk kurir dan layanan yang sama akan
> **memperbarui** baris itu, bukan menggandakannya. Jadi untuk menaikkan tarif,
> cukup masukkan ulang kotanya dengan harga baru.

Kolom pencarian di atas tabel membantu menemukan kota dengan cepat.

Untuk menghapus, tekan ikon tempat sampah merah di kanan baris. Order ke kota
itu selanjutnya akan memakai tarif cadangan.

Aplikasi sudah berisi tarif kota-kota yang paling sering dikirimi jastip.
Tambahkan kota lain begitu ketemu — perkiraan yang memakai tarif asli adalah
angka yang berani kamu sebutkan ke customer, sedangkan yang memakai tarif
cadangan cuma perkiraan kasar.

## 14.4 Pengguna

Mengelola akun tim.

{{img:82-pengaturan-pengguna}}

**Menambah pengguna:**

1. Tekan **Tambah Pengguna**.
2. Isi nama, email, dan password awal.
3. Pilih **Role**:

| Role | Bisa apa |
|---|---|
| **Owner** | Semuanya, termasuk laporan laba dan pengaturan |
| **Admin** | Seluruh operasional harian, tanpa laporan margin dan pengaturan |
| **Tripper** | Hanya daftar belanja dan input pembelian |

4. Tekan **Simpan**.

**Tindakan lain:**

- **Ubah** — mengubah nama, role, dan status aktif. Email tidak bisa diubah.
- **Ikon kunci** — mereset password. Pengguna itu akan dikeluarkan dari semua
  perangkat.
- **Ikon tempat sampah** — menghapus akun.

Dua hal yang tidak diizinkan: menghapus akunmu sendiri, dan menghapus owner
terakhir yang masih aktif.

> **Beri tripper akun sendiri.** Dia cuma butuh dua layar, dan bisa mencatat
> belanja langsung dari ponsel sambil berdiri di toko. Jauh lebih akurat
> daripada mengirim foto struk ke admin untuk diketik ulang nanti.

## 14.5 Jejak Perubahan

Catatan siapa mengubah apa, dan kapan.

{{img:83-pengaturan-audit}}

Buka halaman ini kalau ada angka yang terasa janggal dan kamu perlu tahu
riwayatnya. Saring per jenis data lewat dropdown di kanan atas.

Yang tercatat antara lain: order dibuat, item diubah, status berpindah,
pembayaran dicatat, pembelian dihapus, dan resi diinput.

---

# 15. Alur kerja satu trip, dari awal sampai selesai

Ringkasan urutan kerja lengkap, sebagai daftar periksa.

## Sebelum berangkat

| # | Kerjakan | Menu |
|---|---|---|
| 1 | Buat trip, isi tanggal dan **kurs** | Trip → Buat Trip |
| 2 | Susun katalog produk beserta markup | Trip → tab Katalog |
| 3 | Ubah status menjadi **Buka Order** | Trip → tombol status |
| 4 | Posting katalog ke media sosial | *(di luar aplikasi)* |
| 5 | Catat order yang masuk, sekalian pilih **Asal order** | Order → Catat Order |
| 6 | Tagih dan catat DP tiap customer | Detail order → Tagih DP |

## Selama di luar negeri

| # | Kerjakan | Menu |
|---|---|---|
| 7 | Ubah status trip menjadi **Sedang Belanja** | Trip → tombol status |
| 8 | Buka daftar belanja di ponsel | Daftar Belanja |
| 9 | Catat tiap pembelian begitu keluar dari kasir | Daftar Belanja → Catat Beli |
| 10 | Catat biaya bagasi, transport, akomodasi | Trip → tab Biaya |

## Setelah pulang

| # | Kerjakan | Menu |
|---|---|---|
| 11 | Ubah status trip menjadi **Tiba di Indonesia** | Trip → tombol status |
| 12 | Cocokkan barang dengan tiap pesanan | Detail order → Cocokkan Barang |
| 13 | Kemas per customer: timbang, ukur kardus, hitung ongkir | Detail order → Tandai Dikemas |
| 14 | Terbitkan dan kirim invoice pelunasan | Detail order → Invoice |
| 15 | Catat pelunasan yang masuk | Detail order → Catat Bayar |
| 16 | Input resi JNE dan kabari customer | Detail order → Input Resi |
| 17 | Periksa laba trip | Trip → tab Profit |
| 18 | Jual sisa stok di marketplace | Stok → Jual |
| 19 | Lihat channel mana yang paling menghasilkan | Laporan → Per Channel |
| 20 | Catat siapa pembelanja terbesar untuk trip berikutnya | Laporan → Per Customer |

Menu **Siap Kemas** membantu langkah 12 sampai 16: bukalah menu itu tiap pagi
dan kerjakan antreannya dari atas.

---

# 16. Pertanyaan yang sering muncul

**Customer minta tambah barang, ordernya sudah saya buat. Bagaimana?**
Buka detail ordernya, tekan **Tambah Produk** di panel Item pesanan. Kalau order
sudah lunas, statusnya otomatis turun ke Ditagihkan supaya tagihan barunya
terlihat.

**Tokonya kehabisan barang, cuma dapat sebagian.**
Catat pembelian sejumlah yang benar-benar didapat. Saat mencocokkan barang, isi
kolom Diterima sesuai kenyataan; statusnya otomatis menjadi **Sebagian**. Barang
yang tidak dapat tetap muncul di invoice dengan keterangannya.

**Saya beli lebih banyak dari pesanan. Apakah rugi?**
Tidak. Kelebihannya otomatis masuk stok dan tidak dipotong dari laba trip. Jual
lewat menu Stok kapan saja.

**Kenapa saya tidak bisa input resi?**
Kemungkinan ordernya belum lunas. Catat pelunasannya dulu. Kalau memang mau
dikirim duluan, centang **Kirim walau belum lunas**.

**Kenapa harga di katalog tidak berubah padahal kurs sudah saya ubah?**
Disengaja, supaya harga yang sudah diumumkan ke customer tidak bergeser
diam-diam. Tekan **Hitung Ulang Harga** di tab Katalog kalau memang ingin
mengubahnya.

**Kenapa ongkir yang saya bayar ke JNE jauh lebih mahal dari timbangan?**
Karena kardusnya besar. Ekspedisi menagih mana yang lebih besar antara berat
asli dan berat volume, yaitu (panjang × lebar × tinggi) ÷ 6000. Kardus 40 × 30 ×
25 cm dihitung 5 kg walaupun isinya cuma 800 gram. Isi kolom **Dimensi paket**
saat mengemas lalu tekan **Hitung Ongkir**, supaya ketahuan sebelum kamu
menagihkan ongkir ke customer.

**Kota tujuan customer saya tidak ada tarifnya.**
Aplikasi akan memakai tarif cadangan dan memberi peringatan kuning. Tambahkan
kota itu di **Pengaturan → Ongkir** sesuai daftar harga ekspedisimu, supaya
perkiraan berikutnya tepat.

**Saya lupa mengisi asal order. Bisa diperbaiki?**
Bisa. Buka detail ordernya, tekan **Ubah Order**, lalu ubah kolom **Asal order**.
Laporan Per Channel langsung menyesuaikan.

**Kenapa laporan Per Channel isinya WhatsApp semua?**
Order yang dicatat sebelum kolom asal order ada dihitung sebagai WhatsApp.
Angkanya akan mulai mencerminkan keadaan sebenarnya setelah beberapa trip
berikutnya dicatat dengan channel yang benar.

**Saya salah catat pembayaran.**
Tekan ikon tempat sampah di samping pembayaran tersebut. Sisa tagihan langsung
dihitung ulang.

**Order sudah dikirim tapi jumlahnya salah.**
Order yang sudah dikirim tidak bisa diubah. Kalau perlu koreksi uang, catat
pembayaran jenis **Refund** atau **Penyesuaian**.

**Menu Laporan dan Pengaturan tidak ada di layar saya.**
Menu yang tidak boleh kamu akses memang disembunyikan. Laporan margin dan
Pengaturan hanya untuk owner.

**Bagaimana cara mengganti nomor rekening di invoice?**
Pengaturan → tab Toko → kolom **Rekening pembayaran** → Simpan.
