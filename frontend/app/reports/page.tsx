"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { auditApi } from "@/lib/api";
import type { Audit } from "@/types";

export default function ReportsPage() {
  const [audits, setAudits] = useState<Audit[]>([]);
  const [error, setError] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);

  function load() {
    auditApi
      .list()
      .then(setAudits)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load"));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleDelete(id: string, name: string) {
    if (!confirm(`"${name}" analizini silmek istediğinize emin misiniz?`)) return;
    setDeletingId(id);
    setError("");
    try {
      await auditApi.remove(id);
      setAudits((prev) => prev.filter((a) => a.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Silinemedi");
    } finally {
      setDeletingId(null);
    }
  }

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-5xl px-6 py-10">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">client audits</p>
            <h1 className="mt-2 text-2xl font-medium text-foreground">Reports</h1>
          </div>
          <Link
            href="/reports/new"
            className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-accent-fg"
          >
            New audit
          </Link>
        </div>

        {error && (
          <p className="mb-4 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-500">
            {error}
          </p>
        )}

        <div className="space-y-3">
          {audits.map((a) => (
            <div
              key={a.id}
              className="flex items-center justify-between gap-4 rounded-2xl border border-border bg-surface p-4 transition hover:border-accent"
            >
              <Link href={`/reports/${a.id}`} className="min-w-0 flex-1">
                <p className="font-medium text-foreground">{a.client_name}</p>
                <p className="text-sm text-muted">
                  {a.app_display_name}
                  {a.country ? ` · ${a.country.toUpperCase()}` : ""} · {a.mode} · {a.total_reviews} reviews
                </p>
              </Link>
              <div className="flex shrink-0 items-center gap-3">
                <span className="font-mono text-xs uppercase text-accent">{a.status}</span>
                <button
                  type="button"
                  disabled={deletingId === a.id}
                  onClick={() => void handleDelete(a.id, a.client_name)}
                  className="rounded-lg border border-red-500/30 px-3 py-1.5 text-xs text-red-500 transition hover:bg-red-500/10 disabled:opacity-50"
                >
                  {deletingId === a.id ? "…" : "Sil"}
                </button>
              </div>
            </div>
          ))}
          {audits.length === 0 && !error && (
            <div className="rounded-2xl border border-dashed border-border p-8 text-center text-sm text-muted">
              Henüz analiz yok. Yeni audit ile başlayın.
            </div>
          )}
        </div>
      </div>
    </ProtectedRoute>
  );
}
