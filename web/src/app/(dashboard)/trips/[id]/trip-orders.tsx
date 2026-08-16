"use client";

import { Plus } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { OrderStatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useOrders } from "@/hooks/use-orders";
import { formatDate, formatIDR, formatNumber, toNumber } from "@/lib/utils";
import type { Trip } from "@/types/api";

export function TripOrders({ trip }: { trip: Trip }) {
  const [page, setPage] = useState(1);
  const { data, isLoading, error } = useOrders({ trip_id: trip.id, page });

  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="space-y-1">
          <h2 className="text-lg font-semibold">Order pada trip ini</h2>
          <p className="text-sm text-muted-foreground">
            {formatNumber(data?.meta.total ?? 0)} order dari {formatNumber(trip.total_customers ?? 0)}{" "}
            customer
          </p>
        </div>
        <Button asChild>
          <Link href={`/orders/new?trip_id=${trip.id}`}>
            <Plus />
            Catat Order
          </Link>
        </Button>
      </div>

      <ErrorState error={error} />

      <div>
        <DataTable
          columns={6}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada order"
          emptyDescription="Catat pesanan yang masuk dari sosial media di sini."
          emptyAction={
            <Button asChild>
              <Link href={`/orders/new?trip_id=${trip.id}`}>
                <Plus />
                Catat Order
              </Link>
            </Button>
          }
          head={
            <TR>
              <TH>Order</TH>
              <TH className="hidden sm:table-cell">Customer</TH>
              <TH className="hidden text-right lg:table-cell">Item</TH>
              <TH className="text-right">Total</TH>
              <TH className="hidden text-right sm:table-cell">Sisa Bayar</TH>
              <TH>Status</TH>
            </TR>
          }
        >
          {data?.items.map((order) => (
            <TR key={order.id}>
              <TD>
                <Link href={`/orders/${order.id}`} className="font-medium hover:underline">
                  {order.order_number}
                </Link>
                <p className="text-xs text-muted-foreground">{formatDate(order.order_date)}</p>
                <p className="text-xs text-muted-foreground sm:hidden">{order.customer_name}</p>
              </TD>
              <TD className="hidden sm:table-cell">
                <p className="font-medium">{order.customer_name}</p>
                <p className="text-xs text-muted-foreground">{order.customer_code}</p>
              </TD>
              <TD className="tabular hidden text-right lg:table-cell">
                {formatNumber(order.total_qty ?? 0)} pcs
                <p className="text-xs text-muted-foreground">
                  {formatNumber(order.item_count ?? 0)} produk
                </p>
              </TD>
              <TD className="tabular text-right font-medium">{formatIDR(order.total)}</TD>
              <TD
                className={`tabular hidden text-right sm:table-cell ${
                  toNumber(order.balance_due) > 0 ? "font-medium text-amber-600" : "text-muted-foreground"
                }`}
              >
                {formatIDR(order.balance_due)}
              </TD>
              <TD>
                <OrderStatusBadge status={order.status} settled={toNumber(order.balance_due) <= 0} />
              </TD>
            </TR>
          ))}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}
