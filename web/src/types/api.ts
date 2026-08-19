/**
 * Tipe data yang dikembalikan backend.
 *
 * Catatan penting: seluruh nominal uang bertipe `string`, bukan `number`.
 * Backend menyimpannya sebagai NUMERIC PostgreSQL yang presisinya melebihi
 * `number` JavaScript, jadi nilainya baru dikonversi saat ditampilkan lewat
 * helper `formatIDR`/`toNumber`.
 */

export type Money = string;

export interface Envelope<T> {
  data: T;
  meta?: PageMeta;
  error?: ApiErrorPayload;
}

export interface PageMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface ApiErrorPayload {
  code: string;
  message: string;
  fields?: Record<string, string>;
}

// --- Pengguna & sesi -------------------------------------------------------

export type UserRole = "owner" | "admin" | "tripper";

/** Hak akses per menu; nilainya sama persis dengan yang dipakai backend. */
export type Permission =
  | "trips"
  | "shopping_list"
  | "purchases"
  | "orders"
  | "invoices"
  /** Mencakup antrean kemas sekaligus daftar paket — keduanya satu menu. */
  | "shipments"
  | "customers"
  | "products"
  | "stock"
  | "reports"
  | "settings"
  | "users";

export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  phone: string | null;
  is_active: boolean;
  /** Hak akses khusus pengguna ini. Kosong berarti mengikuti bawaan role. */
  permissions: Permission[];
  /** Hasil gabungan hak khusus dengan bawaan role, dihitung backend. */
  effective_permissions: Permission[];
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Session {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  user: User;
}

// --- Customer --------------------------------------------------------------

export interface Customer {
  id: string;
  code: string;
  name: string;
  phone_wa: string;
  email: string | null;
  instagram: string | null;
  address: string | null;
  city: string | null;
  /** Kecamatan. */
  district: string | null;
  /** Kelurahan atau desa. */
  subdistrict: string | null;
  province: string | null;
  postal_code: string | null;
  notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface CustomerStats {
  total_orders: number;
  total_spent: Money;
  last_order_at: string | null;
}

// --- Produk ----------------------------------------------------------------

export type MarkupType = "percent" | "nominal";

export interface ProductCategory {
  id: string;
  name: string;
  slug: string;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: string;
  sku: string;
  name: string;
  category_id: string | null;
  category_name?: string | null;
  brand: string | null;
  store_name: string | null;
  base_currency: string;
  base_price: Money;
  markup_type: MarkupType;
  markup_value: Money;
  weight_gram: number;
  image_url: string | null;
  notes: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/** Riwayat harga sebuah produk pada satu trip. */
export interface ProductPriceHistory {
  trip_id: string;
  trip_code: string;
  trip_title: string;
  country: string;
  depart_date: string;
  currency: string;
  exchange_rate: string;
  catalog_cost: Money;
  catalog_cost_idr: Money;
  sell_price: Money;
  actual_cost: Money;
  actual_cost_idr: Money;
  qty_purchased: number;
  qty_sold: number;
}

export interface PricePreview {
  cost_price_idr: Money;
  sell_price: Money;
  profit_per_pcs: Money;
}

// --- Trip ------------------------------------------------------------------

/** Status trip cukup dua: masih menerima order atau tidak. */
export type TripStatus = "open" | "closed";

export interface Trip {
  id: string;
  code: string;
  title: string;
  country: string;
  city: string | null;
  tripper_user_id: string | null;
  tripper_name?: string | null;
  depart_date: string;
  return_date: string;
  order_deadline: string | null;
  currency: string;
  exchange_rate: string;
  status: TripStatus;
  notes: string | null;
  created_by: string | null;
  created_at: string;
  updated_at: string;
  total_orders?: number;
  total_customers?: number;
  catalog_items?: number;
}

export interface TripItem {
  id: string;
  trip_id: string;
  product_id: string;
  cost_price: Money;
  cost_price_idr: Money;
  markup_type: MarkupType;
  markup_value: Money;
  sell_price: Money;
  max_qty: number | null;
  is_active: boolean;
  notes: string | null;
  product_name: string;
  product_sku: string;
  brand: string | null;
  image_url: string | null;
  weight_gram: number;
  category_name: string | null;
  qty_ordered: number;
  created_at: string;
  updated_at: string;
}

export type ExpenseCategory =
  | "tiket"
  | "bagasi"
  | "akomodasi"
  | "transport"
  | "visa"
  | "lainnya";

export interface TripExpense {
  id: string;
  trip_id: string;
  category: ExpenseCategory;
  description: string;
  amount: Money;
  spent_at: string;
  receipt_url: string | null;
  created_at: string;
  updated_at: string;
}

// --- Order -----------------------------------------------------------------

/** Asal order, dipakai untuk rekap penjualan per channel. */
export type OrderSource = "whatsapp" | "instagram" | "tiktok" | "marketplace" | "lainnya";

/**
 * Lima tahap perjalanan order ditambah Batal.
 *
 * Mengemas, menetapkan ongkir, dan menerbitkan invoice pelunasan semuanya
 * terjadi di dalam "dp_paid" — kemajuannya dibaca dari data kemasan dan baris
 * invoice, bukan dari status.
 */
export type OrderStatus =
  | "awaiting_dp"
  | "dp_paid"
  | "paid"
  | "shipped"
  | "completed"
  | "cancelled";

export type FulfillmentStatus =
  | "pending"
  | "purchased"
  | "partial"
  | "unavailable"
  | "refunded";

export interface OrderItem {
  id: string;
  order_id: string;
  product_id: string;
  trip_item_id: string | null;
  product_name: string;
  product_sku: string;
  qty: number;
  unit_price: Money;
  unit_cost_est: Money;
  subtotal: Money;
  qty_purchased: number;
  qty_received: number;
  fulfillment_status: FulfillmentStatus;
  notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: string;
  order_number: string;
  trip_id: string;
  customer_id: string;
  order_date: string;
  status: OrderStatus;
  order_source: OrderSource;
  subtotal: Money;
  discount: Money;
  shipping_fee: Money;
  total: Money;
  dp_required: Money;
  paid_amount: Money;
  balance_due: Money;
  recipient_name: string;
  recipient_phone: string;
  shipping_address: string;
  shipping_city: string;
  shipping_district: string | null;
  shipping_subdistrict: string | null;
  shipping_province: string | null;
  shipping_postal_code: string | null;
  notes: string | null;
  cancel_reason: string | null;
  cancelled_at: string | null;
  created_at: string;
  updated_at: string;
  // Hanya ada pada endpoint daftar order.
  customer_name?: string;
  customer_code?: string;
  trip_code?: string;
  trip_title?: string;
  trip_currency?: string;
  trip_exchange_rate?: string;
  item_count?: number;
  total_qty?: number;
}

export type PaymentType = "dp" | "settlement" | "refund" | "adjustment";
export type PaymentMethod = "transfer" | "cash" | "qris" | "ewallet" | "lainnya";

export interface Payment {
  id: string;
  order_id: string;
  type: PaymentType;
  amount: Money;
  method: PaymentMethod;
  reference: string | null;
  proof_url: string | null;
  paid_at: string;
  notes: string | null;
  created_at: string;
}

export interface OrderDetail extends Order {
  customer: Customer;
  trip: Trip;
  items: OrderItem[];
  payments: Payment[];
  invoices: Invoice[];
  shipment: Shipment | null;
  next_statuses: OrderStatus[];
  editable: boolean;
}

// --- Belanja & stok --------------------------------------------------------

export interface ShoppingListEntry {
  product_id: string;
  product_name: string;
  product_sku: string;
  brand: string | null;
  store_name: string | null;
  image_url: string | null;
  category_name: string | null;
  /** Hanya order yang DP-nya sudah masuk; itulah yang boleh dibelanjakan. */
  qty_ordered: number;
  /** Permintaan yang masih menunggu DP, dipisah agar tidak ikut terbeli. */
  qty_awaiting_dp: number;
  qty_purchased: number;
  qty_remaining: number;
  order_count: number;
  est_cost_idr: Money;
  cost_price: Money;
  sell_price_idr: Money;
}

export interface Purchase {
  id: string;
  trip_id: string;
  product_id: string;
  purchase_date: string;
  qty: number;
  unit_cost_foreign: Money;
  currency: string;
  exchange_rate: string;
  unit_cost_idr: Money;
  total_cost_idr: Money;
  store_name: string | null;
  receipt_url: string | null;
  notes: string | null;
  created_at: string;
  updated_at: string;
  product_name?: string;
  product_sku?: string;
  purchaser_name?: string | null;
  qty_to_orders?: number;
  qty_to_stock?: number;
}

export interface PurchaseAllocation {
  id: string;
  purchase_id: string;
  order_item_id: string | null;
  qty: number;
  unit_cost_idr: Money;
  order_number: string | null;
  customer_name: string | null;
  product_name: string;
  created_at: string;
}

export interface PurchaseResult {
  purchase: Purchase;
  allocations: PurchaseAllocation[];
  qty_to_orders: number;
  qty_to_stock: number;
}

export interface StockItem {
  id: string;
  product_id: string;
  qty_on_hand: number;
  avg_cost_idr: Money;
  location: string | null;
  updated_at: string;
  product_name: string;
  product_sku: string;
  brand: string | null;
  image_url: string | null;
  category_name: string | null;
  stock_value: Money;
}

export type StockMovementType =
  | "in_purchase"
  | "out_order"
  | "out_marketplace"
  | "adjustment";

export interface StockMovement {
  id: string;
  product_id: string;
  type: StockMovementType;
  qty: number;
  unit_cost_idr: Money;
  sale_price_idr: Money | null;
  trip_id: string | null;
  ref_type: string | null;
  ref_id: string | null;
  note: string | null;
  created_at: string;
  product_name: string;
  product_sku: string;
  actor_name: string | null;
}

// --- Invoice & pengiriman --------------------------------------------------

export type InvoiceType = "dp" | "final";
export type InvoiceStatus = "draft" | "sent" | "paid" | "void";
export type SentChannel = "wa" | "email" | "manual";

export interface Invoice {
  id: string;
  invoice_number: string;
  order_id: string;
  type: InvoiceType;
  issue_date: string;
  due_date: string | null;
  subtotal: Money;
  discount: Money;
  shipping_fee: Money;
  total: Money;
  /** Uang muka: ditagih pada invoice DP, dikurangkan pada invoice pelunasan. */
  dp_amount: Money;
  amount_paid: Money;
  amount_due: Money;
  status: InvoiceStatus;
  pdf_path: string | null;
  sent_channel: SentChannel | null;
  sent_at: string | null;
  paid_at: string | null;
  notes: string | null;
  created_at: string;
  updated_at: string;
  order_number?: string;
  customer_name?: string;
  customer_phone?: string;
  trip_code?: string;
}

export type ShipmentStatus =
  | "packing"
  | "ready"
  | "shipped"
  | "delivered"
  | "returned";

export interface Shipment {
  id: string;
  order_id: string;
  courier: string;
  service: string;
  tracking_number: string | null;
  weight_gram: number;
  length_cm: number;
  width_cm: number;
  height_cm: number;
  estimated_cost: Money;
  shipping_cost: Money;
  status: ShipmentStatus;
  packed_at: string | null;
  shipped_at: string | null;
  delivered_at: string | null;
  customer_notified_at: string | null;
  notes: string | null;
  created_at: string;
  updated_at: string;
  order_number?: string;
  customer_name?: string;
  recipient_name?: string;
  recipient_phone?: string;
  shipping_city?: string;
  order_status?: OrderStatus;
  order_balance_due?: Money;
}

/** Pesan siap kirim beserta tautan pembukanya. */
export interface NotifyMessage {
  phone: string;
  text: string;
  whatsapp_url: string;
  mailto_url?: string;
}

// --- Ongkir ----------------------------------------------------------------

/** Sisa barang surplus sebuah trip yang masih tersimpan di gudang. */
export interface TripStockOnHand {
  product_name: string;
  sku: string;
  qty: number;
}

/**
 * Apa yang ikut terhapus bersama sebuah trip, dan apa yang menghalanginya.
 *
 * Diambil sebelum tombol hapus ditekan supaya angkanya terlihat lebih dulu —
 * menghapus trip membuang order, invoice yang sudah terkirim, dan catatan uang
 * yang sudah diterima.
 */
export interface TripDeletionImpact {
  trip_id: string;
  trip_code: string;
  orders: number;
  invoices: number;
  /** Uang yang sudah benar-benar diterima pada trip ini. */
  payments_total: Money;
  purchases: number;
  purchases_cost: Money;
  expenses: number;
  catalog_items: number;
  shipments: number;
  /** Order yang sudah diserahkan ke kurir — menghalangi penghapusan. */
  shipped_orders: string[] | null;
  /** Surplus yang masih ada di stok — menghalangi penghapusan. */
  stock_on_hand: TripStockOnHand[];
}

/** Order yang siap ditagih pelunasannya, untuk dialog Buat Invoice. */
export interface InvoiceCandidate {
  order_id: string;
  order_number: string;
  order_date: string;
  customer_name: string;
  trip_code: string;
  total: Money;
  shipping_fee: Money;
  paid_amount: Money;
  balance_due: Money;
}

/**
 * Satu baris di menu Pengiriman.
 *
 * Yang didaftar adalah order, bukan paket — paket baru terbentuk setelah data
 * kemasan diisi, sementara pekerjaannya justru dimulai sebelum itu. Kolom paket
 * karenanya boleh null.
 */
export interface ShippingQueueItem {
  order_id: string;
  order_number: string;
  order_status: OrderStatus;
  order_date: string;
  trip_code: string;
  customer_name: string;
  recipient_name: string;
  recipient_phone: string;
  shipping_city: string;
  total_qty: number;
  total: Money;
  balance_due: Money;
  /** Ongkir yang ditagihkan ke customer. Nol berarti layanan belum dipilih. */
  shipping_fee: Money;

  shipment_id: string | null;
  courier: string | null;
  service: string | null;
  weight_gram: number | null;
  length_cm: number | null;
  width_cm: number | null;
  height_cm: number | null;
  tracking_number: string | null;
  shipment_status: ShipmentStatus | null;
  shipment_notes: string | null;
  packed_at: string | null;
  shipped_at: string | null;
  /** Ongkos yang benar-benar dibayar ke kurir, diisi saat resi dicatat. */
  shipping_cost: Money | null;
  customer_notified_at: string | null;
}

/** Tahap pekerjaan di menu Pengiriman, dipakai menyaring daftar. */
export type ShippingStage = "perlu_kemas" | "siap_kirim" | "terkirim";

/** Satu layanan kurir yang bisa dipilih saat mengemas. */
export interface ShippingOption {
  courier: string;
  service: string;
  cost: Money;
  etd: string;
  destination?: string;
  source: string;
}

/** Hasil hitung ongkir beserta dasar perhitungannya. */
export interface ShippingEstimate {
  courier: string;
  service: string;
  city: string;
  actual_weight_gram: number;
  volumetric_weight_gram: number;
  chargeable_weight_gram: number;
  price_per_kg: Money;
  cost: Money;
  etd: string;
  /** Tujuan seperti dikenali kurir, misalnya "CILANDAK BARAT, JAKARTA SELATAN, 12430". */
  destination?: string;
  /** Nama layanan kurir yang menjawab. */
  source: string;
}

/** Satu tujuan pengiriman di daftar resmi kurir. */
export interface ShippingDestination {
  destination_id: number;
  label: string;
  city_name: string | null;
  province_name: string | null;
  zip_code: string | null;
}

export interface CourierOption {
  code: string;
  name: string;
}

/** Keadaan layanan tarif yang sedang dipakai backend. */
export interface ShippingProviderInfo {
  name: string;
  /** API key terpasang dan layanannya aktif. */
  connected: boolean;
  /** Terhubung sekaligus kota asalnya sudah dipilih. */
  ready: boolean;
  origin_id: number;
  origin_label: string;
  couriers: string[];
  courier_options: CourierOption[];
}

// --- Laporan ---------------------------------------------------------------

export interface ExpenseBreakdown {
  category: ExpenseCategory;
  amount: Money;
}

export interface TripProfitReport {
  trip_id: string;
  trip_code: string;
  title: string;
  country: string;
  status: TripStatus;
  revenue: Money;
  cogs: Money;
  gross_profit: Money;
  trip_expenses: Money;
  net_profit: Money;
  margin_percent: string;
  shipping_fee_collected: Money;
  shipping_cost_paid: Money;
  discount_given: Money;
  surplus_stock_qty: number;
  surplus_stock_value: Money;
  total_capital_out: Money;
  payment_received: Money;
  outstanding: Money;
  order_count: number;
  customer_count: number;
  item_qty: number;
  expense_breakdown: ExpenseBreakdown[];
}

export interface OrderProfit {
  order_id: string;
  order_number: string;
  customer_name: string;
  trip_code: string;
  status: OrderStatus;
  order_date: string;
  revenue: Money;
  cogs: Money;
  profit: Money;
  margin_percent: string;
}

export interface Receivable {
  order_id: string;
  order_number: string;
  customer_id: string;
  customer_name: string;
  customer_phone: string;
  trip_code: string;
  status: OrderStatus;
  order_date: string;
  total: Money;
  paid_amount: Money;
  balance_due: Money;
  days_outstanding: number;
}

export interface CustomerSales {
  customer_id: string;
  customer_code: string;
  customer_name: string;
  customer_phone: string;
  city: string | null;
  order_count: number;
  item_qty: number;
  revenue: Money;
  cogs: Money;
  profit: Money;
  outstanding: Money;
  avg_order_value: Money;
  first_order_at: string | null;
  last_order_at: string | null;
}

export interface ChannelSales {
  order_source: OrderSource;
  order_count: number;
  customer_count: number;
  item_qty: number;
  revenue: Money;
  cogs: Money;
  profit: Money;
  avg_order_value: Money;
  revenue_share: string;
}

export interface ProductSales {
  product_id: string;
  product_name: string;
  product_sku: string;
  category_name: string | null;
  qty_sold: number;
  order_count: number;
  revenue: Money;
  cogs: Money;
  profit: Money;
}

export interface DashboardSummary {
  active_trips: number;
  open_orders: number;
  pending_shipment: number;
  outstanding: Money;
  revenue_this_month: Money;
  profit_this_month: Money;
  orders_this_month: number;
  stock_value: Money;
  stock_qty: number;
  customer_count: number;
  recent_orders: Order[];
  upcoming_trips: Trip[];
  top_products: ProductSales[];
}

// --- Pengaturan & audit ----------------------------------------------------

export interface Setting {
  key: string;
  value: string;
  description: string | null;
  updated_at: string;
}

export interface AuditLog {
  id: number;
  user_id: string | null;
  user_name: string | null;
  entity: string;
  entity_id: string | null;
  action: string;
  changes: Record<string, unknown> | null;
  ip_address: string | null;
  created_at: string;
}

export interface UploadResult {
  url: string;
  path: string;
  content_type: string;
  size: number;
  original_name: string;
}
