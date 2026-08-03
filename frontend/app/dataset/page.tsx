"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import ProtectedRoute from "@/components/ProtectedRoute";
import { useAuth } from "@/context/AuthContext";
import { datasetApi } from "@/lib/api";
import {
  BatchAnalyzeItem,
  BatchAnalyzeResult,
  DatasetReviewPage,
  DatasetReviewRow,
} from "@/types";

const HF_DATASET_URL =
  "https://huggingface.co/datasets/levonov/inferreview-app-reviews";
const PAGE_SIZES = [10, 20, 50] as const;

function storeLabel(store: string) {
  if (store === "google_play") return "Google Play";
  if (store === "app_store") return "App Store";
  return store;
}

function Badge({
  children,
  tone = "neutral",
}: {
  children: React.ReactNode;
  tone?: "neutral" | "good" | "bad" | "accent";
}) {
  const tones = {
    neutral: "border-border bg-surface-2 text-muted",
    good: "border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
    bad: "border-red-500/30 bg-red-500/10 text-red-500",
    accent: "border-accent/30 bg-accent/10 text-accent",
  };
  return (
    <span
      className={`inline-block rounded-md border px-2 py-0.5 font-mono text-xs ${tones[tone]}`}
    >
      {children}
    </span>
  );
}

function ReviewCard({ row }: { row: DatasetReviewRow }) {
  return (
    <article className="rounded-2xl border border-border bg-surface p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="font-mono text-xs text-accent">{row.review_id}</p>
          <p className="mt-1 text-sm font-medium text-foreground">
            {row.app_name}{" "}
            <span className="font-normal text-muted">v{row.app_version}</span>
          </p>
          <p className="mt-0.5 font-mono text-xs text-muted">
            {storeLabel(row.store)} · ★ {row.rating} · {row.language}
          </p>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {row.category && <Badge tone="accent">{row.category}</Badge>}
          {row.sentiment && <Badge>{row.sentiment}</Badge>}
        </div>
      </div>
      <p className="mt-3 text-sm leading-relaxed text-foreground">{row.text}</p>
    </article>
  );
}

function AnalyzeResultRow({ item }: { item: BatchAnalyzeItem }) {
  return (
    <div className="rounded-xl border border-border bg-surface-1 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="font-mono text-xs text-accent">{item.review_id}</span>
        <Badge tone={item.match ? "good" : "bad"}>
          {item.match ? "match" : "mismatch"}
        </Badge>
      </div>
      <p className="mt-2 line-clamp-2 text-sm text-muted">{item.text}</p>
      <div className="mt-3 grid gap-2 text-xs sm:grid-cols-2">
        <div>
          <p className="text-muted">expected</p>
          <p className="font-mono text-foreground">
            {item.expected_category ?? "—"} / {item.expected_sentiment ?? "—"}
          </p>
        </div>
        <div>
          <p className="text-muted">model</p>
          <p className="font-mono text-foreground">
            {item.category} / {item.sentiment}{" "}
            <span className="text-muted">({item.latency_ms} ms)</span>
          </p>
        </div>
      </div>
    </div>
  );
}

export default function DatasetPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [page, setPage] = useState<DatasetReviewPage | null>(null);
  const [offset, setOffset] = useState(0);
  const [limit, setLimit] = useState<number>(20);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [batchLimit, setBatchLimit] = useState(5);
  const [batchLoading, setBatchLoading] = useState(false);
  const [batchError, setBatchError] = useState("");
  const [batchResult, setBatchResult] = useState<BatchAnalyzeResult | null>(
    null,
  );
  const [csvLoading, setCsvLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await datasetApi.list(offset, limit);
      setPage(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load dataset");
    } finally {
      setLoading(false);
    }
  }, [offset, limit]);

  useEffect(() => {
    if (authLoading || !user) return;
    if (!user.is_admin) {
      router.replace("/home");
    }
  }, [authLoading, user, router]);

  useEffect(() => {
    if (authLoading || !user?.is_admin) return;
    void load();
  }, [load, authLoading, user]);

  async function handleExportCSV() {
    setCsvLoading(true);
    setError("");
    try {
      const csv = await datasetApi.exportCSV(offset, limit);
      const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `inferreview-dataset-${offset}-${limit}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : "CSV export failed");
    } finally {
      setCsvLoading(false);
    }
  }

  async function handleBatchAnalyze() {
    setBatchLoading(true);
    setBatchError("");
    setBatchResult(null);
    try {
      const result = await datasetApi.batchAnalyze(offset, batchLimit);
      setBatchResult(result);
    } catch (err) {
      setBatchError(
        err instanceof Error ? err.message : "Batch analyze failed",
      );
    } finally {
      setBatchLoading(false);
    }
  }

  const canPrev = offset > 0;
  const canNext = page ? page.rows.length >= limit : false;

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-6xl px-6 py-10">
        <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
              hugging face · dataset
            </p>
            <h1 className="mt-2 text-2xl font-medium text-foreground">
              App review dataset
            </h1>
            <p className="mt-1 max-w-xl text-sm text-muted">
              Browse labeled reviews from{" "}
              <a
                href={HF_DATASET_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="text-accent underline-offset-2 hover:underline"
              >
                levonov/inferreview-app-reviews
              </a>
              , export CSV, or run batch MLC classification against ground
              truth labels.
            </p>
            {page?.dataset_id && (
              <p className="mt-2 font-mono text-xs text-muted">
                source: {page.dataset_id}
              </p>
            )}
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => void load()}
              disabled={loading}
              className="rounded-lg border border-border px-3 py-1.5 font-mono text-xs text-foreground transition hover:bg-surface-2 disabled:opacity-50"
            >
              {loading ? "loading…" : "refresh"}
            </button>
            <button
              type="button"
              onClick={() => void handleExportCSV()}
              disabled={csvLoading || loading}
              className="rounded-lg border border-border px-3 py-1.5 font-mono text-xs text-foreground transition hover:bg-surface-2 disabled:opacity-50"
            >
              {csvLoading ? "exporting…" : "download csv"}
            </button>
          </div>
        </header>

        {error && (
          <p className="mb-6 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-500">
            {error}
          </p>
        )}

        <div className="mb-6 flex flex-wrap items-center gap-3 rounded-xl border border-border bg-surface-1 p-4">
          <label className="flex items-center gap-2 font-mono text-xs text-muted">
            page size
            <select
              value={limit}
              onChange={(e) => {
                setOffset(0);
                setLimit(Number(e.target.value));
              }}
              className="rounded-lg border border-border bg-background px-2 py-1 text-foreground"
            >
              {PAGE_SIZES.map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
          <div className="flex items-center gap-2">
            <button
              type="button"
              disabled={!canPrev || loading}
              onClick={() => setOffset((o) => Math.max(0, o - limit))}
              className="rounded-lg border border-border px-3 py-1 font-mono text-xs disabled:opacity-40"
            >
              ← prev
            </button>
            <span className="font-mono text-xs text-muted">
              offset {offset}
            </span>
            <button
              type="button"
              disabled={!canNext || loading}
              onClick={() => setOffset((o) => o + limit)}
              className="rounded-lg border border-border px-3 py-1 font-mono text-xs disabled:opacity-40"
            >
              next →
            </button>
          </div>
          {page && (
            <span className="font-mono text-xs text-muted">
              showing {page.rows.length} rows
            </span>
          )}
        </div>

        {loading && !page ? (
          <p className="font-mono text-sm text-muted">Loading dataset…</p>
        ) : (
          <div className="space-y-3">
            {page?.rows.map((row) => (
              <ReviewCard key={row.review_id} row={row} />
            ))}
            {page?.rows.length === 0 && (
              <p className="text-sm text-muted">No rows in this page.</p>
            )}
          </div>
        )}

        <section className="mt-12 rounded-2xl border border-border bg-surface p-6">
          <h2 className="text-lg font-medium text-foreground">
            Batch analyze
          </h2>
          <p className="mt-1 text-sm text-muted">
            Classify rows from the current page via server MLC (hybrid tunnel).
            Each review takes a few seconds — keep the batch small.
          </p>

          <div className="mt-4 flex flex-wrap items-end gap-4">
            <label className="block text-xs text-muted">
              rows to analyze (from current offset)
              <select
                value={batchLimit}
                onChange={(e) => setBatchLimit(Number(e.target.value))}
                className="mt-1 block rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm"
              >
                {[3, 5, 10, 20].map((n) => (
                  <option key={n} value={n}>
                    {n}
                  </option>
                ))}
              </select>
            </label>
            <button
              type="button"
              onClick={() => void handleBatchAnalyze()}
              disabled={batchLoading}
              className="rounded-lg border border-accent/40 bg-accent/10 px-4 py-2 font-mono text-sm text-accent transition hover:bg-accent/20 disabled:opacity-50"
            >
              {batchLoading ? "running MLC…" : "run batch analyze"}
            </button>
          </div>

          {batchError && (
            <p className="mt-4 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-500">
              {batchError}
            </p>
          )}

          {batchResult && (
            <div className="mt-6 space-y-4">
              <div className="flex flex-wrap gap-4 font-mono text-sm">
                <span className="text-foreground">
                  processed:{" "}
                  <strong>{batchResult.processed}</strong>
                </span>
                <span className="text-foreground">
                  accuracy:{" "}
                  <strong>{batchResult.accuracy_pct.toFixed(1)}%</strong>
                </span>
              </div>
              <div className="space-y-3">
                {batchResult.items.map((item) => (
                  <AnalyzeResultRow key={item.review_id} item={item} />
                ))}
              </div>
            </div>
          )}
        </section>
      </div>
    </ProtectedRoute>
  );
}
