"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";

import { ApiError } from "@/lib/api";

export function QueryProvider({ children }: { children: React.ReactNode }) {
  // QueryClient dibuat di dalam state agar tiap sesi browser punya cache
  // sendiri dan tidak ikut terbawa antar-render di server.
  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            // Data back office berubah karena tindakan admin sendiri, bukan
            // dari luar, jadi refetch agresif hanya menambah beban server.
            staleTime: 30_000,
            refetchOnWindowFocus: false,
            retry: (failureCount, error) => {
              // Percuma mengulang request yang ditolak karena input atau hak
              // akses; hanya kegagalan jaringan/server yang layak dicoba lagi.
              if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
                return false;
              }
              return failureCount < 2;
            },
          },
          mutations: {
            retry: false,
          },
        },
      }),
  );

  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
