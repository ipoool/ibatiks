import { NextResponse } from "next/server";

import { backendURL, clearSessionCookies, getRefreshToken, setSessionCookies } from "@/lib/session";
import type { Envelope, Session } from "@/types/api";

/**
 * Menukar refresh token dengan sesi baru dan menuliskannya kembali ke cookie.
 *
 * Dipanggil middleware sebelum halaman dirender di server. Proxy BFF sudah
 * bisa memperbarui token untuk panggilan dari browser, tapi render halaman di
 * server tidak lewat proxy — tanpa jalur ini, membuka ulang halaman setelah
 * access token 15 menit habis akan melempar admin ke halaman login walau
 * refresh token-nya masih berlaku seminggu.
 *
 * Token baru hanya dikirim sebagai cookie httpOnly, tidak pernah sebagai isi
 * response: rute ini bisa dipanggil dari browser, dan mengembalikan token
 * dalam bentuk JSON berarti menyerahkannya ke skrip apa pun yang berjalan di
 * halaman.
 */
export async function POST() {
  const refreshToken = await getRefreshToken();
  if (!refreshToken) {
    return NextResponse.json(
      { error: { code: "UNAUTHORIZED", message: "sesi berakhir, silakan login ulang" } },
      { status: 401 },
    );
  }

  let response: Response;
  try {
    response = await fetch(`${backendURL()}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
      cache: "no-store",
    });
  } catch {
    return NextResponse.json(
      {
        error: {
          code: "BACKEND_UNAVAILABLE",
          message: "server tidak bisa dihubungi, coba lagi sebentar lagi",
        },
      },
      { status: 503 },
    );
  }

  const payload = (await response.json().catch(() => ({}))) as Envelope<Session>;
  if (!response.ok || !payload.data) {
    // Refresh token ditolak backend — bersihkan cookienya supaya penjaga rute
    // langsung mengarahkan ke login alih-alih mencoba lagi tiap navigasi.
    await clearSessionCookies();
    return NextResponse.json(
      { error: { code: "UNAUTHORIZED", message: "sesi berakhir, silakan login ulang" } },
      { status: 401 },
    );
  }

  await setSessionCookies(payload.data);
  return NextResponse.json({ data: { user: payload.data.user } });
}
