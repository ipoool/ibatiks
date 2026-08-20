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
   * Premi asuransi kiriman. Diketik sendiri, bukan dihitung sistem: balasan
   * RajaOngkir hanya berisi nama kurir, layanan, ongkos, dan estimasi tiba —
   * tidak ada data asuransi sama sekali. Menanam rumus premi di kode berarti
   * angka yang tidak pernah ikut berubah saat kurir mengubah tarifnya, persis
   * alasan tabel tarif ongkir dulu dilepas.
   */
  const [pakaiAsuransi, setPakaiAsuransi] = useState(toNumber(item.insurance_fee) > 0);
  const [premi, setPremi] = useState(
    toNumber(item.insurance_fee) > 0 ? String(toNumber(item.insurance_fee)) : "",
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
  const ongkirTersimpan = terpilih ? String(toNumber(terpilih.cost)) : ongkirManual.trim();
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
          toast.success(ongkirTersimpan ? "Data kemasan dan ongkir tersimpan" : "Data kemasan tersimpan");
          onClose();
        },
        onError: (err) => {
          toast.error(err instanceof ApiError ? err.message : "Gagal menyimpan data kemasan");
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

      <Field label="Layanan kurir" hint="Ongkir yang ditagihkan ke customer diambil dari sini.">
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
                  terpilih?.courier === pilihan.courier && terpilih?.service === pilihan.service;
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
                      <Check className={cn("size-4 shrink-0", aktif ? "opacity-100" : "opacity-0")} />
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
                {terpilih.courier} {terpilih.service} · {formatIDR(terpilih.cost)}
              </span>
            </p>
          )}
        </div>
      </Field>

      {/*
       * Satu-satunya jaring pengaman sejak tabel tarif dilepas. Kurir sesekali
       * tidak bisa dihubungi, dan pengemasan tidak boleh berhenti karenanya —
       * angka dari struk konter lebih benar daripada tebakan mana pun.
       */}
      <Field
        label="Atau ketik ongkirnya sendiri (Rp)"
        htmlFor="kemas_ongkir_manual"
        hint="Dipakai kalau daftar layanan tidak bisa diambil. Mengetik di sini melepas pilihan layanan di atas."
      >
        <Input
          id="kemas_ongkir_manual"
          type="number"
          min="0"
          step="any"
          value={ongkirManual}
          onChange={(event) => {
            setOngkirManual(event.target.value);
            // Mengetik angka sendiri berarti melepas pilihan dari daftar;
            // dua sumber angka yang aktif bersamaan hanya membingungkan.
            setTerpilih(null);
          }}
          placeholder="42000"
        />
      </Field>

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
            Ditagihkan ke customer: <span className="font-medium">{formatIDR(totalDitagihkan)}</span>
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
            Daftar barang tidak bisa dimuat. Pengecekan dilewati supaya paket yang sudah siap
            tetap bisa dicatat.
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
                      <span className="block text-sm font-medium">{baris.product_name}</span>
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
                Centang tiap barang setelah dihitung fisiknya. Tombol Simpan terbuka setelah
                semuanya dicek ({formatNumber(jumlahDicek)} dari {formatNumber(daftarItem.length)}).
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
