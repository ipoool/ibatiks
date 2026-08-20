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

**Status sengaja sedikit.** Trip hanya `open`/`closed`; order hanya enam: `awaiting_dp`, `dp_paid`, `paid`, `shipped`, `completed`, `cancelled`. Sebuah status hanya boleh ada kalau ia menjawab pertanyaan yang datanya tidak tercatat di tempat lain. Belanja, penerimaan barang, pengemasan, penetapan ongkir, dan penerbitan invoice pelunasan semuanya terjadi *di dalam* `dp_paid` — masing-masing sudah meninggalkan jejaknya sendiri (data pembelian, data penerimaan, data kemasan, baris invoice), jadi status tersendiri hanya akan jadi salinan yang bisa berbeda dari data aslinya. Jangan menambah status baru untuk menandai kejadian yang datanya sudah tercatat di tempat lain.

**Ongkir ditetapkan saat mengemas, bukan saat order dicatat.** Waktu order masuk, barangnya belum ada dan beratnya belum diketahui; angka yang diisi di situ cuma tebakan. `ShipmentService.Pack` yang menuliskannya ke `orders.shipping_fee` lalu memanggil `RecalculateTotals`. Dua hal yang wajib menyertainya:

- **`dp_required` tidak ikut dihitung ulang.** DP sudah disepakati dan besar kemungkinan sudah dibayar; menaikkannya belakangan berarti customer yang sudah menyetor tiba-tiba dianggap kurang bayar.
- **Status ikut direkonsiliasi** lewat `reconcileOrderStatus`. Customer yang terlanjur melunasi sebelum paketnya ditimbang kembali punya sisa tagihan sebesar ongkirnya; tanpa ini ordernya tetap berlabel Pembayaran Lunas, ikut masuk antrean siap kirim, dan barangnya berangkat sementara ongkirnya tidak pernah tertagih.

Kolom `shipping_fee` pada permintaan edit order bertipe penunjuk. Kalau bertipe nilai biasa, form yang tidak mengirimkannya akan mengirim nol dan menghapus diam-diam ongkir yang baru dihitung kurir.

**Invoice pelunasan menagih seluruh sisa pesanan termasuk ongkir**, jadi ia ditolak selama order masih `awaiting_dp` atau ongkirnya masih nol. Menerbitkannya lebih awal berarti mengirim tagihan yang nilainya masih akan berubah — customer membayar, lalu ditagih lagi, dan dokumen yang sudah ia pegang tidak cocok dengan yang tercatat. Invoice DP diterbitkan dari detail order; invoice pelunasan dari menu Invoice, yang hanya menawarkan order yang benar-benar sudah siap.

**Menandai invoice lunas berarti mencatat pembayarannya**, bukan mengubah label barisnya. Saldo order, status order, dan laporan piutang semuanya dihitung dari tabel `payments`.

**Invoice memuat nilai order seutuhnya.** Baik invoice DP maupun pelunasan menuliskan subtotal, diskon, ongkir, dan total order yang sebenarnya; yang membedakan hanya apa yang ditagih (`dp_amount` untuk DP, sisa tagihan untuk pelunasan). Jangan kembali menyimpan nilai DP sebagai `total` — customer akan menerima dokumen yang seolah-olah menyatakan pesanannya cuma seharga uang muka.

**Barang yang tidak dapat tidak ikut ditagih, dan dasarnya `qty_purchased`.** Untuk item yang tidak terbeli penuh, yang ditagihkan adalah `qty_purchased × unit_price` — bukan jumlah yang dulu dipesan. Aturannya di `RecalculateTotals`, dan `PurchaseService` wajib memanggilnya setiap kali `SyncItemPurchasedQty` dipanggil, baik saat pembelian dicatat maupun dihapus. Kolom `qty` sengaja dibiarkan apa adanya supaya tetap terbaca apa yang dipesan semula; yang berubah hanya apa yang ditagih.

Dulu dasarnya `qty_received`, diisi lewat layar "Cocokkan Barang" tersendiri. Layar itu dilepas: angkanya sudah ada di data pembelian, dan meminta orang mengetiknya ulang berarti satu angka yang sama disimpan dua kali dan bisa berbeda. Sisanya justru berbahaya — selama belum dicocokkan, `qty_received` masih nol, jadi item berstatus `partial` tertagih **nol rupiah** kalau ada jalur lain yang memanggil `RecalculateTotals` lebih dulu (menetapkan ongkir saat mengemas, misalnya). Kolom `qty_received` ditinggalkan di database untuk baris lama dan tidak lagi ditulis siapa pun.

**Akun media sosial customer disimpan sebagai JSONB di kolom `customers.socials`,** bukan tabel tersendiri. Isinya selalu dibaca sebagai satu kesatuan bersama customernya, tidak pernah diagregasi maupun jadi syarat penyaringan — tabel terpisah berarti join pada tiap pembacaan customer untuk data yang tidak pernah butuh dijoin. Daftar platformnya tertutup (`domain.SocialPlatforms`) dan divalidasi di service, bukan lewat CHECK constraint: pesan galat dari constraint tidak bisa menyebut baris keberapa yang salah. Handle disimpan apa adanya termasuk tanda "@" — sebagian platform memakainya, sebagian tidak, dan menormalkannya berarti menebak.

**Laporan laba-rugi melayani satu trip dan seluruh trip lewat satu kueri.** Penyaringnya `($1::uuid IS NULL OR trip_id = $1)` di `TripFinancials` dan `ExpenseBreakdown`; `trip_id` nil berarti semuanya dijumlahkan. Menggandakan kuerinya untuk kasus "semua trip" berarti dua definisi laba yang harus dijaga tetap sama, dan cepat atau lambat salah satunya tertinggal saat rumusnya berubah. Identitas trip pada hasilnya kosong saat lintas trip — tidak ada satu trip yang bisa disebut.

**Snapshot historis.** `order_items` menyimpan salinan nama dan harga produk; `orders` menyimpan salinan alamat kirim. Mengedit master data tidak boleh mengubah dokumen lama.

**Menghapus trip membuang seluruh riwayat di dalamnya**, dan dua hal menghalanginya — keduanya soal kenyataan di luar aplikasi, bukan kerapian data:

- **Order yang sudah diserahkan ke kurir.** Barangnya sudah di jalan atau sudah diterima; penjualannya sudah terjadi, dan menghapus catatannya tidak membatalkan apa pun selain ingatan toko sendiri.
- **Barang surplus yang masih ada di stok.** Barangnya nyata dan masih bisa dijual; membuang catatan asalnya hanya menyisakan stok tanpa asal-usul dan tanpa harga modal yang bisa dipertanggungjawabkan. Stok bersifat fungible, jadi yang dihitung adalah yang lebih kecil antara surplus dari trip itu dan sisa stok sekarang.

Yang tidak menghalangi tapi tetap hilang: **uang yang sudah diterima**. Karena itu nominalnya dicatat ke `audit_logs` sebelum barisnya dihapus — setelah itu jejak tersebut satu-satunya yang tersisa, dan dialog konfirmasinya wajib menyebutkan angkanya.

`orders.trip_id` sengaja tetap `ON DELETE RESTRICT`. Order dihapus lewat `OrderRepo.DeleteByTrip` di dalam transaksi yang sama, supaya tidak ada jalur lain di aplikasi ini yang bisa membuang order tanpa melewati penjagaan di atas.

**Label pengiriman, bukan surat jalan.** Yang dicetak dan ditempel di kardus adalah label pengirim–penerima ukuran 100 × 150 mm (`internal/pdf/label.go`), mengikuti kertas thermal yang dipakai kurir. Tanpa daftar barang dan tanpa nominal: label terbaca siapa pun yang memegang paket di jalan, dan isi belanjaan customer bukan urusan mereka. Blok penerima sengaja jauh lebih besar dari yang lain — itu satu-satunya bagian yang benar-benar dibaca orang.

Nama kurir dan nomor resi juga tidak dicetak. Keduanya sudah ada di label resmi kurir yang ditempel berdampingan; menuliskannya lagi berarti dua nomor pada satu kardus, dan yang dari kami sudah kedaluwarsa begitu paket dibatalkan lalu dikirim ulang lewat kurir lain.

**Dokumen historis tetap bisa dibuka walau master datanya dihapus.** Detail order memakai `GetByIDIncludingDeleted` untuk customernya: menghapus customer tidak menghapus ordernya, jadi menyaring `deleted_at IS NULL` di jalur itu membuat seluruh order lamanya mati dengan "customer tidak ditemukan" — padahal ordernya masih ada dan bisa jadi masih punya sisa tagihan. Daftar dan penyuntingan tetap menyaring yang sudah dihapus.

**Daftar yang ditautkan dari halaman lain membaca kata kuncinya dari alamat** lewat `usePencarianURL` (`src/hooks/use-pencarian-url.ts`). Detail order dan laporan Per Customer menautkan ke `/customers?q=…`; tanpa itu tautannya mendarat di daftar penuh dengan kotak pencarian kosong, dan yang mengeklik mengira customernya tidak ketemu. Nilainya dibaca sekali sebagai keadaan awal — menyetelnya lewat efek dilarang aturan lint dan akan menimpa apa yang sedang diketik — sedangkan alamatnya diperbarui dengan `history.replaceState`, bukan `router.replace`, supaya tiap ketikan tidak memicu navigasi ulang.

**Pencarian customer menormalkan nomor.** Nomor disimpan sebagai `62812…`, sementara admin mengetiknya seperti yang tertera di WhatsApp customer (`0812…`). Kata kunci yang berbentuk nomor ikut dinormalkan sebelum dicocokkan; `domain.LooksLikePhone` sengaja ketat supaya kata kunci berhuruf tetap diperlakukan sebagai pencarian nama dan hasilnya tidak melebar.

**Data PostgreSQL tinggal di folder `./data/postgres`, bukan volume Docker.** Volume bernama ikut terhapus oleh `docker compose down -v` — perintah pembersihan container yang tidak menyebut-nyebut database sama sekali. Jangan mengembalikannya jadi volume. `PGDATA` sengaja menunjuk subdirektori `pgdata` di dalam titik mount, sebab postgres menolak jalan kalau direktori datanya bukan miliknya sendiri. Satu-satunya perintah yang menghapus folder itu adalah `make reset`, dan target itu menghapusnya secara terpisah karena `down -v` tidak lagi bisa.

**Migrasi bersifat tambahan.** Jangan menyunting file migrasi yang sudah pernah dijalankan — buat migrasi baru. Setiap `.up.sql` wajib punya `.down.sql` yang benar-benar membalik; uji dengan `make migrate-down && make migrate-up`.

## Ongkir

**Hanya satu sumber angka: kurir, lewat RajaOngkir.** Toko pernah menyimpan tabel tarif per kota sendiri sebagai cadangan; itu dilepas di migrasi 000019. Tarif yang diisi tangan tidak pernah ikut naik saat kurir menaikkan harganya, dan angka yang salah tapi terlihat resmi lebih berbahaya daripada tidak ada angka sama sekali. Jangan menambahkan sumber tarif kedua.

- **Kegagalan menghubungi kurir kini dilaporkan, bukan ditelan.** Selama masih ada tabel tarif, menelan galat berarti tetap memberi angka. Sekarang tidak ada angka pengganti, jadi menelannya hanya menghasilkan daftar kosong tanpa penjelasan. `Estimate` dan `Options` mengembalikan galat yang menyebutkan sebabnya.
- **Asuransi kiriman preminya diketik, bukan dihitung.** RajaOngkir tidak mengembalikan data asuransi sama sekali — balasan `calculate/domestic-cost` hanya berisi `name`, `code`, `service`, `description`, `cost`, dan `etd`. Menanam rumus premi di kode berarti angka yang tidak pernah ikut berubah saat kurir mengubah tarifnya, persis alasan tabel tarif ongkir dilepas. Preminya disimpan di `shipments.insurance_fee` sebagai rincian, sementara yang ditagihkan ke customer tetap satu angka di `orders.shipping_fee` (ongkir + premi) — invoice dan label sudah membacanya dari sana, dan memecahnya jadi dua baris tagihan berarti mengubah bentuk dokumen yang sudah dipegang customer.
- **Jaring pengamannya manusia, bukan tabel.** Dialog kemas punya centang "Manual Ongkir", dipakai kalau daftar layanan gagal keluar. Angka dari struk konter lebih benar daripada tebakan mana pun. Jangan menggantinya dengan tarif default. Kedua sumber angka tidak pernah tampil bersamaan: menyalakan centang itu menyembunyikan daftar layanan, dan mematikannya mengosongkan angka yang sempat diketik — dua kolom ongkir yang aktif berbarengan membuat admin harus menebak mana yang tersimpan.
- **Kurir menjual ongkos utuh, bukan harga per kilo.** Jangan mengubah balasan RajaOngkir jadi tarif per kg lalu dikalikan — tarifnya berjenjang dan hasilnya meleset pada berat yang bukan kelipatan bulat. `price_per_kg` pada hasil cuma turunan untuk ditampilkan.
- **Pembagi berat volumetrik adalah konstanta** `domain.VolumetricDivisor` (6000, mengikuti JNE). Dulu bisa disetel per toko dan tidak ada yang pernah mengubahnya, sementara salah menyetelnya membuat seluruh perkiraan meleset tanpa gejala.
- **`shipping_couriers` dipisah titik dua** (`jne:jnt:sicepat`) karena itu bentuk yang diminta API-nya. Simpan apa adanya, jangan diterjemahkan bolak-balik.
- **Pemetaan alamat ke ID tujuan disimpan** di `shipping_destinations`. Kuota langganan terbatas dan pemetaan kota ke ID hampir tidak pernah berubah; mencari ulang alamat yang sama membuang kuota. Urutan percobaannya dari yang paling spesifik: kode pos → kelurahan + kota → kecamatan + kota → kota.
- **API key tidak pernah menyentuh browser.** `RAJAONGKIR_API_KEY` tinggal di server; menu Pengaturan hanya menyimpan ID kota asal, labelnya, dan daftar kurir.
- **Pesan penolakan dari kurir diteruskan apa adanya.** "API key tidak valid" hanya bisa dibereskan tim toko, dan "terjadi kesalahan pada server" tidak memberi tahu apa pun.

## Frontend

**shadcn/ui gaya `new-york` (berbasis Radix).** Jangan memasang komponen gaya `base-nova` — mencampur dua pustaka primitif akan mengacaukan token dan perilaku fokus. Komponen dasar ada di `src/components/ui/`.

**Jangan menyetel state di dalam efek** — aturan lint `react-hooks/set-state-in-effect` menolaknya. State turunan dihitung langsung saat render (lihat `nav-menu.tsx`: pilihan pengguna disimpan bersama `pathname` supaya kedaluwarsa sendiri saat pindah halaman).

**Select tidak boleh punya item bernilai string kosong** (batasan Radix). Pakai `FilterSelect` yang sudah menangani sentinel `__all__`, atau `OptionSelect` bila memang tanpa opsi "semua".

**DP bisa diketik sebagai rupiah atau persen** lewat `InputDP` (`src/components/input-dp.tsx`), tapi yang dikirim ke API selalu rupiah. Satuannya disimpan di state pemanggil, bukan di dalam komponen: dalam mode persen nominalnya dihitung ulang tiap render, sehingga menambah item membuat DP ikut menyesuaikan — dan menyimpan state itu di dalam komponen akan menuntut efek yang menyetel state, yang dilarang aturan lint di bawah. Dasar hitungannya nilai barang (subtotal − diskon), bukan total, sebab total sudah memuat ongkir begitu paket ditimbang sementara DP memang tidak ikut dihitung ulang saat itu.

**Input uang memakai `step="any"`.** `step="1000"` membuat nominal seperti Rp348.400 ditolak browser — termasuk DP 50% yang dihitung sistem sendiri dan transfer berkode unik.

**Tabel harus menyusut, bukan menggeser halaman.** Kartu dan kolom grid yang memuat tabel wajib `min-w-0` — anak grid/flex bawaannya tidak boleh lebih sempit dari isinya, sehingga tabel lebar mendorong seluruh halaman ke samping dan `overflow-x-auto` di dalamnya tidak pernah terpakai. Kolom sekunder disembunyikan bertahap (`hidden sm:table-cell` lalu `hidden lg:table-cell`/`xl:table-cell`), dan isian yang ikut hilang dilipat ke kolom utama sebagai baris kecil `sm:hidden`. Sel bawaan shadcn memakai `whitespace-nowrap`; kolom berisi nama panjang perlu `whitespace-normal` supaya boleh turun baris.

**Radix Select memantul balik kalau nilainya disetel ke opsi yang baru muncul di render yang sama.** Ia memanggil `onValueChange("")` dan pilihannya langsung terlepas — gejalanya: produk yang baru dibuat lewat quick-add tampak tersimpan tapi kolomnya kembali ke placeholder. Untuk kolom yang isinya bisa bertambah dari dalam form itu sendiri, pakai `Combobox`: labelnya dibaca dari daftar biasa, jadi tidak punya perilaku itu.

**Tombol beriskon memakai prop `tooltip` milik `Button`**, bukan merangkai `Tooltip`/`TooltipTrigger`/`TooltipContent` sendiri. Ikon tanpa teks baru bisa dibedakan setelah diklik, dan yang dirangkai manual cepat atau lambat ada yang terlewat. Sertakan juga `<span className="sr-only">` — tooltip tidak muncul untuk navigasi papan ketik di semua pembaca layar.

**Dropdown yang isinya bisa panjang** (customer, katalog produk) memakai `Combobox`, dengan `keywords` untuk hal yang dicari orang tapi bukan judul: nomor WA, SKU.

**Mata uang dipilih dari daftar, bukan diketik.** Daftarnya di `src/lib/mata-uang.ts` dan dipakai bersama form trip dan form produk lewat `CurrencySelect`, supaya keduanya tidak pernah menawarkan pilihan yang berbeda. Urutannya ASEAN dulu, baru tujuan lain — yang paling sering dipakai paling sedikit digulir. Salah ketik satu huruf ("IRD" alih-alih "IDR") dulu membuat kurs gagal diambil dan seluruh harga jual trip meleset, sementara tulisannya tetap terlihat wajar sekilas. Kode di luar daftar tetap ditampilkan sebagai pilihan tambahan supaya membuka trip lama tidak diam-diam mengganti mata uangnya.

**`src/lib/utils.ts` adalah pusat helper format** (`cn`, `formatIDR`, `formatDate`, `toNumber`, …) dan dipakai puluhan berkas. `shadcn init` pernah menimpanya — jangan jalankan ulang init; tambahkan komponen satu per satu dengan `shadcn add`.

**Middleware harus berada di `src/middleware.ts`.** Proyek ini memakai direktori `src/`; berkas `middleware.ts` di akar `web/` diabaikan Next tanpa peringatan apa pun. Pastikan `ƒ Proxy (Middleware)` muncul di keluaran `next build`.

**Konstanta yang dibaca komponen server tidak boleh diekspor dari modul `"use client"`.** Nilai yang diimpor dari modul klien berubah jadi rujukan modul, dan pemakaiannya gagal diam-diam (`cookies().get(konstanta)` mengembalikan undefined). Simpan di modul biasa seperti `src/lib/sidebar.ts` dan `src/lib/route-permissions.ts`.

**Penjagaan kiriman ganda memakai `useRef`, bukan `isPending`.** Baik atribut `disabled` maupun `mutation.isPending` baru berubah setelah React merender ulang, jadi dua kiriman di tick yang sama — klik ganda, tombol Enter yang ditahan — sama-sama membaca nilai lama dan lolos berdua. Di form catat order itu berarti dua order kembar untuk customer yang sama. Ref berubah saat itu juga; jangan lupa melepasnya kembali di `onError` supaya admin tetap bisa mencoba ulang.

**Dialog tidak boleh menjalankan submit form halaman di baliknya.** Radix memindahkan isi dialog ke ujung body, tapi React merambatkan event lewat pohon komponen, bukan pohon DOM. `FormDialog` sudah menghentikan perambatan submit-nya; jangan menghapus baris itu. Tanpa itu, menekan Simpan di Tambah Customer pada halaman catat order ikut membuat ordernya.

**Dialog kemas menahan tombol Simpan sampai seluruh barang dicentang.** Centangnya tidak disimpan ke database — itu ritual di meja kemas, bukan data. Kalau daftar barangnya gagal dimuat, penguncian dilepas: menahan Simpan karena jaringan bermasalah berarti paket yang sudah siap tidak bisa dicatat sama sekali.

**Tombol simpan pada dialog berisi `Combobox` wajib memakai `submitDisabled`.** Radix Combobox bukan `<select>` bawaan, jadi validasi bawaan browser tidak melihatnya sama sekali: tombolnya tetap bisa ditekan selagi kosong, dan gelembung yang muncul justru menunjuk kolom lain yang kebetulan berupa input biasa.

**Pesan validasi bawaan browser disetel sendiri.** Bahasanya mengikuti bahasa antarmuka browser, bukan `lang` halaman, jadi `<html lang="id">` tidak menolong. `Input` dan `Textarea` memasang pesan Indonesia lewat `src/lib/validasi-bawaan.ts` dan membuangnya lagi begitu isinya berubah — pembuangan itu wajib, sebab `customValidity` yang tertinggal membuat kolomnya dianggap tidak sah selamanya dan formulirnya tidak akan pernah mau dikirim.

**Bar tab menggulir mendatar di layar sempit, membungkus mulai `sm`.** Membungkus di ponsel membuat lima tab jadi dua baris yang mendorong isi halaman turun, dan barisan yang tingginya berubah-ubah antar halaman terbaca seperti tata letak yang goyah. `scrollbar-hidden` wajib menyertainya: `overflow-x-auto` memaksa `overflow-y` ikut jadi `auto`, dan isi yang satu piksel lebih tinggi memunculkan scrollbar tegak kecil di ujung kanan. Pemicunya juga wajib `shrink-0` — tanpa itu tab menyusut agar muat, barisannya tidak pernah meluap, dan tidak ada yang bisa digulir.

**Tinggi isi dialog dibatasi 670px** (`FormDialog`), lebih dari itu digulir. Wadahnya sengaja tanpa `flex-1`: tinggi yang dibagi flexbox sering jatuh di setengah piksel, `scrollHeight` membulat ke atas sementara `clientHeight` ke bawah, dan scrollbar muncul untuk satu piksel yang tidak ada isinya.

**Jangan membungkus isian form dengan grid dua kolom di dalam grid dua kolom.** Pembungkusnya menempati satu sel, jadi isian di dalamnya tinggal seperempat lebar form — dan pembungkus itu sendiri ikut menggeser urutan sel berikutnya. Taruh tiap `Field` langsung sebagai anak grid formnya; kalau memang harus melebar, pakai `sm:col-span-2`.

**Anak grid yang memakai `col-span-2` wajib memakai awalan `sm:`.** Tanpa awalan itu, span-nya membuat kolom kedua secara implisit bahkan saat induknya satu kolom — dan tidak ada `grid-cols-1` yang bisa menolongnya, karena kolomnya dibuat oleh span itu sendiri. Gejalanya: dua isian yang seharusnya menumpuk di ponsel malah berdesakan bersebelahan.

## Hak akses

Role (`owner`/`admin`/`tripper`) menentukan batas kasar; di dalamnya owner bisa mencentang menu per pengguna lewat **Pengaturan → Pengguna**. Aturannya hanya ditulis sekali, di `internal/domain/permission.go`:

- Daftar kosong di kolom `users.permissions` berarti "ikut bawaan role" — bukan "tanpa akses".
- Centang hanya bisa **mempersempit**; backend menyaring ulang permintaan supaya tripper tidak bisa diberi menu pengaturan.
- Hak akses ikut dibawa di dalam access token, jadi mengubahnya mencabut sesi pengguna itu supaya pembatasannya berlaku saat itu juga.
- Frontend memakai `effective_permissions` yang dihitung backend; jangan menyalin tabel bawaan role ke UI selain untuk menampilkan pilihan centang.
- **Owner tidak bisa mencabut Pengaturan dan Pengguna dari dirinya sendiri** (`OwnerLockedPermissions`). Dua menu itu satu-satunya jalan mengembalikan hak akses siapa pun; sekali hilang, satu-satunya pemulihan adalah `UPDATE users SET permissions = NULL` langsung ke database. Karena dihitung dan bukan disimpan, baris yang terlanjur rusak ikut pulih sendiri.

**Dashboard ikut butuh hak Laporan, dan orang yang tidak punya tidak mendarat di sana.** Seluruh isi Dashboard datang dari satu endpoint laporan; tanpa hak itu halamannya tetap terbuka tapi datanya ditolak, dan yang terbaca adalah deretan angka nol lengkap dengan "Belum ada order" — seolah tokonya kosong. Kena persis pada tripper. Sesudah login dan pada tiap penolakan rute, middleware mengarahkan ke `firstAllowedPath`: halaman pertama yang benar-benar boleh dibuka, mengikuti urutan sidebar. Pengalihannya berhenti sendiri kalau tujuannya halaman itu juga — tanpa itu pengguna tanpa hak apa pun terjebak putaran.

Menu yang tidak dimiliki pengguna juga **dijaga di tingkat rute**, bukan cuma disembunyikan dari sidebar. Petanya di `src/lib/route-permissions.ts` — modul biasa supaya bisa dibaca middleware tanpa menyeret ikon menu ke bundel edge — dan dipakai bersama oleh sidebar dan `src/middleware.ts`, jadi menu yang disembunyikan dan halaman yang ditolak tidak pernah berbeda pendapat. Ini bukan lapisan keamanan; backend tetap menolak endpoint-nya sendiri. Yang dijaga adalah supaya orang tidak mendarat di halaman yang datanya gagal dimuat lalu membaca "Belum ada customer" seolah tokonya kosong.

## Sesi dan autentikasi

Token tidak pernah menyentuh `localStorage`. Alurnya:

- `src/app/api/auth/*` menaruh access + refresh token di cookie httpOnly.
- `src/app/api/proxy/[...path]` menyisipkan `Authorization` untuk panggilan dari browser dan memperbarui token saat 401.
- `src/middleware.ts` memperbarui token **sebelum halaman dirender di server** — render halaman tidak lewat proxy, jadi tanpa ini sesi putus tiap 15 menit.
- Layout dashboard mengarahkan ke `/login?expired=1` bila backend menolak sesi. Penanda itu wajib: tanpanya middleware melihat cookie yang tampak sah dan melempar balik, dan halaman terjebak putaran pengalihan.

**Login dikunci setelah lima kegagalan.** Lima kali gagal dalam lima menit mengunci sebuah email selama lima menit; aturannya di `domain.LoginMaxAttempts` dan `domain.LoginBlockDuration`, hitungannya di tabel `login_attempts`.

- **Dihitung per email, bukan per IP.** Keputusan sadar: penebak password yang berpindah-pindah IP tetap tertahan. Konsekuensinya juga sadar — siapa pun yang tahu email seorang pengguna bisa membuatnya terkunci dengan sengaja salah lima kali. Kuncinya selalu lepas sendiri, tidak pernah permanen.
- **Email yang tidak terdaftar ikut dihitung.** Kalau hanya email terdaftar yang dihitung, pola penguncian membocorkan email mana yang ada di sistem — persis yang dihindari pesan "email atau password salah" yang seragam itu.
- **Hitungannya berjendela.** Kegagalan yang lebih lama dari `LoginBlockDuration` dilupakan. Tanpa itu, satu salah ketik hari ini dan empat lagi bulan depan akan mengunci akun tanpa ada yang menyerang apa pun.
- **Penambahan hitungan dikerjakan satu pernyataan SQL**, bukan baca-lalu-tulis di Go. Dua percobaan yang datang bersamaan akan saling menimpa kalau hitungannya dibaca dulu ke memori, dan penyerang yang menembak paralel justru mendapat percobaan gratis.
- **Akun nonaktif tidak dihitung sebagai kegagalan**: passwordnya benar, dan menguncinya hanya menyusahkan orang yang memang berhak tahu kenapa ia tidak bisa masuk.

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
