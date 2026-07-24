"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import RichResult, {
  DistributionBars,
  QualityHistogram,
} from "@/components/RichResult";
import { reviewApi } from "@/lib/api";
import { Decision, Metrics, Review, Score } from "@/types";

const GRAFANA_URL =
  process.env.NEXT_PUBLIC_GRAFANA_URL ?? "https://grafana.inferreview.com";

export default function MonitoringPage() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [scoresByDecision, setScoresByDecision] = useState<
    Record<string, Score>
  >({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filter, setFilter] = useState<"all" | "scored" | "unscored">("all");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [m, d, scores, revs] = await Promise.all([
        reviewApi.metrics(),
        reviewApi.decisions(),
        reviewApi.scores(),
        reviewApi.list(),
      ]);
      setMetrics(m);
      setDecisions(d);
      setReviews(revs);
      const byDecision: Record<string, Score> = {};
      for (const s of scores) {
        if (!byDecision[s.decision_id]) {
          byDecision[s.decision_id] = s;
        }
      }
      setScoresByDecision(byDecision);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load monitoring data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load().catch(() => {});
  }, [load]);

  const reviewsById = useMemo(() => {
    const map: Record<string, Review> = {};
    for (const r of reviews) map[r.id] = r;
    return map;
  }, [reviews]);

  const filteredDecisions = useMemo(() => {
    return decisions.filter((d) => {
      const scored = Boolean(scoresByDecision[d.id]);
      if (filter === "scored") return scored;
      if (filter === "unscored") return !scored;
      return true;
    });
  }, [decisions, scoresByDecision, filter]);

  const unscoredCount = decisions.length - Object.keys(scoresByDecision).length;
  const compliantCount = metrics
    ? Math.round((metrics.accuracy_pct / 100) * metrics.total_scores)
    : 0;

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-6xl px-6 py-10">
        <div className="mb-8 flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
              monitoring
            </p>
            <h1 className="mt-2 text-2xl font-medium text-foreground">
              LLM output monitoring
            </h1>
            <p className="mt-1 text-sm text-muted">
              Auto-scored decisions, compliance KPIs, distributions, and raw
              output inspection.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => void load()}
              disabled={loading}
              className="rounded-lg border border-border px-3 py-1.5 font-mono text-xs text-foreground transition hover:bg-surface-2 disabled:opacity-50"
            >
              {loading ? "refreshing…" : "refresh"}
            </button>
            <a
              href={GRAFANA_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-lg border border-accent/40 bg-accent/10 px-3 py-1.5 font-mono text-xs text-accent transition hover:bg-accent/20"
            >
              open grafana ↗
            </a>
          </div>
        </div>

        {error && (
          <p className="mb-6 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-500">
            {error}
          </p>
        )}

        {metrics && (
          <>
            <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-6">
              <div className="rounded-2xl border border-accent/30 bg-accent/5 p-5 sm:col-span-2 lg:col-span-1">
                <p className="font-mono text-xs text-muted">compliance</p>
                <p className="mt-2 font-mono text-4xl font-medium text-accent">
                  {metrics.accuracy_pct.toFixed(0)}
                  <span className="text-lg">%</span>
                </p>
                <p className="mt-1 text-xs text-muted">
                  {compliantCount}/{metrics.total_scores} scores ≥ 4/5
                </p>
              </div>

              <Stat label="reviews" value={metrics.total_reviews} />
              <Stat label="decisions" value={metrics.total_decisions} />
              <Stat label="auto scores" value={metrics.total_scores} />
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

            <div className="mb-8 grid gap-4 lg:grid-cols-3">
              <Panel title="category distribution">
                <DistributionBars
                  counts={metrics.category_counts}
                  total={metrics.total_decisions}
                />
              </Panel>
              <Panel title="sentiment distribution">
                <DistributionBars
                  counts={metrics.sentiment_counts}
                  total={metrics.total_decisions}
                />
              </Panel>
              <Panel title="quality distribution (1–5)">
                <QualityHistogram scores={scoresByDecision} />
              </Panel>
            </div>

            <div className="mb-8 rounded-2xl border border-border bg-surface p-5">
              <p className="font-mono text-xs uppercase tracking-wider text-muted">
                coverage
              </p>
              <div className="mt-3 grid gap-3 sm:grid-cols-3">
                <CoverageStat
                  label="scored decisions"
                  value={metrics.total_scores}
                  total={metrics.total_decisions}
                />
                <CoverageStat
                  label="unscored (legacy)"
                  value={Math.max(unscoredCount, 0)}
                  total={metrics.total_decisions}
                />
                <CoverageStat
                  label="reviews analyzed"
                  value={metrics.total_decisions}
                  total={metrics.total_reviews}
                />
              </div>
            </div>
          </>
        )}

        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <p className="font-mono text-xs uppercase tracking-wider text-muted">
            decisions &amp; rich results
          </p>
          <div className="flex gap-1 rounded-lg border border-border p-0.5">
            {(["all", "scored", "unscored"] as const).map((f) => (
              <button
                key={f}
                type="button"
                onClick={() => setFilter(f)}
                className={`rounded-md px-2.5 py-1 font-mono text-xs transition ${
                  filter === f
                    ? "bg-surface-2 text-foreground"
                    : "text-muted hover:text-foreground"
                }`}
              >
                {f}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-4">
          {filteredDecisions.length === 0 && !loading && (
            <p className="rounded-2xl border border-dashed border-border p-8 text-center text-sm text-muted">
              {decisions.length === 0
                ? "No decisions yet. Analyze some reviews on the dashboard first."
                : "No decisions match this filter."}
            </p>
          )}

          {filteredDecisions.map((d) => (
            <div
              key={d.id}
              className="rounded-2xl border border-border bg-surface p-5"
            >
              <div className="mb-1 flex flex-wrap items-center gap-2 font-mono text-[10px] text-muted">
                <span>{new Date(d.created_at).toLocaleString()}</span>
                <span>·</span>
                <span>id {d.id.slice(0, 8)}…</span>
              </div>
              <RichResult
                decision={d}
                score={scoresByDecision[d.id]}
                review={reviewsById[d.review_id]}
              />
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

function Panel({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="rounded-2xl border border-border bg-surface p-6">
      <p className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">
        {title}
      </p>
      {children}
    </div>
  );
}

function CoverageStat({
  label,
  value,
  total,
}: {
  label: string;
  value: number;
  total: number;
}) {
  const pct = total > 0 ? (value / total) * 100 : 0;
  return (
    <div>
      <div className="flex items-baseline justify-between">
        <span className="text-xs text-muted">{label}</span>
        <span className="font-mono text-xs text-foreground">
          {value}/{total}
        </span>
      </div>
      <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-surface-2">
        <div
          className="h-full rounded-full bg-accent"
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
