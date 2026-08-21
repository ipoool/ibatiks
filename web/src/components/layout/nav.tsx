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
  Plane,
  Receipt,
  Settings,
  ShoppingCart,
  Store,
  Truck,
  Users,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";


import { canOpenPath } from "@/lib/route-permissions";

export interface NavItem {
  href: string;
  label: string;
  icon: LucideIcon;
  /*
   * Hak akses yang dibutuhkan menu ini tidak ditulis di sini, melainkan
   * diturunkan dari href lewat src/lib/route-permissions.ts — modul yang sama
   * dipakai middleware untuk menjaga rutenya, supaya menu yang disembunyikan
   * dan halaman yang ditolak tidak pernah berbeda pendapat.
   */
}

export interface NavSection {
  title: string;
  /** Ikon yang mewakili seksi saat isinya diringkas jadi satu baris menu. */
  icon: LucideIcon;
  items: NavItem[];
}

/**
 * Struktur menu mengikuti urutan kerja sehari-hari, bukan urutan tabel di
 * database: dari merencanakan trip, mencatat order, belanja, sampai mengirim.
 */
export const NAV_SECTIONS: NavSection[] = [
  {
    title: "Ringkasan",
    icon: LayoutDashboard,
    items: [{ href: "/", label: "Dashboard", icon: LayoutDashboard }],
  },
  {
    title: "Perjalanan",
    icon: Luggage,
    items: [
      { href: "/trips", label: "Trip", icon: Plane },
      { href: "/shopping-list", label: "Daftar Belanja", icon: ClipboardList },
      { href: "/purchases", label: "Pembelian", icon: ShoppingCart },
    ],
  },
  {
    title: "Penjualan",
    icon: Store,
    items: [
      { href: "/orders", label: "Order", icon: Receipt },
      // Urutannya mengikuti perjalanan satu order: dicatat, dikemas lalu
      // dikirim, baru ditagih pelunasannya. Invoice pelunasan memang terbit
      // paling belakang, jadi menunya duduk paling bawah.
      { href: "/shipments", label: "Pengiriman", icon: Truck },
      { href: "/invoices", label: "Invoice", icon: FileText },
    ],
  },
  {
    title: "Data Master",
    icon: Database,
    items: [
      { href: "/customers", label: "Customer", icon: Users },
      { href: "/products", label: "Produk", icon: Package },
      { href: "/stock", label: "Stok", icon: Boxes },
    ],
  },
  {
    title: "Lainnya",
    icon: MoreHorizontal,
    items: [
      { href: "/reports", label: "Laporan", icon: BarChart3 },
      { href: "/settings", label: "Pengaturan", icon: Settings },
    ],
  },
];

/**
 * Menyaring menu sesuai hak akses, sekaligus membuang seksi yang jadi kosong.
 *
 * Yang menentukan hanya daftar menunya, bukan nama rolenya. Tiap menu dulu
 * membawa daftar role yang boleh melihatnya, dan daftar itu selalu berakhir
 * kembar dengan hak aksesnya — sampai role jadi data: role bikinan toko sendiri
 * bukan owner, admin, maupun tripper, jadi seluruh menu ikut tersaring habis
 * dan sidebarnya kosong melompong.
 *
 * Hak akses efektif dihitung backend dan ikut di dalam data pengguna, jadi
 * aturan siapa boleh apa hanya ditulis di satu tempat.
 */
export function visibleSections(user: {
  effective_permissions?: string[] | null;
}): NavSection[] {
  const granted = user.effective_permissions ?? [];

  return NAV_SECTIONS.map((section) => ({
    ...section,
    items: section.items.filter((item) => canOpenPath(item.href, granted)),
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
