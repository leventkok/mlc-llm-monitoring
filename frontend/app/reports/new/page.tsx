"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import ProtectedRoute from "@/components/ProtectedRoute";
import { auditApi } from "@/lib/api";
import type { StoreApp } from "@/types";

const COUNTRIES = [
  { code: "tr", label: "Türkiye (TR)" },
  { code: "us", label: "United States (US)" },
  { code: "gb", label: "United Kingdom (GB)" },
  { code: "de", label: "Germany (DE)" },
  { code: "fr", label: "France (FR)" },
  { code: "es", label: "Spain (ES)" },
  { code: "it", label: "Italy (IT)" },
  { code: "nl", label: "Netherlands (NL)" },
  { code: "sa", label: "Saudi Arabia (SA)" },
  { code: "ae", label: "UAE (AE)" },
];

const inputClass =
  "mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-foreground outline-none transition focus:border-accent";

export default function NewReportPage() {
  const router = useRouter();
  const [clientName, setClientName] = useState("");
  const [query, setQuery] = useState("");
  const [country, setCountry] = useState("tr");
  const [mode, setMode] = useState<"quick" | "full">("full");
  const [includePlay, setIncludePlay] = useState(true);
  const [includeAppStore, setIncludeAppStore] = useState(true);
  const [playLimit, setPlayLimit] = useState(0);
  const [appStoreLimit, setAppStoreLimit] = useState(0);
  const [playApps, setPlayApps] = useState<StoreApp[]>([]);
  const [appStoreApps, setAppStoreApps] = useState<StoreApp[]>([]);
  const [playApp, setPlayApp] = useState<StoreApp | null>(null);
  const [appStoreApp, setAppStoreApp] = useState<StoreApp | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const modeCap = mode === "full" ? 10000 : 500;

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim()) return;
    setBusy(true);
    setError("");
    try {
      const res = await auditApi.search(query.trim(), country);
      setPlayApps(res.play);
      setAppStoreApps(res.appstore);
      setPlayApp(includePlay ? (res.play[0] ?? null) : null);
      setAppStoreApp(includeAppStore ? (res.appstore[0] ?? null) : null);
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
    const playId = includePlay ? playApp?.app_id : undefined;
    const appStoreId = includeAppStore ? appStoreApp?.app_id : undefined;
    if (!playId && !appStoreId) {
      setError("En az bir mağaza seçin");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const audit = await auditApi.create({
        client_name: clientName.trim(),
        app_display_name: query.trim() || playApp?.app_name || appStoreApp?.app_name || clientName.trim(),
        play_app_id: playId,
        appstore_app_id: appStoreId,
        country,
        play_review_limit: includePlay && playLimit > 0 ? playLimit : undefined,
        appstore_review_limit: includeAppStore && appStoreLimit > 0 ? appStoreLimit : undefined,
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
          Mağaza, ülke ve limit seçerek yorum çekin; MLC ile sınıflandırıp rapor üretin.
        </p>

        <form onSubmit={(e) => void handleSearch(e)} className="mt-8 space-y-4 rounded-2xl border border-border bg-surface p-5">
          <label className="block text-sm text-foreground">
            Uygulama ara
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className={inputClass}
              placeholder="Forge Master"
              required
            />
          </label>

          <label className="block text-sm text-foreground">
            Ülke / mağaza bölgesi
            <select value={country} onChange={(e) => setCountry(e.target.value)} className={inputClass}>
              {COUNTRIES.map((c) => (
                <option key={c.code} value={c.code}>
                  {c.label}
                </option>
              ))}
            </select>
          </label>

          <button
            type="submit"
            disabled={busy}
            className="rounded-lg border border-accent px-4 py-2 text-sm text-accent disabled:opacity-50"
          >
            Mağazalarda ara
          </button>
        </form>

        {(playApps.length > 0 || appStoreApps.length > 0) && (
          <form onSubmit={(e) => void handleStart(e)} className="mt-6 space-y-4 rounded-2xl border border-border bg-surface p-5">
            <label className="block text-sm text-foreground">
              Müşteri adı (rapor kapağı)
              <input value={clientName} onChange={(e) => setClientName(e.target.value)} className={inputClass} required />
            </label>

            <div className="grid gap-3 sm:grid-cols-2">
              <label className="flex items-center gap-2 text-sm text-foreground">
                <input type="checkbox" checked={includePlay} onChange={(e) => setIncludePlay(e.target.checked)} />
                Google Play çek
              </label>
              <label className="flex items-center gap-2 text-sm text-foreground">
                <input
                  type="checkbox"
                  checked={includeAppStore}
                  onChange={(e) => setIncludeAppStore(e.target.checked)}
                />
                App Store çek
              </label>
            </div>

            {includePlay && playApps.length > 0 && (
              <>
                <label className="block text-sm text-foreground">
                  Google Play uygulaması
                  <select
                    value={playApp?.app_id ?? ""}
                    onChange={(e) => setPlayApp(playApps.find((a) => a.app_id === e.target.value) ?? null)}
                    className={inputClass}
                  >
                    {playApps.map((a) => (
                      <option key={a.app_id} value={a.app_id}>
                        {a.app_name} — {a.developer}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block text-sm text-foreground">
                  Play yorum limiti (boş = mod varsayılanı, max {modeCap})
                  <input
                    type="number"
                    min={0}
                    max={10000}
                    value={playLimit || ""}
                    onChange={(e) => setPlayLimit(Number(e.target.value) || 0)}
                    className={inputClass}
                    placeholder={String(modeCap)}
                  />
                </label>
              </>
            )}

            {includeAppStore && appStoreApps.length > 0 && (
              <>
                <label className="block text-sm text-foreground">
                  App Store uygulaması
                  <select
                    value={appStoreApp?.app_id ?? ""}
                    onChange={(e) => setAppStoreApp(appStoreApps.find((a) => a.app_id === e.target.value) ?? null)}
                    className={inputClass}
                  >
                    {appStoreApps.map((a) => (
                      <option key={a.app_id} value={a.app_id}>
                        {a.app_name} — {a.developer}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block text-sm text-foreground">
                  App Store yorum limiti (boş = mod varsayılanı; RSS ~500 üst sınır)
                  <input
                    type="number"
                    min={0}
                    max={10000}
                    value={appStoreLimit || ""}
                    onChange={(e) => setAppStoreLimit(Number(e.target.value) || 0)}
                    className={inputClass}
                    placeholder={String(modeCap)}
                  />
                </label>
              </>
            )}

            <label className="block text-sm text-foreground">
              Mod
              <select
                value={mode}
                onChange={(e) => setMode(e.target.value as "quick" | "full")}
                className={inputClass}
              >
                <option value="quick">Quick — mağaza başına 500 yorum (~3–5 dk)</option>
                <option value="full">Full — mağaza başına 10.000 yorum (Play; App Store RSS sınırlı)</option>
              </select>
            </label>

            {error && (
              <p className="rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-500">{error}</p>
            )}

            <button
              type="submit"
              disabled={busy}
              className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-accent-fg disabled:opacity-50"
            >
              Audit başlat
            </button>
          </form>
        )}
      </div>
    </ProtectedRoute>
  );
}
