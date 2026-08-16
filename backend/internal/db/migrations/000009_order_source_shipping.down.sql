DELETE FROM app_settings
WHERE key IN ('shipping_volumetric_divisor', 'shipping_default_price_per_kg', 'shipping_provider');

DROP TABLE IF EXISTS shipping_rates;

ALTER TABLE shipments
    DROP COLUMN IF EXISTS estimated_cost,
    DROP COLUMN IF EXISTS height_cm,
    DROP COLUMN IF EXISTS width_cm,
    DROP COLUMN IF EXISTS length_cm;

DROP INDEX IF EXISTS idx_orders_source;

ALTER TABLE orders DROP COLUMN IF EXISTS order_source;
