import { NextResponse } from "next/server";

import { backendURL, clearSessionCookies, getRefreshToken } from "@/lib/session";

/**
 * Mencabut sesi di backend lalu menghapus cookie.
 *
 * Cookie tetap dihapus meskipun backend gagal dihubungi: dari sisi pengguna,
 * menekan tombol keluar harus selalu berhasil mengeluarkannya dari aplikasi.
 */
export async function POST() {
  const refreshToken = await getRefreshToken();

  if (refreshToken) {
    try {
      await fetch(`${backendURL()}/api/v1/auth/logout`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
        cache: "no-store",
      });
    } catch {
      // Diabaikan dengan sengaja; sesi tetap dibersihkan di sisi browser.
    }
  }

  await clearSessionCookies();
  return NextResponse.json({ data: { message: "berhasil keluar" } });
}
