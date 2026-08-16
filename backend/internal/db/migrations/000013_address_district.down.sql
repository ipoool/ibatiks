ALTER TABLE orders
    DROP COLUMN shipping_district,
    DROP COLUMN shipping_subdistrict;

ALTER TABLE customers
    DROP COLUMN district,
    DROP COLUMN subdistrict;
