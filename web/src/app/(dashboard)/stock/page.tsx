"use client";

import { History, ShoppingBag, SlidersHorizontal } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { CheckboxField } from "@/components/ui/checkbox-field";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NumberInput } from "@/components/ui/number-input";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { StatCard } from "@/components/ui/stat-card";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useDebounced } from "@/hooks/use-debounced";
import {
  useAdjustStock,
  useSellStock,
  useStock,
  useStockMovements,
} from "@/hooks/use-operations";
import { ApiError } from "@/lib/api";
import { formatDate, formatIDR, formatNumber, toNumber } from "@/lib/utils";
import type { StockItem, StockMovementType } from "@/types/api";

const MOVEMENT_LABEL: Record<StockMovementType, { label: string; tone: "success" | "info" | "warning" | "neutral" }> = {
  in_purchase: { label: "Masuk dari belanja", tone: "success" },
  out_order: { label: "Dipakai pesanan", tone: "info" },
  out_marketplace: { label: "Terjual marketplace", tone: "warning" },
  adjustment: { label: "Penyesuaian", tone: "neutral" },
};

export default function StockPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [inStockOnly, setInStockOnly] = useState(true);
  const [selling, setSelling] = useState<StockItem | null>(null);
  const [adjusting, setAdjusting] = useState<StockItem | null>(null);
  const debouncedSearch = useDebounced(search);

  const { data, isLoading, error } = useStock({
    page,
    q: debouncedSearch,
    in_stock_only: inStockOnly || undefined,
  });

  const totalValue = data?.items.reduce((sum, item) => sum + toNumber(item.stock_value), 0) ?? 0;
  const totalQty = data?.items.reduce((sum, item) => sum + item.qty_on_hand, 0) ?? 0;

  return (
    <>
      <PageHeader
        title="Stok"
        description="Barang sisa belanja yang tidak dipesan siapa pun, siap dijual di marketplace"
      />

      <ErrorState error={error} />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <StatCard
          label="Nilai stok (halaman ini)"
          value={formatIDR(totalValue)}
          isLoading={isLoading}
        />
        <StatCard label="Jumlah unit" value={`${formatNumber(totalQty)} pcs`} isLoading={isLoading} />
        <StatCard
          label="Jenis produk"
          value={formatNumber(data?.meta.total ?? 0)}
          isLoading={isLoading}
        />
      </div>

      <Tabs defaultValue="posisi">
        <TabsList>
          <TabsTrigger value="posisi">Posisi Stok</TabsTrigger>
          <TabsTrigger value="riwayat">
            <History className="size-4" />
            Riwayat Pergerakan
          </TabsTrigger>
        </TabsList>

        <TabsContent value="posisi">
          <div className="flex flex-wrap gap-3">
            <SearchInput
              value={search}
              onChange={(value) => {
                setSearch(value);
                setPage(1);
              }}
              placeholder="Cari produk atau SKU…"
              className="min-w-64 flex-1 sm:max-w-md"
            />
            <CheckboxField
              id="in_stock_only"
              variant="boxed"
              checked={inStockOnly}
              onCheckedChange={(checked) => {
                setInStockOnly(checked);
                setPage(1);
              }}
            >
              Hanya yang masih ada
            </CheckboxField>
          </div>

          <div>
            <DataTable
              columns={5}
              isLoading={isLoading}
              isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
              emptyTitle="Stok kosong"
              emptyDescription="Stok terisi otomatis ketika tripper membeli lebih banyak dari yang dipesan."
              head={
                <TR>
                  <TH>Produk</TH>
                  <TH className="text-right">Qty</TH>
                  <TH className="hidden text-right sm:table-cell">HPP rata-rata</TH>
                  <TH className="text-right">Nilai stok</TH>
                  <TH className="text-right">Aksi</TH>
                </TR>
              }
            >
              {data?.items.map((item) => (
                <TR key={item.id}>
                  <TD className="whitespace-normal">
                    <p className="font-medium">{item.product_name}</p>
                    <p className="text-xs text-muted-foreground">
                      {item.product_sku}
                      {item.category_name ? ` · ${item.category_name}` : ""}
                    </p>
                  </TD>
                  <TD className="tabular text-right font-semibold">
                    {formatNumber(item.qty_on_hand)}
                  </TD>
                  <TD className="tabular hidden text-right sm:table-cell">
                    {formatIDR(item.avg_cost_idr)}
                  </TD>
                  <TD className="tabular text-right font-medium">{formatIDR(item.stock_value)}</TD>
                  <TD>
                    <div className="flex justify-end gap-1">
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={item.qty_on_hand === 0}
                        onClick={() => setSelling(item)}
                      >
                        <ShoppingBag />
                        <span className="hidden sm:inline">Jual</span>
                        <span className="sr-only sm:hidden">Jual {item.product_name}</span>
                      </Button>
                      <Button variant="ghost" size="icon-sm" onClick={() => setAdjusting(item)} tooltip="Sesuaikan">
                        <SlidersHorizontal />
                        <span className="sr-only">Sesuaikan</span>
                      </Button>
                    </div>
                  </TD>
                </TR>
              ))}
            </DataTable>

            <Pagination meta={data?.meta} onPageChange={setPage} />
          </div>
        </TabsContent>

        <TabsContent value="riwayat">
          <MovementHistory />
        </TabsContent>
      </Tabs>

      {selling && <SellDialog item={selling} onClose={() => setSelling(null)} />}
      {adjusting && <AdjustDialog item={adjusting} onClose={() => setAdjusting(null)} />}
    </>
  );
}

function MovementHistory() {
  const [page, setPage] = useState(1);
  const { data, isLoading, error } = useStockMovements({ page });

  return (
    <>
      <ErrorState error={error} />
      <div>
        <DataTable
          columns={5}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada pergerakan stok"
          head={
            <TR>
              <TH>Tanggal</TH>
              <TH>Produk</TH>
              <TH>Jenis</TH>
              <TH className="text-right">Qty</TH>
              <TH className="text-right">Nilai</TH>
            </TR>
          }
        >
          {data?.items.map((movement) => {
            const meta = MOVEMENT_LABEL[movement.type];
            return (
              <TR key={movement.id}>
                <TD className="whitespace-nowrap text-sm">{formatDate(movement.created_at)}</TD>
                <TD>
                  <p className="text-sm font-medium">{movement.product_name}</p>
                  {movement.note && (
                    <p className="text-xs text-muted-foreground">{movement.note}</p>
                  )}
                </TD>
                <TD>
                  <Badge variant={meta.tone}>{meta.label}</Badge>
                </TD>
                <TD
                  className={`tabular text-right font-medium ${
                    movement.qty > 0 ? "text-emerald-600" : "text-red-600"
                  }`}
                >
                  {movement.qty > 0 ? "+" : ""}
                  {formatNumber(movement.qty)}
                </TD>
                <TD className="tabular text-right text-muted-foreground">
                  {movement.sale_price_idr
                    ? formatIDR(movement.sale_price_idr)
                    : formatIDR(movement.unit_cost_idr)}
                </TD>
              </TR>
            );
          })}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}

function SellDialog({ item, onClose }: { item: StockItem; onClose: () => void }) {
  const sell = useSellStock();
  const [form, setForm] = useState({
    qty: 1,
    sale_price: "",
    channel: "Shopee",
    note: "",
  });

  const margin = toNumber(form.sale_price) - toNumber(item.avg_cost_idr);

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    sell.mutate(
      {
        product_id: item.product_id,
        qty: Number(form.qty),
        sale_price: form.sale_price || "0",
        channel: form.channel,
        note: form.note || null,
      },
      {
        onSuccess: () => {
          toast.success("Penjualan stok dicatat");
          onClose();
        },
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal mencatat penjualan");
        },
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title="Catat Penjualan Stok"
      description={`${item.product_name} — tersedia ${formatNumber(item.qty_on_hand)} pcs, HPP ${formatIDR(item.avg_cost_idr)}/pcs.`}
      error={sell.error}
      loading={sell.isPending}
      onSubmit={handleSubmit}
      submitLabel="Catat Penjualan"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Jumlah terjual" htmlFor="sell_qty" required>
          <NumberInput
            id="sell_qty"
            min="1"
            max={item.qty_on_hand}
            value={form.qty}
            onValueChange={(qty) => setForm({ ...form, qty })}
            required
            autoFocus
          />
        </Field>

        <Field
          label="Harga jual per pcs (Rp)"
          htmlFor="sale_price"
          required
          hint={margin !== 0 ? `Margin ${formatIDR(margin)}/pcs` : undefined}
        >
          <Input
            id="sale_price"
            type="number"
            min="0"
            step="any"
            value={form.sale_price}
            onChange={(event) => setForm({ ...form, sale_price: event.target.value })}
            required
          />
        </Field>

        <Field label="Kanal penjualan" htmlFor="channel">
          <Input
            id="channel"
            value={form.channel}
            onChange={(event) => setForm({ ...form, channel: event.target.value })}
            placeholder="Shopee, Tokopedia, Instagram…"
          />
        </Field>

        <Field label="Catatan" htmlFor="sell_note">
          <Textarea
            id="sell_note"
            rows={2}
            value={form.note}
            onChange={(event) => setForm({ ...form, note: event.target.value })}
          />
        </Field>
      </div>
    </FormDialog>
  );
}

function AdjustDialog({ item, onClose }: { item: StockItem; onClose: () => void }) {
  const adjust = useAdjustStock();
  const [form, setForm] = useState({ new_qty: item.qty_on_hand, note: "" });

  const delta = Number(form.new_qty) - item.qty_on_hand;

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    adjust.mutate(
      {
        product_id: item.product_id,
        new_qty: Number(form.new_qty),
        note: form.note || null,
      },
      {
        onSuccess: () => {
          toast.success("Stok disesuaikan");
          onClose();
        },
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title="Sesuaikan Stok"
      description={`${item.product_name} — tercatat ${formatNumber(item.qty_on_hand)} pcs. Isi jumlah hasil hitung fisik.`}
      error={adjust.error}
      loading={adjust.isPending}
      onSubmit={handleSubmit}
      submitLabel="Simpan Penyesuaian"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field
          label="Jumlah sebenarnya"
          htmlFor="new_qty"
          required
          hint={delta !== 0 ? `Selisih ${delta > 0 ? "+" : ""}${delta} pcs` : "Tidak ada selisih"}
        >
          <NumberInput
            id="new_qty"
            min="0"
            value={form.new_qty}
            onValueChange={(newQty) => setForm({ ...form, new_qty: newQty })}
            required
            autoFocus
          />
        </Field>

        <Field label="Alasan" htmlFor="adjust_note">
          <Input
            id="adjust_note"
            value={form.note}
            onChange={(event) => setForm({ ...form, note: event.target.value })}
            placeholder="Stock opname, barang rusak, dan lain-lain"
          />
        </Field>
      </div>
    </FormDialog>
  );
}
