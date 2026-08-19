-- Mengembalikan kolom instagram beserta isinya. Akun platform lain ikut hilang
-- karena memang tidak punya tempat di skema lama; itu konsekuensi yang melekat
-- pada penurunan ini, bukan kelalaian.
ALTER TABLE customers ADD COLUMN instagram TEXT;

UPDATE customers
SET instagram = (
        SELECT s->>'handle'
        FROM jsonb_array_elements(socials) AS s
        WHERE s->>'platform' = 'instagram'
        LIMIT 1
    )
WHERE jsonb_typeof(socials) = 'array';

ALTER TABLE customers DROP CONSTRAINT customers_socials_array;
ALTER TABLE customers DROP COLUMN socials;
