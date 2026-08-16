"use client";

import { Check, Copy, Loader2, Mail, MessageCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

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
import type { NotifyMessage } from "@/types/api";

interface WAMessageDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  message: NotifyMessage | undefined;
  isLoading?: boolean;
  error?: unknown;
  /** Dipanggil setelah admin menekan tombol kirim, untuk mencatat status terkirim. */
  onSent?: (channel: "wa" | "email" | "manual") => void;
  sending?: boolean;
}

/**
 * Menampilkan pesan siap kirim beserta tombol yang membuka WhatsApp.
 *
 * Pengiriman tetap dilakukan admin sendiri dari nomor tokonya, jadi customer
 * menerima pesan dari kontak yang sudah dikenal — bukan dari nomor gateway
 * asing yang gampang disangka spam.
 */
export function WAMessageDialog({
  open,
  onOpenChange,
  title,
  description,
  message,
  isLoading = false,
  error,
  onSent,
  sending = false,
}: WAMessageDialogProps) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    if (!message) return;
    try {
      await navigator.clipboard.writeText(message.text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Browser menolak akses clipboard. Salin manual dari kotak di atas.");
    }
  }

  function handleOpenWhatsApp() {
    if (!message?.whatsapp_url) return;
    window.open(message.whatsapp_url, "_blank", "noopener,noreferrer");
    onSent?.("wa");
  }

  function handleOpenEmail() {
    if (!message?.mailto_url) return;
    window.location.href = message.mailto_url;
    onSent?.("email");
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>

        <ErrorState error={error} />

        {isLoading ? (
          <div className="flex items-center gap-2 py-8 text-sm text-muted-foreground">
            <Loader2 className="size-4 animate-spin" />
            Menyiapkan pesan…
          </div>
        ) : message ? (
          <>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">
                Dikirim ke <span className="font-medium text-foreground">{message.phone}</span>
              </p>
              <pre className="scrollbar-thin max-h-64 overflow-auto whitespace-pre-wrap rounded-lg border border-border bg-muted/50 p-3 font-sans text-sm">
                {message.text}
              </pre>
            </div>

            <DialogFooter className="sm:justify-between">
              <Button variant="outline" onClick={handleCopy}>
                {copied ? <Check /> : <Copy />}
                {copied ? "Tersalin" : "Salin teks"}
              </Button>

              <div className="flex flex-col-reverse gap-2 sm:flex-row">
                {message.mailto_url && (
                  <Button variant="outline" onClick={handleOpenEmail} loading={sending}>
                    <Mail />
                    Email
                  </Button>
                )}
                <Button variant="success" onClick={handleOpenWhatsApp} loading={sending}>
                  <MessageCircle />
                  Buka WhatsApp
                </Button>
              </div>
            </DialogFooter>
          </>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
