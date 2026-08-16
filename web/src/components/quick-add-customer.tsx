"use client";

import { useState } from "react";
import { toast } from "sonner";

import { FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useSaveCustomer } from "@/hooks/use-master";
import { ApiError } from "@/lib/api";

const EMPTY = { name: "", phone_wa: "", address: "", city: "" };

/**
 * Dialog ringkas untuk membuat customer di tengah pencatatan order.
 *
 * Isinya sengaja hanya empat kolom, bukan salinan form customer lengkap: yang
 * dibutuhkan saat mencatat order hanyalah identitas dan alamat kirim, dan
 * memindahkan seluruh form ke sini akan membuat admin mengisi hal-hal yang
 * belum tentu diketahuinya saat itu juga. Sisanya bisa dilengkapi belakangan
 * dari menu Customer.
 */
export function QuickAddCustomerDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Dipanggil dengan id customer baru supaya form pemanggil bisa langsung memilihnya. */
  onCreated: (customerId: string) => void;
}) {
  const [form, setForm] = useState(EMPTY);
  const save = useSaveCustomer();

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    save.mutate(
      {
        name: form.name.trim(),
        phone_wa: form.phone_wa.trim(),
        address: form.address.trim() || null,
        city: form.city.trim() || null,
      },
      {
        onSuccess: (customer) => {
          toast.success(`Customer ${customer.name} ditambahkan`);
          onCreated(customer.id);
          setForm(EMPTY);
          onOpenChange(false);
        },
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal menyimpan customer");
        },
      },
    );
  }

  const fieldError = (name: string) =>
    save.error instanceof ApiError ? save.error.fields?.[name] : undefined;

  return (
    <FormDialog
      open={open}
      onOpenChange={(next) => {
        if (!next) setForm(EMPTY);
        save.reset();
        onOpenChange(next);
      }}
      title="Tambah Customer"
      description="Cukup isi yang perlu sekarang; sisanya bisa dilengkapi dari menu Customer."
      error={save.error}
      loading={save.isPending}
      onSubmit={handleSubmit}
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Nama" htmlFor="qa_name" required error={fieldError("name")}>
          <Input
            id="qa_name"
            value={form.name}
            onChange={(event) => setForm({ ...form, name: event.target.value })}
            placeholder="Nama customer"
            required
            autoFocus
          />
        </Field>

        <Field
          label="Nomor WhatsApp"
          htmlFor="qa_phone"
          required
          error={fieldError("phone_wa")}
          hint="Boleh 0812… atau +62812…"
        >
          <Input
            id="qa_phone"
            value={form.phone_wa}
            onChange={(event) => setForm({ ...form, phone_wa: event.target.value })}
            placeholder="081234567890"
            required
          />
        </Field>

        <Field label="Alamat" htmlFor="qa_address" className="sm:col-span-2">
          <Input
            id="qa_address"
            value={form.address}
            onChange={(event) => setForm({ ...form, address: event.target.value })}
            placeholder="Jalan, nomor rumah, RT/RW"
          />
        </Field>

        <Field label="Kota" htmlFor="qa_city" className="sm:col-span-2">
          <Input
            id="qa_city"
            value={form.city}
            onChange={(event) => setForm({ ...form, city: event.target.value })}
            placeholder="Jakarta Selatan"
          />
        </Field>
      </div>
    </FormDialog>
  );
}
