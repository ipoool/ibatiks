"use client";

import { ExternalLink, MessageCircle, Package, Printer, Truck } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { DataTable, TD, TH, TR } from "@/components/data-table";
import { FilterSelect } from "@/components/filter-select";
import { OrderStatusBadge } from "@/components/status-badge";
import { ShippingInfoButton } from "@/components/shipping-info";
import { Button } from "@/components/ui/button";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useDebounced } from "@/hooks/use-debounced";
import { labelUrl, useShippingQueue } from "@/hooks/use-operations";
import { useTrips } from "@/hooks/use-trips";
import { formatIDR, formatNumber, toNumber } from "@/lib/utils";
import type { ShippingQueueItem, ShippingStage } from "@/types/api";

import { DialogKemas } from "./dialog-kemas";
import { DialogResi } from "./dialog-resi";

/*
 * Tahap kerja, bukan status tersimpan. Yang dicari petugas gudang adalah "hari
 * ini saya kerjakan apa", dan jawabannya tidak selalu sama dengan status order:
 * order Diproses bisa berarti belum ditimbang atau sudah siap ditagih.
 */
const STAGE_OPTIONS = [
  { value: "perlu_kemas", label: "Perlu dikemas" },
  { value: "siap_kirim", label: "Siap dikirim" },
  { value: "terkirim", label: "Sudah dikirim" },
] as const satisfies ReadonlyArray<{ value: ShippingStage; label: string }>;

/**
 * Menu Pengiriman.
 *
 * Satu tabel, bukan dua. Mengemas dan menyerahkan ke kurir adalah satu
 * pekerjaan yang sama di meja yang sama; memisahkannya jadi dua daftar memaksa
 * petugas bolak-balik mencari order yang sama di tempat berbeda.
 *
 * Yang didaftar adalah order yang DP-nya sudah masuk — bukan paket. Paket baru
 * terbentuk setelah data kemasan diisi, jadi mendaftar paket akan
 * menyembunyikan justru pekerjaan yang belum dikerjakan.
 */
export default function ShipmentsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [stage, setStage] = useState<ShippingStage | "">("");
  const [tripId, setTripId] = useState("");
  const [kemasTarget, setKemasTarget] = useState<ShippingQueueItem | null>(null);
  const [resiTarget, setResiTarget] = useState<ShippingQueueItem | null>(null);
  const debouncedSearch = useDebounced(search);

  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];

  const { data, isLoading, error } = useShippingQueue({
    page,
    q: debouncedSearch,
    stage,
    trip_id: tripId || undefined,
  });

  return (
    <>
      <PageHeader
        title="Pengiriman"
        description="Order yang DP-nya sudah masuk: ditimbang, dihitung ongkirnya, lalu diserahkan ke kurir"
      />

      <ErrorState error={error} />

      <div className="flex flex-wrap gap-3">
        <SearchInput
          value={search}
          onChange={(value) => {
            setSearch(value);
            setPage(1);
          }}
          placeholder="Cari nomor order, customer, penerima, atau resi…"
          className="min-w-64 flex-1 sm:max-w-md"
        />
        <FilterSelect
          value={stage}
          onChange={(value) => {
            setStage(value);
            setPage(1);
          }}
          allLabel="Semua tahap"
          options={STAGE_OPTIONS}
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
        <ShippingInfoButton className="self-center" />
      </div>

      <div>
        <DataTable
          columns={6}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Antrean kosong"
          emptyDescription="Order muncul di sini setelah DP-nya diterima."
          head={
            <TR>
              <TH className="min-w-44">Order</TH>
              <TH className="hidden sm:table-cell">Penerima</TH>
              <TH className="hidden lg:table-cell">Paket</TH>
              <TH className="hidden text-right lg:table-cell">Ongkir</TH>
              <TH className="min-w-32">Resi</TH>
              <TH className="text-right">Aksi</TH>
            </TR>
          }
        >
          {data?.items.map((item) => (
            <BarisPengiriman
              key={item.order_id}
              item={item}
              onKemas={() => setKemasTarget(item)}
              onResi={() => setResiTarget(item)}
            />
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>

      {kemasTarget && (
        <DialogKemas item={kemasTarget} onClose={() => setKemasTarget(null)} />
      )}
      {resiTarget && <DialogResi item={resiTarget} onClose={() => setResiTarget(null)} />}
    </>
  );
}

function BarisPengiriman({
  item,
  onKemas,
  onResi,
}: {
  item: ShippingQueueItem;
  onKemas: () => void;
  onResi: () => void;
}) {
  const sudahDikemas = (item.weight_gram ?? 0) > 0;
  const ongkirTerisi = toNumber(item.shipping_fee) > 0;
  const sudahDikirim = Boolean(item.tracking_number);
  const lunas = toNumber(item.balance_due) <= 0;

  return (
    <TR>
      <TD className="whitespace-normal">
        <Link href={`/orders/${item.order_id}`} className="font-medium hover:underline">
          {item.order_number}
        </Link>
        <div className="mt-1">
          <OrderStatusBadge status={item.order_status} settled={lunas} />
        </div>
        <p className="text-xs text-muted-foreground">
          {item.customer_name} · {item.trip_code}
        </p>
        {/* Penerima dan kota menyusul nomor order saat kolomnya disembunyikan;
            itu yang dipakai mencocokkan paket di meja kemas. */}
        <p className="mt-1 text-xs text-muted-foreground sm:hidden">
          {item.recipient_name} · {item.shipping_city}
        </p>
      </TD>

      <TD className="hidden sm:table-cell">
        <p className="font-medium">{item.recipient_name}</p>
        <p className="text-xs text-muted-foreground">{item.shipping_city}</p>
      </TD>

      <TD className="hidden text-sm lg:table-cell">
        {sudahDikemas ? (
          <>
            <p>
              {formatNumber((item.weight_gram ?? 0) / 1000)} kg · {formatNumber(item.total_qty)} pcs
            </p>
            <p className="text-xs text-muted-foreground">
              {item.courier} {item.service}
            </p>
          </>
        ) : (
          <span className="text-muted-foreground">belum ditimbang</span>
        )}
      </TD>

      <TD className="tabular hidden text-right lg:table-cell">
        {ongkirTerisi ? (
          formatIDR(item.shipping_fee)
        ) : (
          <span className="text-amber-600">belum diisi</span>
        )}
      </TD>

      <TD>
        {sudahDikirim ? (
          <div className="flex items-center gap-1.5">
            <span className="tabular font-mono text-sm">{item.tracking_number}</span>
            <Tooltip>
              <TooltipTrigger asChild>
                <a
                  href="https://www.jne.co.id/tracking-package"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-muted-foreground hover:text-foreground"
                >
                  <ExternalLink className="size-3.5" />
                  <span className="sr-only">Lacak paket</span>
                </a>
              </TooltipTrigger>
              <TooltipContent>Lacak paket</TooltipContent>
            </Tooltip>
          </div>
        ) : (
          <span className="text-sm text-muted-foreground">belum ada</span>
        )}
        {item.shipped_at && !item.customer_notified_at && (
          <p className="flex items-center gap-1 text-xs text-amber-600">
            <MessageCircle className="size-3" />
            belum dikabari
          </p>
        )}
      </TD>

      <TD className="text-right">
        <div className="flex items-center justify-end gap-1">
          {/* Label dicetak saat mengemas, jadi tombolnya duduk di baris ini. */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon-sm" asChild>
                <a href={labelUrl(item.order_id)} target="_blank" rel="noopener noreferrer">
                  <Printer />
                  <span className="sr-only">Cetak label</span>
                </a>
              </Button>
            </TooltipTrigger>
            <TooltipContent>Cetak label</TooltipContent>
          </Tooltip>

          {!sudahDikirim && (
            <Button size="sm" variant={ongkirTerisi ? "outline" : "default"} onClick={onKemas}>
              <Package />
              <span className="hidden sm:inline">{sudahDikemas ? "Ubah" : "Kemas"}</span>
              <span className="sr-only sm:hidden">Data kemasan {item.order_number}</span>
            </Button>
          )}

          {/*
           * Tombol resi baru muncul setelah lunas. Sebelum itu ia hanya akan
           * ditolak backend, dan tombol yang selalu gagal lebih membingungkan
           * daripada tombol yang belum ada.
           */}
          {!sudahDikirim && lunas && (
            <Button size="sm" onClick={onResi}>
              <Truck />
              <span className="hidden sm:inline">Kirim</span>
              <span className="sr-only sm:hidden">Input resi {item.order_number}</span>
            </Button>
          )}
        </div>
      </TD>
    </TR>
  );
}
