"use client";

import { AlertCircle, Search } from "lucide-react";
import * as React from "react";

import { Input } from "@/components/ui/input";
import { ApiError } from "@/lib/api";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { cn } from "@/lib/utils";

interface PageHeaderProps {
  title: string;
  description?: string;
  /** Tombol aksi utama halaman, ditempatkan di kanan judul. */
  actions?: React.ReactNode;
  className?: string;
}

export function PageHeader({ title, description, actions, className }: PageHeaderProps) {
  return (
    <div className={cn("flex flex-wrap items-start justify-between gap-4", className)}>
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {description && <p className="text-sm text-muted-foreground">{description}</p>}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
    </div>
  );
}

interface SearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export function SearchInput({ value, onChange, placeholder = "Cari…", className }: SearchInputProps) {
  return (
    <div className={cn("relative", className)}>
      <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        className="pl-9"
      />
    </div>
  );
}

/** Menampilkan pesan error API dengan bahasa yang bisa ditindaklanjuti admin. */
export function ErrorState({ error, className }: { error: unknown; className?: string }) {
  if (!error) return null;

  const message =
    error instanceof ApiError
      ? error.message
      : "Terjadi kesalahan yang tidak terduga. Muat ulang halaman lalu coba lagi.";

  return (
    <Alert variant="destructive" className={className}>
      <AlertCircle />
      <AlertTitle>{message}</AlertTitle>
      {error instanceof ApiError && error.fields && (
        <AlertDescription>
          <ul className="list-inside list-disc text-xs">
            {Object.entries(error.fields).map(([field, hint]) => (
              <li key={field}>
                <span className="font-medium">{field}</span>: {hint}
              </li>
            ))}
          </ul>
        </AlertDescription>
      )}
    </Alert>
  );
}

/** Baris ringkas "label: nilai" untuk panel detail. */
export function DetailRow({
  label,
  value,
  className,
}: {
  label: string;
  value: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex items-start justify-between gap-4 py-1.5 text-sm", className)}>
      <span className="text-muted-foreground">{label}</span>
      <span className="text-right font-medium">{value}</span>
    </div>
  );
}
