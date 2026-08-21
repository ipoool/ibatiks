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

/** Bulan berjalan dalam bentuk "2026-08", bentuk yang dimengerti input bulan. */
export function bulanIni(): string {
  const kini = new Date();
  return `${kini.getFullYear()}-${String(kini.getMonth() + 1).padStart(2, "0")}`;
}

/**
 * Nama bulan yang dibaca orang, misalnya "Agustus 2026".
 *
 * Dipakai judul yang ikut berubah mengikuti penyaringnya. Judul tetap seperti
 * "Omzet bulan ini" akan berbohong begitu orang memilih bulan lain, dan angka
 * yang salah label lebih berbahaya daripada angka yang tidak dijelaskan.
 */
export function labelBulan(bulan: string): string {
  const { from } = rentangBulan(bulan);
  if (!from) return "sepanjang waktu";
  return new Date(from).toLocaleDateString("id-ID", { month: "long", year: "numeric" });
}

/**
 * Penyaring bulan untuk laporan.
 *
 * Satu komponen dipakai seluruh tab supaya "bulan ini" berarti hal yang sama di
 * mana pun ia muncul — dan supaya perhitungan hari terakhir bulan hanya ada di
 * satu tempat. Bulan disimpan di state pemanggil bersama penyaring lainnya;
 * yang dikembalikan lewat onChange sudah berupa rentang tanggal, sebab itu yang
 * dimengerti API.
 *
 * bulanAwal mengisi keadaan awalnya. Laporan memulai tanpa penyaring supaya
 * seluruh riwayat terlihat; Dashboard memulai dari bulan berjalan, sebab yang
 * ditanyakan orang saat membukanya adalah "bulan ini bagaimana".
 */
export function useFilterBulan(bulanAwal = ""): {
  bulan: string;
  periode: Periode;
  kendali: React.ReactNode;
} {
  const [bulan, setBulan] = useState(bulanAwal);

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

/** Awal dan akhir bulan berjalan, dipakai sebagai rentang bawaan. */
function bulanBerjalan(): Periode {
  const kini = new Date();
  const bulan = `${kini.getFullYear()}-${String(kini.getMonth() + 1).padStart(2, "0")}`;
  return rentangBulan(bulan);
}

/**
 * Penyaring rentang tanggal, bawaannya bulan berjalan.
 *
 * Dipakai daftar yang isinya menumpuk terus dan hampir selalu ditengok untuk
 * periode berjalan — daftar invoice, misalnya. Bawaan yang menyaring memang
 * berisiko membuat orang mengira datanya hilang, jadi tanggalnya selalu terlihat
 * di layar dan ada tombol untuk melepasnya sekaligus.
 */
export function useFilterRentang(): { periode: Periode; kendali: React.ReactNode } {
  const [rentang, setRentang] = useState<Periode>(bulanBerjalan);

  const ubah = (bagian: keyof Periode) => (nilai: string) =>
    setRentang({ ...rentang, [bagian]: nilai || undefined });

  return {
    periode: rentang,
    kendali: (
      <>
        <div className="flex items-center gap-2">
          <Input
            type="date"
            aria-label="Tanggal mulai"
            value={rentang.from ?? ""}
            // Batas atas mengikuti tanggal akhir supaya rentang terbalik tidak
            // pernah terbentuk — kosong hasilnya, dan tidak ada di layar yang
            // menjelaskan kenapa.
            max={rentang.to}
            onChange={(event) => ubah("from")(event.target.value)}
            className="sm:w-40"
          />
          <span className="text-sm text-muted-foreground">–</span>
          <Input
            type="date"
            aria-label="Tanggal akhir"
            value={rentang.to ?? ""}
            min={rentang.from}
            onChange={(event) => ubah("to")(event.target.value)}
            className="sm:w-40"
          />
        </div>
        {(rentang.from || rentang.to) && (
          <Button variant="ghost" size="sm" onClick={() => setRentang({})}>
            Semua tanggal
          </Button>
        )}
      </>
    ),
  };
}
