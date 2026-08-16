"use client";

import { createContext, useContext } from "react";

import type { User, UserRole } from "@/types/api";

const CurrentUserContext = createContext<User | null>(null);

/**
 * Menyediakan identitas pengguna yang sudah diambil layout server ke seluruh
 * halaman di bawahnya.
 *
 * Dipakai untuk menyembunyikan bagian layar yang backend-nya akan menolak
 * request dari role tersebut. Ini murni soal pengalaman pemakaian, bukan
 * keamanan: penjagaan yang sesungguhnya tetap ada di backend.
 */
export function CurrentUserProvider({
  user,
  children,
}: {
  user: User;
  children: React.ReactNode;
}) {
  return <CurrentUserContext.Provider value={user}>{children}</CurrentUserContext.Provider>;
}

export function useCurrentUser(): User | null {
  return useContext(CurrentUserContext);
}

/** True kalau pengguna saat ini punya salah satu role yang diizinkan. */
export function useHasRole(...roles: UserRole[]): boolean {
  const user = useContext(CurrentUserContext);
  return user ? roles.includes(user.role) : false;
}
