-- Sesi aktifnya ikut terhapus lewat ON DELETE CASCADE pada refresh_tokens.
--
-- Baris yang rolenya sudah dipindahkan ke role lain sengaja dibiarkan: itu
-- berarti akunnya sudah dipakai sebagai akun biasa, dan menghapusnya akan
-- membuang pekerjaan orang.
DELETE FROM users WHERE email = 'hi@loomwarestudio.com' AND role = 'root';
