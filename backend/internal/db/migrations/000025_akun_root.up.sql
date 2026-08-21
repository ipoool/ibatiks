-- Akun root lahir bersama skemanya, bukan lewat seed.
--
-- Ia akun pertama yang ada saat aplikasi dipasang di production: sebelum ada
-- satu pun pengguna lain, seseorang harus bisa masuk untuk membuat mereka.
-- Menaruhnya di seed berarti langkah terpisah yang bisa terlewat, dan kalau
-- terlewat tidak ada seorang pun yang bisa login.
--
-- Password-nya di-hash bcrypt cost 12 — sama seperti yang dipakai aplikasi.
-- Hash-nya tersimpan di repositori ini, jadi passwordnya harus diganti lewat
-- Pengaturan → Pengguna → Reset Password segera setelah login pertama.
--
-- ON CONFLICT supaya migrasi tetap aman dijalankan pada database yang akunnya
-- sudah pernah dibuat lewat seed.
INSERT INTO users (name, email, password_hash, role)
VALUES ('Root', 'hi@loomwarestudio.com',
        '$2a$12$CNF0uUembhZr7TcmjCEmCek/JEbQHtw3zB8ppoJT.D3KKfwe5hSn.', 'root')
ON CONFLICT (email) DO NOTHING;
