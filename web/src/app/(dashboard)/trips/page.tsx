"use client";

import { CalendarDays, Package, Plus, Users } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { TRIP_STATUS_OPTIONS, TripStatusBadge } from "@/components/status-badge";
import { FilterSelect } from "@/components/filter-select";
import { Button } from "@/components/ui/button";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useDebounced } from "@/hooks/use-debounced";
import { useTrips } from "@/hooks/use-trips";
import { formatDate, formatNumber } from "@/lib/utils";
import type { TripStatus } from "@/types/api";

import { TripFormDialog } from "./trip-form";

export default function TripsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<TripStatus | "">("");
  const [formOpen, setFormOpen] = useState(false);
  const debouncedSearch = useDebounced(search);

  const { data, isLoading, error } = useTrips({ page, q: debouncedSearch, status });

  return (
    <>
      <PageHeader
        title="Trip"
        description="Setiap perjalanan ke luar negeri beserta katalog dan pesanannya"
        actions={
          <Button onClick={() => setFormOpen(true)}>
            <Plus />
            Buat Trip
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
          placeholder="Cari judul, kode, atau negara…"
          className="min-w-64 flex-1 sm:max-w-md"
        />
        <FilterSelect
          value={status}
          onChange={(value) => {
            setStatus(value);
            setPage(1);
          }}
          allLabel="Semua status"
          options={TRIP_STATUS_OPTIONS}
          className="sm:w-56"
        />
      </div>

      <div>
        <DataTable
          columns={6}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle={search || status ? "Trip tidak ditemukan" : "Belum ada trip"}
          emptyDescription={
            search || status
              ? "Coba ubah kata kunci atau filter status."
              : "Buat trip pertama, lalu susun katalog produk yang akan ditawarkan."
          }
          emptyAction={
            !search &&
            !status && (
              <Button onClick={() => setFormOpen(true)}>
                <Plus />
                Buat Trip
              </Button>
            )
          }
          head={
            <TR>
              <TH>Trip</TH>
              <TH className="hidden sm:table-cell">Tanggal</TH>
              <TH className="hidden xl:table-cell">Kurs</TH>
              <TH className="hidden text-right lg:table-cell">Katalog</TH>
              <TH className="hidden text-right sm:table-cell">Order</TH>
              <TH>Status</TH>
            </TR>
          }
        >
          {data?.items.map((trip) => (
            <TR key={trip.id}>
              <TD className="whitespace-normal">
                <Link href={`/trips/${trip.id}`} className="font-medium hover:underline">
                  {trip.title}
                </Link>
                <p className="text-xs text-muted-foreground">
                  {trip.code} · {trip.country}
                  {trip.city ? `, ${trip.city}` : ""}
                </p>
                {/* Tanggal berangkat ikut di kolom trip selama kolom tanggalnya
                    disembunyikan — itu yang paling dicari saat memilih trip. */}
                <p className="text-xs text-muted-foreground sm:hidden">
                  {formatDate(trip.depart_date)} · {formatNumber(trip.total_orders ?? 0)} order
                </p>
              </TD>
              <TD className="hidden whitespace-nowrap text-sm sm:table-cell">
                <span className="inline-flex items-center gap-1.5">
                  <CalendarDays className="size-3.5 text-muted-foreground" />
                  {formatDate(trip.depart_date)}
                </span>
                <p className="pl-5 text-xs text-muted-foreground">
                  s/d {formatDate(trip.return_date)}
                </p>
              </TD>
              <TD className="tabular hidden whitespace-nowrap text-sm xl:table-cell">
                1 {trip.currency} = {formatNumber(trip.exchange_rate)}
              </TD>
              <TD className="tabular hidden text-right lg:table-cell">
                <span className="inline-flex items-center gap-1.5">
                  <Package className="size-3.5 text-muted-foreground" />
                  {formatNumber(trip.catalog_items ?? 0)}
                </span>
              </TD>
              <TD className="tabular hidden text-right sm:table-cell">
                <span className="inline-flex items-center gap-1.5">
                  <Users className="size-3.5 text-muted-foreground" />
                  {formatNumber(trip.total_orders ?? 0)}
                </span>
              </TD>
              <TD>
                <TripStatusBadge status={trip.status} />
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>

      <TripFormDialog open={formOpen} onOpenChange={setFormOpen} />
    </>
  );
}
