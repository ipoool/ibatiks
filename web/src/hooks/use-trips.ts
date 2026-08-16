"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api, buildQuery } from "@/lib/api";
import type { ListParams } from "@/hooks/use-master";
import type {
  ExpenseCategory,
  MarkupType,
  ShoppingListEntry,
  Trip,
  TripExpense,
  TripItem,
  TripStatus,
} from "@/types/api";

export interface TripListParams extends ListParams {
  status?: TripStatus | "";
}

export const tripKeys = {
  all: ["trips"] as const,
  list: (params: TripListParams) => ["trips", "list", params] as const,
  detail: (id: string) => ["trips", "detail", id] as const,
  items: (id: string) => ["trips", "items", id] as const,
  expenses: (id: string) => ["trips", "expenses", id] as const,
  shoppingList: (id: string) => ["trips", "shopping-list", id] as const,
};

export function useTrips(params: TripListParams) {
  return useQuery({
    queryKey: tripKeys.list(params),
    queryFn: () => api.list<Trip>(`/trips${buildQuery({ ...params })}`),
  });
}

export function useTrip(id: string | undefined) {
  return useQuery({
    queryKey: tripKeys.detail(id ?? ""),
    queryFn: () => api.get<Trip>(`/trips/${id}`),
    enabled: Boolean(id),
  });
}

export type TripPayload = {
  title: string;
  country: string;
  city?: string | null;
  tripper_user_id?: string | null;
  depart_date: string;
  return_date: string;
  order_deadline?: string;
  currency: string;
  exchange_rate: string;
  notes?: string | null;
};

export function useSaveTrip(id?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: TripPayload) =>
      id ? api.put<Trip>(`/trips/${id}`, payload) : api.post<Trip>("/trips", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tripKeys.all });
    },
  });
}

/**
 * Mengambil kurs terkini sebuah mata uang terhadap rupiah.
 *
 * Hasilnya hanya mengisi kolom kurs saat trip dibuat. Setelah tersimpan, kurs
 * trip tidak pernah ikut bergerak lagi — laporan laba trip yang sudah selesai
 * tidak boleh berubah karena kurs pasar hari ini berbeda.
 */
export function useFetchExchangeRate() {
  return useMutation({
    mutationFn: (currency: string) =>
      api.get<{ from: string; to: string; rate: string; source: string; fetched_at: string }>(
        `/exchange-rate?from=${encodeURIComponent(currency)}`,
      ),
  });
}

export function useChangeTripStatus(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (status: TripStatus) => api.patch<Trip>(`/trips/${id}/status`, { status }),
    onSuccess: () => {
      // Perubahan status trip ikut menggeser status order di dalamnya, jadi
      // cache order juga harus dianggap basi.
      queryClient.invalidateQueries({ queryKey: tripKeys.all });
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
  });
}

export function useDeleteTrip() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.delete(`/trips/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tripKeys.all });
    },
  });
}

// --- Katalog trip ----------------------------------------------------------

export function useTripItems(tripId: string | undefined) {
  return useQuery({
    queryKey: tripKeys.items(tripId ?? ""),
    queryFn: () => api.get<TripItem[]>(`/trips/${tripId}/items`),
    enabled: Boolean(tripId),
  });
}

export type TripItemPayload = {
  product_id: string;
  cost_price: string;
  markup_type: MarkupType;
  markup_value: string;
  max_qty?: number | null;
  is_active?: boolean;
  notes?: string | null;
};

export function useSaveTripItem(tripId: string, itemId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: TripItemPayload) =>
      itemId
        ? api.put<TripItem>(`/trips/${tripId}/items/${itemId}`, payload)
        : api.post<TripItem>(`/trips/${tripId}/items`, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tripKeys.items(tripId) });
      queryClient.invalidateQueries({ queryKey: tripKeys.detail(tripId) });
    },
  });
}

export function useDeleteTripItem(tripId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (itemId: string) => api.delete(`/trips/${tripId}/items/${itemId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tripKeys.items(tripId) });
      queryClient.invalidateQueries({ queryKey: tripKeys.detail(tripId) });
    },
  });
}

export type SyncExchangeRateResult = {
  trip: Trip;
  previous_rate: string;
  new_rate: string;
  source: string;
  items_updated: number;
};

/**
 * Menyegarkan kurs sebuah trip dari sumber kurs harian.
 *
 * `recalculate_prices` sengaja terpisah: menyegarkan kurs adalah pencatatan,
 * sedangkan menghitung ulang harga katalog berarti mengubah harga yang sudah
 * dilihat customer — dua hal yang tidak boleh terjadi tanpa disadari.
 */
export function useSyncExchangeRate(tripId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (recalculatePrices: boolean) =>
      api.post<SyncExchangeRateResult>(`/trips/${tripId}/sync-exchange-rate`, {
        recalculate_prices: recalculatePrices,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tripKeys.all });
    },
  });
}

/** Menghitung ulang seluruh harga katalog memakai kurs trip terkini. */
export function useRecalculatePrices(tripId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => api.post<TripItem[]>(`/trips/${tripId}/recalculate-prices`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tripKeys.items(tripId) });
    },
  });
}

// --- Biaya perjalanan ------------------------------------------------------

export function useTripExpenses(tripId: string | undefined) {
  return useQuery({
    queryKey: tripKeys.expenses(tripId ?? ""),
    queryFn: () => api.get<TripExpense[]>(`/trips/${tripId}/expenses`),
    enabled: Boolean(tripId),
  });
}

export type TripExpensePayload = {
  category: ExpenseCategory;
  description: string;
  amount: string;
  spent_at?: string;
  receipt_url?: string | null;
};

export function useSaveTripExpense(tripId: string, expenseId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: TripExpensePayload) =>
      expenseId
        ? api.put<TripExpense>(`/trips/${tripId}/expenses/${expenseId}`, payload)
        : api.post<TripExpense>(`/trips/${tripId}/expenses`, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tripKeys.expenses(tripId) });
      // Biaya perjalanan langsung memengaruhi laba bersih trip.
      queryClient.invalidateQueries({ queryKey: ["reports"] });
    },
  });
}

export function useDeleteTripExpense(tripId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (expenseId: string) => api.delete(`/trips/${tripId}/expenses/${expenseId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tripKeys.expenses(tripId) });
      queryClient.invalidateQueries({ queryKey: ["reports"] });
    },
  });
}

// --- Daftar belanja --------------------------------------------------------

export function useShoppingList(tripId: string | undefined) {
  return useQuery({
    queryKey: tripKeys.shoppingList(tripId ?? ""),
    queryFn: () => api.get<ShoppingListEntry[]>(`/trips/${tripId}/shopping-list`),
    enabled: Boolean(tripId),
    // Tripper memakai halaman ini sambil belanja, jadi datanya harus segar.
    staleTime: 0,
  });
}
