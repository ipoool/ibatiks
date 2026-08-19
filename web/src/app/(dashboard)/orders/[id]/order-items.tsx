"use client";

import { Check, Package, Pencil, Plus, Trash2, X } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { FulfillmentBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";

import { OrderReceiveButton } from "./order-actions";
import { Combobox } from "@/components/ui/combobox";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ConfirmDialog, FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NumberInput } from "@/components/ui/number-input";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useAddOrderItem, useDeleteOrderItem, useUpdateOrderItem } from "@/hooks/use-orders";
import { useTripItems } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatIDR, formatNumber } from "@/lib/utils";
import type { OrderDetail, OrderItem } from "@/types/api";

export function OrderItems({ order }: { order: OrderDetail }) {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editQty, setEditQty] = useState(0);
  const [editPrice, setEditPrice] = useState("");
  const [deleting, setDeleting] = useState<OrderItem | null>(null);
  const [addOpen, setAddOpen] = useState(false);
  const [addForm, setAddForm] = useState({ product_id: "", qty: 1 });

  const { data: catalog, isLoading: loadingCatalog } = useTripItems(
    addOpen ? order.trip_id : undefined,
  );
  const catalogOptions =
    catalog
      ?.filter((item) => item.is_active)
      .map((item) => ({
        value: item.product_id,
        label: item.product_name,
        keywords: item.product_sku,
        description: `${item.product_sku} · ${formatIDR(item.sell_price)}`,
      })) ?? [];
  const catalogEmpty = !loadingCatalog && catalogOptions.length === 0;
  const updateItem = useUpdateOrderItem(order.id);
  const deleteItem = useDeleteOrderItem(order.id);
  const addItem = useAddOrderItem(order.id);

  function startEdit(item: OrderItem) {
    setEditingId(item.id);
    setEditQty(item.qty);
    setEditPrice(item.unit_price);
  }

  function saveEdit(item: OrderItem) {
    updateItem.mutate(
      { itemId: item.id, qty: editQty, unit_price: editPrice },
      {
        onSuccess: (updated) => {
          setEditingId(null);
          toast.success("Item diperbarui", {
            description: `Total order sekarang ${formatIDR(updated.total)}.`,
          });
        },
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal mengubah item");
        },
      },
    );
  }

  function handleDelete() {
    if (!deleting) return;
    deleteItem.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Item dihapus dari order");
        setDeleting(null);
      },
    });
  }

  function handleAdd(event: React.FormEvent) {
    event.preventDefault();
    addItem.mutate(
      { product_id: addForm.product_id, qty: Number(addForm.qty) },
      {
        onSuccess: () => {
          toast.success("Produk ditambahkan ke order");
          setAddOpen(false);
          setAddForm({ product_id: "", qty: 1 });
        },
      },
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Item pesanan</CardTitle>
        <CardAction>
          <div className="flex flex-wrap items-center gap-2">
            <OrderReceiveButton order={order} />
            {order.editable && (
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  addItem.reset();
                  setAddOpen(true);
                }}
              >
                <Plus />
                Tambah Produk
              </Button>
            )}
          </div>
        </CardAction>
      </CardHeader>

      <CardContent>
        <DataTable
          columns={5}
          isEmpty={order.items.length === 0}
          emptyTitle="Belum ada item"
          head={
            <TR>
              {/* Kolom produk diberi lebar minimum supaya nama panjang tidak
                  terpecah per huruf saat panel ringkasan mempersempit tabel.
                  Status pemenuhan ikut di kolom ini, bukan kolom sendiri, agar
                  tabel tetap muat di samping panel ringkasan biaya. */}
              <TH className="sm:min-w-40">Produk</TH>
              <TH className="w-12 text-right sm:w-14">Qty</TH>
              <TH className="w-20 text-right sm:w-24">Harga</TH>
              <TH className="hidden w-28 text-right sm:table-cell">Subtotal</TH>
              <TH className="w-16 text-right">Aksi</TH>
            </TR>
          }
        >
          {order.items.map((item) => {
            const editing = editingId === item.id;

            return (
              <TR key={item.id}>
                {/* Nama produk boleh turun baris di layar sempit; dipaksa satu
                    baris, ia mendorong kolom harga dan tombol keluar layar. */}
                <TD className="whitespace-normal">
                  <p className="font-medium">{item.product_name}</p>
                  <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1">
                    <FulfillmentBadge status={item.fulfillment_status} />
                    <span className="text-xs text-muted-foreground">
                      {item.product_sku}
                      {item.qty_purchased > 0 && ` · dibeli ${formatNumber(item.qty_purchased)}`}
                      {item.qty_received > 0 && ` · diterima ${formatNumber(item.qty_received)}`}
                    </span>
                  </div>
                </TD>

                <TD className="text-right">
                  {editing ? (
                    <NumberInput
                      min={Math.max(item.qty_received, 1)}
                      value={editQty}
                      onValueChange={setEditQty}
                      className="h-9 text-right"
                      autoFocus
                    />
                  ) : (
                    <span className="tabular">{formatNumber(item.qty)}</span>
                  )}
                </TD>

                <TD className="text-right">
                  {editing ? (
                    <Input
                      type="number"
                      min="0"
                      step="any"
                      value={editPrice}
                      onChange={(event) => setEditPrice(event.target.value)}
                      className="h-9 text-right"
                    />
                  ) : (
                    <>
                      <span className="tabular">{formatIDR(item.unit_price)}</span>
                      <p className="tabular text-xs font-medium sm:hidden">
                        {formatIDR(item.subtotal)}
                      </p>
                    </>
                  )}
                </TD>

                {/* Subtotal disembunyikan di ponsel, bukan harga satuan: harga
                    satuan bisa disunting langsung di kolomnya, sedangkan subtotal
                    hanyalah hasil kali yang bisa dituliskan sebagai keterangan. */}
                <TD className="tabular hidden text-right font-medium sm:table-cell">
                  {formatIDR(item.subtotal)}
                </TD>

                <TD>
                  <div className="flex justify-end gap-1">
                    {editing ? (
                      <>
                        <Button
                          size="icon-sm"
                          tooltip="Simpan"
                          onClick={() => saveEdit(item)}
                          loading={updateItem.isPending}
                        >
                          <Check />
                          <span className="sr-only">Simpan</span>
                        </Button>
                        <Button variant="ghost" size="icon-sm" onClick={() => setEditingId(null)} tooltip="Batal">
                          <X />
                          <span className="sr-only">Batal</span>
                        </Button>
                      </>
                    ) : (
                      order.editable && (
                        <>
                          <Button variant="ghost" size="icon-sm" onClick={() => startEdit(item)} tooltip="Ubah">
                            <Pencil />
                            <span className="sr-only">Ubah</span>
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            tooltip="Hapus"
                            className="text-destructive hover:text-destructive"
                            onClick={() => setDeleting(item)}
                          >
                            <Trash2 />
                            <span className="sr-only">Hapus</span>
                          </Button>
                        </>
                      )
                    )}
                  </div>
                </TD>
              </TR>
            );
          })}
        </DataTable>

        {!order.editable && (
          <p className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
            <Package className="size-3.5" />
            {/* Alasannya disebut apa adanya: order batal tidak pernah
                diserahkan ke kurir, dan menuliskannya begitu membuat admin
                mengira ada paket yang terlanjur jalan. */}
            {order.status === "cancelled"
              ? "Order dibatalkan, isinya dibekukan sebagai catatan."
              : "Order sudah diserahkan ke kurir, isinya tidak bisa diubah lagi."}
          </p>
        )}
      </CardContent>

      <FormDialog
        open={addOpen}
        onOpenChange={setAddOpen}
        title="Tambah Produk ke Order"
        description="Harga diambil dari katalog trip. Produk yang sudah ada akan ditambah jumlahnya."
        error={addItem.error}
        loading={addItem.isPending}
        // Tanpa ini tombol Simpan bisa ditekan dengan produk kosong, dan yang
        // muncul adalah penolakan dari server — padahal formnya sendiri sudah
        // tahu isiannya belum lengkap.
        submitDisabled={!addForm.product_id || addForm.qty < 1}
        onSubmit={handleAdd}
      >
        <div className="grid gap-4 sm:grid-cols-3">
          <Field
            label="Produk"
            htmlFor="add_product_id"
            required
            className="sm:col-span-2"
            hint={
              catalogEmpty
                ? "Katalog trip ini masih kosong — isi dulu lewat tab Katalog pada halaman trip."
                : undefined
            }
          >
            <Combobox
              id="add_product_id"
              value={addForm.product_id}
              onChange={(value) => setAddForm({ ...addForm, product_id: value })}
              options={catalogOptions}
              isLoading={loadingCatalog}
              disabled={catalogEmpty}
              placeholder={catalogEmpty ? "Katalog trip masih kosong" : "Pilih produk…"}
              searchPlaceholder="Cari nama produk atau SKU…"
              emptyLabel="Produk tidak ada di katalog trip ini"
            />
          </Field>

          <Field label="Jumlah" htmlFor="add_qty" required>
            <NumberInput
              id="add_qty"
              min="1"
              value={addForm.qty}
              onValueChange={(qty) => setAddForm({ ...addForm, qty })}
              required
            />
          </Field>
        </div>
      </FormDialog>

      <ConfirmDialog
        open={Boolean(deleting)}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Hapus item dari order?"
        description={`${deleting?.product_name ?? ""} akan dihapus dan total order dihitung ulang. Barang yang sudah terlanjur dibeli untuk item ini dipindahkan ke stok.`}
        confirmLabel="Hapus item"
        loading={deleteItem.isPending}
        error={deleteItem.error}
        onConfirm={handleDelete}
      />
    </Card>
  );
}
