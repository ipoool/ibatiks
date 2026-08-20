"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { ListParams } from "@/hooks/use-master";
import { api, buildQuery } from "@/lib/api";
import type {
  AuditLog,
  ChannelSales,
  CustomerSales,
  DashboardSummary,
  OrderProfit,
  ProductSales,
  Receivable,
  Setting,
  TripProfitReport,
  User,
} from "@/types/api";

export const reportKeys = {
  all: ["reports"] as const,
  dashboard: ["reports", "dashboard"] as const,
  tripProfit: (tripId: string) => ["reports", "trip-profit", tripId] as const,
  orderProfits: (params: ListParams & { trip_id?: string }) =>
    ["reports", "order-profits", params] as const,
  receivables: (params: ListParams) => ["reports", "receivables", params] as const,
  products: (params: { trip_id?: string; from?: string; to?: string; limit?: number }) =>
    ["reports", "products", params] as const,
  customers: (params: ListParams & { trip_id?: string }) =>
    ["reports", "customers", params] as const,
  channels: (params: { trip_id?: string; from?: string; to?: string }) =>
    ["reports", "channels", params] as const,
};

export function useDashboard() {
  return useQuery({
    queryKey: reportKeys.dashboard,
    queryFn: () => api.get<DashboardSummary>("/reports/dashboard"),
  });
}

/**
 * Laporan laba-rugi. Tanpa tripId, seluruh trip dijumlahkan jadi satu laporan —
 * itu keadaan yang sah, jadi kuerinya tidak lagi menunggu adanya trip.
 */
export function useTripProfit(params: { trip_id?: string; from?: string; to?: string } = {}) {
  return useQuery({
    queryKey: [...reportKeys.tripProfit(params.trip_id ?? "semua"), params.from ?? "", params.to ?? ""],
    queryFn: () => api.get<TripProfitReport>(`/reports/profit${buildQuery({ ...params })}`),
  });
}

export function useOrderProfits(params: ListParams & { trip_id?: string }) {
  return useQuery({
    queryKey: reportKeys.orderProfits(params),
    queryFn: () => api.list<OrderProfit>(`/reports/orders${buildQuery({ ...params })}`),
  });
}

export function useReceivables(params: ListParams) {
  return useQuery({
    queryKey: reportKeys.receivables(params),
    queryFn: () => api.list<Receivable>(`/reports/receivables${buildQuery({ ...params })}`),
  });
}

/** Rekap penjualan per customer. */
export function useCustomerSales(params: ListParams & { trip_id?: string }) {
  return useQuery({
    queryKey: reportKeys.customers(params),
    queryFn: () => api.list<CustomerSales>(`/reports/customers${buildQuery({ ...params })}`),
  });
}

/** Rekap penjualan per asal order (WhatsApp, Instagram, dan lainnya). */
export function useChannelSales(params: { trip_id?: string; from?: string; to?: string } = {}) {
  return useQuery({
    queryKey: reportKeys.channels(params),
    queryFn: () => api.get<ChannelSales[]>(`/reports/channels${buildQuery({ ...params })}`),
  });
}

export function useProductSales(params: { trip_id?: string; from?: string; to?: string; limit?: number }) {
  return useQuery({
    queryKey: reportKeys.products(params),
    queryFn: () => api.get<ProductSales[]>(`/reports/products${buildQuery({ ...params })}`),
  });
}

/** URL ekspor CSV; dibuka sebagai tautan unduhan biasa. */
export const csvUrl = (path: string, params: Record<string, string | number | undefined> = {}) =>
  api.downloadURL(`${path}${buildQuery({ ...params, format: "csv" })}`);

// --- Pengaturan ------------------------------------------------------------

export const settingKeys = {
  all: ["settings"] as const,
};

export function useSettings() {
  return useQuery({
    queryKey: settingKeys.all,
    queryFn: () => api.get<Setting[]>("/settings"),
    staleTime: 5 * 60_000,
  });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (settings: Record<string, string>) => api.put("/settings", { settings }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingKeys.all });
    },
  });
}

// --- Pengguna --------------------------------------------------------------

export const userKeys = {
  all: ["users"] as const,
  list: (params: ListParams) => ["users", "list", params] as const,
};

export function useUsers(params: ListParams) {
  return useQuery({
    queryKey: userKeys.list(params),
    queryFn: () => api.list<User>(`/users${buildQuery({ ...params })}`),
  });
}

export function useSaveUser(id?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      id ? api.put<User>(`/users/${id}`, payload) : api.post<User>("/users", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all });
    },
  });
}

export function useResetUserPassword() {
  return useMutation({
    mutationFn: ({ id, password }: { id: string; password: string }) =>
      api.post(`/users/${id}/reset-password`, { new_password: password }),
  });
}

export function useDeleteUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.delete(`/users/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.all });
    },
  });
}

// --- Audit log -------------------------------------------------------------

export function useAuditLogs(params: ListParams & { entity?: string; entity_id?: string }) {
  return useQuery({
    queryKey: ["audit-logs", params],
    queryFn: () => api.list<AuditLog>(`/audit-logs${buildQuery({ ...params })}`),
  });
}
