"use client";

import { Calculator, Pencil, Plus, RefreshCw, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { CheckboxField } from "@/components/ui/checkbox-field";
import { Combobox } from "@/components/ui/combobox";
import { OptionSelect } from "@/components/filter-select";
import { LastPriceHint } from "@/components/price-history";
import { QuickAddProductDialog } from "@/components/quick-add-product";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { ConfirmDialog, FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { ErrorState } from "@/components/ui/page";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useProducts } from "@/hooks/use-master";
import {
  useDeleteTripItem,
  useRecalculatePrices,
  useSaveTripItem,
  useTripItems,
  type TripItemPayload,
} from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatIDR, formatNumber, toNumber } from "@/lib/utils";
import type { MarkupType, Product, Trip, TripItem } from "@/types/api";

const EMPTY_FORM: TripItemPayload = {
  product_id: "",
  cost_price: "",
  markup_type: "percent",
  markup_value: "",
  max_qty: null,
  is_active: true,
  notes: "",
};

const MARKUP_OPTIONS: ReadonlyArray<{ value: MarkupType; label: string }> = [
  { value: "percent", label: "Persen (%)" },
  { value: "nominal", label: "Nominal (Rp)" },
];

export function TripCatalog({ trip }: { trip: Trip }) {
  const { data: items, isLoading, error } = useTripItems(trip.id);
  const { data: products } = useProducts({ per_page: 200, active_only: true });

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<TripItem | null>(null);
  const [form, setForm] = useState<TripItemPayload>(EMPTY_FORM);
  const [quickAddOpen, setQuickAddOpen] = useState(false);
  const [deleting, setDeleting] = useState<TripItem | null>(null);
  const [recalcOpen, setRecalcOpen] = useState(false);

  const save = useSaveTripItem(trip.id, editing?.id);
  const remove = useDeleteTripItem(trip.id);
  const recalculate = useRecalculatePrices(trip.id);

  // Harga jual dihitung ulang di sisi klien memakai rumus yang sama dengan
  // backend, supaya admin langsung melihat hasilnya sambil mengetik.
  const previewCostIDR = Math.round(toNumber(form.cost_price) * toNumber(trip.exchange_rate));
  const previewSell = (() => {
    const markup = toNumber(form.markup_value);
    const raw =
      form.markup_type === "percent"
        ? previewCostIDR + (previewCostIDR * markup) / 100
        : previewCostIDR + markup;
    return raw === 0 ? 0 : Math.ceil(raw / 100) * 100;
  })();

  function openCreate() {
    setEditing(null);
    setForm(EMPTY_FORM);
    save.reset();
    setFormOpen(true);
  }

  function openEdit(item: TripItem) {
    setEditing(item);
    setForm({
      product_id: item.product_id,
      cost_price: item.cost_price,
      markup_type: item.markup_type,
      markup_value: item.markup_value,
      max_qty: item.max_qty,
      is_active: item.is_active,
      notes: item.notes ?? "",
    });
    save.reset();
    setFormOpen(true);
  }

  // Saat produk dipilih, markup diisikan dari master produk sebagai titik awal
  // yang masuk akal. Dikerjakan di handler pilihan, bukan lewat efek, supaya
  // angka yang sudah disesuaikan admin tidak tertimpa ulang.
  //
  // Harga modal hanya ikut terisi kalau mata uang master sama dengan mata uang
  // trip ini. Produk yang sama dibeli di negara berbeda punya harga yang tidak
  // sebanding: mengisi 880 JPY ke trip berkurs KRW akan terlihat wajar padahal
  // salah tiga digit, dan kesalahan seperti itu baru ketahuan setelah harga
  // jualnya terlanjur diumumkan.
  /*
   * Produk yang baru dibuat disimpan sebentar di sini.
   *
   * Daftar produk dari server baru menyusul beberapa saat setelah produknya
   * tersimpan, sementara pilihannya harus ditetapkan sekarang juga. Radix
   * Select yang menerima nilai tanpa opsi yang cocok menampilkan placeholder
   * dan tidak memperbaruinya lagi walau opsinya belakangan datang — jadi
   * opsinya disediakan sendiri, tidak menunggu jaringan.
   */
  const [produkBaru, setProdukBaru] = useState<Product | null>(null);

  function pilihProdukBaru(product: Product) {
    setProdukBaru(product);
    setForm((current) => ({
      ...current,
      product_id: product.id,
      // Produk baru belum punya harga modal; yang berlaku untuk trip ini
      // diketik admin di dialog katalog ini juga.
      cost_price: "",
      markup_type: product.markup_type,
      markup_value: product.markup_value,
    }));
  }

  function selectProduct(productId: string) {
    const product = products?.items.find((candidate) => candidate.id === productId);
    const sameCurrency = product?.base_currency === trip.currency;

    setForm((current) => ({
      ...current,
      product_id: productId,
      ...(product
        ? {
            cost_price: sameCurrency ? product.base_price : "",
            markup_type: product.markup_type,
            markup_value: product.markup_value,
          }
        : {}),
    }));
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    save.mutate(
      {
        ...form,
        // Kolom angka dibiarkan kosong selagi diketik supaya angka 0 tidak
        // menempel di depan ketikan; nilainya baru dijadikan "0" di sini.
        cost_price: form.cost_price || "0",
        markup_value: form.markup_value || "0",
        max_qty: form.max_qty ? Number(form.max_qty) : null,
        notes: form.notes || null,
      },
      {
        onSuccess: () => {
          toast.success(editing ? "Item katalog diperbarui" : "Produk ditambahkan ke katalog");
          setFormOpen(false);
        },
      },
    );
  }

  function handleDelete() {
    if (!deleting) return;
    remove.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Produk dikeluarkan dari katalog");
        setDeleting(null);
      },
    });
  }

  function handleRecalculate() {
    recalculate.mutate(undefined, {
      onSuccess: (updated) => {
        toast.success(`${updated.length} harga katalog dihitung ulang`);
        setRecalcOpen(false);
      },
      onError: (err) => {
        toast.error(err instanceof ApiError ? err.message : "Gagal menghitung ulang harga");
      },
    });
  }

  // Produk yang sudah ada di katalog disembunyikan dari pilihan agar tidak
  // menabrak batasan satu produk satu kali per trip.
  const availableProducts = products?.items.filter(
    (product) => editing?.product_id === product.id || !items?.some((item) => item.product_id === product.id),
  );

  const daftarProduk = availableProducts ?? [];
  // Produk yang baru dibuat ikut ditawarkan sampai daftar dari server memuatnya.
  const semuaProduk =
    produkBaru && !daftarProduk.some((p) => p.id === produkBaru.id)
      ? [produkBaru, ...daftarProduk]
      : daftarProduk;

  const productOptions = semuaProduk.map((product) => ({
    value: product.id,
    label: product.name,
    // SKU ikut dicari tapi tidak jadi judul: orang mengingat nama produknya,
    // dan mengetik SKU tetap harus ketemu.
    keywords: product.sku,
    description: product.sku,
  }));
  const fieldError = (name: string) =>
    save.error instanceof ApiError ? save.error.fields?.[name] : undefined;

  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="space-y-1">
          <h2 className="text-lg font-semibold">Katalog trip</h2>
          <p className="text-sm text-muted-foreground">
            Harga jual dikunci saat disimpan, jadi tidak berubah walau kurs trip diedit belakangan.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" onClick={() => setRecalcOpen(true)} disabled={!items?.length}>
            <RefreshCw />
            Hitung Ulang Harga
          </Button>
          <Button onClick={openCreate}>
            <Plus />
            Tambah Produk
          </Button>
        </div>
      </div>

      <ErrorState error={error} />

      <DataTable
        columns={7}
        isLoading={isLoading}
        isEmpty={!isLoading && (items?.length ?? 0) === 0}
        emptyTitle="Katalog masih kosong"
        emptyDescription="Tambahkan produk yang akan ditawarkan pada trip ini beserta markup-nya."
        emptyAction={
          <Button onClick={openCreate}>
            <Plus />
            Tambah Produk
          </Button>
        }
        head={
          <TR>
            <TH>Produk</TH>
            <TH className="hidden text-right sm:table-cell">Modal ({trip.currency})</TH>
            <TH className="hidden text-right lg:table-cell">Modal (Rp)</TH>
            <TH className="hidden text-right lg:table-cell">Markup</TH>
            <TH className="text-right">Harga Jual</TH>
            <TH className="hidden text-right sm:table-cell">Dipesan</TH>
            <TH className="text-right">Aksi</TH>
          </TR>
        }
      >
        {items?.map((item) => (
          <TR key={item.id}>
            <TD>
              <div className="flex items-center gap-2">
                <p className="font-medium">{item.product_name}</p>
                {!item.is_active && <Badge variant="neutral">Nonaktif</Badge>}
              </div>
              <p className="text-xs text-muted-foreground">
                {item.product_sku}
                {item.brand ? ` · ${item.brand}` : ""}
              </p>
            </TD>
            <TD className="tabular hidden text-right sm:table-cell">
              {formatNumber(item.cost_price)}
            </TD>
            <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
              {formatIDR(item.cost_price_idr)}
            </TD>
            <TD className="tabular hidden text-right lg:table-cell">
              {item.markup_type === "percent"
                ? `${formatNumber(item.markup_value)}%`
                : formatIDR(item.markup_value)}
            </TD>
            <TD className="tabular text-right font-semibold">{formatIDR(item.sell_price)}</TD>
            <TD className="tabular hidden text-right sm:table-cell">
              {formatNumber(item.qty_ordered)}
              {item.max_qty ? (
                <span className="text-muted-foreground"> / {formatNumber(item.max_qty)}</span>
              ) : null}
            </TD>
            <TD>
              <div className="flex justify-end gap-1">
                <Button variant="ghost" size="icon-sm" onClick={() => openEdit(item)} tooltip="Ubah">
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
              </div>
            </TD>
          </TR>
        ))}
      </DataTable>

      <FormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        title={editing ? "Ubah Item Katalog" : "Tambah ke Katalog"}
        error={save.error}
        loading={save.isPending}
        onSubmit={handleSubmit}
      >
        <div className="space-y-4">
          <Field label="Produk" htmlFor="product_id" required error={fieldError("product_id")}>
            {/*
              Tombol produk baru duduk persis di sebelah pilihannya. Produk yang
              belum terdaftar biasanya baru ketahuan saat menyusun katalog —
              memaksa admin membatalkan dialog ini, pergi ke menu Produk, lalu
              mengulang dari awal berarti kehilangan harga dan markup yang sudah
              diketik.
            */}
            <div className="flex min-w-0 gap-2">
              {/*
                Combobox, bukan Select. Dua alasan, dan keduanya nyata.

                Katalog produk bisa memanjang, dan aturan proyek ini memang
                menyebut Combobox untuk daftar semacam itu — mengetik sebagian
                nama jauh lebih cepat daripada menggulir ratusan baris.

                Yang kedua baru ketahuan saat menguji tombol Produk baru: Radix
                Select memanggil onValueChange("") kalau nilainya disetel ke opsi
                yang baru muncul pada render yang sama, sehingga produk yang baru
                dibuat langsung terlepas lagi. Combobox membaca labelnya dari
                daftar biasa, jadi tidak punya perilaku itu.
              */}
              <Combobox
                value={form.product_id}
                onChange={selectProduct}
                options={productOptions}
                placeholder="Pilih produk…"
                searchPlaceholder="Cari nama atau SKU…"
                emptyLabel="Produk tidak ditemukan"
                disabled={Boolean(editing)}
                className="min-w-0 flex-1"
              />
              {!editing && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      onClick={() => setQuickAddOpen(true)}
                    >
                      <Plus />
                      <span className="sr-only">Produk baru</span>
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Produk baru</TooltipContent>
                </Tooltip>
              )}
            </div>
          </Field>

          {/* Harga trip sebelumnya ditempel di sini supaya admin tidak perlu
              membuka catatan lama untuk mengingat modal produk ini. */}
          <LastPriceHint
            productId={form.product_id || undefined}
            currency={trip.currency}
            onUse={(costPrice) => setForm((current) => ({ ...current, cost_price: costPrice }))}
          />

          <div className="grid gap-4 sm:grid-cols-3">
            <Field label={`Harga modal (${trip.currency})`} htmlFor="cost_price" required>
              <Input
                id="cost_price"
                type="number"
                min="0"
                step="any"
                value={form.cost_price}
                onChange={(event) => setForm({ ...form, cost_price: event.target.value })}
                required
              />
            </Field>

            <Field label="Jenis markup" htmlFor="markup_type">
              <OptionSelect
                id="markup_type"
                value={form.markup_type}
                onChange={(value) => setForm({ ...form, markup_type: value })}
                options={MARKUP_OPTIONS}
              />
            </Field>

            <Field
              label={form.markup_type === "percent" ? "Markup (%)" : "Markup (Rp)"}
              htmlFor="markup_value"
            >
              <Input
                id="markup_value"
                type="number"
                min="0"
                step="any"
                value={form.markup_value}
                onChange={(event) => setForm({ ...form, markup_value: event.target.value })}
              />
            </Field>
          </div>

          <Card className="bg-muted/50 py-0">
            <CardContent className="flex flex-wrap items-center justify-between gap-4 py-3">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Calculator className="size-4" />
                Perkiraan harga
              </div>
              <div className="flex flex-wrap items-center gap-6 text-sm">
                <div>
                  <p className="text-xs text-muted-foreground">Modal (Rp)</p>
                  <p className="tabular font-medium">{formatIDR(previewCostIDR)}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Harga jual</p>
                  <p className="tabular text-base font-semibold">{formatIDR(previewSell)}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Margin/pcs</p>
                  <p className="tabular font-medium text-emerald-600">
                    {formatIDR(previewSell - previewCostIDR)}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-4 sm:grid-cols-2">
            <Field
              label="Batas kuota"
              htmlFor="max_qty"
              hint="Kosongkan kalau tidak dibatasi"
            >
              <Input
                id="max_qty"
                type="number"
                min="1"
                value={form.max_qty ?? ""}
                onChange={(event) =>
                  setForm({ ...form, max_qty: event.target.value ? Number(event.target.value) : null })
                }
                placeholder="mis. 10"
              />
            </Field>

            <Field label="Catatan" htmlFor="item_notes">
              <Textarea
                id="item_notes"
                rows={2}
                value={form.notes ?? ""}
                onChange={(event) => setForm({ ...form, notes: event.target.value })}
                placeholder="Varian, batasan toko, dan lain-lain"
              />
            </Field>
          </div>

          <CheckboxField
            id="catalog_is_active"
            checked={form.is_active ?? true}
            onCheckedChange={(checked) => setForm({ ...form, is_active: checked })}
          >
            Aktif dan bisa dipesan customer
          </CheckboxField>
        </div>
      </FormDialog>

      {/*
        Dirender sebagai saudara dialog katalog, bukan di dalamnya. Radix
        memindahkan isi tiap dialog ke ujung body, jadi menumpuknya di dalam
        hanya membuat pohon komponennya berlapis tanpa manfaat — sementara
        FormDialog sudah menghentikan perambatan submit, sehingga menyimpan
        produk di sini tidak ikut mengirim form katalog di belakangnya.
      */}
      <QuickAddProductDialog
        open={quickAddOpen}
        onOpenChange={setQuickAddOpen}
        onCreated={pilihProdukBaru}
      />

      <ConfirmDialog
        open={recalcOpen}
        onOpenChange={setRecalcOpen}
        title="Hitung ulang semua harga jual?"
        description={`Seluruh harga katalog akan dihitung ulang memakai kurs trip saat ini (1 ${trip.currency} = Rp${formatNumber(trip.exchange_rate)}). Harga yang sudah terlanjur diumumkan ke customer bisa berubah.`}
        confirmLabel="Hitung ulang"
        destructive={false}
        loading={recalculate.isPending}
        error={recalculate.error}
        onConfirm={handleRecalculate}
      />

      <ConfirmDialog
        open={Boolean(deleting)}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Keluarkan dari katalog?"
        description={`${deleting?.product_name ?? ""} tidak akan bisa dipesan lagi pada trip ini. Produk yang sudah terlanjur dipesan tidak bisa dihapus — nonaktifkan saja.`}
        confirmLabel="Keluarkan"
        loading={remove.isPending}
        error={remove.error}
        onConfirm={handleDelete}
      />
    </>
  );
}
