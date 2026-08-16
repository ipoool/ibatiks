"use client";

import {
  BarChart3,
  Boxes,
  ClipboardList,
  Database,
  FileText,
  LayoutDashboard,
  Luggage,
  MoreHorizontal,
  Package,
  PackageCheck,
  Plane,
  Receipt,
  Settings,
  ShoppingCart,
  Store,
  Truck,
  Users,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import type { Permission, UserRole } from "@/types/api";

export interface NavItem {
  href: string;
  label: string;
  icon: LucideIcon;
  /** Role yang boleh melihat menu ini. */
  roles: UserRole[];
  /**
   * Hak akses yang harus dimiliki pengguna. Owner bisa mencabutnya per
   * pengguna lewat Pengaturan, dan backend menolak endpoint-nya dengan aturan
   * yang sama — menu yang hilang di sini bukan sekadar disembunyikan.
   *
   * Kosong berarti menu itu terbuka untuk siapa pun yang sudah login, seperti
   * Dashboard yang isinya menyesuaikan hak masing-masing.
   */
  permission?: Permission;
}

export interface NavSection {
  title: string;
  /** Ikon yang mewakili seksi saat isinya diringkas jadi satu baris menu. */
  icon: LucideIcon;
  items: NavItem[];
}

const ALL: UserRole[] = ["owner", "admin", "tripper"];
const STAFF: UserRole[] = ["owner", "admin"];
const OWNER: UserRole[] = ["owner"];

/**
 * Struktur menu mengikuti urutan kerja sehari-hari, bukan urutan tabel di
 * database: dari merencanakan trip, mencatat order, belanja, sampai mengirim.
 */
export const NAV_SECTIONS: NavSection[] = [
  {
    title: "Ringkasan",
    icon: LayoutDashboard,
    items: [{ href: "/", label: "Dashboard", icon: LayoutDashboard, roles: ALL }],
  },
  {
    title: "Perjalanan",
    icon: Luggage,
    items: [
      { href: "/trips", label: "Trip", icon: Plane, roles: ALL, permission: "trips" },
      { href: "/shopping-list", label: "Daftar Belanja", icon: ClipboardList, roles: ALL, permission: "shopping_list" },
      { href: "/purchases", label: "Pembelian", icon: ShoppingCart, roles: ALL, permission: "purchases" },
    ],
  },
  {
    title: "Penjualan",
    icon: Store,
    items: [
      { href: "/orders", label: "Order", icon: Receipt, roles: STAFF, permission: "orders" },
      { href: "/invoices", label: "Invoice", icon: FileText, roles: STAFF, permission: "invoices" },
      // Barang dikemas dulu baru diserahkan ke kurir, jadi menunya berurutan
      // sama seperti pekerjaannya.
      { href: "/packing", label: "Siap Kemas", icon: PackageCheck, roles: STAFF, permission: "packing" },
      { href: "/shipments", label: "Pengiriman", icon: Truck, roles: STAFF, permission: "shipments" },
    ],
  },
  {
    title: "Data Master",
    icon: Database,
    items: [
      { href: "/customers", label: "Customer", icon: Users, roles: STAFF, permission: "customers" },
      { href: "/products", label: "Produk", icon: Package, roles: ALL, permission: "products" },
      { href: "/stock", label: "Stok", icon: Boxes, roles: STAFF, permission: "stock" },
    ],
  },
  {
    title: "Lainnya",
    icon: MoreHorizontal,
    items: [
      { href: "/reports", label: "Laporan", icon: BarChart3, roles: STAFF, permission: "reports" },
      { href: "/settings", label: "Pengaturan", icon: Settings, roles: OWNER, permission: "settings" },
    ],
  },
];

/**
 * Menyaring menu sesuai role dan hak akses, sekaligus membuang seksi yang jadi
 * kosong.
 *
 * Hak akses efektif dihitung backend dan ikut di dalam data pengguna, jadi
 * aturan siapa boleh apa hanya ditulis di satu tempat.
 */
export function visibleSections(user: {
  role: UserRole;
  effective_permissions?: Permission[] | null;
}): NavSection[] {
  const granted = user.effective_permissions ?? [];

  return NAV_SECTIONS.map((section) => ({
    ...section,
    items: section.items.filter(
      (item) =>
        item.roles.includes(user.role) &&
        (!item.permission || granted.includes(item.permission)),
    ),
  })).filter((section) => section.items.length > 0);
}

/** Menandai menu aktif; "/" hanya cocok persis agar tidak selalu tersorot. */
export function isActiveHref(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/";
  return pathname === href || pathname.startsWith(`${href}/`);
}

/**
 * Menu yang sedang terbuka di dalam sebuah seksi. Dipakai baris seksi untuk
 * menandai dirinya aktif sekaligus menuliskan nama halaman yang sedang dibuka —
 * tanpa itu, meringkas submenu ke dalam popup membuat pengguna kehilangan
 * petunjuk sedang berada di mana.
 */
export function activeItem(pathname: string, section: NavSection): NavItem | undefined {
  return section.items.find((item) => isActiveHref(pathname, item.href));
}
