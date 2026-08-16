"use client";

import { History } from "lucide-react";

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ErrorState } from "@/components/ui/page";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useProductPriceHistory } from "@/hooks/use-master";
import { formatDate, formatIDR, formatNumber, toNumber } from "@/lib/utils";

/**
 * Riwayat harga sebuah produk dari trip ke trip.
 *
 * Ditampilkan berdampingan: harga yang dipasang di katalog dan harga yang
 * benar-benar dibayar tripper. Selisih keduanya yang paling berguna saat
 * menyusun harga trip berikutnya — kalau harga di kasir ternyata selalu lebih
 * tinggi dari yang dipasang, markup-nya yang perlu disesuaikan, bukan
 * harganya ditebak ulang dari nol.
 */
export function PriceHistoryDialog({
  productId,
  productName,
  open,
  onOpenChange,
}: {
  productId: string | undefined;
  productName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { data, isLoading, error } = useProductPriceHistory(productId, open);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>Riwayat harga — {productName}</DialogTitle>
          <DialogDescription>
            Harga modal berbeda antar trip karena negara asal dan kursnya berbeda. Angka lama di
            sini adalah acuan, bukan harga yang otomatis dipakai lagi.
          </DialogDescription>
        </DialogHeader>

        <ErrorState error={error} />

        <DataTable
          columns={6}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.length ?? 0) === 0}
          emptyTitle="Belum ada riwayat"
          emptyDescription="Produk ini belum pernah masuk katalog trip mana pun."
          head={
            <TR>
              <TH className="min-w-40">Trip</TH>
              <TH className="w-28 text-right">Katalog</TH>
              <TH className="w-28 text-right">Beli riil</TH>
              <TH className="w-28 text-right">Harga jual</TH>
              <TH className="w-20 text-right">Dibeli</TH>
              <TH className="w-20 text-right">Terjual</TH>
            </TR>
          }
        >
          {data?.map((row) => {
            const actual = toNumber(row.actual_cost);
            const catalog = toNumber(row.catalog_cost);
            // Selisih ditandai hanya kalau keduanya benar-benar ada, supaya
            // produk yang masuk katalog tapi belum dibeli tidak terlihat rugi.
            const overpaid = actual > 0 && catalog > 0 && actual > catalog;

            return (
              <TR key={row.trip_id}>
                <TD>
                  <p className="font-medium">{row.trip_code}</p>
                  <p className="text-xs text-muted-foreground">
                    {row.country} · {formatDate(row.depart_date)} · 1 {row.currency} = Rp
                    {formatNumber(row.exchange_rate)}
                  </p>
                </TD>
                <TD className="tabular text-right">
                  {catalog > 0 ? `${row.currency} ${formatNumber(catalog)}` : "—"}
                </TD>
                <TD
                  className={`tabular text-right ${overpaid ? "font-medium text-amber-600" : ""}`}
                >
                  {actual > 0 ? `${row.currency} ${formatNumber(actual)}` : "—"}
                </TD>
                <TD className="tabular text-right">
                  {toNumber(row.sell_price) > 0 ? formatIDR(row.sell_price) : "—"}
                </TD>
                <TD className="tabular text-right text-muted-foreground">
                  {formatNumber(row.qty_purchased)}
                </TD>
                <TD className="tabular text-right text-muted-foreground">
                  {formatNumber(row.qty_sold)}
                </TD>
              </TR>
            );
          })}
        </DataTable>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Ringkasan satu baris berisi harga terakhir produk ini, untuk ditempel di form
 * katalog trip. Inilah yang membuat harga tidak perlu digali ulang dari catatan
 * trip sebelumnya.
 */
export function LastPriceHint({
  productId,
  currency,
  onUse,
}: {
  productId: string | undefined;
  currency: string;
  onUse: (costPrice: string) => void;
}) {
  const { data } = useProductPriceHistory(productId, Boolean(productId));
  const last = data?.[0];
  if (!last) return null;

  // Yang ditawarkan adalah harga yang benar-benar dibayar di kasir kalau ada,
  // karena itu angka yang terbukti; harga katalog lama hanya perkiraan.
  const actual = toNumber(last.actual_cost);
  const suggested = actual > 0 ? last.actual_cost : last.catalog_cost;
  if (toNumber(suggested) <= 0) return null;

  const sameCurrency = last.currency === currency;

  return (
    <div className="flex flex-wrap items-center gap-x-2 gap-y-1 rounded-lg border border-border bg-muted/40 px-3 py-2 text-xs">
      <History className="size-3.5 shrink-0 text-muted-foreground" />
      <span className="text-muted-foreground">
        Trip {last.trip_code} ({last.country}):{" "}
        <span className="font-medium text-foreground">
          {last.currency} {formatNumber(suggested)}
        </span>
        {actual > 0 ? " dibayar di kasir" : " di katalog"}
      </span>

      {sameCurrency ? (
        <button
          type="button"
          onClick={() => onUse(String(suggested))}
          className="font-medium text-primary hover:underline"
        >
          Pakai harga ini
        </button>
      ) : (
        <span className="text-muted-foreground">
          — mata uangnya berbeda dari trip ini, jadi tidak bisa dipakai langsung
        </span>
      )}
    </div>
  );
}
