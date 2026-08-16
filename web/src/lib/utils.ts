import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/** Menggabungkan className bersyarat sekaligus menyelesaikan konflik utility Tailwind. */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Nominal uang dikirim backend sebagai string (mis. "1250000.00"), bukan angka.
 * Pilihan itu disengaja: NUMERIC di PostgreSQL bisa melampaui presisi number
 * JavaScript, jadi nilainya dikonversi hanya saat hendak ditampilkan.
 */
export function toNumber(value: string | number | null | undefined): number {
  if (value === null || value === undefined || value === "") return 0;
  const parsed = typeof value === "number" ? value : Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

const rupiahFormatter = new Intl.NumberFormat("id-ID", {
  style: "currency",
  currency: "IDR",
  minimumFractionDigits: 0,
  maximumFractionDigits: 0,
});

/** Format rupiah penuh: 1250000 -> "Rp1.250.000". */
export function formatIDR(value: string | number | null | undefined): string {
  return rupiahFormatter.format(Math.round(toNumber(value)));
}

/**
 * Format nominal dalam mata uang trip, mis. 4500 JPY -> "JPY 4.500".
 * Dipakai untuk menampilkan kembali harga dalam mata uang negara belanja
 * ketika admin ingin mencocokkan angka dengan nota dari toko di sana.
 */
export function formatForeign(
  value: string | number | null | undefined,
  currency: string,
): string {
  const amount = toNumber(value);
  // Mata uang bernilai besar per unit (mis. dolar) perlu sen agar tidak
  // terlihat dibulatkan habis; nol tetap ditulis bulat supaya tidak berisik.
  const abs = Math.abs(amount);
  const fractionDigits = abs > 0 && abs < 100 ? 2 : 0;
  return `${currency} ${new Intl.NumberFormat("id-ID", {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(amount)}`;
}

/**
 * Mengubah nominal rupiah kembali ke mata uang trip memakai kurs yang dikunci
 * saat trip dibuat. Kurs nol/kosong berarti order tanpa trip valid, jadi
 * nilainya dikembalikan apa adanya agar tidak muncul Infinity di layar.
 */
export function idrToForeign(
  value: string | number | null | undefined,
  exchangeRate: string | number | null | undefined,
): number {
  const rate = toNumber(exchangeRate);
  if (rate <= 0) return toNumber(value);
  return toNumber(value) / rate;
}

/** Format ringkas untuk kartu statistik: 1250000 -> "Rp1,3 jt". */
export function formatIDRCompact(value: string | number | null | undefined): string {
  const amount = toNumber(value);
  const abs = Math.abs(amount);
  const sign = amount < 0 ? "-" : "";

  if (abs >= 1_000_000_000) return `${sign}Rp${(abs / 1_000_000_000).toFixed(1).replace(".", ",")} m`;
  if (abs >= 1_000_000) return `${sign}Rp${(abs / 1_000_000).toFixed(1).replace(".", ",")} jt`;
  if (abs >= 1_000) return `${sign}Rp${(abs / 1_000).toFixed(0)} rb`;
  return formatIDR(amount);
}

/** Format angka biasa dengan pemisah ribuan Indonesia. */
export function formatNumber(value: string | number | null | undefined): string {
  return new Intl.NumberFormat("id-ID").format(toNumber(value));
}

const dateFormatter = new Intl.DateTimeFormat("id-ID", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});

const dateTimeFormatter = new Intl.DateTimeFormat("id-ID", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

export function formatDate(value: string | Date | null | undefined): string {
  if (!value) return "-";
  const date = typeof value === "string" ? new Date(value) : value;
  if (Number.isNaN(date.getTime())) return "-";
  return dateFormatter.format(date);
}

export function formatDateTime(value: string | Date | null | undefined): string {
  if (!value) return "-";
  const date = typeof value === "string" ? new Date(value) : value;
  if (Number.isNaN(date.getTime())) return "-";
  return dateTimeFormatter.format(date);
}

/** Ubah tanggal menjadi "YYYY-MM-DD" untuk input type=date dan payload API. */
export function toDateInput(value: string | Date | null | undefined): string {
  if (!value) return "";
  const date = typeof value === "string" ? new Date(value) : value;
  if (Number.isNaN(date.getTime())) return "";

  // Pakai komponen waktu lokal, bukan toISOString(): konversi ke UTC bisa
  // menggeser tanggal satu hari untuk zona waktu Indonesia.
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}

export function todayInput(): string {
  return toDateInput(new Date());
}

/** Jarak hari dari hari ini, dipakai menandai tagihan yang mulai menua. */
export function daysAgo(value: string | Date | null | undefined): number {
  if (!value) return 0;
  const date = typeof value === "string" ? new Date(value) : value;
  if (Number.isNaN(date.getTime())) return 0;
  return Math.floor((Date.now() - date.getTime()) / 86_400_000);
}

/** Inisial untuk avatar sederhana pada daftar customer. */
export function initials(name: string): string {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}
