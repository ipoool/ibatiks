"use client";

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { MATA_UANG_ASEAN, MATA_UANG_LAIN, labelMataUang, mataUangDikenal } from "@/lib/mata-uang";

/**
 * Pemilih mata uang, dikelompokkan ASEAN lebih dulu.
 *
 * Dulu kolom ini isian teks bebas. Salah ketik satu huruf — "IRD" alih-alih
 * "IDR" — membuat kurs gagal diambil dan seluruh harga jual trip meleset,
 * sementara tulisannya tetap terlihat wajar sekilas.
 */
export function CurrencySelect({
  value,
  onChange,
  id,
  disabled,
  className,
}: {
  value: string;
  onChange: (value: string) => void;
  id?: string;
  disabled?: boolean;
  className?: string;
}) {
  /*
   * Trip lama bisa saja memakai mata uang di luar daftar — kursnya sudah
   * dikunci dan tidak boleh berubah hanya karena daftarnya menyusut. Kodenya
   * disisipkan sebagai pilihan tambahan supaya membuka form edit tidak
   * diam-diam menggantinya.
   */
  const asing = value && !mataUangDikenal(value) ? value : null;

  return (
    <Select value={value} onValueChange={onChange} disabled={disabled}>
      <SelectTrigger id={id} className={cn("w-full", className)}>
        <SelectValue placeholder="Pilih mata uang…" />
      </SelectTrigger>
      <SelectContent>
        {asing && (
          <SelectGroup>
            <SelectLabel>Tersimpan sebelumnya</SelectLabel>
            <SelectItem value={asing}>{labelMataUang(asing)}</SelectItem>
          </SelectGroup>
        )}
        <SelectGroup>
          <SelectLabel>ASEAN</SelectLabel>
          {MATA_UANG_ASEAN.map((m) => (
            <SelectItem key={m.kode} value={m.kode}>
              {m.label}
            </SelectItem>
          ))}
        </SelectGroup>
        <SelectGroup>
          <SelectLabel>Tujuan lain</SelectLabel>
          {MATA_UANG_LAIN.map((m) => (
            <SelectItem key={m.kode} value={m.kode}>
              {m.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
