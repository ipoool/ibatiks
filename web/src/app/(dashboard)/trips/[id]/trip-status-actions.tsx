"use client";

import { ArrowRight } from "lucide-react";
import { toast } from "sonner";

import { tripStatusLabel } from "@/components/status-badge";
import { ConfirmButton } from "@/components/ui/confirm-button";
import { useChangeTripStatus } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import type { Trip, TripStatus } from "@/types/api";

/**
 * Status berikutnya yang sah untuk tiap status trip. Daftar ini sengaja
 * mencerminkan peta transisi di backend supaya tombol yang tampil pasti
 * diterima server, bukan menampilkan pilihan yang berujung penolakan.
 */
const NEXT_STATUS: Record<TripStatus, TripStatus[]> = {
  draft: ["open"],
  open: ["closed", "shopping"],
  closed: ["shopping", "open"],
  shopping: ["in_transit", "arrived"],
  in_transit: ["arrived"],
  arrived: ["settled"],
  settled: [],
  cancelled: [],
};

/**
 * Penjelasan dampak tiap perpindahan status trip, ditampilkan pada dialog
 * konfirmasi. Dua di antaranya mengubah status banyak order sekaligus, dan
 * itulah alasan utama perpindahan status trip perlu dikonfirmasi.
 */
const STATUS_EFFECT: Record<TripStatus, string> = {
  draft: "Trip dikembalikan ke draft.",
  open: "Trip dibuka untuk menerima order. Katalognya sudah bisa dipakai saat mencatat pesanan.",
  closed:
    "Order baru untuk trip ini ditutup. Order yang sudah masuk tidak terpengaruh dan tetap diproses seperti biasa.",
  shopping:
    "Trip ditandai sedang belanja. Semua order pada trip ini yang DP-nya sudah masuk otomatis berpindah ke tahap Dibelikan sekaligus.",
  in_transit: "Trip ditandai dalam perjalanan pulang. Belanja dianggap sudah selesai.",
  arrived:
    "Barang dianggap sudah sampai di Indonesia. Semua order pada trip ini yang sedang dibelikan otomatis berpindah ke Barang Tiba sekaligus, siap dicocokkan.",
  settled:
    "Trip ditutup dan dibukukan. Tidak ada perpindahan status lagi setelah ini, jadi pastikan seluruh belanja dan biaya perjalanan sudah tercatat.",
  cancelled: "Trip dibatalkan.",
};

export function TripStatusActions({ trip }: { trip: Trip }) {
  const changeStatus = useChangeTripStatus(trip.id);
  const nextStatuses = NEXT_STATUS[trip.status] ?? [];

  if (nextStatuses.length === 0) return null;

  // mutateAsync dipakai supaya dialog konfirmasi bisa menunggu hasilnya:
  // ditutup kalau berhasil, tetap terbuka beserta pesan galat kalau gagal.
  async function handleChange(status: TripStatus) {
    await changeStatus.mutateAsync(status, {
      onSuccess: () => {
        toast.success(`Trip berpindah ke "${tripStatusLabel(status)}"`);
      },
      onError: (error) => {
        toast.error(error instanceof ApiError ? error.message : "Gagal mengubah status trip");
      },
    });
  }

  return (
    <div className="flex flex-wrap gap-2">
      {nextStatuses.map((status, index) => (
        <ConfirmButton
          key={status}
          // Pilihan pertama adalah langkah paling lazim, jadi ditonjolkan.
          variant={index === 0 ? "default" : "outline"}
          title={`Pindahkan trip ke "${tripStatusLabel(status)}"?`}
          description={STATUS_EFFECT[status]}
          confirmLabel={`Ya, ${tripStatusLabel(status).toLowerCase()}`}
          // Pembukuan trip tidak bisa diurungkan.
          destructive={status === "settled"}
          error={changeStatus.error}
          onConfirm={() => handleChange(status)}
        >
          <ArrowRight />
          {tripStatusLabel(status)}
        </ConfirmButton>
      ))}
    </div>
  );
}
