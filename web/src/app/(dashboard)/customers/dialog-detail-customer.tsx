"use client";

import { labelPlatform } from "@/components/social-fields";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { DetailRow } from "@/components/ui/page";
import { formatDate } from "@/lib/utils";
import type { Customer } from "@/types/api";

/**
 * Seluruh data seorang customer, hanya untuk dibaca.
 *
 * Daftar customer memangkas isinya supaya barisnya tetap terbaca: alamat cuma
 * baris jalannya, kelurahan dan kecamatan hilang sama sekali, catatan tidak
 * pernah muncul. Padahal justru itu yang dicari saat hendak mengirim paket atau
 * menelepon. Sebelumnya satu-satunya cara melihatnya adalah membuka dialog Ubah
 * — memakai formulir penyuntingan untuk sekadar membaca, dengan tombol Simpan
 * yang sewaktu-waktu tertekan.
 */
export function DialogDetailCustomer({
  customer,
  onClose,
}: {
  customer: Customer;
  onClose: () => void;
}) {
  const alamatLengkap = [
    customer.address,
    customer.subdistrict,
    customer.district,
    customer.city,
    customer.province,
    customer.postal_code,
  ]
    .map((bagian) => bagian?.trim())
    .filter(Boolean)
    .join(", ");

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{customer.name}</DialogTitle>
          <DialogDescription>
            {customer.code} · terdaftar {formatDate(customer.created_at)}
          </DialogDescription>
        </DialogHeader>

        <div className="divide-y divide-border">
          <DetailRow
            label="Nomor WhatsApp"
            value={
              <a
                href={`https://wa.me/${customer.phone_wa}`}
                target="_blank"
                rel="noopener noreferrer"
                className="hover:underline"
              >
                {customer.phone_wa}
              </a>
            }
          />
          <DetailRow label="Email" value={customer.email || "—"} />
          <DetailRow
            label="Media sosial"
            value={
              customer.socials.length > 0
                ? customer.socials
                    .map((akun) => `${labelPlatform(akun.platform)} ${akun.handle}`)
                    .join(" · ")
                : "—"
            }
          />
          {/* Alamat ditulis utuh dalam satu baris, urut seperti orang menulisnya
              di amplop — itu bentuk yang tinggal disalin saat mengisi resi. */}
          <DetailRow
            label="Alamat pengiriman"
            value={
              alamatLengkap ? (
                <span className="whitespace-normal">{alamatLengkap}</span>
              ) : (
                "belum diisi"
              )
            }
          />
        </div>

        {customer.notes && (
          <div className="rounded-lg border border-border bg-muted/40 p-3">
            <p className="text-xs text-muted-foreground">Catatan</p>
            <p className="mt-0.5 text-sm whitespace-pre-line">{customer.notes}</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
