-- Order jastip milik satu customer pada satu trip.
--
-- Alamat pengiriman disalin (snapshot) ke sini, tidak sekadar mereferensikan
-- customers: kalau customer pindah rumah tahun depan, order lama tetap
-- menyimpan alamat ke mana barang benar-benar dikirim dulu.
CREATE TABLE orders (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number         TEXT          NOT NULL UNIQUE,
    trip_id              UUID          NOT NULL REFERENCES trips (id) ON DELETE RESTRICT,
    customer_id          UUID          NOT NULL REFERENCES customers (id) ON DELETE RESTRICT,
    order_date           DATE          NOT NULL DEFAULT CURRENT_DATE,
    status               TEXT          NOT NULL DEFAULT 'draft'
                                       CHECK (status IN ('draft', 'awaiting_dp', 'dp_paid', 'purchasing',
                                                         'arrived', 'packed', 'invoiced', 'paid',
                                                         'shipped', 'completed', 'cancelled')),

    subtotal             NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    discount             NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (discount >= 0),
    shipping_fee         NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (shipping_fee >= 0),
    total                NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (total >= 0),
    dp_required          NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (dp_required >= 0),
    paid_amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    -- Sisa tagihan selalu turunan dari total dan pembayaran, jadi dihitung oleh
    -- database agar tidak mungkin melenceng dari kedua kolom sumbernya.
    balance_due          NUMERIC(18,2) GENERATED ALWAYS AS (total - paid_amount) STORED,

    recipient_name       TEXT          NOT NULL,
    recipient_phone      TEXT          NOT NULL,
    shipping_address     TEXT          NOT NULL,
    shipping_city        TEXT          NOT NULL,
    shipping_province    TEXT,
    shipping_postal_code TEXT,

    notes                TEXT,
    cancel_reason        TEXT,
    cancelled_at         TIMESTAMPTZ,
    created_by           UUID          REFERENCES users (id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_orders_trip_status ON orders (trip_id, status);
CREATE INDEX idx_orders_customer ON orders (customer_id, order_date DESC);
CREATE INDEX idx_orders_number ON orders (order_number);
-- Indeks parsial untuk laporan piutang: hanya order yang masih punya sisa bayar.
CREATE INDEX idx_orders_receivable ON orders (balance_due DESC) WHERE balance_due > 0;

-- Baris pesanan. product_name/sku ikut disalin supaya invoice lama tetap
-- menampilkan nama produk sebagaimana saat dipesan.
CREATE TABLE order_items (
    id                 UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id           UUID          NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id         UUID          NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    trip_item_id       UUID          REFERENCES trip_items (id) ON DELETE SET NULL,
    product_name       TEXT          NOT NULL,
    product_sku        TEXT          NOT NULL,
    qty                INTEGER       NOT NULL CHECK (qty > 0),
    unit_price         NUMERIC(18,2) NOT NULL CHECK (unit_price >= 0),
    -- Estimasi modal saat order dibuat. HPP sesungguhnya diambil dari
    -- purchase_allocations setelah tripper benar-benar belanja.
    unit_cost_est      NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (unit_cost_est >= 0),
    subtotal           NUMERIC(18,2) GENERATED ALWAYS AS (qty * unit_price) STORED,
    qty_purchased      INTEGER       NOT NULL DEFAULT 0 CHECK (qty_purchased >= 0),
    qty_received       INTEGER       NOT NULL DEFAULT 0 CHECK (qty_received >= 0),
    fulfillment_status TEXT          NOT NULL DEFAULT 'pending'
                                     CHECK (fulfillment_status IN ('pending', 'purchased', 'partial',
                                                                   'unavailable', 'refunded')),
    notes              TEXT,
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_order_items_updated_at
    BEFORE UPDATE ON order_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_order_items_order ON order_items (order_id);
CREATE INDEX idx_order_items_product ON order_items (product_id);

-- Semua uang masuk/keluar yang terkait order: DP, pelunasan, refund, koreksi.
-- amount selalu positif; arah uangnya ditentukan oleh kolom type.
CREATE TABLE payments (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID          NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    type        TEXT          NOT NULL
                              CHECK (type IN ('dp', 'settlement', 'refund', 'adjustment')),
    amount      NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    method      TEXT          NOT NULL DEFAULT 'transfer'
                              CHECK (method IN ('transfer', 'cash', 'qris', 'ewallet', 'lainnya')),
    reference   TEXT,
    proof_url   TEXT,
    paid_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    notes       TEXT,
    recorded_by UUID          REFERENCES users (id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order ON payments (order_id, paid_at DESC);
