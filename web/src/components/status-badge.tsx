"use client";

import { Badge, type badgeVariants } from "@/components/ui/badge";
import type {
  FulfillmentStatus,
  InvoiceStatus,
  OrderSource,
  OrderStatus,
  ShipmentStatus,
  TripStatus,
} from "@/types/api";
import type { VariantProps } from "class-variance-authority";

/**
 * Lencana status khusus domain jastip.
 *
 * Dipisahkan dari primitif Badge shadcn karena isinya kosakata bisnis, bukan
 * urusan tampilan: label yang dilihat admin, dan warna yang mewakili artinya.
 * Primitifnya tetap bisa diperbarui dari registry shadcn tanpa menyentuh berkas
 * ini.
 */
type BadgeTone = NonNullable<VariantProps<typeof badgeVariants>["variant"]>;

/*
 * Label dan warna status dipusatkan di sini supaya istilah yang dilihat admin
 * konsisten di seluruh halaman. Warna mengikuti arti: abu-abu untuk yang belum
 * berjalan, biru/ungu untuk yang sedang berjalan, hijau untuk selesai, kuning
 * untuk butuh perhatian, merah untuk batal.
 */

const ORDER_STATUS: Record<OrderStatus, { label: string; tone: BadgeTone }> = {
  awaiting_dp: { label: "Menunggu DP", tone: "warning" },
  // "Diproses" memuat seluruh pekerjaan di tengah: belanja, penerimaan barang,
  // pengemasan, penetapan ongkir, sampai invoice pelunasan terbit. Status
  // inilah yang masuk hitungan daftar belanja tripper.
  dp_paid: { label: "Diproses", tone: "info" },
  paid: { label: "Pembayaran Lunas", tone: "success" },
  shipped: { label: "Dikirim", tone: "info" },
  completed: { label: "Selesai", tone: "success" },
  cancelled: { label: "Batal", tone: "danger" },
};

const TRIP_STATUS: Record<TripStatus, { label: string; tone: BadgeTone }> = {
  open: { label: "Open", tone: "success" },
  closed: { label: "Closed", tone: "warning" },
};

const INVOICE_STATUS: Record<InvoiceStatus, { label: string; tone: BadgeTone }> = {
  draft: { label: "Draft", tone: "neutral" },
  sent: { label: "Terkirim", tone: "info" },
  paid: { label: "Lunas", tone: "success" },
  void: { label: "Dibatalkan", tone: "danger" },
};

const SHIPMENT_STATUS: Record<ShipmentStatus, { label: string; tone: BadgeTone }> = {
  packing: { label: "Dikemas", tone: "neutral" },
  ready: { label: "Siap Kirim", tone: "warning" },
  shipped: { label: "Dikirim", tone: "info" },
  delivered: { label: "Diterima", tone: "success" },
  returned: { label: "Retur", tone: "danger" },
};

const FULFILLMENT_STATUS: Record<FulfillmentStatus, { label: string; tone: BadgeTone }> = {
  pending: { label: "Belum Dibeli", tone: "neutral" },
  purchased: { label: "Sudah Dibeli", tone: "success" },
  partial: { label: "Sebagian", tone: "warning" },
  unavailable: { label: "Tidak Ada", tone: "danger" },
  refunded: { label: "Direfund", tone: "danger" },
};

/*
 * Asal order dibedakan warnanya supaya sekilas terlihat channel mana yang
 * paling ramai saat menelusuri daftar order.
 */
const ORDER_SOURCE: Record<OrderSource, { label: string; tone: BadgeTone }> = {
  whatsapp: { label: "WhatsApp", tone: "success" },
  instagram: { label: "Instagram", tone: "progress" },
  tiktok: { label: "TikTok", tone: "info" },
  marketplace: { label: "Marketplace", tone: "warning" },
  lainnya: { label: "Lainnya", tone: "neutral" },
};

export function OrderSourceBadge({ source }: { source: OrderSource }) {
  const meta = ORDER_SOURCE[source] ?? { label: source, tone: "neutral" as BadgeTone };
  return <Badge variant={meta.tone}>{meta.label}</Badge>;
}

/**
 * Status order, dengan penanda lunas terpisah bila perlu.
 *
 * Status order menceritakan posisi barangnya — belum dibeli, sudah dikemas,
 * sudah dikirim — jadi order yang dilunasi saat barangnya belum siap tidak
 * boleh melompat ke "Lunas": kalau itu terjadi, ia muncul di antrean siap
 * kirim padahal barangnya belum ada. Karena itu pelunasan lebih awal ditandai
 * chip tersendiri, supaya admin tetap melihat uangnya sudah masuk.
 */
export function OrderStatusBadge({
  status,
  settled = false,
}: {
  status: OrderStatus;
  /** Sisa tagihannya sudah nol walau statusnya belum "paid". */
  settled?: boolean;
}) {
  const meta = ORDER_STATUS[status] ?? { label: status, tone: "neutral" as BadgeTone };
  const showSettled =
    settled && !["paid", "shipped", "completed", "cancelled"].includes(status);

  return (
    <span className="inline-flex flex-wrap items-center gap-1">
      <Badge variant={meta.tone}>{meta.label}</Badge>
      {showSettled && <Badge variant="success">Lunas</Badge>}
    </span>
  );
}

export function TripStatusBadge({ status }: { status: TripStatus }) {
  const meta = TRIP_STATUS[status] ?? { label: status, tone: "neutral" as BadgeTone };
  return <Badge variant={meta.tone}>{meta.label}</Badge>;
}

export function InvoiceStatusBadge({ status }: { status: InvoiceStatus }) {
  const meta = INVOICE_STATUS[status] ?? { label: status, tone: "neutral" as BadgeTone };
  return <Badge variant={meta.tone}>{meta.label}</Badge>;
}

export function ShipmentStatusBadge({ status }: { status: ShipmentStatus }) {
  const meta = SHIPMENT_STATUS[status] ?? { label: status, tone: "neutral" as BadgeTone };
  return <Badge variant={meta.tone}>{meta.label}</Badge>;
}

export function FulfillmentBadge({ status }: { status: FulfillmentStatus }) {
  const meta = FULFILLMENT_STATUS[status] ?? { label: status, tone: "neutral" as BadgeTone };
  return <Badge variant={meta.tone}>{meta.label}</Badge>;
}

export const orderStatusLabel = (status: OrderStatus) => ORDER_STATUS[status]?.label ?? status;
export const tripStatusLabel = (status: TripStatus) => TRIP_STATUS[status]?.label ?? status;

export const orderSourceLabel = (source: OrderSource) => ORDER_SOURCE[source]?.label ?? source;

export const ORDER_SOURCE_OPTIONS = Object.entries(ORDER_SOURCE).map(([value, meta]) => ({
  value: value as OrderSource,
  label: meta.label,
}));

export const ORDER_STATUS_OPTIONS = Object.entries(ORDER_STATUS).map(([value, meta]) => ({
  value: value as OrderStatus,
  label: meta.label,
}));

export const TRIP_STATUS_OPTIONS = Object.entries(TRIP_STATUS).map(([value, meta]) => ({
  value: value as TripStatus,
  label: meta.label,
}));
