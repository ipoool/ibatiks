import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // standalone menghasilkan bundel server minimal berisi hanya dependensi yang
  // benar-benar dipakai, sehingga image Docker jauh lebih kecil dan tidak perlu
  // menyalin seluruh node_modules.
  output: "standalone",

  // Foto produk umumnya berupa tautan dari toko/marketplace, jadi optimizer
  // gambar Next dimatikan agar tidak perlu mendaftarkan setiap host asal.
  images: {
    unoptimized: true,
  },

  // Header keamanan dasar. Aplikasi ini back office internal, jadi tidak boleh
  // di-embed situs lain dan tidak perlu mengirim referrer ke luar.
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "same-origin" },
          { key: "X-DNS-Prefetch-Control", value: "off" },
        ],
      },
    ];
  },
};

export default nextConfig;
