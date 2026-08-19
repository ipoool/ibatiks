"use client";

import Link from "next/link";

import { CopyButton } from "@/components/copy-button";
import { OrderStatusBadge } from "@/components/status-badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useCustomerStats } from "@/hooks/use-master";
import { useOrders } from "@/hooks/use-orders";
import { formatDate, formatIDR, toNumber } from "@/lib/utils";
import type { Customer } from "@/types/api";

/** Sepuluh order terakhir cukup untuk menilai seorang customer; sisanya di menu Order. */
const JUMLAH_DITAMPILKAN = 10;

/**
 * Riwayat belanja seorang customer, dibuka dari daftar customer.
 *
 * Pertanyaan yang dijawab dialog ini muncul saat mencatat order: pelanggan lama
 * atau baru, pernah menunggak atau tidak, terakhir belanja kapan. Sebelumnya
 * jawabannya harus dicari dengan pindah ke menu Order lalu mengetik namanya di
 * pencarian — dan orang yang sedang mengetik order jarang mau meninggalkan
 * formulirnya untuk itu.
 *
 * Angka ringkasannya diambil dari endpoint stats, bukan dijumlah dari daftar di
 * bawahnya. Daftarnya hanya memuat sepuluh baris pertama, jadi menjumlahkannya
 * sendiri akan menghasilkan total yang terlihat pasti padahal salah begitu
 * customernya punya lebih dari sepuluh order.
 */
export function DialogRiwayatOrder({
  customer,
  onClose,
}: {
  customer: Customer;
  onClose: () => void;
}) {
  const stats = useCustomerStats(customer.id);
  const orders = useOrders({ customer_id: customer.id, per_page: JUMLAH_DITAMPILKAN });

  const daftar = orders.data?.items ?? [];
  const total = orders.data?.meta?.total ?? 0;
  const memuat = stats.isLoading || orders.isLoading;

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Riwayat Order</DialogTitle>
          <DialogDescription>
            {customer.name} · {customer.code}
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-3 gap-3 rounded-lg border border-border bg-muted/40 p-3">
          <Ringkas label="Total order" nilai={memuat ? "…" : String(stats.data?.total_orders ?? 0)} />
          <Ringkas
            label="Total belanja"
            nilai={memuat ? "…" : formatIDR(stats.data?.total_spent ?? 0)}
          />
          <Ringkas
            label="Terakhir order"
            nilai={
              memuat ? "…" : stats.data?.last_order_at ? formatDate(stats.data.last_order_at) : "—"
            }
          />
        </div>

        {memuat ? (
          <p className="py-8 text-center text-sm text-muted-foreground">Memuat riwayat…</p>
        ) : daftar.length === 0 ? (
          <div className="py-8 text-center">
            <p className="text-sm font-medium">Belum ada order</p>
            <p className="text-xs text-muted-foreground">
              Order customer ini akan muncul di sini setelah dicatat.
            </p>
          </div>
        ) : (
          <div className="scrollbar-thin max-h-[380px] divide-y divide-border overflow-y-auto">
            {daftar.map((order) => {
              const sisa = toNumber(order.balance_due);
              return (
                <div
                  key={order.id}
                  className="flex items-start justify-between gap-3 py-2.5 first:pt-0"
                >
                  <div className="min-w-0">
                    <div className="flex items-center gap-0.5">
                      <Link
                        href={`/orders/${order.id}`}
                        onClick={onClose}
                        className="text-sm font-medium whitespace-nowrap hover:underline"
                      >
                        {order.order_number}
                      </Link>
                      <CopyButton value={order.order_number} label="Nomor order" />
                    </div>
                    <p className="truncate text-xs text-muted-foreground">
                      {formatDate(order.order_date)} · {order.trip_code}
                    </p>
                  </div>
                  <div className="shrink-0 text-right">
                    <p className="tabular text-sm font-medium">{formatIDR(order.total)}</p>
                    {/* Sisa tagihan hanya disebut kalau memang ada. Pada order
                        yang sudah lunas, "sisa Rp0" cuma menambah baris yang
                        harus dibaca untuk memastikan tidak ada apa-apa. */}
                    {sisa > 0 && order.status !== "cancelled" && (
                      <p className="tabular text-xs text-amber-600">sisa {formatIDR(sisa)}</p>
                    )}
                  </div>
                  <div className="shrink-0">
                    <OrderStatusBadge status={order.status} settled={sisa <= 0} />
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {total > daftar.length && (
          <p className="text-xs text-muted-foreground">
            Menampilkan {daftar.length} order terakhir dari {total}. Selebihnya ada di menu Order.
          </p>
        )}
      </DialogContent>
    </Dialog>
  );
}

function Ringkas({ label, nilai }: { label: string; nilai: string }) {
  return (
    <div className="min-w-0">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="truncate text-sm font-semibold">{nilai}</p>
    </div>
  );
}
