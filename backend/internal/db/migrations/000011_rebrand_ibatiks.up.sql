-- Rebranding Jastipin -> Ibatiks.
--
-- Hanya nilai yang masih sama persis dengan bawaan lama yang diganti. Toko yang
-- sudah menyunting sendiri nama, email, atau rekeningnya tidak boleh tertimpa
-- oleh migrasi: yang tertulis di invoice mereka adalah keputusan pemiliknya,
-- bukan bawaan aplikasi.
UPDATE app_settings SET value = 'Ibatiks'
    WHERE key = 'store_name' AND value = 'Jastipin';

UPDATE app_settings SET value = 'halo@ibatiks.id'
    WHERE key = 'store_email' AND value = 'halo@jastipin.id';

UPDATE app_settings SET value = 'BCA 1234567890 a/n Ibatiks'
    WHERE key = 'bank_account' AND value = 'BCA 1234567890 a/n Jastipin';
