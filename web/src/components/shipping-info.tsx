"use client";

import { Info } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

/**
 * Penjelasan cara ongkir dihitung.
 *
 * Angka ongkir sering dipertanyakan customer maupun admin baru — terutama saat
 * paket ringan tapi besar ditagih lebih mahal dari beratnya. Penjelasannya
 * ditaruh persis di sebelah tombol hitung supaya jawabannya ada di tempat
 * pertanyaannya muncul, bukan di dokumen terpisah yang tidak pernah dibuka.
 */
export function ShippingInfoButton({ className }: { className?: string }) {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className={className}
            onClick={() => setOpen(true)}
            aria-label="Cara ongkir dihitung"
          >
            <Info />
          </Button>
        </TooltipTrigger>
        <TooltipContent>Cara ongkir dihitung</TooltipContent>
      </Tooltip>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-h-[calc(100dvh-2rem)] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Cara ongkir dihitung</DialogTitle>
            <DialogDescription>
              Kurir tidak selalu menagih berat timbangan. Paket besar tapi ringan tetap memakan
              ruang di mobil, jadi yang ditagih adalah yang lebih besar antara berat asli dan berat
              volume.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-5 text-sm">
            <section className="space-y-2">
              <h3 className="font-semibold">1. Berat volume</h3>
              <p className="text-muted-foreground">
                Dihitung dari dimensi paket, bukan dari timbangan:
              </p>
              <pre className="scrollbar-thin overflow-x-auto rounded-lg bg-muted px-3 py-2 text-xs">
                berat volume (kg) = panjang × lebar × tinggi (cm) ÷ 6.000
              </pre>
              <p className="text-muted-foreground">
                Angka 6.000 adalah pembagi bawaan JNE, dan hampir semua kurir dalam negeri
                mengikutinya.
              </p>
            </section>

            <section className="space-y-2">
              <h3 className="font-semibold">2. Berat yang ditagih</h3>
              <p className="text-muted-foreground">
                Diambil yang paling besar, lalu dibulatkan <em>ke atas</em> ke kilogram penuh, dan
                tidak pernah kurang dari berat minimum tarif kota tujuan:
              </p>
              <pre className="scrollbar-thin overflow-x-auto rounded-lg bg-muted px-3 py-2 text-xs">
                berat ditagih = pembulatan_ke_atas( maks(berat asli, berat volume) )
              </pre>
              <p className="text-muted-foreground">
                Pembulatan ke atas berarti 1,2 kg ditagih 2 kg. Itu memang cara kurir menagih, jadi
                perkiraan di sini mengikutinya supaya tidak ada selisih yang harus ditombok toko.
              </p>
            </section>

            <section className="space-y-2">
              <h3 className="font-semibold">3. Biayanya</h3>
              <p className="text-muted-foreground">
                Angkanya datang dari <span className="font-medium text-foreground">kurir</span>,
                lewat RajaOngkir. Setelah berat dan dimensi diisi di dialog kemas, daftar layanan
                beserta harganya keluar dan kamu memilih salah satu — kurir menjual ongkos utuh
                untuk berat itu, bukan harga per kilogram, jadi angkanya sama dengan yang nanti
                dibayar di konter.
              </p>
              <p className="text-muted-foreground">
                Toko tidak lagi menyimpan tabel tarif sendiri. Tarif yang diketik tangan tidak
                pernah ikut naik saat kurir menaikkan harganya, dan angka yang salah tapi terlihat
                resmi lebih berbahaya daripada tidak ada angka sama sekali.
              </p>
              <p className="text-muted-foreground">
                Kalau kurirnya sedang tidak bisa dihubungi, ongkirnya{" "}
                <span className="font-medium text-foreground">diketik sendiri</span> di dialog yang
                sama — dari struk konter atau aplikasi kurir. Pengemasan tidak perlu berhenti
                menunggu layanannya pulih.
              </p>
            </section>

            <section className="space-y-2">
              <h3 className="font-semibold">Contoh</h3>
              <div className="rounded-lg border border-border">
                <table className="w-full text-sm">
                  <tbody className="[&_td]:px-3 [&_td]:py-2 [&_tr]:border-b [&_tr:last-child]:border-0">
                    <tr>
                      <td className="text-muted-foreground">Paket</td>
                      <td className="text-right">40 × 30 × 20 cm, timbangan 1.500 g</td>
                    </tr>
                    <tr>
                      <td className="text-muted-foreground">Berat volume</td>
                      <td className="text-right">40×30×20 ÷ 6.000 = 4 kg</td>
                    </tr>
                    <tr>
                      <td className="text-muted-foreground">Berat ditagih</td>
                      <td className="text-right">maks(1,5 kg; 4 kg) = 4 kg</td>
                    </tr>
                    <tr>
                      <td className="text-muted-foreground">Ditanyakan ke kurir</td>
                      <td className="text-right">JNE REG, 4 kg ke Jakarta Selatan</td>
                    </tr>
                    <tr className="font-medium">
                      <td>Ongkir</td>
                      <td className="text-right">seharga yang dijawab kurir untuk 4 kg</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p className="text-muted-foreground">
                Karena itu mengisi dimensi paket bukan pekerjaan tambahan yang sia-sia: tanpa
                dimensi, paket besar-ringan akan diperkirakan jauh lebih murah daripada tagihan
                kurir yang sebenarnya.
              </p>
            </section>

            <section className="space-y-2">
              <h3 className="font-semibold">Yang perlu diingat</h3>
              <ul className="list-disc space-y-1 pl-5 text-muted-foreground">
                <li>
                  Angka ini <span className="font-medium">perkiraan</span>. Setelah paket
                  ditimbang di konter kurir, isi ongkir sebenarnya pada form kirim supaya laporan
                  laba memakai biaya yang nyata.
                </li>
                <li>
                  Ongkir yang ditagihkan ke customer dan ongkir yang dibayar ke kurir disimpan
                  terpisah — toko bebas menanggung selisihnya.
                </li>
                <li>
                  Berat produk diambil dari master produk. Produk yang beratnya masih 0 membuat
                  perkiraan terlalu rendah.
                </li>
              </ul>
            </section>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
