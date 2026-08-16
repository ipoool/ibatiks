-- Ekstensi dan helper yang dipakai seluruh skema.

-- gen_random_uuid() sebenarnya sudah bawaan sejak PG13, tapi pgcrypto tetap
-- dipasang untuk fungsi hash yang mungkin dipakai belakangan.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- pg_trgm dipakai untuk pencarian nama/telepon customer yang toleran typo.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- citext supaya email tidak perlu di-lower() setiap kali dibandingkan.
CREATE EXTENSION IF NOT EXISTS citext;

-- Satu trigger function dipakai ulang oleh semua tabel yang punya updated_at,
-- supaya kolom itu tidak pernah lupa di-update dari sisi aplikasi.
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
