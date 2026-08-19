"use client";

import { Download, MessageCircle } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { FilterSelect } from "@/components/filter-select";
import { useHasRole } from "@/components/layout/user-context";
import { OrderSourceBadge, OrderStatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { ErrorState, PageHeader } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { StatCard } from "@/components/ui/stat-card";
import { CopyButton } from "@/components/copy-button";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  csvUrl,
  useChannelSales,
  useCustomerSales,
  useOrderProfits,
  useProductSales,
  useReceivables,
} from "@/hooks/use-reports";
import { useTrips } from "@/hooks/use-trips";
import { formatDate, formatIDR, formatNumber, toNumber } from "@/lib/utils";

export default function ReportsPage() {
  // Laporan margin hanya boleh dilihat owner. Tabnya disembunyikan dari admin
  // supaya mereka tidak membuka tab yang pasti dijawab "tidak punya akses" oleh
  // backend. Penjagaan sesungguhnya tetap ada di server.
  const canSeeMargin = useHasRole("owner");

  return (
    <>
      <PageHeader
        title="Laporan"
        description={
          canSeeMargin
            ? "Piutang berjalan, margin per order, performa produk, customer, dan channel"
            : "Piutang berjalan serta performa produk, customer, dan channel"
        }
      />

      <Tabs defaultValue="piutang">
        <TabsList>
          <TabsTrigger value="customer">Per Customer</TabsTrigger>
          <TabsTrigger value="channel">Per Channel</TabsTrigger>
          <TabsTrigger value="piutang">Piutang</TabsTrigger>
          {canSeeMargin && <TabsTrigger value="profit">Profit per Order</TabsTrigger>}
          <TabsTrigger value="produk">Performa Produk</TabsTrigger>
        </TabsList>

        {/* Urutan isi disamakan dengan urutan tabnya supaya berkas ini tetap
            terbaca berurutan; Radix memilih isi berdasarkan value, bukan posisi. */}
        <TabsContent value="customer">
          <CustomerSalesReport />
        </TabsContent>
        <TabsContent value="channel">
          <ChannelSalesReport />
        </TabsContent>
        <TabsContent value="piutang">
          <ReceivablesReport />
        </TabsContent>
        {canSeeMargin && (
          <TabsContent value="profit">
            <OrderProfitReport />
          </TabsContent>
        )}
        <TabsContent value="produk">
          <ProductSalesReport />
        </TabsContent>
      </Tabs>
    </>
  );
}

function ReceivablesReport() {
  const [page, setPage] = useState(1);
  const { data, isLoading, error } = useReceivables({ page });

  const totalOutstanding =
    data?.items.reduce((sum, item) => sum + toNumber(item.balance_due), 0) ?? 0;

  return (
    <>
      <ErrorState error={error} />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Order yang masih punya sisa tagihan, diurutkan dari yang paling lama menunggu.
        </p>
        <Button variant="outline" size="sm" asChild>
          <a href={csvUrl("/reports/receivables")} download>
            <Download />
            Ekspor CSV
          </a>
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <StatCard
          label="Total piutang (halaman ini)"
          value={formatIDR(totalOutstanding)}
          tone={totalOutstanding > 0 ? "warning" : "success"}
        />
        <StatCard label="Jumlah order menunggak" value={formatNumber(data?.meta.total ?? 0)} />
      </div>

      <div>
        <DataTable
          columns={7}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Tidak ada piutang"
          emptyDescription="Semua order sudah lunas. "
          head={
            <TR>
              <TH>Order</TH>
              <TH className="hidden sm:table-cell">Customer</TH>
              <TH className="hidden text-right lg:table-cell">Total</TH>
              <TH className="hidden text-right lg:table-cell">Sudah bayar</TH>
              <TH className="text-right">Sisa</TH>
              <TH className="hidden text-right sm:table-cell">Umur</TH>
              <TH className="w-16 text-right">Aksi</TH>
            </TR>
          }
        >
          {data?.items.map((item) => (
            <TR key={item.order_id}>
              <TD>
                <div className="flex items-center gap-0.5">
                  <Link
                    href={`/orders/${item.order_id}`}
                    className="font-medium whitespace-nowrap hover:underline"
                  >
                    {item.order_number}
                  </Link>
                  <CopyButton value={item.order_number} label="Nomor order" />
                </div>
                <div className="mt-1">
                  <OrderStatusBadge status={item.status} />
                </div>
                {/* Nama customer dan umur tagihan menyusul nomor order saat
                    kolomnya disembunyikan — dua hal itu yang menentukan siapa
                    yang perlu ditagih lebih dulu. */}
                <p className="mt-1 text-xs text-muted-foreground sm:hidden">
                  {item.customer_name} · {formatNumber(item.days_outstanding)} hari
                </p>
              </TD>
              <TD className="hidden sm:table-cell">
                <p className="text-sm font-medium">{item.customer_name}</p>
                <p className="text-xs text-muted-foreground">{item.customer_phone}</p>
              </TD>
              <TD className="tabular hidden text-right lg:table-cell">{formatIDR(item.total)}</TD>
              <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
                {formatIDR(item.paid_amount)}
              </TD>
              <TD className="tabular text-right font-semibold text-amber-600">
                {formatIDR(item.balance_due)}
              </TD>
              <TD className="tabular hidden text-right sm:table-cell">
                <span className={item.days_outstanding > 14 ? "font-medium text-destructive" : ""}>
                  {formatNumber(item.days_outstanding)} hari
                </span>
                <p className="text-xs text-muted-foreground">{formatDate(item.order_date)}</p>
              </TD>
              {/* Menagih adalah tindakan, bukan keterangan customer, jadi
                  tempatnya di kolom aksi bersama tombol baris lain. */}
              <TD className="text-right">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      className="text-emerald-600 hover:text-emerald-700"
                      asChild
                    >
                      <a
                        href={`https://wa.me/${item.customer_phone}`}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <MessageCircle />
                        <span className="sr-only">Tagih via WhatsApp</span>
                      </a>
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Tagih via WhatsApp</TooltipContent>
                </Tooltip>
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}

function OrderProfitReport() {
  const [page, setPage] = useState(1);
  const [tripId, setTripId] = useState("");

  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];
  const { data, isLoading, error } = useOrderProfits({ page, trip_id: tripId || undefined });

  return (
    <>
      <ErrorState error={error} />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <FilterSelect
          value={tripId}
          onChange={(value) => {
            setTripId(value);
            setPage(1);
          }}
          allLabel="Semua trip"
          options={tripOptions}
          className="sm:w-64"
        />

        <Button variant="outline" size="sm" asChild>
          <a href={csvUrl("/reports/orders", { trip_id: tripId || undefined })} download>
            <Download />
            Ekspor CSV
          </a>
        </Button>
      </div>

      <p className="text-sm text-muted-foreground">
        HPP diambil dari biaya belanja yang benar-benar dialokasikan ke order. Order yang HPP-nya
        masih nol berarti pembeliannya belum diinput.
      </p>

      <div>
        <DataTable
          columns={6}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada data"
          head={
            <TR>
              <TH>Order</TH>
              <TH className="hidden sm:table-cell">Customer</TH>
              <TH className="hidden text-right sm:table-cell">Omzet</TH>
              <TH className="hidden text-right lg:table-cell">HPP</TH>
              <TH className="text-right">Profit</TH>
              <TH className="text-right">Margin</TH>
            </TR>
          }
        >
          {data?.items.map((item) => {
            const profit = toNumber(item.profit);
            return (
              <TR key={item.order_id}>
                <TD>
                  <div className="flex items-center gap-0.5">
                    <Link
                      href={`/orders/${item.order_id}`}
                      className="font-medium whitespace-nowrap hover:underline"
                    >
                      {item.order_number}
                    </Link>
                    <CopyButton value={item.order_number} label="Nomor order" />
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {item.trip_code} · {formatDate(item.order_date)}
                  </p>
                  <p className="text-xs text-muted-foreground sm:hidden">{item.customer_name}</p>
                </TD>
                <TD className="hidden text-sm sm:table-cell">{item.customer_name}</TD>
                <TD className="tabular hidden text-right sm:table-cell">
                  {formatIDR(item.revenue)}
                </TD>
                <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
                  {formatIDR(item.cogs)}
                </TD>
                <TD
                  className={`tabular text-right font-semibold ${
                    profit >= 0 ? "text-emerald-600" : "text-red-600"
                  }`}
                >
                  {formatIDR(item.profit)}
                </TD>
                <TD className="tabular text-right text-sm">{item.margin_percent}%</TD>
              </TR>
            );
          })}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}

function ProductSalesReport() {
  const [tripId, setTripId] = useState("");
  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];
  const { data, isLoading, error } = useProductSales({ trip_id: tripId || undefined, limit: 50 });

  return (
    <>
      <ErrorState error={error} />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <FilterSelect
          value={tripId}
          onChange={(value) => {
            setTripId(value);
          }}
          allLabel="Semua trip"
          options={tripOptions}
          className="sm:w-64"
        />

        <Button variant="outline" size="sm" asChild>
          <a href={csvUrl("/reports/products", { trip_id: tripId || undefined, limit: 200 })} download>
            <Download />
            Ekspor CSV
          </a>
        </Button>
      </div>

      <DataTable
        columns={6}
        isLoading={isLoading}
        isEmpty={!isLoading && (data?.length ?? 0) === 0}
        emptyTitle="Belum ada penjualan"
        head={
          <TR>
            <TH>Produk</TH>
            <TH className="text-right">Terjual</TH>
            <TH className="hidden text-right lg:table-cell">Order</TH>
            <TH className="hidden text-right sm:table-cell">Omzet</TH>
            <TH className="hidden text-right lg:table-cell">HPP</TH>
            <TH className="text-right">Profit</TH>
          </TR>
        }
      >
        {data?.map((item) => {
          const profit = toNumber(item.profit);
          return (
            <TR key={item.product_id}>
              <TD>
                <p className="font-medium">{item.product_name}</p>
                <p className="text-xs text-muted-foreground">
                  {item.product_sku}
                  {item.category_name ? ` · ${item.category_name}` : ""}
                </p>
              </TD>
              <TD className="tabular text-right font-medium">{formatNumber(item.qty_sold)}</TD>
              <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
                {formatNumber(item.order_count)}
              </TD>
              <TD className="tabular hidden text-right sm:table-cell">{formatIDR(item.revenue)}</TD>
              <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
                {formatIDR(item.cogs)}
              </TD>
              <TD
                className={`tabular text-right font-semibold ${
                  profit >= 0 ? "text-emerald-600" : "text-red-600"
                }`}
              >
                {formatIDR(item.profit)}
              </TD>
            </TR>
          );
        })}
      </DataTable>
    </>
  );
}

function CustomerSalesReport() {
  const [page, setPage] = useState(1);
  const [tripId, setTripId] = useState("");
  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];
  const { data, isLoading, error } = useCustomerSales({
    page,
    trip_id: tripId || undefined,
  });

  return (
    <>
      <ErrorState error={error} />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <FilterSelect
          value={tripId}
          onChange={(value) => {
            setTripId(value);
            setPage(1);
          }}
          allLabel="Semua trip"
          options={tripOptions}
          className="sm:w-64"
        />

        <Button variant="outline" size="sm" asChild>
          <a href={csvUrl("/reports/customers", { trip_id: tripId || undefined })} download>
            <Download />
            Ekspor CSV
          </a>
        </Button>
      </div>

      <p className="text-sm text-muted-foreground">
        Diurutkan dari pembelanja terbesar. Berguna untuk menentukan siapa yang layak dapat
        prioritas slot atau harga khusus pada trip berikutnya.{" "}
        {/* Disebutkan supaya jumlah baris berkasnya tidak dikira keliru: satu
            customer yang memesan lewat dua kanal muncul dua kali. */}
        <span className="text-muted-foreground/80">
          Berkas CSV-nya dipecah per channel, jadi satu customer bisa muncul lebih dari sekali.
        </span>
      </p>

      <div>
        <DataTable
          columns={7}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada transaksi"
          emptyDescription="Customer akan muncul di sini setelah ordernya tercatat."
          head={
            <TR>
              <TH className="min-w-40">Customer</TH>
              <TH className="hidden w-20 text-right sm:table-cell">Order</TH>
              <TH className="hidden w-20 text-right lg:table-cell">Pcs</TH>
              <TH className="w-32 text-right">Omzet</TH>
              <TH className="hidden w-32 text-right lg:table-cell">Rata-rata</TH>
              <TH className="hidden w-32 text-right sm:table-cell">Profit</TH>
              <TH className="hidden w-32 text-right sm:table-cell">Piutang</TH>
            </TR>
          }
        >
          {data?.items.map((item) => {
            const profit = toNumber(item.profit);
            const outstanding = toNumber(item.outstanding);
            return (
              <TR key={item.customer_id}>
                <TD>
                  <Link
                    href={`/customers?q=${encodeURIComponent(item.customer_code)}`}
                    className="font-medium hover:underline"
                  >
                    {item.customer_name}
                  </Link>
                  <p className="text-xs text-muted-foreground">
                    {item.customer_code}
                    {item.city ? ` · ${item.city}` : ""}
                    {item.last_order_at ? ` · terakhir ${formatDate(item.last_order_at)}` : ""}
                  </p>
                </TD>
                <TD className="tabular hidden text-right sm:table-cell">
                  {formatNumber(item.order_count)}
                </TD>
                <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
                  {formatNumber(item.item_qty)}
                </TD>
                <TD className="tabular text-right font-medium">{formatIDR(item.revenue)}</TD>
                <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
                  {formatIDR(item.avg_order_value)}
                </TD>
                <TD
                  className={`tabular hidden text-right font-semibold sm:table-cell ${
                    profit >= 0 ? "text-emerald-600" : "text-red-600"
                  }`}
                >
                  {formatIDR(item.profit)}
                </TD>
                <TD
                  className={`tabular hidden text-right sm:table-cell ${
                    outstanding > 0 ? "font-medium text-amber-600" : "text-muted-foreground"
                  }`}
                >
                  {formatIDR(item.outstanding)}
                </TD>
              </TR>
            );
          })}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}

function ChannelSalesReport() {
  const [tripId, setTripId] = useState("");
  const { data: trips } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];
  const { data, isLoading, error } = useChannelSales({ trip_id: tripId || undefined });

  const totalRevenue = data?.reduce((sum, row) => sum + toNumber(row.revenue), 0) ?? 0;
  const totalOrders = data?.reduce((sum, row) => sum + row.order_count, 0) ?? 0;

  return (
    <>
      <ErrorState error={error} />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <FilterSelect
          value={tripId}
          onChange={(value) => {
            setTripId(value);
          }}
          allLabel="Semua trip"
          options={tripOptions}
          className="sm:w-64"
        />

        <Button variant="outline" size="sm" asChild>
          <a href={csvUrl("/reports/channels", { trip_id: tripId || undefined })} download>
            <Download />
            Ekspor CSV
          </a>
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <StatCard label="Total omzet" value={formatIDR(totalRevenue)} />
        <StatCard label="Total order" value={formatNumber(totalOrders)} />
      </div>

      <DataTable
        columns={7}
        isLoading={isLoading}
        isEmpty={!isLoading && (data?.length ?? 0) === 0}
        emptyTitle="Belum ada penjualan"
        emptyDescription="Isi kolom channel saat mencatat pesanan supaya rekap ini terisi."
        head={
          <TR>
            <TH className="w-32">Channel</TH>
            <TH className="w-20 text-right">Order</TH>
            <TH className="hidden w-24 text-right lg:table-cell">Customer</TH>
            <TH className="w-32 text-right">Omzet</TH>
            <TH className="hidden w-32 text-right lg:table-cell">Rata-rata</TH>
            <TH className="hidden w-32 text-right sm:table-cell">Profit</TH>
            <TH className="hidden min-w-40 sm:table-cell">Porsi omzet</TH>
          </TR>
        }
      >
        {data?.map((row) => {
          const profit = toNumber(row.profit);
          const share = toNumber(row.revenue_share);
          return (
            <TR key={row.order_source}>
              <TD>
                <OrderSourceBadge source={row.order_source} />
              </TD>
              <TD className="tabular text-right">{formatNumber(row.order_count)}</TD>
              <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
                {formatNumber(row.customer_count)}
              </TD>
              <TD className="tabular text-right font-medium">{formatIDR(row.revenue)}</TD>
              <TD className="tabular hidden text-right text-muted-foreground lg:table-cell">
                {formatIDR(row.avg_order_value)}
              </TD>
              <TD
                className={`tabular hidden text-right font-semibold sm:table-cell ${
                  profit >= 0 ? "text-emerald-600" : "text-red-600"
                }`}
              >
                {formatIDR(row.profit)}
              </TD>
              <TD className="hidden sm:table-cell">
                {/* Bar sederhana supaya perbandingan antar channel langsung terbaca. */}
                <div className="flex items-center gap-2">
                  <div className="h-2 flex-1 overflow-hidden rounded-full bg-muted">
                    <div
                      className="h-full rounded-full bg-primary"
                      style={{ width: `${Math.min(share, 100)}%` }}
                    />
                  </div>
                  <span className="tabular w-12 text-right text-xs text-muted-foreground">
                    {share.toFixed(1)}%
                  </span>
                </div>
              </TD>
            </TR>
          );
        })}
      </DataTable>
    </>
  );
}
