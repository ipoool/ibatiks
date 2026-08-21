-- Dashboard mendapat hak aksesnya sendiri.
--
-- Sebelumnya ia menumpang hak Laporan, karena seluruh isinya memang datang dari
-- satu endpoint laporan. Akibatnya tim toko tidak bisa memberi seseorang
-- ringkasan harian tanpa sekaligus membuka rekap piutang, penjualan per
-- customer, dan performa produk — dan tidak bisa mencabut Dashboard dari orang
-- yang memang perlu membuka Laporan.
--
-- Diberikan ke setiap role yang sudah memegang Laporan, supaya tidak ada yang
-- kehilangan halaman yang selama ini terbuka untuknya.
UPDATE roles
SET permissions = permissions || ARRAY['dashboard']
WHERE 'reports' = ANY (permissions)
  AND NOT ('dashboard' = ANY (permissions));

-- Centang per pengguna diperlakukan sama. Daftar yang kosong berarti mengikuti
-- rolenya, jadi tidak perlu disentuh.
UPDATE users
SET permissions = permissions || ARRAY['dashboard']
WHERE permissions IS NOT NULL
  AND 'reports' = ANY (permissions)
  AND NOT ('dashboard' = ANY (permissions));
