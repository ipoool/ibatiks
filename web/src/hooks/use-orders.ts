"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { ListParams } from "@/hooks/use-master";
import { api, buildQuery } from "@/lib/api";
import type {
  FulfillmentStatus,
  Invoice,
  NotifyMessage,
  Order,
  OrderDetail,
  OrderSource,
  OrderStatus,
  PaymentMethod,
  PaymentType,
} from "@/types/api";

export interface OrderListParams extends ListParams {
  trip_id?: string;
  customer_id?: string;
  status?: OrderStatus | "";
  source?: OrderSource | "";
  unpaid_only?: boolean;
  ready_to_ship?: boolean;
}

export const orderKeys = {
  all: ["orders"] as const,
  list: (params: OrderListParams) => ["orders", "list", params] as const,
  detail: (id: string) => ["orders", "detail", id] as const,
};

/**
 * Hampir semua aksi order mengubah banyak hal sekaligus (total, status,
 * alokasi stok, invoice), jadi setelah mutasi seluruh cache yang berkaitan
 * ditandai basi alih-alih ditambal satu per satu.
 */
function invalidateOrderRelated(queryClient: ReturnType<typeof useQueryClient>, orderId?: string) {
  queryClient.invalidateQueries({ queryKey: orderKeys.all });
  queryClient.invalidateQueries({ queryKey: ["trips"] });
  queryClient.invalidateQueries({ queryKey: ["stock"] });
  queryClient.invalidateQueries({ queryKey: ["reports"] });
  queryClient.invalidateQueries({ queryKey: ["invoices"] });
  queryClient.invalidateQueries({ queryKey: ["shipments"] });
  if (orderId) {
    queryClient.invalidateQueries({ queryKey: orderKeys.detail(orderId) });
  }
}

export function useOrders(params: OrderListParams) {
  return useQuery({
    queryKey: orderKeys.list(params),
    queryFn: () => api.list<Order>(`/orders${buildQuery({ ...params })}`),
  });
}

export function useOrder(id: string | undefined) {
  return useQuery({
    queryKey: orderKeys.detail(id ?? ""),
    queryFn: () => api.get<OrderDetail>(`/orders/${id}`),
    enabled: Boolean(id),
  });
}

export type OrderItemPayload = {
  product_id: string;
  qty: number;
  unit_price?: string;
  notes?: string | null;
};

export type CreateOrderPayload = {
  trip_id: string;
  customer_id: string;
  order_date?: string;
  order_source?: OrderSource;
  items: OrderItemPayload[];
  discount?: string;
  shipping_fee?: string;
  dp_required?: string;
  recipient_name?: string | null;
  recipient_phone?: string | null;
  shipping_address?: string | null;
  shipping_city?: string | null;
  shipping_district?: string | null;
  shipping_subdistrict?: string | null;
  shipping_province?: string | null;
  shipping_postal_code?: string | null;
  notes?: string | null;
};

export function useCreateOrder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateOrderPayload) => api.post<OrderDetail>("/orders", payload),
    onSuccess: () => invalidateOrderRelated(queryClient),
  });
}

export type UpdateOrderPayload = {
  order_date?: string;
  order_source?: OrderSource;
  discount?: string;
  shipping_fee?: string;
  dp_required?: string;
  recipient_name: string;
  recipient_phone: string;
  shipping_address: string;
  shipping_city: string;
  shipping_district?: string | null;
  shipping_subdistrict?: string | null;
  shipping_province?: string | null;
  shipping_postal_code?: string | null;
  notes?: string | null;
};

export function useUpdateOrder(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: UpdateOrderPayload) => api.put<OrderDetail>(`/orders/${id}`, payload),
    onSuccess: () => invalidateOrderRelated(queryClient, id),
  });
}

export function useAddOrderItem(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: OrderItemPayload) =>
      api.post<OrderDetail>(`/orders/${orderId}/items`, payload),
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

/** Mengubah jumlah pesanan — aksi paling sering dipakai pada operasional harian. */
export function useUpdateOrderItem(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      itemId,
      ...payload
    }: {
      itemId: string;
      qty: number;
      unit_price?: string;
      notes?: string | null;
    }) => api.put<OrderDetail>(`/orders/${orderId}/items/${itemId}`, payload),
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

export function useDeleteOrderItem(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (itemId: string) => api.delete(`/orders/${orderId}/items/${itemId}`),
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

export function useChangeOrderStatus(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (status: OrderStatus) =>
      api.patch<OrderDetail>(`/orders/${orderId}/status`, { status }),
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

export function useCancelOrder(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (reason?: string) => api.post<OrderDetail>(`/orders/${orderId}/cancel`, { reason }),
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

export type PaymentPayload = {
  type: PaymentType;
  amount: string;
  method: PaymentMethod;
  reference?: string | null;
  proof_url?: string | null;
  paid_at?: string;
  notes?: string | null;
};

export function useRecordPayment(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: PaymentPayload) =>
      api.post<OrderDetail>(`/orders/${orderId}/payments`, payload),
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

export function useDeletePayment(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (paymentId: string) =>
      api.delete(`/orders/${orderId}/payments/${paymentId}`) as Promise<void>,
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

export type ReceiptPayload = {
  items: Array<{ item_id: string; qty_received: number; status?: FulfillmentStatus }>;
};

/** Mencatat pencocokan barang yang datang dengan isi pesanan. */
export function useReceiveOrder(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: ReceiptPayload) =>
      api.post<OrderDetail>(`/orders/${orderId}/receive`, payload),
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

export function useCreateInvoice(orderId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: { type: "dp" | "final"; due_date?: string; notes?: string | null }) =>
      api.post<Invoice>(`/orders/${orderId}/invoices`, payload),
    onSuccess: () => invalidateOrderRelated(queryClient, orderId),
  });
}

/** Pesan permintaan DP siap kirim beserta tautan wa.me. */
export function useDPMessage(orderId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: ["orders", "dp-message", orderId],
    queryFn: () => api.get<NotifyMessage>(`/orders/${orderId}/dp-message`),
    enabled: Boolean(orderId) && enabled,
  });
}
