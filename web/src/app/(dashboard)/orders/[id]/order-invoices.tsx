"use client";

import { Ban, FileText, MessageCircle, Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { OptionSelect } from "@/components/filter-select";
import { InvoiceStatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ConfirmButton } from "@/components/ui/confirm-button";
import { FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { WAMessageDialog } from "@/components/wa-message-dialog";
import { useCreateInvoice } from "@/hooks/use-orders";
import {
  invoicePDFUrl,
  useInvoiceMessage,
  useMarkInvoiceSent,
  useVoidInvoice,
} from "@/hooks/use-operations";
import { ApiError } from "@/lib/api";
import { formatDate, formatIDR } from "@/lib/utils";
import type { Invoice, InvoiceType, OrderDetail, SentChannel } from "@/types/api";

const INVOICE_TYPE_OPTIONS: ReadonlyArray<{ value: InvoiceType; label: string }> = [
  { value: "final", label: "Pelunasan (seluruh nilai order)" },
  { value: "dp", label: "DP (hanya uang muka)" },
];

export function OrderInvoices({ order }: { order: OrderDetail }) {
  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<{ type: InvoiceType; due_date: string; notes: string }>({
    type: "final",
    due_date: "",
    notes: "",
  });
  const [messageInvoiceId, setMessageInvoiceId] = useState<string | null>(null);

  const create = useCreateInvoice(order.id);
  const markSent = useMarkInvoiceSent();
  const voidInvoice = useVoidInvoice();
  const message = useInvoiceMessage(messageInvoiceId ?? undefined, Boolean(messageInvoiceId));

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    create.mutate(
      {
        type: form.type,
        due_date: form.due_date || undefined,
        notes: form.notes || null,
      },
      {
        onSuccess: (invoice) => {
          toast.success(`Invoice ${invoice.invoice_number} diterbitkan`);
          setFormOpen(false);
          // Langsung tawarkan pengirimannya, karena itu langkah berikutnya.
          setMessageInvoiceId(invoice.id);
        },
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal menerbitkan invoice");
        },
      },
    );
  }

  async function handleVoid(invoice: Invoice) {
    await voidInvoice.mutateAsync(invoice.id, {
      onSuccess: () => toast.success(`Invoice ${invoice.invoice_number} dibatalkan`),
      onError: (error) => {
        toast.error(error instanceof ApiError ? error.message : "Gagal membatalkan invoice");
      },
    });
  }

  function handleSent(channel: SentChannel) {
    if (!messageInvoiceId) return;
    markSent.mutate(
      { id: messageInvoiceId, channel },
      { onSuccess: () => toast.success("Invoice ditandai sudah dikirim") },
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Invoice</CardTitle>
        <CardAction>
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              create.reset();
              setFormOpen(true);
            }}
          >
            <Plus />
            Terbitkan
          </Button>
        </CardAction>
      </CardHeader>

      <CardContent>
        {order.invoices.length === 0 ? (
          <div className="flex flex-col items-center gap-2 rounded-lg border border-dashed border-border py-8 text-center">
            <FileText className="size-6 text-muted-foreground/60" />
            <p className="text-sm text-muted-foreground">
              Belum ada invoice. Terbitkan setelah barang sampai untuk menagih pelunasan.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {order.invoices.map((invoice) => (
              <div key={invoice.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-medium">{invoice.invoice_number}</span>
                    <InvoiceStatusBadge status={invoice.status} />
                    <span className="text-xs uppercase text-muted-foreground">
                      {invoice.type === "dp" ? "DP" : "Pelunasan"}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Terbit {formatDate(invoice.issue_date)}
                    {invoice.due_date ? ` · jatuh tempo ${formatDate(invoice.due_date)}` : ""}
                    {invoice.sent_at ? ` · dikirim ${formatDate(invoice.sent_at)}` : ""}
                  </p>
                </div>

                <div className="flex items-center gap-2">
                  <div className="text-right">
                    {/* Angka besar adalah yang ditagih invoice ini — pada
                        invoice DP itu uang mukanya, bukan total order — supaya
                        cocok dengan dokumen yang diterima customer. */}
                    <p className="tabular font-semibold">
                      {formatIDR(invoice.type === "dp" ? invoice.dp_amount : invoice.total)}
                    </p>
                    <p className="tabular text-xs text-muted-foreground">
                      {invoice.type === "dp" ? "DP dari " : "total "}
                      {formatIDR(invoice.total)} · sisa {formatIDR(invoice.amount_due)}
                    </p>
                  </div>
                  <Button variant="outline" size="sm" asChild>
                    <a href={invoicePDFUrl(invoice.id)} target="_blank" rel="noopener noreferrer">
                      <FileText />
                      PDF
                    </a>
                  </Button>
                  {/* Invoice batal tidak lagi ditawarkan untuk dikirim maupun
                      dibatalkan ulang; yang lunas juga ditolak backend. */}
                  {invoice.status !== "void" && (
                    <Button
                      variant="success"
                      size="sm"
                      onClick={() => setMessageInvoiceId(invoice.id)}
                    >
                      <MessageCircle />
                      Kirim
                    </Button>
                  )}

                  {invoice.status !== "void" && invoice.status !== "paid" && (
                    <ConfirmButton
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive"
                      title={`Batalkan invoice ${invoice.invoice_number}?`}
                      description="Invoice ditandai batal dan tidak berlaku lagi sebagai tagihan. Dokumen PDF-nya tetap bisa dibuka sebagai jejak. Status order tidak ikut berubah, jadi invoice pengganti bisa langsung diterbitkan setelah ini."
                      confirmLabel="Ya, batalkan invoice"
                      destructive
                      error={voidInvoice.error}
                      onConfirm={() => handleVoid(invoice)}
                    >
                      <Ban />
                      Batalkan
                    </ConfirmButton>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>

      <FormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        title="Terbitkan Invoice"
        description="Nominal disalin saat invoice dibuat, jadi tidak ikut berubah kalau order diedit setelah ini."
        error={create.error}
        loading={create.isPending}
        onSubmit={handleSubmit}
        submitLabel="Terbitkan"
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Jenis invoice" htmlFor="invoice_type" required>
            <OptionSelect
              id="invoice_type"
              value={form.type}
              onChange={(value) => setForm({ ...form, type: value })}
              options={INVOICE_TYPE_OPTIONS}
            />
          </Field>

          <Field
            label="Jatuh tempo"
            htmlFor="due_date"
            hint="Kosongkan untuk memakai default pengaturan"
          >
            <Input
              id="due_date"
              type="date"
              value={form.due_date}
              onChange={(event) => setForm({ ...form, due_date: event.target.value })}
            />
          </Field>

          <Field label="Catatan invoice" htmlFor="invoice_notes" className="sm:col-span-2">
            <Textarea
              id="invoice_notes"
              rows={2}
              value={form.notes}
              onChange={(event) => setForm({ ...form, notes: event.target.value })}
              placeholder="Muncul pada catatan invoice"
            />
          </Field>
        </div>
      </FormDialog>

      <WAMessageDialog
        open={Boolean(messageInvoiceId)}
        onOpenChange={(open) => !open && setMessageInvoiceId(null)}
        title="Kirim invoice ke customer"
        description="Kirim dari nomor toko sendiri supaya customer mengenali pengirimnya."
        message={message.data}
        isLoading={message.isLoading}
        error={message.error}
        onSent={handleSent}
        sending={markSent.isPending}
      />
    </Card>
  );
}
