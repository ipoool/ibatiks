-- Premi yang sudah tercatat ikut hilang; nilainya sudah menyatu di
-- orders.shipping_fee, jadi yang ditagihkan ke customer tidak berubah.
ALTER TABLE shipments DROP COLUMN insurance_fee;
