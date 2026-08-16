"use client";

import { RefreshCw } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { ConfirmDialog } from "@/components/ui/form-dialog";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useSyncExchangeRate } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatNumber } from "@/lib/utils";
import type { Trip } from "@/types/api";

/**
 * Menyegarkan kurs sebuah trip dari kurs harian.
 *
 * Kurs trip dikunci saat trip dibuat karena seluruh harga jual diturunkan
 * darinya. Menyegarkannya adalah keputusan sadar, jadi selalu lewat dialog
 * konfirmasi — lengkap dengan peringatan bahwa menghitung ulang katalog berarti
 * harga yang sudah dilihat customer ikut berubah.
 */
export function SyncRateButton({ trip }: { trip: Trip }) {
  const [open, setOpen] = useState(false);
  const [recalculate, setRecalculate] = useState(false);
  const sync = useSyncExchangeRate(trip.id);

  async function handleConfirm() {
    const result = await sync.mutateAsync(recalculate);

    const changed = result.previous_rate !== result.new_rate;
    toast.success(
      changed
        ? `Kurs ${trip.currency} diperbarui: ${formatNumber(result.previous_rate)} → ${formatNumber(result.new_rate)}`
        : `Kurs ${trip.currency} sudah sesuai dengan kurs hari ini`,
      {
        description:
          result.items_updated > 0
            ? `${result.items_updated} harga katalog ikut dihitung ulang.`
            : "Harga katalog tidak diubah.",
      },
    );
    // ConfirmDialog tidak menutup dirinya sendiri: yang tahu aksinya berhasil
    // hanya pemanggilnya.
    setRecalculate(false);
    setOpen(false);
  }

  // Rupiah tidak punya kurs terhadap dirinya sendiri.
  if (trip.currency === "IDR") return null;

  return (
    <>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setOpen(true)}
            aria-label={`Segarkan kurs ${trip.currency}`}
          >
            <RefreshCw />
          </Button>
        </TooltipTrigger>
        <TooltipContent>Segarkan kurs dari kurs harian</TooltipContent>
      </Tooltip>

      <ConfirmDialog
        open={open}
        onOpenChange={(next) => {
          if (!next) sync.reset();
          setOpen(next);
        }}
        title={`Segarkan kurs ${trip.currency}?`}
        description={`Kurs trip ${trip.code} sekarang 1 ${trip.currency} = Rp${formatNumber(trip.exchange_rate)}. Kurs akan diambil ulang dari sumber kurs harian.`}
        confirmLabel="Ya, segarkan kurs"
        destructive={false}
        error={sync.error instanceof ApiError ? sync.error : sync.error}
        loading={sync.isPending}
        onConfirm={handleConfirm}
      >
        <div className="space-y-3">
          <CheckboxField
            id={`recalc_${trip.id}`}
            checked={recalculate}
            onCheckedChange={setRecalculate}
          >
            Hitung ulang harga jual katalog dengan kurs baru
          </CheckboxField>

          {recalculate && (
            <div className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
              <p className="font-medium">Akan mempengaruhi harga jual ke customer.</p>
              <p className="mt-1">
                Seluruh harga katalog trip ini dihitung ulang. Customer yang sudah melihat harga
                lama akan menemukan angka yang berbeda. Order yang terlanjur dibuat tetap memakai
                harga lamanya, karena harga disalin saat order dicatat.
              </p>
            </div>
          )}
        </div>
      </ConfirmDialog>
    </>
  );
}
