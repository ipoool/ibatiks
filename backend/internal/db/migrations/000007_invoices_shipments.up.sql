-- Invoice yang dikirim ke customer. Satu order bisa punya invoice DP dan
-- invoice pelunasan, karena itu tidak ada unique constraint pada order_id.
CREATE TABLE invoices (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number TEXT          NOT NULL UNIQUE,
    order_id       UUID          NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    type           TEXT          NOT NULL CHECK (type IN ('dp', 'final')),
    issue_date     DATE          NOT NULL DEFAULT CURRENT_DATE,
    due_date       DATE,

    -- Seluruh nominal di-snapshot saat invoice dibuat supaya PDF yang sudah
    -- terkirim ke customer tidak berubah isinya kalau order diedit belakangan.
    subtotal       NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    discount       NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (discount >= 0),
    shipping_fee   NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (shipping_fee >= 0),
    total          NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (total >= 0),
    amount_paid    NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (amount_paid >= 0),
    amount_due     NUMERIC(18,2) NOT NULL DEFAULT 0,

    status         TEXT          NOT NULL DEFAULT 'draft'
                                 CHECK (status IN ('draft', 'sent', 'paid', 'void')),
    pdf_path       TEXT,
    sent_channel   TEXT          CHECK (sent_channel IS NULL OR sent_channel IN ('wa', 'email', 'manual')),
    sent_at        TIMESTAMPTZ,
    paid_at        TIMESTAMPTZ,
    notes          TEXT,
    created_by     UUID          REFERENCES users (id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_invoices_order ON invoices (order_id, created_at DESC);
CREATE INDEX idx_invoices_status ON invoices (status, issue_date DESC);

-- Paket pengiriman. Satu order dikemas jadi satu paket, sehingga order_id unik.
-- Resi dan ongkir diisi manual oleh admin (tanpa integrasi API kurir).
CREATE TABLE shipments (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id             UUID          NOT NULL UNIQUE REFERENCES orders (id) ON DELETE CASCADE,
    courier              TEXT          NOT NULL DEFAULT 'JNE',
    service              TEXT          NOT NULL DEFAULT 'REG',
    tracking_number      TEXT,
    weight_gram          INTEGER       NOT NULL DEFAULT 0 CHECK (weight_gram >= 0),
    shipping_cost        NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (shipping_cost >= 0),
    status               TEXT          NOT NULL DEFAULT 'packing'
                                       CHECK (status IN ('packing', 'ready', 'shipped', 'delivered', 'returned')),
    packed_at            TIMESTAMPTZ,
    packed_by            UUID          REFERENCES users (id) ON DELETE SET NULL,
    shipped_at           TIMESTAMPTZ,
    delivered_at         TIMESTAMPTZ,
    customer_notified_at TIMESTAMPTZ,
    notes                TEXT,
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),

    -- Status 'shipped' tanpa resi berarti admin lupa mengisi nomor resi;
    -- database menolaknya supaya customer tidak pernah dikabari resi kosong.
    CONSTRAINT shipments_tracking_required CHECK (
        status NOT IN ('shipped', 'delivered') OR tracking_number IS NOT NULL
    )
);

CREATE TRIGGER trg_shipments_updated_at
    BEFORE UPDATE ON shipments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_shipments_status ON shipments (status, created_at DESC);
CREATE INDEX idx_shipments_tracking ON shipments (tracking_number)
    WHERE tracking_number IS NOT NULL;
