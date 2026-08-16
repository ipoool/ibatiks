"use client";

import { ExternalLink, MessageCircle, Printer } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { FilterSelect } from "@/components/filter-select";
import { ShipmentStatusBadge } from "@/components/status-badge";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useDebounced } from "@/hooks/use-debounced";
import { deliveryNoteUrl, useShipments } from "@/hooks/use-operations";
import { useTrips } from "@/hooks/use-trips";
import { formatDate, formatIDR, toNumber } from "@/lib/utils";
import type { ShipmentStatus } from "@/types/api";

const STATUS_OPTIONS: Array<{ value: ShipmentStatus; label: string }> = [
  { value: "packing", label: "Dikemas" },
  { value: "ready", label: "Siap kirim" },
  { value: "shipped", label: "Dikirim" },
  { value: "delivered", label: "Diterima" },
  { value: "returned", label: "Retur" },
];

export default function ShipmentsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<ShipmentStatus | "">("");
  const [tripId, setTripId] = useState("");
  const debouncedSearch = useDebounced(search);

  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];
  const { data, isLoading, error } = useShipments({
    page,
    q: debouncedSearch,
    status,
    trip_id: tripId || undefined,
  });

  return (
    <>
      <PageHeader
        title="Pengiriman"
        description="Paket yang sudah dikemas beserta nomor resi JNE-nya"
      />

      <ErrorState error={error} />

      <div className="flex flex-wrap gap-3">
        <SearchInput
          value={search}
          onChange={(value) => {
            setSearch(value);
            setPage(1);
          }}
          placeholder="Cari nomor order, customer, atau resi…"
          className="min-w-64 flex-1 sm:max-w-md"
        />
        <FilterSelect
          value={status}
          onChange={(value) => {
            setStatus(value);
            setPage(1);
          }}
          allLabel="Semua status"
          options={STATUS_OPTIONS}
          className="sm:w-48"
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

      <div>
        <DataTable
          columns={6}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada paket"
          emptyDescription="Paket muncul di sini setelah order ditandai dikemas."
          head={
            <TR>
              <TH>Order</TH>
              <TH className="hidden sm:table-cell">Penerima</TH>
              <TH>Resi</TH>
              <TH className="hidden text-right lg:table-cell">Ongkir</TH>
              <TH className="hidden lg:table-cell">Dikirim</TH>
              <TH>Status</TH>
            </TR>
          }
        >
          {data?.items.map((shipment) => (
            <TR key={shipment.id}>
              <TD className="whitespace-normal">
                <Link href={`/orders/${shipment.order_id}`} className="font-medium hover:underline">
                  {shipment.order_number}
                </Link>
                <p className="text-xs text-muted-foreground">{shipment.customer_name}</p>
                {/* Kota tujuan menyusul nomor order saat kolom penerima
                    disembunyikan — itu yang membedakan satu paket dari lainnya
                    ketika mencocokkan tumpukan kiriman. */}
                <p className="text-xs text-muted-foreground sm:hidden">
                  {shipment.shipping_city}
                </p>
              </TD>
              <TD className="hidden sm:table-cell">
                <p className="text-sm font-medium">{shipment.recipient_name}</p>
                <p className="text-xs text-muted-foreground">{shipment.shipping_city}</p>
              </TD>
              <TD>
                {shipment.tracking_number ? (
                  <div className="flex items-center gap-1.5">
                    <span className="tabular font-mono text-sm">{shipment.tracking_number}</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <a
                          href="https://www.jne.co.id/tracking-package"
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-muted-foreground hover:text-foreground"
                        >
                          <ExternalLink className="size-3.5" />
                          <span className="sr-only">Lacak di situs JNE</span>
                        </a>
                      </TooltipTrigger>
                      <TooltipContent>Lacak di situs JNE</TooltipContent>
                    </Tooltip>
                  </div>
                ) : (
                  <span className="text-sm text-muted-foreground">belum ada</span>
                )}
                <p className="text-xs text-muted-foreground">
                  {shipment.courier} {shipment.service}
                </p>
              </TD>
              {/* Paket yang belum diserahkan ke kurir belum punya ongkir asli;
                  yang ditampilkan perkiraannya, ditandai supaya tidak dikira
                  angka final saat merekap biaya. */}
              <TD className="tabular hidden text-right lg:table-cell">
                {toNumber(shipment.shipping_cost) > 0 ? (
                  formatIDR(shipment.shipping_cost)
                ) : toNumber(shipment.estimated_cost) > 0 ? (
                  <span className="text-muted-foreground">
                    {formatIDR(shipment.estimated_cost)}
                    <span className="block text-xs">estimasi</span>
                  </span>
                ) : (
                  <span className="text-muted-foreground">{formatIDR(0)}</span>
                )}
              </TD>
              <TD className="hidden text-sm lg:table-cell">
                {formatDate(shipment.shipped_at)}
                {shipment.shipped_at && !shipment.customer_notified_at && (
                  <p className="flex items-center gap-1 text-xs text-amber-600">
                    <MessageCircle className="size-3" />
                    belum dikabari
                  </p>
                )}
              </TD>
              <TD>
                <div className="flex items-center justify-between gap-1">
                  <ShipmentStatusBadge status={shipment.status} />
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button variant="ghost" size="icon-sm" asChild>
                        <a
                          href={deliveryNoteUrl(shipment.order_id)}
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
                </div>
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}
