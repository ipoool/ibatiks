-- Dari mana order itu datang. Dibutuhkan untuk rekap penjualan per channel,
-- karena biaya promosi dan cara melayaninya berbeda antara WhatsApp dan
-- Instagram.
ALTER TABLE orders
    ADD COLUMN order_source TEXT NOT NULL DEFAULT 'whatsapp'
        CHECK (order_source IN ('whatsapp', 'instagram', 'tiktok', 'marketplace', 'lainnya'));

CREATE INDEX idx_orders_source ON orders (order_source, order_date DESC);

-- Dimensi paket untuk menghitung berat volumetrik. Ekspedisi menagih
-- berdasarkan yang lebih besar antara berat asli dan berat volumetrik, jadi
-- kardus besar berisi barang ringan tetap mahal.
ALTER TABLE shipments
    ADD COLUMN length_cm INTEGER NOT NULL DEFAULT 0 CHECK (length_cm >= 0),
    ADD COLUMN width_cm  INTEGER NOT NULL DEFAULT 0 CHECK (width_cm  >= 0),
    ADD COLUMN height_cm INTEGER NOT NULL DEFAULT 0 CHECK (height_cm >= 0),
    -- Ongkir hasil hitungan sistem, disimpan terpisah dari shipping_cost yang
    -- berisi angka yang benar-benar dibayar ke kurir. Selisih keduanya berguna
    -- untuk mengoreksi tarif yang sudah usang.
    ADD COLUMN estimated_cost NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (estimated_cost >= 0);

-- Tarif kirim per kota tujuan. Tabel ini adalah sumber estimasi ongkir bawaan.
-- Kalau nanti dipasang integrasi API kurir, tabel ini tetap dipakai sebagai
-- cadangan ketika API sedang tidak bisa dihubungi.
CREATE TABLE shipping_rates (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    courier          TEXT          NOT NULL DEFAULT 'JNE',
    service          TEXT          NOT NULL DEFAULT 'REG',
    destination_city TEXT          NOT NULL,
    province         TEXT,
    price_per_kg     NUMERIC(18,2) NOT NULL CHECK (price_per_kg >= 0),
    min_weight_gram  INTEGER       NOT NULL DEFAULT 1000 CHECK (min_weight_gram > 0),
    etd              TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),

    -- Kota disimpan huruf kecil agar pencocokan tidak bergantung cara mengetik.
    CONSTRAINT shipping_rates_city_lower CHECK (destination_city = lower(destination_city)),
    CONSTRAINT shipping_rates_unique UNIQUE (courier, service, destination_city)
);

CREATE TRIGGER trg_shipping_rates_updated_at
    BEFORE UPDATE ON shipping_rates
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_shipping_rates_city ON shipping_rates (destination_city);

-- Tarif awal untuk kota-kota tujuan yang paling sering. Angkanya perkiraan
-- kasar dan wajib disesuaikan sendiri oleh pemilik toko lewat menu Pengaturan.
INSERT INTO shipping_rates (courier, service, destination_city, province, price_per_kg, etd) VALUES
    ('JNE', 'REG', 'jakarta',        'DKI Jakarta',      12000, '1-2 hari'),
    ('JNE', 'REG', 'jakarta selatan','DKI Jakarta',      12000, '1-2 hari'),
    ('JNE', 'REG', 'jakarta pusat',  'DKI Jakarta',      12000, '1-2 hari'),
    ('JNE', 'REG', 'jakarta barat',  'DKI Jakarta',      12000, '1-2 hari'),
    ('JNE', 'REG', 'jakarta timur',  'DKI Jakarta',      12000, '1-2 hari'),
    ('JNE', 'REG', 'jakarta utara',  'DKI Jakarta',      12000, '1-2 hari'),
    ('JNE', 'REG', 'tangerang',      'Banten',           13000, '1-2 hari'),
    ('JNE', 'REG', 'bekasi',         'Jawa Barat',       13000, '1-2 hari'),
    ('JNE', 'REG', 'depok',          'Jawa Barat',       13000, '1-2 hari'),
    ('JNE', 'REG', 'bogor',          'Jawa Barat',       14000, '1-2 hari'),
    ('JNE', 'REG', 'bandung',        'Jawa Barat',       16000, '2-3 hari'),
    ('JNE', 'REG', 'semarang',       'Jawa Tengah',      18000, '2-3 hari'),
    ('JNE', 'REG', 'yogyakarta',     'DI Yogyakarta',    19000, '2-3 hari'),
    ('JNE', 'REG', 'surabaya',       'Jawa Timur',       21000, '2-3 hari'),
    ('JNE', 'REG', 'malang',         'Jawa Timur',       22000, '2-4 hari'),
    ('JNE', 'REG', 'denpasar',       'Bali',             26000, '3-4 hari'),
    ('JNE', 'REG', 'medan',          'Sumatera Utara',   32000, '3-5 hari'),
    ('JNE', 'REG', 'palembang',      'Sumatera Selatan', 28000, '3-4 hari'),
    ('JNE', 'REG', 'pekanbaru',      'Riau',             33000, '3-5 hari'),
    ('JNE', 'REG', 'makassar',       'Sulawesi Selatan', 35000, '3-5 hari'),
    ('JNE', 'REG', 'balikpapan',     'Kalimantan Timur', 36000, '3-5 hari'),
    ('JNE', 'REG', 'banjarmasin',    'Kalimantan Selatan', 34000, '3-5 hari'),
    ('JNE', 'YES', 'jakarta',        'DKI Jakarta',      22000, '1 hari'),
    ('JNE', 'YES', 'bandung',        'Jawa Barat',       28000, '1 hari'),
    ('JNE', 'YES', 'surabaya',       'Jawa Timur',       35000, '1 hari'),
    ('JNE', 'OKE', 'jakarta',        'DKI Jakarta',       9000, '2-3 hari'),
    ('JNE', 'OKE', 'bandung',        'Jawa Barat',       13000, '3-4 hari'),
    ('JNE', 'OKE', 'surabaya',       'Jawa Timur',       17000, '3-5 hari');

-- Pengaturan pendukung perhitungan ongkir.
INSERT INTO app_settings (key, value, description) VALUES
    ('shipping_volumetric_divisor', '6000',
     'Pembagi berat volumetrik: (P x L x T dalam cm) / pembagi = kg. JNE memakai 6000'),
    ('shipping_default_price_per_kg', '25000',
     'Tarif per kg yang dipakai kalau kota tujuan belum ada di tabel tarif'),
    ('shipping_provider', 'internal',
     'Sumber estimasi ongkir: internal (tabel tarif) atau nama vendor jika sudah diintegrasikan');
