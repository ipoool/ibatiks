/*
 * Kunci pengaturan yang punya form khusus, dikelompokkan agar mudah dibaca.
 *
 * Modul biasa, bukan bagian dari halamannya, karena labelnya dipakai dua tempat:
 * form pengaturan dan Jejak Perubahan — yang menampilkan kunci apa saja yang
 * ikut berubah. Menyalinnya ke dua tempat berarti daftar yang cepat atau lambat
 * berbeda, dan jejak perubahan akan menyebut nama kolom yang sudah tidak ada.
 */
export const STORE_FIELDS = [
  {
    key: "store_name",
    label: "Nama toko",
    hint: "Tampil di header invoice",
    required: true,
    minLength: 2,
    maxLength: 120,
  },
  {
    key: "store_phone",
    label: "Nomor WA toko",
    hint: "Format internasional, contoh 6281234567890",
    inputMode: "tel" as const,
    // Angka saja, boleh diawali tanda plus. Sengaja tidak memaksa awalan 62:
    // nomornya cuma dicetak di invoice dan label, bukan dijadikan tautan wa.me,
    // dan menolak nomor yang sah karena bentuknya berbeda lebih merugikan.
    pattern: "\\+?[0-9]{8,20}",
    title: "Nomor WA hanya berisi angka, boleh diawali tanda +. Contoh: 6281234567890",
  },
  {
    key: "store_email",
    label: "Email toko",
    type: "email" as const,
    placeholder: "halo@tokokamu.id",
  },
  { key: "store_address", label: "Alamat toko", maxLength: 300 },
  {
    key: "bank_account",
    label: "Rekening pembayaran",
    hint: "Tampil di invoice dan pesan penagihan",
    maxLength: 200,
  },
  { key: "invoice_footer", label: "Catatan penutup invoice", maxLength: 300 },
  {
    key: "invoice_due_days",
    label: "Jatuh tempo invoice (hari)",
    hint: "Dihitung sejak invoice terbit. Kosong atau tidak wajar dianggap 3 hari.",
    type: "number" as const,
    min: 0,
    max: 365,
    step: 1,
  },
] as const;

export const TEMPLATE_FIELDS = [
  {
    key: "wa_template_dp",
    label: "Pesan permintaan DP",
    placeholders: "{{customer_name}} {{trip_title}} {{total}} {{dp_amount}} {{bank_account}}",
  },
  {
    key: "wa_template_invoice",
    label: "Pesan penagihan pelunasan",
    placeholders:
      "{{customer_name}} {{invoice_number}} {{total}} {{amount_paid}} {{amount_due}} {{bank_account}} {{due_date}}",
  },
  {
    key: "wa_template_shipped",
    label: "Pesan informasi pengiriman",
    placeholders: "{{customer_name}} {{order_number}} {{courier}} {{service}} {{tracking_number}}",
  },
] as const;

/*
 * Pengaturan pengiriman tidak punya form berbasis daftar seperti dua di atas —
 * isinya dirakit sendiri oleh tab Ongkir dari pencarian kota dan centang kurir.
 * Namanya tetap perlu ada di sini supaya Jejak Perubahan tidak menyebut nama
 * kolom database mentah kepada orang yang membacanya.
 */
const LABEL_ONGKIR: Record<string, string> = {
  shipping_origin_id: "Kota asal pengiriman",
  shipping_origin_label: "Nama kota asal",
  shipping_couriers: "Kurir yang dipakai",
};

/** Nama kolom pengaturan yang dibaca orang, dikunci nama kuncinya. */
export const LABEL_PENGATURAN: Record<string, string> = {
  ...Object.fromEntries(
    [...STORE_FIELDS, ...TEMPLATE_FIELDS].map((field) => [field.key, field.label]),
  ),
  ...LABEL_ONGKIR,
};
