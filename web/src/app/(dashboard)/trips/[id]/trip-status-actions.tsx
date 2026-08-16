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
  open: ["closed"],
  closed: ["open"],
};

/** Penjelasan dampak perpindahan status, ditampilkan pada dialog konfirmasi. */
const STATUS_EFFECT: Record<TripStatus, string> = {
  open: "Trip kembali menerima order. Katalognya bisa dipakai lagi saat mencatat pesanan baru.",
  closed:
    "Order baru untuk trip ini ditutup. Order yang sudah masuk tidak terpengaruh dan tetap diproses seperti biasa — termasuk belanja, pengemasan, dan pengiriman.",
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
          // Menutup dan membuka order sama-sama bisa diurungkan, jadi tidak
          // perlu ditampilkan sebagai aksi merah.
          destructive={false}
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
