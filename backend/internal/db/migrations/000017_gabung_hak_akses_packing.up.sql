-- Antrean kemas digabung menjadi tab di dalam menu Pengiriman, jadi hak akses
-- "packing" tidak punya menu lagi untuk dijaga.
--
-- Baris yang menyimpannya diterjemahkan ke "shipments", bukan sekadar dibuang:
-- pengguna yang dulu hanya dicentang "packing" masih punya pekerjaan yang sama
-- persis, dan membuang centangnya begitu saja akan mencabut satu-satunya menu
-- yang mereka pakai tiap hari.
UPDATE users
SET permissions = (
    SELECT array_agg(DISTINCT CASE WHEN hak = 'packing' THEN 'shipments' ELSE hak END)
    FROM unnest(permissions) AS hak
)
WHERE permissions IS NOT NULL
  AND 'packing' = ANY (permissions);
