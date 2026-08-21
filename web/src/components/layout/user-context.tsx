"use client";

import { createContext, useContext } from "react";

import type { Permission, User } from "@/types/api";

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

/**
 * True kalau pengguna saat ini boleh membuka menu tersebut.
 *
 * Menggantikan pemeriksaan berbasis nama role. Sejak role jadi data, nama role
 * tidak lagi bisa jadi pegangan: role bikinan toko sendiri bukan owner, admin,
 * maupun tripper, jadi `role === "owner"` akan menyembunyikan bagian layar dari
 * orang yang sebenarnya berhak — dan sebaliknya menampilkannya untuk yang
 * menunya sudah dicabut.
 *
 * Dipakai untuk menyembunyikan bagian layar yang backend-nya akan menolak.
 * Penjagaan yang sesungguhnya tetap di backend.
 */
export function useHasPermission(permission: Permission): boolean {
  const user = useContext(CurrentUserContext);
  return user?.effective_permissions?.includes(permission) ?? false;
}
