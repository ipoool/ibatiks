"use client";

import { Info, Loader2, TrendingDown, TrendingUp } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DetailRow, ErrorState } from "@/components/ui/page";
import { StatCard } from "@/components/ui/stat-card";
import { useTripProfit } from "@/hooks/use-reports";
import { formatIDR, formatNumber, toNumber } from "@/lib/utils";
import type { ExpenseCategory } from "@/types/api";

const CATEGORY_LABEL: Record<ExpenseCategory, string> = {
  tiket: "Tiket pesawat",
  bagasi: "Bagasi",
  akomodasi: "Akomodasi",
  transport: "Transport lokal",
  visa: "Visa",
  lainnya: "Lainnya",
};

/**
 * Laporan laba-rugi, satu trip atau seluruhnya sekaligus.
 *
 * Dulu hanya ada sebagai tab di detail trip, jadi laba seluruh usaha tidak
 * pernah bisa dibaca dalam satu layar — harus dibuka trip demi trip lalu
 * dijumlah sendiri. Tata letaknya sengaja tidak diubah: yang berpindah cuma
 * tempatnya, dan orang yang sudah hafal susunannya tidak perlu belajar lagi.
 */
export function ProfitLossReport({
  tripId,
  from,
  to,
}: {
  tripId?: string;
  from?: string;
  to?: string;
}) {
  const { data: report, isLoading, error } = useTripProfit({ trip_id: tripId, from, to });

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-12 text-muted-foreground">
        <Loader2 className="size-4 animate-spin" />
        Menghitung laporan…
      </div>
    );
  }

  if (error) {
    return <ErrorState error={error} />;
  }
  if (!report) return null;

  const netProfit = toNumber(report.net_profit);
  const profitable = netProfit >= 0;

  return (
    <>
      <div className="space-y-1">
        <h2 className="text-lg font-semibold">
          {report.trip_code ? `Laba trip ${report.trip_code}` : "Laba seluruh trip"}
        </h2>
        <p className="text-sm text-muted-foreground">
          Dihitung dari omzet order, HPP belanja yang benar-benar terjadi, dan biaya perjalanan.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Omzet" value={formatIDR(report.revenue)} hint={`${formatNumber(report.order_count)} order`} />
        <StatCard label="HPP barang pesanan" value={formatIDR(report.cogs)} />
        <StatCard label="Biaya perjalanan" value={formatIDR(report.trip_expenses)} />
        <StatCard
          label="Laba bersih"
          value={formatIDR(report.net_profit)}
          hint={`Margin ${report.margin_percent}%`}
          icon={profitable ? TrendingUp : TrendingDown}
          tone={profitable ? "success" : "danger"}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Rincian perhitungan</CardTitle>
          </CardHeader>
          <CardContent className="divide-y divide-border">
            <DetailRow label="Omzet order" value={formatIDR(report.revenue)} />
            <DetailRow
              label="HPP barang pesanan"
              value={<span className="text-destructive">−{formatIDR(report.cogs)}</span>}
            />
            <DetailRow
              label="Laba kotor"
              value={<span className="font-semibold">{formatIDR(report.gross_profit)}</span>}
            />
            <DetailRow
              label="Biaya perjalanan"
              value={<span className="text-destructive">−{formatIDR(report.trip_expenses)}</span>}
            />
            <DetailRow
              label="Laba bersih"
              value={
                <span
                  className={`text-base font-semibold ${
                    profitable ? "text-emerald-600" : "text-red-600"
                  }`}
                >
                  {formatIDR(report.net_profit)}
                </span>
              }
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Arus kas & posisi</CardTitle>
          </CardHeader>
          <CardContent className="divide-y divide-border">
            <DetailRow label="Total modal keluar" value={formatIDR(report.total_capital_out)} />
            <DetailRow label="Uang masuk dari customer" value={formatIDR(report.payment_received)} />
            <DetailRow
              label="Sisa tagihan belum masuk"
              value={
                <span className={toNumber(report.outstanding) > 0 ? "text-amber-600" : undefined}>
                  {formatIDR(report.outstanding)}
                </span>
              }
            />
            <DetailRow label="Ongkir ditagihkan ke customer" value={formatIDR(report.shipping_fee_collected)} />
            <DetailRow label="Ongkir dibayar ke kurir" value={formatIDR(report.shipping_cost_paid)} />
            <DetailRow label="Diskon diberikan" value={formatIDR(report.discount_given)} />
          </CardContent>
        </Card>
      </div>

      {report.surplus_stock_qty > 0 && (
        <Card className="border-sky-200 bg-sky-50/60">
          <CardContent className="flex gap-3 py-4">
            <Info className="mt-0.5 size-4 shrink-0 text-sky-600" />
            <div className="space-y-1 text-sm">
              <p className="font-medium text-sky-900">
                {formatNumber(report.surplus_stock_qty)} unit senilai{" "}
                {formatIDR(report.surplus_stock_value)} masuk stok
              </p>
              <p className="text-sky-800">
                Barang yang dibeli lebih banyak dari pesanan tidak dibebankan sebagai HPP trip ini.
                Uangnya memang sudah keluar (ikut terhitung di modal keluar), tapi nilainya masih
                dipegang sebagai barang dan baru menjadi HPP ketika stoknya terjual.
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Lintas trip, biayanya dipecah per trip dulu. Disatukan per kategori,
          "Tiket pesawat Rp18.500.000" tidak memberi tahu itu tiket trip yang
          mana — padahal perbandingan antar tripnya yang dicari saat membuka
          laporan gabungan. */}
      {report.expense_by_trip.length > 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>Rincian biaya perjalanan</CardTitle>
          </CardHeader>
          <CardContent className="divide-y divide-border">
            {report.expense_by_trip.map((trip) => (
              <div key={trip.trip_id} className="py-3 first:pt-0 last:pb-0">
                <div className="flex items-baseline justify-between gap-3">
                  <p className="min-w-0 text-sm font-medium">
                    {trip.trip_code}
                    <span className="ml-1.5 font-normal text-muted-foreground">{trip.title}</span>
                  </p>
                  <p className="tabular shrink-0 text-sm font-semibold">{formatIDR(trip.total)}</p>
                </div>
                <div className="mt-1 space-y-0.5">
                  {trip.items.map((expense) => (
                    <div
                      key={expense.category}
                      className="flex items-baseline justify-between gap-3 text-xs text-muted-foreground"
                    >
                      <span>{CATEGORY_LABEL[expense.category] ?? expense.category}</span>
                      <span className="tabular shrink-0">{formatIDR(expense.amount)}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      ) : (
        report.expense_breakdown.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Rincian biaya perjalanan</CardTitle>
            </CardHeader>
            <CardContent className="divide-y divide-border">
              {report.expense_breakdown.map((expense) => (
                <DetailRow
                  key={expense.category}
                  label={CATEGORY_LABEL[expense.category] ?? expense.category}
                  value={formatIDR(expense.amount)}
                />
              ))}
            </CardContent>
          </Card>
        )
      )}
    </>
  );
}
