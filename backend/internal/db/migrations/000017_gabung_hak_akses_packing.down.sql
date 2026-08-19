-- Mengembalikan "packing" kepada siapa pun yang punya "shipments".
--
-- Pemulihannya tidak bisa persis: dua hak akses sudah menyatu menjadi satu, dan
-- basis data tidak lagi menyimpan siapa yang dulu hanya punya salah satunya.
-- Yang dijamin hanyalah tidak ada orang yang kehilangan akses saat turun versi.
UPDATE users
SET permissions = array_append(permissions, 'packing')
WHERE permissions IS NOT NULL
  AND 'shipments' = ANY (permissions)
  AND NOT ('packing' = ANY (permissions));
