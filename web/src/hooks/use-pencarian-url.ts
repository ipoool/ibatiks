"use client";

import { useSearchParams } from "next/navigation";
import { useState } from "react";

/**
 * Kata kunci pencarian yang berasal dari alamat halaman, dan ikut menulisinya
 * kembali saat diubah.
 *
 * Beberapa tempat di aplikasi menautkan ke daftar yang sudah tersaring —
 * nama customer pada detail order, kode customer pada laporan. Tanpa ini
 * tautannya mendarat di daftar penuh: kotak pencariannya kosong, dan yang
 * mengeklik mengira customernya tidak ketemu.
 *
 * Nilainya dibaca sekali sebagai keadaan awal, bukan disetel lewat efek —
 * aturan lint proyek ini melarangnya, dan menyetel state dari alamat pada tiap
 * render akan menimpa apa yang sedang diketik orang.
 *
 * Alamatnya diperbarui dengan history.replaceState, bukan router.replace: yang
 * dibutuhkan hanya supaya alamat di bilah peramban tetap jujur dan bisa disalin,
 * tanpa memicu navigasi ulang pada tiap ketikan.
 */
export function usePencarianURL(param = "q"): [string, (value: string) => void] {
  const searchParams = useSearchParams();
  const [nilai, setNilai] = useState(() => searchParams.get(param) ?? "");

  function ubah(baru: string) {
    setNilai(baru);

    const url = new URL(window.location.href);
    if (baru) url.searchParams.set(param, baru);
    else url.searchParams.delete(param);
    window.history.replaceState(null, "", url);
  }

  return [nilai, ubah];
}
