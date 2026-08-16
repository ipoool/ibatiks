"use client";

import { useState } from "react";
import { toast } from "sonner";

import { OptionSelect } from "@/components/filter-select";
import { Combobox } from "@/components/ui/combobox";
import { Field } from "@/components/ui/field";
import { FormDialog } from "@/components/ui/form-dialog";
import { Input } from "@/components/ui/input";
import { useProducts } from "@/hooks/use-master";
import { useSaveTripItem, useTripItems } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatIDR, toNumber } from "@/lib/utils";
import type { MarkupType, Trip } from "@/types/api";

const MARKUP_OPTIONS: ReadonlyArray<{ value: MarkupType; label: string }> = [
  { value: "percent", label: "Persen (%)" },
  { value: "nominal", label: "Nominal (Rp)" },
];

const EMPTY = { product_id: "", cost_price: "", markup_type: "percent" as MarkupType, markup_value: "" };

/**
 * Dialog ringkas untuk memasukkan produk ke katalog trip tanpa meninggalkan
 * form order.
 *
 * Yang ditawarkan hanya produk dari master data yang belum ada di katalog trip
 * ini — memasukkan produk yang sama dua kali ditolak backend, dan menampilkannya
 * di daftar pilihan hanya mengundang kesalahan.
 */
export function QuickAddCatalogItemDialog({
  trip,
  open,
  onOpenChange,
  onCreated,
}: {
  trip: Trip;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Dipanggil dengan product_id supaya pemanggil bisa langsung memakainya. */
  onCreated: (productId: string) => void;
}) {
  const [form, setForm] = useState(EMPTY);

  const { data: products, isLoading: loadingProducts } = useProducts({ per_page: 200 });
  const { data: catalog } = useTripItems(open ? trip.id : undefined);
  const save = useSaveTripItem(trip.id);

  const alreadyInCatalog = new Set(catalog?.map((item) => item.product_id));
  const productOptions =
    products?.items
      .filter((product) => product.is_active && !alreadyInCatalog.has(product.id))
      .map((product) => ({
        value: product.id,
        label: product.name,
        keywords: `${product.sku} ${product.brand ?? ""}`,
        description: `${product.sku} · modal ${product.base_currency} ${product.base_price}`,
      })) ?? [];

  const selectedProduct = products?.items.find((product) => product.id === form.product_id);

  // Perkiraan harga jual ditampilkan sebelum menyimpan supaya admin tahu angka
  // yang akan dilihat customer, bukan baru sadar setelah masuk katalog.
  const costIDR = Math.round(toNumber(form.cost_price) * toNumber(trip.exchange_rate));
  const sellPrice =
    form.markup_type === "percent"
      ? Math.round((costIDR * (100 + toNumber(form.markup_value))) / 100 / 100) * 100
      : Math.round((costIDR + toNumber(form.markup_value)) / 100) * 100;

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    save.mutate(
      {
        product_id: form.product_id,
        cost_price: form.cost_price || "0",
        markup_type: form.markup_type,
        markup_value: form.markup_value || "0",
      },
      {
        onSuccess: () => {
          toast.success(`${selectedProduct?.name ?? "Produk"} masuk katalog trip`);
          onCreated(form.product_id);
          setForm(EMPTY);
          onOpenChange(false);
        },
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal menambah ke katalog");
        },
      },
    );
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={(next) => {
        if (!next) setForm(EMPTY);
        save.reset();
        onOpenChange(next);
      }}
      title="Tambah Produk ke Katalog Trip"
      description={`Harga modal diisi dalam ${trip.currency}; konversinya memakai kurs trip 1 ${trip.currency} = Rp${trip.exchange_rate}.`}
      error={save.error}
      loading={save.isPending}
      submitLabel="Tambah ke katalog"
      onSubmit={handleSubmit}
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Produk" htmlFor="qa_product" required className="sm:col-span-2">
          <Combobox
            id="qa_product"
            value={form.product_id}
            onChange={(value) => setForm({ ...form, product_id: value })}
            options={productOptions}
            isLoading={loadingProducts}
            placeholder="Pilih produk dari master data…"
            searchPlaceholder="Cari nama, SKU, atau brand…"
            emptyLabel="Semua produk sudah ada di katalog trip ini"
          />
        </Field>

        <Field label={`Harga modal (${trip.currency})`} htmlFor="qa_cost" required>
          <Input
            id="qa_cost"
            type="number"
            min="0"
            step="any"
            value={form.cost_price}
            onChange={(event) => setForm({ ...form, cost_price: event.target.value })}
            required
          />
        </Field>

        <Field label="Jenis markup" htmlFor="qa_markup_type">
          <OptionSelect
            id="qa_markup_type"
            value={form.markup_type}
            onChange={(value) => setForm({ ...form, markup_type: value })}
            options={MARKUP_OPTIONS}
          />
        </Field>

        <Field
          label={form.markup_type === "percent" ? "Markup (%)" : "Markup (Rp)"}
          htmlFor="qa_markup_value"
          className="sm:col-span-2"
        >
          <Input
            id="qa_markup_value"
            type="number"
            min="0"
            step="any"
            value={form.markup_value}
            onChange={(event) => setForm({ ...form, markup_value: event.target.value })}
          />
        </Field>

        {costIDR > 0 && (
          <div className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-border bg-muted/40 px-3 py-2 text-sm sm:col-span-2">
            <span className="text-muted-foreground">
              Modal Rp{costIDR.toLocaleString("id-ID")}
            </span>
            <span className="font-medium">Harga jual {formatIDR(sellPrice)}</span>
          </div>
        )}
      </div>
    </FormDialog>
  );
}
