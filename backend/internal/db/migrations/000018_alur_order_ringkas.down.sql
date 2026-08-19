-- Mengembalikan batasan dan bawaan yang lama.
--
-- Status aslinya tidak bisa dipulihkan: tiga nilai sudah dilebur menjadi dua,
-- dan basis data tidak lagi menyimpan mana yang dulunya draft, packed, atau
-- invoiced. Yang dijamin turunnya hanyalah batasan yang kembali menerima nilai
-- lama, sehingga versi sebelumnya bisa berjalan lagi tanpa baris yang ditolak.
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;

ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('draft', 'awaiting_dp', 'dp_paid', 'packed', 'invoiced',
                      'paid', 'shipped', 'completed', 'cancelled'));

ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'draft';
