-- Satu baris = satu perjalanan jastip ke luar negeri.
-- exchange_rate dikunci di level trip: semua harga jual dan HPP pada trip ini
-- memakai kurs yang sama, sehingga laporan profit tidak berubah kalau kurs
-- pasar bergerak setelah trip selesai.
CREATE TABLE trips (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    code            TEXT          NOT NULL UNIQUE,
    title           TEXT          NOT NULL,
    country         TEXT          NOT NULL,
    city            TEXT,
    tripper_user_id UUID          REFERENCES users (id) ON DELETE SET NULL,
    depart_date     DATE          NOT NULL,
    return_date     DATE          NOT NULL,
    order_deadline  DATE,
    currency        TEXT          NOT NULL DEFAULT 'IDR',
    exchange_rate   NUMERIC(18,6) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
    status          TEXT          NOT NULL DEFAULT 'draft'
                                  CHECK (status IN ('draft', 'open', 'closed', 'shopping',
                                                    'in_transit', 'arrived', 'settled', 'cancelled')),
    notes           TEXT,
    created_by      UUID          REFERENCES users (id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT trips_date_order CHECK (return_date >= depart_date)
);

CREATE TRIGGER trg_trips_updated_at
    BEFORE UPDATE ON trips
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_trips_status ON trips (status, depart_date DESC);
CREATE INDEX idx_trips_tripper ON trips (tripper_user_id);

-- Katalog produk yang ditawarkan pada satu trip, lengkap dengan harga modal
-- dan markup-nya. sell_price disimpan (bukan dihitung saat render) supaya
-- perubahan markup di kemudian hari tidak mengubah harga yang sudah terlanjur
-- dipublikasikan ke customer.
CREATE TABLE trip_items (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id        UUID          NOT NULL REFERENCES trips (id) ON DELETE CASCADE,
    product_id     UUID          NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    cost_price     NUMERIC(18,2) NOT NULL CHECK (cost_price >= 0),
    cost_price_idr NUMERIC(18,2) NOT NULL CHECK (cost_price_idr >= 0),
    markup_type    TEXT          NOT NULL DEFAULT 'percent'
                                 CHECK (markup_type IN ('percent', 'nominal')),
    markup_value   NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (markup_value >= 0),
    sell_price     NUMERIC(18,2) NOT NULL CHECK (sell_price >= 0),
    max_qty        INTEGER       CHECK (max_qty IS NULL OR max_qty > 0),
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    notes          TEXT,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT trip_items_unique_product UNIQUE (trip_id, product_id)
);

CREATE TRIGGER trg_trip_items_updated_at
    BEFORE UPDATE ON trip_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_trip_items_product ON trip_items (product_id);

-- Modal perjalanan di luar harga barang: tiket, bagasi, akomodasi, dsb.
-- Nilai ini dikurangkan dari profit trip.
CREATE TABLE trip_expenses (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id     UUID          NOT NULL REFERENCES trips (id) ON DELETE CASCADE,
    category    TEXT          NOT NULL
                              CHECK (category IN ('tiket', 'bagasi', 'akomodasi',
                                                  'transport', 'visa', 'lainnya')),
    description TEXT          NOT NULL,
    amount      NUMERIC(18,2) NOT NULL CHECK (amount >= 0),
    spent_at    DATE          NOT NULL DEFAULT CURRENT_DATE,
    receipt_url TEXT,
    created_by  UUID          REFERENCES users (id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_trip_expenses_updated_at
    BEFORE UPDATE ON trip_expenses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_trip_expenses_trip ON trip_expenses (trip_id, spent_at DESC);
