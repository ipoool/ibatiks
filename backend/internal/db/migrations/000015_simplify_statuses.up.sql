-- Menyederhanakan status trip dan order.
--
-- Status yang sebelumnya mengikuti posisi barang (trip sedang belanja, dalam
-- perjalanan, sudah sampai) ternyata menuntut admin memperbarui dua tempat
-- untuk satu kejadian yang sama, dan jarang dipakai. Yang benar-benar dipakai
-- hanya: trip masih menerima order atau tidak.
--
-- Trip: draft/closed/in_transit/arrived/settled/cancelled -> closed,
--       open/shopping -> open.
ALTER TABLE trips DROP CONSTRAINT IF EXISTS trips_status_check;

UPDATE trips SET status = 'open'   WHERE status IN ('open', 'shopping');
UPDATE trips SET status = 'closed' WHERE status NOT IN ('open');

ALTER TABLE trips ADD CONSTRAINT trips_status_check
    CHECK (status IN ('open', 'closed'));

-- Trip dibuat untuk menerima order, jadi bawaannya langsung terbuka. Bawaan
-- lamanya 'draft' — status yang sudah tidak ada, sehingga trip baru akan
-- ditolak batasan di atas.
ALTER TABLE trips ALTER COLUMN status SET DEFAULT 'open';

-- Order: "purchasing" dan "arrived" dilebur ke "dp_paid" (Diproses).
--
-- Keduanya sama-sama berarti pesanan sedang dikerjakan dan belum dikemas.
-- Melebur ke Diproses lebih jujur daripada menaikkannya ke Sedang Dikemas,
-- karena barangnya memang belum masuk kardus.
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;

UPDATE orders SET status = 'dp_paid' WHERE status IN ('purchasing', 'arrived');

ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('draft', 'awaiting_dp', 'dp_paid', 'packed', 'invoiced',
                      'paid', 'shipped', 'completed', 'cancelled'));
