"use client";

import { useEffect, useState, useCallback } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { reviewApi } from "@/lib/api";
import { Decision, Metrics } from "@/types";
import { categoryBadge, sentimentBadge } from "@/lib/badges";

const CATEGORIES = ["bug", "feature", "praise", "spam", "other"];

export default function MonitoringPage() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [scored, setScored] = useState<Record<string, boolean>>({});

  const load = useCallback(async () => {
    const [m, d] = await Promise.all([
      reviewApi.metrics(),
      reviewApi.decisions(),
    ]);
    setMetrics(m);
    setDecisions(d);
  }, []);

  useEffect(() => {
    load().catch(() => {});
  }, [load]);

  async function handleScore(
    decisionId: string,
    quality: number,
    correct?: string,
  ) {
    try {
      await reviewApi.score({
        decision_id: decisionId,
        quality,
        correct_category: correct,
      });
      setScored((prev) => ({ ...prev, [decisionId]: true }));
      load(); // refresh metrics after scoring
    } catch {
      // ignore for now
    }
  }

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-6xl px-6 py-10">
        <div className="mb-8">
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
            monitoring
          </p>
          <h1 className="mt-2 text-2xl font-medium text-foreground">
            Model performance
          </h1>
          <p className="mt-1 text-sm text-muted">
            Observe the raw model&apos;s decisions and score their quality.
          </p>
        </div>

        {/* Metrics */}
        {metrics && (
          <>
            {/* Accuracy hero + stat cards */}
            <div className="mb-6 grid gap-4 sm:grid-cols-4">
              <div className="rounded-2xl border border-accent/30 bg-accent/5 p-5">
                <p className="font-mono text-xs text-muted">accuracy</p>
                <p className="mt-2 font-mono text-4xl font-medium text-accent">
                  {metrics.accuracy_pct.toFixed(0)}
                  <span className="text-lg">%</span>
                </p>
                <p className="mt-1 text-xs text-muted">
                  vs. human ground truth
                </p>
              </div>

              <Stat label="reviews" value={metrics.total_reviews} />
              <Stat label="decisions" value={metrics.total_decisions} />
              <Stat
                label="avg quality"
                value={metrics.avg_quality.toFixed(1)}
                suffix="/5"
              />
            </div>

            {/* Category distribution bars */}
            <div className="mb-8 rounded-2xl border border-border bg-surface p-6">
              <p className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">
                category distribution
              </p>
              <Distribution
                counts={metrics.category_counts}
                total={metrics.total_decisions}
              />
            </div>
          </>
        )}

        {/* Decisions to score */}
        <p className="mb-3 font-mono text-xs uppercase tracking-wider text-muted">
          decisions
        </p>
        <div className="space-y-3">
          {decisions.length === 0 && (
            <p className="rounded-2xl border border-dashed border-border p-8 text-center text-sm text-muted">
              No decisions yet. Analyze some reviews on the dashboard first.
            </p>
          )}

          {decisions.map((d) => (
            <div
              key={d.id}
              className="rounded-2xl border border-border bg-surface p-5"
            >
              <div className="flex flex-wrap items-center gap-2">
                <span
                  className={`rounded-md border px-2 py-0.5 font-mono text-xs ${categoryBadge(d.category)}`}
                >
                  {d.category}
                </span>
                <span
                  className={`rounded-md border px-2 py-0.5 font-mono text-xs ${sentimentBadge(d.sentiment)}`}
                >
                  {d.sentiment}
                </span>
                <span className="ml-auto font-mono text-xs text-muted">
                  {d.latency_ms}ms
                </span>
              </div>

              {scored[d.id] ? (
                <p className="mt-4 font-mono text-xs text-emerald-500">
                  ✓ scored
                </p>
              ) : (
                <div className="mt-4 border-t border-border pt-4">
                  <p className="mb-2 font-mono text-xs text-muted">
                    rate this decision &amp; set the correct category
                  </p>
                  <div className="flex flex-wrap items-center gap-2">
                    {/* Quality 1-5 */}
                    <div className="flex gap-1">
                      {[1, 2, 3, 4, 5].map((q) => (
                        <button
                          key={q}
                          onClick={() => handleScore(d.id, q, d.category)}
                          className="h-8 w-8 rounded-lg border border-border font-mono text-xs text-muted transition hover:border-accent hover:text-accent"
                          title={`Quality ${q}, category correct`}
                        >
                          {q}
                        </button>
                      ))}
                    </div>
                    <span className="font-mono text-xs text-muted">
                      or mark correct category:
                    </span>
                    <select
                      onChange={(e) =>
                        e.target.value && handleScore(d.id, 3, e.target.value)
                      }
                      defaultValue=""
                      className="rounded-lg border border-border bg-background px-2 py-1 font-mono text-xs text-foreground outline-none focus:border-accent"
                    >
                      <option value="" disabled>
                        ground truth…
                      </option>
                      {CATEGORIES.map((c) => (
                        <option key={c} value={c}>
                          {c}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </ProtectedRoute>
  );
}

function Stat({
  label,
  value,
  suffix,
}: {
  label: string;
  value: number | string;
  suffix?: string;
}) {
  return (
    <div className="rounded-2xl border border-border bg-surface p-5">
      <p className="font-mono text-xs text-muted">{label}</p>
      <p className="mt-2 font-mono text-4xl font-medium text-foreground">
        {value}
        {suffix && <span className="text-lg text-muted">{suffix}</span>}
      </p>
    </div>
  );
}

function Distribution({
  counts,
  total,
}: {
  counts: Record<string, number>;
  total: number;
}) {
  const entries = Object.entries(counts);
  if (entries.length === 0) {
    return <p className="text-sm text-muted">No data yet.</p>;
  }
  const colors: Record<string, string> = {
    bug: "bg-red-500",
    feature: "bg-blue-500",
    praise: "bg-emerald-500",
    spam: "bg-amber-500",
    other: "bg-zinc-500",
  };
  return (
    <div className="space-y-3">
      {entries.map(([cat, n]) => {
        const pct = total > 0 ? (n / total) * 100 : 0;
        return (
          <div key={cat} className="flex items-center gap-3">
            <span className="w-16 font-mono text-xs text-muted">{cat}</span>
            <div className="h-2 flex-1 overflow-hidden rounded-full bg-surface-2">
              <div
                className={`h-full rounded-full ${colors[cat] || "bg-zinc-500"}`}
                style={{ width: `${pct}%` }}
              />
            </div>
            <span className="w-8 text-right font-mono text-xs text-foreground">
              {n}
            </span>
          </div>
        );
      })}
    </div>
  );
}
