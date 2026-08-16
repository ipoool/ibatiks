"use client";

import { ArrowLeft, Loader2, MapPin, MessageCircle, Plane, User } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";

import { OrderSourceBadge, OrderStatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardAction } from "@/components/ui/card";
import { DetailRow, ErrorState } from "@/components/ui/page";
import { useOrder } from "@/hooks/use-orders";
import { formatDate, formatIDR, toNumber } from "@/lib/utils";

import { OrderActions, OrderEditButton } from "./order-actions";
import { OrderInvoices } from "./order-invoices";
import { OrderItems } from "./order-items";
import { OrderPayments } from "./order-payments";
import { OrderShipment } from "./order-shipment";

export default function OrderDetailPage() {
  const params = useParams<{ id: string }>();
  const { data: order, isLoading, error } = useOrder(params.id);

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-16 text-muted-foreground">
        <Loader2 className="size-4 animate-spin" />
        Memuat order…
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/orders">
            <ArrowLeft />
            Kembali ke daftar order
          </Link>
        </Button>
        <ErrorState error={error ?? new Error("Order tidak ditemukan")} />
      </div>
    );
  }

  const balanceDue = toNumber(order.balance_due);

  return (
    <>
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild className="-ml-2">
          <Link href="/orders">
            <ArrowLeft />
            Semua order
          </Link>
        </Button>

        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-2xl font-semibold tracking-tight">{order.order_number}</h1>
              <OrderStatusBadge status={order.status} settled={toNumber(order.balance_due) <= 0} />
              <OrderSourceBadge source={order.order_source} />
            </div>
            <p className="text-sm text-muted-foreground">
              {formatDate(order.order_date)} · {order.customer.name} ·{" "}
              <Link href={`/trips/${order.trip_id}`} className="hover:underline">
                {order.trip.code}
              </Link>
            </p>
          </div>

          <OrderActions order={order} />
        </div>
      </div>

      {order.status === "cancelled" && (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">
          <p className="font-medium">Order ini sudah dibatalkan.</p>
          {order.cancel_reason && <p>Alasan: {order.cancel_reason}</p>}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="min-w-0 space-y-6 lg:col-span-2">
          <OrderItems order={order} />
          <OrderPayments order={order} />
          <OrderInvoices order={order} />
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Ringkasan biaya</CardTitle>
              <CardAction>
                <OrderEditButton order={order} />
              </CardAction>
            </CardHeader>
            <CardContent className="divide-y divide-border">
              <DetailRow label="Subtotal" value={formatIDR(order.subtotal)} />
              {toNumber(order.discount) > 0 && (
                <DetailRow label="Diskon" value={`−${formatIDR(order.discount)}`} />
              )}
              {toNumber(order.shipping_fee) > 0 && (
                <DetailRow label="Ongkir" value={formatIDR(order.shipping_fee)} />
              )}
              <DetailRow
                label="Total"
                value={<span className="text-base font-semibold">{formatIDR(order.total)}</span>}
              />
              <DetailRow label="Sudah dibayar" value={formatIDR(order.paid_amount)} />
              <DetailRow
                label="Sisa tagihan"
                value={
                  <span
                    className={`text-base font-semibold ${
                      balanceDue > 0 ? "text-amber-600" : "text-emerald-600"
                    }`}
                  >
                    {formatIDR(order.balance_due)}
                  </span>
                }
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Customer</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex items-start gap-2">
                <User className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                <div className="min-w-0">
                  <Link
                    href={`/customers?q=${encodeURIComponent(order.customer.name)}`}
                    className="font-medium hover:underline"
                  >
                    {order.customer.name}
                  </Link>
                  <p className="text-muted-foreground">{order.customer.code}</p>
                </div>
              </div>

              <div className="flex items-start gap-2">
                <MessageCircle className="mt-0.5 size-4 shrink-0 text-emerald-600" />
                <a
                  href={`https://wa.me/${order.customer.phone_wa}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:underline"
                >
                  {order.customer.phone_wa}
                </a>
              </div>

              <div className="flex items-start gap-2">
                <MapPin className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                <div className="min-w-0">
                  <p className="font-medium">{order.recipient_name}</p>
                  <p className="text-muted-foreground">{order.recipient_phone}</p>
                  <p className="text-muted-foreground">{order.shipping_address}</p>
                  {/* Urutannya mengikuti cara alamat Indonesia ditulis:
                      kelurahan, kecamatan, kota, provinsi, kode pos. */}
                  <p className="text-muted-foreground">
                    {[
                      order.shipping_subdistrict,
                      order.shipping_district,
                      order.shipping_city,
                      order.shipping_province,
                      order.shipping_postal_code,
                    ]
                      .filter(Boolean)
                      .join(", ")}
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-2">
                <Plane className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                <div className="min-w-0">
                  <Link href={`/trips/${order.trip_id}`} className="font-medium hover:underline">
                    {order.trip.title}
                  </Link>
                  <p className="text-muted-foreground">
                    {order.trip.country} · {formatDate(order.trip.depart_date)}
                  </p>
                </div>
              </div>

              {order.notes && (
                <div className="rounded-lg bg-muted/60 p-3 text-muted-foreground">
                  <p className="mb-1 text-xs font-medium uppercase tracking-wide">Catatan</p>
                  {order.notes}
                </div>
              )}
            </CardContent>
          </Card>

          <OrderShipment order={order} />
        </div>
      </div>
    </>
  );
}
