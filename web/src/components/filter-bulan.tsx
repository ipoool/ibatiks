"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export interface Periode {
  from?: string;
  to?: string;
}

/**
 * Mengubah "2026-08" menjadi rentang tanggal awal dan akhir bulannya.
 *
 * Hari terakhir dihitung lewat tanggal 0 bulan berikutnya, jadi Februari dan
 * tahun kabisat ikut benar tanpa daftar jumlah hari yang harus dijaga sendiri.
 */
export function rentangBulan(bulan: string): Periode {
  const cocok = /^(\d{4})-(\d{2})$/.exec(bulan);
  if (!cocok) return {};

  const akhir = new Date(Number(cocok[1]), Number(cocok[2]), 0).getDate();
  return { from: `${bulan}-01`, to: `${bulan}-${String(akhir).padStart(2, "0")}` };
}

/**
 * Penyaring bulan untuk laporan.
 *
 * Satu komponen dipakai seluruh tab supaya "bulan ini" berarti hal yang sama di
 * mana pun ia muncul — dan supaya perhitungan hari terakhir bulan hanya ada di
 * satu tempat. Bulan disimpan di state pemanggil bersama penyaring lainnya;
 * yang dikembalikan lewat onChange sudah berupa rentang tanggal, sebab itu yang
 * dimengerti API.
 */
export function useFilterBulan(): { bulan: string; periode: Periode; kendali: React.ReactNode } {
  const [bulan, setBulan] = useState("");

  return {
    bulan,
    periode: rentangBulan(bulan),
    kendali: (
      <>
        {/* Input bulan bawaan browser: pemilih bulannya sudah disediakan sistem,
            dan menyusun dropdown bulan-tahun sendiri hanya menambah dua kolom
            yang harus dijaga tetap sinkron. */}
        <Input
          type="month"
          aria-label="Bulan"
          value={bulan}
          onChange={(event) => setBulan(event.target.value)}
          className="sm:w-44"
        />
        {bulan && (
          <Button variant="ghost" size="sm" onClick={() => setBulan("")}>
            Semua bulan
          </Button>
        )}
      </>
    ),
  };
}
