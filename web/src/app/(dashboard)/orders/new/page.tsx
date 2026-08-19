"use client";

import { ArrowLeft, PackagePlus, Plus, Trash2, UserPlus } from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useMemo, useRef, useState } from "react";
import { toast } from "sonner";

import {
  DP_KOSONG,
  InputDP,
  keteranganDP,
  dpKeRupiah,
  type NilaiDP,
} from "@/components/input-dp";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { OptionSelect } from "@/components/filter-select";
import { ORDER_SOURCE_OPTIONS } from "@/components/status-badge";
import { QuickAddCatalogItemDialog } from "@/components/quick-add-catalog-item";
import { QuickAddCustomerDialog } from "@/components/quick-add-customer";
import { Button } from "@/components/ui/button";
import { Combobox } from "@/components/ui/combobox";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { DetailRow, ErrorState } from "@/components/ui/page";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import { useCustomers } from "@/hooks/use-master";
import { useCreateOrder } from "@/hooks/use-orders";
import { useTripItems, useTrips } from "@/hooks/use-trips";
import { ApiError } from "@/lib/api";
import { formatIDR, formatNumber, todayInput, toNumber } from "@/lib/utils";
import type { OrderSource } from "@/types/api";

interface DraftItem {
  product_id: string;
  name: string;
  sku: string;
  qty: number;
  unit_price: string;
  max_qty: number | null;
  qty_ordered: number;
}

export default function NewOrderPage() {
  return (
    <Suspense fallback={<p className="text-sm text-muted-foreground">Memuat…</p>}>
      <NewOrderForm />
    </Suspense>
  );
}

function NewOrderForm() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const [tripId, setTripId] = useState(searchParams.get("trip_id") ?? "");
  const [customerId, setCustomerId] = useState("");
  const [orderDate, setOrderDate] = useState(todayInput());
  const [orderSource, setOrderSource] = useState<OrderSource>("whatsapp");
  const [items, setItems] = useState<DraftItem[]>([]);
  const [discount, setDiscount] = useState("0");
  const [dp, setDp] = useState<NilaiDP>(DP_KOSONG);
  const [notes, setNotes] = useState("");
  const [overrideAddress, setOverrideAddress] = useState(false);
  const [customerFormOpen, setCustomerFormOpen] = useState(false);
  const [catalogFormOpen, setCatalogFormOpen] = useState(false);
  const [address, setAddress] = useState({
    recipient_name: "",
    recipient_phone: "",
    shipping_address: "",
    shipping_city: "",
    shipping_district: "",
    shipping_subdistrict: "",
    shipping_province: "",
    shipping_postal_code: "",
  });

  const { data: trips } = useTrips({ per_page: 100 });
  const { data: customers } = useCustomers({ per_page: 200 });
  const tripOptions =
    trips?.items.map((trip) => ({
      value: trip.id,
      label: trip.title,
      keywords: trip.code,
      description: trip.code,
    })) ?? [];
  const customerOptions =
    customers?.items.map((customer) => ({
      value: customer.id,
      label: customer.name,
      keywords: customer.phone_wa,
      description: customer.phone_wa,
    })) ?? [];
  const { data: catalog } = useTripItems(tripId || undefined);
  const createOrder = useCreateOrder();

  /*
   * Penanda "sedang mengirim" yang berubah saat itu juga, bukan setelah render
   * berikutnya. Lihat handleSubmit: tanpa ini, klik ganda membuat order kembar.
   */
  const sedangMengirim = useRef(false);

  const selectedCustomer = customers?.items.find((customer) => customer.id === customerId);
  const selectedTrip = trips?.items.find((trip) => trip.id === tripId);

  // Alamat customer disalin sebagai titik awal begitu opsi "kirim ke alamat
  // lain" dinyalakan, supaya admin tinggal mengubah bagian yang berbeda.
  // Penyalinan dilakukan di sini, bukan lewat efek yang mengamati state:
  // dengan begitu alamat yang sudah diketik admin tidak tertimpa ulang.
  function toggleOverrideAddress(enabled: boolean) {
    setOverrideAddress(enabled);
    if (enabled && selectedCustomer) {
      setAddress({
        recipient_name: selectedCustomer.name,
        recipient_phone: selectedCustomer.phone_wa,
        shipping_address: selectedCustomer.address ?? "",
        shipping_city: selectedCustomer.city ?? "",
        shipping_district: selectedCustomer.district ?? "",
        shipping_subdistrict: selectedCustomer.subdistrict ?? "",
        shipping_province: selectedCustomer.province ?? "",
        shipping_postal_code: selectedCustomer.postal_code ?? "",
      });
    }
  }

  // Ganti trip berarti katalognya berbeda, jadi item yang sudah dipilih dibuang.
  function changeTrip(nextTripId: string) {
    setTripId(nextTripId);
    setItems([]);
  }

  const subtotal = useMemo(
    () => items.reduce((sum, item) => sum + item.qty * toNumber(item.unit_price), 0),
    [items],
  );
  /*
   * Ongkir belum ada di tahap ini. Beratnya baru diketahui setelah barangnya
   * datang dan paketnya ditimbang, jadi ongkir ditetapkan belakangan di menu
   * Pengiriman dan total order ikut naik saat itu.
   */
  const total = Math.max(subtotal - toNumber(discount), 0);
  const suggestedDP = Math.round(total / 2);
  // Selalu rupiah, apa pun satuan yang sedang dipakai admin — nilai inilah yang
  // dikirim ke API dan yang dibandingkan dengan patokan setengah.
  const dpRupiah = dpKeRupiah(dp, total);
  const dpDiTawar = dpRupiah !== "" && toNumber(dpRupiah) < suggestedDP;

  function addItem(productId: string) {
    const catalogItem = catalog?.find((item) => item.product_id === productId);
    if (!catalogItem) return;

    setItems((current) => {
      const existing = current.find((item) => item.product_id === productId);
      if (existing) {
        return current.map((item) =>
          item.product_id === productId ? { ...item, qty: item.qty + 1 } : item,
        );
      }
      return [
        ...current,
        {
          product_id: productId,
          name: catalogItem.product_name,
          sku: catalogItem.product_sku,
          qty: 1,
          unit_price: catalogItem.sell_price,
          max_qty: catalogItem.max_qty,
          qty_ordered: catalogItem.qty_ordered,
        },
      ];
    });
  }

  function updateItem(productId: string, patch: Partial<DraftItem>) {
    setItems((current) =>
      current.map((item) => (item.product_id === productId ? { ...item, ...patch } : item)),
    );
  }

  function removeItem(productId: string) {
    setItems((current) => current.filter((item) => item.product_id !== productId));
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    /*
     * Tombolnya memang dinonaktifkan selama pengiriman berjalan, tapi baik
     * atribut disabled maupun createOrder.isPending baru berubah setelah React
     * merender ulang. Dua kiriman di tick yang sama — klik ganda, tombol Enter
     * yang ditahan — sama-sama membaca nilai lama dan lolos berdua, lalu jadi
     * order kembar untuk customer yang sama.
     *
     * Ref berubah saat itu juga, jadi penjagaannya tidak bergantung pada waktu
     * render.
     */
    if (sedangMengirim.current) {
      return;
    }
    sedangMengirim.current = true;

    const batalkanPenjagaan = () => {
      sedangMengirim.current = false;
    };

    if (!tripId || !customerId) {
      batalkanPenjagaan();
      toast.error("Pilih trip dan customer terlebih dahulu");
      return;
    }
    if (items.length === 0) {
      batalkanPenjagaan();
      toast.error("Tambahkan minimal satu produk");
      return;
    }

    createOrder.mutate(
      {
        trip_id: tripId,
        customer_id: customerId,
        order_date: orderDate,
        order_source: orderSource,
        items: items.map((item) => ({
          product_id: item.product_id,
          qty: item.qty,
          unit_price: item.unit_price,
        })),
        discount,
        dp_required: dpRupiah || undefined,
        notes: notes || null,
        ...(overrideAddress ? address : {}),
      },
      {
        onSuccess: (order) => {
          toast.success(`Order ${order.order_number} dibuat`);
          router.push(`/orders/${order.id}`);
        },
        onError: (error) => {
          // Gagal berarti belum ada order yang terbentuk, jadi admin harus
          // boleh mencoba lagi.
          batalkanPenjagaan();
          toast.error(error instanceof ApiError ? error.message : "Gagal membuat order");
        },
      },
    );
  }

  const catalogOptions =
    catalog
      ?.filter((item) => item.is_active)
      .map((item) => ({
        value: item.product_id,
        label: item.product_name,
        keywords: item.product_sku,
        description: `${item.product_sku} · ${formatIDR(item.sell_price)}`,
      })) ?? [];

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild className="-ml-2">
          <Link href="/orders">
            <ArrowLeft />
            Semua order
          </Link>
        </Button>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Catat Order Baru</h1>
          <p className="text-sm text-muted-foreground">
            Harga otomatis diambil dari katalog trip dan disalin ke order, jadi tidak berubah
            walau katalog diedit belakangan.
          </p>
        </div>
      </div>

      <ErrorState error={createOrder.error} />

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="min-w-0 space-y-6 lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>Trip &amp; customer</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2">
              <Field label="Trip" htmlFor="trip_id" required>
                <Combobox
                  id="trip_id"
                  value={tripId}
                  onChange={changeTrip}
                  options={tripOptions}
                  placeholder="Pilih trip…"
                  searchPlaceholder="Cari kode atau judul trip…"
                  emptyLabel="Trip tidak ditemukan"
                />
              </Field>

              <Field label="Customer" htmlFor="customer_id" required>
                <div className="flex gap-2">
                  <Combobox
                    id="customer_id"
                    value={customerId}
                    onChange={setCustomerId}
                    options={customerOptions}
                    placeholder="Pilih customer…"
                    searchPlaceholder="Cari nama atau nomor WA…"
                    emptyLabel="Customer tidak ditemukan"
                    className="min-w-0 flex-1"
                  />
                  {/* Customer baru sering muncul saat order sedang dicatat;
                      tanpa jalan pintas ini admin harus meninggalkan form yang
                      belum tersimpan untuk membuatnya lebih dulu. */}
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        onClick={() => setCustomerFormOpen(true)}
                      >
                        <UserPlus />
                        <span className="sr-only">Tambah customer baru</span>
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Tambah customer baru</TooltipContent>
                  </Tooltip>
                </div>
              </Field>

              <Field label="Tanggal order" htmlFor="order_date">
                <Input
                  id="order_date"
                  type="date"
                  value={orderDate}
                  onChange={(event) => setOrderDate(event.target.value)}
                />
              </Field>

              <Field
                label="Channel"
                htmlFor="order_source"
                hint="Dipakai laporan penjualan per channel"
              >
                <OptionSelect
                  id="order_source"
                  value={orderSource}
                  onChange={setOrderSource}
                  options={ORDER_SOURCE_OPTIONS}
                />
              </Field>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Produk yang dipesan</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Field label="Tambah dari katalog trip" htmlFor="add_product">
                {/*
                  Select ini bukan penyimpan nilai, melainkan pemicu: begitu
                  produk dipilih ia langsung masuk tabel dan pilihannya
                  dikosongkan lagi. Karena itu `value` sengaja dibiarkan kosong
                  supaya placeholder-nya kembali muncul dan produk yang sama
                  bisa ditambahkan dua kali berturut-turut.
                */}
                <div className="flex gap-2">
                  <Combobox
                    id="add_product"
                    // Bukan penyimpan nilai melainkan pemicu: begitu produk
                    // dipilih ia langsung masuk tabel dan pilihannya dikosongkan
                    // lagi, supaya produk yang sama bisa ditambah berturut-turut.
                    value=""
                    onChange={addItem}
                    options={catalogOptions}
                    placeholder={tripId ? "Pilih produk untuk ditambahkan…" : "Pilih trip dulu"}
                    searchPlaceholder="Cari nama produk atau SKU…"
                    emptyLabel="Produk tidak ada di katalog trip ini"
                    disabled={!tripId}
                    className="min-w-0 flex-1"
                  />
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        disabled={!tripId}
                        onClick={() => setCatalogFormOpen(true)}
                      >
                        <PackagePlus />
                        <span className="sr-only">Tambah produk ke katalog trip</span>
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Tambah produk ke katalog trip</TooltipContent>
                  </Tooltip>
                </div>
              </Field>

              <DataTable
                columns={5}
                isEmpty={items.length === 0}
                emptyTitle="Belum ada produk"
                emptyDescription="Pilih produk dari katalog trip di atas."
                head={
                  <TR>
                    <TH>Produk</TH>
                    <TH className="w-28 text-right">Qty</TH>
                    <TH className="w-40 text-right">Harga</TH>
                    <TH className="hidden text-right sm:table-cell">Subtotal</TH>
                    <TH className="w-12" />
                  </TR>
                }
              >
                {items.map((item) => {
                  const remaining =
                    item.max_qty === null ? null : item.max_qty - item.qty_ordered;
                  const overQuota = remaining !== null && item.qty > remaining;

                  return (
                    <TR key={item.product_id}>
                      <TD>
                        <p className="font-medium">{item.name}</p>
                        <p className="text-xs text-muted-foreground">{item.sku}</p>
                        {overQuota && (
                          <p className="text-xs font-medium text-destructive">
                            Kuota tersisa {formatNumber(Math.max(remaining ?? 0, 0))} pcs
                          </p>
                        )}
                      </TD>
                      <TD>
                        <Input
                          type="number"
                          min="1"
                          value={item.qty}
                          onChange={(event) =>
                            updateItem(item.product_id, { qty: Math.max(Number(event.target.value), 1) })
                          }
                          className="h-9 text-right"
                        />
                      </TD>
                      <TD>
                        <Input
                          type="number"
                          min="0"
                          step="any"
                          value={item.unit_price}
                          onChange={(event) =>
                            updateItem(item.product_id, { unit_price: event.target.value })
                          }
                          className="h-9 text-right"
                        />
                        <p className="tabular mt-1 text-right text-xs font-medium sm:hidden">
                          {formatIDR(item.qty * toNumber(item.unit_price))}
                        </p>
                      </TD>
                      {/* Subtotal tetap terbaca di ponsel lewat keterangan di
                          bawah harga; kolomnya sendiri disembunyikan karena dua
                          kolom isian sudah memenuhi lebar layar. */}
                      <TD className="tabular hidden text-right font-medium sm:table-cell">
                        {formatIDR(item.qty * toNumber(item.unit_price))}
                      </TD>
                      <TD>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          tooltip="Hapus dari pesanan"
                          className="text-destructive hover:text-destructive"
                          onClick={() => removeItem(item.product_id)}
                        >
                          <Trash2 />
                          <span className="sr-only">Hapus dari pesanan</span>
                        </Button>
                      </TD>
                    </TR>
                  );
                })}
              </DataTable>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Alamat pengiriman</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <CheckboxField
                id="override_address"
                checked={overrideAddress}
                onCheckedChange={toggleOverrideAddress}
              >
                Kirim ke alamat lain (hadiah, kantor, titip teman)
              </CheckboxField>

              {overrideAddress ? (
                <div className="grid gap-4 sm:grid-cols-2">
                  <Field label="Nama penerima" htmlFor="recipient_name" required>
                    <Input
                      id="recipient_name"
                      value={address.recipient_name}
                      onChange={(event) =>
                        setAddress({ ...address, recipient_name: event.target.value })
                      }
                      required
                    />
                  </Field>
                  <Field label="Nomor HP penerima" htmlFor="recipient_phone" required>
                    <Input
                      id="recipient_phone"
                      value={address.recipient_phone}
                      onChange={(event) =>
                        setAddress({ ...address, recipient_phone: event.target.value })
                      }
                      required
                    />
                  </Field>
                  <Field label="Alamat" htmlFor="shipping_address" required className="sm:col-span-2">
                    <Textarea
                      id="shipping_address"
                      value={address.shipping_address}
                      onChange={(event) =>
                        setAddress({ ...address, shipping_address: event.target.value })
                      }
                      required
                    />
                  </Field>
                  {/* Kelurahan dan kecamatan berdiri sendiri, bukan dijejalkan
                      ke kolom alamat: keduanya yang dibaca kurir saat menyortir. */}
                  <Field label="Kelurahan" htmlFor="shipping_subdistrict">
                    <Input
                      id="shipping_subdistrict"
                      value={address.shipping_subdistrict}
                      onChange={(event) =>
                        setAddress({ ...address, shipping_subdistrict: event.target.value })
                      }
                    />
                  </Field>
                  <Field label="Kecamatan" htmlFor="shipping_district">
                    <Input
                      id="shipping_district"
                      value={address.shipping_district}
                      onChange={(event) =>
                        setAddress({ ...address, shipping_district: event.target.value })
                      }
                    />
                  </Field>
                  <Field label="Kota/Kabupaten" htmlFor="shipping_city" required>
                    <Input
                      id="shipping_city"
                      value={address.shipping_city}
                      onChange={(event) =>
                        setAddress({ ...address, shipping_city: event.target.value })
                      }
                      required
                    />
                  </Field>
                  {/* Tiap isian alamat berdiri langsung di grid formnya, bukan dibungkus
                      grid dua kolom lagi. Grid dua kolom di dalam satu kolom grid dua kolom
                      menyisakan seperempat lebar untuk tiap isian — dan justru nama kelurahan
                      dan kecamatan yang paling panjang. */}
                  <Field label="Provinsi" htmlFor="shipping_province">
                    <Input
                      id="shipping_province"
                      value={address.shipping_province}
                      onChange={(event) =>
                        setAddress({ ...address, shipping_province: event.target.value })
                      }
                    />
                  </Field>
                  <Field label="Kode pos" htmlFor="shipping_postal_code">
                    <Input
                      id="shipping_postal_code"
                      value={address.shipping_postal_code}
                      onChange={(event) =>
                        setAddress({ ...address, shipping_postal_code: event.target.value })
                      }
                    />
                  </Field>
                </div>
              ) : selectedCustomer ? (
                <div className="rounded-lg border border-border bg-muted/40 p-4 text-sm">
                  <p className="font-medium">{selectedCustomer.name}</p>
                  <p className="text-muted-foreground">{selectedCustomer.phone_wa}</p>
                  <p className="mt-1 text-muted-foreground">
                    {selectedCustomer.address || (
                      <span className="text-destructive">
                        Alamat customer masih kosong — lengkapi dulu atau isi alamat lain di sini.
                      </span>
                    )}
                  </p>
                  <p className="text-muted-foreground">
                    {[selectedCustomer.city, selectedCustomer.province, selectedCustomer.postal_code]
                      .filter(Boolean)
                      .join(", ")}
                  </p>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">
                  Pilih customer untuk melihat alamat pengirimannya.
                </p>
              )}
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card className="lg:sticky lg:top-6">
            <CardHeader>
              <CardTitle>Ringkasan</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="divide-y divide-border">
                <DetailRow label="Subtotal" value={formatIDR(subtotal)} />
                <DetailRow label="Diskon" value={`−${formatIDR(discount)}`} />
                <DetailRow
                  label="Total barang"
                  value={<span className="text-base font-semibold">{formatIDR(total)}</span>}
                />
              </div>

              <Field label="Diskon (Rp)" htmlFor="discount">
                <Input
                  id="discount"
                  type="number"
                  min="0"
                  step="any"
                  value={discount}
                  onChange={(event) => setDiscount(event.target.value)}
                />
              </Field>

              <Field
                label="DP diminta"
                htmlFor="dp_required"
                hint={
                  keteranganDP(dp, total) ??
                  `Kosongkan untuk 50% otomatis (${formatIDR(suggestedDP)})`
                }
              >
                <InputDP
                  id="dp_required"
                  value={dp}
                  onChange={setDp}
                  nilaiBarang={total}
                  placeholder={dp.satuan === "persen" ? "50" : String(suggestedDP)}
                />
                {/* Peringatan, bukan penolakan. Setengah harga barang adalah
                    patokan yang menutup modal belanja; menurunkannya untuk
                    customer lama tetap boleh, asal disadari. */}
                {dpDiTawar && (
                  <p className="text-xs text-amber-600">
                    Di bawah setengah nilai barang ({formatIDR(suggestedDP)}). Modal belanja
                    sebagian ditalangi toko sampai pelunasan masuk.
                  </p>
                )}
              </Field>

              <Field label="Catatan" htmlFor="order_notes">
                <Textarea
                  id="order_notes"
                  rows={2}
                  value={notes}
                  onChange={(event) => setNotes(event.target.value)}
                  placeholder="Permintaan khusus dari customer"
                />
              </Field>

              {selectedTrip && (
                <p className="text-xs text-muted-foreground">
                  Trip {selectedTrip.code} · kurs 1 {selectedTrip.currency} = Rp
                  {formatNumber(selectedTrip.exchange_rate)}
                </p>
              )}

              <Button type="submit" className="w-full" loading={createOrder.isPending}>
                <Plus />
                Buat Order
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>

      <QuickAddCustomerDialog
        open={customerFormOpen}
        onOpenChange={setCustomerFormOpen}
        onCreated={setCustomerId}
      />

      {selectedTrip && (
        <QuickAddCatalogItemDialog
          trip={selectedTrip}
          open={catalogFormOpen}
          onOpenChange={setCatalogFormOpen}
          onCreated={addItem}
        />
      )}
    </form>
  );
}
