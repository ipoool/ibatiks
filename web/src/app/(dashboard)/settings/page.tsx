"use client";

import { KeyRound, Plus, Save, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { ShippingProviderCard, ShippingTestPanel } from "./shipping-provider";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { FilterSelect, OptionSelect, toOptions } from "@/components/filter-select";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ConfirmDialog, FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState, PageHeader } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
import { formatDateTime } from "@/lib/utils";
import type { Permission, User, UserRole } from "@/types/api";

/*
 * Kunci pengaturan yang punya form khusus, dikelompokkan agar mudah dibaca.
 *
 * Tiap isian membawa aturan validasinya sendiri. Nilai-nilai ini tercetak di
 * invoice dan label paket yang dipegang customer, jadi salah ketik di sini tidak
 * ketahuan sampai dokumennya sudah terkirim. Pesan galatnya berbahasa Indonesia
 * lewat mekanisme yang sama seperti form lain (src/lib/validasi-bawaan.ts);
 * `title` dipakai untuk menjelaskan pola yang diminta.
 */
const STORE_FIELDS = [
  {
    key: "store_name",
    label: "Nama toko",
    hint: "Tampil di header invoice",
    required: true,
    minLength: 2,
    maxLength: 120,
  },
  {
    key: "store_phone",
    label: "Nomor WA toko",
    hint: "Format internasional, contoh 6281234567890",
    inputMode: "tel" as const,
    // Angka saja, boleh diawali tanda plus. Sengaja tidak memaksa awalan 62:
    // nomornya cuma dicetak di invoice dan label, bukan dijadikan tautan wa.me,
    // dan menolak nomor yang sah karena bentuknya berbeda lebih merugikan.
    pattern: "\\+?[0-9]{8,20}",
    title: "Nomor WA hanya berisi angka, boleh diawali tanda +. Contoh: 6281234567890",
  },
  {
    key: "store_email",
    label: "Email toko",
    type: "email" as const,
    placeholder: "halo@tokokamu.id",
  },
  { key: "store_address", label: "Alamat toko", maxLength: 300 },
  {
    key: "bank_account",
    label: "Rekening pembayaran",
    hint: "Tampil di invoice dan pesan penagihan",
    maxLength: 200,
  },
  { key: "invoice_footer", label: "Catatan penutup invoice", maxLength: 300 },
  {
    key: "invoice_due_days",
    label: "Jatuh tempo invoice (hari)",
    hint: "Dihitung sejak invoice terbit. Kosong atau tidak wajar dianggap 3 hari.",
    type: "number" as const,
    min: 0,
    max: 365,
    step: 1,
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


export default function SettingsPage() {
  return (
    <>
      <PageHeader
        title="Pengaturan"
        description="Identitas toko, template pesan WhatsApp, pengiriman, dan pengguna back office"
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

        <TabsContent value="ongkir" className="space-y-6">
          <ShippingProviderCard />
          <ShippingTestPanel />
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
  /** Atribut validasi bawaan peramban; pesannya diterjemahkan oleh Input. */
  type?: "email" | "number" | "text";
  inputMode?: "tel" | "numeric";
  pattern?: string;
  title?: string;
  placeholder?: string;
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  min?: number;
  max?: number;
  step?: number;
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
                  type={field.type}
                  inputMode={field.inputMode}
                  pattern={field.pattern}
                  title={field.title}
                  placeholder={field.placeholder}
                  required={field.required}
                  minLength={field.minLength}
                  maxLength={field.maxLength}
                  min={field.min}
                  max={field.max}
                  step={field.step}
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

const PERMISSION_LABEL: Record<Permission, string> = {
  trips: "Trip",
  shopping_list: "Daftar Belanja",
  purchases: "Pembelian",
  orders: "Order",
  invoices: "Invoice",
  shipments: "Pengiriman",
  customers: "Customer",
  products: "Produk",
  stock: "Stok",
  reports: "Laporan",
  settings: "Pengaturan",
  users: "Manajemen Pengguna",
};

/**
 * Menu bawaan tiap role — salinan dari `domain.DefaultPermissions` di backend.
 *
 * Dipakai hanya untuk menentukan centang apa yang ditawarkan; keputusan akhir
 * tetap di backend, yang menyaring ulang permintaan supaya centang tidak bisa
 * melebarkan batas role.
 */
const DEFAULT_PERMISSIONS: Record<UserRole, Permission[]> = {
  owner: [
    "trips",
    "shopping_list",
    "purchases",
    "orders",
    "invoices",
    "shipments",
    "customers",
    "products",
    "stock",
    "reports",
    "settings",
    "users",
  ],
  admin: [
    "trips",
    "shopping_list",
    "purchases",
    "orders",
    "invoices",
    "shipments",
    "customers",
    "products",
    "stock",
    "reports",
  ],
  tripper: ["trips", "shopping_list", "purchases", "products"],
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
    permissions: [] as Permission[],
  });

  const { data, isLoading, error } = useUsers({ per_page: 100 });
  const save = useSaveUser(editing?.id);
  const remove = useDeleteUser();
  const resetPassword = useResetUserPassword();

  function openCreate() {
    setEditing(null);
    setForm({
      name: "",
      email: "",
      password: "",
      role: "admin",
      phone: "",
      is_active: true,
      // Dicentang penuh sesuai bawaan role: pengguna baru memang memulai
      // dengan seluruh menu rolenya, dan kotak yang semuanya kosong akan
      // terbaca seolah-olah ia tidak diberi akses apa pun.
      permissions: DEFAULT_PERMISSIONS.admin,
    });
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
      // Yang ditampilkan adalah hak efektifnya, jadi centang di dialog persis
      // menggambarkan menu yang sekarang bisa dibuka pengguna itu.
      permissions: user.effective_permissions ?? [],
    });
    save.reset();
    setFormOpen(true);
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    // Saat mengubah, email dan password tidak ikut dikirim: keduanya punya
    // alur tersendiri agar tidak berubah tanpa disengaja.
    // Mencentang semua yang boleh dibuka role ini sama saja dengan tidak
    // menyetel apa-apa, jadi dikirim kosong supaya pengguna itu ikut berubah
    // otomatis kalau bawaan role diperluas nanti.
    const roleDefaults = DEFAULT_PERMISSIONS[form.role];
    const permissions =
      form.permissions.length === roleDefaults.length ? [] : form.permissions;

    const payload = editing
      ? {
          name: form.name,
          role: form.role,
          phone: form.phone || null,
          is_active: form.is_active,
          permissions,
        }
      : {
          name: form.name,
          email: form.email,
          password: form.password,
          role: form.role,
          phone: form.phone || null,
          permissions,
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
            <TH className="hidden sm:table-cell">Email</TH>
            <TH>Role</TH>
            <TH className="hidden lg:table-cell">Login terakhir</TH>
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
              {/* Email menyusul nama saat kolomnya disembunyikan: itu yang
                  dipakai login, jadi tidak boleh hilang dari daftar. */}
              <p className="text-xs text-muted-foreground sm:hidden">{user.email}</p>
            </TD>
            <TD className="hidden text-sm sm:table-cell">{user.email}</TD>
            <TD>
              <Badge variant={user.role === "owner" ? "info" : "neutral"}>
                {ROLE_LABEL[user.role]}
              </Badge>
            </TD>
            <TD className="hidden text-sm text-muted-foreground lg:table-cell">
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
              // Ganti role berarti daftar menunya berbeda, jadi centangnya
              // dikembalikan ke bawaan role yang baru alih-alih menyisakan
              // pilihan lama yang belum tentu masih berlaku.
              onChange={(value) =>
                setForm({ ...form, role: value, permissions: DEFAULT_PERMISSIONS[value] })
              }
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

          {/* Hak akses per menu: batas kasarnya tetap role, centang di sini
              hanya bisa mempersempit. Backend menerapkan aturan yang sama, jadi
              menu yang dimatikan benar-benar tertutup, bukan sekadar hilang
              dari daftar. */}
          <div className="space-y-2 sm:col-span-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm font-medium">Menu yang boleh dibuka</p>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setForm({ ...form, permissions: DEFAULT_PERMISSIONS[form.role] })}
              >
                Ikuti bawaan role
              </Button>
            </div>

            <div className="grid gap-x-4 gap-y-2 rounded-lg border border-border p-3 sm:grid-cols-2">
              {DEFAULT_PERMISSIONS[form.role].map((permission) => (
                <CheckboxField
                  key={permission}
                  id={`perm_${permission}`}
                  checked={form.permissions.includes(permission)}
                  onCheckedChange={(checked) =>
                    setForm({
                      ...form,
                      permissions: checked
                        ? [...form.permissions, permission]
                        : form.permissions.filter((item) => item !== permission),
                    })
                  }
                >
                  {PERMISSION_LABEL[permission]}
                </CheckboxField>
              ))}
            </div>

            {editing && (
              <p className="text-xs text-muted-foreground">
                Mengubah hak akses mengeluarkan pengguna ini dari seluruh perangkatnya, supaya
                pembatasannya berlaku saat itu juga.
              </p>
            )}
          </div>

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
              <TH className="hidden sm:table-cell">Pengguna</TH>
              <TH>Aksi</TH>
              <TH className="hidden lg:table-cell">Detail</TH>
            </TR>
          }
        >
          {data?.items.map((log) => (
            <TR key={log.id}>
              <TD className="text-sm">
                <span className="whitespace-nowrap">{formatDateTime(log.created_at)}</span>
                <p className="text-xs text-muted-foreground sm:hidden">
                  {log.user_name ?? "sistem"}
                </p>
              </TD>
              <TD className="hidden text-sm sm:table-cell">{log.user_name ?? "sistem"}</TD>
              <TD>
                <Badge variant="neutral">{ACTION_LABEL[log.action] ?? log.action}</Badge>
                <p className="mt-1 text-xs text-muted-foreground">{log.entity}</p>
              </TD>
              <TD className="hidden max-w-md lg:table-cell">
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
