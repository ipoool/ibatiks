-- Integrasi tarif kirim RajaOngkir.
--
-- RajaOngkir tidak mengenal nama kota sebagai teks; tiap tujuan punya ID
-- numerik sendiri. Alamat order disimpan sebagai teks bebas, jadi ID-nya harus
-- dicari lewat API. Hasil pencarian itu disimpan di sini supaya alamat yang
-- sama tidak memicu panggilan API berulang — kuota langganan terbatas, dan
-- pemetaan kota ke ID hampir tidak pernah berubah.
CREATE TABLE shipping_destinations (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Kunci pencarian yang sudah dinormalkan, hasil dari domain.NormalizeCity
    -- atas gabungan bagian alamat yang tersedia.
    query         TEXT        NOT NULL,
    destination_id INTEGER    NOT NULL CHECK (destination_id > 0),
    label         TEXT        NOT NULL,
    city_name     TEXT,
    province_name TEXT,
    zip_code      TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT shipping_destinations_query_lower CHECK (query = lower(query)),
    CONSTRAINT shipping_destinations_query_unique UNIQUE (query)
);

CREATE TRIGGER trg_shipping_destinations_updated_at
    BEFORE UPDATE ON shipping_destinations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Setelan yang dipakai menghitung ongkir lewat RajaOngkir. Kota asal disimpan
-- sebagai ID sekaligus labelnya: ID untuk dikirim ke API, label supaya admin
-- bisa memastikan yang dipilih memang gudangnya.
INSERT INTO app_settings (key, value, description) VALUES
    ('shipping_origin_id', '',
     'ID kota asal pengiriman di RajaOngkir'),
    ('shipping_origin_label', '',
     'Nama kota asal pengiriman, hanya untuk ditampilkan'),
    ('shipping_couriers', 'jne:jnt:sicepat',
     'Kurir yang ditawarkan RajaOngkir, dipisah titik dua')
ON CONFLICT (key) DO NOTHING;
