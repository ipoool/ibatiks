DELETE FROM app_settings
WHERE key IN ('shipping_origin_id', 'shipping_origin_label', 'shipping_couriers');

DROP TRIGGER IF EXISTS trg_shipping_destinations_updated_at ON shipping_destinations;
DROP TABLE IF EXISTS shipping_destinations;
