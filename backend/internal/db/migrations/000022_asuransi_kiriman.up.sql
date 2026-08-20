-- Premi asuransi kiriman, dicatat terpisah dari ongkirnya.
--
-- Yang ditagihkan ke customer tetap satu angka di orders.shipping_fee — invoice
-- dan label sudah membacanya dari sana, dan memecahnya jadi dua baris tagihan
-- berarti mengubah bentuk invoice yang sudah dipegang customer. Rinciannya
-- disimpan di sini supaya tetap bisa dijawab berapa yang ongkir dan berapa yang
-- asuransi ketika ada yang bertanya.
--
-- Preminya diketik admin, bukan dihitung sistem. RajaOngkir tidak mengembalikan
-- data asuransi sama sekali — balasannya hanya berisi nama kurir, layanan,
-- ongkos, dan estimasi tiba — dan menanam rumus premi sendiri di kode berarti
-- angka yang tidak pernah ikut berubah saat kurir mengubah tarifnya.
ALTER TABLE shipments
    ADD COLUMN insurance_fee NUMERIC(18,2) NOT NULL DEFAULT 0
        CHECK (insurance_fee >= 0);
