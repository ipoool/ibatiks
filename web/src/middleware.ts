import { NextResponse, type NextRequest } from "next/server";

import { canOpenPath } from "@/lib/route-permissions";

const ACCESS_TOKEN_COOKIE = "jastipin_at";
const REFRESH_TOKEN_COOKIE = "jastipin_rt";

/** Ambang aman sebelum kedaluwarsa, supaya token tidak habis di tengah render. */
const EXPIRY_SKEW_SECONDS = 30;

/**
 * Penjaga rute halaman.
 *
 * Middleware hanya memeriksa keberadaan cookie sesi, bukan keabsahannya —
 * verifikasi tanda tangan token tetap dilakukan backend pada setiap request.
 * Tujuannya sekadar mengarahkan orang yang belum login ke halaman login,
 * bukan menjadi lapisan keamanan.
 */
export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasSession = Boolean(request.cookies.get(REFRESH_TOKEN_COOKIE)?.value);
  const isLoginPage = pathname === "/login";

  if (!hasSession && !isLoginPage) {
    // Tujuan awal disimpan supaya setelah login pengguna kembali ke halaman
    // yang tadi ingin dibuka.
    return NextResponse.redirect(loginRedirectURL(request));
  }

  /*
   * Halaman dashboard melaporkan bahwa backend menolak sesi ini lewat
   * `?expired=1`. Cookienya dibuang di sini, bukan dibiarkan, supaya tidak
   * terjadi tarik-menarik pengalihan antara halaman yang tahu sesinya mati dan
   * middleware yang hanya melihat cookienya masih ada.
   */
  if (isLoginPage && request.nextUrl.searchParams.has("expired")) {
    const response = NextResponse.next();
    response.cookies.delete(ACCESS_TOKEN_COOKIE);
    response.cookies.delete(REFRESH_TOKEN_COOKIE);
    return response;
  }

  /*
   * Access token diperbarui di sini, sebelum halaman dirender.
   *
   * Render halaman berjalan di server dan tidak lewat proxy BFF, sehingga
   * pembaruan token yang sudah ada di proxy tidak menolongnya: begitu access
   * token 15 menit habis, layout gagal mengambil identitas pengguna dan
   * melempar admin ke halaman login padahal refresh token-nya masih berlaku
   * seminggu. Middleware satu-satunya tempat yang bisa menulis cookie sekaligus
   * membetulkan request yang sedang dilayani.
   */
  if (hasSession && isAccessTokenStale(request)) {
    const refreshed = await refreshSession(request);

    if (refreshed.status === "gagal") {
      /*
       * Refresh token ditolak backend — misalnya sesinya sudah dicabut atau
       * database di-reset. Pemeriksaan ini sengaja dijalankan juga di halaman
       * login: kalau tidak, cookie basi membuat /login dilempar ke "/", lalu
       * "/" melemparnya balik ke /login, dan pengguna terjebak dalam putaran
       * pengalihan yang hanya bisa diputus dengan menghapus cookie manual.
       */
      const response = isLoginPage
        ? NextResponse.next()
        : NextResponse.redirect(loginRedirectURL(request));
      response.cookies.delete(ACCESS_TOKEN_COOKIE);
      response.cookies.delete(REFRESH_TOKEN_COOKIE);
      return response;
    }

    if (refreshed.status === "berhasil" && isLoginPage) {
      const response = NextResponse.redirect(new URL("/", request.url));
      for (const cookie of refreshed.setCookies) response.headers.append("set-cookie", cookie);
      return response;
    }

    return refreshed.response;
  }

  if (hasSession && isLoginPage) {
    return NextResponse.redirect(new URL("/", request.url));
  }

  /*
   * Menu yang tidak dimiliki pengguna disembunyikan dari sidebar, tapi
   * alamatnya masih bisa diketik langsung. Halamannya lalu tetap dirender:
   * datanya ditolak backend, dan yang terbaca di layar adalah "Belum ada
   * customer" lengkap dengan tombol Tambah Customer — padahal customernya
   * ratusan dan tombolnya pasti gagal.
   *
   * Ini bukan lapisan keamanan; backend tetap menolak endpoint-nya sendiri.
   * Yang dijaga di sini adalah supaya orang tidak mendarat di halaman yang
   * berbohong tentang isi tokonya.
   */
  const granted = grantedPermissions(request);
  if (hasSession && granted && !canOpenPath(pathname, granted)) {
    return NextResponse.redirect(new URL("/", request.url));
  }

  return NextResponse.next();
}

/**
 * Hak akses efektif yang dibawa access token, atau null bila tidak terbaca.
 *
 * Isinya hanya dibaca, tidak diverifikasi — sama seperti pembacaan tanggal
 * kedaluwarsa di atas. Null berarti penjagaan rute dilewati sepenuhnya:
 * menghalangi orang gara-gara token yang tidak terbaca jauh lebih merugikan
 * daripada membiarkannya lewat, sebab backend tetap yang memutuskan.
 */
function grantedPermissions(request: NextRequest): string[] | null {
  const token = request.cookies.get(ACCESS_TOKEN_COOKIE)?.value;
  const payload = token?.split(".")[1];
  if (!payload) return null;

  try {
    const claims = JSON.parse(atob(payload.replace(/-/g, "+").replace(/_/g, "/"))) as {
      perms?: unknown;
    };
    if (!Array.isArray(claims.perms)) return null;
    return claims.perms.filter((p): p is string => typeof p === "string");
  } catch {
    return null;
  }
}

/** Halaman login beserta tujuan awal, supaya setelah masuk kembali ke sana. */
function loginRedirectURL(request: NextRequest): URL {
  const { pathname, search } = request.nextUrl;
  const loginURL = new URL("/login", request.url);
  if (pathname !== "/" && pathname !== "/login") {
    loginURL.searchParams.set("next", pathname + search);
  }
  return loginURL;
}

/** Benar bila access token hilang, cacat, atau sebentar lagi kedaluwarsa. */
function isAccessTokenStale(request: NextRequest): boolean {
  const token = request.cookies.get(ACCESS_TOKEN_COOKIE)?.value;
  if (!token) return true;

  // Isi token hanya dibaca, tidak diverifikasi — keabsahan tanda tangannya
  // tetap urusan backend pada tiap request. Di sini yang dibutuhkan sekadar
  // tanggal kedaluwarsanya.
  const [, payload] = token.split(".");
  if (!payload) return true;

  try {
    const claims = JSON.parse(atob(payload.replace(/-/g, "+").replace(/_/g, "/"))) as {
      exp?: number;
    };
    if (typeof claims.exp !== "number") return true;
    return claims.exp - EXPIRY_SKEW_SECONDS <= Math.floor(Date.now() / 1000);
  } catch {
    return true;
  }
}

type RefreshResult =
  | { status: "berhasil"; response: NextResponse; setCookies: string[] }
  | { status: "gagal" }
  | { status: "tertunda"; response: NextResponse };

/**
 * Menukar refresh token lewat route handler sendiri, lalu meneruskan cookie
 * barunya ke browser sekaligus ke request yang sedang dilayani.
 *
 * Pembaruan sengaja dititipkan ke `/api/auth/refresh`, bukan memanggil backend
 * langsung dari sini: middleware dibundel untuk edge runtime dan nilai
 * `process.env` di dalamnya ikut terkunci saat build, sehingga alamat backend
 * di server production belum tentu sama dengan yang tertanam di image.
 */
async function refreshSession(request: NextRequest): Promise<RefreshResult> {
  let refreshed: Response;
  try {
    refreshed = await fetch(new URL("/api/auth/refresh", request.url), {
      method: "POST",
      headers: { cookie: request.headers.get("cookie") ?? "" },
      cache: "no-store",
    });
  } catch {
    // Backend sedang tidak bisa dihubungi; biarkan halaman mencoba sendiri
    // daripada memaksa keluar dari sesi yang sebenarnya masih sah.
    return { status: "tertunda", response: NextResponse.next() };
  }

  if (!refreshed.ok) return { status: "gagal" };

  const setCookies = refreshed.headers.getSetCookie();

  // Request yang sedang berjalan ikut dibetulkan supaya layout membaca token
  // yang baru, bukan yang barusan kedaluwarsa — tanpa ini halaman pertama
  // setelah pembaruan tetap gagal dan pengguna tetap terlempar ke login.
  const newAccessToken = readCookieValue(setCookies, ACCESS_TOKEN_COOKIE);
  if (newAccessToken) request.cookies.set(ACCESS_TOKEN_COOKIE, newAccessToken);

  const response = NextResponse.next({ request: { headers: request.headers } });
  for (const cookie of setCookies) response.headers.append("set-cookie", cookie);
  return { status: "berhasil", response, setCookies };
}

function readCookieValue(setCookies: string[], name: string): string | undefined {
  const header = setCookies.find((cookie) => cookie.startsWith(`${name}=`));
  return header?.slice(name.length + 1).split(";")[0];
}

export const config = {
  matcher: [
    /*
     * Semua rute halaman kecuali:
     *   api        - route handler punya penjagaannya sendiri
     *   _next/*    - aset build Next.js
     *   berkas statis di public/ (gambar, robots.txt, ikon, dan sejenisnya)
     *
     * robots.txt harus tetap bisa diakses tanpa login; kalau ikut diarahkan ke
     * halaman login, crawler justru tidak pernah membaca aturan larangannya.
     */
    "/((?!api|_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico|txt|xml|json|woff2?)$).*)",
  ],
};
