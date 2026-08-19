"use client";

import * as React from "react";

import { Input } from "@/components/ui/input";
import { cn, formatIDR, toNumber } from "@/lib/utils";

export type SatuanDP = "rp" | "persen";

export interface NilaiDP {
  satuan: SatuanDP;
  nilai: string;
}

export const DP_KOSONG: NilaiDP = { satuan: "rp", nilai: "" };

/** Membaca DP dari database, yang selalu tersimpan sebagai rupiah. */
export function dpDariRupiah(rupiah: string | number | null | undefined): NilaiDP {
  return { satuan: "rp", nilai: rupiah == null ? "" : String(rupiah) };
}

/**
 * Nominal rupiah yang dikirim ke API. String kosong berarti admin tidak
 * menentukan, dan backend memakai persentase bawaannya sendiri.
 *
 * Persen tidak pernah ikut dikirim: yang disepakati dengan customer adalah
 * angka rupiahnya. Kalau yang tersimpan persen, sebuah item yang ditambahkan
 * belakangan akan diam-diam menggeser DP yang sudah dibayar.
 */
export function dpKeRupiah(dp: NilaiDP, nilaiBarang: number): string {
  if (dp.nilai === "") return "";
  if (dp.satuan === "rp") return dp.nilai;
  return String(Math.round((nilaiBarang * toNumber(dp.nilai)) / 100));
}

/** Membulatkan persen ke dua desimal, supaya tidak muncul 49,999999999. */
function persenRapi(n: number): number {
  return Math.round(n * 100) / 100;
}

function tukarSatuan(dp: NilaiDP, ke: SatuanDP, nilaiBarang: number): NilaiDP {
  // Nominalnya dipertahankan, yang berubah cuma cara menulisnya — Rp1.727.700
  // menjadi 50, bukan menjadi 1.727.700 persen. Tanpa nilai barang, persen
  // tidak punya arti apa-apa, jadi angkanya dibiarkan sampai ada isinya.
  if (dp.satuan === ke || dp.nilai === "" || nilaiBarang <= 0) {
    return { satuan: ke, nilai: dp.nilai };
  }
  const n = toNumber(dp.nilai);
  return {
    satuan: ke,
    nilai:
      ke === "persen"
        ? String(persenRapi((n / nilaiBarang) * 100))
        : String(Math.round((nilaiBarang * n) / 100)),
  };
}

interface InputDPProps {
  id: string;
  value: NilaiDP;
  onChange: (next: NilaiDP) => void;
  /** Subtotal dikurangi diskon — dasar hitungan persen. */
  nilaiBarang: number;
  placeholder?: string;
  disabled?: boolean;
}

/**
 * Isian DP yang bisa diketik sebagai rupiah atau sebagai persen.
 *
 * Kesepakatan dengan customer hampir selalu berbunyi "DP setengah dulu", bukan
 * "DP satu juta tujuh ratus dua puluh tujuh ribu tujuh ratus". Memaksa admin
 * menghitung sendiri persen menjadi rupiah adalah pekerjaan yang tidak perlu,
 * dan salah ketik satu digit di situ tidak kelihatan salah.
 *
 * Yang keluar dari komponen ini tetap `NilaiDP`; pemanggilnya yang mengubahnya
 * jadi rupiah lewat `dpKeRupiah` saat menyusun permintaan.
 */
export function InputDP({
  id,
  value,
  onChange,
  nilaiBarang,
  placeholder,
  disabled,
}: InputDPProps) {
  const persen = value.satuan === "persen";

  return (
    <div className="flex items-center gap-2">
      <Input
        id={id}
        type="number"
        min="0"
        // Persen di atas 100 berarti DP melebihi nilai barang — backend
        // menolaknya, jadi lebih baik tertahan di sini sebelum terkirim.
        max={persen ? "100" : undefined}
        step="any"
        inputMode="decimal"
        value={value.nilai}
        onChange={(event) => onChange({ ...value, nilai: event.target.value })}
        placeholder={placeholder}
        disabled={disabled}
        className="min-w-0 flex-1"
      />
      <div className="flex shrink-0 rounded-md border border-input p-0.5">
        {(["rp", "persen"] as const).map((satuan) => (
          <button
            // Wajib type="button". Tanpa itu tombolnya ikut mengirim form yang
            // membungkusnya, dan mengganti satuan malah membuat ordernya.
            type="button"
            key={satuan}
            disabled={disabled}
            aria-pressed={value.satuan === satuan}
            onClick={() => onChange(tukarSatuan(value, satuan, nilaiBarang))}
            className={cn(
              "rounded px-2.5 py-1 text-xs font-medium transition-colors",
              "focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-none",
              "disabled:cursor-not-allowed disabled:opacity-50",
              value.satuan === satuan
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            {satuan === "rp" ? "Rp" : "%"}
          </button>
        ))}
      </div>
    </div>
  );
}

/**
 * Terjemahan isian ke satuan yang satunya, untuk ditaruh di bawah kolomnya.
 *
 * Admin mengetik salah satu, tapi yang disebut ke customer bisa yang mana saja
 * — jadi keduanya selalu terlihat tanpa perlu menukar satuan bolak-balik.
 */
export function keteranganDP(dp: NilaiDP, nilaiBarang: number): string | null {
  if (dp.nilai === "" || nilaiBarang <= 0) return null;
  const rupiah = toNumber(dpKeRupiah(dp, nilaiBarang));
  if (dp.satuan === "persen") return `${formatIDR(rupiah)} dari nilai barang`;
  return `${persenRapi((rupiah / nilaiBarang) * 100)}% dari nilai barang`;
}
