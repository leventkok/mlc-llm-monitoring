"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { auditApi } from "@/lib/api";
import type { Audit } from "@/types";

export default function ReportsPage() {
  const [audits, setAudits] = useState<Audit[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    auditApi
      .list()
      .then(setAudits)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load"));
  }, []);

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-5xl px-6 py-10">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
              client audits
            </p>
            <h1 className="mt-2 text-2xl font-medium text-foreground">Reports</h1>
          </div>
          <Link
            href="/reports/new"
            className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-accent-fg"
          >
            New audit
          </Link>
        </div>

        {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

        <div className="space-y-3">
          {audits.map((a) => (
            <Link
              key={a.id}
              href={`/reports/${a.id}`}
              className="block rounded-xl border border-border bg-surface p-4 transition hover:border-accent"
            >
              <div className="flex items-center justify-between gap-4">
                <div>
                  <p className="font-medium text-foreground">{a.client_name}</p>
                  <p className="text-sm text-muted">
                    {a.app_display_name} · {a.mode} · {a.total_reviews} reviews
                  </p>
                </div>
                <span className="font-mono text-xs uppercase text-accent">{a.status}</span>
              </div>
            </Link>
          ))}
          {audits.length === 0 && !error && (
            <p className="text-sm text-muted">No audits yet. Start with Avva or any client app.</p>
          )}
        </div>
      </div>
    </ProtectedRoute>
  );
}
