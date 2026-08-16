"use client";

import { ClipboardList } from "lucide-react";
import { useState } from "react";

import { FilterSelect } from "@/components/filter-select";
import { Card, CardContent } from "@/components/ui/card";
import { ErrorState, PageHeader } from "@/components/ui/page";
import { useTrips } from "@/hooks/use-trips";

import { TripShopping } from "../trips/[id]/trip-shopping";

/**
 * Halaman khusus tripper: pilih trip, lalu belanja dari daftarnya.
 *
 * Isinya memakai komponen yang sama dengan tab "Belanja" di detail trip supaya
 * perilakunya tidak pernah berbeda antara dua pintu masuk yang sama-sama sah.
 */
export default function ShoppingListPage() {
  const [chosenTripId, setChosenTripId] = useState("");
  const { data: trips, isLoading, error } = useTrips({ per_page: 100 });
  const tripOptions =
    trips?.items.map((trip) => ({ value: trip.id, label: `${trip.code} — ${trip.title}` })) ?? [];

  // Trip yang sedang berjalan dipakai sebagai pilihan awal supaya tripper tidak
  // perlu memilih apa pun saat membuka halaman ini di tengah belanja. Nilainya
  // diturunkan langsung dari data, bukan disalin ke state lewat efek, sehingga
  // pilihan tripper sendiri selalu menang atas nilai bawaan.
  // Trip yang sedang berjalan dipilih lebih dulu; kalau semuanya sudah
  // ditutup, yang terbaru tetap ditawarkan supaya halaman tidak kosong.
  const defaultTrip = trips?.items.find((trip) => trip.status === "open") ?? trips?.items[0];
  const tripId = chosenTripId || defaultTrip?.id || "";

  const selectedTrip = trips?.items.find((trip) => trip.id === tripId);

  return (
    <>
      <PageHeader
        title="Daftar Belanja"
        description="Apa saja yang harus dibeli di negara tujuan, diringkas dari seluruh pesanan"
        actions={
          <FilterSelect
            value={tripId}
            onChange={setChosenTripId}
            allLabel="Pilih trip…"
            options={tripOptions}
            className="sm:w-72"
          />
        }
      />

      <ErrorState error={error} />

      {!isLoading && trips?.items.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center gap-2 py-16 text-center">
            <ClipboardList className="size-8 text-muted-foreground/60" />
            <p className="font-medium">Belum ada trip</p>
            <p className="max-w-sm text-sm text-muted-foreground">
              Daftar belanja terbentuk otomatis dari pesanan pada sebuah trip.
            </p>
          </CardContent>
        </Card>
      )}

      {selectedTrip && <TripShopping trip={selectedTrip} />}
    </>
  );
}
