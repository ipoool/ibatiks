"use client";

import {
  AlertTriangle,
  Boxes,
  Plane,
  Receipt,
  TrendingUp,
  Truck,
  Wallet,
} from "lucide-react";
import Link from "next/link";

import { OrderStatusBadge, TripStatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ErrorState, PageHeader } from "@/components/ui/page";
import { StatCard } from "@/components/ui/stat-card";
import { CopyButton } from "@/components/copy-button";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useDashboard } from "@/hooks/use-reports";
import {
  formatDate,
  formatIDR,
  formatIDRCompact,
  formatNumber,
  toNumber,
} from "@/lib/utils";

export default function DashboardPage() {
  const { data, isLoading, error } = useDashboard();

  const profit = toNumber(data?.profit_this_month);

  return (
    <>
      <PageHeader
        title="Dashboard"
        description="Ringkasan kondisi bisnis jastip kamu hari ini"
        actions={
          <Button asChild>
            <Link href="/orders/new">
              <Receipt />
              Catat Order
            </Link>
          </Button>
        }
      />

      {/*
        Ringkasan tidak ikut dirender kalau datanya gagal diambil. Kartu-kartu
        di bawah menampilkan nol saat kosong, dan nol yang sebenarnya berarti
        "tidak terbaca" terbaca sebagai "tokonya sepi" — kesimpulan yang salah,
        dan tidak ada apa pun di layar yang membantahnya.

        Orang yang memang tidak berhak melihat Dashboard sudah dialihkan ke menu
        pertamanya oleh middleware, jadi yang sampai ke sini biasanya gangguan
        jaringan sesaat.
      */}
      {error ? (
        <ErrorState error={error} />
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <StatCard
              label="Trip aktif"
              value={formatNumber(data?.active_trips)}
              hint="Sedang buka order atau dalam perjalanan"
              icon={Plane}
              isLoading={isLoading}
            />
            <StatCard
              label="Order berjalan"
              value={formatNumber(data?.open_orders)}
              hint={`${formatNumber(data?.orders_this_month)} order bulan ini`}
              icon={Receipt}
              isLoading={isLoading}
            />
            <StatCard
              label="Siap dikirim"
              value={formatNumber(data?.pending_shipment)}
              hint="Sudah lunas, menunggu resi"
              icon={Truck}
              tone={(data?.pending_shipment ?? 0) > 0 ? "warning" : "default"}
              isLoading={isLoading}
            />
            <StatCard
              label="Piutang berjalan"
              value={formatIDRCompact(data?.outstanding)}
              hint="Total sisa tagihan customer"
              icon={AlertTriangle}
              tone={toNumber(data?.outstanding) > 0 ? "warning" : "default"}
              isLoading={isLoading}
            />
          </div>

          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            <StatCard
              label="Omzet bulan ini"
              value={formatIDR(data?.revenue_this_month)}
              icon={Wallet}
              isLoading={isLoading}
            />
            <StatCard
              label="Laba kotor bulan ini"
              value={formatIDR(data?.profit_this_month)}
              hint="Omzet dikurangi HPP barang yang sudah dibeli"
              icon={TrendingUp}
              tone={profit >= 0 ? "success" : "danger"}
              isLoading={isLoading}
            />
            <StatCard
              label="Nilai stok"
              value={formatIDR(data?.stock_value)}
              hint={`${formatNumber(data?.stock_qty)} unit siap dijual`}
              icon={Boxes}
              isLoading={isLoading}
            />
          </div>

          <div className="grid gap-4 lg:grid-cols-3">
            <Card className="lg:col-span-2">
              <CardHeader>
                <CardTitle>Order terbaru</CardTitle>
                <CardAction>
                  <Button variant="ghost" size="sm" asChild>
                    <Link href="/orders">Lihat semua</Link>
                  </Button>
                </CardAction>
              </CardHeader>
              <CardContent>
                <DataTable
                  columns={4}
                  isLoading={isLoading}
                  isEmpty={
                    !isLoading && (data?.recent_orders?.length ?? 0) === 0
                  }
                  emptyTitle="Belum ada order"
                  emptyDescription="Order yang kamu catat akan muncul di sini."
                  head={
                    <TR>
                      <TH>Order</TH>
                      <TH className="hidden sm:table-cell">Customer</TH>
                      <TH className="text-right">Total</TH>
                      <TH>Status</TH>
                    </TR>
                  }
                >
                  {data?.recent_orders?.map((order) => (
                    <TR key={order.id}>
                      <TD>
                        <div className="flex items-center gap-0.5">
                          <Link
                            href={`/orders/${order.id}`}
                            className="font-medium whitespace-nowrap hover:underline"
                          >
                            {order.order_number}
                          </Link>
                          <CopyButton
                            value={order.order_number}
                            label="Nomor order"
                          />
                        </div>
                        <p className="text-xs text-muted-foreground">
                          {formatDate(order.order_date)}
                        </p>
                        <p className="text-xs text-muted-foreground sm:hidden">
                          {order.customer_name}
                        </p>
                      </TD>
                      <TD className="hidden sm:table-cell">
                        <p className="font-medium">{order.customer_name}</p>
                        <p className="text-xs text-muted-foreground">
                          {order.trip_code}
                        </p>
                      </TD>
                      <TD className="tabular text-right font-medium">
                        {formatIDR(order.total)}
                      </TD>
                      <TD>
                        <OrderStatusBadge status={order.status} />
                      </TD>
                    </TR>
                  ))}
                </DataTable>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Trip mendatang</CardTitle>
                <CardAction>
                  <Button variant="ghost" size="sm" asChild>
                    <Link href="/trips">Lihat semua</Link>
                  </Button>
                </CardAction>
              </CardHeader>
              <CardContent className="space-y-3">
                {isLoading && (
                  <p className="text-sm text-muted-foreground">Memuat…</p>
                )}
                {!isLoading && (data?.upcoming_trips?.length ?? 0) === 0 && (
                  <p className="text-sm text-muted-foreground">
                    Belum ada trip yang dibuka untuk order.
                  </p>
                )}
                {data?.upcoming_trips?.map((trip) => (
                  <Link
                    key={trip.id}
                    href={`/trips/${trip.id}`}
                    className="block rounded-lg border border-border p-3 transition-colors hover:bg-accent"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="truncate font-medium">{trip.title}</p>
                        <p className="text-xs text-muted-foreground">
                          {trip.country} · {formatDate(trip.depart_date)}
                        </p>
                      </div>
                      <TripStatusBadge status={trip.status} />
                    </div>
                  </Link>
                ))}
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Produk terlaris</CardTitle>
            </CardHeader>
            <CardContent>
              <DataTable
                columns={5}
                isLoading={isLoading}
                isEmpty={!isLoading && (data?.top_products?.length ?? 0) === 0}
                emptyTitle="Belum ada penjualan"
                emptyDescription="Angka akan muncul setelah ada order yang tercatat."
                head={
                  <TR>
                    <TH>Produk</TH>
                    <TH className="text-right">Terjual</TH>
                    <TH className="hidden text-right sm:table-cell">Omzet</TH>
                    <TH className="hidden text-right xl:table-cell">HPP</TH>
                    <TH className="text-right">Profit</TH>
                  </TR>
                }
              >
                {data?.top_products?.map((product) => (
                  <TR key={product.product_id}>
                    <TD>
                      <p className="font-medium">{product.product_name}</p>
                      <p className="text-xs text-muted-foreground">
                        {product.product_sku}
                      </p>
                    </TD>
                    <TD className="tabular text-right">
                      {formatNumber(product.qty_sold)}
                    </TD>
                    <TD className="tabular hidden text-right sm:table-cell">
                      {formatIDR(product.revenue)}
                    </TD>
                    <TD className="tabular hidden text-right text-muted-foreground xl:table-cell">
                      {formatIDR(product.cogs)}
                    </TD>
                    <TD
                      className={`tabular text-right font-medium ${
                        toNumber(product.profit) >= 0
                          ? "text-emerald-600"
                          : "text-red-600"
                      }`}
                    >
                      {formatIDR(product.profit)}
                    </TD>
                  </TR>
                ))}
              </DataTable>
            </CardContent>
          </Card>
        </>
      )}
    </>
  );
}
