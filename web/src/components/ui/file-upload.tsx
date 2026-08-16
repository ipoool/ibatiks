"use client";

import { FileText, Paperclip, Trash2 } from "lucide-react";
import { useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { api, ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

/**
 * Pengunggah berkas untuk bukti transfer dan struk belanja.
 *
 * Hasil unggahan disimpan sebagai URL pada kolom bukti, jadi berkas lama yang
 * dulu diisi dengan menempel tautan tetap terbaca. Berkas gambar ditampilkan
 * sebagai pratinjau kecil supaya admin bisa memastikan yang terkirim memang
 * bukti yang benar sebelum pembayaran dicatat.
 */
export function FileUpload({
  value,
  onChange,
  accept = "image/*,application/pdf",
  disabled,
  className,
}: {
  value: string | null;
  onChange: (url: string | null) => void;
  accept?: string;
  disabled?: boolean;
  className?: string;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleFile(file: File | undefined) {
    if (!file) return;

    setError(null);
    setUploading(true);
    try {
      const result = await api.upload(file);
      onChange(result.url);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal mengunggah berkas");
    } finally {
      setUploading(false);
      // Nilai input dikosongkan supaya memilih berkas yang sama dua kali tetap
      // memicu onChange — kalau tidak, percobaan ulang setelah gagal diam saja.
      if (inputRef.current) inputRef.current.value = "";
    }
  }

  const isImage = value ? /\.(jpe?g|png|webp|gif)$/i.test(value) : false;

  return (
    <div className={cn("space-y-2", className)}>
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        className="hidden"
        onChange={(event) => handleFile(event.target.files?.[0])}
      />

      {value ? (
        <div className="flex items-center gap-3 rounded-lg border border-border bg-muted/30 p-2">
          {isImage ? (
            // Berkas diunggah ke server sendiri, jadi tidak ada host luar yang
            // perlu dioptimasi Next; <img> biasa sudah cukup dan menghindari
            // konfigurasi domain gambar.
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={value}
              alt="Bukti transfer"
              className="size-14 shrink-0 rounded-md border border-border object-cover"
            />
          ) : (
            <div className="flex size-14 shrink-0 items-center justify-center rounded-md border border-border bg-card">
              <FileText className="size-5 text-muted-foreground" />
            </div>
          )}

          <a
            href={value}
            target="_blank"
            rel="noopener noreferrer"
            className="min-w-0 flex-1 truncate text-sm text-primary hover:underline"
          >
            Lihat bukti
          </a>

          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            onClick={() => onChange(null)}
            disabled={disabled}
            aria-label="Hapus bukti"
          >
            <Trash2 className="text-destructive" />
          </Button>
        </div>
      ) : (
        <Button
          type="button"
          variant="outline"
          className="w-full justify-center"
          loading={uploading}
          disabled={disabled}
          onClick={() => inputRef.current?.click()}
        >
          <Paperclip />
          {uploading ? "Mengunggah…" : "Unggah bukti"}
        </Button>
      )}

      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}
