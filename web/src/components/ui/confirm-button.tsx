"use client";

import * as React from "react";

import { Button, type ButtonProps } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/form-dialog";

interface ConfirmButtonProps extends Omit<ButtonProps, "onClick"> {
  title: string;
  description: string;
  confirmLabel?: string;
  /** Merah untuk aksi yang merugikan kalau salah; bawaannya biru. */
  destructive?: boolean;
  error?: unknown;
  /**
   * Aksi yang dijalankan setelah pengguna menekan tombol konfirmasi. Harus
   * mengembalikan promise (mis. `mutateAsync`) supaya dialog tahu kapan
   * prosesnya selesai.
   */
  onConfirm: () => Promise<unknown>;
}

/**
 * Tombol yang menahan aksinya di balik dialog konfirmasi.
 *
 * Dipakai untuk perpindahan status, yang tidak punya tombol "urungkan": sekali
 * order berpindah ke Dikirim, isinya terkunci. Karena itu tiap dialog memuat
 * penjelasan dampaknya, bukan sekadar pertanyaan "yakin?" — konfirmasi yang
 * tidak memberi informasi baru hanya melatih orang menekan Ya tanpa membaca.
 */
export function ConfirmButton({
  title,
  description,
  confirmLabel,
  destructive = false,
  error,
  onConfirm,
  children,
  ...buttonProps
}: ConfirmButtonProps) {
  const [open, setOpen] = React.useState(false);
  const [running, setRunning] = React.useState(false);
  // Galat dari percobaan sebelumnya tidak boleh muncul saat dialog baru dibuka,
  // karena error mutation bertahan sampai percobaan berikutnya.
  const [attempted, setAttempted] = React.useState(false);

  async function handleConfirm() {
    setAttempted(true);
    setRunning(true);
    try {
      await onConfirm();
      setOpen(false);
    } catch {
      // Dialog sengaja dibiarkan terbuka agar pengguna bisa mencoba lagi tanpa
      // mencari tombolnya dari awal. Pesan galatnya ditampilkan di dalam dialog.
    } finally {
      setRunning(false);
    }
  }

  return (
    <>
      <Button
        {...buttonProps}
        onClick={() => {
          setAttempted(false);
          setOpen(true);
        }}
      >
        {children}
      </Button>

      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title={title}
        description={description}
        confirmLabel={confirmLabel}
        destructive={destructive}
        loading={running}
        error={attempted ? error : undefined}
        onConfirm={handleConfirm}
      />
    </>
  );
}
