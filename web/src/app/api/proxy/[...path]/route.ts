import { NextResponse, type NextRequest } from "next/server";

import {
  backendURL,
  clearSessionCookies,
  getAccessToken,
  getRefreshToken,
  setSessionCookies,
} from "@/lib/session";
import type { Envelope, Session } from "@/types/api";

/**
 * BFF proxy: seluruh panggilan API dari browser lewat sini.
 *
 * Tugasnya tiga: menyisipkan access token dari cookie httpOnly, memperbarui
 * token yang kedaluwarsa lalu mengulang request aslinya, dan meneruskan
 * response apa adanya termasuk berkas PDF maupun CSV.
 */

type RouteContext = { params: Promise<{ path: string[] }> };

export async function GET(request: NextRequest, context: RouteContext) {
  return handle(request, context);
}
export async function POST(request: NextRequest, context: RouteContext) {
  return handle(request, context);
}
export async function PUT(request: NextRequest, context: RouteContext) {
  return handle(request, context);
}
export async function PATCH(request: NextRequest, context: RouteContext) {
  return handle(request, context);
}
export async function DELETE(request: NextRequest, context: RouteContext) {
  return handle(request, context);
}

async function handle(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  const target = `${backendURL()}/api/v1/${path.join("/")}${request.nextUrl.search}`;

  // Body dibaca sekali ke buffer supaya request bisa diulang setelah token
  // diperbarui; stream request hanya bisa dikonsumsi satu kali.
  const body =
    request.method === "GET" || request.method === "DELETE"
      ? undefined
      : await request.arrayBuffer();

  let accessToken = await getAccessToken();
  let response: Response;

  try {
    response = await forward(target, request, accessToken, body);
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

  if (response.status === 401) {
    accessToken = await refreshAccessToken();
    if (!accessToken) {
      await clearSessionCookies();
      return NextResponse.json(
        { error: { code: "UNAUTHORIZED", message: "sesi berakhir, silakan login ulang" } },
        { status: 401 },
      );
    }

    try {
      response = await forward(target, request, accessToken, body);
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
  }

  return passthrough(response);
}

function forward(
  target: string,
  request: NextRequest,
  accessToken: string | undefined,
  body: ArrayBuffer | undefined,
): Promise<Response> {
  const headers = new Headers();

  // Hanya header yang relevan yang diteruskan; sisanya (termasuk cookie sesi
  // Next) tidak ada urusannya dengan backend.
  const contentType = request.headers.get("content-type");
  if (contentType) headers.set("Content-Type", contentType);
  headers.set("Accept", request.headers.get("accept") ?? "application/json");
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);

  return fetch(target, {
    method: request.method,
    headers,
    body: body && body.byteLength > 0 ? body : undefined,
    cache: "no-store",
    redirect: "manual",
  });
}

/** Menukar refresh token dengan sesi baru dan memperbarui cookie. */
async function refreshAccessToken(): Promise<string | undefined> {
  const refreshToken = await getRefreshToken();
  if (!refreshToken) return undefined;

  try {
    const response = await fetch(`${backendURL()}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
      cache: "no-store",
    });
    if (!response.ok) return undefined;

    const payload = (await response.json()) as Envelope<Session>;
    if (!payload.data) return undefined;

    await setSessionCookies(payload.data);
    return payload.data.access_token;
  } catch {
    return undefined;
  }
}

/** Meneruskan response backend apa adanya, termasuk PDF dan CSV. */
function passthrough(response: Response) {
  const headers = new Headers();
  for (const key of ["content-type", "content-disposition", "content-length", "cache-control"]) {
    const value = response.headers.get(key);
    if (value) headers.set(key, value);
  }

  return new NextResponse(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}
