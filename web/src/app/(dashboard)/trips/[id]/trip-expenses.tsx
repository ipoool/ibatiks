"use client";

import { Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { OptionSelect, toOptions } from "@/components/filter-select";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { ConfirmDialog, FormDialog } from "@/components/ui/form-dialog";
import { FileUpload } from "@/components/ui/file-upload";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { ErrorState } from "@/components/ui/page";
import { DataTable, TD, TH, TR } from "@/components/data-table";
import {
  useDeleteTripExpense,
  useSaveTripExpense,
  useTripExpenses,
  type TripExpensePayload,
} from "@/hooks/use-trips";
import { formatDate, formatIDR, toDateInput, todayInput, toNumber } from "@/lib/utils";
import type { ExpenseCategory, Trip, TripExpense } from "@/types/api";

const CATEGORY_LABEL: Record<ExpenseCategory, string> = {
  tiket: "Tiket pesawat",
  bagasi: "Bagasi",
  akomodasi: "Akomodasi",
  transport: "Transport lokal",
  visa: "Visa",
  lainnya: "Lainnya",
};

function emptyForm(): TripExpensePayload {
  return {
    category: "tiket",
    description: "",
    amount: "",
    spent_at: todayInput(),
    receipt_url: "",
  };
}

export function TripExpenses({ trip }: { trip: Trip }) {
  const { data: expenses, isLoading, error } = useTripExpenses(trip.id);

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<TripExpense | null>(null);
  const [form, setForm] = useState<TripExpensePayload>(emptyForm);
  const [deleting, setDeleting] = useState<TripExpense | null>(null);

  const save = useSaveTripExpense(trip.id, editing?.id);
  const remove = useDeleteTripExpense(trip.id);

  const total = expenses?.reduce((sum, expense) => sum + toNumber(expense.amount), 0) ?? 0;

  function openCreate() {
    setEditing(null);
    setForm(emptyForm());
    save.reset();
    setFormOpen(true);
  }

  function openEdit(expense: TripExpense) {
    setEditing(expense);
    setForm({
      category: expense.category,
      description: expense.description,
      amount: expense.amount,
      spent_at: toDateInput(expense.spent_at),
      receipt_url: expense.receipt_url ?? "",
    });
    save.reset();
    setFormOpen(true);
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    save.mutate(
      { ...form, receipt_url: form.receipt_url || null },
      {
        onSuccess: () => {
          toast.success(editing ? "Biaya diperbarui" : "Biaya dicatat");
          setFormOpen(false);
        },
      },
    );
  }

  function handleDelete() {
    if (!deleting) return;
    remove.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Biaya dihapus");
        setDeleting(null);
      },
    });
  }

  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="space-y-1">
          <h2 className="text-lg font-semibold">Biaya perjalanan</h2>
          <p className="text-sm text-muted-foreground">
            Modal di luar harga barang. Nilainya dikurangkan dari laba trip.
          </p>
        </div>
        <Button onClick={openCreate}>
          <Plus />
          Catat Biaya
        </Button>
      </div>

      <ErrorState error={error} />

      <Card className="bg-muted/50">
        <CardContent className="flex items-center justify-between py-4">
          <span className="text-sm text-muted-foreground">Total biaya perjalanan</span>
          <span className="tabular text-xl font-semibold">{formatIDR(total)}</span>
        </CardContent>
      </Card>

      <DataTable
        columns={5}
        isLoading={isLoading}
        isEmpty={!isLoading && (expenses?.length ?? 0) === 0}
        emptyTitle="Belum ada biaya tercatat"
        emptyDescription="Catat tiket, bagasi, akomodasi, dan transport agar laba trip mencerminkan modal sebenarnya."
        emptyAction={
          <Button onClick={openCreate}>
            <Plus />
            Catat Biaya
          </Button>
        }
        head={
          <TR>
            <TH className="hidden sm:table-cell">Tanggal</TH>
            <TH>Kategori</TH>
            <TH className="hidden lg:table-cell">Keterangan</TH>
            <TH className="text-right">Nominal</TH>
            <TH className="text-right">Aksi</TH>
          </TR>
        }
      >
        {expenses?.map((expense) => (
          <TR key={expense.id}>
            <TD className="hidden whitespace-nowrap text-sm sm:table-cell">
              {formatDate(expense.spent_at)}
            </TD>
            <TD className="text-sm">
              {CATEGORY_LABEL[expense.category]}
              {/* Tanggal dan keterangan menyusul kategori saat kolomnya
                  disembunyikan; satu kategori bisa dicatat berkali-kali. */}
              <p className="text-xs text-muted-foreground sm:hidden">
                {formatDate(expense.spent_at)}
              </p>
              <p className="text-xs text-muted-foreground lg:hidden">{expense.description}</p>
            </TD>
            <TD className="hidden lg:table-cell">
              <p className="text-sm">{expense.description}</p>
              {expense.receipt_url && (
                <a
                  href={expense.receipt_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs text-muted-foreground hover:underline"
                >
                  Lihat bukti
                </a>
              )}
            </TD>
            <TD className="tabular text-right font-medium">{formatIDR(expense.amount)}</TD>
            <TD>
              <div className="flex justify-end gap-1">
                <Button variant="ghost" size="icon-sm" onClick={() => openEdit(expense)} tooltip="Ubah">
                  <Pencil />
                  <span className="sr-only">Ubah</span>
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  tooltip="Hapus"
                  className="text-destructive hover:text-destructive"
                  onClick={() => setDeleting(expense)}
                >
                  <Trash2 />
                  <span className="sr-only">Hapus</span>
                </Button>
              </div>
            </TD>
          </TR>
        ))}
      </DataTable>

      <FormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        title={editing ? "Ubah Biaya" : "Catat Biaya Perjalanan"}
        error={save.error}
        loading={save.isPending}
        onSubmit={handleSubmit}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Kategori" htmlFor="category" required>
            <OptionSelect
              id="category"
              value={form.category}
              onChange={(value) => setForm({ ...form, category: value })}
              options={toOptions(CATEGORY_LABEL)}
            />
          </Field>

          <Field label="Tanggal" htmlFor="spent_at" required>
            <Input
              id="spent_at"
              type="date"
              value={form.spent_at ?? ""}
              onChange={(event) => setForm({ ...form, spent_at: event.target.value })}
              required
            />
          </Field>

          <Field label="Keterangan" htmlFor="description" required className="sm:col-span-2">
            <Input
              id="description"
              value={form.description}
              onChange={(event) => setForm({ ...form, description: event.target.value })}
              placeholder="Extra baggage 10kg"
              required
            />
          </Field>

          <Field label="Nominal (Rp)" htmlFor="amount" required>
            <Input
              id="amount"
              type="number"
              min="0"
              step="any"
              value={form.amount}
              onChange={(event) => setForm({ ...form, amount: event.target.value })}
              placeholder="850000"
              required
            />
          </Field>

          <Field label="Bukti pembayaran" htmlFor="receipt_url" hint="Foto atau PDF struk">
            <FileUpload
              value={form.receipt_url || null}
              onChange={(url) => setForm({ ...form, receipt_url: url ?? "" })}
            />
          </Field>
        </div>
      </FormDialog>

      <ConfirmDialog
        open={Boolean(deleting)}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Hapus biaya ini?"
        description={`${deleting?.description ?? ""} (${formatIDR(deleting?.amount ?? 0)}) akan dihapus dan laba trip dihitung ulang.`}
        confirmLabel="Hapus"
        loading={remove.isPending}
        error={remove.error}
        onConfirm={handleDelete}
      />
    </>
  );
}
