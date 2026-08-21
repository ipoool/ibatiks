-- Dashboard kembali menumpang hak Laporan.
--
-- Role yang memegang Dashboard tanpa Laporan kehilangan halaman itu setelah
-- migrasi ini — tidak ada hak lain yang bisa mewakilinya di skema lama.
UPDATE roles
SET permissions = array_remove(permissions, 'dashboard')
WHERE 'dashboard' = ANY (permissions);

UPDATE users
SET permissions = array_remove(permissions, 'dashboard')
WHERE permissions IS NOT NULL
  AND 'dashboard' = ANY (permissions);
