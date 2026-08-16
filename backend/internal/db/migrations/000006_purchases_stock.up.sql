-- Realisasi belanja tripper di lapangan. Berbeda dengan trip_items yang berisi
-- rencana harga, tabel ini mencatat berapa yang benar-benar dibayar di toko.
CREATE TABLE purchases (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id           UUID          NOT NULL REFERENCES trips (id) ON DELETE CASCADE,
    product_id        UUID          NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    purchase_date     DATE          NOT NULL DEFAULT CURRENT_DATE,
    qty               INTEGER       NOT NULL CHECK (qty > 0),
    unit_cost_foreign NUMERIC(18,2) NOT NULL CHECK (unit_cost_foreign >= 0),
    currency          TEXT          NOT NULL DEFAULT 'IDR',
    -- Kurs di-snapshot per pembelian: kalau tripper belanja beberapa hari dan
    -- kurs berubah, tiap transaksi tetap membawa kurs saat itu.
    exchange_rate     NUMERIC(18,6) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
    unit_cost_idr     NUMERIC(18,2) NOT NULL CHECK (unit_cost_idr >= 0),
    total_cost_idr    NUMERIC(18,2) NOT NULL CHECK (total_cost_idr >= 0),
    store_name        TEXT,
    receipt_url       TEXT,
    notes             TEXT,
    purchased_by      UUID          REFERENCES users (id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_purchases_updated_at
    BEFORE UPDATE ON purchases
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_purchases_trip ON purchases (trip_id, purchase_date DESC);
CREATE INDEX idx_purchases_product ON purchases (product_id);

-- Inti dari perhitungan profit: ke mana tiap unit hasil belanja pergi.
-- order_item_id NULL berarti unit tersebut tidak dipesan siapa pun dan masuk
-- stok untuk dijual di marketplace.
CREATE TABLE purchase_allocations (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_id   UUID          NOT NULL REFERENCES purchases (id) ON DELETE CASCADE,
    order_item_id UUID          REFERENCES order_items (id) ON DELETE CASCADE,
    qty           INTEGER       NOT NULL CHECK (qty > 0),
    unit_cost_idr NUMERIC(18,2) NOT NULL CHECK (unit_cost_idr >= 0),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_purchase_allocations_purchase ON purchase_allocations (purchase_id);
CREATE INDEX idx_purchase_allocations_order_item ON purchase_allocations (order_item_id)
    WHERE order_item_id IS NOT NULL;

-- Posisi stok berjalan per produk, memakai moving average cost.
CREATE TABLE stock_items (
    id           UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id   UUID          NOT NULL UNIQUE REFERENCES products (id) ON DELETE CASCADE,
    qty_on_hand  INTEGER       NOT NULL DEFAULT 0 CHECK (qty_on_hand >= 0),
    avg_cost_idr NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (avg_cost_idr >= 0),
    location     TEXT,
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_stock_items_updated_at
    BEFORE UPDATE ON stock_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Buku besar pergerakan stok. qty bertanda: positif menambah, negatif mengurangi.
CREATE TABLE stock_movements (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id     UUID          NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    type           TEXT          NOT NULL
                                 CHECK (type IN ('in_purchase', 'out_order', 'out_marketplace', 'adjustment')),
    qty            INTEGER       NOT NULL CHECK (qty <> 0),
    unit_cost_idr  NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (unit_cost_idr >= 0),
    sale_price_idr NUMERIC(18,2) CHECK (sale_price_idr IS NULL OR sale_price_idr >= 0),
    trip_id        UUID          REFERENCES trips (id) ON DELETE SET NULL,
    ref_type       TEXT,
    ref_id         UUID,
    note           TEXT,
    created_by     UUID          REFERENCES users (id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_stock_movements_product ON stock_movements (product_id, created_at DESC);
CREATE INDEX idx_stock_movements_ref ON stock_movements (ref_type, ref_id);
