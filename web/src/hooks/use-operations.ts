"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { ListParams } from "@/hooks/use-master";
import { api, buildQuery } from "@/lib/api";
import type {
  Invoice,
  InvoiceCandidate,
  PaymentMethod,
  NotifyMessage,
  Purchase,
  PurchaseAllocation,
  PurchaseResult,
  SentChannel,
  Shipment,
  ShippingDestination,
  ShippingEstimate,
  ShippingOption,
  TrackingInfo,
  ShippingProviderInfo,
  ShippingQueueItem,
  ShippingStage,
  StockItem,
  StockMovement,
} from "@/types/api";

// --- Pembelian -------------------------------------------------------------

export interface PurchaseListParams extends ListParams {
  trip_id?: string;
  product_id?: string;
}

export const purchaseKeys = {
  all: ["purchases"] as const,
  list: (params: PurchaseListParams) => ["purchases", "list", params] as const,
  allocations: (id: string) => ["purchases", "allocations", id] as const,
};

export function usePurchases(params: PurchaseListParams) {
  return useQuery({
    queryKey: purchaseKeys.list(params),
    queryFn: () => api.list<Purchase>(`/purchases${buildQuery({ ...params })}`),
  });
}

export function usePurchaseAllocations(purchaseId: string | undefined) {
  return useQuery({
    queryKey: purchaseKeys.allocations(purchaseId ?? ""),
    queryFn: () => api.get<PurchaseAllocation[]>(`/purchases/${purchaseId}/allocations`),
    enabled: Boolean(purchaseId),
  });
}

export type PurchasePayload = {
  product_id: string;
  purchase_date?: string;
  qty: number;
  unit_cost_foreign: string;
  exchange_rate?: string;
  store_name?: string | null;
  receipt_url?: string | null;
  notes?: string | null;
};

/** Mencatat belanja tripper; backend langsung mengalokasikannya ke pesanan dan stok. */
export function useRecordPurchase(tripId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: PurchasePayload) =>
      api.post<PurchaseResult>(`/trips/${tripId}/purchases`, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: purchaseKeys.all });
      queryClient.invalidateQueries({ queryKey: ["trips"] });
      queryClient.invalidateQueries({ queryKey: ["orders"] });
      queryClient.invalidateQueries({ queryKey: ["stock"] });
      queryClient.invalidateQueries({ queryKey: ["reports"] });
    },
  });
}

export function useDeletePurchase() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.delete(`/purchases/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: purchaseKeys.all });
      queryClient.invalidateQueries({ queryKey: ["trips"] });
      queryClient.invalidateQueries({ queryKey: ["orders"] });
      queryClient.invalidateQueries({ queryKey: ["stock"] });
      queryClient.invalidateQueries({ queryKey: ["reports"] });
    },
  });
}

// --- Stok ------------------------------------------------------------------

export interface StockListParams extends ListParams {
  in_stock_only?: boolean;
}

export const stockKeys = {
  all: ["stock"] as const,
  list: (params: StockListParams) => ["stock", "list", params] as const,
  movements: (params: ListParams & { product_id?: string }) =>
    ["stock", "movements", params] as const,
};

export function useStock(params: StockListParams) {
  return useQuery({
    queryKey: stockKeys.list(params),
    queryFn: () => api.list<StockItem>(`/stock${buildQuery({ ...params })}`),
  });
}

export function useStockMovements(params: ListParams & { product_id?: string }) {
  return useQuery({
    queryKey: stockKeys.movements(params),
    queryFn: () => api.list<StockMovement>(`/stock/movements${buildQuery({ ...params })}`),
  });
}

export function useSellStock() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: {
      product_id: string;
      qty: number;
      sale_price: string;
      channel?: string;
      note?: string | null;
    }) => api.post<StockMovement>("/stock/sell", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stockKeys.all });
    },
  });
}

export function useAdjustStock() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: { product_id: string; new_qty: number; note?: string | null }) =>
      api.post<StockItem>("/stock/adjust", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stockKeys.all });
    },
  });
}

// --- Invoice ---------------------------------------------------------------

export interface InvoiceListParams extends ListParams {
  status?: string;
  type?: string;
  from?: string;
  to?: string;
}

export const invoiceKeys = {
  candidates: (search: string) => ["invoices", "candidates", search] as const,
  all: ["invoices"] as const,
  list: (params: InvoiceListParams) => ["invoices", "list", params] as const,
  message: (id: string) => ["invoices", "message", id] as const,
};

/**
 * Order yang siap ditagih pelunasannya: DP-nya sudah masuk, ongkirnya sudah
 * ditetapkan, dan belum punya invoice pelunasan yang berlaku.
 */
export function useInvoiceCandidates(search: string, enabled: boolean) {
  return useQuery({
    queryKey: invoiceKeys.candidates(search),
    queryFn: () => api.get<InvoiceCandidate[]>(`/invoices/candidates${buildQuery({ q: search })}`),
    enabled,
  });
}

/** Menerbitkan invoice pelunasan untuk sebuah order dari menu Invoice. */
export function useIssueFinalInvoice() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      orderId,
      ...payload
    }: {
      orderId: string;
      due_date?: string;
      notes?: string | null;
    }) => api.post<Invoice>(`/orders/${orderId}/invoices`, { type: "final", ...payload }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: invoiceKeys.all });
      queryClient.invalidateQueries({ queryKey: ["orders"] });
      queryClient.invalidateQueries({ queryKey: shipmentKeys.all });
    },
  });
}

/**
 * Mencatat pelunasan sebuah invoice sebagai pembayaran pada ordernya.
 *
 * Sengaja tidak sekadar menandai baris invoice jadi "paid": saldo order dan
 * laporan piutang dihitung dari tabel pembayaran, jadi menandai dokumennya
 * tanpa mencatat uangnya akan membuat keduanya berbohong.
 */
export function useSettleInvoice() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      orderId,
      ...payload
    }: {
      orderId: string;
      amount: string;
      method: PaymentMethod;
      paid_at?: string;
      reference?: string | null;
    }) => api.post(`/orders/${orderId}/payments`, { type: "settlement", ...payload }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: invoiceKeys.all });
      queryClient.invalidateQueries({ queryKey: ["orders"] });
      queryClient.invalidateQueries({ queryKey: shipmentKeys.all });
      queryClient.invalidateQueries({ queryKey: ["reports"] });
    },
  });
}

export function useInvoices(params: InvoiceListParams) {
  return useQuery({
    queryKey: invoiceKeys.list(params),
    queryFn: () => api.list<Invoice>(`/invoices${buildQuery({ ...params })}`),
  });
}

/** Teks penagihan siap kirim; hanya diambil saat dialog kirim dibuka. */
export function useInvoiceMessage(invoiceId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: invoiceKeys.message(invoiceId ?? ""),
    queryFn: () => api.get<NotifyMessage>(`/invoices/${invoiceId}/message`),
    enabled: Boolean(invoiceId) && enabled,
  });
}

export function useMarkInvoiceSent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, channel }: { id: string; channel: SentChannel }) =>
      api.post<Invoice>(`/invoices/${id}/mark-sent`, { channel }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: invoiceKeys.all });
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
  });
}

export function useVoidInvoice() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.post<Invoice>(`/invoices/${id}/void`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: invoiceKeys.all });
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
  });
}

export const invoicePDFUrl = (id: string) => api.downloadURL(`/invoices/${id}/pdf`);

/** Label pengirim–penerima siap tempel di kardus. */
export const labelUrl = (orderID: string) => api.downloadURL(`/orders/${orderID}/label`);

// --- Pengiriman ------------------------------------------------------------

export interface ShipmentListParams extends ListParams {
  /** Tahap pekerjaan, bukan status tersimpan. Kosong berarti semuanya. */
  stage?: ShippingStage | "";
  trip_id?: string;
}

export const shipmentKeys = {
  all: ["shipments"] as const,
  list: (params: ShipmentListParams) => ["shipments", "list", params] as const,
  message: (orderId: string) => ["shipments", "message", orderId] as const,
  options: (orderId: string) => ["shipments", "options", orderId] as const,
};

/** Antrean pengiriman: order yang DP-nya sudah masuk beserta data kemasannya. */
export function useShippingQueue(params: ShipmentListParams) {
  return useQuery({
    queryKey: shipmentKeys.list(params),
    queryFn: () => api.list<ShippingQueueItem>(`/shipments${buildQuery({ ...params })}`),
  });
}

/**
 * Daftar layanan kurir beserta harganya untuk satu paket.
 *
 * Sengaja mutation, bukan query: dijalankan saat admin menekan tombolnya
 * setelah mengisi berat dan dimensi, bukan tiap kali dialognya terbuka —
 * tiap panggilan memakan kuota langganan RajaOngkir.
 */
export function useShippingOptions(orderId: string) {
  return useMutation({
    mutationFn: (payload: {
      weight_gram: number;
      length_cm?: number;
      width_cm?: number;
      height_cm?: number;
    }) => api.post<ShippingOption[]>(`/orders/${orderId}/shipping-options`, payload),
  });
}

function invalidateShipmentRelated(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: shipmentKeys.all });
  queryClient.invalidateQueries({ queryKey: ["orders"] });
  queryClient.invalidateQueries({ queryKey: ["reports"] });
}

/**
 * Melacak posisi paket lewat resi yang tersimpan.
 *
 * Dibuat sebagai mutation, bukan query: pengecekannya memakai kuota RajaOngkir,
 * jadi hanya berjalan saat admin memintanya — bukan tiap kali daftarnya dibuka.
 */
export function useTrackOrder(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => api.get<TrackingInfo>(`/orders/${orderId}/tracking`),
    onSuccess: (info) => {
      // Statusnya bisa berubah jadi Selesai di server, jadi daftarnya disegarkan.
      if (info.order_completed) invalidateShipmentRelated(queryClient);
    },
  });
}

export function usePackOrder(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: {
      courier?: string;
      service?: string;
      weight_gram?: number;
      length_cm?: number;
      width_cm?: number;
      height_cm?: number;
      /** Ongkir yang ditagihkan; dikosongkan berarti tidak ikut diubah. */
      shipping_fee?: string;
      insurance_fee?: string;
      notes?: string | null;
    }) => api.post<Shipment>(`/orders/${orderId}/pack`, payload),
    onSuccess: () => invalidateShipmentRelated(queryClient),
  });
}

/** Mencatat nomor resi JNE dan menandai paket sudah diserahkan ke kurir. */
export function useShipOrder(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: {
      tracking_number: string;
      shipping_cost?: string;
      shipped_at?: string;
      allow_unpaid?: boolean;
    }) => api.post<Shipment>(`/orders/${orderId}/ship`, payload),
    onSuccess: () => invalidateShipmentRelated(queryClient),
  });
}

export function useMarkDelivered(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => api.post<Shipment>(`/orders/${orderId}/delivered`),
    onSuccess: () => invalidateShipmentRelated(queryClient),
  });
}

export function useShipmentMessage(orderId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: shipmentKeys.message(orderId ?? ""),
    queryFn: () => api.get<NotifyMessage>(`/orders/${orderId}/shipment-message`),
    enabled: Boolean(orderId) && enabled,
  });
}

export function useMarkShipmentNotified(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => api.post<Shipment>(`/orders/${orderId}/shipment-notified`),
    onSuccess: () => invalidateShipmentRelated(queryClient),
  });
}

export function useUpdateShipment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      ...payload
    }: {
      id: string;
      courier?: string;
      service?: string;
      weight_gram?: number;
      shipping_cost?: string;
      tracking_number?: string | null;
      notes?: string | null;
    }) => api.put<Shipment>(`/shipments/${id}`, payload),
    onSuccess: () => invalidateShipmentRelated(queryClient),
  });
}

// --- Ongkir ----------------------------------------------------------------

export type EstimatePayload = {
  courier?: string;
  service?: string;
  city?: string;
  weight_gram?: number;
  length_cm?: number;
  width_cm?: number;
  height_cm?: number;
};

/**
 * Menghitung perkiraan ongkir untuk sebuah order. Kota tujuan diambil dari
 * alamat kirim order, jadi tidak perlu diketik ulang.
 */
export function useEstimateShipping(orderId: string) {
  return useMutation({
    mutationFn: (payload: EstimatePayload) =>
      api.post<ShippingEstimate>(`/orders/${orderId}/shipping-estimate`, payload),
  });
}

/**
 * Uji coba hitung ongkir tanpa order, dipakai panel percobaan di Pengaturan
 * supaya tim toko bisa memastikan sambungannya benar sebelum ada order jalan.
 */
export function useTestShippingEstimate() {
  return useMutation({
    mutationFn: (payload: {
      courier?: string;
      service?: string;
      city: string;
      district?: string;
      subdistrict?: string;
      postal_code?: string;
      weight_gram: number;
    }) => api.post<ShippingEstimate>("/shipping/estimate", payload),
  });
}

export const shippingProviderKeys = {
  provider: ["shipping-provider"] as const,
  destinations: (q: string) => ["shipping-destinations", q] as const,
};

/** Keadaan sambungan ke RajaOngkir, satu-satunya sumber ongkir. */
export function useShippingProvider() {
  return useQuery({
    queryKey: shippingProviderKeys.provider,
    queryFn: () => api.get<ShippingProviderInfo>("/shipping/provider"),
  });
}

/**
 * Pencarian kota asal di daftar tujuan kurir.
 *
 * Baru jalan pada tiga huruf: tiap pencarian memakan kuota langganan, jadi
 * mengetik satu huruf tidak boleh langsung menembak API.
 */
export function useShippingDestinations(q: string, enabled = true) {
  const keyword = q.trim();
  return useQuery({
    queryKey: shippingProviderKeys.destinations(keyword),
    queryFn: () =>
      api.get<ShippingDestination[]>(`/shipping/destinations${buildQuery({ q: keyword })}`),
    enabled: enabled && keyword.length >= 3,
    staleTime: 5 * 60 * 1000,
  });
}

