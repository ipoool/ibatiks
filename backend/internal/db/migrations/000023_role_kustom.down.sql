-- Mengembalikan role jadi daftar tertutup di dalam kode.
--
-- Role bikinan toko tidak punya padanan di daftar lama, jadi penggunanya
-- dipindahkan ke role bawaan yang paling mendekati wewenangnya: yang ber-scope
-- 'field' jadi tripper, sisanya jadi admin. Root jadi owner. Perpindahan ini
-- tidak bisa dibalik lagi oleh migrasi naiknya — nama role aslinya hilang.
UPDATE users u
SET role = CASE
        WHEN u.role = 'root' THEN 'owner'
        WHEN r.scope = 'field' THEN 'tripper'
        ELSE 'admin'
    END
FROM roles r
WHERE r.name = u.role
  AND u.role NOT IN ('owner', 'admin', 'tripper');

-- reports_finance tidak dikenal skema lama; laba-rugi kembali dijaga oleh nama
-- role, jadi hak aksesnya dibuang dari daftar tiap pengguna.
UPDATE users
SET permissions = array_remove(permissions, 'reports_finance')
WHERE permissions IS NOT NULL
  AND 'reports_finance' = ANY (permissions);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('owner', 'admin', 'tripper'));

DROP TABLE roles;
