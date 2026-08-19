"use client";

import { Plus, Trash2 } from "lucide-react";

import { OptionSelect } from "@/components/filter-select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Social, SocialPlatform } from "@/types/api";

/**
 * Platform yang bisa dipilih, beserta contoh penulisan handle-nya.
 *
 * Daftar tertutup, bukan isian bebas: salah ketik satu huruf pada nama platform
 * membuat labelnya meleset sementara tulisannya tetap terlihat wajar sekilas —
 * persis alasan mata uang trip juga dipilih dari daftar.
 */
export const SOCIAL_PLATFORMS: ReadonlyArray<{
  value: SocialPlatform;
  label: string;
  contoh: string;
}> = [
  { value: "instagram", label: "Instagram", contoh: "@username" },
  { value: "tiktok", label: "TikTok", contoh: "@username" },
  { value: "facebook", label: "Facebook", contoh: "nama profil atau tautan" },
  { value: "lainnya", label: "Lainnya", contoh: "tulis nama platform dan akunnya" },
];

export const SOCIAL_PLATFORM_OPTIONS = SOCIAL_PLATFORMS.map(({ value, label }) => ({
  value,
  label,
}));

export function labelPlatform(platform: string): string {
  return SOCIAL_PLATFORMS.find((p) => p.value === platform)?.label ?? platform;
}

/** Platform yang belum terpakai, supaya baris baru tidak langsung kembar. */
function platformBerikutnya(terpakai: readonly Social[]): SocialPlatform {
  const adaYangPakai = (value: SocialPlatform) => terpakai.some((s) => s.platform === value);
  return SOCIAL_PLATFORMS.find((p) => !adaYangPakai(p.value))?.value ?? "lainnya";
}

/**
 * Daftar akun media sosial customer yang bisa ditambah dan dikurangi.
 *
 * Satu customer sering dihubungi lewat lebih dari satu tempat: Instagram untuk
 * melihat katalog, TikTok tempat ia menemukan toko, Telegram untuk mengirim
 * foto struk. Sebelumnya hanya Instagram yang punya kolom, dan sisanya terselip
 * di catatan tempat tidak ada yang mencarinya.
 */
export function SocialFields({
  value,
  onChange,
}: {
  value: Social[];
  onChange: (next: Social[]) => void;
}) {
  function ubah(index: number, patch: Partial<Social>) {
    onChange(value.map((akun, i) => (i === index ? { ...akun, ...patch } : akun)));
  }

  return (
    <div className="space-y-2">
      {value.map((akun, index) => (
        <div key={index} className="flex items-center gap-2">
          <OptionSelect
            value={akun.platform}
            onChange={(platform) => ubah(index, { platform: platform as SocialPlatform })}
            options={SOCIAL_PLATFORM_OPTIONS}
            className="w-36 shrink-0"
          />
          <Input
            value={akun.handle}
            onChange={(event) => ubah(index, { handle: event.target.value })}
            placeholder={SOCIAL_PLATFORMS.find((p) => p.value === akun.platform)?.contoh}
            className="min-w-0 flex-1"
          />
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            tooltip="Hapus akun"
            className="shrink-0 text-destructive hover:text-destructive"
            onClick={() => onChange(value.filter((_, i) => i !== index))}
          >
            <Trash2 />
            <span className="sr-only">Hapus akun</span>
          </Button>
        </div>
      ))}

      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={() =>
          onChange([...value, { platform: platformBerikutnya(value), handle: "" }])
        }
      >
        <Plus />
        Tambah akun
      </Button>
    </div>
  );
}
