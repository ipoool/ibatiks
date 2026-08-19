"use client";

import { LogOut, Menu, PanelLeftClose, PanelLeftOpen, X } from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";

import { FooterCredit } from "@/components/layout/footer-credit";
import { Logo } from "@/components/layout/logo";
import { visibleSections } from "@/components/layout/nav";
import { NavMenu } from "@/components/layout/nav-menu";
import { CurrentUserProvider } from "@/components/layout/user-context";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/form-dialog";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { SIDEBAR_COOKIE } from "@/lib/sidebar";
import { cn, initials } from "@/lib/utils";
import type { User } from "@/types/api";

const ROLE_LABEL: Record<User["role"], string> = {
  owner: "Owner",
  admin: "Admin",
  tripper: "Tripper",
};

export function AppShell({
  user,
  children,
  defaultCollapsed = false,
}: {
  user: User;
  children: React.ReactNode;
  defaultCollapsed?: boolean;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(defaultCollapsed);
  const [confirmLogout, setConfirmLogout] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);

  const sections = visibleSections(user);

  async function handleLogout() {
    setLoggingOut(true);
    try {
      await fetch("/api/auth/logout", { method: "POST" });
      router.replace("/login");
      router.refresh();
    } finally {
      // Dikembalikan juga saat gagal, supaya tombolnya tidak tertinggal dalam
      // keadaan memuat dan pengguna bisa mencoba lagi.
      setLoggingOut(false);
    }
  }

  // Menekan menu apa pun berarti pengguna memang ingin pindah halaman, jadi
  // laci di layar kecil sekalian ditutup.
  const nav = (
    <NavMenu sections={sections} pathname={pathname} onNavigate={() => setMobileOpen(false)} />
  );

  function toggleSidebar() {
    const next = !collapsed;
    setCollapsed(next);
    // Setahun, dan `SameSite=Lax` supaya ikut terkirim pada navigasi biasa —
    // ini sekadar preferensi tampilan, tidak ada isi rahasia di dalamnya.
    document.cookie = `${SIDEBAR_COOKIE}=${next ? "1" : "0"}; path=/; max-age=31536000; samesite=lax`;
  }

  return (
    <div className="flex min-h-screen">
      {/*
        Sidebar tetap di layar lebar.

        Tingginya dipatok setinggi layar dan dibuat sticky, bukan dibiarkan
        mengikuti tinggi halaman: tanpa itu, pada halaman panjang seperti
        Dashboard sidebar ikut memanjang sehingga panel akun dan tombol Keluar
        terdorong ke dasar dokumen dan baru terlihat setelah menggulir jauh.
        Daftar menunya sendiri yang menggulir di dalam, lewat `overflow-y-auto`
        pada <nav>.
      */}
      <aside
        className={cn(
          "hidden shrink-0 flex-col border-r border-border bg-card transition-[width] duration-200 lg:sticky lg:top-0 lg:flex lg:h-screen",
          collapsed ? "w-16" : "w-64",
        )}
      >
        <div
          className={cn(
            "flex items-center border-b border-border",
            // Saat menciut, logo dan tombol tidak muat berdampingan di lebar
            // 64 piksel, jadi ditumpuk dan tinggi headernya dibiarkan mengikuti
            // isinya alih-alih dipatok setinggi header halaman.
            collapsed ? "flex-col gap-1 px-1 py-2" : "h-16 gap-2 px-5",
          )}
        >
          {collapsed ? (
            <Logo size={24} priority />
          ) : (
            <>
              <Logo size={28} priority />
              <span className="flex-1 truncate text-lg font-semibold tracking-tight">Ibatiks</span>
            </>
          )}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon-sm" onClick={toggleSidebar}>
                {collapsed ? <PanelLeftOpen /> : <PanelLeftClose />}
                <span className="sr-only">{collapsed ? "Lebarkan menu" : "Ciutkan menu"}</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">
              {collapsed ? "Lebarkan menu" : "Ciutkan menu"}
            </TooltipContent>
          </Tooltip>
        </div>
        <NavMenu
          sections={sections}
          pathname={pathname}
          onNavigate={() => setMobileOpen(false)}
          collapsed={collapsed}
        />
        <UserPanel user={user} collapsed={collapsed} onLogout={() => setConfirmLogout(true)} />
      </aside>

      {/* Sidebar geser di layar kecil */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <button
            type="button"
            aria-label="Tutup menu"
            className="absolute inset-0 bg-black/40"
            onClick={() => setMobileOpen(false)}
          />
          <aside className="relative flex h-full w-64 flex-col border-r border-border bg-card">
            <div className="flex h-16 items-center justify-between border-b border-border px-5">
              <span className="flex items-center gap-2 text-lg font-semibold tracking-tight">
                <Logo size={28} />
                Ibatiks
              </span>
              <Button variant="ghost" size="icon-sm" onClick={() => setMobileOpen(false)}>
                <X />
                <span className="sr-only">Tutup menu</span>
              </Button>
            </div>
            {nav}
            <UserPanel user={user} onLogout={() => setConfirmLogout(true)} />
          </aside>
        </div>
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-border bg-card/95 px-4 backdrop-blur lg:hidden">
          <Button variant="ghost" size="icon-sm" onClick={() => setMobileOpen(true)}>
            <Menu />
            <span className="sr-only">Buka menu</span>
          </Button>
          <Logo size={24} />
          <span className="font-semibold">Ibatiks</span>
        </header>

        <main className="mx-auto w-full max-w-[1400px] flex-1 space-y-6 p-4 sm:p-6 lg:p-8">
          <CurrentUserProvider user={user}>{children}</CurrentUserProvider>
        </main>

        {/*
          Ditaruh setelah <main> yang flex-1, jadi ia terdorong ke dasar layar
          saat isi halamannya pendek dan mengikuti di bawah isi saat panjang —
          bukan melayang menutupi konten seperti elemen yang dipatok fixed.
        */}
        <footer className="mx-auto w-full max-w-[1400px] px-4 pb-4 sm:px-6 lg:px-8">
          <FooterCredit className="text-right" />
        </footer>
      </div>

      {/*
        Dialog dirender sekali di sini, bukan di dalam UserPanel: panel itu
        muncul dua kali (sidebar desktop dan laci mobile), sehingga dialognya
        akan ikut tergandakan kalau ditaruh di dalamnya.
      */}
      <ConfirmDialog
        open={confirmLogout}
        onOpenChange={setConfirmLogout}
        title="Keluar dari Ibatiks?"
        description={`Kamu akan keluar dari akun ${user.name}. Pekerjaan yang belum disimpan pada form yang sedang terbuka akan hilang.`}
        confirmLabel="Ya, keluar"
        destructive={false}
        loading={loggingOut}
        onConfirm={handleLogout}
      />
    </div>
  );
}

function UserPanel({
  user,
  onLogout,
  collapsed = false,
}: {
  user: User;
  onLogout: () => void;
  collapsed?: boolean;
}) {
  if (collapsed) {
    return (
      <div className="mt-auto flex flex-col items-center gap-2 border-t border-border p-2">
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-semibold text-primary-foreground">
              {initials(user.name)}
            </div>
          </TooltipTrigger>
          <TooltipContent side="right">
            {user.name} · {ROLE_LABEL[user.role]}
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon-sm"
              className="text-destructive hover:bg-destructive/5 hover:text-destructive"
              onClick={onLogout}
            >
              <LogOut />
              <span className="sr-only">Keluar</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent side="right">Keluar</TooltipContent>
        </Tooltip>
      </div>
    );
  }

  return (
    // mt-auto mendorong panel ini ke dasar sidebar walau daftar menunya pendek,
    // sehingga tombol keluar selalu berada di tempat yang sama.
    <div className="mt-auto space-y-2 border-t border-border p-3">
      <div className="flex items-center gap-3 rounded-lg px-2 py-1">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-semibold text-primary-foreground">
          {initials(user.name)}
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{user.name}</p>
          <p className="truncate text-xs text-muted-foreground">{ROLE_LABEL[user.role]}</p>
        </div>
      </div>

      {/*
        Tombol keluar dibuat bergaris dan berwarna merah, bukan teks abu-abu
        samar seperti sebelumnya: keluar adalah aksi yang dicari orang ketika
        selesai bekerja, dan tombol yang menyaru dengan latar membuatnya sulit
        ditemukan justru saat dibutuhkan.
      */}
      <Button
        variant="outline"
        className="w-full justify-center text-destructive hover:bg-destructive/5 hover:text-destructive"
        onClick={onLogout}
      >
        <LogOut />
        Keluar
      </Button>
    </div>
  );
}
