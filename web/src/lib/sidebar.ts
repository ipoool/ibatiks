/**
 * Nama cookie penyimpan lebar sidebar, dipakai bersama oleh layout server
 * (membaca) dan AppShell di klien (menulis).
 *
 * Diletakkan di modul tersendiri, bukan diekspor dari `shell.tsx`: berkas itu
 * ber-"use client", dan nilai yang diimpor komponen server dari modul klien
 * bukan lagi string biasa melainkan rujukan modul — `cookies().get()` yang
 * menerimanya diam-diam mengembalikan undefined, sehingga sidebar selalu
 * terlukis lebar walau pengguna sudah menciutkannya.
 */
export const SIDEBAR_COOKIE = "jastipin_sidebar_collapsed";
