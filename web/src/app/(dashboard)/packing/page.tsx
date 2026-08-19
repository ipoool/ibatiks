import { redirect } from "next/navigation";

/**
 * Antrean kemas sudah pindah menjadi tab di dalam Pengiriman.
 *
 * Alamat lamanya dibiarkan hidup dan dialihkan, bukan dihapus: tim toko
 * menyimpan tautan halaman yang dipakai tiap hari di bookmark browser, dan
 * "404 Not Found" bukan cara yang baik untuk memberi tahu bahwa menunya cuma
 * berpindah tempat.
 */
export default function PackingPage() {
  redirect("/shipments");
}
