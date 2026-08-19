-- Satu customer bisa dihubungi lewat lebih dari satu akun: Instagram untuk
-- melihat katalog, TikTok tempat ia menemukan toko, Telegram untuk mengirim
-- foto struk. Sebelumnya hanya Instagram yang punya tempat, dan sisanya
-- terselip di kolom catatan tempat tidak ada yang mencarinya.
--
-- Disimpan sebagai JSONB, bukan tabel tersendiri. Isinya cuma dibaca sebagai
-- satu kesatuan bersama customernya, tidak pernah diagregasi maupun dijadikan
-- syarat penyaringan; tabel terpisah berarti join pada tiap pembacaan customer
-- untuk data yang tidak pernah butuh dijoin.
ALTER TABLE customers ADD COLUMN socials JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Akun Instagram yang sudah tercatat dipindahkan apa adanya, jadi tidak ada
-- data yang hilang saat kolom lamanya dilepas.
UPDATE customers
SET socials = jsonb_build_array(
        jsonb_build_object('platform', 'instagram', 'handle', trim(instagram))
    )
WHERE instagram IS NOT NULL AND trim(instagram) <> '';

ALTER TABLE customers DROP COLUMN instagram;

-- Penjagaan bentuk: harus array. Isi tiap elemennya divalidasi di service,
-- bukan di sini — pesan galat dari CHECK constraint tidak bisa menyebut baris
-- keberapa yang salah.
ALTER TABLE customers ADD CONSTRAINT customers_socials_array
    CHECK (jsonb_typeof(socials) = 'array');
