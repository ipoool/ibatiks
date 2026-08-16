-- Customer jastip. Nomor WA disimpan dalam format internasional tanpa tanda
-- baca (contoh: 6281234567890) supaya link wa.me bisa dibentuk langsung.
CREATE TABLE customers (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    phone_wa    TEXT        NOT NULL,
    email       TEXT,
    instagram   TEXT,
    address     TEXT,
    city        TEXT,
    province    TEXT,
    postal_code TEXT,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TRIGGER trg_customers_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Pencarian customer hampir selalu lewat potongan nama atau nomor HP, jadi
-- indeks trigram jauh lebih berguna di sini ketimbang b-tree biasa.
CREATE INDEX idx_customers_name_trgm ON customers USING GIN (name gin_trgm_ops);
CREATE INDEX idx_customers_phone_trgm ON customers USING GIN (phone_wa gin_trgm_ops);
CREATE INDEX idx_customers_alive ON customers (created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE product_categories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_product_categories_updated_at
    BEFORE UPDATE ON product_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Master produk. base_price adalah harga beli acuan dalam mata uang asal
-- (base_currency); harga jual sesungguhnya dihitung per trip di trip_items
-- karena kurs dan harga toko berubah tiap perjalanan.
CREATE TABLE products (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    sku           TEXT          NOT NULL UNIQUE,
    name          TEXT          NOT NULL,
    category_id   UUID          REFERENCES product_categories (id) ON DELETE SET NULL,
    brand         TEXT,
    store_name    TEXT,
    base_currency TEXT          NOT NULL DEFAULT 'IDR',
    base_price    NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (base_price >= 0),
    markup_type   TEXT          NOT NULL DEFAULT 'percent'
                                CHECK (markup_type IN ('percent', 'nominal')),
    markup_value  NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (markup_value >= 0),
    weight_gram   INTEGER       NOT NULL DEFAULT 0 CHECK (weight_gram >= 0),
    image_url     TEXT,
    notes         TEXT,
    is_active     BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_products_name_trgm ON products USING GIN (name gin_trgm_ops);
CREATE INDEX idx_products_category ON products (category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_active ON products (is_active, name) WHERE deleted_at IS NULL;
