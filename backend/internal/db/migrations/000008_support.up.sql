-- Penomoran dokumen (TRP/ORD/INV/CUS) per tahun. Nomor diambil dengan
-- SELECT ... FOR UPDATE di dalam transaksi yang sama dengan insert dokumennya,
-- sehingga dua admin yang menyimpan bersamaan tidak bisa mendapat nomor kembar.
CREATE TABLE document_counters (
    doc_type    TEXT        NOT NULL,
    year        INTEGER     NOT NULL,
    last_number INTEGER     NOT NULL DEFAULT 0 CHECK (last_number >= 0),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (doc_type, year)
);

-- Jejak perubahan. Order dan invoice boleh diedit, jadi harus ada catatan
-- siapa mengubah apa — terutama untuk perubahan qty dan nominal.
CREATE TABLE audit_logs (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    UUID        REFERENCES users (id) ON DELETE SET NULL,
    entity     TEXT        NOT NULL,
    entity_id  UUID,
    action     TEXT        NOT NULL,
    changes    JSONB,
    ip_address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_entity ON audit_logs (entity, entity_id, created_at DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs (user_id, created_at DESC);

-- Pengaturan toko: identitas di invoice, rekening, dan template pesan WA.
CREATE TABLE app_settings (
    key         TEXT        PRIMARY KEY,
    value       TEXT        NOT NULL,
    description TEXT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_app_settings_updated_at
    BEFORE UPDATE ON app_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Nilai awal. Placeholder {{...}} diganti saat pesan dibentuk oleh backend.
INSERT INTO app_settings (key, value, description) VALUES
    ('store_name',        'Jastipin',
     'Nama toko yang tampil di header invoice'),
    ('store_phone',       '6281234567890',
     'Nomor WA toko (format internasional tanpa +)'),
    ('store_email',       'halo@jastipin.id',
     'Email toko yang tampil di invoice'),
    ('store_address',     'Jakarta, Indonesia',
     'Alamat toko yang tampil di invoice'),
    ('bank_account',      'BCA 1234567890 a/n Jastipin',
     'Rekening tujuan pembayaran yang tampil di invoice'),
    ('invoice_footer',    'Terima kasih sudah titip belanja bersama kami!',
     'Catatan penutup pada invoice'),
    ('invoice_due_days',  '3',
     'Jatuh tempo invoice pelunasan dalam hari sejak diterbitkan'),
    ('wa_template_dp',
     'Halo {{customer_name}}, terima kasih sudah order di trip {{trip_title}}.'
     || E'\nTotal pesanan: {{total}}'
     || E'\nDP yang perlu ditransfer: {{dp_amount}}'
     || E'\nTransfer ke: {{bank_account}}'
     || E'\nMohon konfirmasi setelah transfer ya. Terima kasih!',
     'Template pesan WA permintaan DP'),
    ('wa_template_invoice',
     'Halo {{customer_name}}, barang pesananmu sudah sampai di Indonesia.'
     || E'\nInvoice {{invoice_number}}'
     || E'\nTotal: {{total}}'
     || E'\nSudah dibayar: {{amount_paid}}'
     || E'\nSisa pelunasan: {{amount_due}}'
     || E'\nTransfer ke: {{bank_account}}'
     || E'\nSetelah lunas barang langsung kami kirim. Terima kasih!',
     'Template pesan WA penagihan pelunasan'),
    ('wa_template_shipped',
     'Halo {{customer_name}}, pesananmu {{order_number}} sudah dikirim.'
     || E'\nKurir: {{courier}} {{service}}'
     || E'\nNo. resi: {{tracking_number}}'
     || E'\nCek posisi paket di https://www.jne.co.id/tracking-package'
     || E'\nTerima kasih sudah belanja bersama kami!',
     'Template pesan WA informasi pengiriman');
