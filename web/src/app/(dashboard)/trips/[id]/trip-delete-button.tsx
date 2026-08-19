"use client";

import { Loader2, Trash2, TriangleAlert } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { FormDialog } from "@/components/ui/form-dialog";
import { Input } from "@/components/ui/input";
import { ErrorState } from "@/components/ui/page";
import { useDeleteTrip, useTripDeletionImpact } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatIDR, formatNumber, toNumber } from "@/lib/utils";
import type { Trip, TripDeletionImpact } from "@/types/api";

/**
 * Menghapus trip beserta seluruh riwayat di dalamnya.
 *
 * Dampaknya diambil dari server saat dialog dibuka, bukan disimpulkan dari data
 * yang sudah ada di halaman: yang dihapus mencakup order, invoice, pembayaran,
 * pengiriman, dan pembelian, dan admin berhak melihat angkanya sebelum memutuskan.
 */
export function TripDeleteButton({ trip }: { trip: Trip }) {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
        <Trash2 className="text-destructive" />
        Hapus Trip
      </Button>
      {open && <DialogHapusTrip trip={trip} onClose={() => setOpen(false)} />}
    </>
  );
}

function DialogHapusTrip({ trip, onClose }: { trip: Trip; onClose: () => void }) {
  const router = useRouter();
  const { data: impact, isLoading, error } = useTripDeletionImpact(trip.id, true);
  const remove = useDeleteTrip();
  const [konfirmasi, setKonfirmasi] = useState("");

  const terhalang = Boolean(
    (impact?.shipped_orders?.length ?? 0) > 0 || (impact?.stock_on_hand?.length ?? 0) > 0,
  );
  const kosong =
    impact !== undefined &&
    impact.orders === 0 &&
    impact.purchases === 0 &&
    impact.expenses === 0;

  /*
   * Trip yang sudah berisi order menuntut kode tripnya diketik ulang. Menghapus
   * di sini membuang catatan uang yang sudah diterima — satu klik yang keliru
   * pada tombol merah terlalu murah untuk akibat sebesar itu.
   */
  const perluKetik = !kosong && !terhalang;
  const kodeCocok = konfirmasi.trim().toUpperCase() === trip.code.toUpperCase();

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    if (terhalang || (perluKetik && !kodeCocok)) return;

    remove.mutate(trip.id, {
      onSuccess: () => {
        toast.success(`Trip ${trip.code} dihapus`);
        router.push("/trips");
      },
      onError: (err) => {
        toast.error(err instanceof ApiError ? err.message : "Gagal menghapus trip");
      },
    });
  }

  return (
    <FormDialog
      open
      onOpenChange={(o) => !o && onClose()}
      title={`Hapus trip ${trip.code}?`}
      description={trip.title}
      onSubmit={handleSubmit}
      loading={remove.isPending}
      submitLabel="Hapus trip"
      submitDisabled={isLoading || terhalang || (perluKetik && !kodeCocok)}
      destructive
    >
      <ErrorState error={error ?? remove.error} />

      {isLoading && (
        <p className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" />
          Menghitung apa saja yang ikut terhapus…
        </p>
      )}

      {impact && terhalang && <Penghalang impact={impact} />}

      {impact && !terhalang && kosong && (
        <p className="text-sm text-muted-foreground">
          Trip ini belum berisi order, pembelian, maupun biaya perjalanan. Katalognya
          {impact.catalog_items > 0
            ? ` (${formatNumber(impact.catalog_items)} produk) ikut terhapus.`
            : " kosong."}
        </p>
      )}

      {impact && !terhalang && !kosong && (
        <>
          <DaftarDampak impact={impact} />

          <Field
            label={`Ketik ${trip.code} untuk memastikan`}
            htmlFor="konfirmasi_hapus"
            hint="Penghapusan tidak bisa dibatalkan."
          >
            <Input
              id="konfirmasi_hapus"
              value={konfirmasi}
              onChange={(event) => setKonfirmasi(event.target.value)}
              placeholder={trip.code}
              autoComplete="off"
            />
          </Field>
        </>
      )}
    </FormDialog>
  );
}

/** Alasan trip ini tidak boleh dihapus, beserta jalan keluarnya. */
function Penghalang({ impact }: { impact: TripDeletionImpact }) {
  const terkirim = impact.shipped_orders ?? [];

  return (
    <div className="space-y-3 rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm">
      <p className="flex items-center gap-2 font-medium text-destructive">
        <TriangleAlert className="size-4 shrink-0" />
        Trip ini tidak bisa dihapus
      </p>

      {terkirim.length > 0 && (
        <div>
          <p className="font-medium">
            {formatNumber(terkirim.length)} order sudah diserahkan ke kurir
          </p>
          <p className="text-muted-foreground">
            {terkirim.join(", ")} — barangnya sudah di jalan atau sudah diterima customer.
            Penjualan yang sudah jadi tidak bisa dihapus.
          </p>
        </div>
      )}

      {impact.stock_on_hand.length > 0 && (
        <div>
          <p className="font-medium">Barang surplus masih ada di stok</p>
          <ul className="list-disc pl-5 text-muted-foreground">
            {impact.stock_on_hand.map((item) => (
              <li key={item.sku}>
                {item.product_name} — {formatNumber(item.qty)} pcs
              </li>
            ))}
          </ul>
          <p className="text-muted-foreground">
            Barangnya nyata dan masih bisa dijual. Habiskan atau sesuaikan stoknya dulu lewat
            menu Stok, baru tripnya bisa dihapus.
          </p>
        </div>
      )}
    </div>
  );
}

/** Rincian apa saja yang ikut terhapus. */
function DaftarDampak({ impact }: { impact: TripDeletionImpact }) {
  const baris = [
    { label: "Order", nilai: impact.orders },
    { label: "Invoice terbit", nilai: impact.invoices },
    { label: "Paket", nilai: impact.shipments },
    { label: "Pembelian", nilai: impact.purchases },
    { label: "Biaya perjalanan", nilai: impact.expenses },
    { label: "Produk di katalog", nilai: impact.catalog_items },
  ].filter((b) => b.nilai > 0);

  const uangMasuk = toNumber(impact.payments_total);

  return (
    <div className="space-y-3 rounded-lg border border-amber-500/40 bg-amber-500/5 p-4 text-sm">
      <p className="flex items-center gap-2 font-medium text-amber-700">
        <TriangleAlert className="size-4 shrink-0" />
        Yang ikut terhapus bersama trip ini
      </p>

      <ul className="grid gap-x-6 gap-y-1 sm:grid-cols-2">
        {baris.map((b) => (
          <li key={b.label} className="flex justify-between gap-4">
            <span className="text-muted-foreground">{b.label}</span>
            <span className="tabular font-medium">{formatNumber(b.nilai)}</span>
          </li>
        ))}
      </ul>

      {/*
       * Uang yang sudah diterima ditonjolkan terpisah. Inilah akibat yang paling
       * mudah terlewat: nominalnya hilang dari pembukuan dan laporan, sementara
       * uangnya tetap ada di rekening toko.
       */}
      {uangMasuk > 0 && (
        <p className="border-t border-amber-500/30 pt-3">
          <span className="font-medium">{formatIDR(impact.payments_total)}</span> pembayaran yang
          sudah diterima ikut hilang dari pembukuan dan laporan. Uangnya tetap ada di rekening,
          tapi aplikasi tidak lagi mengingat dari mana asalnya.
        </p>
      )}
    </div>
  );
}
