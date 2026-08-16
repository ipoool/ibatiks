-- Mengembalikan identitas bawaan ke nama lama, dengan penjagaan yang sama:
-- hanya nilai yang belum disunting pemilik toko yang dikembalikan.
UPDATE app_settings SET value = 'Jastipin'
    WHERE key = 'store_name' AND value = 'Ibatiks';

UPDATE app_settings SET value = 'halo@jastipin.id'
    WHERE key = 'store_email' AND value = 'halo@ibatiks.id';

UPDATE app_settings SET value = 'BCA 1234567890 a/n Jastipin'
    WHERE key = 'bank_account' AND value = 'BCA 1234567890 a/n Ibatiks';
