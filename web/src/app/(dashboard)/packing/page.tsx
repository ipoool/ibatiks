"use client";

import { PackageCheck, Printer } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { FilterSelect, OptionSelect } from "@/components/filter-select";
import { BalanceDue } from "@/components/balance-due";
import { OrderStatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { ErrorState, PageHeader } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { deliveryNoteUrl } from "@/hooks/use-operations";
import { useOrders } from "@/hooks/use-orders";
import { useTrips } from "@/hooks/use-trips";
import { formatNumber, toNumber } from "@/lib/utils";
import type { OrderStatus } from "@/types/api";

/**
 * Antrean kerja gudang: order yang barangnya sudah datang tapi belum dikirim.
 *
 * Halaman ini menjawab pertanyaan "hari ini saya harus mengemas apa saja",
 * yang tidak enak dicari lewat daftar order umum.
 */
/* Label antrean memakai bahasa gudang, bukan nama status order: yang dicari
   petugas adalah "hari ini saya kerjakan apa", bukan status formalnya. */
const QUEUE_OPTIONS: ReadonlyArray<{ value: OrderStatus; label: string }> = [
  { value: "dp_paid", label: "Siap dikemas" },
  { value: "packed", label: "Sedang dikemas" },
  { value: "invoiced", label: "Menunggu pelunasan" },
  { value: "paid", label: "Siap dikirim" },
];

export default function PackingPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<OrderStatus>("dp_paid");
  const [tripId, setTripId] = useState("");

  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];
  const { data, isLoading, error } = useOrders({
    page,
    status,
    trip_id: tripId || undefined,
    per_page: 50,
  });

  const STATUS_HINT: Record<string, string> = {
    arrived: "Barang sudah dicocokkan, siap dikemas atas nama customer.",
    packed: "Sudah dikemas — terbitkan invoice pelunasan.",
    invoiced: "Invoice terkirim, menunggu pelunasan masuk.",
    paid: "Sudah lunas, tinggal input nomor resi.",
  };

  return (
    <>
      <PageHeader
        title="Antrean Kemas &amp; Kirim"
        description="Order yang barangnya sudah tiba dan sedang menunggu diproses"
      />

      <ErrorState error={error} />

      <div className="flex flex-wrap gap-3">
        <OptionSelect
          value={status}
          onChange={(value) => {
            setStatus(value);
            setPage(1);
          }}
          options={QUEUE_OPTIONS}
          className="sm:w-64"
        />

        <FilterSelect
          value={tripId}
          onChange={(value) => {
            setTripId(value);
            setPage(1);
          }}
          allLabel="Semua trip"
          options={tripOptions}
          className="sm:w-56"
        />
      </div>

      <p className="text-sm text-muted-foreground">{STATUS_HINT[status]}</p>

      <div>
        <DataTable
          columns={6}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Antrean kosong"
          emptyDescription="Tidak ada order pada tahap ini. Selamat, pekerjaan gudang sedang bersih."
          head={
            <TR>
              <TH>Order</TH>
              <TH className="hidden sm:table-cell">Penerima</TH>
              <TH className="hidden lg:table-cell">Tujuan</TH>
              <TH className="hidden text-right lg:table-cell">Item</TH>
              <TH className="text-right">Sisa Bayar</TH>
              <TH className="text-right">Aksi</TH>
            </TR>
          }
        >
          {data?.items.map((order) => (
            <TR key={order.id}>
              <TD className="whitespace-normal">
                <Link href={`/orders/${order.id}`} className="font-medium hover:underline">
                  {order.order_number}
                </Link>
                <div className="mt-1">
                  <OrderStatusBadge status={order.status} settled={toNumber(order.balance_due) <= 0} />
                </div>
                {/* Penerima dan kota menyusul nomor order saat kolomnya
                    disembunyikan; itu yang dipakai mencocokkan paket. */}
                <p className="mt-1 text-xs text-muted-foreground sm:hidden">
                  {order.recipient_name} · {order.shipping_city}
                </p>
              </TD>
              <TD className="hidden sm:table-cell">
                <p className="font-medium">{order.recipient_name}</p>
                <p className="text-xs text-muted-foreground">{order.recipient_phone}</p>
              </TD>
              <TD className="hidden max-w-xs lg:table-cell">
                <p className="truncate text-sm">{order.shipping_address}</p>
                <p className="text-xs text-muted-foreground">{order.shipping_city}</p>
              </TD>
              <TD className="tabular hidden text-right lg:table-cell">
                {formatNumber(order.total_qty ?? 0)} pcs
              </TD>
              <TD className="tabular text-right font-medium">
                <BalanceDue amount={order.balance_due} status={order.status} />
              </TD>
              <TD className="text-right">
                {/* Surat jalan dicetak berbarengan dengan mengemas, jadi
                    tombolnya duduk di antrean ini juga. */}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button variant="ghost" size="icon-sm" asChild className="mr-1">
                      <a
                        href={deliveryNoteUrl(order.id)}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <Printer />
                        <span className="sr-only">Cetak surat jalan</span>
                      </a>
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Cetak surat jalan</TooltipContent>
                </Tooltip>
                <Button size="sm" asChild>
                  <Link href={`/orders/${order.id}`}>
                    <PackageCheck />
                    <span className="hidden sm:inline">Proses</span>
                    <span className="sr-only sm:hidden">Proses order {order.order_number}</span>
                  </Link>
                </Button>
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}
