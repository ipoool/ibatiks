"use client";

import { useState } from "react";
import { toast } from "sonner";

import { Field } from "@/components/ui/field";
import { FormDialog } from "@/components/ui/form-dialog";
import { Input } from "@/components/ui/input";
import { ErrorState } from "@/components/ui/page";
import { useSettleInvoice } from "@/hooks/use-operations";
import { ApiError } from "@/lib/api";
import { formatIDR, todayInput, toNumber } from "@/lib/utils";
import type { Invoice } from "@/types/api";

/**
 * Menandai invoice sudah dibayar customer.
 *
 * Yang dicatat adalah pembayarannya, bukan sekadar label pada barisnya. Saldo
 * order, status order, dan laporan piutang semuanya dihitung dari tabel
 * pembayaran — menandai dokumennya lunas tanpa mencatat uangnya akan membuat
 * ketiganya berbohong, dan order tetap terlihat menunggak selamanya.
 */
export function DialogTandaiLunas({
  invoice,
  onClose,
}: {
  invoice: Invoice;
  onClose: () => void;
}) {
  const settle = useSettleInvoice();
  const [form, setForm] = useState({
    // Bawaannya sisa tagihan invoice ini — yang paling sering terjadi adalah
    // customer membayar persis sejumlah yang ditagih.
    amount: String(toNumber(invoice.amount_due)),
    paid_at: todayInput(),
    reference: "",
  });

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    settle.mutate(
      {
        orderId: invoice.order_id,
        amount: form.amount || "0",
        method: "transfer",
        paid_at: form.paid_at || undefined,
        reference: form.reference || null,
      },
      {
        onSuccess: () => {
          toast.success("Pembayaran tercatat");
          onClose();
        },
        onError: (err) => {
          toast.error(err instanceof ApiError ? err.message : "Gagal mencatat pembayaran");
        },
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={`Catat pembayaran ${invoice.invoice_number}`}
      description={`${invoice.customer_name} · ditagih ${formatIDR(invoice.amount_due)}`}
      onSubmit={handleSubmit}
      loading={settle.isPending}
      submitLabel="Catat pembayaran"
    >
      <ErrorState error={settle.error} />

      <Field
        label="Nominal diterima (Rp)"
        htmlFor="lunas_nominal"
        required
        hint="Boleh kurang dari yang ditagih; sisanya tetap tercatat sebagai piutang."
      >
        <Input
          id="lunas_nominal"
          type="number"
          min="0"
          step="any"
          value={form.amount}
          onChange={(event) => setForm({ ...form, amount: event.target.value })}
          required
        />
      </Field>

      <Field label="Tanggal masuk" htmlFor="lunas_tanggal">
        <Input
          id="lunas_tanggal"
          type="date"
          value={form.paid_at}
          onChange={(event) => setForm({ ...form, paid_at: event.target.value })}
        />
      </Field>

      <Field
        label="Keterangan transfer"
        htmlFor="lunas_referensi"
        hint="Nama pengirim atau berita transfer, untuk mencocokkan dengan mutasi rekening."
      >
        <Input
          id="lunas_referensi"
          value={form.reference}
          onChange={(event) => setForm({ ...form, reference: event.target.value })}
          placeholder="BCA a/n Rina Kartika"
        />
      </Field>
    </FormDialog>
  );
}
