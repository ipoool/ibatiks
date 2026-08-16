"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import { formatNumber } from "@/lib/utils";
import type { PageMeta } from "@/types/api";

interface PaginationProps {
  meta: PageMeta | undefined;
  onPageChange: (page: number) => void;
}

export function Pagination({ meta, onPageChange }: PaginationProps) {
  if (!meta || meta.total_pages <= 1) return null;

  const from = (meta.page - 1) * meta.per_page + 1;
  const to = Math.min(meta.page * meta.per_page, meta.total);

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 px-1 py-3">
      <p className="text-sm text-muted-foreground">
        Menampilkan <span className="font-medium text-foreground">{formatNumber(from)}</span>–
        <span className="font-medium text-foreground">{formatNumber(to)}</span> dari{" "}
        <span className="font-medium text-foreground">{formatNumber(meta.total)}</span> data
      </p>

      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(meta.page - 1)}
          disabled={meta.page <= 1}
        >
          <ChevronLeft />
          Sebelumnya
        </Button>
        <span className="px-2 text-sm text-muted-foreground">
          Hal. {meta.page} / {meta.total_pages}
        </span>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(meta.page + 1)}
          disabled={meta.page >= meta.total_pages}
        >
          Berikutnya
          <ChevronRight />
        </Button>
      </div>
    </div>
  );
}
