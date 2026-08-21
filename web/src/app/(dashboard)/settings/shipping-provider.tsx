"use client";

import { Calculator, MapPin, Save, Search } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NumberInput } from "@/components/ui/number-input";
import { ErrorState } from "@/components/ui/page";
import { useHasPermission } from "@/components/layout/user-context";
import { useDebounced } from "@/hooks/use-debounced";
import {
  useShippingDestinations,
  useShippingProvider,
  useTestShippingEstimate,
} from "@/hooks/use-operations";
import { useUpdateSettings } from "@/hooks/use-reports";
import { ApiError } from "@/lib/api";
import { formatIDR } from "@/lib/utils";
import type { ShippingDestination } from "@/types/api";

/**
 * Sambungan ke layanan tarif kurir (RajaOngkir).
 *
 * Ini satu-satunya sumber ongkir sejak tabel tarif dilepas. Kalau belum
 * terhubung, ongkirnya diketik sendiri di dialog kemas — bukan diambil dari
 * angka cadangan yang dirawat entah sejak kapan.
 *
 * Yang disimpan hanya tiga hal: ID kota asal, labelnya, dan daftar kurir yang
 * ditanyakan. API key-nya sendiri tidak pernah lewat sini — kunci itu tinggal di
 * `RAJAONGKIR_API_KEY` pada server, sehingga tidak pernah terkirim ke browser
 * siapa pun yang membuka halaman ini.
 */
export function ShippingProviderCard() {
  const canEdit = useHasPermission("settings");
  const { data: provider, isLoading, error } = useShippingProvider();
  const update = useUpdateSettings();

  const [pilihanAsal, setPilihanAsal] = useState<ShippingDestination | null>(null);
  const [pilihanKurir, setPilihanKurir] = useState<string[] | null>(null);
  const [gantiAsal, setGantiAsal] = useState(false);

  // Nilai tersimpan dari server dipakai selama pengguna belum mengubah apa pun,
  // supaya penyegaran cache di latar belakang tidak menghapus pilihan yang
  // sedang disusun.
  const asalTersimpan = provider?.origin_id
    ? { id: provider.origin_id, label: provider.origin_label || `ID ${provider.origin_id}` }
    : null;
  const asal = pilihanAsal
    ? { id: pilihanAsal.destination_id, label: pilihanAsal.label }
    : asalTersimpan;
  const kurir = pilihanKurir ?? provider?.couriers ?? [];

  const belumDiubah = pilihanAsal === null && pilihanKurir === null;

  function simpan() {
    if (!asal) {
      toast.error("Pilih kota asal pengiriman dulu");
      return;
    }
    if (kurir.length === 0) {
      toast.error("Pilih minimal satu kurir");
      return;
    }

    update.mutate(
      {
        shipping_origin_id: String(asal.id),
        shipping_origin_label: asal.label,
        // Pemisahnya titik dua karena itu bentuk yang diminta RajaOngkir.
        shipping_couriers: kurir.join(":"),
      },
      {
        onSuccess: () => {
          toast.success("Pengaturan pengiriman disimpan");
          setPilihanAsal(null);
          setPilihanKurir(null);
          setGantiAsal(false);
        },
        onError: (err) => {
          toast.error(err instanceof ApiError ? err.message : "Gagal menyimpan pengaturan");
        },
      },
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex flex-wrap items-center gap-2">
          Sambungan RajaOngkir
          {!isLoading &&
            (provider?.connected ? (
              <Badge variant="secondary">Terhubung</Badge>
            ) : (
              <Badge variant="outline">Belum terhubung</Badge>
            ))}
        </CardTitle>
        <CardDescription>
          {provider?.connected
            ? "Ongkir diambil langsung dari kurir sesuai berat paket. Kalau layanannya sedang tidak bisa dihubungi, ongkirnya diketik sendiri saat mengemas."
            : "API key belum dipasang di server. Isi RAJAONGKIR_API_KEY pada berkas .env lalu jalankan ulang backend; sampai itu terjadi, ongkir harus diketik sendiri saat mengemas."}
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-5">
        <ErrorState error={error ?? update.error} />

        <Field
          label="Kota asal pengiriman"
          hint="Titik berangkat paket. Asal yang keliru membuat seluruh perkiraan ongkir meleset."
        >
          {asal && !gantiAsal ? (
            <div className="flex flex-wrap items-center gap-2 rounded-md border border-input bg-card px-3 py-2">
              <MapPin className="size-4 shrink-0 text-muted-foreground" />
              <span className="min-w-0 flex-1 text-sm">{asal.label}</span>
              {canEdit && provider?.connected && (
                <Button type="button" variant="outline" size="sm" onClick={() => setGantiAsal(true)}>
                  Ganti
                </Button>
              )}
            </div>
          ) : (
            <PencarianAsal
              disabled={!canEdit || !provider?.connected}
              onPilih={(tujuan) => {
                setPilihanAsal(tujuan);
                setGantiAsal(false);
              }}
              onBatal={asal ? () => setGantiAsal(false) : undefined}
            />
          )}
        </Field>

        <Field
          label="Kurir yang ditanyakan"
          hint="Makin banyak kurir, makin banyak pilihan harga — tapi juga makin lama perhitungannya."
        >
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
            {(provider?.courier_options ?? []).map((pilihan) => (
              <CheckboxField
                key={pilihan.code}
                id={`kurir-${pilihan.code}`}
                checked={kurir.includes(pilihan.code)}
                disabled={!canEdit || !provider?.connected}
                onCheckedChange={(checked) =>
                  setPilihanKurir(
                    checked
                      ? [...kurir, pilihan.code]
                      : kurir.filter((kode) => kode !== pilihan.code),
                  )
                }
              >
                {pilihan.name}
              </CheckboxField>
            ))}
          </div>
        </Field>

        {canEdit && (
          <Button
            type="button"
            onClick={simpan}
            loading={update.isPending}
            disabled={!provider?.connected || belumDiubah}
          >
            <Save />
            Simpan
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

/** Kolom cari kota asal beserta daftar hasilnya. */
function PencarianAsal({
  disabled,
  onPilih,
  onBatal,
}: {
  disabled?: boolean;
  onPilih: (tujuan: ShippingDestination) => void;
  onBatal?: () => void;
}) {
  const [kataKunci, setKataKunci] = useState("");
  const pencarian = useDebounced(kataKunci);
  const { data, isFetching, error } = useShippingDestinations(pencarian, !disabled);

  return (
    <div className="space-y-2">
      <div className="flex gap-2">
        <div className="relative min-w-0 flex-1">
          <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={kataKunci}
            onChange={(event) => setKataKunci(event.target.value)}
            placeholder="Ketik nama kota atau kecamatan…"
            disabled={disabled}
            className="pl-9"
          />
        </div>
        {onBatal && (
          <Button type="button" variant="ghost" onClick={onBatal}>
            Batal
          </Button>
        )}
      </div>

      <ErrorState error={error} />

      {!disabled && kataKunci.trim().length > 0 && kataKunci.trim().length < 3 && (
        <p className="text-xs text-muted-foreground">Ketik minimal 3 huruf.</p>
      )}

      {isFetching && <p className="text-xs text-muted-foreground">Mencari…</p>}

      {/* Pesan "tidak ketemu" hanya kalau pencariannya memang berhasil dijalankan —
          menumpuknya di atas pesan galat membuat sebabnya jadi kabur. */}
      {!isFetching && !error && pencarian.trim().length >= 3 && (data?.length ?? 0) === 0 && (
        <p className="text-xs text-muted-foreground">
          Tidak ada yang cocok. Coba nama kecamatan atau kode posnya.
        </p>
      )}

      {(data?.length ?? 0) > 0 && (
        <ul className="max-h-64 divide-y overflow-y-auto rounded-md border">
          {data?.map((tujuan) => (
            <li key={tujuan.destination_id}>
              <button
                type="button"
                onClick={() => onPilih(tujuan)}
                className="w-full px-3 py-2 text-left text-sm hover:bg-accent focus-visible:bg-accent focus-visible:outline-none"
              >
                {tujuan.label}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

/**
 * Panel coba hitung.
 *
 * Ada supaya tim toko bisa memastikan sambungannya benar sebelum ada order
 * berjalan — daripada baru ketahuan salah asal saat customer sudah menunggu
 * angka ongkir.
 */
export function ShippingTestPanel() {
  const { data: provider } = useShippingProvider();
  const hitung = useTestShippingEstimate();
  const [form, setForm] = useState({ city: "", postal_code: "", weight_gram: 1000 });

  const hasil = hitung.data;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Coba hitung ongkir</CardTitle>
        <CardDescription>
          Uji satu alamat tujuan tanpa menyentuh order mana pun, untuk memastikan sambungannya
          benar sebelum ada order berjalan.
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-4">
        <ErrorState error={hitung.error} />

        <form
          className="grid gap-4 sm:grid-cols-3"
          onSubmit={(event) => {
            event.preventDefault();
            hitung.mutate({
              city: form.city,
              postal_code: form.postal_code || undefined,
              weight_gram: form.weight_gram,
            });
          }}
        >
          <Field label="Kota tujuan" htmlFor="uji-kota" required>
            <Input
              id="uji-kota"
              value={form.city}
              onChange={(event) => setForm({ ...form, city: event.target.value })}
              placeholder="Bandung"
              required
            />
          </Field>

          <Field label="Kode pos" htmlFor="uji-kodepos" hint="Membuat tujuannya lebih tepat">
            <Input
              id="uji-kodepos"
              value={form.postal_code}
              onChange={(event) => setForm({ ...form, postal_code: event.target.value })}
              placeholder="40115"
            />
          </Field>

          <Field label="Berat (gram)" htmlFor="uji-berat" required>
            <NumberInput
              id="uji-berat"
              value={form.weight_gram}
              onValueChange={(value) => setForm({ ...form, weight_gram: value })}
              min={1}
              step="any"
              blankWhenZero
              required
            />
          </Field>

          <div className="sm:col-span-3">
            <Button type="submit" loading={hitung.isPending}>
              <Calculator />
              Hitung
            </Button>
          </div>
        </form>

        {hasil && (
          <div className="space-y-2 rounded-md border bg-muted/40 p-4 text-sm">
            <p className="text-2xl font-semibold tabular">{formatIDR(hasil.cost)}</p>
            <p>
              {hasil.courier} {hasil.service}
              {hasil.etd && ` · estimasi ${hasil.etd}`}
            </p>
            {hasil.destination && (
              <p className="text-muted-foreground">Tujuan: {hasil.destination}</p>
            )}
            <p className="text-muted-foreground">
              Berat ditagih {(hasil.chargeable_weight_gram / 1000).toLocaleString("id-ID")} kg ·
              sumber angka: {hasil.source}
            </p>
          </div>
        )}

        {/*
         * Tidak ada lagi sumber angka cadangan, jadi tanpa sambungan yang siap
         * tombol Hitung pasti gagal. Alasannya disebutkan lebih dulu supaya
         * admin tidak menekannya lalu mengira ada yang rusak.
         */}
        {!provider?.ready && (
          <p className="text-xs text-amber-600">
            {provider?.connected
              ? "Kota asal belum dipilih di kartu di atas, jadi ongkir belum bisa dihitung."
              : "RajaOngkir belum terhubung, jadi ongkir belum bisa dihitung. Isi RAJAONGKIR_API_KEY di server."}
          </p>
        )}
      </CardContent>
    </Card>
  );
}
