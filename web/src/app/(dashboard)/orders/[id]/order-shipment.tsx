"use client";

import {
  Calculator,
  CheckCircle2,
  MessageCircle,
  PackageCheck,
  Printer,
  Truck,
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { OptionSelect } from "@/components/filter-select";
import { ShipmentStatusBadge } from "@/components/status-badge";
import { ShippingInfoButton } from "@/components/shipping-info";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { ConfirmButton } from "@/components/ui/confirm-button";
import { FormDialog } from "@/components/ui/form-dialog";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NumberInput } from "@/components/ui/number-input";
import { Textarea } from "@/components/ui/textarea";
import { DetailRow } from "@/components/ui/page";
import { WAMessageDialog } from "@/components/wa-message-dialog";
import {
  useEstimateShipping,
  useMarkDelivered,
  useMarkShipmentNotified,
  usePackOrder,
  deliveryNoteUrl,
  useShipOrder,
  useShipmentMessage,
} from "@/hooks/use-operations";
import { ApiError } from "@/lib/api";
import { cn, formatDate, formatIDR, formatNumber, todayInput, toNumber } from "@/lib/utils";
import type { OrderDetail } from "@/types/api";

const JNE_SERVICE_OPTIONS = ["REG", "YES", "OKE", "JTR"].map((service) => ({
  value: service,
  label: service,
}));

export function OrderShipment({ order }: { order: OrderDetail }) {
  const [packOpen, setPackOpen] = useState(false);
  const [shipOpen, setShipOpen] = useState(false);
  const [messageOpen, setMessageOpen] = useState(false);

  // Berat default dijumlahkan dari berat produk kalau paketnya belum ditimbang.
  const [packForm, setPackForm] = useState({
    courier: order.shipment?.courier ?? "JNE",
    service: order.shipment?.service ?? "REG",
    weight_gram: order.shipment?.weight_gram ?? 0,
    length_cm: order.shipment?.length_cm ?? 0,
    width_cm: order.shipment?.width_cm ?? 0,
    height_cm: order.shipment?.height_cm ?? 0,
    notes: order.shipment?.notes ?? "",
  });

  const [shipForm, setShipForm] = useState({
    tracking_number: "",
    shipping_cost: order.shipping_fee,
    shipped_at: todayInput(),
    allow_unpaid: false,
  });

  const pack = usePackOrder(order.id);
  const estimate = useEstimateShipping(order.id);
  const ship = useShipOrder(order.id);
  const markDelivered = useMarkDelivered(order.id);
  const markNotified = useMarkShipmentNotified(order.id);
  const message = useShipmentMessage(order.id, messageOpen);

  const shipment = order.shipment;
  const balanceDue = toNumber(order.balance_due);

  function handlePack(event: React.FormEvent) {
    event.preventDefault();
    pack.mutate(
      {
        ...packForm,
        weight_gram: Number(packForm.weight_gram),
        length_cm: Number(packForm.length_cm),
        width_cm: Number(packForm.width_cm),
        height_cm: Number(packForm.height_cm),
        notes: packForm.notes || null,
      },
      {
        onSuccess: () => {
          toast.success("Paket ditandai sudah dikemas");
          setPackOpen(false);
        },
      },
    );
  }

  /*
   * Ongkir dihitung dari berat tercatat dan berat volume dimensi paket; yang
   * dipakai adalah yang lebih besar, sesuai cara ekspedisi menagih. Kota tujuan
   * diambil backend dari alamat kirim order sehingga tidak perlu diketik ulang.
   */
  function handleEstimate() {
    estimate.mutate(
      {
        courier: packForm.courier,
        service: packForm.service,
        weight_gram: Number(packForm.weight_gram),
        length_cm: Number(packForm.length_cm),
        width_cm: Number(packForm.width_cm),
        height_cm: Number(packForm.height_cm),
      },
      {
        onError: (error) => {
          toast.error(error instanceof ApiError ? error.message : "Gagal menghitung ongkir");
        },
      },
    );
  }

  function handleShip(event: React.FormEvent) {
    event.preventDefault();
    ship.mutate(shipForm, {
      onSuccess: () => {
        toast.success("Resi tersimpan, order ditandai dikirim");
        setShipOpen(false);
        // Customer perlu segera tahu nomor resinya.
        setMessageOpen(true);
      },
      onError: (error) => {
        toast.error(error instanceof ApiError ? error.message : "Gagal menyimpan resi");
      },
    });
  }

  // Order yang dibatalkan tidak dikemas dan tidak dikirim. Backend menolak
  // keduanya, jadi menampilkan tombolnya hanya mengundang klik yang berujung
  // pesan galat — dan surat jalan untuk paket yang tidak jadi berangkat justru
  // berbahaya kalau terlanjur tercetak dan ikut ditempel di kardus.
  const cancelled = order.status === "cancelled";

  return (
    <Card>
      <CardHeader>
        <CardTitle>Pengiriman</CardTitle>
        {shipment && (
          <CardAction>
            <ShipmentStatusBadge status={shipment.status} />
          </CardAction>
        )}
      </CardHeader>

      <CardContent className="space-y-4">
        {cancelled && !shipment ? (
          <p className="rounded-lg border border-border bg-muted/40 px-3 py-3 text-sm text-muted-foreground">
            Order ini dibatalkan, jadi tidak ada paket yang dikirim. Uang yang sudah diterima
            dikembalikan lewat pencatatan refund di kartu Pembayaran.
          </p>
        ) : shipment ? (
          <div className="divide-y divide-border">
            <DetailRow label="Kurir" value={`${shipment.courier} ${shipment.service}`} />
            <DetailRow
              label="Nomor resi"
              value={
                shipment.tracking_number ? (
                  <span className="tabular font-mono">{shipment.tracking_number}</span>
                ) : (
                  <span className="text-muted-foreground">belum diisi</span>
                )
              }
            />
            <DetailRow label="Berat" value={`${formatNumber(shipment.weight_gram)} g`} />
            {shipment.length_cm > 0 && (
              <DetailRow
                label="Dimensi"
                value={`${shipment.length_cm} × ${shipment.width_cm} × ${shipment.height_cm} cm`}
              />
            )}
            {toNumber(shipment.estimated_cost) > 0 && (
              <DetailRow
                label="Estimasi ongkir"
                value={formatIDR(shipment.estimated_cost)}
              />
            )}
            <DetailRow label="Ongkir dibayar" value={formatIDR(shipment.shipping_cost)} />
            {shipment.packed_at && (
              <DetailRow label="Dikemas" value={formatDate(shipment.packed_at)} />
            )}
            {shipment.shipped_at && (
              <DetailRow label="Dikirim" value={formatDate(shipment.shipped_at)} />
            )}
            {shipment.customer_notified_at && (
              <DetailRow
                label="Customer dikabari"
                value={formatDate(shipment.customer_notified_at)}
              />
            )}
          </div>
        ) : (
          <div className="flex flex-col items-center gap-2 rounded-lg border border-dashed border-border py-8 text-center">
            <Truck className="size-6 text-muted-foreground/60" />
            <p className="text-sm text-muted-foreground">
              Paket belum dikemas. Kemas dulu setelah barang selesai dicocokkan.
            </p>
          </div>
        )}

        {balanceDue > 0 && !cancelled && shipment?.status !== "shipped" && (
          <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
            Sisa tagihan {formatIDR(order.balance_due)} belum masuk. Catat pelunasan dulu sebelum
            paket dikirim.
          </p>
        )}

        <div className={cn("flex flex-wrap gap-2", cancelled && "hidden")}>
          <Button
            variant="outline"
            onClick={() => {
              pack.reset();
              setPackOpen(true);
            }}
          >
            <PackageCheck />
            {shipment ? "Ubah Data Kemasan" : "Tandai Dikemas"}
          </Button>

          {shipment?.status !== "shipped" && shipment?.status !== "delivered" && (
            <Button
              onClick={() => {
                ship.reset();
                setShipOpen(true);
              }}
            >
              <Truck />
              Input Resi &amp; Kirim
            </Button>
          )}

          {shipment?.tracking_number && (
            <Button variant="success" onClick={() => setMessageOpen(true)}>
              <MessageCircle />
              Kabari Customer
            </Button>
          )}

          {/* Surat jalan boleh dicetak kapan saja, termasuk sebelum resi ada:
              lembarnya dipakai sebagai pendamping saat mengemas dan saat
              menyerahkan paket ke konter kurir. */}
          <Button variant="outline" asChild>
            <a
              href={deliveryNoteUrl(order.id)}
              target="_blank"
              rel="noopener noreferrer"
            >
              <Printer />
              Surat Jalan
            </a>
          </Button>

          {shipment?.status === "shipped" && (
            <ConfirmButton
              variant="outline"
              title="Tandai paket sudah diterima?"
              description="Paket ditandai sampai di tangan customer dan order ditutup sebagai Selesai. Lakukan setelah customer benar-benar mengonfirmasi, bukan sekadar karena resinya sudah menunjukkan terkirim."
              confirmLabel="Ya, sudah diterima"
              error={markDelivered.error}
              onConfirm={() =>
                markDelivered.mutateAsync(undefined, {
                  onSuccess: () => toast.success("Order ditandai selesai"),
                  onError: (error) => {
                    toast.error(
                      error instanceof ApiError ? error.message : "Gagal menandai paket diterima",
                    );
                  },
                })
              }
            >
              <CheckCircle2 />
              Tandai Diterima
            </ConfirmButton>
          )}
        </div>
      </CardContent>

      <FormDialog
        open={packOpen}
        onOpenChange={setPackOpen}
        title="Data Kemasan"
        description="Catat kurir, layanan, dan berat paket sebelum diserahkan ke ekspedisi."
        error={pack.error}
        loading={pack.isPending}
        onSubmit={handlePack}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Kurir" htmlFor="courier">
            <Input
              id="courier"
              value={packForm.courier}
              onChange={(event) => setPackForm({ ...packForm, courier: event.target.value })}
              placeholder="JNE"
            />
          </Field>

          <Field label="Layanan" htmlFor="service">
            <OptionSelect
              id="service"
              value={packForm.service}
              onChange={(value) => setPackForm({ ...packForm, service: value })}
              options={JNE_SERVICE_OPTIONS}
            />
          </Field>

          <Field label="Berat (gram)" htmlFor="weight_gram">
            <NumberInput
              id="weight_gram"
              min="0"
              blankWhenZero
              placeholder="0"
              value={packForm.weight_gram}
              onValueChange={(weight) => setPackForm({ ...packForm, weight_gram: weight })}
            />
          </Field>

          <Field
            label="Dimensi paket (cm)"
            htmlFor="length_cm"
            hint="Panjang × lebar × tinggi. Kosongkan kalau paketnya kecil dan ringan."
            className="sm:col-span-2"
          >
            <div className="flex items-center gap-2">
              <NumberInput
                id="length_cm"
                min="0"
                blankWhenZero
                placeholder="0"
                value={packForm.length_cm}
                onValueChange={(length) => setPackForm({ ...packForm, length_cm: length })}
                aria-label="Panjang (cm)"
              />
              <span className="text-muted-foreground">×</span>
              <NumberInput
                min="0"
                blankWhenZero
                placeholder="0"
                value={packForm.width_cm}
                onValueChange={(width) => setPackForm({ ...packForm, width_cm: width })}
                aria-label="Lebar (cm)"
              />
              <span className="text-muted-foreground">×</span>
              <NumberInput
                min="0"
                blankWhenZero
                placeholder="0"
                value={packForm.height_cm}
                onValueChange={(height) => setPackForm({ ...packForm, height_cm: height })}
                aria-label="Tinggi (cm)"
              />
            </div>
          </Field>

          <div className="sm:col-span-2">
            <Button
              type="button"
              variant="outline"
              onClick={handleEstimate}
              loading={estimate.isPending}
            >
              <Calculator />
              Hitung Ongkir ke {order.shipping_city}
            </Button>
            <ShippingInfoButton className="ml-1 align-middle" />

            {estimate.data && (
              <div className="mt-3 space-y-1 rounded-lg border border-border bg-muted/40 px-3 py-2 text-sm">
                <p className="font-medium">
                  Perkiraan {formatIDR(estimate.data.cost)}
                  <span className="font-normal text-muted-foreground">
                    {" "}
                    · {estimate.data.courier} {estimate.data.service}
                    {estimate.data.etd && ` · ${estimate.data.etd}`}
                  </span>
                </p>
                <p className="text-xs text-muted-foreground">
                  Berat asli {formatNumber(estimate.data.actual_weight_gram)} g · berat volume{" "}
                  {formatNumber(estimate.data.volumetric_weight_gram)} g · ditagih{" "}
                  {formatNumber(estimate.data.chargeable_weight_gram / 1000)} kg ×{" "}
                  {formatIDR(estimate.data.price_per_kg)}
                </p>
                {!estimate.data.rate_found && (
                  <p className="text-xs text-amber-700">
                    Tarif kota ini belum ada, dipakai tarif umum. Tambahkan di Pengaturan → Tarif
                    ongkir supaya lebih akurat.
                  </p>
                )}
              </div>
            )}
          </div>

          <Field label="Catatan kemasan" htmlFor="pack_notes" className="sm:col-span-2">
            <Textarea
              id="pack_notes"
              rows={2}
              value={packForm.notes}
              onChange={(event) => setPackForm({ ...packForm, notes: event.target.value })}
              placeholder="Bubble wrap dobel, jangan ditumpuk, dan lain-lain"
            />
          </Field>
        </div>
      </FormDialog>

      <FormDialog
        open={shipOpen}
        onOpenChange={setShipOpen}
        title="Input Resi JNE"
        description="Setelah tersimpan, order langsung ditandai dikirim dan isinya tidak bisa diubah lagi."
        error={ship.error}
        loading={ship.isPending}
        onSubmit={handleShip}
        submitLabel="Simpan Resi"
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Nomor resi" htmlFor="tracking_number" required className="sm:col-span-2">
            <Input
              id="tracking_number"
              value={shipForm.tracking_number}
              onChange={(event) =>
                setShipForm({ ...shipForm, tracking_number: event.target.value.toUpperCase() })
              }
              placeholder="JNE0012345678"
              required
              autoFocus
            />
          </Field>

          <Field
            label="Ongkir dibayar (Rp)"
            htmlFor="shipping_cost"
            hint="Yang dibayar ke kurir, bukan yang ditagihkan ke customer"
          >
            <Input
              id="shipping_cost"
              type="number"
              min="0"
              step="any"
              value={shipForm.shipping_cost}
              onChange={(event) => setShipForm({ ...shipForm, shipping_cost: event.target.value })}
            />
            {toNumber(shipment?.estimated_cost) > 0 && (
              <button
                type="button"
                className="text-left text-xs text-primary hover:underline"
                onClick={() =>
                  setShipForm({ ...shipForm, shipping_cost: shipment!.estimated_cost })
                }
              >
                Pakai estimasi {formatIDR(shipment?.estimated_cost)}
              </button>
            )}
          </Field>

          <Field label="Tanggal kirim" htmlFor="shipped_at">
            <Input
              id="shipped_at"
              type="date"
              value={shipForm.shipped_at}
              onChange={(event) => setShipForm({ ...shipForm, shipped_at: event.target.value })}
            />
          </Field>

          {balanceDue > 0 && (
            <div className="flex items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm sm:col-span-2">
              <Checkbox
                id="allow_unpaid"
                className="mt-0.5"
                checked={shipForm.allow_unpaid}
                onCheckedChange={(checked) =>
                  setShipForm({ ...shipForm, allow_unpaid: checked === true })
                }
              />
              <Label htmlFor="allow_unpaid" className="cursor-pointer font-normal text-amber-900">
                Kirim walau belum lunas (sisa {formatIDR(order.balance_due)}). Pakai ini hanya untuk
                pelanggan yang memang dipercaya membayar setelah barang diterima.
              </Label>
            </div>
          )}
        </div>
      </FormDialog>

      <WAMessageDialog
        open={messageOpen}
        onOpenChange={setMessageOpen}
        title="Kabari nomor resi"
        message={message.data}
        isLoading={message.isLoading}
        error={message.error}
        onSent={() =>
          markNotified.mutate(undefined, {
            onSuccess: () => toast.success("Ditandai sudah dikabari"),
          })
        }
        sending={markNotified.isPending}
      />
    </Card>
  );
}
