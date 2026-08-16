"use client";

import { Inbox, Loader2 } from "lucide-react";
import * as React from "react";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";

/**
 * Tabel data dengan penanganan kondisi memuat dan kosong yang seragam.
 *
 * Dibangun di atas primitif Table shadcn dan diletakkan di luar `components/ui`
 * karena isinya keputusan aplikasi, bukan primitif: berapa lama pesan memuat
 * ditampilkan, seperti apa keadaan kosong terlihat, dan tabel lebar digulir
 * sendiri alih-alih menggeser seluruh halaman.
 */
interface DataTableProps {
  /** Jumlah kolom, dipakai merentangkan baris kosong dan baris memuat. */
  columns: number;
  isLoading?: boolean;
  isEmpty?: boolean;
  emptyTitle?: string;
  emptyDescription?: string;
  emptyAction?: React.ReactNode;
  head: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}

export function DataTable({
  columns,
  isLoading = false,
  isEmpty = false,
  emptyTitle = "Belum ada data",
  emptyDescription,
  emptyAction,
  head,
  children,
  className,
}: DataTableProps) {
  return (
    <div
      className={cn(
        "scrollbar-thin overflow-x-auto rounded-xl border border-border bg-card",
        className,
      )}
    >
      <Table>
        <TableHeader className="bg-muted/60">{head}</TableHeader>
        <TableBody>
          {isLoading ? (
            <TableRow className="hover:bg-transparent">
              <TableCell colSpan={columns} className="py-16 text-center">
                <span className="inline-flex items-center gap-2 text-sm text-muted-foreground">
                  <Loader2 className="size-4 animate-spin" />
                  Memuat data…
                </span>
              </TableCell>
            </TableRow>
          ) : isEmpty ? (
            <TableRow className="hover:bg-transparent">
              <TableCell colSpan={columns} className="py-16 text-center">
                <div className="flex flex-col items-center gap-2">
                  <Inbox className="size-8 text-muted-foreground/60" />
                  <p className="text-sm font-medium">{emptyTitle}</p>
                  {emptyDescription && (
                    <p className="max-w-sm text-sm text-muted-foreground">{emptyDescription}</p>
                  )}
                  {emptyAction && <div className="mt-2">{emptyAction}</div>}
                </div>
              </TableCell>
            </TableRow>
          ) : (
            children
          )}
        </TableBody>
      </Table>
    </div>
  );
}

/*
 * Alias pendek untuk sel tabel.
 *
 * Tabel di aplikasi ini bisa punya delapan kolom dengan kelas lebar di tiap
 * header; menuliskan `TableCell` dan `TableHead` penuh membuat satu baris JSX
 * membentang jauh melewati batas layar dan justru menyulitkan pembacaan
 * strukturnya.
 */
export const TR = TableRow;
export const TD = TableCell;

export function TH({ className, ...props }: React.ComponentProps<typeof TableHead>) {
  return (
    <TableHead
      className={cn(
        "px-4 py-3 text-left text-xs font-semibold tracking-wide text-muted-foreground uppercase",
        className,
      )}
      {...props}
    />
  );
}
