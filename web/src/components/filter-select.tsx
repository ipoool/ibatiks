"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

/**
 * Nilai pengganti untuk pilihan "semua".
 *
 * Radix menolak SelectItem bernilai string kosong — nilai kosong dipakainya
 * untuk menandai belum ada yang dipilih. Sementara di sisi API, "tanpa filter"
 * memang dikirim sebagai string kosong. Sentinel ini yang menjembatani keduanya,
 * dan sengaja ditulis mencolok supaya tidak pernah tertukar dengan nilai asli.
 */
const ALL = "__all__";

interface FilterSelectProps<T extends string> {
  value: T | "";
  onChange: (value: T | "") => void;
  /** Label untuk pilihan tanpa filter, misalnya "Semua status". */
  allLabel: string;
  options: ReadonlyArray<{ value: T; label: string }>;
  className?: string;
  disabled?: boolean;
  "aria-label"?: string;
}

/**
 * Select untuk menyaring daftar, dengan pilihan "semua" di posisi teratas.
 *
 * Pola ini berulang di hampir setiap halaman daftar. Dibungkus supaya sentinel
 * "semua" ditangani di satu tempat: kalau tiap halaman menuliskannya sendiri,
 * cepat atau lambat ada yang lupa mengubahnya kembali menjadi string kosong dan
 * filternya diam-diam mengirim "__all__" ke server.
 */
export function FilterSelect<T extends string>({
  value,
  onChange,
  allLabel,
  options,
  className,
  disabled,
  "aria-label": ariaLabel,
}: FilterSelectProps<T>) {
  return (
    <Select
      value={value === "" ? ALL : value}
      onValueChange={(next) => onChange(next === ALL ? "" : (next as T))}
      disabled={disabled}
    >
      <SelectTrigger className={cn("w-full", className)} aria-label={ariaLabel ?? allLabel}>
        <SelectValue placeholder={allLabel} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={ALL}>{allLabel}</SelectItem>
        {options.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            {option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

interface OptionSelectProps<T extends string> {
  value: T;
  onChange: (value: T) => void;
  options: ReadonlyArray<{ value: T; label: string }>;
  id?: string;
  className?: string;
  disabled?: boolean;
  placeholder?: string;
  "aria-label"?: string;
}

/**
 * Select untuk isian form yang seluruh pilihannya sah — tidak ada "semua",
 * jadi tidak perlu sentinel. Dipisah dari FilterSelect supaya tipenya tegas:
 * nilainya tidak pernah string kosong.
 */
export function OptionSelect<T extends string>({
  value,
  onChange,
  options,
  id,
  className,
  disabled,
  placeholder,
  "aria-label": ariaLabel,
}: OptionSelectProps<T>) {
  return (
    <Select value={value} onValueChange={(next) => onChange(next as T)} disabled={disabled}>
      <SelectTrigger id={id} className={cn("w-full", className)} aria-label={ariaLabel}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {options.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            {option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

/**
 * Mengubah peta label menjadi daftar opsi.
 *
 * Banyak label di aplikasi ini sudah tersimpan sebagai Record<kode, label>
 * karena dipakai juga untuk menampilkan nilai tunggal. Helper ini menghindari
 * penulisan daftar yang sama untuk kedua keperluan — dan dengan itu, kesempatan
 * keduanya berbeda tanpa disadari.
 */
export function toOptions<T extends string>(
  labels: Record<T, string>,
): ReadonlyArray<{ value: T; label: string }> {
  return (Object.entries(labels) as Array<[T, string]>).map(([value, label]) => ({ value, label }));
}
