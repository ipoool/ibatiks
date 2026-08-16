-- Mengembalikan invoice DP ke bentuk lama: nilai uang muka kembali ditulis
-- sebagai subtotal dan total invoice.
UPDATE invoices
SET subtotal     = dp_amount,
    discount     = 0,
    shipping_fee = 0,
    total        = dp_amount
WHERE type = 'dp';

ALTER TABLE invoices DROP COLUMN dp_amount;
