-- Alamat Indonesia butuh kecamatan dan kelurahan.
--
-- Tanpa keduanya, alamat kirim yang dicetak untuk kurir sering kurang lengkap
-- dan paket tertahan di gudang sortir — apalagi untuk nama jalan yang sama di
-- dua kelurahan berbeda. Kolomnya dibuat opsional supaya data lama tetap sah.
ALTER TABLE customers
    ADD COLUMN district    text,
    ADD COLUMN subdistrict text;

COMMENT ON COLUMN customers.district IS 'Kecamatan';
COMMENT ON COLUMN customers.subdistrict IS 'Kelurahan atau desa';

-- Order menyimpan salinan alamat kirim supaya dokumen lama tidak berubah saat
-- data customer disunting, jadi kolomnya ikut ditambahkan di sini.
ALTER TABLE orders
    ADD COLUMN shipping_district    text,
    ADD COLUMN shipping_subdistrict text;

COMMENT ON COLUMN orders.shipping_district IS 'Kecamatan tujuan kirim';
COMMENT ON COLUMN orders.shipping_subdistrict IS 'Kelurahan tujuan kirim';
