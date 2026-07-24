"use client";

import { useState } from "react";
import { Decision, Review, Score } from "@/types";
import { categoryBadge, sentimentBadge } from "@/lib/badges";
import {
  analyzeClassification,
  computeAutoQuality,
  getLatencyBand,
  latencyBandLabel,
  parseClassificationJSON,
  qualityLabel,
} from "@/lib/classification";

interface RichResultProps {
  decision: Decision;
  score?: Score;
  review?: Review;
  compact?: boolean;
}

export default function RichResult({
  decision,
  score,
  review,
  compact = false,
}: RichResultProps) {
  const [showRaw, setShowRaw] = useState(false);

  const analysis = analyzeClassification(
    decision.category,
    decision.sentiment,
    decision.raw_output,
  );
  const parsed = parseClassificationJSON(decision.raw_output);
  const quality = score?.quality ?? computeAutoQuality(
    decision.category,
    decision.sentiment,
    decision.raw_output,
    decision.latency_ms,
  );
  const latencyBand = getLatencyBand(decision.latency_ms);

  const latencyColors: Record<string, string> = {
    fast: "text-emerald-500",
    ok: "text-foreground",
    slow: "text-amber-500",
    critical: "text-red-500",
  };

  return (
    <div className={compact ? "space-y-3" : "mt-4 space-y-4"}>
      <div className="flex flex-wrap items-center gap-2">
        <span className="font-mono text-xs text-muted">rich result</span>
        <span
          className={`rounded-md border px-2 py-0.5 font-mono text-xs ${categoryBadge(decision.category)}`}
        >
          {decision.category}
        </span>
        <span
          className={`rounded-md border px-2 py-0.5 font-mono text-xs ${sentimentBadge(decision.sentiment)}`}
        >
          {decision.sentiment}
        </span>
        <QualityPill quality={quality} scored={Boolean(score)} />
        <span
          className={`ml-auto font-mono text-xs ${latencyColors[latencyBand]}`}
        >
          {decision.latency_ms}ms · {latencyBandLabel(latencyBand)}
        </span>
      </div>

      {review && (
        <div className="rounded-lg border border-border bg-background/50 px-3 py-2">
          <p className="font-mono text-[10px] uppercase tracking-wider text-muted">
            source review
          </p>
          <p className="mt-1 text-sm text-foreground">{review.text}</p>
          <p className="mt-1 font-mono text-xs text-muted">
            {review.app_name} · {review.store} · {review.rating}★
          </p>
        </div>
      )}

      <div className="grid gap-3 sm:grid-cols-2">
        <ResultTable
          rows={[
            ["category", decision.category],
            ["sentiment", decision.sentiment],
            ["json parsed", analysis.jsonMatch ? "yes" : "no"],
            ["fields consistent", analysis.consistent ? "yes" : "no"],
            ["output valid", analysis.valid ? "yes" : "no"],
          ]}
        />
        <div className="rounded-xl border border-border bg-background p-4">
          <p className="mb-3 font-mono text-[10px] uppercase tracking-wider text-muted">
            quality score
          </p>
          <QualityMeter quality={quality} />
          <p className="mt-2 text-xs text-muted">
            {qualityLabel(quality)} — auto-scored output health (format + latency)
          </p>
        </div>
      </div>

      {parsed && (
        <div className="rounded-xl border border-accent/20 bg-accent/5 p-4">
          <p className="mb-2 font-mono text-[10px] uppercase tracking-wider text-accent">
            parsed classification
          </p>
          <div className="grid gap-2 sm:grid-cols-2">
            <Field label="category" value={parsed.category} />
            <Field label="sentiment" value={parsed.sentiment} />
          </div>
        </div>
      )}

      <div>
        <button
          type="button"
          onClick={() => setShowRaw((v) => !v)}
          className="font-mono text-xs text-accent transition hover:opacity-80"
        >
          {showRaw ? "hide raw output" : "show raw output"}
        </button>
        {showRaw && (
          <pre className="mt-2 overflow-x-auto rounded-lg border border-border bg-background p-3 font-mono text-xs text-muted">
            {decision.raw_output || "(empty)"}
          </pre>
        )}
      </div>
    </div>
  );
}

function QualityPill({
  quality,
  scored,
}: {
  quality: number;
  scored: boolean;
}) {
  return (
    <span className="rounded-md border border-accent/40 bg-accent/10 px-2 py-0.5 font-mono text-xs text-accent">
      {scored ? "auto score" : "computed"} {quality}/5
    </span>
  );
}

function QualityMeter({ quality }: { quality: number }) {
  return (
    <div className="flex items-center gap-1">
      {[1, 2, 3, 4, 5].map((n) => (
        <div
          key={n}
          className={`h-2 flex-1 rounded-full ${
            n <= quality ? "bg-accent" : "bg-surface-2"
          }`}
        />
      ))}
      <span className="ml-2 font-mono text-sm font-medium text-foreground">
        {quality}/5
      </span>
    </div>
  );
}

function ResultTable({ rows }: { rows: [string, string][] }) {
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-background">
      <table className="w-full text-left text-sm">
        <tbody>
          {rows.map(([label, value]) => (
            <tr key={label} className="border-b border-border last:border-0">
              <td className="w-2/5 px-3 py-2 font-mono text-xs text-muted">
                {label}
              </td>
              <td className="px-3 py-2 font-mono text-xs text-foreground">
                {value}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-background px-3 py-2">
      <p className="font-mono text-[10px] uppercase text-muted">{label}</p>
      <p className="mt-0.5 font-mono text-sm text-foreground">{value}</p>
    </div>
  );
}

export function DistributionBars({
  counts,
  total,
  colors,
}: {
  counts: Record<string, number>;
  total: number;
  colors?: Record<string, string>;
}) {
  const entries = Object.entries(counts).sort((a, b) => b[1] - a[1]);
  if (entries.length === 0) {
    return <p className="text-sm text-muted">No data yet.</p>;
  }

  const defaultColors: Record<string, string> = {
    bug: "bg-red-500",
    feature: "bg-blue-500",
    praise: "bg-emerald-500",
    spam: "bg-amber-500",
    other: "bg-zinc-500",
    positive: "bg-emerald-500",
    negative: "bg-red-500",
    neutral: "bg-zinc-500",
  };

  const palette = colors ?? defaultColors;

  return (
    <div className="space-y-3">
      {entries.map(([key, n]) => {
        const pct = total > 0 ? (n / total) * 100 : 0;
        return (
          <div key={key} className="flex items-center gap-3">
            <span className="w-20 truncate font-mono text-xs text-muted">
              {key}
            </span>
            <div className="h-2 flex-1 overflow-hidden rounded-full bg-surface-2">
              <div
                className={`h-full rounded-full ${palette[key] || "bg-zinc-500"}`}
                style={{ width: `${pct}%` }}
              />
            </div>
            <span className="w-10 text-right font-mono text-xs text-foreground">
              {n}
            </span>
            <span className="w-12 text-right font-mono text-[10px] text-muted">
              {pct.toFixed(0)}%
            </span>
          </div>
        );
      })}
    </div>
  );
}

export function QualityHistogram({
  scores,
}: {
  scores: Record<string, Score>;
}) {
  const counts = [0, 0, 0, 0, 0, 0];
  for (const s of Object.values(scores)) {
    if (s.quality >= 1 && s.quality <= 5) {
      counts[s.quality]++;
    }
  }
  const total = counts.slice(1).reduce((a, b) => a + b, 0);
  if (total === 0) {
    return <p className="text-sm text-muted">No scores yet.</p>;
  }

  return (
    <div className="flex items-end gap-2">
      {[1, 2, 3, 4, 5].map((q) => {
        const n = counts[q];
        const pct = total > 0 ? (n / total) * 100 : 0;
        return (
          <div key={q} className="flex flex-1 flex-col items-center gap-1">
            <div className="flex h-24 w-full items-end rounded-t-lg bg-surface-2">
              <div
                className="w-full rounded-t-lg bg-accent transition-all"
                style={{ height: `${Math.max(pct, n > 0 ? 8 : 0)}%` }}
                title={`${n} decisions`}
              />
            </div>
            <span className="font-mono text-xs text-muted">{q}</span>
            <span className="font-mono text-[10px] text-muted">{n}</span>
          </div>
        );
      })}
    </div>
  );
}
