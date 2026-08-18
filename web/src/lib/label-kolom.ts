/**
 * Nama kolom versi manusia untuk pesan galat dari API.
 *
 * Backend menyebut kolomnya dengan nama teknis (`tracking_number`,
 * `phone_wa`), dan itu yang ikut terkirim di daftar `fields` pada galat
 * validasi. Menampilkannya apa adanya berarti tim toko membaca nama kolom
 * database di tengah formulir yang seluruh labelnya berbahasa Indonesia.
 *
 * Kunci yang belum terdaftar sengaja tidak ditebak-tebak: pemanggilnya
 * menampilkan pesannya saja, tanpa nama kolom. Lebih baik kehilangan satu
 * petunjuk daripada memunculkan istilah yang tidak dikenali siapa pun.
 */
const LABEL_KOLOM: Record<string, string> = {
  // Identitas dan kontak
  name: "Nama",
  email: "Email",
  phone: "Nomor telepon",
  phone_wa: "Nomor WhatsApp",
  instagram: "Instagram",
  password: "Password",
  current_password: "Password saat ini",
  new_password: "Password baru",
  role: "Role",
  permissions: "Hak akses menu",

  // Alamat
  address: "Alamat",
  subdistrict: "Kelurahan",
  district: "Kecamatan",
  city: "Kota/Kabupaten",
  province: "Provinsi",
  postal_code: "Kode pos",

  // Produk dan katalog
  sku: "SKU",
  brand: "Brand",
  category: "Kategori",
  category_id: "Kategori",
  store_name: "Toko langganan",
  base_currency: "Mata uang",
  base_price: "Harga modal",
  cost_price: "Harga modal",
  markup_type: "Jenis markup",
  markup_value: "Markup",
  max_qty: "Batas qty",
  weight_gram: "Berat",
  image_url: "URL gambar",
  product_id: "Produk",
  is_active: "Status aktif",

  // Trip
  title: "Judul trip",
  country: "Negara",
  currency: "Mata uang",
  exchange_rate: "Kurs",
  depart_date: "Tanggal berangkat",
  return_date: "Tanggal pulang",
  order_deadline: "Deadline order",
  tripper_user_id: "Tripper",
  trip_id: "Trip",

  // Order
  customer_id: "Customer",
  order_date: "Tanggal order",
  order_source: "Asal order",
  items: "Produk yang dipesan",
  item_id: "Item pesanan",
  qty: "Qty",
  qty_received: "Qty diterima",
  unit_price: "Harga satuan",
  unit_cost_foreign: "Harga modal",
  discount: "Diskon",
  shipping_fee: "Ongkir ditagihkan",
  dp_required: "DP diminta",
  status: "Status",
  reason: "Alasan",
  notes: "Catatan",
  note: "Catatan",

  // Alamat kirim pada order
  recipient_name: "Nama penerima",
  recipient_phone: "Nomor penerima",
  shipping_address: "Alamat kirim",
  shipping_city: "Kota tujuan",
  shipping_district: "Kecamatan",
  shipping_subdistrict: "Kelurahan",
  shipping_province: "Provinsi",
  shipping_postal_code: "Kode pos",

  // Pembayaran dan invoice
  amount: "Nominal",
  method: "Metode pembayaran",
  reference: "Referensi",
  proof_url: "Bukti transfer",
  paid_at: "Tanggal bayar",
  type: "Jenis",
  due_date: "Jatuh tempo",

  // Pengiriman
  courier: "Kurir",
  service: "Layanan",
  tracking_number: "Nomor resi",
  shipping_cost: "Ongkir dibayar",
  shipped_at: "Tanggal kirim",
  length_cm: "Panjang",
  width_cm: "Lebar",
  height_cm: "Tinggi",

  // Belanja, stok, dan tarif
  purchase_date: "Tanggal belanja",
  receipt_url: "Struk belanja",
  new_qty: "Qty baru",
  sale_price: "Harga jual",
  channel: "Channel",
  destination_city: "Kota tujuan",
  price_per_kg: "Tarif per kg",
  min_weight_gram: "Berat minimum",
  etd: "Estimasi tiba",

  // Biaya perjalanan
  description: "Keterangan",
  spent_at: "Tanggal keluar",
};

/** Nama kolom yang bisa dibaca orang, atau null bila belum terdaftar. */
export function labelKolom(kunci: string): string | null {
  return LABEL_KOLOM[kunci] ?? null;
}
