"use client";

import Link from "next/link";
import { useState } from "react";

import { FilterSelect } from "@/components/filter-select";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { Badge } from "@/components/ui/badge";
import { ErrorState } from "@/components/ui/page";
import { Pagination } from "@/components/ui/pagination";
import { useAuditActors, useAuditLogs } from "@/hooks/use-reports";
import { formatDateTime, formatIDR, formatNumber } from "@/lib/utils";
import type { AuditLog } from "@/types/api";
import { LABEL_PENGATURAN } from "./setting-fields";

/*
 * Jejak perubahan, ditulis untuk dibaca tim toko — bukan untuk dibaca programer.
 *
 * Kolom `changes` di database berisi JSON dengan nama kunci apa adanya
 * ("weight_gram", "shipping_cost"). Menampilkannya mentah membuat kolom Detail
 * jadi deretan tanda kurung yang tidak menjawab pertanyaan yang membuat orang
 * membuka layar ini: apa persisnya yang berubah, dan jadi berapa.
 */

const ENTITY_LABEL: Record<string, string> = {
  order: "Order",
  invoice: "Invoice",
  shipment: "Pengiriman",
  purchase: "Pembelian",
  trip: "Trip",
  settings: "Pengaturan",
  payment: "Pembayaran",
};

const ENTITY_OPTIONS = [
  { value: "order", label: "Order" },
  { value: "invoice", label: "Invoice" },
  { value: "shipment", label: "Pengiriman" },
  { value: "purchase", label: "Pembelian" },
  { value: "trip", label: "Trip" },
  { value: "settings", label: "Pengaturan" },
];

/**
 * Judul aksi, dibaca bersama entitasnya.
 *
 * "Dibuat" pada invoice dan "Dibuat" pada order adalah dua pekerjaan yang
 * berbeda di meja toko, jadi keduanya diberi kalimatnya sendiri. Pasangan yang
 * tidak terdaftar jatuh ke label aksi umum — jejak yang belum sempat diberi
 * kalimat tetap harus terbaca, bukan hilang.
 */
const AKSI_LABEL: Record<string, string> = {
  "order:create": "Order dicatat",
  "order:update": "Order diubah",
  "order:delete": "Order dihapus",
  "order:status_change": "Status order diubah",
  "order:payment_record": "Pembayaran dicatat",
  "order:item_change": "Item order diubah",
  "order:receive": "Barang diterima",
  "invoice:create": "Invoice diterbitkan",
  "invoice:sent": "Invoice dikirim ke customer",
  "invoice:void": "Invoice dibatalkan",
  "shipment:pack": "Paket dikemas",
  "shipment:ship": "Paket diserahkan ke kurir",
  "shipment:delivered": "Paket diterima customer",
  "purchase:create": "Pembelian dicatat",
  "purchase:delete": "Pembelian dihapus",
  "trip:create": "Trip dibuat",
  "trip:update": "Trip diubah",
  "trip:delete": "Trip dihapus",
  "settings:update": "Pengaturan toko diubah",
};

const AKSI_UMUM: Record<string, string> = {
  create: "Dibuat",
  update: "Diubah",
  delete: "Dihapus",
  status_change: "Ubah status",
  payment_record: "Catat pembayaran",
  item_change: "Ubah item",
  receive: "Terima barang",
  pack: "Kemas",
  ship: "Kirim",
  sent: "Kirim invoice",
  void: "Batalkan invoice",
  delivered: "Diterima customer",
};

function labelAksi(entity: string, action: string): string {
  return AKSI_LABEL[`${entity}:${action}`] ?? AKSI_UMUM[action] ?? action;
}

/** Aksi yang menghapus sesuatu ditandai merah — itu yang paling dicari orang. */
const AKSI_MERAH = new Set(["delete", "void"]);

/*
 * Urutan bacanya sengaja tetap, mengikuti urutan penulisan di bawah — bukan
 * urutan kunci di dalam JSON-nya. Postgres mengembalikan kunci jsonb menurut
 * panjang lalu abjad, jadi tanpa ini satu jenis kejadian yang sama bisa terbaca
 * dengan urutan berbeda di tiap baris.
 *
 * Yang menjawab "apa" lebih dulu (nomor dan jenis), baru "berapa".
 */
const KOLOM_LABEL: Record<string, string> = {
  trip_code: "Kode trip",
  order_number: "Nomor order",
  invoice_number: "Nomor invoice",
  tracking_number: "Nomor resi",
  entity: "Jenis data",
  type: "Jenis",
  method: "Metode",
  channel: "Dikirim lewat",
  courier: "Kurir",
  service: "Layanan",
  item_count: "Jumlah item",
  qty: "Qty",
  weight_gram: "Berat",
  amount: "Nominal",
  total: "Total",
  discount: "Diskon",
  shipping: "Ongkir",
  shipping_cost: "Ongkir",
  unit_cost_idr: "Harga modal satuan",
  to_orders: "Dialokasikan ke order",
  to_stock: "Masuk ke stok",
  allow_unpaid: "Dikirim walau belum lunas",
  orders: "Order ikut terhapus",
  invoices: "Invoice ikut terhapus",
  purchases: "Pembelian ikut terhapus",
  shipments: "Paket ikut terhapus",
  expenses: "Biaya trip ikut terhapus",
  catalog_items: "Item katalog ikut terhapus",
  payments_total: "Uang yang sudah diterima",
  purchases_cost: "Nilai pembelian",
  keys: "Kolom yang diubah",
};

const KOLOM_URUTAN = Object.keys(KOLOM_LABEL);

/*
 * Hitungan bernilai nol dilewati, tapi nominal rupiah tidak pernah.
 *
 * Menghapus trip membuang seluruh riwayat di dalamnya, dan "uang yang sudah
 * diterima Rp 0" adalah keterangan yang justru harus terbaca — itu satu-satunya
 * jejak yang tersisa setelah barisnya hilang. Sementara deretan "Order: 0,
 * Invoice: 0, Paket: 0" cuma membuat baris yang penting tenggelam.
 */
const KOLOM_SEMBUNYI_NOL = new Set([
  "orders",
  "invoices",
  "purchases",
  "shipments",
  "expenses",
  "catalog_items",
  "to_orders",
  "to_stock",
]);

/** Kolom yang isinya rupiah. Nilainya string, sesuai aturan uang di proyek ini. */
const KOLOM_UANG = new Set([
  "amount",
  "discount",
  "payments_total",
  "purchases_cost",
  "shipping",
  "shipping_cost",
  "total",
  "unit_cost_idr",
]);

/** Nilai berkode yang punya nama sendiri di layar. */
const NILAI_LABEL: Record<string, Record<string, string>> = {
  type: {
    dp: "DP",
    down_payment: "DP",
    final: "Pelunasan",
    settlement: "Pelunasan",
    refund: "Pengembalian dana",
  },
  method: {
    transfer: "Transfer bank",
    cash: "Tunai",
    ewallet: "E-wallet",
    qris: "QRIS",
    other: "Lainnya",
  },
  channel: { wa: "WhatsApp", email: "Email" },
  entity: ENTITY_LABEL,
};

interface Rincian {
  label: string;
  nilai: string;
}

function nilaiTeks(kolom: string, nilai: unknown): string {
  if (nilai === null || nilai === undefined) return "—";
  if (typeof nilai === "boolean") return nilai ? "Ya" : "Tidak";

  if (Array.isArray(nilai)) {
    // Daftar kunci pengaturan diterjemahkan ke nama yang dibaca orang.
    const isi = nilai.map((item) =>
      typeof item === "string" ? (LABEL_PENGATURAN[item] ?? item) : String(item),
    );
    return isi.length > 0 ? isi.join(", ") : "—";
  }

  if (KOLOM_UANG.has(kolom)) return formatIDR(nilai as string);
  if (kolom === "weight_gram") return `${formatNumber(nilai as number)} gram`;

  const terjemahan = NILAI_LABEL[kolom]?.[String(nilai)];
  if (terjemahan) return terjemahan;

  if (typeof nilai === "number") return formatNumber(nilai);
  return String(nilai);
}

/**
 * Mengubah isi kolom `changes` jadi baris-baris yang bisa dibaca.
 *
 * Pasangan `x_from` dan `x_to` digabung jadi satu baris "sebelum → sesudah".
 * Dipisah dua baris, angka lama dan angka baru terbaca seperti dua nilai yang
 * berdiri sendiri, padahal yang dicari orang justru selisihnya.
 */
function bacaPerubahan(changes: AuditLog["changes"]): Rincian[] {
  if (!changes || typeof changes !== "object") return [];

  const isi = changes as Record<string, unknown>;
  const out: Rincian[] = [];
  const sudah = new Set<string>();

  for (const [kolom, nilai] of Object.entries(isi)) {
    if (sudah.has(kolom)) continue;

    if (kolom.endsWith("_from")) {
      const dasar = kolom.slice(0, -"_from".length);
      const pasangan = `${dasar}_to`;
      if (pasangan in isi) {
        sudah.add(kolom);
        sudah.add(pasangan);
        out.push({
          label: KOLOM_LABEL[dasar] ?? dasar,
          nilai: `${nilaiTeks(dasar, nilai)} → ${nilaiTeks(dasar, isi[pasangan])}`,
        });
        continue;
      }
    }

    sudah.add(kolom);
    if (KOLOM_SEMBUNYI_NOL.has(kolom) && nilai === 0) continue;
    out.push({ label: KOLOM_LABEL[kolom] ?? kolom, nilai: nilaiTeks(kolom, nilai) });
  }

  // Kolom yang tidak dikenal ikut ditampilkan apa adanya, di urutan paling
  // belakang: jejak yang belum sempat diberi nama tetap harus terbaca.
  return out.sort((a, b) => urutanKolom(a.label) - urutanKolom(b.label));
}

function urutanKolom(label: string): number {
  const posisi = KOLOM_URUTAN.findIndex((kolom) => KOLOM_LABEL[kolom] === label);
  return posisi === -1 ? KOLOM_URUTAN.length : posisi;
}

/** Alamat halaman yang memuat entitas ini, kalau ada. */
function tautanEntitas(log: AuditLog): string | null {
  if (!log.entity_id) return null;
  if (log.entity === "order") return `/orders/${log.entity_id}`;
  if (log.entity === "trip") return `/trips/${log.entity_id}`;
  return null;
}

export function AuditTrail() {
  const [page, setPage] = useState(1);
  const [entity, setEntity] = useState("");
  const [userID, setUserID] = useState("");

  const { data, isLoading, error } = useAuditLogs({
    page,
    entity: entity || undefined,
    user_id: userID || undefined,
  });
  const { data: actors } = useAuditActors();

  const actorOptions = (actors ?? []).map((actor) => ({
    value: actor.id,
    label: `${actor.name} (${formatNumber(actor.log_count)})`,
  }));

  function ganti(setter: (value: string) => void) {
    return (value: string) => {
      setter(value);
      // Halaman dikembalikan ke satu: halaman lima dari hasil lama hampir
      // selalu kosong pada hasil yang baru, dan yang terbaca "tidak ada jejak".
      setPage(1);
    };
  }

  return (
    <>
      <ErrorState error={error} />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Catatan siapa mengubah apa — berguna saat menelusuri perubahan qty atau nominal order.
        </p>
        <div className="flex flex-wrap items-center gap-2">
          <FilterSelect
            value={userID}
            onChange={ganti(setUserID)}
            allLabel="Semua akun"
            options={actorOptions}
            className="sm:w-52"
          />
          <FilterSelect
            value={entity}
            onChange={ganti(setEntity)}
            allLabel="Semua jenis data"
            options={ENTITY_OPTIONS}
            className="sm:w-44"
          />
        </div>
      </div>

      <div>
        <DataTable
          columns={4}
          isLoading={isLoading}
          isEmpty={!isLoading && (data?.items.length ?? 0) === 0}
          emptyTitle="Belum ada jejak perubahan"
          emptyDescription="Perubahan yang dilakukan tim toko akan tercatat di sini."
          head={
            <TR>
              <TH>Waktu</TH>
              <TH className="hidden md:table-cell">Pengguna</TH>
              <TH>Yang dikerjakan</TH>
              <TH className="hidden sm:table-cell">Detail</TH>
            </TR>
          }
        >
          {data?.items.map((log) => {
            const rincian = bacaPerubahan(log.changes);
            const tautan = tautanEntitas(log);

            return (
              <TR key={log.id}>
                <TD className="text-sm">
                  <span className="whitespace-nowrap">{formatDateTime(log.created_at)}</span>
                  {/* Nama pengguna menyusul waktu saat kolomnya disembunyikan:
                      "siapa" adalah setengah dari pertanyaan yang dijawab layar
                      ini, jadi ia tidak boleh hilang di layar sempit. */}
                  <p className="text-xs text-muted-foreground md:hidden">
                    {log.user_name ?? "sistem"}
                  </p>
                </TD>
                <TD className="hidden text-sm md:table-cell">{log.user_name ?? "sistem"}</TD>
                <TD className="whitespace-normal">
                  <Badge variant={AKSI_MERAH.has(log.action) ? "danger" : "neutral"}>
                    {labelAksi(log.entity, log.action)}
                  </Badge>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {tautan ? (
                      <Link
                        href={tautan}
                        className="underline decoration-dotted underline-offset-2 hover:text-foreground"
                      >
                        {ENTITY_LABEL[log.entity] ?? log.entity}
                      </Link>
                    ) : (
                      (ENTITY_LABEL[log.entity] ?? log.entity)
                    )}
                  </p>
                  {/* Detail ikut ke kolom utama di layar sempit — ia justru
                      bagian yang paling dicari, jadi menyembunyikannya saja
                      menyisakan tabel yang tidak menjawab apa-apa. */}
                  {rincian.length > 0 && (
                    <div className="mt-1 space-y-0.5 sm:hidden">
                      {rincian.map((baris) => (
                        <p key={baris.label} className="text-xs text-muted-foreground">
                          {baris.label}: <span className="text-foreground">{baris.nilai}</span>
                        </p>
                      ))}
                    </div>
                  )}
                </TD>
                <TD className="hidden whitespace-normal sm:table-cell">
                  {rincian.length > 0 ? (
                    <div className="space-y-0.5">
                      {rincian.map((baris) => (
                        <p key={baris.label} className="text-xs text-muted-foreground">
                          {baris.label}: <span className="text-foreground">{baris.nilai}</span>
                        </p>
                      ))}
                    </div>
                  ) : (
                    <span className="text-xs text-muted-foreground">
                      Tidak ada rincian tambahan
                    </span>
                  )}
                </TD>
              </TR>
            );
          })}
        </DataTable>

        <Pagination meta={data?.meta} onPageChange={setPage} />
      </div>
    </>
  );
}
