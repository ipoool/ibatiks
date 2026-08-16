import { cn, formatIDR, toNumber } from "@/lib/utils";
import type { Money, OrderStatus } from "@/types/api";

/**
 * Menampilkan sisa tagihan sebuah order.
 *
 * Ada dua keadaan yang, kalau dicetak apa adanya, membuat admin salah baca:
 *
 * 1. Customer membayar lebih — sering terjadi karena kode unik transfer atau
 *    pembulatan. Angka sisa tagihannya menjadi negatif, dan "-Rp500.000" di
 *    kolom tagihan terbaca seperti kesalahan sistem. Yang benar: tagihannya
 *    nol, dan ada kelebihan yang harus dikembalikan ke customer.
 *
 * 2. Order dibatalkan — sisa tagihannya secara hitungan masih ada, tapi tidak
 *    akan pernah ditagih. Menampilkannya berwarna oranye seperti piutang lain
 *    membuat order itu ikut terkejar-kejar di layar kerja.
 */
export function BalanceDue({
  amount,
  status,
  className,
  align = "right",
}: {
  amount: Money;
  status?: OrderStatus;
  className?: string;
  /** Perataan baris keterangan di bawah angkanya. */
  align?: "left" | "right";
}) {
  const value = toNumber(amount);
  const cancelled = status === "cancelled";
  const catatan = align === "right" ? "text-right" : "text-left";

  if (cancelled) {
    return (
      <span className={cn("block text-muted-foreground", className)}>
        {formatIDR(0)}
        <span className={cn("block text-xs font-normal", catatan)}>order dibatalkan</span>
      </span>
    );
  }

  if (value < 0) {
    return (
      <span className={cn("block text-emerald-600", className)}>
        {formatIDR(0)}
        <span className={cn("block text-xs font-normal text-amber-600", catatan)}>
          lebih bayar {formatIDR(Math.abs(value))}
        </span>
      </span>
    );
  }

  return (
    <span className={cn("block", value > 0 ? "text-amber-600" : "text-muted-foreground", className)}>
      {formatIDR(value)}
    </span>
  );
}

/** Benar bila customer membayar melebihi tagihannya. */
export function isOverpaid(amount: Money): boolean {
  return toNumber(amount) < 0;
}
