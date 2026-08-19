"use client";

import { Ban, FileText, MessageCircle, Plus, Wallet } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";

import { FilterSelect } from "@/components/filter-select";
import { InvoiceStatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { ConfirmButton } from "@/components/ui/confirm-button";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { WAMessageDialog } from "@/components/wa-message-dialog";
import { useDebounced } from "@/hooks/use-debounced";

import { DialogBuatInvoice } from "./dialog-buat-invoice";
import { DialogTandaiLunas } from "./dialog-tandai-lunas";
import {
  invoicePDFUrl,
  useInvoiceMessage,
  useInvoices,
  useMarkInvoiceSent,
  useVoidInvoice,
} from "@/hooks/use-operations";
import { ApiError } from "@/lib/api";
import { daysAgo, formatDate, formatIDR, toNumber } from "@/lib/utils";
import type { Invoice, SentChannel } from "@/types/api";

const INVOICE_STATUS_FILTER = [
  { value: "draft", label: "Draft" },
  { value: "sent", label: "Terkirim" },
  { value: "paid", label: "Lunas" },
  { value: "void", label: "Dibatalkan" },
] as const;

const INVOICE_TYPE_FILTER = [
  { value: "dp", label: "DP" },
  { value: "final", label: "Pelunasan" },
] as const;

export default function InvoicesPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");
  const [type, setType] = useState("");
  const [messageId, setMessageId] = useState<string | null>(null);
  const [buatOpen, setBuatOpen] = useState(false);
  const [lunasTarget, setLunasTarget] = useState<Invoice | null>(null);
  const debouncedSearch = useDebounced(search);

  const { data, isLoading, error } = useInvoices({ page, q: debouncedSearch, status, type });
  const message = useInvoiceMessage(messageId ?? undefined, Boolean(messageId));
  const markSent = useMarkInvoiceSent();
  const voidInvoice = useVoidInvoice();

  async function handleVoid(invoice: Invoice) {
    await voidInvoice.mutateAsync(invoice.id, {
      onSuccess: () => toast.success(`Invoice ${invoice.invoice_number} dibatalkan`),
      onError: (error) => {
        toast.error(error instanceof ApiError ? error.message : "Gagal membatalkan invoice");
      },
    });
  }

  function handleSent(channel: SentChannel) {
    if (!messageId) return;
    markSent.mutate(
      { id: messageId, channel },
      { onSuccess: () => toast.success("Invoice ditandai sudah dikirim") },
    );
  }

  return (
    <>
      <PageHeader
        title="Invoice"
        description="Semua tagihan yang sudah diterbitkan ke customer"
        actions={
          <Button onClick={() => setBuatOpen(true)}>
            <Plus />
            Buat Invoice
          </Button>
        }
      />

      <ErrorState error={error} />

      <div className="flex flex-wrap gap-3">
        <SearchInput
          value={search}
          onChange={(value) => {
            setSearch(value);
            setPage(1);
          }}
          placeholder="Cari nomor invoice, order, atau customer…"
          className="min-w-64 flex-1 sm:max-w-md"
        />
        <FilterSelect
          value={status}
          onChange={(value) => {
            setStatus(value);
            setPage(1);
          }}
          allLabel="Semua status"
          options={INVOICE_STATUS_FILTER}
          className="sm:w-44"
        />
        <FilterSelect
          value={type}
          onChange={(value) => {
            setType(value);
            setPage(1);
          }}
          allLabel="Semua jenis"
          options={INVOICE_TYPE_FILTER}
          className="sm:w-44"
        />
      </div>

      <div>
        <DataTable
          columns={9}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada invoice"
          emptyDescription="Invoice DP diterbitkan dari detail order; invoice pelunasan dari tombol Buat Invoice di atas."
          head={
            <TR>
              <TH>Invoice</TH>
              <TH className="hidden sm:table-cell">Jenis</TH>
              <TH className="hidden xl:table-cell">Order</TH>
              <TH className="hidden md:table-cell">Customer</TH>
              <TH className="hidden text-right lg:table-cell">Total</TH>
              <TH className="text-right">Sisa</TH>
              <TH className="hidden sm:table-cell">Status</TH>
              <TH className="hidden text-right whitespace-nowrap sm:table-cell">Terbit</TH>
              <TH className="text-right">Aksi</TH>
            </TR>
          }
        >
          {data?.items.map((invoice) => {
            const overdue =
              invoice.due_date &&
              invoice.status !== "paid" &&
              invoice.status !== "void" &&
              daysAgo(invoice.due_date) > 0;

            return (
              <TR key={invoice.id}>
                <TD className="whitespace-normal">
                  <p className="font-medium">{invoice.invoice_number}</p>
                  {/* Jenis, tanggal terbit, customer, dan status dilipat ke sini
                      selama kolom masing-masing disembunyikan; tanpa itu daftar
                      invoice di ponsel tinggal deretan nomor tanpa penanda. */}
                  <p className="text-xs text-muted-foreground sm:hidden">
                    {invoice.type === "dp" ? "DP" : "Pelunasan"} · {formatDate(invoice.issue_date)}
                  </p>
                  <p className="text-xs text-muted-foreground md:hidden">{invoice.customer_name}</p>
                  <div className="mt-1 sm:hidden">
                    <InvoiceStatusBadge status={invoice.status} />
                  </div>
                </TD>
                <TD className="hidden text-sm sm:table-cell">
                  {invoice.type === "dp" ? "DP" : "Pelunasan"}
                </TD>
                <TD className="hidden xl:table-cell">
                  <Link href={`/orders/${invoice.order_id}`} className="text-sm hover:underline">
                    {invoice.order_number}
                  </Link>
                  <p className="text-xs text-muted-foreground">{invoice.trip_code}</p>
                </TD>
                <TD className="hidden text-sm whitespace-normal md:table-cell">
                  {invoice.customer_name}
                </TD>
                <TD className="tabular hidden text-right font-medium lg:table-cell">
                  {formatIDR(invoice.type === "dp" ? invoice.dp_amount : invoice.total)}
                  {invoice.type === "dp" && (
                    <p className="text-xs font-normal text-muted-foreground">
                      dari {formatIDR(invoice.total)}
                    </p>
                  )}
                </TD>
                <TD
                  className={`tabular text-right ${
                    toNumber(invoice.amount_due) > 0 ? "text-amber-600" : "text-muted-foreground"
                  }`}
                >
                  {formatIDR(invoice.amount_due)}
                  {overdue && <p className="text-xs font-medium text-destructive">lewat tempo</p>}
                </TD>
                <TD className="hidden sm:table-cell">
                  <InvoiceStatusBadge status={invoice.status} />
                </TD>
                <TD className="hidden text-right text-sm whitespace-nowrap text-muted-foreground sm:table-cell">
                  {formatDate(invoice.issue_date)}
                </TD>
                <TD>
                  <div className="flex justify-end gap-1">
                    <Button variant="ghost" size="icon-sm" asChild>
                      <a href={invoicePDFUrl(invoice.id)} target="_blank" rel="noopener noreferrer">
                        <FileText />
                        <span className="sr-only">PDF</span>
                      </a>
                    </Button>
                    {invoice.status !== "void" && (
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        className="text-emerald-600 hover:text-emerald-700"
                        onClick={() => setMessageId(invoice.id)}
                      >
                        <MessageCircle />
                        <span className="sr-only">Kirim</span>
                      </Button>
                    )}

                    {/* Menandai lunas berarti mencatat uangnya masuk ke order,
                        bukan sekadar mengubah label barisnya. */}
                    {invoice.status !== "void" && toNumber(invoice.amount_due) > 0 && (
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        className="text-emerald-600 hover:text-emerald-700"
                        onClick={() => setLunasTarget(invoice)}
                      >
                        <Wallet />
                        <span className="sr-only">Catat pembayaran</span>
                      </Button>
                    )}

                    {invoice.status !== "void" && invoice.status !== "paid" && (
                      <ConfirmButton
                        variant="ghost"
                        size="icon-sm"
                        className="text-destructive hover:text-destructive"
                        title={`Batalkan invoice ${invoice.invoice_number}?`}
                        description="Invoice ditandai batal dan tidak berlaku lagi sebagai tagihan. Dokumen PDF-nya tetap bisa dibuka sebagai jejak. Status order tidak ikut berubah, jadi invoice pengganti bisa langsung diterbitkan setelah ini."
                        confirmLabel="Ya, batalkan invoice"
                        destructive
                        error={voidInvoice.error}
                        onConfirm={() => handleVoid(invoice)}
                      >
                        <Ban />
                        <span className="sr-only">Batalkan invoice</span>
                      </ConfirmButton>
                    )}
                  </div>
                </TD>
              </TR>
            );
          })}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>

      {buatOpen && (
        <DialogBuatInvoice
          onClose={() => setBuatOpen(false)}
          // Langsung tawarkan pengirimannya, karena itu langkah berikutnya.
          onIssued={(invoice) => setMessageId(invoice.id)}
        />
      )}

      {lunasTarget && (
        <DialogTandaiLunas invoice={lunasTarget} onClose={() => setLunasTarget(null)} />
      )}

      <WAMessageDialog
        open={Boolean(messageId)}
        onOpenChange={(open) => !open && setMessageId(null)}
        title="Kirim invoice ke customer"
        message={message.data}
        isLoading={message.isLoading}
        error={message.error}
        onSent={handleSent}
        sending={markSent.isPending}
      />
    </>
  );
}
