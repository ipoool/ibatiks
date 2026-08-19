"use client";

import { ArrowLeft, Loader2, Pencil } from "lucide-react";
import Link from "next/link";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";

import { TripStatusBadge } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/page";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useTrip } from "@/hooks/use-trips";
import { formatDate, formatNumber } from "@/lib/utils";

import { TripFormDialog } from "../trip-form";
import { TripCatalog } from "./trip-catalog";
import { TripDeleteButton } from "./trip-delete-button";
import { TripExpenses } from "./trip-expenses";
import { TripOrders } from "./trip-orders";
import { TripProfit } from "./trip-profit";
import { TripShopping } from "./trip-shopping";
import { TripStatusActions } from "./trip-status-actions";

export default function TripDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const searchParams = useSearchParams();
  const tripId = params.id;

  const [editOpen, setEditOpen] = useState(false);
  const { data: trip, isLoading, error } = useTrip(tripId);

  // Tab aktif disimpan di URL supaya halaman bisa dibagikan dan tombol
  // "kembali" browser bekerja seperti yang diharapkan.
  const activeTab = searchParams.get("tab") ?? "katalog";

  function handleTabChange(value: string) {
    const next = new URLSearchParams(searchParams.toString());
    next.set("tab", value);
    router.replace(`/trips/${tripId}?${next.toString()}`, { scroll: false });
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-16 text-muted-foreground">
        <Loader2 className="size-4 animate-spin" />
        Memuat trip…
      </div>
    );
  }

  if (error || !trip) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/trips">
            <ArrowLeft />
            Kembali ke daftar trip
          </Link>
        </Button>
        <ErrorState error={error ?? new Error("Trip tidak ditemukan")} />
      </div>
    );
  }

  return (
    <>
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild className="-ml-2">
          <Link href="/trips">
            <ArrowLeft />
            Semua trip
          </Link>
        </Button>

        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-2xl font-semibold tracking-tight">{trip.title}</h1>
              <TripStatusBadge status={trip.status} />
            </div>
            <p className="text-sm text-muted-foreground">
              {trip.code} · {trip.country}
              {trip.city ? `, ${trip.city}` : ""} · {formatDate(trip.depart_date)} –{" "}
              {formatDate(trip.return_date)} · 1 {trip.currency} = Rp
              {formatNumber(trip.exchange_rate)}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <Button variant="outline" onClick={() => setEditOpen(true)}>
              <Pencil />
              Ubah
            </Button>
            <TripStatusActions trip={trip} />
            {/* Hapus duduk paling kanan, terpisah dari aksi sehari-hari. */}
            <TripDeleteButton trip={trip} />
          </div>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList>
          <TabsTrigger value="katalog">Katalog</TabsTrigger>
          <TabsTrigger value="order">Order</TabsTrigger>
          <TabsTrigger value="belanja">Belanja</TabsTrigger>
          <TabsTrigger value="biaya">Biaya</TabsTrigger>
          <TabsTrigger value="profit">Profit</TabsTrigger>
        </TabsList>

        <TabsContent value="katalog">
          <TripCatalog trip={trip} />
        </TabsContent>
        <TabsContent value="order">
          <TripOrders trip={trip} />
        </TabsContent>
        <TabsContent value="belanja">
          <TripShopping trip={trip} />
        </TabsContent>
        <TabsContent value="biaya">
          <TripExpenses trip={trip} />
        </TabsContent>
        <TabsContent value="profit">
          <TripProfit trip={trip} />
        </TabsContent>
      </Tabs>

      {editOpen && (
        // Dibuat ulang tiap dibuka supaya isian form selalu mulai dari data terkini.
        <TripFormDialog open={editOpen} onOpenChange={setEditOpen} trip={trip} />
      )}
    </>
  );
}
