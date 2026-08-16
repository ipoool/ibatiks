"use client";

import * as React from "react";

import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface FieldProps {
  label: string;
  htmlFor?: string;
  /** Pesan error dari validasi form atau dari backend. */
  error?: string;
  hint?: string;
  required?: boolean;
  className?: string;
  children: React.ReactNode;
}

/**
 * Pembungkus satu isian form: label, kontrol, petunjuk, dan pesan error.
 *
 * Petunjuk disembunyikan begitu ada error, bukan ditumpuk di bawahnya. Dua baris
 * teks kecil di bawah satu kolom membuat pesan yang penting — yang menjelaskan
 * kenapa formnya ditolak — jadi sulit dibedakan dari keterangan biasa.
 */
export function Field({ label, htmlFor, error, hint, required, className, children }: FieldProps) {
  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      <Label htmlFor={htmlFor}>
        {label}
        {required && <span className="ml-0.5 text-destructive">*</span>}
      </Label>
      {children}
      {hint && !error && <p className="text-xs text-muted-foreground">{hint}</p>}
      {error && <p className="text-xs font-medium text-destructive">{error}</p>}
    </div>
  );
}
