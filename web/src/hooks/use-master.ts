"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api, buildQuery } from "@/lib/api";
import type {
  Customer,
  CustomerStats,
  Social,
  PricePreview,
  Product,
  ProductCategory,
  ProductPriceHistory,
} from "@/types/api";

export interface ListParams {
  page?: number;
  per_page?: number;
  q?: string;
  sort?: string;
  order?: "asc" | "desc";
}

// --- Customer --------------------------------------------------------------

export const customerKeys = {
  all: ["customers"] as const,
  list: (params: ListParams) => ["customers", "list", params] as const,
  detail: (id: string) => ["customers", "detail", id] as const,
  stats: (id: string) => ["customers", "stats", id] as const,
};

export function useCustomers(params: ListParams) {
  return useQuery({
    queryKey: customerKeys.list(params),
    queryFn: () => api.list<Customer>(`/customers${buildQuery({ ...params })}`),
  });
}

export function useCustomer(id: string | undefined) {
  return useQuery({
    queryKey: customerKeys.detail(id ?? ""),
    queryFn: () => api.get<Customer>(`/customers/${id}`),
    enabled: Boolean(id),
  });
}

export function useCustomerStats(id: string | undefined) {
  return useQuery({
    queryKey: customerKeys.stats(id ?? ""),
    queryFn: () => api.get<CustomerStats>(`/customers/${id}/stats`),
    enabled: Boolean(id),
  });
}

export type CustomerPayload = {
  name: string;
  phone_wa: string;
  email?: string | null;
  socials?: Social[];
  address?: string | null;
  city?: string | null;
  district?: string | null;
  subdistrict?: string | null;
  province?: string | null;
  postal_code?: string | null;
  notes?: string | null;
};

export function useSaveCustomer(id?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CustomerPayload) =>
      id
        ? api.put<Customer>(`/customers/${id}`, payload)
        : api.post<Customer>("/customers", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customerKeys.all });
    },
  });
}

export function useDeleteCustomer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.delete(`/customers/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customerKeys.all });
    },
  });
}

// --- Kategori produk -------------------------------------------------------

export const categoryKeys = {
  all: ["product-categories"] as const,
};

export function useCategories() {
  return useQuery({
    queryKey: categoryKeys.all,
    queryFn: () => api.get<ProductCategory[]>("/product-categories"),
    // Kategori jarang berubah, jadi tidak perlu sering diambil ulang.
    staleTime: 5 * 60_000,
  });
}

export function useSaveCategory(id?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: { name: string; description?: string | null }) =>
      id
        ? api.put<ProductCategory>(`/product-categories/${id}`, payload)
        : api.post<ProductCategory>("/product-categories", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: categoryKeys.all });
    },
  });
}

export function useDeleteCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.delete(`/product-categories/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: categoryKeys.all });
    },
  });
}

// --- Produk ----------------------------------------------------------------

export interface ProductListParams extends ListParams {
  category_id?: string;
  active_only?: boolean;
}

export const productKeys = {
  all: ["products"] as const,
  list: (params: ProductListParams) => ["products", "list", params] as const,
  detail: (id: string) => ["products", "detail", id] as const,
};

/**
 * Riwayat harga produk dari trip ke trip. Diambil sesuai permintaan, bukan ikut
 * daftar produk, karena hanya dibutuhkan saat admin sedang menimbang harga.
 */
export function useProductPriceHistory(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: ["products", "price-history", id],
    queryFn: () => api.get<ProductPriceHistory[]>(`/products/${id}/price-history`),
    enabled: Boolean(id) && enabled,
  });
}

export function useProducts(params: ProductListParams) {
  return useQuery({
    queryKey: productKeys.list(params),
    queryFn: () => api.list<Product>(`/products${buildQuery({ ...params })}`),
  });
}

export function useProduct(id: string | undefined) {
  return useQuery({
    queryKey: productKeys.detail(id ?? ""),
    queryFn: () => api.get<Product>(`/products/${id}`),
    enabled: Boolean(id),
  });
}

export type ProductPayload = {
  sku?: string;
  name: string;
  category_id?: string | null;
  brand?: string | null;
  store_name?: string | null;
  base_currency?: string;
  base_price?: string;
  markup_type: "percent" | "nominal";
  markup_value?: string;
  weight_gram?: number;
  image_url?: string | null;
  notes?: string | null;
  is_active?: boolean;
};

export function useSaveProduct(id?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: ProductPayload) =>
      id ? api.put<Product>(`/products/${id}`, payload) : api.post<Product>("/products", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: productKeys.all });
    },
  });
}

export function useDeleteProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.delete(`/products/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: productKeys.all });
    },
  });
}

/** Menghitung harga jual di server agar rumus markup persis sama dengan saat disimpan. */
export function usePricePreview() {
  return useMutation({
    mutationFn: (payload: {
      cost_price: string;
      exchange_rate: string;
      markup_type: "percent" | "nominal";
      markup_value: string;
    }) => api.post<PricePreview>("/products/price-preview", payload),
  });
}
