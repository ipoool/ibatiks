-- Melepas tabel tarif yang diisi sendiri oleh toko.
--
-- Sejak ongkir diambil dari RajaOngkir, tabel ini cuma jadi sumber kedua yang
-- diam-diam berbeda: tarif yang diketik tangan tidak pernah ikut naik saat kurir
-- menaikkan harganya, dan angka yang salah tapi terlihat resmi lebih berbahaya
-- daripada tidak ada angka sama sekali.
--
-- Kalau kurirnya tidak bisa dihubungi, admin mengetik ongkirnya sendiri saat
-- mengemas — dari struk konter atau aplikasi kurir. Itu angka yang benar-benar
-- dibayar, bukan tebakan yang dirawat entah sejak kapan.
DROP TRIGGER IF EXISTS trg_shipping_rates_updated_at ON shipping_rates;
DROP TABLE IF EXISTS shipping_rates;

-- Pengaturan yang hanya melayani tabel itu ikut dilepas.
--
-- shipping_volumetric_divisor jadi konstanta di kode (domain.VolumetricDivisor,
-- 6000 mengikuti JNE): tidak ada toko yang pernah mengubahnya, sementara salah
-- menyetelnya membuat seluruh perkiraan ongkir meleset tanpa gejala apa pun.
DELETE FROM app_settings
WHERE key IN (
    'shipping_default_price_per_kg',
    'shipping_volumetric_divisor',
    'shipping_provider'
);
