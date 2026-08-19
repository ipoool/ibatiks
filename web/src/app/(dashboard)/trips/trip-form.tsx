"use client";

import { RefreshCw } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { CurrencySelect } from "@/components/currency-select";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useFetchExchangeRate, useSaveTrip, type TripPayload } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatDateTime, toDateInput, todayInput } from "@/lib/utils";
import type { Trip } from "@/types/api";

function emptyForm(): TripPayload {
  return {
    title: "",
    country: "",
    city: "",
    depart_date: todayInput(),
    return_date: todayInput(),
    order_deadline: "",
    currency: "JPY",
    exchange_rate: "",
    notes: "",
  };
}

interface TripFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Diisi saat mengubah trip yang sudah ada. */
  trip?: Trip;
}

export function TripFormDialog({ open, onOpenChange, trip }: TripFormDialogProps) {
  const router = useRouter();
  const save = useSaveTrip(trip?.id);
  const fetchRate = useFetchExchangeRate();

  function handleFetchRate() {
    fetchRate.mutate(form.currency.trim().toUpperCase(), {
      onSuccess: (result) => {
        setForm((current) => ({ ...current, exchange_rate: result.rate }));
        toast.success(`Kurs 1 ${result.from} = Rp${result.rate}`);
      },
      onError: (error) => {
        toast.error(error instanceof ApiError ? error.message : "Gagal mengambil kurs");
      },
    });
  }

  const [form, setForm] = useState<TripPayload>(() =>
    trip
      ? {
          title: trip.title,
          country: trip.country,
          city: trip.city ?? "",
          tripper_user_id: trip.tripper_user_id,
          depart_date: toDateInput(trip.depart_date),
          return_date: toDateInput(trip.return_date),
          order_deadline: toDateInput(trip.order_deadline),
          currency: trip.currency,
          exchange_rate: trip.exchange_rate,
          notes: trip.notes ?? "",
        }
      : emptyForm(),
  );

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    save.mutate(
      {
        ...form,
        city: form.city || null,
        // order_deadline kosong berarti tidak ada batas waktu order.
        order_deadline: form.order_deadline || undefined,
        notes: form.notes || null,
      },
      {
        onSuccess: (saved) => {
          toast.success(trip ? "Trip diperbarui" : "Trip dibuat");
          onOpenChange(false);
          if (!trip) {
            // Langsung buka detailnya supaya admin bisa menyusun katalog.
            router.push(`/trips/${saved.id}`);
          }
        },
      },
    );
  }

  const fieldError = (name: string) =>
    save.error instanceof ApiError ? save.error.fields?.[name] : undefined;

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={trip ? "Ubah Trip" : "Buat Trip Baru"}
      description="Kurs yang diisi di sini dipakai untuk menghitung seluruh harga jual pada trip ini."
      error={save.error}
      loading={save.isPending}
      onSubmit={handleSubmit}
      submitLabel={trip ? "Simpan" : "Buat Trip"}
      // Radix Select bukan <select> bawaan, jadi validasi browser tidak
      // melihatnya sama sekali. Tanpa penjagaan ini tombolnya tetap bisa
      // ditekan selagi mata uang kosong, dan gelembung yang muncul justru
      // menunjuk kolom lain yang kebetulan berupa input biasa.
      submitDisabled={form.currency.trim().length !== 3}
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Judul trip" htmlFor="title" required error={fieldError("title")} className="sm:col-span-2">
          <Input
            id="title"
            value={form.title}
            onChange={(event) => setForm({ ...form, title: event.target.value })}
            placeholder="Jastip Tokyo Maret 2026"
            required
          />
        </Field>

        <Field label="Negara" htmlFor="country" required error={fieldError("country")}>
          <Input
            id="country"
            value={form.country}
            onChange={(event) => setForm({ ...form, country: event.target.value })}
            placeholder="Jepang"
            required
          />
        </Field>

        <Field label="Kota" htmlFor="city">
          <Input
            id="city"
            value={form.city ?? ""}
            onChange={(event) => setForm({ ...form, city: event.target.value })}
            placeholder="Tokyo"
          />
        </Field>

        <Field label="Tanggal berangkat" htmlFor="depart_date" required error={fieldError("depart_date")}>
          <Input
            id="depart_date"
            type="date"
            value={form.depart_date}
            onChange={(event) => setForm({ ...form, depart_date: event.target.value })}
            required
          />
        </Field>

        <Field label="Tanggal pulang" htmlFor="return_date" required error={fieldError("return_date")}>
          <Input
            id="return_date"
            type="date"
            /*
             * Batasnya mengikuti tanggal berangkat supaya tanggal yang mundur
             * ditolak saat dipilih, bukan setelah formnya dikirim. Backend
             * tetap memeriksanya sendiri; ini hanya memindahkan penolakannya ke
             * tempat admin masih ingat sedang mengisi apa.
             */
            min={form.depart_date || undefined}
            value={form.return_date}
            onChange={(event) => setForm({ ...form, return_date: event.target.value })}
            required
          />
        </Field>

        <Field
          label="Batas terima order"
          htmlFor="order_deadline"
          error={fieldError("order_deadline")}
          hint="Kosongkan kalau tidak dibatasi"
        >
          <Input
            id="order_deadline"
            type="date"
            // Cerminan aturan backend: batas order tidak boleh melewati
            // tanggal pulang.
            max={form.return_date || undefined}
            value={form.order_deadline ?? ""}
            onChange={(event) => setForm({ ...form, order_deadline: event.target.value })}
          />
        </Field>

        <div className="grid grid-cols-2 gap-4">
          <Field label="Mata uang" htmlFor="currency" required error={fieldError("currency")}>
            <CurrencySelect
              id="currency"
              value={form.currency}
              onChange={(currency) => setForm({ ...form, currency })}
            />
          </Field>

          <Field
            label="Kurs ke Rupiah"
            htmlFor="exchange_rate"
            required
            error={fieldError("exchange_rate")}
          >
            <Input
              id="exchange_rate"
              type="number"
              min="0.000001"
              step="any"
              value={form.exchange_rate}
              onChange={(event) => setForm({ ...form, exchange_rate: event.target.value })}
              placeholder="108.5"
              required
            />
          </Field>

          {/*
            Kurs diambil dari layanan luar supaya tidak perlu diketik dari hasil
            mengintip aplikasi lain. Yang diambil hanya nilai awalnya: begitu
            trip tersimpan, kursnya terkunci dan seluruh harga pada trip ini
            memakai angka itu sampai selesai.
          */}
          <div className="col-span-2 flex flex-wrap items-center gap-x-3 gap-y-1">
            <Button
              type="button"
              variant="outline"
              size="sm"
              loading={fetchRate.isPending}
              disabled={form.currency.trim().length !== 3}
              onClick={handleFetchRate}
            >
              <RefreshCw />
              Ambil kurs terkini
            </Button>

            {fetchRate.data ? (
              <span className="text-xs text-muted-foreground">
                Diambil {formatDateTime(fetchRate.data.fetched_at)} dari {fetchRate.data.source}
              </span>
            ) : (
              <span className="text-xs text-muted-foreground">
                Kurs dikunci setelah trip tersimpan
              </span>
            )}
          </div>
        </div>

        <Field label="Catatan" htmlFor="notes" className="sm:col-span-2">
          <Textarea
            id="notes"
            rows={2}
            value={form.notes ?? ""}
            onChange={(event) => setForm({ ...form, notes: event.target.value })}
            placeholder="Rencana toko yang dikunjungi, batas bagasi, dan lain-lain"
          />
        </Field>
      </div>
    </FormDialog>
  );
}
