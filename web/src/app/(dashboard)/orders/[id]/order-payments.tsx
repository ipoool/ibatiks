"use client";

import { MessageCircle, Paperclip, Plus, Trash2, Wallet } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { OptionSelect, toOptions } from "@/components/filter-select";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ConfirmDialog, FormDialog } from "@/components/ui/form-dialog";
import { FileUpload } from "@/components/ui/file-upload";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { WAMessageDialog } from "@/components/wa-message-dialog";
import {
  useDPMessage,
  useDeletePayment,
  useRecordPayment,
  type PaymentPayload,
} from "@/hooks/use-orders";
import { ApiError } from "@/lib/api";
import { formatDate, formatIDR, todayInput, toNumber } from "@/lib/utils";
import type { OrderDetail, Payment, PaymentMethod, PaymentType } from "@/types/api";

const TYPE_LABEL: Record<PaymentType, string> = {
  dp: "DP",
  settlement: "Pelunasan",
  refund: "Refund",
  adjustment: "Penyesuaian",
};

const METHOD_LABEL: Record<PaymentMethod, string> = {
  transfer: "Transfer bank",
  cash: "Tunai",
  qris: "QRIS",
  ewallet: "E-wallet",
  lainnya: "Lainnya",
};

export function OrderPayments({ order }: { order: OrderDetail }) {
  const [formOpen, setFormOpen] = useState(false);
  const [deleting, setDeleting] = useState<Payment | null>(null);
  const [dpMessageOpen, setDpMessageOpen] = useState(false);

  const balanceDue = toNumber(order.balance_due);
  const dpPaid = order.payments
    .filter((payment) => payment.type === "dp")
    .reduce((sum, payment) => sum + toNumber(payment.amount), 0);
  const dpOutstanding = Math.max(toNumber(order.dp_required) - dpPaid, 0);

  // Jenis pembayaran default mengikuti tahap order: selama DP belum lunas,
  // yang paling mungkin dicatat adalah DP.
  const [form, setForm] = useState<PaymentPayload>({
    type: dpOutstanding > 0 ? "dp" : "settlement",
    amount: "",
    method: "transfer",
    reference: "",
    paid_at: todayInput(),
    notes: "",
  });

  const record = useRecordPayment(order.id);
  const removePayment = useDeletePayment(order.id);
  const dpMessage = useDPMessage(order.id, dpMessageOpen);

  function openForm(type: PaymentType, amount: number) {
    setForm({
      type,
      amount: amount > 0 ? String(Math.round(amount)) : "",
      method: "transfer",
      reference: "",
      paid_at: todayInput(),
      notes: "",
    });
    record.reset();
    setFormOpen(true);
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    record.mutate(
      { ...form, reference: form.reference || null, notes: form.notes || null },
      {
        onSuccess: (updated) => {
          toast.success("Pembayaran dicatat", {
            description:
              toNumber(updated.balance_due) <= 0
                ? "Order sudah lunas dan siap dikirim."
                : `Sisa tagihan ${formatIDR(updated.balance_due)}.`,
          });
          setFormOpen(false);
        },
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal mencatat pembayaran");
        },
      },
    );
  }

  function handleDelete() {
    if (!deleting) return;
    removePayment.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Pembayaran dihapus");
        setDeleting(null);
      },
    });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Pembayaran</CardTitle>
        <CardAction className="flex gap-2">
          {dpOutstanding > 0 && (
            <Button size="sm" variant="outline" onClick={() => setDpMessageOpen(true)}>
              <MessageCircle />
              Tagih DP
            </Button>
          )}
          <Button
            size="sm"
            onClick={() => openForm(dpOutstanding > 0 ? "dp" : "settlement", dpOutstanding > 0 ? dpOutstanding : balanceDue)}
          >
            <Plus />
            Catat Bayar
          </Button>
        </CardAction>
      </CardHeader>

      <CardContent className="space-y-4">
        <div className="grid gap-3 sm:grid-cols-3">
          <div className="rounded-lg border border-border p-3">
            <p className="text-xs text-muted-foreground">DP diminta</p>
            <p className="tabular font-semibold">{formatIDR(order.dp_required)}</p>
            {dpOutstanding > 0 && (
              <p className="text-xs text-amber-600">Kurang {formatIDR(dpOutstanding)}</p>
            )}
          </div>
          <div className="rounded-lg border border-border p-3">
            <p className="text-xs text-muted-foreground">Sudah dibayar</p>
            <p className="tabular font-semibold text-emerald-600">{formatIDR(order.paid_amount)}</p>
          </div>
          <div className="rounded-lg border border-border p-3">
            <p className="text-xs text-muted-foreground">Sisa tagihan</p>
            <p
              className={`tabular font-semibold ${balanceDue > 0 ? "text-amber-600" : "text-emerald-600"}`}
            >
              {formatIDR(order.balance_due)}
            </p>
          </div>
        </div>

        {order.payments.length === 0 ? (
          <div className="flex flex-col items-center gap-2 rounded-lg border border-dashed border-border py-8 text-center">
            <Wallet className="size-6 text-muted-foreground/60" />
            <p className="text-sm text-muted-foreground">Belum ada pembayaran tercatat.</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {order.payments.map((payment) => (
              <div key={payment.id} className="flex items-center justify-between gap-3 py-2.5">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant={payment.type === "refund" ? "danger" : "success"}>
                      {TYPE_LABEL[payment.type]}
                    </Badge>
                    <span className="text-sm text-muted-foreground">
                      {METHOD_LABEL[payment.method]} · {formatDate(payment.paid_at)}
                    </span>
                  </div>
                  {payment.reference && (
                    <p className="truncate text-xs text-muted-foreground">Ref: {payment.reference}</p>
                  )}
                  {/* Bukti transfer tetap bisa dibuka setelah pembayaran
                      tercatat, supaya bisa dicocokkan ulang dengan mutasi
                      rekening kalau ada selisih. */}
                  {payment.proof_url && (
                    <a
                      href={payment.proof_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                    >
                      <Paperclip className="size-3" />
                      Lihat bukti transfer
                    </a>
                  )}
                  {payment.notes && (
                    <p className="truncate text-xs text-muted-foreground">{payment.notes}</p>
                  )}
                </div>

                <div className="flex items-center gap-2">
                  <span
                    className={`tabular font-semibold ${
                      payment.type === "refund" ? "text-destructive" : ""
                    }`}
                  >
                    {payment.type === "refund" ? "−" : ""}
                    {formatIDR(payment.amount)}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    className="text-destructive hover:text-destructive"
                    onClick={() => setDeleting(payment)}
                  >
                    <Trash2 />
                    <span className="sr-only">Hapus</span>
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>

      <FormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        title="Catat Pembayaran"
        description="Status order otomatis menyesuaikan setelah uang tercatat masuk."
        error={record.error}
        loading={record.isPending}
        onSubmit={handleSubmit}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Jenis" htmlFor="payment_type" required>
            <OptionSelect
              id="payment_type"
              value={form.type}
              onChange={(value) => setForm({ ...form, type: value })}
              options={toOptions(TYPE_LABEL)}
            />
          </Field>

          <Field label="Nominal (Rp)" htmlFor="amount" required>
            <Input
              id="amount"
              type="number"
              min="1"
              step="any"
              value={form.amount}
              onChange={(event) => setForm({ ...form, amount: event.target.value })}
              required
              autoFocus
            />
          </Field>

          <Field label="Metode" htmlFor="method" required>
            <OptionSelect
              id="method"
              value={form.method}
              onChange={(value) => setForm({ ...form, method: value })}
              options={toOptions(METHOD_LABEL)}
            />
          </Field>

          <Field label="Tanggal bayar" htmlFor="paid_at">
            <Input
              id="paid_at"
              type="date"
              value={form.paid_at ?? ""}
              onChange={(event) => setForm({ ...form, paid_at: event.target.value })}
            />
          </Field>

          <Field label="Referensi" htmlFor="reference" hint="Nomor transaksi atau nama pengirim">
            <Input
              id="reference"
              value={form.reference ?? ""}
              onChange={(event) => setForm({ ...form, reference: event.target.value })}
            />
          </Field>

          <Field
            label="Bukti transfer"
            htmlFor="proof_url"
            hint="Foto atau PDF struk transfer dari customer"
          >
            <FileUpload
              value={form.proof_url ?? null}
              onChange={(url) => setForm({ ...form, proof_url: url })}
            />
          </Field>

          <Field label="Catatan" htmlFor="payment_notes" className="sm:col-span-2">
            <Textarea
              id="payment_notes"
              rows={2}
              value={form.notes ?? ""}
              onChange={(event) => setForm({ ...form, notes: event.target.value })}
            />
          </Field>
        </div>
      </FormDialog>

      <WAMessageDialog
        open={dpMessageOpen}
        onOpenChange={setDpMessageOpen}
        title="Tagih DP ke customer"
        description="Teks sudah terisi lengkap. Tinggal buka WhatsApp lalu tekan kirim."
        message={dpMessage.data}
        isLoading={dpMessage.isLoading}
        error={dpMessage.error}
      />

      <ConfirmDialog
        open={Boolean(deleting)}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Hapus catatan pembayaran?"
        description={`Pembayaran ${formatIDR(deleting?.amount ?? 0)} akan dihapus dan sisa tagihan dihitung ulang.`}
        confirmLabel="Hapus"
        loading={removePayment.isPending}
        error={removePayment.error}
        onConfirm={handleDelete}
      />
    </Card>
  );
}
