/**
 * Kredit pembuat aplikasi.
 *
 * Sengaja kecil dan tenang: ini bukan informasi yang dicari tim toko saat
 * bekerja, jadi tidak boleh bersaing dengan isi halaman. Tetap memakai ukuran
 * teks kecil aplikasi (12px) dan warna teks sekunder yang sama dengan seluruh
 * keterangan lain — bukan warna yang lebih pudar lagi, supaya tetap terbaca.
 */
export function FooterCredit({ className }: { className?: string }) {
  return (
    <p className={className}>
      <span className="text-xs text-muted-foreground">
        Developed by <span className="font-medium text-foreground/70">Loomware Studio</span>
      </span>
    </p>
  );
}
