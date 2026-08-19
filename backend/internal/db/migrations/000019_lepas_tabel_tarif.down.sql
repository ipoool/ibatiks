-- Mengembalikan struktur tabel tarif beserta pengaturannya.
--
-- Isinya tidak bisa dipulihkan: baris tarif per kota sudah terbuang, dan basis
-- data tidak menyimpan salinannya. Yang dijamin turunnya hanyalah versi
-- sebelumnya bisa berjalan lagi tanpa tabel yang hilang — tarifnya sendiri harus
-- diisi ulang lewat menu Pengaturan.
CREATE TABLE IF NOT EXISTS shipping_rates (
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

    CONSTRAINT shipping_rates_city_lower CHECK (destination_city = lower(destination_city)),
    CONSTRAINT shipping_rates_unique UNIQUE (courier, service, destination_city)
);

CREATE INDEX IF NOT EXISTS idx_shipping_rates_city ON shipping_rates (destination_city);

CREATE TRIGGER trg_shipping_rates_updated_at
    BEFORE UPDATE ON shipping_rates
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO app_settings (key, value, description) VALUES
    ('shipping_volumetric_divisor', '6000',
     'Pembagi berat volumetrik: (P x L x T dalam cm) / pembagi = kg. JNE memakai 6000'),
    ('shipping_default_price_per_kg', '25000',
     'Tarif per kg yang dipakai kalau kota tujuan belum ada di tabel tarif'),
    ('shipping_provider', 'internal',
     'Sumber estimasi ongkir: internal (tabel tarif) atau nama vendor jika sudah diintegrasikan')
ON CONFLICT (key) DO NOTHING;
