"use client";

import Link from "next/link";
import { useState } from "react";

import { activeItem, isActiveHref, type NavSection } from "@/components/layout/nav";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

const rowClass =
  "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors";
const activeRowClass = "bg-primary text-primary-foreground";
const idleRowClass = "text-muted-foreground hover:bg-accent hover:text-accent-foreground";

/**
 * Menu samping berbentuk accordion: tiap kelompok hanya satu baris, isinya
 * terbentang ke bawah saat ditekan.
 *
 * Daftar penuh sebelumnya memakan seluruh tinggi layar dan membuat menu yang
 * jarang dipakai terlihat sama mendesaknya dengan yang dipakai tiap hari.
 * Hanya satu kelompok yang terbuka pada satu waktu, sebab membiarkan semuanya
 * terbuka mengembalikan panjang daftar seperti semula.
 */
export function NavMenu({
  sections,
  pathname,
  onNavigate,
  collapsed = false,
}: {
  sections: NavSection[];
  pathname: string;
  onNavigate: () => void;
  /** Sidebar menyempit jadi deretan ikon; label pindah ke tooltip. */
  collapsed?: boolean;
}) {
  // Dipecah jadi dua komponen, bukan satu dengan percabangan di dalamnya:
  // ragam terlipat tidak memerlukan state accordion sama sekali.
  return collapsed ? (
    <NavRail sections={sections} pathname={pathname} onNavigate={onNavigate} />
  ) : (
    <NavAccordion sections={sections} pathname={pathname} onNavigate={onNavigate} />
  );
}

/**
 * Ragam sempit: seluruh menu dijejer sebagai ikon, dikelompokkan garis.
 *
 * Accordion ditinggalkan sama sekali di sini. Melipat kelompok tidak ada
 * gunanya kalau judulnya sendiri tak terbaca — yang tersisa hanya ikon, jadi
 * semua menu dibuat sejajar. Hasilnya justru satu klik ke mana pun, bukan
 * buka-kelompok lalu pilih.
 */
function NavRail({
  sections,
  pathname,
  onNavigate,
}: {
  sections: NavSection[];
  pathname: string;
  onNavigate: () => void;
}) {
  return (
    <nav className="flex flex-1 flex-col gap-1 overflow-y-auto px-2 py-4">
      {sections.map((section, index) => (
        <div key={section.title} className="flex flex-col gap-1">
          {index > 0 && <div className="my-1 border-t border-border" />}
          {section.items.map((item) => {
            const active = isActiveHref(pathname, item.href);
            return (
              <Tooltip key={item.href}>
                <TooltipTrigger asChild>
                  <Link
                    href={item.href}
                    onClick={onNavigate}
                    aria-current={active ? "page" : undefined}
                    className={cn(
                      "flex size-10 items-center justify-center rounded-lg transition-colors",
                      active ? activeRowClass : idleRowClass,
                    )}
                  >
                    <item.icon className="size-4" />
                    <span className="sr-only">{item.label}</span>
                  </Link>
                </TooltipTrigger>
                <TooltipContent side="right">
                  {section.title} · {item.label}
                </TooltipContent>
              </Tooltip>
            );
          })}
        </div>
      ))}
    </nav>
  );
}

function NavAccordion({
  sections,
  pathname,
  onNavigate,
}: {
  sections: NavSection[];
  pathname: string;
  onNavigate: () => void;
}) {
  // Kelompok yang memuat halaman terbuka dibentangkan lebih dulu, supaya
  // sesampainya di sebuah halaman pengguna langsung melihat posisinya.
  const activeSection = sections.find((section) => activeItem(pathname, section))?.title ?? "";

  /*
   * Pilihan pengguna disimpan bersama pathname tempat pilihan itu dibuat.
   * Begitu pindah halaman, catatannya otomatis tidak berlaku lagi dan menu
   * kembali mengikuti kelompok yang aktif — tanpa perlu efek yang menyetel
   * state setiap kali pathname berubah.
   */
  const [choice, setChoice] = useState<{
    pathname: string;
    title: string;
  } | null>(null);
  const openTitle = choice && choice.pathname === pathname ? choice.title : activeSection;

  const foldable = sections.filter((section) => section.items.length > 1);
  const single = sections.filter((section) => section.items.length === 1);

  return (
    <nav className="flex flex-1 flex-col gap-1 overflow-y-auto px-3 py-4">
      {/*
        Kelompok berisi satu menu tidak dilipat: membukanya hanya menambah satu
        klik tanpa menyembunyikan apa pun. Ditempatkan lebih dulu karena satu-
        satunya kelompok seperti itu, Dashboard, memang selalu di paling atas.
      */}
      {single.map((section) => {
        const [only] = section.items;
        if (!only) return null;
        const active = isActiveHref(pathname, only.href);
        return (
          <Link
            key={section.title}
            href={only.href}
            onClick={onNavigate}
            aria-current={active ? "page" : undefined}
            className={cn(rowClass, active ? activeRowClass : idleRowClass)}
          >
            <only.icon className="size-4 shrink-0" />
            <span className="flex-1 text-left">{only.label}</span>
          </Link>
        );
      })}

      <Accordion
        type="single"
        collapsible
        value={openTitle}
        onValueChange={(title) => setChoice({ pathname, title })}
        className="flex flex-col gap-1"
      >
        {foldable.map((section) => {
          const current = activeItem(pathname, section);

          return (
            <AccordionItem key={section.title} value={section.title} className="border-b-0">
              <AccordionTrigger
                className={cn(
                  rowClass,
                  "py-2 hover:no-underline",
                  current && openTitle !== section.title ? activeRowClass : idleRowClass,
                )}
              >
                <span className="flex min-w-0 flex-1 items-center gap-3">
                  <section.icon className="size-4 shrink-0" />
                  <span className="min-w-0 flex-1 text-left">
                    {section.title}
                    {/* Saat terlipat, nama halaman yang sedang dibuka ikut
                        dituliskan; begitu terbentang, menunya sudah kelihatan
                        sendiri. */}
                    {current && openTitle !== section.title && (
                      <span className="block truncate text-xs font-normal opacity-80">
                        {current.label}
                      </span>
                    )}
                  </span>
                </span>
              </AccordionTrigger>

              <AccordionContent className="pb-0">
                <div className="mt-1 ml-4 space-y-1 border-l border-border pl-3">
                  {section.items.map((item) => {
                    const active = isActiveHref(pathname, item.href);
                    return (
                      <Link
                        key={item.href}
                        href={item.href}
                        onClick={onNavigate}
                        aria-current={active ? "page" : undefined}
                        className={cn(rowClass, "py-1.5", active ? activeRowClass : idleRowClass)}
                      >
                        <item.icon className="size-4 shrink-0" />
                        {item.label}
                      </Link>
                    );
                  })}
                </div>
              </AccordionContent>
            </AccordionItem>
          );
        })}
      </Accordion>
    </nav>
  );
}
