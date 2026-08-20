"use client";

import { Calculator, Check } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { FormDialog } from "@/components/ui/form-dialog";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NumberInput } from "@/components/ui/number-input";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState } from "@/components/ui/page";
import { Checkbox } from "@/components/ui/checkbox";
import { usePackOrder, useShippingOptions } from "@/hooks/use-operations";
import { useOrder } from "@/hooks/use-orders";
import { ApiError } from "@/lib/api";
import { cn, formatIDR, formatNumber, toNumber } from "@/lib/utils";
import type { ShippingOption, ShippingQueueItem } from "@/types/api";

/**
 * Data kemasan sekaligus penetapan ongkir.
 *
 * Dua hal ini satu dialog karena satu pekerjaan: paket ditimbang, lalu
 * harganya keluar. Memisahkannya berarti admin menimbang di satu tempat dan
 * mencari harganya di tempat lain, padahal harga itu ditentukan oleh berat yang
 * baru saja ia masukkan.
 */
export function DialogKemas({
  item,
  onClose,
}: {
  item: ShippingQueueItem;
  onClose: () => void;
}) {
  const pack = usePackOrder(item.order_id);
  const options = useShippingOptions(item.order_id);
  const detail = useOrder(item.order_id);

  /*
   * Centang pengecekan barang. Sengaja tidak disimpan ke database: ini ritual
   * di meja kemas — hitung fisiknya, centang, lalu segel kardusnya. Yang punya
   * arti setelah dialog ditutup adalah paketnya sudah dikemas, dan itu sudah
   * tercatat sendiri lewat data kemasan.
   */
  const [dicek, setDicek] = useState<Record<string, boolean>>({});
  const daftarItem = detail.data?.items ?? [];
  const jumlahDicek = daftarItem.filter((baris) => dicek[baris.id]).length;
  const semuaDicek = daftarItem.length > 0 && jumlahDicek === daftarItem.length;
  /*
   * Kalau daftarnya gagal dimuat, penguncian dilepas. Menahan tombol Simpan
   * karena jaringan sedang bermasalah berarti paket yang sudah siap tidak bisa
   * dicatat sama sekali — jauh lebih merugikan daripada satu pengecekan yang
   * terlewat.
   */
  const terkunci = detail.isLoading || (daftarItem.length > 0 && !semuaDicek);

  const [form, setForm] = useState({
    weight_gram: item.weight_gram ?? 0,
    length_cm: item.length_cm ?? 0,
    width_cm: item.width_cm ?? 0,
    height_cm: item.height_cm ?? 0,
    notes: item.shipment_notes ?? "",
  });
  // Ongkir yang diketik sendiri. Dipakai kalau RajaOngkir tidak bisa menjawab —
  // API key belum diisi, kuota habis, atau layanannya sedang down. Angkanya dari
  // struk konter atau aplikasi kurir; itu yang benar-benar dibayar.
  const [ongkirManual, setOngkirManual] = useState(
    toNumber(item.shipping_fee) > 0 ? String(toNumber(item.shipping_fee)) : "",
  );
  /*
   * Ongkir punya dua sumber yang tidak boleh aktif bersamaan: daftar layanan
   * dari kurir, atau angka yang diketik sendiri. Sebelumnya keduanya tampil
   * berdampingan dan admin harus menyimpulkan sendiri mana yang dipakai —
   * ditambah satu kalimat yang menjelaskan bahwa mengetik akan melepas pilihan
   * di atasnya. Sekarang dipilih lewat satu centang, dan yang tidak dipakai
   * tidak ditampilkan sama sekali.
   *
   * Bawaannya daftar kurir. Mengetik sendiri adalah jalan keluar saat kurir
   * tidak bisa dijangkau, bukan cara yang biasa; kalau order ini sudah punya
   * ongkir tanpa nama kurir, berarti dulu memang diketik dan modenya menyala.
   */
  const [ongkirManualAktif, setOngkirManualAktif] = useState(
    toNumber(item.shipping_fee) > 0 && !item.courier,
  );

  /*
   * Premi asuransi kiriman. Diketik sendiri, bukan dihitung sistem: balasan
   * RajaOngkir hanya berisi nama kurir, layanan, ongkos, dan estimasi tiba —
   * tidak ada data asuransi sama sekali. Menanam rumus premi di kode berarti
   * angka yang tidak pernah ikut berubah saat kurir mengubah tarifnya, persis
   * alasan tabel tarif ongkir dulu dilepas.
   */
  const [pakaiAsuransi, setPakaiAsuransi] = useState(
    toNumber(item.insurance_fee) > 0,
  );
  const [premi, setPremi] = useState(
    toNumber(item.insurance_fee) > 0
      ? String(toNumber(item.insurance_fee))
      : "",
  );

  const [terpilih, setTerpilih] = useState<ShippingOption | null>(
    item.courier && toNumber(item.shipping_fee) > 0
      ? {
          courier: item.courier,
          service: item.service ?? "",
          cost: item.shipping_fee,
          etd: "",
          source: "tersimpan",
        }
      : null,
  );

  const daftar = options.data ?? [];
  const [pertama] = daftar;

  // Yang tersimpan: layanan yang dipilih kalau ada, kalau tidak angka ketikan.
  const ongkirTersimpan = terpilih
    ? String(toNumber(terpilih.cost))
    : ongkirManual.trim();
  // Dicentang tapi kosong berarti nol, bukan "tidak diubah".
  const premiTersimpan = pakaiAsuransi ? String(toNumber(premi)) : "0";
  const totalDitagihkan = toNumber(ongkirTersimpan) + toNumber(premiTersimpan);

  function ambilPilihan() {
    if (form.weight_gram <= 0) {
      toast.error("Isi berat paketnya dulu");
      return;
    }
    options.mutate(
      {
        weight_gram: form.weight_gram,
        length_cm: form.length_cm || undefined,
        width_cm: form.width_cm || undefined,
        height_cm: form.height_cm || undefined,
      },
      {
        onSuccess: () => setTerpilih(null),
        onError: (err) => {
          toast.error(
            err instanceof ApiError
              ? `${err.message} — ongkirnya bisa diketik sendiri di bawah`
              : "Gagal mengambil daftar layanan",
          );
        },
      },
    );
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    pack.mutate(
      {
        courier: terpilih?.courier,
        service: terpilih?.service,
        weight_gram: form.weight_gram,
        length_cm: form.length_cm,
        width_cm: form.width_cm,
        height_cm: form.height_cm,
        // Dikosongkan berarti admin baru menyimpan ukuran paketnya; ongkir di
        // order dibiarkan apa adanya.
        shipping_fee: ongkirTersimpan || undefined,
        insurance_fee: premiTersimpan,
        notes: form.notes || null,
      },
      {
        onSuccess: () => {
          toast.success(
            ongkirTersimpan
              ? "Data kemasan dan ongkir tersimpan"
              : "Data kemasan tersimpan",
          );
          onClose();
        },
        onError: (err) => {
          toast.error(
            err instanceof ApiError
              ? err.message
              : "Gagal menyimpan data kemasan",
          );
        },
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={`Data kemasan ${item.order_number}`}
      description={`${item.recipient_name} · ${item.shipping_city}`}
      onSubmit={handleSubmit}
      loading={pack.isPending}
      submitLabel="Simpan"
      submitDisabled={terkunci}
    >
      <ErrorState error={pack.error} />

      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Berat (gram)" htmlFor="kemas_berat" required>
          <NumberInput
            id="kemas_berat"
            value={form.weight_gram}
            onValueChange={(value) => setForm({ ...form, weight_gram: value })}
            min={1}
            step="any"
            blankWhenZero
            required
          />
        </Field>

        <Field
          label="Dimensi kardus (cm)"
          hint="Panjang × lebar × tinggi. Boleh dikosongkan."
        >
          <div className="flex items-center gap-1">
            <NumberInput
              value={form.length_cm}
              onValueChange={(value) => setForm({ ...form, length_cm: value })}
              min={0}
              step="any"
              blankWhenZero
              aria-label="Panjang (cm)"
            />
            <span className="text-muted-foreground">×</span>
            <NumberInput
              value={form.width_cm}
              onValueChange={(value) => setForm({ ...form, width_cm: value })}
              min={0}
              step="any"
              blankWhenZero
              aria-label="Lebar (cm)"
            />
            <span className="text-muted-foreground">×</span>
            <NumberInput
              value={form.height_cm}
              onValueChange={(value) => setForm({ ...form, height_cm: value })}
              min={0}
              step="any"
              blankWhenZero
              aria-label="Tinggi (cm)"
            />
          </div>
        </Field>
      </div>

      {!ongkirManualAktif && (
        <Field
          label="Layanan kurir"
          hint="Ongkir yang ditagihkan ke customer diambil dari sini."
        >
          <div className="space-y-2">
            <Button
              type="button"
              variant="outline"
              onClick={ambilPilihan}
              loading={options.isPending}
              className="w-full sm:w-auto"
            >
              <Calculator />
              Ambil daftar layanan
            </Button>

            {daftar.length > 0 && (
              <ul className="max-h-56 divide-y overflow-y-auto rounded-md border">
                {daftar.map((pilihan) => {
                  const aktif =
                    terpilih?.courier === pilihan.courier &&
                    terpilih?.service === pilihan.service;
                  return (
                    <li key={`${pilihan.courier}-${pilihan.service}`}>
                      <button
                        type="button"
                        onClick={() => setTerpilih(pilihan)}
                        className={cn(
                          "flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-accent",
                          aktif && "bg-accent",
                        )}
                      >
                        <Check
                          className={cn(
                            "size-4 shrink-0",
                            aktif ? "opacity-100" : "opacity-0",
                          )}
                        />
                        <span className="min-w-0 flex-1">
                          <span className="font-medium">
                            {pilihan.courier} {pilihan.service}
                          </span>
                          {pilihan.etd && (
                            <span className="block text-xs text-muted-foreground">
                              estimasi {pilihan.etd}
                            </span>
                          )}
                        </span>
                        <span className="tabular shrink-0 font-medium">
                          {formatIDR(pilihan.cost)}
                        </span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}

            {pertama?.source && (
              <p className="text-xs text-muted-foreground">
                Sumber angka: {pertama.source}
                {pertama.destination && ` · tujuan ${pertama.destination}`}
              </p>
            )}

            {terpilih && (
              <p className="text-sm">
                Dipilih:{" "}
                <span className="font-medium">
                  {terpilih.courier} {terpilih.service} ·{" "}
                  {formatIDR(terpilih.cost)}
                </span>
              </p>
            )}
          </div>
        </Field>
      )}

      {/*
       * Satu-satunya jaring pengaman sejak tabel tarif dilepas. Kurir sesekali
       * tidak bisa dihubungi, dan pengemasan tidak boleh berhenti karenanya —
       * angka dari struk konter lebih benar daripada tebakan mana pun.
       */}
      <div className="space-y-2">
        <label className="flex cursor-pointer items-center gap-2">
          <Checkbox
            checked={ongkirManualAktif}
            onCheckedChange={(nilai) => {
              const aktif = nilai === true;
              setOngkirManualAktif(aktif);
              // Yang ditinggalkan ikut dikosongkan supaya tidak ada angka
              // tersembunyi dari mode sebelumnya yang diam-diam ikut tersimpan.
              if (aktif) setTerpilih(null);
              else setOngkirManual("");
            }}
          />
          <span className="text-sm font-medium">Ketik ongkir sendiri</span>
        </label>

        {ongkirManualAktif ? (
          <Field
            label="Ongkir (Rp)"
            htmlFor="kemas_ongkir_manual"
            hint="Angka dari struk atau aplikasi kurir. Dipakai kalau daftar layanan tidak bisa diambil."
          >
            <Input
              id="kemas_ongkir_manual"
              type="number"
              min="0"
              step="any"
              value={ongkirManual}
              onChange={(event) => setOngkirManual(event.target.value)}
              placeholder="42000"
            />
          </Field>
        ) : (
          <p className="text-xs text-muted-foreground">
            Nyalakan kalau daftar layanan di atas tidak bisa diambil.
          </p>
        )}
      </div>

      {/*
       * Asuransi kiriman. Preminya diketik dari struk kurir, bukan dihitung
       * sistem: RajaOngkir tidak mengembalikan data asuransi sama sekali, dan
       * menanam rumus premi di sini berarti angka yang diam-diam meleset begitu
       * kurir mengubah tarifnya.
       */}
      <div className="space-y-2">
        <label className="flex cursor-pointer items-center gap-2">
          <Checkbox
            checked={pakaiAsuransi}
            onCheckedChange={(nilai) => setPakaiAsuransi(nilai === true)}
          />
          <span className="text-sm font-medium">Pakai asuransi kiriman</span>
        </label>

        {pakaiAsuransi && (
          <Field
            label="Premi asuransi (Rp)"
            htmlFor="kemas_premi"
            hint="Angka dari struk atau aplikasi kurir. Ditambahkan ke ongkir yang ditagihkan ke customer."
          >
            <Input
              id="kemas_premi"
              type="number"
              min="0"
              step="any"
              value={premi}
              onChange={(event) => setPremi(event.target.value)}
              placeholder="5000"
            />
          </Field>
        )}

        {/* Yang masuk ke tagihan customer adalah gabungannya, jadi angka itu
            yang ditampilkan — bukan dua angka yang harus dijumlah sendiri. */}
        {totalDitagihkan > 0 && (
          <p className="text-xs text-muted-foreground">
            Ditagihkan ke customer:{" "}
            <span className="font-medium">{formatIDR(totalDitagihkan)}</span>
            {toNumber(premiTersimpan) > 0 &&
              ` (ongkir ${formatIDR(ongkirTersimpan)} + asuransi ${formatIDR(premiTersimpan)})`}
          </p>
        )}
      </div>

      {/* Daftar periksa isi paket. Ditaruh sebelum catatan supaya urutannya
          mengikuti pekerjaannya: hitung barangnya dulu, baru tulis catatan. */}
      <div className="space-y-2">
        <p className="text-sm font-medium">Barang yang dipesan</p>

        {detail.isLoading ? (
          <p className="text-sm text-muted-foreground">Memuat daftar barang…</p>
        ) : daftarItem.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Daftar barang tidak bisa dimuat. Pengecekan dilewati supaya paket
            yang sudah siap tetap bisa dicatat.
          </p>
        ) : (
          <>
            <div className="divide-y divide-border rounded-lg border border-border">
              {daftarItem.map((baris) => {
                const kurang = baris.qty_purchased < baris.qty;
                return (
                  <label
                    key={baris.id}
                    className="flex cursor-pointer items-center gap-3 px-3 py-2 hover:bg-accent/50"
                  >
                    <Checkbox
                      checked={Boolean(dicek[baris.id])}
                      onCheckedChange={(nilai) =>
                        setDicek({ ...dicek, [baris.id]: nilai === true })
                      }
                    />
                    <span className="min-w-0 flex-1">
                      <span className="block text-sm font-medium">
                        {baris.product_name}
                      </span>
                      {/* Yang perlu dihitung adalah yang benar-benar terbeli,
                          bukan yang dulu dipesan — dan bedanya justru yang
                          paling gampang terlewat saat mengemas. */}
                      <span className="block text-xs text-muted-foreground">
                        {kurang
                          ? `Terbeli ${formatNumber(baris.qty_purchased)} dari ${formatNumber(baris.qty)} pcs dipesan`
                          : `${formatNumber(baris.qty)} pcs`}
                      </span>
                    </span>
                  </label>
                );
              })}
            </div>

            {semuaDicek ? (
              <p className="flex items-center gap-1.5 text-xs font-medium text-emerald-600">
                <Check className="size-3.5" />
                Semua barang sudah dicek — order siap dikemas.
              </p>
            ) : (
              <p className="text-xs text-muted-foreground">
                Centang tiap barang setelah dihitung fisiknya. Tombol Simpan
                terbuka setelah semuanya dicek ({formatNumber(jumlahDicek)} dari{" "}
                {formatNumber(daftarItem.length)}).
              </p>
            )}
          </>
        )}
      </div>

      <Field label="Catatan kemasan" htmlFor="kemas_catatan">
        <Textarea
          id="kemas_catatan"
          rows={2}
          value={form.notes}
          onChange={(event) => setForm({ ...form, notes: event.target.value })}
          placeholder="Barang pecah belah, dobel bubble wrap"
        />
      </Field>
    </FormDialog>
  );
}
