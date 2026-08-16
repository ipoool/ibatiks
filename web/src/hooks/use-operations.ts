"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { ListParams } from "@/hooks/use-master";
import { api, buildQuery } from "@/lib/api";
import type {
  Invoice,
  NotifyMessage,
  Purchase,
  PurchaseAllocation,
  PurchaseResult,
  SentChannel,
  Shipment,
  ShipmentStatus,
  ShippingEstimate,
  ShippingRate,
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
}

export const invoiceKeys = {
  all: ["invoices"] as const,
  list: (params: InvoiceListParams) => ["invoices", "list", params] as const,
  message: (id: string) => ["invoices", "message", id] as const,
};

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

// --- Pengiriman ------------------------------------------------------------

export interface ShipmentListParams extends ListParams {
  status?: ShipmentStatus | "";
  trip_id?: string;
}

export const shipmentKeys = {
  all: ["shipments"] as const,
  list: (params: ShipmentListParams) => ["shipments", "list", params] as const,
  message: (orderId: string) => ["shipments", "message", orderId] as const,
};

export function useShipments(params: ShipmentListParams) {
  return useQuery({
    queryKey: shipmentKeys.list(params),
    queryFn: () => api.list<Shipment>(`/shipments${buildQuery({ ...params })}`),
  });
}

function invalidateShipmentRelated(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: shipmentKeys.all });
  queryClient.invalidateQueries({ queryKey: ["orders"] });
  queryClient.invalidateQueries({ queryKey: ["reports"] });
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

export const shippingRateKeys = {
  all: ["shipping-rates"] as const,
  list: (params: { courier?: string; q?: string }) => ["shipping-rates", params] as const,
};

export function useShippingRates(params: { courier?: string; q?: string } = {}) {
  return useQuery({
    queryKey: shippingRateKeys.list(params),
    queryFn: () => api.get<ShippingRate[]>(`/shipping/rates${buildQuery({ ...params })}`),
    staleTime: 5 * 60_000,
  });
}

export function useSaveShippingRate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: {
      courier?: string;
      service?: string;
      destination_city: string;
      province?: string | null;
      price_per_kg: string;
      min_weight_gram?: number;
      etd?: string | null;
    }) => api.post<ShippingRate>("/shipping/rates", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: shippingRateKeys.all });
    },
  });
}

export function useDeleteShippingRate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.delete(`/shipping/rates/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: shippingRateKeys.all });
    },
  });
}
