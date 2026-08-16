import { NextResponse } from "next/server";

import { backendURL, setSessionCookies } from "@/lib/session";
import type { Envelope, Session } from "@/types/api";

/**
 * Menukar email dan password dengan sesi, lalu menyimpan tokennya di cookie
 * httpOnly. Token tidak pernah dikirim balik ke browser sebagai JSON.
 */
export async function POST(request: Request) {
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json(
      { error: { code: "VALIDATION_ERROR", message: "body request tidak valid" } },
      { status: 400 },
    );
  }

  let response: Response;
  try {
    response = await fetch(`${backendURL()}/api/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
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

  const payload = (await response.json()) as Envelope<Session>;
  if (!response.ok || !payload.data) {
    return NextResponse.json(payload, { status: response.status });
  }

  await setSessionCookies(payload.data);

  // Hanya data pengguna yang dikembalikan; tokennya tinggal di cookie.
  return NextResponse.json({ data: { user: payload.data.user } });
}
