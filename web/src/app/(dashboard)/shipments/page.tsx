"use client";

import { PageHeader } from "@/components/ui/page";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { AntreanKemas } from "./antrean-kemas";
import { DaftarPaket } from "./daftar-paket";

/**
 * Mengemas dan menyerahkan ke kurir adalah satu pekerjaan yang sama, dikerjakan
 * orang yang sama di meja yang sama. Dulu keduanya jadi dua menu terpisah,
 * sehingga petugas gudang harus bolak-balik: melihat apa yang harus dikemas di
 * satu menu, lalu mencari paketnya di menu lain untuk mengisi nomor resi.
 *
 * Antrean dijadikan tab pertama karena itu yang dibuka lebih dulu tiap pagi —
 * daftar paket adalah hasilnya, bukan titik mulainya.
 */
export default function ShipmentsPage() {
  return (
    <>
      <PageHeader
        title="Pengiriman"
        description="Antrean order yang menunggu dikemas, dan paket yang sudah punya nomor resi"
      />

      <Tabs defaultValue="antrean">
        <TabsList>
          <TabsTrigger value="antrean">Antrean Kemas</TabsTrigger>
          <TabsTrigger value="paket">Paket &amp; Resi</TabsTrigger>
        </TabsList>

        <TabsContent value="antrean">
          <AntreanKemas />
        </TabsContent>

        <TabsContent value="paket">
          <DaftarPaket />
        </TabsContent>
      </Tabs>
    </>
  );
}
