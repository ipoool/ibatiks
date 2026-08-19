"use client";

import Link from "next/link";
import { ChevronDown, ChevronRight, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { FilterSelect } from "@/components/filter-select";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/form-dialog";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useDebounced } from "@/hooks/use-debounced";
import { useDeletePurchase, usePurchaseAllocations, usePurchases } from "@/hooks/use-operations";
import { useTrips } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatDate, formatIDR, formatNumber } from "@/lib/utils";
import type { Purchase } from "@/types/api";

export default function PurchasesPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [tripId, setTripId] = useState("");
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<Purchase | null>(null);
  const debouncedSearch = useDebounced(search);

  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];
  const { data, isLoading, error } = usePurchases({
    page,
    q: debouncedSearch,
    trip_id: tripId || undefined,
  });
  const remove = useDeletePurchase();

  function handleDelete() {
    if (!deleting) return;
    remove.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Pembelian dihapus, stok dan alokasi dikembalikan");
        setDeleting(null);
      },
      onError: (err) => {
        toast.error(err instanceof ApiError ? err.message : "Gagal menghapus pembelian");
      },
    });
  }

  return (
    <>
      <PageHeader
        title="Pembelian"
        description="Realisasi belanja tripper di lapangan beserta ke mana tiap unitnya dialokasikan"
      />

      <ErrorState error={error} />

      <div className="flex flex-wrap gap-3">
        <SearchInput
          value={search}
          onChange={(value) => {
            setSearch(value);
            setPage(1);
          }}
          placeholder="Cari produk atau toko…"
          className="min-w-64 flex-1 sm:max-w-md"
        />
        <FilterSelect
          value={tripId}
          onChange={(value) => {
            setTripId(value);
            setPage(1);
          }}
          allLabel="Semua trip"
          options={tripOptions}
          className="sm:w-64"
        />
      </div>

      <div>
        <DataTable
          columns={7}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada pembelian"
          emptyDescription="Catat belanja lewat halaman Daftar Belanja saat tripper sedang di toko."
          head={
            <TR>
              <TH className="w-10" />
              <TH>Produk</TH>
              <TH className="hidden lg:table-cell">Tanggal</TH>
              <TH className="text-right">Qty</TH>
              <TH className="hidden text-right sm:table-cell">Modal/pcs</TH>
              <TH className="text-right">Total</TH>
              <TH className="text-right">Aksi</TH>
            </TR>
          }
        >
          {data?.items.map((purchase) => (
            <PurchaseRow
              key={purchase.id}
              purchase={purchase}
              expanded={expandedId === purchase.id}
              onToggle={() => setExpandedId(expandedId === purchase.id ? null : purchase.id)}
              onDelete={() => setDeleting(purchase)}
            />
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>

      <ConfirmDialog
        open={Boolean(deleting)}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Hapus pembelian ini?"
        description="Alokasi ke pesanan dilepas dan stok yang sempat bertambah akan ditarik kembali."
        confirmLabel="Hapus"
        loading={remove.isPending}
        error={remove.error}
        onConfirm={handleDelete}
      />
    </>
  );
}

function PurchaseRow({
  purchase,
  expanded,
  onToggle,
  onDelete,
}: {
  purchase: Purchase;
  expanded: boolean;
  onToggle: () => void;
  onDelete: () => void;
}) {
  // Alokasi hanya diambil ketika barisnya dibuka, supaya daftar tetap ringan.
  const { data: allocations, isLoading } = usePurchaseAllocations(expanded ? purchase.id : undefined);

  return (
    <>
      <TR>
        <TD>
          <Button variant="ghost" size="icon-sm" onClick={onToggle} tooltip="Lihat alokasi">
            {expanded ? <ChevronDown /> : <ChevronRight />}
            <span className="sr-only">Lihat alokasi</span>
          </Button>
        </TD>
        <TD className="whitespace-normal">
          <p className="font-medium">{purchase.product_name}</p>
          <p className="text-xs text-muted-foreground">
            {purchase.product_sku}
            {purchase.store_name ? ` · ${purchase.store_name}` : ""}
          </p>
          {/* Tanggal belanja menyusul nama produk saat kolomnya disembunyikan:
              satu produk bisa dibeli berkali-kali dalam satu trip. */}
          <p className="text-xs text-muted-foreground lg:hidden">
            {formatDate(purchase.purchase_date)}
          </p>
        </TD>
        <TD className="hidden whitespace-nowrap text-sm lg:table-cell">
          {formatDate(purchase.purchase_date)}
        </TD>
        <TD className="text-right">
          <span className="tabular font-medium">{formatNumber(purchase.qty)}</span>
          {/* Rincian alokasi memakan lebar yang tidak ada di ponsel; angkanya
              tetap bisa dilihat dengan membuka baris ini. */}
          <div className="mt-0.5 hidden justify-end gap-1 sm:flex">
            {(purchase.qty_to_orders ?? 0) > 0 && (
              <Badge variant="info">{purchase.qty_to_orders} pesanan</Badge>
            )}
            {(purchase.qty_to_stock ?? 0) > 0 && (
              <Badge variant="warning">{purchase.qty_to_stock} stok</Badge>
            )}
          </div>
        </TD>
        <TD className="tabular hidden text-right sm:table-cell">
          {formatIDR(purchase.unit_cost_idr)}
          <p className="text-xs text-muted-foreground">
            {formatNumber(purchase.unit_cost_foreign)} {purchase.currency}
          </p>
        </TD>
        <TD className="tabular text-right font-medium">{formatIDR(purchase.total_cost_idr)}</TD>
        <TD>
          <div className="flex justify-end">
            <Button
              variant="ghost"
              size="icon-sm"
              tooltip="Hapus"
              className="text-destructive hover:text-destructive"
              onClick={onDelete}
            >
              <Trash2 />
              <span className="sr-only">Hapus</span>
            </Button>
          </div>
        </TD>
      </TR>

      {expanded && (
        <TR className="bg-muted/30 hover:bg-muted/30">
          <TD colSpan={7} className="py-3">
            <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Alokasi unit
            </p>
            {isLoading ? (
              <p className="text-sm text-muted-foreground">Memuat…</p>
            ) : (
              <div className="space-y-1">
                {allocations?.map((allocation) => (
                  <div
                    key={allocation.id}
                    className="flex items-center justify-between rounded-md bg-card px-3 py-2 text-sm"
                  >
                    <span>
                      {allocation.order_number ? (
                        <>
                          {/* Ditautkan ke ordernya: daftar ini yang menjawab
                              "unit ini buat siapa", dan pertanyaan berikutnya
                              hampir selalu soal isi ordernya. */}
                          {allocation.order_id ? (
                            <Link
                              href={`/orders/${allocation.order_id}`}
                              className="font-medium hover:underline"
                            >
                              {allocation.order_number}
                            </Link>
                          ) : (
                            <span className="font-medium">{allocation.order_number}</span>
                          )}
                          <span className="text-muted-foreground"> — {allocation.customer_name}</span>
                        </>
                      ) : (
                        <span className="text-amber-700">Masuk stok (tidak dipesan siapa pun)</span>
                      )}
                    </span>
                    <span className="tabular font-medium">{formatNumber(allocation.qty)} pcs</span>
                  </div>
                ))}
              </div>
            )}
          </TD>
        </TR>
      )}
    </>
  );
}
