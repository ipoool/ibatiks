"use client";

import { KeyRound, Plus, Save, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { FilterSelect, OptionSelect, toOptions } from "@/components/filter-select";
import { useHasRole } from "@/components/layout/user-context";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ConfirmDialog, FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState, PageHeader } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useDebounced } from "@/hooks/use-debounced";
import {
  useDeleteShippingRate,
  useSaveShippingRate,
  useShippingRates,
} from "@/hooks/use-operations";
import {
  useAuditLogs,
  useDeleteUser,
  useResetUserPassword,
  useSaveUser,
  useSettings,
  useUpdateSettings,
  useUsers,
} from "@/hooks/use-reports";
import { ApiError } from "@/lib/api";
import { formatDateTime, formatIDR } from "@/lib/utils";
import type { ShippingRate, User, UserRole } from "@/types/api";

/** Kunci pengaturan yang punya form khusus, dikelompokkan agar mudah dibaca. */
const STORE_FIELDS = [
  { key: "store_name", label: "Nama toko", hint: "Tampil di header invoice" },
  { key: "store_phone", label: "Nomor WA toko", hint: "Format internasional, contoh 6281234567890" },
  { key: "store_email", label: "Email toko" },
  { key: "store_address", label: "Alamat toko" },
  { key: "bank_account", label: "Rekening pembayaran", hint: "Tampil di invoice dan pesan penagihan" },
  { key: "invoice_footer", label: "Catatan penutup invoice" },
  { key: "invoice_due_days", label: "Jatuh tempo invoice (hari)" },
] as const;

/*
 * Parameter perhitungan ongkir. Pembagi volume mengikuti kebiasaan ekspedisi
 * dalam negeri (6000 untuk JNE); tarif cadangan dipakai saat kota tujuan belum
 * terdaftar di tabel tarif.
 */
const SHIPPING_FIELDS = [
  {
    key: "shipping_volumetric_divisor",
    label: "Pembagi berat volume",
    hint: "JNE memakai 6000: (P × L × T dalam cm) ÷ 6000 = kg volume",
  },
  {
    key: "shipping_default_price_per_kg",
    label: "Tarif cadangan per kg (Rp)",
    hint: "Dipakai kalau kota tujuan belum ada di tabel tarif",
  },
] as const;

const TEMPLATE_FIELDS = [
  {
    key: "wa_template_dp",
    label: "Pesan permintaan DP",
    placeholders: "{{customer_name}} {{trip_title}} {{total}} {{dp_amount}} {{bank_account}}",
  },
  {
    key: "wa_template_invoice",
    label: "Pesan penagihan pelunasan",
    placeholders:
      "{{customer_name}} {{invoice_number}} {{total}} {{amount_paid}} {{amount_due}} {{bank_account}} {{due_date}}",
  },
  {
    key: "wa_template_shipped",
    label: "Pesan informasi pengiriman",
    placeholders: "{{customer_name}} {{order_number}} {{courier}} {{service}} {{tracking_number}}",
  },
] as const;

const AUDIT_ENTITY_OPTIONS = [
  { value: "order", label: "Order" },
  { value: "trip", label: "Trip" },
  { value: "purchase", label: "Pembelian" },
  { value: "invoice", label: "Invoice" },
  { value: "shipment", label: "Pengiriman" },
  { value: "settings", label: "Pengaturan" },
] as const;

const COURIER_SERVICE_OPTIONS = ["REG", "YES", "OKE", "JTR"].map((service) => ({
  value: service,
  label: service,
}));

export default function SettingsPage() {
  return (
    <>
      <PageHeader
        title="Pengaturan"
        description="Identitas toko, template pesan WhatsApp, tarif ongkir, dan pengguna back office"
      />

      <Tabs defaultValue="toko">
        <TabsList>
          <TabsTrigger value="toko">Toko</TabsTrigger>
          <TabsTrigger value="template">Template Pesan</TabsTrigger>
          <TabsTrigger value="ongkir">Ongkir</TabsTrigger>
          <TabsTrigger value="pengguna">Pengguna</TabsTrigger>
          <TabsTrigger value="audit">Jejak Perubahan</TabsTrigger>
        </TabsList>

        <TabsContent value="toko">
          <SettingsForm
            title="Identitas toko"
            description="Data ini muncul pada invoice PDF dan pesan yang dikirim ke customer."
            fields={STORE_FIELDS.map((field) => ({ ...field, multiline: false }))}
          />
        </TabsContent>

        <TabsContent value="template">
          <SettingsForm
            title="Template pesan WhatsApp"
            description="Teks di bawah dipakai saat menyiapkan pesan. Placeholder diganti otomatis dengan data order."
            fields={TEMPLATE_FIELDS.map((field) => ({
              key: field.key,
              label: field.label,
              hint: `Placeholder: ${field.placeholders}`,
              multiline: true,
            }))}
          />
        </TabsContent>

        <TabsContent value="ongkir">
          <SettingsForm
            title="Perhitungan ongkir"
            description="Dipakai saat menghitung perkiraan ongkir pada dialog pengemasan."
            fields={SHIPPING_FIELDS.map((field) => ({ ...field, multiline: false }))}
          />
          <ShippingRates />
        </TabsContent>

        <TabsContent value="pengguna">
          <UserManagement />
        </TabsContent>

        <TabsContent value="audit">
          <AuditTrail />
        </TabsContent>
      </Tabs>
    </>
  );
}

interface SettingField {
  key: string;
  label: string;
  hint?: string;
  multiline: boolean;
}

function SettingsForm({
  title,
  description,
  fields,
}: {
  title: string;
  description: string;
  fields: SettingField[];
}) {
  const { data: settings, isLoading, error } = useSettings();
  const update = useUpdateSettings();

  // Yang disimpan di state hanya isian yang benar-benar diubah pengguna; sisanya
  // dibaca langsung dari data server. Dengan begitu ketikan yang belum disimpan
  // tidak pernah tertimpa saat cache di-refresh di latar belakang.
  const [edits, setEdits] = useState<Record<string, string>>({});

  const serverValues = Object.fromEntries(
    (settings ?? []).map((setting) => [setting.key, setting.value]),
  );
  const valueOf = (key: string) => edits[key] ?? serverValues[key] ?? "";

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    const payload = Object.fromEntries(fields.map((field) => [field.key, valueOf(field.key)]));
    update.mutate(payload, {
      onSuccess: () => {
        toast.success("Pengaturan disimpan");
        // Setelah tersimpan, nilai server menjadi sumber kebenaran lagi.
        setEdits({});
      },
      onError: (err) => {
        toast.error(err instanceof ApiError ? err.message : "Gagal menyimpan pengaturan");
      },
    });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <ErrorState error={error ?? update.error} className="mb-4" />

        <form onSubmit={handleSubmit} className="space-y-4">
          {fields.map((field) => (
            <Field key={field.key} label={field.label} htmlFor={field.key} hint={field.hint}>
              {field.multiline ? (
                <Textarea
                  id={field.key}
                  rows={6}
                  value={valueOf(field.key)}
                  onChange={(event) => setEdits({ ...edits, [field.key]: event.target.value })}
                  disabled={isLoading}
                />
              ) : (
                <Input
                  id={field.key}
                  value={valueOf(field.key)}
                  onChange={(event) => setEdits({ ...edits, [field.key]: event.target.value })}
                  disabled={isLoading}
                />
              )}
            </Field>
          ))}

          <Button type="submit" loading={update.isPending}>
            <Save />
            Simpan
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

const ROLE_LABEL: Record<UserRole, string> = {
  owner: "Owner",
  admin: "Admin",
  tripper: "Tripper",
};

const ROLE_HINT: Record<UserRole, string> = {
  owner: "Akses penuh termasuk laporan laba dan manajemen pengguna",
  admin: "Seluruh operasional harian: trip, order, invoice, kirim, stok",
  tripper: "Hanya daftar belanja dan input pembelian di lapangan",
};

function UserManagement() {
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [deleting, setDeleting] = useState<User | null>(null);
  const [resetting, setResetting] = useState<User | null>(null);
  const [newPassword, setNewPassword] = useState("");
  const [form, setForm] = useState({
    name: "",
    email: "",
    password: "",
    role: "admin" as UserRole,
    phone: "",
    is_active: true,
  });

  const { data, isLoading, error } = useUsers({ per_page: 100 });
  const save = useSaveUser(editing?.id);
  const remove = useDeleteUser();
  const resetPassword = useResetUserPassword();

  function openCreate() {
    setEditing(null);
    setForm({ name: "", email: "", password: "", role: "admin", phone: "", is_active: true });
    save.reset();
    setFormOpen(true);
  }

  function openEdit(user: User) {
    setEditing(user);
    setForm({
      name: user.name,
      email: user.email,
      password: "",
      role: user.role,
      phone: user.phone ?? "",
      is_active: user.is_active,
    });
    save.reset();
    setFormOpen(true);
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    // Saat mengubah, email dan password tidak ikut dikirim: keduanya punya
    // alur tersendiri agar tidak berubah tanpa disengaja.
    const payload = editing
      ? { name: form.name, role: form.role, phone: form.phone || null, is_active: form.is_active }
      : {
          name: form.name,
          email: form.email,
          password: form.password,
          role: form.role,
          phone: form.phone || null,
        };

    save.mutate(payload, {
      onSuccess: () => {
        toast.success(editing ? "Pengguna diperbarui" : "Pengguna ditambahkan");
        setFormOpen(false);
      },
    });
  }

  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Beri tripper akun sendiri agar bisa mencatat belanja langsung dari lapangan.
        </p>
        <Button onClick={openCreate}>
          <Plus />
          Tambah Pengguna
        </Button>
      </div>

      <ErrorState error={error} />

      <DataTable
        columns={5}
        isLoading={isLoading}
        isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
        emptyTitle="Belum ada pengguna lain"
        head={
          <TR>
            <TH>Nama</TH>
            <TH>Email</TH>
            <TH>Role</TH>
            <TH>Login terakhir</TH>
            <TH className="text-right">Aksi</TH>
          </TR>
        }
      >
        {data?.items.map((user) => (
          <TR key={user.id}>
            <TD>
              <div className="flex items-center gap-2">
                <p className="font-medium">{user.name}</p>
                {!user.is_active && <Badge variant="neutral">Nonaktif</Badge>}
              </div>
              {user.phone && <p className="text-xs text-muted-foreground">{user.phone}</p>}
            </TD>
            <TD className="text-sm">{user.email}</TD>
            <TD>
              <Badge variant={user.role === "owner" ? "info" : "neutral"}>
                {ROLE_LABEL[user.role]}
              </Badge>
            </TD>
            <TD className="text-sm text-muted-foreground">
              {user.last_login_at ? formatDateTime(user.last_login_at) : "belum pernah"}
            </TD>
            <TD>
              <div className="flex justify-end gap-1">
                <Button variant="ghost" size="sm" onClick={() => openEdit(user)}>
                  Ubah
                </Button>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => {
                        setResetting(user);
                        setNewPassword("");
                      }}
                    >
                      <KeyRound />
                      <span className="sr-only">Reset password</span>
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Reset password</TooltipContent>
                </Tooltip>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  className="text-destructive hover:text-destructive"
                  onClick={() => setDeleting(user)}
                >
                  <Trash2 />
                </Button>
              </div>
            </TD>
          </TR>
        ))}
      </DataTable>

      <FormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        title={editing ? "Ubah Pengguna" : "Tambah Pengguna"}
        error={save.error}
        loading={save.isPending}
        onSubmit={handleSubmit}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Nama" htmlFor="user_name" required>
            <Input
              id="user_name"
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
              required
            />
          </Field>

          <Field label="Email" htmlFor="user_email" required>
            <Input
              id="user_email"
              type="email"
              value={form.email}
              onChange={(event) => setForm({ ...form, email: event.target.value })}
              disabled={Boolean(editing)}
              required
            />
          </Field>

          {!editing && (
            <Field label="Password" htmlFor="user_password" required hint="Minimal 8 karakter">
              <Input
                id="user_password"
                type="password"
                value={form.password}
                onChange={(event) => setForm({ ...form, password: event.target.value })}
                minLength={8}
                required
              />
            </Field>
          )}

          <Field label="Role" htmlFor="user_role" required hint={ROLE_HINT[form.role]}>
            <OptionSelect
              id="user_role"
              value={form.role}
              onChange={(value) => setForm({ ...form, role: value })}
              options={toOptions(ROLE_LABEL)}
            />
          </Field>

          <Field label="Nomor HP" htmlFor="user_phone">
            <Input
              id="user_phone"
              value={form.phone}
              onChange={(event) => setForm({ ...form, phone: event.target.value })}
            />
          </Field>

          {editing && (
            <CheckboxField
              id="user_is_active"
              className="sm:col-span-2"
              checked={form.is_active}
              onCheckedChange={(checked) => setForm({ ...form, is_active: checked })}
            >
              Akun aktif dan bisa login
            </CheckboxField>
          )}
        </div>
      </FormDialog>

      <FormDialog
        open={Boolean(resetting)}
        onOpenChange={(open) => !open && setResetting(null)}
        title="Reset Password"
        description={`Password baru untuk ${resetting?.name ?? ""}. Seluruh sesi aktifnya akan dikeluarkan.`}
        error={resetPassword.error}
        loading={resetPassword.isPending}
        submitLabel="Reset Password"
        onSubmit={(event) => {
          event.preventDefault();
          if (!resetting) return;
          resetPassword.mutate(
            { id: resetting.id, password: newPassword },
            {
              onSuccess: () => {
                toast.success("Password direset");
                setResetting(null);
              },
            },
          );
        }}
      >
        <Field label="Password baru" htmlFor="reset_password" required hint="Minimal 8 karakter">
          <Input
            id="reset_password"
            type="password"
            value={newPassword}
            onChange={(event) => setNewPassword(event.target.value)}
            minLength={8}
            required
            autoFocus
          />
        </Field>
      </FormDialog>

      <ConfirmDialog
        open={Boolean(deleting)}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Hapus pengguna?"
        description={`${deleting?.name ?? ""} tidak akan bisa login lagi. Jejak perubahan yang pernah dibuatnya tetap tersimpan.`}
        confirmLabel="Hapus"
        loading={remove.isPending}
        error={remove.error}
        onConfirm={() => {
          if (!deleting) return;
          remove.mutate(deleting.id, {
            onSuccess: () => {
              toast.success("Pengguna dihapus");
              setDeleting(null);
            },
          });
        }}
      />
    </>
  );
}

const ACTION_LABEL: Record<string, string> = {
  create: "Dibuat",
  update: "Diubah",
  delete: "Dihapus",
  status_change: "Ubah status",
  payment_record: "Catat pembayaran",
  item_change: "Ubah item",
  receive: "Terima barang",
  pack: "Kemas",
  ship: "Kirim",
  sent: "Kirim invoice",
  void: "Batalkan invoice",
  delivered: "Diterima customer",
};

function AuditTrail() {
  const [page, setPage] = useState(1);
  const [entity, setEntity] = useState("");
  const { data, isLoading, error } = useAuditLogs({ page, entity: entity || undefined });

  return (
    <>
      <ErrorState error={error} />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Catatan siapa mengubah apa — berguna saat menelusuri perubahan qty atau nominal order.
        </p>
        <FilterSelect
          value={entity}
          onChange={(value) => {
            setEntity(value);
            setPage(1);
          }}
          allLabel="Semua entitas"
          options={AUDIT_ENTITY_OPTIONS}
          className="sm:w-48"
        />
      </div>

      <div>
        <DataTable
          columns={4}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada jejak perubahan"
          head={
            <TR>
              <TH>Waktu</TH>
              <TH>Pengguna</TH>
              <TH>Aksi</TH>
              <TH>Detail</TH>
            </TR>
          }
        >
          {data?.items.map((log) => (
            <TR key={log.id}>
              <TD className="whitespace-nowrap text-sm">{formatDateTime(log.created_at)}</TD>
              <TD className="text-sm">{log.user_name ?? "sistem"}</TD>
              <TD>
                <Badge variant="neutral">{ACTION_LABEL[log.action] ?? log.action}</Badge>
                <p className="mt-1 text-xs text-muted-foreground">{log.entity}</p>
              </TD>
              <TD className="max-w-md">
                {log.changes ? (
                  <code className="block truncate text-xs text-muted-foreground">
                    {JSON.stringify(log.changes)}
                  </code>
                ) : (
                  <span className="text-xs text-muted-foreground">—</span>
                )}
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}

function ShippingRates() {
  const canEdit = useHasRole("owner");
  const [search, setSearch] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<ShippingRate | null>(null);
  const debouncedSearch = useDebounced(search);

  const { data, isLoading, error } = useShippingRates({ q: debouncedSearch || undefined });
  const remove = useDeleteShippingRate();

  return (
    <Card className="mt-6">
      <CardHeader>
        <CardTitle>Tarif per kota tujuan</CardTitle>
        <CardDescription>
          Tarif dipakai berurutan: kota tujuan order dicocokkan dulu di sini, kalau tidak ketemu
          barulah tarif cadangan di atas yang dipakai.
        </CardDescription>
        {canEdit && (
          <CardAction>
            <Button size="sm" onClick={() => setFormOpen(true)}>
              <Plus />
              Tambah Tarif
            </Button>
          </CardAction>
        )}
      </CardHeader>

      <CardContent className="space-y-4">
        <ErrorState error={error} />

        <Input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Cari kota…"
          className="sm:max-w-xs"
        />

        <DataTable
          columns={canEdit ? 6 : 5}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.length ?? 0) === 0}
          emptyTitle="Belum ada tarif"
          emptyDescription="Tambahkan kota tujuan yang sering dikirimi paket."
          head={
            <TR>
              <TH className="min-w-40">Kota</TH>
              <TH className="w-24">Kurir</TH>
              <TH className="w-24">Layanan</TH>
              <TH className="w-32 text-right">Per kg</TH>
              <TH className="w-24">Estimasi</TH>
              {canEdit && <TH className="w-16 text-right">Aksi</TH>}
            </TR>
          }
        >
          {data?.map((rate) => (
            <TR key={rate.id}>
              <TD>
                <p className="font-medium capitalize">{rate.destination_city}</p>
                {rate.province && (
                  <p className="text-xs text-muted-foreground">{rate.province}</p>
                )}
              </TD>
              <TD className="text-sm">{rate.courier}</TD>
              <TD className="text-sm">{rate.service}</TD>
              <TD className="tabular text-right font-medium">{formatIDR(rate.price_per_kg)}</TD>
              <TD className="text-sm text-muted-foreground">{rate.etd || "—"}</TD>
              {canEdit && (
                <TD className="text-right">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setDeleteTarget(rate)}
                    aria-label={`Hapus tarif ${rate.destination_city}`}
                  >
                    <Trash2 className="text-red-600" />
                  </Button>
                </TD>
              )}
            </TR>
          ))}
        </DataTable>
      </CardContent>

      {formOpen && <ShippingRateDialog onClose={() => setFormOpen(false)} />}

      <ConfirmDialog
        open={Boolean(deleteTarget)}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Hapus tarif ongkir?"
        description={`Tarif untuk ${deleteTarget?.destination_city ?? ""} akan dihapus. Order ke kota ini nanti memakai tarif cadangan.`}
        confirmLabel="Hapus"
        destructive
        error={remove.error}
        loading={remove.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          remove.mutate(deleteTarget.id, {
            onSuccess: () => {
              toast.success("Tarif dihapus");
              setDeleteTarget(null);
            },
            onError: (err) => {
              toast.error(err instanceof ApiError ? err.message : "Gagal menghapus tarif");
            },
          });
        }}
      />
    </Card>
  );
}

function ShippingRateDialog({ onClose }: { onClose: () => void }) {
  const save = useSaveShippingRate();
  const [form, setForm] = useState({
    courier: "JNE",
    service: "REG",
    destination_city: "",
    province: "",
    price_per_kg: "",
    min_weight_gram: 1000,
    etd: "",
  });

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    save.mutate(
      {
        ...form,
        province: form.province || null,
        etd: form.etd || null,
        min_weight_gram: Number(form.min_weight_gram),
      },
      {
        onSuccess: () => {
          toast.success("Tarif tersimpan");
          onClose();
        },
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title="Tambah Tarif Ongkir"
      description="Kota yang sudah ada akan diperbarui tarifnya, bukan digandakan."
      error={save.error}
      loading={save.isPending}
      onSubmit={handleSubmit}
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Kota tujuan" htmlFor="destination_city" required className="sm:col-span-2">
          <Input
            id="destination_city"
            value={form.destination_city}
            onChange={(event) => setForm({ ...form, destination_city: event.target.value })}
            placeholder="Bandung"
            required
            autoFocus
          />
        </Field>

        <Field label="Provinsi" htmlFor="province">
          <Input
            id="province"
            value={form.province}
            onChange={(event) => setForm({ ...form, province: event.target.value })}
            placeholder="Jawa Barat"
          />
        </Field>

        <Field label="Kurir" htmlFor="rate_courier">
          <Input
            id="rate_courier"
            value={form.courier}
            onChange={(event) => setForm({ ...form, courier: event.target.value })}
          />
        </Field>

        <Field label="Layanan" htmlFor="rate_service">
          <OptionSelect
            id="rate_service"
            value={form.service}
            onChange={(value) => setForm({ ...form, service: value })}
            options={COURIER_SERVICE_OPTIONS}
          />
        </Field>

        <Field label="Tarif per kg (Rp)" htmlFor="price_per_kg" required>
          <Input
            id="price_per_kg"
            type="number"
            min="0"
            step="any"
            value={form.price_per_kg}
            onChange={(event) => setForm({ ...form, price_per_kg: event.target.value })}
            required
          />
        </Field>

        <Field
          label="Berat minimum (gram)"
          htmlFor="min_weight_gram"
          hint="Ekspedisi menagih minimal 1 kg"
        >
          <Input
            id="min_weight_gram"
            type="number"
            min="0"
            step="any"
            value={form.min_weight_gram}
            onChange={(event) => setForm({ ...form, min_weight_gram: Number(event.target.value) })}
          />
        </Field>

        <Field label="Estimasi tiba" htmlFor="etd" hint="Tulis lengkap dengan satuannya">
          <Input
            id="etd"
            value={form.etd}
            onChange={(event) => setForm({ ...form, etd: event.target.value })}
            placeholder="2-3 hari"
          />
        </Field>
      </div>
    </FormDialog>
  );
}
