"use client";

import { useEffect, useState } from "react";

/**
 * Menunda perubahan nilai sampai pengguna berhenti mengetik.
 *
 * Dipakai pada kotak pencarian supaya setiap huruf yang diketik tidak
 * langsung memicu satu request ke server.
 */
export function useDebounced<T>(value: T, delay = 350): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debounced;
}
