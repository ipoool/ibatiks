"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { DetailRow } from "@/components/ui/page";
import { formatDateTime, formatIDR, formatNumber, toNumber } from "@/lib/utils";
import type { ShippingQueueItem } from "@/types/api";

/**
 * Data kemasan yang sudah tersimpan, hanya untuk dibaca.
 *
 * Order yang sudah diserahkan ke kurir tidak lagi menampilkan tombol Kemas —
 * barangnya sudah di jalan dan mengubah beratnya tidak mengubah apa pun kecuali
 * catatan toko sendiri. Tapi angkanya masih sering ditanya: berapa beratnya,
 * layanan apa yang dipakai, ongkirnya berapa, ada asuransinya atau tidak.
 * Sebelumnya jawabannya cuma ada di database.
 */
export function DialogDetailKemasan({
  item,
  onClose,
}: {
  item: ShippingQueueItem;
  onClose: () => void;
}) {
  const dimensi = [item.length_cm, item.width_cm, item.height_cm].filter(
    (sisi) => (sisi ?? 0) > 0,
  );
  const premi = toNumber(item.insurance_fee);
  const ongkirDitagih = toNumber(item.shipping_fee);

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Data kemasan {item.order_number}</DialogTitle>
          <DialogDescription>
            {item.recipient_name} · {item.shipping_city}
          </DialogDescription>
        </DialogHeader>

        <div className="divide-y divide-border">
          <DetailRow
            label="Berat"
            value={
              (item.weight_gram ?? 0) > 0
                ? `${formatNumber((item.weight_gram ?? 0) / 1000)} kg`
                : "—"
            }
          />
          <DetailRow
            label="Dimensi kardus"
            value={dimensi.length === 3 ? `${dimensi.join(" × ")} cm` : "tidak dicatat"}
          />
          <DetailRow
            label="Kurir"
            value={[item.courier, item.service].filter(Boolean).join(" ") || "—"}
          />
          {/* Ongkir dan preminya dipisah di sini walau ditagihkan sebagai satu
              angka — pertanyaan "berapa yang asuransi" justru muncul setelah
              paketnya berangkat. */}
          <DetailRow
            label="Ongkir ditagihkan"
            value={premi > 0 ? formatIDR(ongkirDitagih - premi) : formatIDR(ongkirDitagih)}
          />
          {premi > 0 && <DetailRow label="Premi asuransi" value={formatIDR(premi)} />}
          {premi > 0 && (
            <DetailRow
              label="Total ditagihkan"
              value={<span className="font-semibold">{formatIDR(ongkirDitagih)}</span>}
            />
          )}
          <DetailRow
            label="Ongkir dibayar ke kurir"
            value={toNumber(item.shipping_cost) > 0 ? formatIDR(item.shipping_cost) : "—"}
          />
          <DetailRow label="Nomor resi" value={item.tracking_number || "—"} />
          <DetailRow
            label="Diserahkan ke kurir"
            value={item.shipped_at ? formatDateTime(item.shipped_at) : "—"}
          />
        </div>

        {item.shipment_notes && (
          <div className="rounded-lg border border-border bg-muted/40 p-3">
            <p className="text-xs text-muted-foreground">Catatan kemasan</p>
            <p className="mt-0.5 text-sm">{item.shipment_notes}</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
