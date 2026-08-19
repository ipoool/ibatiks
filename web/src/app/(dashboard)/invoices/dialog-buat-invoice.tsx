"use client";

import { Check, Search } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Field } from "@/components/ui/field";
import { FormDialog } from "@/components/ui/form-dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState } from "@/components/ui/page";
import { useDebounced } from "@/hooks/use-debounced";
import { useInvoiceCandidates, useIssueFinalInvoice } from "@/hooks/use-operations";
import { ApiError } from "@/lib/api";
import { cn, formatDate, formatIDR } from "@/lib/utils";
import type { Invoice, InvoiceCandidate } from "@/types/api";

/**
 * Menerbitkan invoice pelunasan dari menu Invoice.
 *
 * Yang muncul hanya order yang benar-benar siap ditagih: DP-nya sudah masuk,
 * ongkirnya sudah ditetapkan, masih ada sisa tagihan, dan belum punya invoice
 * pelunasan yang berlaku. Menawarkan order yang belum memenuhi syarat hanya
 * akan berakhir dengan penolakan dari backend, dan admin tidak akan tahu apa
 * yang kurang.
 */
export function DialogBuatInvoice({
  onClose,
  onIssued,
}: {
  onClose: () => void;
  onIssued: (invoice: Invoice) => void;
}) {
  const [search, setSearch] = useState("");
  const [terpilih, setTerpilih] = useState<InvoiceCandidate | null>(null);
  const [form, setForm] = useState({ due_date: "", notes: "" });
  const debouncedSearch = useDebounced(search);

  const { data, isLoading, error } = useInvoiceCandidates(debouncedSearch, true);
  const issue = useIssueFinalInvoice();

  const kandidat = data ?? [];

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    if (!terpilih) return;

    issue.mutate(
      {
        orderId: terpilih.order_id,
        due_date: form.due_date || undefined,
        notes: form.notes || null,
      },
      {
        onSuccess: (invoice) => {
          toast.success(`Invoice ${invoice.invoice_number} diterbitkan`);
          onIssued(invoice);
          onClose();
        },
        onError: (err) => {
          toast.error(err instanceof ApiError ? err.message : "Gagal menerbitkan invoice");
        },
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title="Buat Invoice Pelunasan"
      description="Menagih sisa pesanan beserta ongkirnya. Nominalnya disalin saat invoice dibuat, jadi tidak ikut berubah kalau order diedit setelah ini."
      onSubmit={handleSubmit}
      loading={issue.isPending}
      submitLabel="Terbitkan"
      submitDisabled={!terpilih}
    >
      <ErrorState error={error ?? issue.error} />

      <Field label="Order yang ditagih" required>
        <div className="space-y-2">
          <div className="relative">
            <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Cari nomor order atau nama customer…"
              className="pl-9"
            />
          </div>

          {isLoading && <p className="text-xs text-muted-foreground">Memuat…</p>}

          {!isLoading && kandidat.length === 0 && (
            <p className="text-xs text-muted-foreground">
              Belum ada order yang siap ditagih. Order baru muncul di sini setelah DP-nya masuk dan
              ongkirnya ditetapkan di menu Pengiriman.
            </p>
          )}

          {kandidat.length > 0 && (
            <ul className="max-h-64 divide-y overflow-y-auto rounded-md border">
              {kandidat.map((order) => {
                const aktif = terpilih?.order_id === order.order_id;
                return (
                  <li key={order.order_id}>
                    <button
                      type="button"
                      onClick={() => setTerpilih(order)}
                      className={cn(
                        "flex w-full items-start gap-2 px-3 py-2 text-left text-sm hover:bg-accent",
                        aktif && "bg-accent",
                      )}
                    >
                      <Check
                        className={cn("mt-0.5 size-4 shrink-0", aktif ? "opacity-100" : "opacity-0")}
                      />
                      <span className="min-w-0 flex-1">
                        <span className="block font-medium">{order.order_number}</span>
                        <span className="block text-xs text-muted-foreground">
                          {order.customer_name} · {order.trip_code} ·{" "}
                          {formatDate(order.order_date)}
                        </span>
                      </span>
                      <span className="shrink-0 text-right">
                        <span className="tabular block font-medium">
                          {formatIDR(order.balance_due)}
                        </span>
                        <span className="block text-xs text-muted-foreground">
                          ongkir {formatIDR(order.shipping_fee)}
                        </span>
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </Field>

      <div className="grid gap-4 sm:grid-cols-2">
        <Field
          label="Jatuh tempo"
          htmlFor="buat_due_date"
          hint="Kosongkan untuk memakai default pengaturan"
        >
          <Input
            id="buat_due_date"
            type="date"
            value={form.due_date}
            onChange={(event) => setForm({ ...form, due_date: event.target.value })}
          />
        </Field>

        <Field label="Catatan invoice" htmlFor="buat_notes">
          <Textarea
            id="buat_notes"
            rows={2}
            value={form.notes}
            onChange={(event) => setForm({ ...form, notes: event.target.value })}
            placeholder="Muncul pada catatan invoice"
          />
        </Field>
      </div>
    </FormDialog>
  );
}
