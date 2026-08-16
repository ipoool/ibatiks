"use client";

import * as React from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ErrorState } from "@/components/ui/page";
import { cn } from "@/lib/utils";

interface FormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  /** Error dari mutasi; ditampilkan di dalam dialog agar tidak hilang saat dialog terbuka. */
  error?: unknown;
  submitLabel?: string;
  cancelLabel?: string;
  loading?: boolean;
  /**
   * Menonaktifkan tombol simpan selagi isian wajib belum lengkap. Lebih baik
   * tombolnya mati daripada ditekan lalu dijawab penolakan server: yang kedua
   * membuat orang mengira ada yang rusak, bukan ada yang kurang diisi.
   */
  submitDisabled?: boolean;
  onSubmit: (event: React.FormEvent) => void;
  className?: string;
  children: React.ReactNode;
  /** Aksi tambahan di kiri footer, misalnya tombol hapus. */
  secondaryAction?: React.ReactNode;
}

/** Dialog berisi form dengan penanganan error dan tombol simpan yang seragam. */
export function FormDialog({
  open,
  onOpenChange,
  title,
  description,
  error,
  submitLabel = "Simpan",
  cancelLabel = "Batal",
  loading = false,
  submitDisabled = false,
  onSubmit,
  className,
  children,
  secondaryAction,
}: FormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {/*
        Dialog dibuat kolom penuh: judul dan tombol aksinya tetap di tempat,
        hanya isian yang menggulir. Pada form panjang seperti Tambah Produk,
        footer yang ikut menggulir membuat tombol Simpan menghilang dari layar
        dan orang mengira formnya buntu.
      */}
      <DialogContent
        className={cn(
          "flex max-h-[calc(100dvh-2rem)] flex-col overflow-hidden sm:max-w-xl",
          className,
        )}
      >
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>

        <form onSubmit={onSubmit} className="flex min-h-0 flex-1 flex-col gap-4">
          <div className="scrollbar-thin min-h-0 flex-1 space-y-4 overflow-y-auto px-1">
            <ErrorState error={error} />
            {children}
          </div>

          <DialogFooter className="border-t border-border pt-4 sm:justify-between">
            <div>{secondaryAction}</div>
            <div className="flex flex-col-reverse gap-2 sm:flex-row">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                {cancelLabel}
              </Button>
              <Button type="submit" loading={loading} disabled={submitDisabled}>
                {submitLabel}
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmLabel?: string;
  destructive?: boolean;
  loading?: boolean;
  error?: unknown;
  onConfirm: () => void;
}

/** Konfirmasi untuk aksi yang sulit dibatalkan, seperti menghapus data. */
export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = "Ya, lanjutkan",
  destructive = true,
  loading = false,
  error,
  onConfirm,
}: ConfirmDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        <ErrorState error={error} />

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Batal
          </Button>
          <Button
            variant={destructive ? "destructive" : "default"}
            onClick={onConfirm}
            loading={loading}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
