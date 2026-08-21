-- Role berhenti jadi daftar tertutup di dalam kode dan pindah jadi data.
--
-- Sebelumnya nama role ditulis langsung di penjaga rute (ownerOnly,
-- staffOnly), jadi role bikinan toko sendiri akan ditolak seluruh endpoint
-- operasional walaupun menunya sudah dicentang. Yang dibaca penjaga sekarang
-- adalah kolom scope, yang ikut dimiliki role baru.
--
-- scope 'full'  = staf toko, boleh mengubah data
-- scope 'field' = petugas lapangan, hanya membaca produk dan mencatat belanja
CREATE TABLE roles (
    name        TEXT        PRIMARY KEY,
    label       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    scope       TEXT        NOT NULL DEFAULT 'full'
                            CHECK (scope IN ('full', 'field')),
    permissions TEXT[]      NOT NULL DEFAULT '{}',
    -- Role bawaan tidak bisa dihapus atau diganti namanya: nama-namanya
    -- dirujuk kode untuk penjagaan khusus (owner terakhir, kunci root).
    is_system   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Laporan laba-rugi dulu dijaga "khusus owner" lewat nama role, tanpa hak akses
-- tersendiri. Begitu role jadi data, penjagaan itu tidak punya pegangan lagi —
-- jadi ia diangkat menjadi hak akses biasa yang bisa dicentang.
INSERT INTO roles (name, label, description, scope, permissions, is_system) VALUES
    ('root', 'Root', 'Akses penuh ke seluruh menu tanpa kecuali', 'full', ARRAY[
        'trips', 'shopping_list', 'purchases', 'orders', 'invoices', 'shipments',
        'customers', 'products', 'stock', 'reports', 'reports_finance',
        'settings', 'users'
    ], TRUE),
    ('owner', 'Owner', 'Pemilik toko: seluruh menu termasuk laba-rugi dan pengguna', 'full', ARRAY[
        'trips', 'shopping_list', 'purchases', 'orders', 'invoices', 'shipments',
        'customers', 'products', 'stock', 'reports', 'reports_finance',
        'settings', 'users'
    ], TRUE),
    ('admin', 'Admin', 'Seluruh operasional harian: trip, order, invoice, kirim, stok', 'full', ARRAY[
        'trips', 'shopping_list', 'purchases', 'orders', 'invoices', 'shipments',
        'customers', 'products', 'stock', 'reports'
    ], TRUE),
    ('tripper', 'Tripper', 'Petugas lapangan: daftar belanja dan input pembelian', 'field', ARRAY[
        'trips', 'shopping_list', 'purchases', 'products'
    ], TRUE);

-- Owner yang hak aksesnya pernah dipersempit sendiri tetap menyimpan daftarnya
-- di users.permissions. Daftar itu dibuat sebelum reports_finance ada, jadi
-- tanpa penambahan ini mereka kehilangan tab Profit / Loss yang selama ini
-- terbuka untuknya.
UPDATE users
SET permissions = permissions || ARRAY['reports_finance']
WHERE role = 'owner'
  AND permissions IS NOT NULL
  AND 'reports' = ANY (permissions)
  AND NOT ('reports_finance' = ANY (permissions));

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey FOREIGN KEY (role) REFERENCES roles (name)
    ON UPDATE CASCADE ON DELETE RESTRICT;

COMMENT ON TABLE roles IS
    'Role pengguna beserta menu bawaannya. Centang per pengguna di users.permissions hanya boleh mempersempit daftar ini.';
COMMENT ON COLUMN roles.scope IS
    'full = staf toko (boleh mengubah data), field = petugas lapangan (produk hanya baca).';
