"use client";

import { useState } from "react";
import { toast } from "sonner";

import { FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { ErrorState } from "@/components/ui/page";
import { useShipOrder } from "@/hooks/use-operations";
import { ApiError } from "@/lib/api";
import { formatIDR, todayInput, toNumber } from "@/lib/utils";
import type { ShippingQueueItem } from "@/types/api";

/**
 * Mencatat nomor resi dan menyerahkan paket ke kurir.
 *
 * Ongkir yang diisi di sini adalah yang benar-benar dibayar ke kurir di konter,
 * bukan yang ditagihkan ke customer. Keduanya sengaja dipisah: toko boleh
 * menanggung selisihnya, dan laporan laba memakai angka yang nyata.
 */
export function DialogResi({
  item,
  onClose,
}: {
  item: ShippingQueueItem;
  onClose: () => void;
}) {
  const ship = useShipOrder(item.order_id);
  const [form, setForm] = useState({
    tracking_number: item.tracking_number ?? "",
    // Bawaannya ongkir yang ditagihkan — paling sering sama, dan admin tinggal
    // membetulkan kalau timbangan konter berbeda.
    shipping_cost: String(toNumber(item.shipping_fee)),
    shipped_at: todayInput(),
  });

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    ship.mutate(
      {
        tracking_number: form.tracking_number,
        shipping_cost: form.shipping_cost || "0",
        shipped_at: form.shipped_at || undefined,
      },
      {
        onSuccess: () => {
          toast.success("Paket ditandai sudah diserahkan ke kurir");
          onClose();
        },
        onError: (err) => {
          toast.error(err instanceof ApiError ? err.message : "Gagal menyimpan nomor resi");
        },
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={`Serahkan ${item.order_number} ke kurir`}
      description={`${item.courier ?? "Kurir"} ${item.service ?? ""} · ditagihkan ${formatIDR(item.shipping_fee)}`}
      onSubmit={handleSubmit}
      loading={ship.isPending}
      submitLabel="Simpan resi"
    >
      <ErrorState error={ship.error} />

      <Field label="Nomor resi" htmlFor="resi_nomor" required>
        <Input
          id="resi_nomor"
          value={form.tracking_number}
          onChange={(event) => setForm({ ...form, tracking_number: event.target.value })}
          placeholder="JNE0012345678"
          required
        />
      </Field>

      <Field
        label="Ongkir dibayar ke kurir (Rp)"
        htmlFor="resi_ongkir"
        hint="Angka di struk konter. Boleh berbeda dari yang ditagihkan ke customer."
      >
        <Input
          id="resi_ongkir"
          type="number"
          min="0"
          step="any"
          value={form.shipping_cost}
          onChange={(event) => setForm({ ...form, shipping_cost: event.target.value })}
        />
      </Field>

      <Field label="Tanggal diserahkan" htmlFor="resi_tanggal">
        <Input
          id="resi_tanggal"
          type="date"
          value={form.shipped_at}
          onChange={(event) => setForm({ ...form, shipped_at: event.target.value })}
        />
      </Field>
    </FormDialog>
  );
}
