"use client";

import * as React from "react";

import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface CheckboxFieldProps {
  id: string;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  children: React.ReactNode;
  /**
   * `boxed` memberi bingkai supaya sebuah centang berdiri sejajar dengan kolom
   * saring di sebelahnya; `plain` untuk centang yang berada di dalam form.
   */
  variant?: "plain" | "boxed";
  className?: string;
  disabled?: boolean;
}

/**
 * Centang beserta labelnya, dengan seluruh label bisa diklik.
 *
 * Checkbox shadcn adalah tombol Radix, bukan input bawaan, sehingga tidak ikut
 * terpicu saat teks di sebelahnya diklik — perilaku yang orang harapkan dari
 * `<label>`. Komponen ini menyambungkan keduanya lewat `htmlFor` supaya sasaran
 * kliknya tetap seluas teksnya.
 */
export function CheckboxField({
  id,
  checked,
  onCheckedChange,
  children,
  variant = "plain",
  className,
  disabled,
}: CheckboxFieldProps) {
  return (
    <div
      className={cn(
        "flex items-center gap-2 text-sm",
        variant === "boxed" && "h-9 rounded-md border border-input bg-card px-3",
        className,
      )}
    >
      <Checkbox
        id={id}
        checked={checked}
        onCheckedChange={(value) => onCheckedChange(value === true)}
        disabled={disabled}
      />
      <Label htmlFor={id} className="cursor-pointer font-normal">
        {children}
      </Label>
    </div>
  );
}
