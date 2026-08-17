import * as React from "react"

import { cn } from "@/lib/utils"
import { bersihkanPesanValidasi, pasangPesanValidasi } from "@/lib/validasi-bawaan"

function Textarea({ className, onInvalid, onInput, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      // Alasannya sama seperti pada Input: gelembung bawaan browser berbahasa
      // Inggris, dan `lang` halaman tidak mengubahnya.
      onInvalid={(event) => {
        pasangPesanValidasi(event.currentTarget)
        onInvalid?.(event)
      }}
      onInput={(event) => {
        bersihkanPesanValidasi(event.currentTarget)
        onInput?.(event)
      }}
      className={cn(
        "flex field-sizing-content min-h-16 w-full rounded-md border border-input bg-card px-3 py-2 text-base shadow-xs transition-[color,box-shadow] outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 md:text-sm dark:bg-input/30 dark:aria-invalid:ring-destructive/40",
        className
      )}
      {...props}
    />
  )
}

export { Textarea }
