"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import ProtectedRoute from "@/components/ProtectedRoute";
import { auditApi } from "@/lib/api";
import type { StoreApp } from "@/types";

export default function NewReportPage() {
  const router = useRouter();
  const [clientName, setClientName] = useState("");
  const [query, setQuery] = useState("");
  const [mode, setMode] = useState<"quick" | "full">("quick");
  const [playApps, setPlayApps] = useState<StoreApp[]>([]);
  const [appStoreApps, setAppStoreApps] = useState<StoreApp[]>([]);
  const [playApp, setPlayApp] = useState<StoreApp | null>(null);
  const [appStoreApp, setAppStoreApp] = useState<StoreApp | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim()) return;
    setBusy(true);
    setError("");
    try {
      const res = await auditApi.search(query.trim());
      setPlayApps(res.play);
      setAppStoreApps(res.appstore);
      setPlayApp(res.play[0] ?? null);
      setAppStoreApp(res.appstore[0] ?? null);
      if (!clientName.trim() && res.play[0]) setClientName(res.play[0].app_name);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Search failed");
    } finally {
      setBusy(false);
    }
  }

  async function handleStart(e: React.FormEvent) {
    e.preventDefault();
    if (!clientName.trim()) return;
    if (!playApp && !appStoreApp) {
      setError("Select at least one store app");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const audit = await auditApi.create({
        client_name: clientName.trim(),
        app_display_name: query.trim() || playApp?.app_name || appStoreApp?.app_name || clientName.trim(),
        play_app_id: playApp?.app_id,
        appstore_app_id: appStoreApp?.app_id,
        mode,
      });
      router.push(`/reports/${audit.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not start audit");
    } finally {
      setBusy(false);
    }
  }

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-3xl px-6 py-10">
        <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">new audit</p>
        <h1 className="mt-2 text-2xl font-medium text-foreground">Client store intelligence</h1>
        <p className="mt-2 text-sm text-muted">
          Search Play Store and App Store, crawl reviews, classify with MLC, and generate a consulting report.
        </p>

        <form onSubmit={(e) => void handleSearch(e)} className="mt-8 space-y-4 rounded-xl border border-border bg-surface p-5">
          <label className="block text-sm">
            App search (e.g. Avva)
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2"
              placeholder="Avva"
              required
            />
          </label>
          <button
            type="submit"
            disabled={busy}
            className="rounded-lg border border-accent px-4 py-2 text-sm text-accent disabled:opacity-50"
          >
            Search stores
          </button>
        </form>

        {(playApps.length > 0 || appStoreApps.length > 0) && (
          <form onSubmit={(e) => void handleStart(e)} className="mt-6 space-y-4 rounded-xl border border-border bg-surface p-5">
            <label className="block text-sm">
              Client name (for report cover)
              <input
                value={clientName}
                onChange={(e) => setClientName(e.target.value)}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2"
                required
              />
            </label>

            {playApps.length > 0 && (
              <label className="block text-sm">
                Google Play app
                <select
                  value={playApp?.app_id ?? ""}
                  onChange={(e) =>
                    setPlayApp(playApps.find((a) => a.app_id === e.target.value) ?? null)
                  }
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2"
                >
                  {playApps.map((a) => (
                    <option key={a.app_id} value={a.app_id}>
                      {a.app_name} — {a.developer}
                    </option>
                  ))}
                </select>
              </label>
            )}

            {appStoreApps.length > 0 && (
              <label className="block text-sm">
                App Store app
                <select
                  value={appStoreApp?.app_id ?? ""}
                  onChange={(e) =>
                    setAppStoreApp(appStoreApps.find((a) => a.app_id === e.target.value) ?? null)
                  }
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2"
                >
                  {appStoreApps.map((a) => (
                    <option key={a.app_id} value={a.app_id}>
                      {a.app_name} — {a.developer}
                    </option>
                  ))}
                </select>
              </label>
            )}

            <label className="block text-sm">
              Mode
              <select
                value={mode}
                onChange={(e) => setMode(e.target.value as "quick" | "full")}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2"
              >
                <option value="quick">Quick — up to 500 reviews (~3–5 min)</option>
                <option value="full">Full — up to 10,000 reviews (~25–30 min)</option>
              </select>
            </label>

            {error && <p className="text-sm text-red-500">{error}</p>}

            <button
              type="submit"
              disabled={busy}
              className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-accent-fg disabled:opacity-50"
            >
              Start audit & analyze
            </button>
          </form>
        )}
      </div>
    </ProtectedRoute>
  );
}
