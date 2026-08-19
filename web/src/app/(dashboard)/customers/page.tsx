"use client";

import { History, MessageCircle, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog, FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState, PageHeader, SearchInput } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

import { DialogRiwayatOrder } from "./dialog-riwayat-order";
import { useDebounced } from "@/hooks/use-debounced";
import {
  useCustomers,
  useDeleteCustomer,
  useSaveCustomer,
  type CustomerPayload,
} from "@/hooks/use-master";
import { usePencarianURL } from "@/hooks/use-pencarian-url";
import { ApiError } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import { SocialFields, labelPlatform } from "@/components/social-fields";
import type { Customer } from "@/types/api";

const EMPTY_FORM: CustomerPayload = {
  name: "",
  phone_wa: "",
  email: "",
  socials: [],
  address: "",
  city: "",
  district: "",
  subdistrict: "",
  province: "",
  postal_code: "",
  notes: "",
};

export default function CustomersPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = usePencarianURL();
  const debouncedSearch = useDebounced(search);

  const [editing, setEditing] = useState<Customer | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<CustomerPayload>(EMPTY_FORM);
  const [deleting, setDeleting] = useState<Customer | null>(null);
  const [riwayat, setRiwayat] = useState<Customer | null>(null);

  const { data, isLoading, error } = useCustomers({ page, q: debouncedSearch });
  const save = useSaveCustomer(editing?.id);
  const remove = useDeleteCustomer();

  function openCreate() {
    setEditing(null);
    setForm(EMPTY_FORM);
    save.reset();
    setFormOpen(true);
  }

  function openEdit(customer: Customer) {
    setEditing(customer);
    setForm({
      name: customer.name,
      phone_wa: customer.phone_wa,
      email: customer.email ?? "",
      socials: customer.socials ?? [],
      address: customer.address ?? "",
      city: customer.city ?? "",
      district: customer.district ?? "",
      subdistrict: customer.subdistrict ?? "",
      province: customer.province ?? "",
      postal_code: customer.postal_code ?? "",
      notes: customer.notes ?? "",
    });
    save.reset();
    setFormOpen(true);
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    // Field opsional yang dikosongkan dikirim sebagai null supaya kolomnya
    // benar-benar kosong di database, bukan berisi string kosong.
    const payload: CustomerPayload = {
      ...form,
      email: form.email || null,
      address: form.address || null,
      city: form.city || null,
      district: form.district || null,
      subdistrict: form.subdistrict || null,
      province: form.province || null,
      postal_code: form.postal_code || null,
      notes: form.notes || null,
    };

    save.mutate(payload, {
      onSuccess: () => {
        toast.success(editing ? "Customer diperbarui" : "Customer ditambahkan");
        setFormOpen(false);
      },
    });
  }

  function handleDelete() {
    if (!deleting) return;
    remove.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Customer dihapus");
        setDeleting(null);
      },
      onError: (err) => {
        toast.error(err instanceof ApiError ? err.message : "Gagal menghapus customer");
      },
    });
  }

  const fieldError = (name: string) =>
    save.error instanceof ApiError ? save.error.fields?.[name] : undefined;

  return (
    <>
      <PageHeader
        title="Customer"
        description="Daftar pelanggan jastip beserta alamat pengirimannya"
        actions={
          <Button onClick={openCreate}>
            <Plus />
            Tambah Customer
          </Button>
        }
      />

      <ErrorState error={error} />

      <SearchInput
        value={search}
        onChange={(value) => {
          setSearch(value);
          setPage(1);
        }}
        placeholder="Cari nama, nomor WA, atau kode customer…"
        className="max-w-md"
      />

      <div>
        <DataTable
          columns={5}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle={search ? "Customer tidak ditemukan" : "Belum ada customer"}
          emptyDescription={
            search
              ? "Coba kata kunci lain, atau tambahkan customer baru."
              : "Tambahkan customer pertama untuk mulai mencatat order."
          }
          emptyAction={
            !search && (
              <Button onClick={openCreate}>
                <Plus />
                Tambah Customer
              </Button>
            )
          }
          head={
            <TR>
              <TH>Customer</TH>
              <TH className="hidden sm:table-cell">Kontak</TH>
              <TH className="hidden lg:table-cell">Alamat</TH>
              <TH className="hidden lg:table-cell">Terdaftar</TH>
              <TH className="text-right">Aksi</TH>
            </TR>
          }
        >
          {data?.items.map((customer) => (
            <TR key={customer.id}>
              <TD>
                <p className="font-medium">{customer.name}</p>
                <p className="text-xs text-muted-foreground">{customer.code}</p>
                {/* Nomor WA menyusul nama saat kolom kontak disembunyikan —
                    itu satu-satunya cara menghubungi customer dari daftar. */}
                <p className="text-xs text-muted-foreground sm:hidden">
                  {customer.phone_wa}
                  {customer.city ? ` · ${customer.city}` : ""}
                </p>
              </TD>
              <TD className="hidden sm:table-cell">
                <a
                  href={`https://wa.me/${customer.phone_wa}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 text-sm hover:underline"
                >
                  <MessageCircle className="size-3.5 text-emerald-600" />
                  {customer.phone_wa}
                </a>
                {customer.email && (
                  <p className="text-xs text-muted-foreground">{customer.email}</p>
                )}
              
                {/* Akun sosialnya ikut ditampilkan: dari daftar inilah admin
                    memutuskan lewat mana customer ini dihubungi, dan membuka
                    detailnya dulu untuk itu adalah satu klik yang tidak perlu. */}
                {customer.socials.length > 0 && (
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {customer.socials
                      .map((akun) => `${labelPlatform(akun.platform)} ${akun.handle}`)
                      .join(" · ")}
                  </p>
                )}
              </TD>
              <TD className="hidden max-w-xs lg:table-cell">
                <p className="truncate text-sm">{customer.address || "—"}</p>
                <p className="text-xs text-muted-foreground">
                  {[customer.city, customer.province].filter(Boolean).join(", ") || "—"}
                </p>
              </TD>
              <TD className="hidden text-sm text-muted-foreground lg:table-cell">
                {formatDate(customer.created_at)}
              </TD>
              <TD>
                <div className="flex justify-end gap-1">
                  {/* Riwayat dibuka dari daftar, bukan dari halaman tersendiri:
                      pertanyaannya muncul justru saat sedang menelusuri nama. */}
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setRiwayat(customer)}
                      >
                        <History />
                        <span className="sr-only">Riwayat order</span>
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Riwayat order</TooltipContent>
                  </Tooltip>
                  <Button variant="ghost" size="icon-sm" onClick={() => openEdit(customer)}>
                    <Pencil />
                    <span className="sr-only">Ubah</span>
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    tooltip="Hapus"
                    className="text-destructive hover:text-destructive"
                    onClick={() => setDeleting(customer)}
                  >
                    <Trash2 />
                    <span className="sr-only">Hapus</span>
                  </Button>
                </div>
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />

        {riwayat && (
          <DialogRiwayatOrder customer={riwayat} onClose={() => setRiwayat(null)} />
        )}
      </div>

      <FormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        title={editing ? "Ubah Customer" : "Tambah Customer"}
        description="Alamat di sini dipakai sebagai alamat kirim default saat membuat order."
        error={save.error}
        loading={save.isPending}
        onSubmit={handleSubmit}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Nama" htmlFor="name" required error={fieldError("name")}>
            <Input
              id="name"
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
              placeholder="Nama lengkap customer"
              required
            />
          </Field>

          <Field
            label="Nomor WhatsApp"
            htmlFor="phone_wa"
            required
            error={fieldError("phone_wa")}
            hint="Boleh ditulis 0812… atau +62812…, sistem yang merapikan"
          >
            <Input
              id="phone_wa"
              value={form.phone_wa}
              onChange={(event) => setForm({ ...form, phone_wa: event.target.value })}
              placeholder="081234567890"
              required
            />
          </Field>

          <Field label="Email" htmlFor="email" error={fieldError("email")}>
            <Input
              id="email"
              type="email"
              value={form.email ?? ""}
              onChange={(event) => setForm({ ...form, email: event.target.value })}
              placeholder="opsional"
            />
          </Field>

          <Field label="Media sosial" className="sm:col-span-2">
            <SocialFields
              value={form.socials ?? []}
              onChange={(socials) => setForm({ ...form, socials })}
            />
          </Field>

          <Field label="Alamat" htmlFor="address" className="sm:col-span-2">
            <Textarea
              id="address"
              value={form.address ?? ""}
              onChange={(event) => setForm({ ...form, address: event.target.value })}
              placeholder="Jalan, nomor rumah, RT/RW"
            />
          </Field>

          {/* Kelurahan dan kecamatan dipisah dari alamat jalan: kurir memakai
              keduanya untuk menentukan gudang sortir, dan alamat yang menumpuk
              semuanya dalam satu baris sering terpotong di label paket. */}
          {/* Tiap isian alamat berdiri langsung di grid formnya, bukan dibungkus
              grid dua kolom lagi. Grid dua kolom di dalam satu kolom grid dua kolom
              menyisakan seperempat lebar untuk tiap isian — dan justru nama kelurahan
              dan kecamatan yang paling panjang. */}
          <Field label="Kelurahan" htmlFor="subdistrict">
            <Input
              id="subdistrict"
              value={form.subdistrict ?? ""}
              onChange={(event) => setForm({ ...form, subdistrict: event.target.value })}
              placeholder="Cilandak Barat"
            />
          </Field>

          <Field label="Kecamatan" htmlFor="district">
            <Input
              id="district"
              value={form.district ?? ""}
              onChange={(event) => setForm({ ...form, district: event.target.value })}
              placeholder="Cilandak"
            />
          </Field>

          <Field label="Kota/Kabupaten" htmlFor="city">
            <Input
              id="city"
              value={form.city ?? ""}
              onChange={(event) => setForm({ ...form, city: event.target.value })}
              placeholder="Jakarta Selatan"
            />
          </Field>

          <Field label="Provinsi" htmlFor="province">
            <Input
              id="province"
              value={form.province ?? ""}
              onChange={(event) => setForm({ ...form, province: event.target.value })}
              placeholder="DKI Jakarta"
            />
          </Field>

          <Field label="Kode Pos" htmlFor="postal_code">
            <Input
              id="postal_code"
              value={form.postal_code ?? ""}
              onChange={(event) => setForm({ ...form, postal_code: event.target.value })}
              placeholder="12140"
            />
          </Field>

          <Field label="Catatan" htmlFor="notes" className="sm:col-span-2">
            <Textarea
              id="notes"
              value={form.notes ?? ""}
              onChange={(event) => setForm({ ...form, notes: event.target.value })}
              placeholder="Preferensi packing, patokan alamat, dan lain-lain"
              rows={2}
            />
          </Field>
        </div>
      </FormDialog>

      <ConfirmDialog
        open={Boolean(deleting)}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Hapus customer?"
        description={`${deleting?.name ?? ""} akan disembunyikan dari daftar. Riwayat order lamanya tetap tersimpan.`}
        confirmLabel="Hapus"
        loading={remove.isPending}
        error={remove.error}
        onConfirm={handleDelete}
      />
    </>
  );
}
