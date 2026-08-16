import "server-only";

import { cookies } from "next/headers";

import type { Session } from "@/types/api";

/**
 * Token JWT tidak pernah disimpan di localStorage.
 *
 * Semua panggilan dari browser lewat route handler Next (BFF) yang membaca
 * token dari cookie httpOnly dan menyisipkannya sebagai header Authorization.
 * Konsekuensinya, skrip pihak ketiga yang berhasil masuk ke halaman tetap
 * tidak bisa membaca token sesi.
 */
export const ACCESS_TOKEN_COOKIE = "jastipin_at";
export const REFRESH_TOKEN_COOKIE = "jastipin_rt";

/** URL backend dari sisi server Next (di dalam jaringan Docker). */
export function backendURL(): string {
  return (
    process.env.BACKEND_INTERNAL_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    "http://localhost:8080"
  );
}

function cookieOptions(maxAge: number) {
  return {
    httpOnly: true,
    // Cookie hanya dikirim lewat HTTPS di production; di development http
    // lokal tetap perlu jalan.
    secure: process.env.NODE_ENV === "production",
    // lax cukup untuk aplikasi back office satu domain, dan tetap menahan
    // pengiriman cookie pada request lintas situs yang berbahaya.
    sameSite: "lax" as const,
    path: "/",
    maxAge,
  };
}

export async function setSessionCookies(session: Session) {
  const store = await cookies();

  // Access token berumur pendek (mengikuti JWT_ACCESS_TTL backend); refresh
  // token yang menentukan berapa lama admin bisa tetap login.
  store.set(ACCESS_TOKEN_COOKIE, session.access_token, cookieOptions(60 * 30));
  store.set(REFRESH_TOKEN_COOKIE, session.refresh_token, cookieOptions(60 * 60 * 24 * 7));
}

export async function clearSessionCookies() {
  const store = await cookies();
  store.delete(ACCESS_TOKEN_COOKIE);
  store.delete(REFRESH_TOKEN_COOKIE);
}

export async function getAccessToken(): Promise<string | undefined> {
  return (await cookies()).get(ACCESS_TOKEN_COOKIE)?.value;
}

export async function getRefreshToken(): Promise<string | undefined> {
  return (await cookies()).get(REFRESH_TOKEN_COOKIE)?.value;
}
