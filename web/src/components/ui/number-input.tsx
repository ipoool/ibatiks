"use client";

import * as React from "react";

import { Input } from "@/components/ui/input";

interface NumberInputProps extends Omit<React.ComponentProps<typeof Input>, "value" | "onChange"> {
  value: number;
  /** Dipanggil dengan angka hasil ketikan; `raw` kosong berarti kolomnya dikosongkan. */
  onValueChange: (value: number, raw: string) => void;
  /**
   * Menampilkan kolom kosong ketika nilainya 0. Dipakai untuk isian yang 0-nya
   * berarti "belum diisi" — berat, dimensi paket — supaya pengguna disambut
   * kolom kosong alih-alih angka nol yang harus dihapus dulu.
   */
  blankWhenZero?: boolean;
}

/**
 * Kolom angka yang boleh dikosongkan saat sedang diketik.
 *
 * `<Input type="number">` biasa dipasangkan dengan `Number(e.target.value)`
 * punya satu kebiasaan yang menjengkelkan: begitu isinya dihapus, nilainya
 * menjadi 0 dan angka 0 itu langsung terlukis kembali ke kolom. Pengguna yang
 * hendak mengganti "0" dengan "500" berakhir mengetik "0500", dan yang hendak
 * mengosongkan kolom tidak pernah bisa.
 *
 * Di sini teks yang diketik disimpan apa adanya, sementara pemanggil tetap
 * menerima angka. Nilai dari luar (misalnya saat form diisi untuk mengedit)
 * tetap menang: perubahannya dideteksi saat render, bukan lewat efek, sesuai
 * pola "menyesuaikan state ketika prop berubah".
 */
export function NumberInput({
  value,
  onValueChange,
  blankWhenZero = false,
  ...props
}: NumberInputProps) {
  const asText = (n: number) => (blankWhenZero && n === 0 ? "" : String(n));

  const [draft, setDraft] = React.useState(() => asText(value));
  const [lastValue, setLastValue] = React.useState(value);

  if (value !== lastValue) {
    setLastValue(value);
    // Ketikan yang sedang berjalan tidak ditimpa selama angkanya sama; yang
    // ditimpa hanya saat nilai dari luar benar-benar berbeda.
    if (Number(draft || 0) !== value) setDraft(asText(value));
  }

  return (
    <Input
      {...props}
      type="number"
      inputMode="numeric"
      value={draft}
      onFocus={(event) => {
        // Isi "0" disorot begitu kolomnya diklik supaya ketikan berikutnya
        // menggantikannya, bukan menempel jadi "0500".
        if (event.target.value === "0") event.target.select();
        props.onFocus?.(event);
      }}
      onChange={(event) => {
        const raw = event.target.value;
        setDraft(raw);
        onValueChange(raw === "" ? 0 : Number(raw), raw);
      }}
    />
  );
}
