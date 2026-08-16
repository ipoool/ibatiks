"use client";

import { Plus } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { CheckboxField } from "@/components/ui/checkbox-field";
import { ORDER_SOURCE_OPTIONS, ORDER_STATUS_OPTIONS, OrderSourceBadge, OrderStatusBadge } from "@/components/status-badge";
import { FilterSelect } from "@/components/filter-select";
import { Button } from "@/components/ui/button";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useDebounced } from "@/hooks/use-debounced";
import { useOrders } from "@/hooks/use-orders";
import { useTrips } from "@/hooks/use-trips";
import {
  formatDate,
  formatForeign,
  formatIDR,
  formatNumber,
  idrToForeign,
  toNumber,
} from "@/lib/utils";
import type { Order, OrderSource, OrderStatus } from "@/types/api";

export default function OrdersPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<OrderStatus | "">("");
  const [source, setSource] = useState<OrderSource | "">("");
  const [tripId, setTripId] = useState("");
  const [unpaidOnly, setUnpaidOnly] = useState(false);
  const [showForeign, setShowForeign] = useState(false);
  const debouncedSearch = useDebounced(search);

  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];
  const { data, isLoading, error } = useOrders({
    page,
    q: debouncedSearch,
    status,
    source,
    trip_id: tripId || undefined,
    unpaid_only: unpaidOnly || undefined,
  });

  function resetPage<T>(setter: (value: T) => void) {
    return (value: T) => {
      setter(value);
      setPage(1);
    };
  }

  /*
   * Nominal disimpan dalam rupiah; tampilan mata uang trip dihitung ulang dari
   * kurs yang dikunci pada trip. Order lintas trip bisa punya mata uang berbeda,
   * jadi konversi dilakukan per baris, bukan sekali untuk seluruh tabel.
   */
  function money(value: string, order: Order) {
    const currency = order.trip_currency;
    if (!showForeign || !currency || currency === "IDR") return formatIDR(value);
    return formatForeign(idrToForeign(value, order.trip_exchange_rate), currency);
  }

  return (
    <>
      <PageHeader
        title="Order"
        description="Seluruh pesanan customer beserta status pembayaran dan pengirimannya"
        actions={
          <Button asChild>
            <Link href="/orders/new">
              <Plus />
              Catat Order
            </Link>
          </Button>
        }
      />

      <ErrorState error={error} />

      <div className="flex flex-wrap gap-3">
        <SearchInput
          value={search}
          onChange={resetPage(setSearch)}
          placeholder="Cari nomor order, nama customer, atau penerima…"
          className="min-w-64 flex-1 sm:max-w-sm"
        />

        <FilterSelect
          value={tripId}
          onChange={resetPage(setTripId)}
          allLabel="Semua trip"
          options={tripOptions}
          className="sm:w-56"
        />

        <FilterSelect
          value={status}
          onChange={resetPage(setStatus)}
          allLabel="Semua status"
          options={ORDER_STATUS_OPTIONS}
          className="sm:w-48"
        />

        <FilterSelect
          value={source}
          onChange={resetPage(setSource)}
          allLabel="Semua channel"
          options={ORDER_SOURCE_OPTIONS}
          className="sm:w-44"
        />

        <CheckboxField
          id="unpaid_only"
          variant="boxed"
          checked={unpaidOnly}
          onCheckedChange={resetPage(setUnpaidOnly)}
        >
          Belum lunas
        </CheckboxField>

        <CheckboxField
          id="show_foreign"
          variant="boxed"
          checked={showForeign}
          onCheckedChange={setShowForeign}
        >
          Mata uang trip
        </CheckboxField>
      </div>

      <div>
        <DataTable
          columns={8}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada order"
          emptyDescription="Catat pesanan yang masuk dari WhatsApp atau sosial media di sini."
          emptyAction={
            <Button asChild>
              <Link href="/orders/new">
                <Plus />
                Catat Order
              </Link>
            </Button>
          }
          head={
            <TR>
              <TH className="min-w-32">Order</TH>
              <TH className="hidden min-w-32 sm:table-cell">Customer</TH>
              <TH className="hidden min-w-28 xl:table-cell">Trip</TH>
              <TH className="hidden w-24 xl:table-cell">Channel</TH>
              <TH className="hidden w-16 text-right lg:table-cell">Pcs</TH>
              <TH className="w-28 text-right">Total</TH>
              <TH className="hidden w-28 text-right sm:table-cell">Sisa Bayar</TH>
              <TH className="w-28">Status</TH>
            </TR>
          }
        >
          {data?.items.map((order) => (
            <TR key={order.id}>
              <TD>
                <Link
                  href={`/orders/${order.id}`}
                  className="font-medium whitespace-nowrap hover:underline"
                >
                  {order.order_number}
                </Link>
                <p className="text-xs text-muted-foreground">{formatDate(order.order_date)}</p>
                {/* Nama customer ikut di kolom order selama kolomnya sendiri
                    disembunyikan; tanpa itu daftar order di ponsel cuma berisi
                    nomor yang tidak bisa dibedakan satu sama lain. */}
                <p className="text-xs text-muted-foreground sm:hidden">{order.customer_name}</p>
              </TD>
              <TD className="hidden sm:table-cell">
                <p className="font-medium">{order.customer_name}</p>
                <p className="text-xs text-muted-foreground">{order.shipping_city}</p>
              </TD>
              <TD className="hidden text-sm xl:table-cell">
                <p>{order.trip_code}</p>
                <p className="max-w-40 truncate text-xs text-muted-foreground">{order.trip_title}</p>
              </TD>
              <TD className="hidden xl:table-cell">
                <OrderSourceBadge source={order.order_source} />
              </TD>
              <TD className="tabular hidden text-right text-sm lg:table-cell">
                {formatNumber(order.total_qty ?? 0)}
              </TD>
              <TD className="tabular whitespace-nowrap text-right font-medium">
                {money(order.total, order)}
                {toNumber(order.balance_due) > 0 && (
                  <p className="text-xs font-normal text-amber-600 sm:hidden">
                    sisa {money(order.balance_due, order)}
                  </p>
                )}
              </TD>
              <TD
                className={`tabular hidden whitespace-nowrap text-right sm:table-cell ${
                  toNumber(order.balance_due) > 0
                    ? "font-medium text-amber-600"
                    : "text-muted-foreground"
                }`}
              >
                {money(order.balance_due, order)}
              </TD>
              <TD>
                <OrderStatusBadge status={order.status} />
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}
