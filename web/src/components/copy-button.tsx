"use client";

import { Check, Copy } from "lucide-react";
import * as React from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

/**
 * Menyalin teks ke papan klip.
 *
 * `navigator.clipboard` hanya ada di konteks aman — https atau localhost. Kalau
 * aplikasinya dibuka lewat alamat IP di jaringan toko, ia tidak ada sama sekali,
 * dan tombol yang diam saja lebih membingungkan daripada tombol yang tidak ada.
 * Karena itu ada cara cadangan lewat textarea tersembunyi.
 */
async function salin(teks: string): Promise<boolean> {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(teks);
      return true;
    }
  } catch {
    // Jatuh ke cara di bawah.
  }

  try {
    const kotak = document.createElement("textarea");
    kotak.value = teks;
    // Di luar layar, bukan display:none — kotak yang tidak dirender tidak bisa
    // diseleksi, dan tanpa seleksi tidak ada yang tersalin.
    kotak.style.position = "fixed";
    kotak.style.left = "-9999px";
    document.body.appendChild(kotak);
    kotak.select();
    const berhasil = document.execCommand("copy");
    document.body.removeChild(kotak);
    return berhasil;
  } catch {
    return false;
  }
}

interface CopyButtonProps {
  /** Teks yang disalin. */
  value: string;
  /** Disebut di tooltip dan pesan berhasil, misalnya "Nomor order". */
  label?: string;
  className?: string;
}

export function CopyButton({ value, label = "Teks", className }: CopyButtonProps) {
  const [tersalin, setTersalin] = React.useState(false);
  const jeda = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  // Hanya membersihkan timer saat komponennya dilepas — tidak menyetel state di
  // dalam efek, yang memang dilarang aturan lint proyek ini.
  React.useEffect(() => () => {
    if (jeda.current) clearTimeout(jeda.current);
  }, []);

  async function handleClick(event: React.MouseEvent) {
    // Nomor order biasanya berupa tautan ke detailnya; tanpa ini, menyalin ikut
    // memindahkan halaman dan salinannya tidak sempat terlihat.
    event.preventDefault();
    event.stopPropagation();

    if (!(await salin(value))) {
      toast.error(`${label} gagal disalin`);
      return;
    }

    setTersalin(true);
    if (jeda.current) clearTimeout(jeda.current);
    jeda.current = setTimeout(() => setTersalin(false), 1500);
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          onClick={handleClick}
          className={cn("text-muted-foreground hover:text-foreground", className)}
        >
          {tersalin ? <Check className="text-emerald-600" /> : <Copy />}
          <span className="sr-only">Salin {label.toLowerCase()}</span>
        </Button>
      </TooltipTrigger>
      <TooltipContent>{tersalin ? "Tersalin" : `Salin ${label.toLowerCase()}`}</TooltipContent>
    </Tooltip>
  );
}
