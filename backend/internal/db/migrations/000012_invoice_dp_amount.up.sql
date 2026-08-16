-- Invoice menampilkan nilai order seutuhnya, dengan uang muka sebagai barisnya
-- sendiri.
--
-- Sebelumnya invoice DP menyimpan nilai uang muka pada kolom subtotal dan
-- total, sehingga customer menerima dokumen yang seolah-olah menyatakan harga
-- pesanannya hanya sebesar DP. Sekarang subtotal/total selalu berisi nilai
-- order, dan besaran DP punya kolom sendiri.
ALTER TABLE invoices ADD COLUMN dp_amount NUMERIC(18,2) NOT NULL DEFAULT 0;

-- Invoice DP lama: nilai yang tersimpan di total sebenarnya adalah uang muka.
UPDATE invoices SET dp_amount = total WHERE type = 'dp';

UPDATE invoices i
SET subtotal     = o.subtotal,
    discount     = o.discount,
    shipping_fee = o.shipping_fee,
    total        = o.total
FROM orders o
WHERE i.order_id = o.id AND i.type = 'dp';

-- Invoice pelunasan lama: uang muka yang sudah diterima dicatat sebagai acuan
-- baris "Down payment".
UPDATE invoices i
SET dp_amount = COALESCE((
        SELECT sum(p.amount) FROM payments p
        WHERE p.order_id = i.order_id AND p.type = 'dp'
    ), 0)
WHERE i.type = 'final';
