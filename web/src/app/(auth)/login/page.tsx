import { Suspense } from "react";
import type { Metadata } from "next";

import { Logo } from "@/components/layout/logo";

import { LoginForm } from "./login-form";

export const metadata: Metadata = { title: "Masuk" };

export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-muted px-4 py-12">
      <div className="w-full max-w-sm space-y-6">
        <div className="flex flex-col items-center">
          <Logo variant="full" size={128} priority />
        </div>

        {/* useSearchParams butuh batas Suspense saat halaman dirender statis. */}
        <Suspense fallback={<div className="h-72 rounded-xl border border-border bg-card" />}>
          <LoginForm />
        </Suspense>
      </div>
    </main>
  );
}
