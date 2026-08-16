-- Nomor WhatsApp dipakai sebagai identitas customer pada laporan penjualan per
-- customer. Tanpa jaminan keunikan, satu orang yang terlanjur dicatat dua kali
-- (mis. beda ejaan nama) akan muncul sebagai dua baris berbeda di laporan, dan
-- angka belanjanya terpecah — persis masalah yang ingin dihindari.
--
-- Indeks dibuat parsial supaya customer yang sudah dihapus tidak menahan
-- nomornya: nomor yang sama boleh dipakai lagi oleh entri baru.
CREATE UNIQUE INDEX idx_customers_phone_unique
    ON customers (phone_wa)
    WHERE deleted_at IS NULL;
