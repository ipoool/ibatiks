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
 * Rute yang tidak terdaftar di sini terbuka untuk siapa pun yang sudah login.
 */
export const ROUTE_PERMISSIONS: ReadonlyArray<{ prefix: string; permission: Permission }> = [
  /*
   * Dashboard punya hak aksesnya sendiri walaupun isinya datang dari endpoint
   * laporan. Tanpa hak itu, halamannya tetap terbuka tapi datanya ditolak, dan
   * yang terbaca adalah deretan angka nol lengkap dengan "Belum ada order" —
   * seolah tokonya kosong, padahal orangnya cuma tidak berhak melihat.
   */
  { prefix: "/", permission: "dashboard" },
  { prefix: "/trips", permission: "trips" },
  { prefix: "/shopping-list", permission: "shopping_list" },
  { prefix: "/purchases", permission: "purchases" },
  { prefix: "/orders", permission: "orders" },
  { prefix: "/invoices", permission: "invoices" },
  // /packing sudah dialihkan ke /shipments, tapi tetap terdaftar di sini:
  // pengalihannya berjalan setelah middleware, jadi tanpa baris ini orang yang
  // membuka bookmark lamanya akan ditolak sebelum sempat dialihkan.
  { prefix: "/packing", permission: "shipments" },
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

/**
 * Urutan halaman yang dicoba sebagai tempat mendarat, mengikuti urutan sidebar.
 *
 * Dipakai saat seseorang membuka alamat yang bukan haknya, dan sesudah login.
 * Dulu keduanya diarahkan ke "/" begitu saja; sejak Dashboard punya syarat hak
 * sendiri, itu berarti melempar orang ke halaman yang juga akan menolaknya.
 */
const LANDING_PATHS: readonly string[] = [
  "/",
  "/trips",
  "/shopping-list",
  "/purchases",
  "/orders",
  "/shipments",
  "/invoices",
  "/customers",
  "/products",
  "/stock",
  "/reports",
  "/settings",
];

/**
 * Halaman pertama yang boleh dibuka pengguna ini.
 *
 * Mengembalikan "/" kalau tidak ada satu pun yang cocok — termasuk saat hak
 * aksesnya benar-benar kosong. Dashboard yang menjelaskan keadaannya; menahan
 * orang tanpa halaman tujuan hanya menghasilkan putaran pengalihan.
 */
export function firstAllowedPath(granted: readonly string[]): string {
  return LANDING_PATHS.find((path) => canOpenPath(path, granted)) ?? "/";
}
