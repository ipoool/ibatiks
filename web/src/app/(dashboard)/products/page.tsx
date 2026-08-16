"use client";

import { History, Pencil, Plus, Tags, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FilterSelect, OptionSelect } from "@/components/filter-select";
import { PriceHistoryDialog } from "@/components/price-history";
import { ConfirmDialog, FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useDebounced } from "@/hooks/use-debounced";
import {
  useCategories,
  useDeleteProduct,
  useProducts,
  useSaveProduct,
  type ProductPayload,
} from "@/hooks/use-master";
import { ApiError } from "@/lib/api";
import { formatIDR, formatNumber } from "@/lib/utils";
import type { MarkupType, Product } from "@/types/api";

import { CategoryManager } from "./category-manager";

const EMPTY_FORM: ProductPayload = {
  sku: "",
  name: "",
  category_id: "",
  brand: "",
  store_name: "",
  base_currency: "IDR",
  base_price: "0",
  markup_type: "percent",
  markup_value: "0",
  weight_gram: 0,
  notes: "",
  is_active: true,
};

const MARKUP_OPTIONS: ReadonlyArray<{ value: MarkupType; label: string }> = [
  { value: "percent", label: "Persen (%)" },
  { value: "nominal", label: "Nominal (Rp)" },
];

export default function ProductsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const debouncedSearch = useDebounced(search);

  const [editing, setEditing] = useState<Product | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<ProductPayload>(EMPTY_FORM);
  const [deleting, setDeleting] = useState<Product | null>(null);
  const [historyOf, setHistoryOf] = useState<Product | null>(null);
  const [categoryOpen, setCategoryOpen] = useState(false);

  const { data, isLoading, error } = useProducts({
    page,
    q: debouncedSearch,
    category_id: categoryFilter || undefined,
  });
  const { data: categories } = useCategories();
  const categoryOptions =
    categories?.map((category) => ({ value: category.id, label: category.name })) ?? [];
  const save = useSaveProduct(editing?.id);
  const remove = useDeleteProduct();

  function openCreate() {
    setEditing(null);
    setForm(EMPTY_FORM);
    save.reset();
    setFormOpen(true);
  }

  function openEdit(product: Product) {
    setEditing(product);
    setForm({
      sku: product.sku,
      name: product.name,
      category_id: product.category_id ?? "",
      brand: product.brand ?? "",
      store_name: product.store_name ?? "",
      base_currency: product.base_currency,
      base_price: product.base_price,
      markup_type: product.markup_type,
      markup_value: product.markup_value,
      weight_gram: product.weight_gram,
      image_url: product.image_url ?? "",
      notes: product.notes ?? "",
      is_active: product.is_active,
    });
    save.reset();
    setFormOpen(true);
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    save.mutate(
      {
        ...form,
        sku: form.sku || undefined,
        category_id: form.category_id || null,
        brand: form.brand || null,
        store_name: form.store_name || null,
        image_url: form.image_url || null,
        notes: form.notes || null,
      },
      {
        onSuccess: () => {
          toast.success(editing ? "Produk diperbarui" : "Produk ditambahkan");
          setFormOpen(false);
        },
      },
    );
  }

  function handleDelete() {
    if (!deleting) return;
    remove.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Produk dinonaktifkan");
        setDeleting(null);
      },
    });
  }

  const fieldError = (name: string) =>
    save.error instanceof ApiError ? save.error.fields?.[name] : undefined;

  return (
    <>
      <PageHeader
        title="Produk"
        description="Katalog master. Harga jual sesungguhnya ditentukan per trip mengikuti kurs saat itu."
        actions={
          <>
            <Button variant="outline" onClick={() => setCategoryOpen(true)}>
              <Tags />
              Kategori
            </Button>
            <Button onClick={openCreate}>
              <Plus />
              Tambah Produk
            </Button>
          </>
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
          placeholder="Cari nama, SKU, atau brand…"
          className="min-w-64 flex-1 sm:max-w-md"
        />
        <FilterSelect
          value={categoryFilter}
          onChange={(value) => {
            setCategoryFilter(value);
            setPage(1);
          }}
          allLabel="Semua kategori"
          options={categoryOptions}
          className="sm:w-56"
        />
      </div>

      <div>
        <DataTable
          columns={6}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle={search ? "Produk tidak ditemukan" : "Belum ada produk"}
          emptyDescription={
            search
              ? "Coba kata kunci lain."
              : "Tambahkan produk untuk bisa memasukkannya ke katalog trip."
          }
          emptyAction={
            !search && (
              <Button onClick={openCreate}>
                <Plus />
                Tambah Produk
              </Button>
            )
          }
          head={
            <TR>
              <TH>Produk</TH>
              <TH className="hidden xl:table-cell">Kategori</TH>
              <TH className="text-right">Harga Modal</TH>
              <TH className="hidden text-right sm:table-cell">Markup</TH>
              <TH className="hidden text-right xl:table-cell">Berat</TH>
              <TH className="text-right">Aksi</TH>
            </TR>
          }
        >
          {data?.items.map((product) => (
            <TR key={product.id}>
              {/* Nama produk panjang: dibiarkan turun baris supaya tidak
                  mendorong kolom harga dan tombol keluar dari kartunya. */}
              <TD className="whitespace-normal">
                <div className="flex items-center gap-2">
                  <p className="font-medium">{product.name}</p>
                  {!product.is_active && <Badge variant="neutral">Nonaktif</Badge>}
                </div>
                <p className="text-xs text-muted-foreground">
                  {product.sku}
                  {product.brand ? ` · ${product.brand}` : ""}
                </p>
              </TD>
              <TD className="hidden text-sm text-muted-foreground xl:table-cell">
                {product.category_name ?? "—"}
              </TD>
              <TD className="tabular text-right">
                {product.base_currency === "IDR"
                  ? formatIDR(product.base_price)
                  : `${product.base_currency} ${formatNumber(product.base_price)}`}
              </TD>
              <TD className="tabular hidden text-right sm:table-cell">
                {product.markup_type === "percent"
                  ? `${formatNumber(product.markup_value)}%`
                  : formatIDR(product.markup_value)}
              </TD>
              <TD className="tabular hidden text-right text-muted-foreground xl:table-cell">
                {formatNumber(product.weight_gram)} g
              </TD>
              <TD>
                <div className="flex justify-end gap-1">
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setHistoryOf(product)}
                      >
                        <History />
                        <span className="sr-only">Riwayat harga</span>
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Riwayat harga antar trip</TooltipContent>
                  </Tooltip>
                  <Button variant="ghost" size="icon-sm" onClick={() => openEdit(product)}>
                    <Pencil />
                    <span className="sr-only">Ubah</span>
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    className="text-destructive hover:text-destructive"
                    onClick={() => setDeleting(product)}
                  >
                    <Trash2 />
                    <span className="sr-only">Nonaktifkan</span>
                  </Button>
                </div>
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>

      <FormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        title={editing ? "Ubah Produk" : "Tambah Produk"}
        description="Harga modal diisi dalam mata uang negara asal; konversinya mengikuti kurs trip."
        error={save.error}
        loading={save.isPending}
        onSubmit={handleSubmit}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Nama produk" htmlFor="name" required error={fieldError("name")} className="sm:col-span-2">
            <Input
              id="name"
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
              placeholder="Hada Labo Gokujyun Lotion 170ml"
              required
            />
          </Field>

          <Field label="SKU" htmlFor="sku" hint="Kosongkan untuk dibuatkan otomatis">
            <Input
              id="sku"
              value={form.sku ?? ""}
              onChange={(event) => setForm({ ...form, sku: event.target.value })}
              placeholder="SKN-0001"
            />
          </Field>

          <Field label="Kategori" htmlFor="category_id">
            <FilterSelect
              value={form.category_id ?? ""}
              onChange={(value) => setForm({ ...form, category_id: value })}
              allLabel="Tanpa kategori"
              options={categoryOptions}
            />
          </Field>

          <Field label="Brand" htmlFor="brand">
            <Input
              id="brand"
              value={form.brand ?? ""}
              onChange={(event) => setForm({ ...form, brand: event.target.value })}
              placeholder="Hada Labo"
            />
          </Field>

          <Field label="Toko langganan" htmlFor="store_name" hint="Membantu tripper mencari di lapangan">
            <Input
              id="store_name"
              value={form.store_name ?? ""}
              onChange={(event) => setForm({ ...form, store_name: event.target.value })}
              placeholder="Don Quijote"
            />
          </Field>

          <Field label="Mata uang" htmlFor="base_currency" error={fieldError("base_currency")}>
            <Input
              id="base_currency"
              value={form.base_currency ?? "IDR"}
              onChange={(event) =>
                setForm({ ...form, base_currency: event.target.value.toUpperCase() })
              }
              placeholder="JPY"
              maxLength={3}
            />
          </Field>

          <Field label="Harga modal" htmlFor="base_price" hint="Dalam mata uang di atas">
            <Input
              id="base_price"
              type="number"
              min="0"
              step="any"
              value={form.base_price ?? "0"}
              onChange={(event) => setForm({ ...form, base_price: event.target.value })}
            />
          </Field>

          <Field label="Jenis markup" htmlFor="markup_type" required>
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
              step={form.markup_type === "percent" ? "0.5" : "1000"}
              value={form.markup_value ?? "0"}
              onChange={(event) => setForm({ ...form, markup_value: event.target.value })}
            />
          </Field>

          <Field label="Berat (gram)" htmlFor="weight_gram" hint="Dipakai memperkirakan ongkir">
            <Input
              id="weight_gram"
              type="number"
              min="0"
              value={form.weight_gram ?? 0}
              onChange={(event) => setForm({ ...form, weight_gram: Number(event.target.value) })}
            />
          </Field>

          <Field label="URL gambar" htmlFor="image_url" error={fieldError("image_url")}>
            <Input
              id="image_url"
              type="url"
              value={form.image_url ?? ""}
              onChange={(event) => setForm({ ...form, image_url: event.target.value })}
              placeholder="https://…"
            />
          </Field>

          <Field label="Catatan" htmlFor="notes" className="sm:col-span-2">
            <Textarea
              id="notes"
              rows={2}
              value={form.notes ?? ""}
              onChange={(event) => setForm({ ...form, notes: event.target.value })}
              placeholder="Varian, ukuran, atau catatan khusus"
            />
          </Field>

          <CheckboxField
            id="is_active"
            className="sm:col-span-2"
            checked={form.is_active ?? true}
            onCheckedChange={(checked) => setForm({ ...form, is_active: checked })}
          >
            Produk aktif dan bisa dimasukkan ke katalog trip
          </CheckboxField>
        </div>
      </FormDialog>

      <CategoryManager open={categoryOpen} onOpenChange={setCategoryOpen} />

      <ConfirmDialog
        open={Boolean(deleting)}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Nonaktifkan produk?"
        description={`${deleting?.name ?? ""} tidak akan muncul lagi saat menyusun katalog trip. Riwayat order lama tetap utuh.`}
        confirmLabel="Nonaktifkan"
        loading={remove.isPending}
        error={remove.error}
        onConfirm={handleDelete}
      />

      <PriceHistoryDialog
        productId={historyOf?.id}
        productName={historyOf?.name ?? ""}
        open={Boolean(historyOf)}
        onOpenChange={(open) => !open && setHistoryOf(null)}
      />
    </>
  );
}
