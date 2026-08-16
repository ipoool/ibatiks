import type { Envelope, PageMeta } from "@/types/api";

/** Error API yang membawa kode dan pesan per field dari backend. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly fields?: Record<string, string>;

  constructor(status: number, code: string, message: string, fields?: Record<string, string>) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.fields = fields;
  }

  /** True kalau errornya karena input pengguna, bukan kegagalan sistem. */
  get isValidation(): boolean {
    return this.code === "VALIDATION_ERROR";
  }

  /** True kalau operasi ditolak karena status data saat ini tidak mengizinkan. */
  get isConflict(): boolean {
    return this.code === "CONFLICT" || this.code === "INVALID_STATE";
  }
}

/** Seluruh panggilan browser lewat BFF proxy, bukan langsung ke backend. */
const BASE_PATH = "/api/proxy";

export interface Paginated<T> {
  items: T[];
  meta: PageMeta;
}

type QueryValue = string | number | boolean | null | undefined;

/** Menyusun query string, melewatkan nilai kosong supaya URL tetap bersih. */
export function buildQuery(params: Record<string, QueryValue>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") continue;
    search.set(key, String(value));
  }
  const query = search.toString();
  return query ? `?${query}` : "";
}

async function request<T>(path: string, init: RequestInit = {}): Promise<Envelope<T>> {
  let response: Response;
  try {
    response = await fetch(`${BASE_PATH}${path}`, {
      ...init,
      headers: {
        Accept: "application/json",
        ...(init.body ? { "Content-Type": "application/json" } : {}),
        ...init.headers,
      },
    });
  } catch {
    throw new ApiError(0, "NETWORK_ERROR", "Koneksi ke server terputus. Cek jaringan lalu coba lagi.");
  }

  if (response.status === 204) {
    return {} as Envelope<T>;
  }

  const raw = await response.text();
  let payload: Envelope<T> | undefined;
  if (raw) {
    try {
      payload = JSON.parse(raw) as Envelope<T>;
    } catch {
      payload = undefined;
    }
  }

  if (!response.ok) {
    const error = payload?.error;
    // Sesi habis: paksa kembali ke halaman login alih-alih menampilkan error
    // yang tidak bisa ditindaklanjuti pengguna.
    //
    // Sengaja memakai navigasi penuh (bukan router.push): memuat ulang halaman
    // membuang seluruh cache query dan state komponen milik sesi yang sudah
    // tidak berlaku, sehingga tidak ada sisa data pengguna sebelumnya.
    if (response.status === 401 && typeof window !== "undefined") {
      const loginURL = new URL("/login", window.location.origin);
      loginURL.searchParams.set("next", window.location.pathname);
      window.location.assign(loginURL.toString());
    }
    throw new ApiError(
      response.status,
      error?.code ?? "UNKNOWN_ERROR",
      error?.message ?? "Terjadi kesalahan yang tidak terduga.",
      error?.fields,
    );
  }

  return payload ?? ({} as Envelope<T>);
}

export const api = {
  async get<T>(path: string): Promise<T> {
    const payload = await request<T>(path, { method: "GET" });
    return payload.data;
  },

  /** Versi get untuk endpoint berhalaman: ikut mengembalikan metadata halaman. */
  async list<T>(path: string): Promise<Paginated<T>> {
    const payload = await request<T[]>(path, { method: "GET" });
    return {
      items: payload.data ?? [],
      meta: payload.meta ?? { page: 1, per_page: 20, total: 0, total_pages: 0 },
    };
  },

  async post<T>(path: string, body?: unknown): Promise<T> {
    const payload = await request<T>(path, {
      method: "POST",
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    return payload.data;
  },

  async put<T>(path: string, body?: unknown): Promise<T> {
    const payload = await request<T>(path, {
      method: "PUT",
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    return payload.data;
  },

  async patch<T>(path: string, body?: unknown): Promise<T> {
    const payload = await request<T>(path, {
      method: "PATCH",
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    return payload.data;
  },

  async delete(path: string): Promise<void> {
    await request<void>(path, { method: "DELETE" });
  },

  /** URL absolut untuk unduhan (PDF invoice, ekspor CSV). */
  downloadURL(path: string): string {
    return `${BASE_PATH}${path}`;
  },

  /** Unggah berkas bukti transfer atau foto struk. */
  async upload(file: File): Promise<{ url: string }> {
    const form = new FormData();
    form.append("file", file);

    const response = await fetch(`${BASE_PATH}/uploads`, { method: "POST", body: form });
    const payload = (await response.json()) as Envelope<{ url: string }>;

    if (!response.ok) {
      throw new ApiError(
        response.status,
        payload.error?.code ?? "UPLOAD_FAILED",
        payload.error?.message ?? "Gagal mengunggah berkas.",
        payload.error?.fields,
      );
    }
    return payload.data;
  },
};
