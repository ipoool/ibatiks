-- Hak akses per pengguna, di luar role.
--
-- Role tetap menentukan bawaan, tapi owner kadang perlu mengatur lebih rinci:
-- admin yang hanya boleh mengurus order tanpa melihat laporan laba, atau
-- tripper yang sekalian dititipi input pembelian. NULL berarti "ikut bawaan
-- role" — itu sebabnya kolomnya boleh kosong dan bukan diisi daftar kosong,
-- sebab daftar kosong berarti pengguna tanpa akses ke menu mana pun.
ALTER TABLE users ADD COLUMN permissions text[];

COMMENT ON COLUMN users.permissions IS
    'Daftar hak akses menu. NULL = ikut bawaan role, bukan tanpa akses.';
