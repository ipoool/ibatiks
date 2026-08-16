"use client";

import { useSearchParams } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import type { Envelope, User } from "@/types/api";

/**
 * Membatasi tujuan setelah login pada rute di dalam aplikasi ini.
 *
 * Nilainya datang dari query string, jadi tanpa penyaringan siapa pun bisa
 * mengirim tautan `/login?next=https://situs-lain` dan memakai halaman login
 * yang tampak sah sebagai batu loncatan. Diawali "//" pun ditolak karena
 * browser membacanya sebagai alamat host lain.
 */
function safeNextPath(raw: string | null): string {
  if (!raw || !raw.startsWith("/") || raw.startsWith("//")) return "/";
  return raw;
}

export function LoginForm() {
  const searchParams = useSearchParams();
  const nextPath = safeNextPath(searchParams.get("next"));

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setError(null);
    setLoading(true);

    try {
      // Login lewat route handler Next, bukan langsung ke backend: token
      // ditulis ke cookie httpOnly di sisi server.
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const payload = (await response.json()) as Envelope<{ user: User }>;

      if (!response.ok) {
        setError(payload.error?.message ?? "Gagal masuk. Coba lagi.");
        return;
      }

      /*
       * Sengaja memuat ulang halaman penuh, bukan `router.replace()` +
       * `router.refresh()`: pasangan itu membuat navigasi batal di tengah jalan
       * — cookie sesi sudah tertulis, tapi browser tetap tertinggal di halaman
       * login seolah-olah passwordnya salah. Login hanya terjadi sekali, jadi
       * satu kali muat penuh tidak ada ruginya, sekaligus menjamin layout
       * server membaca cookie yang baru.
       */
      window.location.replace(nextPath);
    } catch {
      setError("Tidak bisa menghubungi server. Cek koneksi lalu coba lagi.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card>
      <CardContent className="pt-5">
        <form onSubmit={handleSubmit} className="space-y-4">
          <Field label="Email" htmlFor="email" required>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="owner@ibatiks.id"
              autoComplete="username"
              required
              autoFocus
            />
          </Field>

          <Field label="Password" htmlFor="password" required>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="••••••••"
              autoComplete="current-password"
              required
            />
          </Field>

          {error && (
            <p className="rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
              {error}
            </p>
          )}

          <Button type="submit" className="w-full" loading={loading}>
            Masuk
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
