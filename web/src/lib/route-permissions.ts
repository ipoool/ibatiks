import type { Permission } from "@/types/api";

/**
 * Hak akses yang dibutuhkan tiap rute halaman.
 *
 * Ditulis di modul biasa, bukan di `nav.tsx`, karena dua pemakainya berada di
 * dunia yang berbeda: menu sidebar berjalan di browser, sedangkan penjaga rute
 * berjalan di middleware. Menyimpannya di modul `"use client"` akan menyeret
 * seluruh ikon menu ke bundel edge, dan konstanta yang diimpor dari sana gagal
 * diam-diam saat dibaca kode server.
 *
 * Rute yang tidak terdaftar di sini terbuka untuk siapa pun yang sudah login —
 * seperti Dashboard, yang isinya sudah menyesuaikan hak masing-masing.
 */
export const ROUTE_PERMISSIONS: ReadonlyArray<{ prefix: string; permission: Permission }> = [
  { prefix: "/trips", permission: "trips" },
  { prefix: "/shopping-list", permission: "shopping_list" },
  { prefix: "/purchases", permission: "purchases" },
  { prefix: "/orders", permission: "orders" },
  { prefix: "/invoices", permission: "invoices" },
  { prefix: "/packing", permission: "packing" },
  { prefix: "/shipments", permission: "shipments" },
  { prefix: "/customers", permission: "customers" },
  { prefix: "/products", permission: "products" },
  { prefix: "/stock", permission: "stock" },
  { prefix: "/reports", permission: "reports" },
  { prefix: "/settings", permission: "settings" },
];

/**
 * Hak akses yang dibutuhkan sebuah alamat, termasuk halaman detail di
 * bawahnya: `/orders/abc` mengikuti aturan `/orders`.
 */
export function permissionForPath(pathname: string): Permission | undefined {
  return ROUTE_PERMISSIONS.find(
    ({ prefix }) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  )?.permission;
}

/**
 * Benar bila pengguna dengan daftar hak ini boleh membuka alamat tersebut.
 *
 * Ini bukan lapisan keamanan — backend tetap menolak endpoint-nya sendiri.
 * Gunanya supaya orang yang menu-nya memang tidak ada tidak mendarat di
 * halaman yang datanya pasti gagal dimuat, lalu membaca "belum ada data"
 * seolah tokonya kosong.
 */
export function canOpenPath(pathname: string, granted: readonly string[]): boolean {
  const needed = permissionForPath(pathname);
  return !needed || granted.includes(needed);
}
