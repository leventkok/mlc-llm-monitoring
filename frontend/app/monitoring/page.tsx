"use client";

import { useEffect, useState, useCallback } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { reviewApi } from "@/lib/api";
import { Decision, Metrics, Score } from "@/types";
import { categoryBadge, sentimentBadge } from "@/lib/badges";

export default function MonitoringPage() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [scoresByDecision, setScoresByDecision] = useState<Record<string, Score>>(
    {},
  );

  const load = useCallback(async () => {
    const [m, d, scores] = await Promise.all([
      reviewApi.metrics(),
      reviewApi.decisions(),
      reviewApi.scores(),
    ]);
    setMetrics(m);
    setDecisions(d);
    const byDecision: Record<string, Score> = {};
    for (const s of scores) {
      if (!byDecision[s.decision_id]) {
        byDecision[s.decision_id] = s;
      }
    }
    setScoresByDecision(byDecision);
  }, []);

  useEffect(() => {
    load().catch(() => {});
  }, [load]);

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-6xl px-6 py-10">
        <div className="mb-8">
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
            monitoring
          </p>
          <h1 className="mt-2 text-2xl font-medium text-foreground">
            RAW LLM monitoring
          </h1>
          <p className="mt-1 text-sm text-muted">
            Auto-scored decisions: output format, latency, and distribution.
          </p>
        </div>

        {metrics && (
          <>
            <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
              <div className="rounded-2xl border border-accent/30 bg-accent/5 p-5 sm:col-span-2 lg:col-span-1">
                <p className="font-mono text-xs text-muted">compliance</p>
                <p className="mt-2 font-mono text-4xl font-medium text-accent">
                  {metrics.accuracy_pct.toFixed(0)}
                  <span className="text-lg">%</span>
                </p>
                <p className="mt-1 text-xs text-muted">quality ≥ 4 / 5</p>
              </div>

              <Stat label="reviews" value={metrics.total_reviews} />
              <Stat label="decisions" value={metrics.total_decisions} />
              <Stat
                label="avg quality"
                value={metrics.avg_quality.toFixed(1)}
                suffix="/5"
              />
              <Stat
                label="avg latency"
                value={Math.round(metrics.avg_latency_ms)}
                suffix="ms"
              />
            </div>

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

        <p className="mb-3 font-mono text-xs uppercase tracking-wider text-muted">
          decisions &amp; raw output
        </p>
        <div className="space-y-3">
          {decisions.length === 0 && (
            <p className="rounded-2xl border border-dashed border-border p-8 text-center text-sm text-muted">
              No decisions yet. Analyze some reviews on the dashboard first.
            </p>
          )}

          {decisions.map((d) => {
            const score = scoresByDecision[d.id];
            return (
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
                  {score ? (
                    <span className="rounded-md border border-accent/40 bg-accent/10 px-2 py-0.5 font-mono text-xs text-accent">
                      auto score {score.quality}/5
                    </span>
                  ) : (
                    <span className="rounded-md border border-border px-2 py-0.5 font-mono text-xs text-muted">
                      not scored
                    </span>
                  )}
                  <span className="ml-auto font-mono text-xs text-muted">
                    {d.latency_ms}ms
                  </span>
                </div>

                {d.raw_output && (
                  <pre className="mt-4 overflow-x-auto rounded-lg border border-border bg-background p-3 font-mono text-xs text-muted">
                    {d.raw_output}
                  </pre>
                )}
              </div>
            );
          })}
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
