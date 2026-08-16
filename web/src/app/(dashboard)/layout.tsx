import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { AppShell } from "@/components/layout/shell";
import { getCurrentUser } from "@/lib/auth";
import { SIDEBAR_COOKIE } from "@/lib/sidebar";

// Seluruh halaman dashboard bergantung pada sesi, jadi tidak boleh di-cache.
export const dynamic = "force-dynamic";

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser();

  /*
   * Middleware sudah menyaring pengunjung tanpa cookie dan memperbarui token
   * yang kedaluwarsa; yang tersisa di sini adalah kasus cookie masih ada tapi
   * backend menolaknya — sesi dicabut, akunnya dinonaktifkan, atau database
   * diganti.
   *
   * Penanda `expired` wajib ikut. Tanpa itu middleware melihat cookie sesi yang
   * dari luar masih tampak sah, mengira pengguna sudah login, dan melempar
   * /login kembali ke sini — halaman lalu melemparnya lagi ke /login, dan
   * browser berhenti dengan galat "terlalu banyak pengalihan".
   */
  if (!user) {
    redirect("/login?expired=1");
  }

  /*
   * Lebar sidebar dibaca dari cookie di server, bukan dari localStorage di
   * klien: kalau dibaca setelah halaman terpasang, sidebar sempat terlukis
   * lebar lalu menciut sendiri di depan mata pengguna setiap kali pindah
   * halaman. Lewat cookie, tampilan pertama sudah benar.
   */
  const collapsed = (await cookies()).get(SIDEBAR_COOKIE)?.value === "1";


  return (
    <AppShell user={user} defaultCollapsed={collapsed}>
      {children}
    </AppShell>
  );
}
