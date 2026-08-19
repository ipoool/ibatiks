/**
 * Mata uang yang bisa dipilih saat menyusun trip atau mendata produk.
 *
 * Ditulis sekali di sini dan dipakai kedua form, supaya keduanya tidak pernah
 * menawarkan daftar yang berbeda tanpa disadari.
 *
 * Urutannya bukan alfabet melainkan kedekatan: ASEAN lebih dulu karena itu
 * tetangga terdekat dan tujuan jastip yang paling sering, baru tujuan lain yang
 * lazim. Yang paling sering dipakai harus paling sedikit digulir.
 */
export interface MataUang {
  kode: string;
  label: string;
}

/** Negara ASEAN, diawali Rupiah karena itu mata uang tokonya sendiri. */
export const MATA_UANG_ASEAN: readonly MataUang[] = [
  { kode: "IDR", label: "IDR — Rupiah Indonesia" },
  { kode: "SGD", label: "SGD — Dolar Singapura" },
  { kode: "MYR", label: "MYR — Ringgit Malaysia" },
  { kode: "THB", label: "THB — Baht Thailand" },
  { kode: "VND", label: "VND — Dong Vietnam" },
  { kode: "PHP", label: "PHP — Peso Filipina" },
  { kode: "BND", label: "BND — Dolar Brunei" },
  { kode: "KHR", label: "KHR — Riel Kamboja" },
  { kode: "LAK", label: "LAK — Kip Laos" },
  { kode: "MMK", label: "MMK — Kyat Myanmar" },
];

/** Tujuan jastip lain yang lazim di luar ASEAN. */
export const MATA_UANG_LAIN: readonly MataUang[] = [
  { kode: "JPY", label: "JPY — Yen Jepang" },
  { kode: "KRW", label: "KRW — Won Korea Selatan" },
  { kode: "CNY", label: "CNY — Yuan Tiongkok" },
  { kode: "HKD", label: "HKD — Dolar Hong Kong" },
  { kode: "TWD", label: "TWD — Dolar Taiwan" },
  { kode: "AUD", label: "AUD — Dolar Australia" },
  { kode: "USD", label: "USD — Dolar Amerika Serikat" },
  { kode: "EUR", label: "EUR — Euro" },
  { kode: "GBP", label: "GBP — Pound Inggris" },
  { kode: "SAR", label: "SAR — Riyal Arab Saudi" },
  { kode: "AED", label: "AED — Dirham Uni Emirat Arab" },
];

const SEMUA = [...MATA_UANG_ASEAN, ...MATA_UANG_LAIN];

/** Label lengkap sebuah kode, atau kodenya sendiri kalau tidak dikenali. */
export function labelMataUang(kode: string): string {
  return SEMUA.find((m) => m.kode === kode)?.label ?? kode;
}

/** Benar bila kode itu ada di daftar pilihan. */
export function mataUangDikenal(kode: string): boolean {
  return SEMUA.some((m) => m.kode === kode);
}
