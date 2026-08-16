"use client";

import { Check, Pencil, Plus, Trash2, X } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ErrorState } from "@/components/ui/page";
import { useCategories, useDeleteCategory, useSaveCategory } from "@/hooks/use-master";
import { ApiError } from "@/lib/api";

/**
 * Pengelolaan kategori dijadikan dialog kecil, bukan halaman tersendiri:
 * jumlahnya sedikit dan hampir selalu diubah sambil menata produk.
 */
export function CategoryManager({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { data: categories, isLoading } = useCategories();
  const [newName, setNewName] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState("");

  const create = useSaveCategory();
  const update = useSaveCategory(editingId ?? undefined);
  const remove = useDeleteCategory();

  function handleCreate(event: React.FormEvent) {
    event.preventDefault();
    if (!newName.trim()) return;

    create.mutate(
      { name: newName.trim() },
      {
        onSuccess: () => {
          setNewName("");
          toast.success("Kategori ditambahkan");
        },
      },
    );
  }

  function handleUpdate() {
    if (!editingId || !editingName.trim()) return;

    update.mutate(
      { name: editingName.trim() },
      {
        onSuccess: () => {
          setEditingId(null);
          toast.success("Kategori diperbarui");
        },
      },
    );
  }

  function handleDelete(id: string) {
    remove.mutate(id, {
      onSuccess: () => toast.success("Kategori dihapus"),
      onError: (error) => {
        toast.error(
          error instanceof ApiError
            ? error.message
            : "Gagal menghapus kategori",
        );
      },
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Kategori Produk</DialogTitle>
          <DialogDescription>
            Kategori memudahkan menyaring produk dan membaca laporan penjualan.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleCreate} className="flex gap-2">
          <Input
            value={newName}
            onChange={(event) => setNewName(event.target.value)}
            placeholder="Nama kategori baru"
          />
          <Button type="submit" loading={create.isPending} disabled={!newName.trim()}>
            <Plus />
            Tambah
          </Button>
        </form>

        <ErrorState error={create.error ?? update.error ?? remove.error} />

        <div className="max-h-72 space-y-1 overflow-y-auto">
          {isLoading && <p className="text-sm text-muted-foreground">Memuat…</p>}
          {!isLoading && (categories?.length ?? 0) === 0 && (
            <p className="py-4 text-center text-sm text-muted-foreground">Belum ada kategori.</p>
          )}

          {categories?.map((category) => (
            <div
              key={category.id}
              className="flex items-center gap-2 rounded-lg border border-border px-3 py-2"
            >
              {editingId === category.id ? (
                <>
                  <Input
                    value={editingName}
                    onChange={(event) => setEditingName(event.target.value)}
                    className="h-8"
                    autoFocus
                  />
                  <Button size="icon-sm" onClick={handleUpdate} loading={update.isPending}>
                    <Check />
                  </Button>
                  <Button variant="ghost" size="icon-sm" onClick={() => setEditingId(null)}>
                    <X />
                  </Button>
                </>
              ) : (
                <>
                  <span className="flex-1 text-sm font-medium">{category.name}</span>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => {
                      setEditingId(category.id);
                      setEditingName(category.name);
                    }}
                  >
                    <Pencil />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    className="text-destructive hover:text-destructive"
                    onClick={() => handleDelete(category.id)}
                  >
                    <Trash2 />
                  </Button>
                </>
              )}
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
