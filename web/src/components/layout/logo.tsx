import Image from "next/image";

import { cn } from "@/lib/utils";

/**
 * Logo Ibatiks.
 *
 * Tersedia dua ragam karena satu berkas tidak cukup untuk dua ukuran: `full`
 * memuat tulisan "Ibatiks co." yang baru terbaca pada ukuran besar, sedangkan
 * `mark` hanya guratan "iba." yang tetap dikenali saat mengecil jadi ikon di
 * sidebar yang diciutkan.
 */
export function Logo({
  variant = "mark",
  size = 32,
  className,
  priority = false,
}: {
  variant?: "mark" | "full";
  /** Sisi terpanjang gambar dalam piksel. */
  size?: number;
  className?: string;
  priority?: boolean;
}) {
  const full = variant === "full";
  const src = full ? "/logo-ibatiks.png" : "/logo-ibatiks-mark.png";

  // Rasio asli berkas dipertahankan supaya guratannya tidak gepeng; tinggi
  // dihitung dari lebar, bukan dipatok bersama-sama.
  const ratio = full ? 584 / 512 : 271 / 256;

  return (
    <Image
      src={src}
      alt="Ibatiks"
      width={size}
      height={Math.round(size * ratio)}
      priority={priority}
      className={cn("object-contain", className)}
    />
  );
}
