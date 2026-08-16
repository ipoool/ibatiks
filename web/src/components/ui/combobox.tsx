"use client";

import { Check, ChevronsUpDown } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

export interface ComboboxOption {
  value: string;
  label: string;
  /**
   * Teks tambahan yang ikut dicari tapi tidak ditampilkan sebagai judul —
   * misalnya nomor WhatsApp customer atau SKU produk. Admin lebih sering ingat
   * nomor pemesan ketimbang ejaan namanya.
   */
  keywords?: string;
  /** Baris kedua di bawah label, untuk keterangan seperti nomor atau harga. */
  description?: string;
}

interface ComboboxProps {
  value: string;
  onChange: (value: string) => void;
  options: ReadonlyArray<ComboboxOption>;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyLabel?: string;
  /** Ditampilkan menggantikan daftar selagi pilihannya masih diambil. */
  isLoading?: boolean;
  disabled?: boolean;
  id?: string;
  className?: string;
}

/**
 * Pemilih dengan kolom pencarian.
 *
 * Dipakai menggantikan Select biasa pada daftar yang bisa memanjang — customer
 * dan katalog produk. Select biasa memaksa orang menggulir mencari satu nama di
 * antara ratusan; di sini cukup mengetik sebagian nama atau nomornya.
 */
export function Combobox({
  value,
  onChange,
  options,
  placeholder = "Pilih…",
  searchPlaceholder = "Cari…",
  emptyLabel = "Tidak ada yang cocok",
  isLoading = false,
  disabled,
  id,
  className,
}: ComboboxProps) {
  const [open, setOpen] = React.useState(false);
  const selected = options.find((option) => option.value === value);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          id={id}
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            "w-full justify-between bg-card font-normal",
            !selected && "text-muted-foreground",
            className,
          )}
        >
          <span className="truncate">{selected ? selected.label : placeholder}</span>
          <ChevronsUpDown className="opacity-50" />
        </Button>
      </PopoverTrigger>

      <PopoverContent
        className="w-(--radix-popover-trigger-width) p-0"
        align="start"
        // Pencarian dilakukan atas seluruh daftar yang sudah ada di memori,
        // jadi tidak ada permintaan jaringan tiap ketikan.
      >
        <Command>
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList>
            {isLoading ? (
              <div className="px-3 py-6 text-center text-sm text-muted-foreground">Memuat…</div>
            ) : (
              <>
                <CommandEmpty>{emptyLabel}</CommandEmpty>
                <CommandGroup>
                  {options.map((option) => (
                    <CommandItem
                      key={option.value}
                      // `value` yang dipakai pencarian sengaja digabung dengan
                      // kata kunci: cmdk mencocokkan string ini, bukan `value`
                      // asli yang berupa UUID dan tidak pernah diketik orang.
                      value={`${option.label} ${option.keywords ?? ""}`}
                      onSelect={() => {
                        onChange(option.value);
                        setOpen(false);
                      }}
                    >
                      <Check
                        className={cn(
                          "shrink-0",
                          option.value === value ? "opacity-100" : "opacity-0",
                        )}
                      />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate">{option.label}</span>
                        {option.description && (
                          <span className="block truncate text-xs text-muted-foreground">
                            {option.description}
                          </span>
                        )}
                      </span>
                    </CommandItem>
                  ))}
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
