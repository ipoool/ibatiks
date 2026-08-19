"use client";

import { useState } from "react";
import { toast } from "sonner";

import { FilterSelect } from "@/components/filter-select";
import { Field } from "@/components/ui/field";
import { FormDialog } from "@/components/ui/form-dialog";
import { Input } from "@/components/ui/input";
import { NumberInput } from "@/components/ui/number-input";
import { useCategories, useSaveProduct } from "@/hooks/use-master";
import { ApiError } from "@/lib/api";
import type { Product } from "@/types/api";

const EMPTY = {
  name: "",
  category_id: "",
  brand: "",
  store_name: "",
  weight_gram: 0,
};

/**
 * Dialog ringkas untuk mendaftarkan produk baru di tengah menyusun katalog trip.
 *
 * Isinya sengaja bukan salinan penuh form Produk. Harga modal, mata uang, dan
 * markup tidak diminta di sini karena dialog katalog yang memanggilnya sudah
 * menanyakan ketiganya untuk trip ini — dua kolom "harga modal" pada dua dialog
 * yang bertumpuk hanya membuat orang menebak mana yang dipakai.
 *
 * Berat ikut diminta karena ia yang dipakai memperkirakan ongkir nanti, dan
 * produk tanpa berat membuat perkiraannya terlalu rendah tanpa gejala apa pun.
 * Sisanya — SKU yang dibuat otomatis, gambar, catatan — bisa dilengkapi
 * belakangan dari menu Produk.
 */
export function QuickAddProductDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /**
   * Dipanggil dengan produk yang baru dibuat, bukan sekadar id-nya.
   *
   * Pemanggil butuh seluruh datanya: daftar produk di form pemanggil baru
   * menyusul beberapa saat kemudian, dan Radix Select yang menerima nilai tanpa
   * opsi yang cocok akan menampilkan placeholder lalu tidak memperbaruinya lagi
   * walau opsinya belakangan datang.
   */
  onCreated: (product: Product) => void;
}) {
  const [form, setForm] = useState(EMPTY);
  const { data: categories } = useCategories();
  const save = useSaveProduct();

  const categoryOptions =
    categories?.map((category) => ({ value: category.id, label: category.name })) ?? [];

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    save.mutate(
      {
        name: form.name.trim(),
        category_id: form.category_id || undefined,
        brand: form.brand.trim() || undefined,
        store_name: form.store_name.trim() || undefined,
        weight_gram: form.weight_gram,
        // Harga dan markup dibiarkan kosong: yang berlaku untuk trip ini diisi
        // di dialog katalog yang memanggil dialog ini.
        base_currency: "IDR",
        markup_type: "percent",
        is_active: true,
      },
      {
        onSuccess: (product) => {
          toast.success(`Produk ${product.name} dibuat`);
          onCreated(product);
          setForm(EMPTY);
          onOpenChange(false);
        },
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal menyimpan produk");
        },
      },
    );
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Produk Baru"
      description="Cukup keterangan dasarnya. Harga modal dan markup untuk trip ini diisi di dialog katalog setelah produk tersimpan."
      error={save.error}
      loading={save.isPending}
      onSubmit={handleSubmit}
      submitLabel="Simpan Produk"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Nama produk" htmlFor="qp_name" required className="sm:col-span-2">
          <Input
            id="qp_name"
            value={form.name}
            onChange={(event) => setForm({ ...form, name: event.target.value })}
            placeholder="Tokyo Banana 8pcs"
            required
          />
        </Field>

        <Field label="Kategori" htmlFor="qp_category">
          <FilterSelect
            value={form.category_id}
            onChange={(category_id) => setForm({ ...form, category_id })}
            allLabel="Tanpa kategori"
            options={categoryOptions}
          />
        </Field>

        <Field label="Brand" htmlFor="qp_brand">
          <Input
            id="qp_brand"
            value={form.brand}
            onChange={(event) => setForm({ ...form, brand: event.target.value })}
            placeholder="Tokyo Banana"
          />
        </Field>

        <Field
          label="Toko langganan"
          htmlFor="qp_store"
          hint="Membantu tripper mencari di lapangan"
        >
          <Input
            id="qp_store"
            value={form.store_name}
            onChange={(event) => setForm({ ...form, store_name: event.target.value })}
            placeholder="Don Quijote Shibuya"
          />
        </Field>

        <Field
          label="Berat (gram)"
          htmlFor="qp_weight"
          hint="Dipakai memperkirakan ongkir. Produk tanpa berat membuat perkiraannya terlalu rendah."
        >
          <NumberInput
            id="qp_weight"
            value={form.weight_gram}
            onValueChange={(weight_gram) => setForm({ ...form, weight_gram })}
            min={0}
            step="any"
            blankWhenZero
          />
        </Field>
      </div>
    </FormDialog>
  );
}
