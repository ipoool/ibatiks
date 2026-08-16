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

import type { UserRole } from "@/types/api";

export interface NavItem {
  href: string;
  label: string;
  icon: LucideIcon;
  /** Role yang boleh melihat menu ini. */
  roles: UserRole[];
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
      { href: "/trips", label: "Trip", icon: Plane, roles: ALL },
      { href: "/shopping-list", label: "Daftar Belanja", icon: ClipboardList, roles: ALL },
      { href: "/purchases", label: "Pembelian", icon: ShoppingCart, roles: ALL },
    ],
  },
  {
    title: "Penjualan",
    icon: Store,
    items: [
      { href: "/orders", label: "Order", icon: Receipt, roles: STAFF },
      { href: "/invoices", label: "Invoice", icon: FileText, roles: STAFF },
      { href: "/shipments", label: "Pengiriman", icon: Truck, roles: STAFF },
      { href: "/packing", label: "Siap Kemas", icon: PackageCheck, roles: STAFF },
    ],
  },
  {
    title: "Data Master",
    icon: Database,
    items: [
      { href: "/customers", label: "Customer", icon: Users, roles: STAFF },
      { href: "/products", label: "Produk", icon: Package, roles: ALL },
      { href: "/stock", label: "Stok", icon: Boxes, roles: STAFF },
    ],
  },
  {
    title: "Lainnya",
    icon: MoreHorizontal,
    items: [
      { href: "/reports", label: "Laporan", icon: BarChart3, roles: STAFF },
      { href: "/settings", label: "Pengaturan", icon: Settings, roles: OWNER },
    ],
  },
];

/** Menyaring menu sesuai role, sekaligus membuang seksi yang jadi kosong. */
export function visibleSections(role: UserRole): NavSection[] {
  return NAV_SECTIONS.map((section) => ({
    ...section,
    items: section.items.filter((item) => item.roles.includes(role)),
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
