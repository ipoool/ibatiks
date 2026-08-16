import "server-only";

import { backendURL, getAccessToken } from "@/lib/session";
import type { Envelope, User } from "@/types/api";

/**
 * Mengambil pengguna yang sedang login dari sisi server.
 *
 * Dipanggil di layout dashboard sehingga identitas dan role sudah tersedia
 * pada render pertama — tanpa kedipan menu yang muncul lalu hilang karena
 * hak aksesnya baru diketahui belakangan.
 */
export async function getCurrentUser(): Promise<User | null> {
  const accessToken = await getAccessToken();
  if (!accessToken) return null;

  try {
    const response = await fetch(`${backendURL()}/api/v1/auth/me`, {
      headers: { Authorization: `Bearer ${accessToken}` },
      cache: "no-store",
    });
    if (!response.ok) return null;

    const payload = (await response.json()) as Envelope<User>;
    return payload.data ?? null;
  } catch {
    // Backend belum siap saat halaman dibuka; layout akan mengarahkan ke login.
    return null;
  }
}
