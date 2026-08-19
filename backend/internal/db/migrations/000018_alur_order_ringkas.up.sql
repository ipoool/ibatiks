-- Menyederhanakan status order menjadi lima tahap perjalanan ditambah Batal.
--
-- Tiga status dilepas karena masing-masing menandai kejadian yang datanya
-- sudah tercatat di tempat lain:
--
--   draft    -> awaiting_dp. Order dicatat setelah customer setuju, jadi
--               sejak lahir ia memang sedang menunggu DP. Tahap "masih
--               disusun admin" tidak pernah benar-benar dipakai.
--   packed   -> dp_paid. Sudah dikemas atau belum terbaca dari data kemasan
--               (berat, dimensi, waktu dikemas), bukan dari status order.
--   invoiced -> dp_paid. Invoice pelunasan yang terkirim tapi belum dibayar
--               bukan kemajuan. Yang memindahkan order ke Pembayaran Lunas
--               adalah uang yang benar-benar masuk.
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;

UPDATE orders SET status = 'awaiting_dp' WHERE status = 'draft';
UPDATE orders SET status = 'dp_paid'     WHERE status IN ('packed', 'invoiced');

ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('awaiting_dp', 'dp_paid', 'paid', 'shipped',
                      'completed', 'cancelled'));

-- Bawaan lamanya 'draft' — status yang sudah tidak ada, sehingga order baru
-- akan ditolak batasan di atas.
ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'awaiting_dp';
