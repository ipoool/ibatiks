"use client";

import { CheckCircle2, ShoppingCart } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NumberInput } from "@/components/ui/number-input";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState } from "@/components/ui/page";
import { StatCard } from "@/components/ui/stat-card";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useRecordPurchase, type PurchasePayload } from "@/hooks/use-operations";
import { useShoppingList } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatIDR, formatNumber, todayInput, toNumber } from "@/lib/utils";
import type { ShoppingListEntry, Trip } from "@/types/api";

export function TripShopping({ trip }: { trip: Trip }) {
  const { data: list, isLoading, error } = useShoppingList(trip.id);
  const record = useRecordPurchase(trip.id);

  const [buying, setBuying] = useState<ShoppingListEntry | null>(null);
  const [form, setForm] = useState<PurchasePayload | null>(null);

  const totalOrdered = list?.reduce((sum, entry) => sum + entry.qty_ordered, 0) ?? 0;
  const totalPurchased = list?.reduce((sum, entry) => sum + entry.qty_purchased, 0) ?? 0;
  const totalRemaining = list?.reduce((sum, entry) => sum + entry.qty_remaining, 0) ?? 0;
  const totalAwaitingDP = list?.reduce((sum, entry) => sum + entry.qty_awaiting_dp, 0) ?? 0;
  const estimatedCost = list?.reduce((sum, entry) => sum + toNumber(entry.est_cost_idr), 0) ?? 0;

  function openBuy(entry: ShoppingListEntry) {
    setBuying(entry);
    setForm({
      product_id: entry.product_id,
      purchase_date: todayInput(),
      // Default-nya sisa yang belum terbeli — angka yang paling sering dipakai.
      qty: Math.max(entry.qty_remaining, 1),
      unit_cost_foreign: entry.cost_price,
      store_name: entry.store_name ?? "",
      notes: "",
    });
    record.reset();
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    if (!form) return;

    record.mutate(
      {
        ...form,
        qty: Number(form.qty),
        store_name: form.store_name || null,
        notes: form.notes || null,
      },
      {
        onSuccess: (result) => {
          toast.success("Pembelian dicatat", {
            description:
              result.qty_to_stock > 0
                ? `${result.qty_to_orders} unit untuk pesanan, ${result.qty_to_stock} unit masuk stok.`
                : `${result.qty_to_orders} unit dialokasikan ke pesanan customer.`,
          });
          setBuying(null);
        },
        onError: (err) => {
          toast.error(err instanceof ApiError ? err.message : "Gagal mencatat pembelian");
        },
      },
    );
  }

  return (
    <>
      <div className="space-y-1">
        <h2 className="text-lg font-semibold">Daftar belanja</h2>
        <p className="text-sm text-muted-foreground">
          Diringkas otomatis dari seluruh pesanan yang masuk. Kalau ada order yang diubah, daftar ini
          ikut menyesuaikan.
        </p>
      </div>

      <ErrorState error={error} />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Total dipesan" value={`${formatNumber(totalOrdered)} pcs`} />
        <StatCard label="Sudah dibeli" value={`${formatNumber(totalPurchased)} pcs`} tone="success" />
        <StatCard
          label="Sisa belanja"
          value={`${formatNumber(totalRemaining)} pcs`}
          tone={totalRemaining > 0 ? "warning" : "success"}
        />
        <StatCard label="Perkiraan modal" value={formatIDR(estimatedCost)} />
      </div>

      {/*
        Tanpa keterangan ini, tripper yang tahu ada pesanan masuk akan bingung
        melihat angka di sini lebih kecil daripada pesanan yang dia dengar.
      */}
      {totalAwaitingDP > 0 && (
        <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
          Ada {formatNumber(totalAwaitingDP)} pcs lagi dari order yang DP-nya belum masuk. Jumlah itu
          belum dihitung sebagai belanja — tunggu DP-nya diverifikasi dulu supaya tidak menalangi
          pembelian dengan uang toko.
        </p>
      )}

      <DataTable
        columns={7}
        isLoading={isLoading}
        isEmpty={!isLoading && (list?.length ?? 0) === 0}
        emptyTitle="Belum ada yang perlu dibeli"
        emptyDescription="Daftar ini terisi otomatis begitu ada order yang masuk pada trip ini."
        head={
          <TR>
            <TH>Produk</TH>
            <TH className="hidden lg:table-cell">Toko</TH>
            <TH className="hidden text-right sm:table-cell">Dipesan</TH>
            <TH className="hidden text-right lg:table-cell">Menunggu DP</TH>
            <TH className="hidden text-right sm:table-cell">Dibeli</TH>
            <TH className="text-right">Sisa</TH>
            <TH className="text-right">Aksi</TH>
          </TR>
        }
      >
        {list?.map((entry) => {
          const done = entry.qty_remaining === 0;
          return (
            <TR key={entry.product_id} className={done ? "opacity-60" : undefined}>
              <TD className="whitespace-normal">
                <div className="flex items-center gap-2">
                  <p className="font-medium">{entry.product_name}</p>
                  {done && (
                    <Badge variant="success">
                      <CheckCircle2 className="size-3" />
                      Lengkap
                    </Badge>
                  )}
                </div>
                <p className="text-xs text-muted-foreground">
                  {entry.product_sku} · dipesan {formatNumber(entry.order_count)} customer · modal{" "}
                  {formatNumber(entry.cost_price)} {trip.currency}
                </p>
                {/* Di ponsel yang tersisa hanya kolom "Sisa", jadi angka pesanan
                    dan yang sudah dibeli dituliskan di sini — tanpa keduanya
                    tripper tidak tahu sisa itu dari berapa. */}
                <p className="text-xs text-muted-foreground sm:hidden">
                  {formatNumber(entry.qty_purchased)} dari {formatNumber(entry.qty_ordered)} dibeli
                </p>
              </TD>
              <TD className="hidden text-sm text-muted-foreground lg:table-cell">
                {entry.store_name ?? "—"}
              </TD>
              <TD className="tabular hidden text-right font-medium sm:table-cell">
                {formatNumber(entry.qty_ordered)}
              </TD>
              {/* Permintaan yang DP-nya belum masuk sengaja dipisah: angkanya
                  berguna sebagai ancang-ancang, tapi belum boleh dibelanjakan. */}
              <TD
                className={`tabular hidden text-right lg:table-cell ${
                  entry.qty_awaiting_dp > 0 ? "text-amber-600" : "text-muted-foreground"
                }`}
              >
                {formatNumber(entry.qty_awaiting_dp)}
              </TD>
              <TD className="tabular hidden text-right text-emerald-600 sm:table-cell">
                {formatNumber(entry.qty_purchased)}
              </TD>
              <TD
                className={`tabular text-right font-semibold ${
                  entry.qty_remaining > 0 ? "text-amber-600" : "text-muted-foreground"
                }`}
              >
                {formatNumber(entry.qty_remaining)}
              </TD>
              <TD className="text-right">
                {/* Di ponsel tombolnya tinggal ikon: label "Catat Beli" memakan
                    lebar yang lebih berguna untuk nama produk. */}
                <Button size="sm" variant={done ? "outline" : "default"} onClick={() => openBuy(entry)}>
                  <ShoppingCart />
                  <span className="hidden sm:inline">Catat Beli</span>
                  <span className="sr-only sm:hidden">Catat pembelian {entry.product_name}</span>
                </Button>
              </TD>
            </TR>
          );
        })}
      </DataTable>

      {buying && form && (
        <FormDialog
          open
          onOpenChange={(open) => !open && setBuying(null)}
          title="Catat Pembelian"
          description={`${buying.product_name} — sistem otomatis membagi ke pesanan customer, sisanya masuk stok.`}
          error={record.error}
          loading={record.isPending}
          onSubmit={handleSubmit}
          submitLabel="Catat Pembelian"
        >
          <div className="grid gap-4 sm:grid-cols-2">
            <Field
              label="Jumlah dibeli"
              htmlFor="qty"
              required
              hint={`Dipesan ${buying.qty_ordered} pcs, sisa ${buying.qty_remaining} pcs`}
            >
              <NumberInput
                id="qty"
                min="1"
                value={form.qty}
                onValueChange={(qty) => setForm({ ...form, qty })}
                required
                autoFocus
              />
            </Field>

            <Field
              label={`Harga satuan (${trip.currency})`}
              htmlFor="unit_cost_foreign"
              required
              hint="Harga yang benar-benar dibayar di toko"
            >
              <Input
                id="unit_cost_foreign"
                type="number"
                min="0"
                step="any"
                value={form.unit_cost_foreign}
                onChange={(event) => setForm({ ...form, unit_cost_foreign: event.target.value })}
                required
              />
            </Field>

            <Field label="Tanggal beli" htmlFor="purchase_date">
              <Input
                id="purchase_date"
                type="date"
                value={form.purchase_date ?? ""}
                onChange={(event) => setForm({ ...form, purchase_date: event.target.value })}
              />
            </Field>

            <Field
              label="Kurs khusus"
              htmlFor="exchange_rate"
              hint={`Kosongkan untuk pakai kurs trip (${formatNumber(trip.exchange_rate)})`}
            >
              <Input
                id="exchange_rate"
                type="number"
                min="0"
                step="any"
                value={form.exchange_rate ?? ""}
                onChange={(event) => setForm({ ...form, exchange_rate: event.target.value })}
                placeholder={trip.exchange_rate}
              />
            </Field>

            <Field label="Toko" htmlFor="store_name">
              <Input
                id="store_name"
                value={form.store_name ?? ""}
                onChange={(event) => setForm({ ...form, store_name: event.target.value })}
                placeholder="Don Quijote Shibuya"
              />
            </Field>

            <Field label="Catatan" htmlFor="purchase_notes">
              <Textarea
                id="purchase_notes"
                rows={2}
                value={form.notes ?? ""}
                onChange={(event) => setForm({ ...form, notes: event.target.value })}
                placeholder="Varian, promo, atau catatan lain"
              />
            </Field>
          </div>

          <Card className="bg-muted/50 py-0">
            <CardContent className="flex items-center justify-between py-3 text-sm">
              <span className="text-muted-foreground">Perkiraan total belanja</span>
              <span className="tabular font-semibold">
                {formatIDR(
                  Number(form.qty || 0) *
                    toNumber(form.unit_cost_foreign) *
                    toNumber(form.exchange_rate || trip.exchange_rate),
                )}
              </span>
            </CardContent>
          </Card>
        </FormDialog>
      )}
    </>
  );
}
