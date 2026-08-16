-- Mengembalikan daftar status yang lebih panjang.
--
-- Data tidak bisa dipulihkan ke status aslinya: peleburannya searah, dan trip
-- yang tadinya "shopping" sudah tidak bisa dibedakan dari trip "open". Yang
-- dikembalikan hanya batasan kolomnya, supaya status lama bisa dipakai lagi.
ALTER TABLE trips DROP CONSTRAINT IF EXISTS trips_status_check;

ALTER TABLE trips ADD CONSTRAINT trips_status_check
    CHECK (status IN ('draft', 'open', 'closed', 'shopping',
                      'in_transit', 'arrived', 'settled', 'cancelled'));

ALTER TABLE trips ALTER COLUMN status SET DEFAULT 'draft';

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;

ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('draft', 'awaiting_dp', 'dp_paid', 'purchasing',
                      'arrived', 'packed', 'invoiced', 'paid',
                      'shipped', 'completed', 'cancelled'));
