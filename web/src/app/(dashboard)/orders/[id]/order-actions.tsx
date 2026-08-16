"use client";

import { ArrowRight, Ban, PackageSearch, Pencil } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { OptionSelect } from "@/components/filter-select";
import { ORDER_SOURCE_OPTIONS, orderStatusLabel } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { ConfirmButton } from "@/components/ui/confirm-button";
import { FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  useCancelOrder,
  useChangeOrderStatus,
  useReceiveOrder,
  useUpdateOrder,
} from "@/hooks/use-orders";
import { ApiError } from "@/lib/api";
import { formatNumber, toDateInput } from "@/lib/utils";
import type { FulfillmentStatus, OrderDetail, OrderStatus } from "@/types/api";

/**
 * Penjelasan dampak tiap perpindahan status, ditampilkan pada dialog
 * konfirmasi. Yang ditulis di sini adalah akibatnya bagi pekerjaan admin,
 * bukan pengulangan nama statusnya.
 */
const STATUS_EFFECT: Record<OrderStatus, string> = {
  draft:
    "Order dikembalikan ke draft dan tidak lagi masuk antrean penagihan DP. Pakai ini kalau pesanannya ternyata belum pasti.",
  awaiting_dp:
    "Order dianggap sudah pasti dan masuk antrean penagihan DP. Isinya masih bisa diubah.",
  dp_paid:
    "Order ditandai DP-nya sudah masuk dan terkunci untuk dibelikan. Pastikan uangnya benar-benar sudah tercatat di daftar pembayaran.",
  purchasing: "Order ditandai sedang dibelikan tripper di luar negeri.",
  arrived:
    "Barang untuk order ini dianggap sudah tiba dan siap dicocokkan satu per satu dengan pesanannya.",
  packed: "Order ditandai sudah dikemas atas nama customer dan siap ditagih pelunasannya.",
  invoiced: "Order ditandai sudah ditagihkan. Pakai setelah invoice pelunasan benar-benar dikirim.",
  paid: "Order ditandai lunas dan boleh dikirim. Pastikan pelunasannya sudah tercatat di daftar pembayaran.",
  shipped:
    "Order ditandai sudah diserahkan ke kurir. Setelah ini isi order tidak bisa diubah lagi, karena dokumen pengiriman dan invoice harus tetap cocok dengan barang yang sudah jalan.",
  completed: "Order ditutup sebagai selesai. Tidak ada perpindahan status lagi setelah ini.",
  cancelled: "Order dibatalkan.",
};

export function OrderActions({ order }: { order: OrderDetail }) {
  const [cancelOpen, setCancelOpen] = useState(false);
  const [cancelReason, setCancelReason] = useState("");
  const [editOpen, setEditOpen] = useState(false);
  const [receiveOpen, setReceiveOpen] = useState(false);

  const changeStatus = useChangeOrderStatus(order.id);
  const cancel = useCancelOrder(order.id);

  // Pembatalan punya dialognya sendiri karena butuh alasan, jadi dikeluarkan
  // dari daftar tombol status biasa.
  const nextStatuses = order.next_statuses.filter((status) => status !== "cancelled");
  const canCancel = order.next_statuses.includes("cancelled");
  const canReceive = ["dp_paid", "purchasing", "arrived"].includes(order.status);

  async function handleChangeStatus(status: OrderStatus) {
    // mutateAsync dipakai supaya dialog konfirmasi bisa menunggu hasilnya:
    // ditutup kalau berhasil, tetap terbuka beserta pesan galat kalau gagal.
    await changeStatus.mutateAsync(status, {
      onSuccess: () => toast.success(`Order berpindah ke "${orderStatusLabel(status)}"`),
      onError: (error) => {
        toast.error(error instanceof ApiError ? error.message : "Gagal mengubah status");
      },
    });
  }

  return (
    <>
      <div className="flex flex-wrap items-center gap-2">
        {order.editable && (
          <Button variant="outline" onClick={() => setEditOpen(true)}>
            <Pencil />
            Ubah Order
          </Button>
        )}

        {canReceive && (
          <Button variant="outline" onClick={() => setReceiveOpen(true)}>
            <PackageSearch />
            Cocokkan Barang
          </Button>
        )}

        {nextStatuses.map((status, index) => (
          <ConfirmButton
            key={status}
            variant={index === 0 ? "default" : "outline"}
            title={`Pindahkan order ke "${orderStatusLabel(status)}"?`}
            description={STATUS_EFFECT[status]}
            confirmLabel={`Ya, ${orderStatusLabel(status).toLowerCase()}`}
            // Perpindahan ke Dikirim membekukan isi order, jadi ditandai merah
            // agar tidak ditekan sambil lalu.
            destructive={status === "shipped"}
            error={changeStatus.error}
            onConfirm={() => handleChangeStatus(status)}
          >
            <ArrowRight />
            {orderStatusLabel(status)}
          </ConfirmButton>
        ))}

        {canCancel && (
          <Button
            variant="ghost"
            className="text-destructive hover:text-destructive"
            onClick={() => setCancelOpen(true)}
          >
            <Ban />
            Batalkan
          </Button>
        )}
      </div>

      <FormDialog
        open={cancelOpen}
        onOpenChange={setCancelOpen}
        title="Batalkan order ini?"
        description="Barang yang sudah terlanjur dibeli untuk order ini akan dipindahkan ke stok. Uang yang sudah diterima perlu dikembalikan lewat pencatatan refund."
        submitLabel="Batalkan order"
        cancelLabel="Jangan batalkan"
        loading={cancel.isPending}
        error={cancel.error}
        onSubmit={(event) => {
          event.preventDefault();
          cancel.mutate(cancelReason || undefined, {
            onSuccess: () => {
              toast.success("Order dibatalkan");
              setCancelOpen(false);
            },
          });
        }}
      >
        <Field
          label="Alasan pembatalan"
          htmlFor="cancel_reason"
          hint="Opsional, tapi berguna saat menelusuri riwayat order"
        >
          <Input
            id="cancel_reason"
            value={cancelReason}
            onChange={(event) => setCancelReason(event.target.value)}
            placeholder="Customer membatalkan, barang habis, dan lain-lain"
          />
        </Field>
      </FormDialog>

      {editOpen && <EditOrderDialog order={order} onClose={() => setEditOpen(false)} />}
      {receiveOpen && <ReceiveDialog order={order} onClose={() => setReceiveOpen(false)} />}
    </>
  );
}

function EditOrderDialog({ order, onClose }: { order: OrderDetail; onClose: () => void }) {
  const update = useUpdateOrder(order.id);
  const [form, setForm] = useState({
    order_date: toDateInput(order.order_date),
    order_source: order.order_source,
    discount: order.discount,
    shipping_fee: order.shipping_fee,
    dp_required: order.dp_required,
    recipient_name: order.recipient_name,
    recipient_phone: order.recipient_phone,
    shipping_address: order.shipping_address,
    shipping_city: order.shipping_city,
    shipping_province: order.shipping_province ?? "",
    shipping_postal_code: order.shipping_postal_code ?? "",
    notes: order.notes ?? "",
  });

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    update.mutate(
      {
        ...form,
        shipping_province: form.shipping_province || null,
        shipping_postal_code: form.shipping_postal_code || null,
        notes: form.notes || null,
      },
      {
        onSuccess: () => {
          toast.success("Order diperbarui");
          onClose();
        },
      },
    );
  }

  const fieldError = (name: string) =>
    update.error instanceof ApiError ? update.error.fields?.[name] : undefined;

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title="Ubah Order"
      description="Untuk mengubah jumlah produk, edit langsung pada tabel item pesanan."
      error={update.error}
      loading={update.isPending}
      onSubmit={handleSubmit}
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Tanggal order" htmlFor="order_date">
          <Input
            id="order_date"
            type="date"
            value={form.order_date}
            onChange={(event) => setForm({ ...form, order_date: event.target.value })}
          />
        </Field>

        <Field label="Asal order" htmlFor="order_source">
          <OptionSelect
            id="order_source"
            value={form.order_source}
            onChange={(value) => setForm({ ...form, order_source: value })}
            options={ORDER_SOURCE_OPTIONS}
          />
        </Field>

        <Field label="DP diminta (Rp)" htmlFor="dp_required">
          <Input
            id="dp_required"
            type="number"
            min="0"
            step="any"
            value={form.dp_required}
            onChange={(event) => setForm({ ...form, dp_required: event.target.value })}
          />
        </Field>

        <Field label="Diskon (Rp)" htmlFor="edit_discount" error={fieldError("discount")}>
          <Input
            id="edit_discount"
            type="number"
            min="0"
            step="any"
            value={form.discount}
            onChange={(event) => setForm({ ...form, discount: event.target.value })}
          />
        </Field>

        <Field label="Ongkir ditagihkan (Rp)" htmlFor="edit_shipping_fee">
          <Input
            id="edit_shipping_fee"
            type="number"
            min="0"
            step="any"
            value={form.shipping_fee}
            onChange={(event) => setForm({ ...form, shipping_fee: event.target.value })}
          />
        </Field>

        <Field
          label="Nama penerima"
          htmlFor="edit_recipient_name"
          required
          error={fieldError("recipient_name")}
        >
          <Input
            id="edit_recipient_name"
            value={form.recipient_name}
            onChange={(event) => setForm({ ...form, recipient_name: event.target.value })}
            required
          />
        </Field>

        <Field
          label="HP penerima"
          htmlFor="edit_recipient_phone"
          required
          error={fieldError("recipient_phone")}
        >
          <Input
            id="edit_recipient_phone"
            value={form.recipient_phone}
            onChange={(event) => setForm({ ...form, recipient_phone: event.target.value })}
            required
          />
        </Field>

        <Field
          label="Alamat"
          htmlFor="edit_shipping_address"
          required
          className="sm:col-span-2"
          error={fieldError("shipping_address")}
        >
          <Textarea
            id="edit_shipping_address"
            value={form.shipping_address}
            onChange={(event) => setForm({ ...form, shipping_address: event.target.value })}
            required
          />
        </Field>

        <Field label="Kota" htmlFor="edit_shipping_city" required error={fieldError("shipping_city")}>
          <Input
            id="edit_shipping_city"
            value={form.shipping_city}
            onChange={(event) => setForm({ ...form, shipping_city: event.target.value })}
            required
          />
        </Field>

        <div className="grid grid-cols-2 gap-4">
          <Field label="Provinsi" htmlFor="edit_shipping_province">
            <Input
              id="edit_shipping_province"
              value={form.shipping_province}
              onChange={(event) => setForm({ ...form, shipping_province: event.target.value })}
            />
          </Field>
          <Field label="Kode pos" htmlFor="edit_shipping_postal_code">
            <Input
              id="edit_shipping_postal_code"
              value={form.shipping_postal_code}
              onChange={(event) => setForm({ ...form, shipping_postal_code: event.target.value })}
            />
          </Field>
        </div>

        <Field label="Catatan" htmlFor="edit_notes" className="sm:col-span-2">
          <Textarea
            id="edit_notes"
            rows={2}
            value={form.notes}
            onChange={(event) => setForm({ ...form, notes: event.target.value })}
          />
        </Field>
      </div>
    </FormDialog>
  );
}

const FULFILLMENT_OPTIONS: Array<{ value: FulfillmentStatus; label: string }> = [
  { value: "purchased", label: "Lengkap" },
  { value: "partial", label: "Sebagian" },
  { value: "unavailable", label: "Tidak tersedia" },
];

function ReceiveDialog({ order, onClose }: { order: OrderDetail; onClose: () => void }) {
  const receive = useReceiveOrder(order.id);

  // Default-nya diterima penuh: itu yang paling sering terjadi, admin tinggal
  // mengubah baris yang bermasalah saja.
  const [rows, setRows] = useState(
    order.items.map((item) => ({
      item_id: item.id,
      name: item.product_name,
      ordered: item.qty,
      qty_received: item.qty_received > 0 ? item.qty_received : item.qty,
      status: "purchased" as FulfillmentStatus,
    })),
  );

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    receive.mutate(
      {
        items: rows.map((row) => ({
          item_id: row.item_id,
          qty_received: Number(row.qty_received),
          status: row.status,
        })),
      },
      {
        onSuccess: () => {
          toast.success("Barang dicocokkan dengan pesanan");
          onClose();
        },
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal mencatat penerimaan");
        },
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title="Cocokkan Barang yang Datang"
      description="Isi berapa unit yang benar-benar diterima untuk tiap produk."
      error={receive.error}
      loading={receive.isPending}
      onSubmit={handleSubmit}
      submitLabel="Simpan Penerimaan"
      className="max-w-2xl"
    >
      <div className="space-y-3">
        {rows.map((row, index) => (
          <div
            key={row.item_id}
            className="grid items-end gap-3 rounded-lg border border-border p-3 sm:grid-cols-[1fr_7rem_10rem]"
          >
            <div className="min-w-0">
              <p className="truncate font-medium">{row.name}</p>
              <p className="text-xs text-muted-foreground">
                Dipesan {formatNumber(row.ordered)} pcs
              </p>
            </div>

            <Field label="Diterima" htmlFor={`qty_${row.item_id}`}>
              <Input
                id={`qty_${row.item_id}`}
                type="number"
                min="0"
                max={row.ordered}
                value={row.qty_received}
                onChange={(event) => {
                  const qty = Number(event.target.value);
                  setRows((current) =>
                    current.map((item, i) =>
                      i === index
                        ? {
                            ...item,
                            qty_received: qty,
                            // Status ikut menyesuaikan supaya admin tidak perlu
                            // mengubah dua kolom untuk satu kenyataan.
                            status:
                              qty === 0
                                ? "unavailable"
                                : qty >= item.ordered
                                  ? "purchased"
                                  : "partial",
                          }
                        : item,
                    ),
                  );
                }}
              />
            </Field>

            <Field label="Status" htmlFor={`status_${row.item_id}`}>
              <OptionSelect
                id={`status_${row.item_id}`}
                value={row.status}
                onChange={(value) =>
                  setRows((current) =>
                    current.map((item, i) => (i === index ? { ...item, status: value } : item)),
                  )
                }
                options={FULFILLMENT_OPTIONS}
              />
            </Field>
          </div>
        ))}
      </div>
    </FormDialog>
  );
}
