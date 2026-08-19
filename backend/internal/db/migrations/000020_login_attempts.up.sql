-- Menahan tebakan password: lima kegagalan berturut-turut mengunci sebuah email
-- selama lima menit.
--
-- Kuncinya email, bukan alamat IP. Penebak yang berpindah-pindah IP karenanya
-- tetap tertahan, dengan konsekuensi yang disadari: siapa pun yang tahu alamat
-- email seorang pengguna bisa membuatnya terkunci dengan sengaja salah lima
-- kali. Penguncian selalu lepas sendiri setelah lima menit, tidak pernah
-- permanen.
--
-- Email yang tidak terdaftar pun dicatat. Kalau hanya email terdaftar yang
-- dihitung, pola penguncian justru membocorkan email mana yang ada di sistem —
-- persis yang dihindari pesan "email atau password salah" yang seragam itu.
CREATE TABLE login_attempts (
    -- Disimpan huruf kecil supaya "Owner@Ibatiks.id" tidak jadi jalan memutar.
    email          TEXT        PRIMARY KEY CHECK (email = lower(email)),
    failed_count   INTEGER     NOT NULL DEFAULT 0 CHECK (failed_count >= 0),
    last_failed_at TIMESTAMPTZ,
    blocked_until  TIMESTAMPTZ,
    -- Alamat IP percobaan terakhir, supaya owner bisa melihat dari mana
    -- datangnya kalau akunnya terkunci berulang kali.
    last_ip        TEXT,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Dipakai membersihkan baris lama yang sudah tidak berarti apa-apa.
CREATE INDEX idx_login_attempts_last_failed ON login_attempts (last_failed_at);
