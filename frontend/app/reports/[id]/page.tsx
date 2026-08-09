"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import ProtectedRoute from "@/components/ProtectedRoute";
import { auditApi } from "@/lib/api";
import type { AuditReport } from "@/types";

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="rounded-xl border border-border bg-surface p-4">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 text-2xl font-medium text-foreground">{value}</p>
    </div>
  );
}

export default function ReportDetailPage() {
  const params = useParams();
  const id = String(params.id ?? "");
  const [report, setReport] = useState<AuditReport | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!id) return;
    let cancelled = false;

    async function poll() {
      try {
        const data = await auditApi.report(id);
        if (cancelled) return;
        setReport(data);
        if (data.audit.status === "running" || data.audit.status === "queued") {
          setTimeout(poll, 3000);
        }
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load");
      }
    }

    void poll();
    return () => {
      cancelled = true;
    };
  }, [id]);

  const audit = report?.audit;
  const insights = report?.insights;
  const stats = insights?.statistics as Record<string, unknown> | undefined;

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-5xl px-6 py-10">
        {!audit && !error && <p className="text-muted">Loading audit…</p>}
        {error && <p className="text-red-500">{error}</p>}

        {audit && (
          <>
            <div className="mb-8">
              <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
                {audit.status} · {audit.step}
              </p>
              <h1 className="mt-2 text-2xl font-medium text-foreground">{audit.client_name}</h1>
              <p className="text-sm text-muted">{audit.app_display_name}</p>
            </div>

            {(audit.status === "running" || audit.status === "queued") && (
              <div className="mb-8 rounded-xl border border-accent/30 bg-accent/5 p-5 text-sm">
                <p>Crawling and analyzing…</p>
                <p className="mt-2 font-mono text-muted">
                  Play {audit.play_fetched} · App Store {audit.appstore_fetched} · analyzed{" "}
                  {audit.analyzed_count}/{audit.total_reviews}
                </p>
                {audit.truncated && (
                  <p className="mt-2 text-xs text-muted">Capped to newest reviews (10k limit).</p>
                )}
              </div>
            )}

            {audit.status === "failed" && (
              <p className="mb-6 rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-sm text-red-500">
                {audit.error_message || "Audit failed"}
              </p>
            )}

            {audit.status === "completed" && insights && (
              <div className="space-y-8">
                <section className="rounded-xl border border-border bg-surface p-6">
                  <h2 className="font-mono text-sm text-accent">Executive summary</h2>
                  <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-foreground">
                    {insights.executive_summary}
                  </p>
                </section>

                <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                  <StatCard label="Total reviews" value={String(stats?.total_reviews ?? audit.total_reviews)} />
                  <StatCard label="Avg rating" value={Number(stats?.avg_rating ?? 0).toFixed(2)} />
                  <StatCard label="Play reviews" value={String(stats?.play_count ?? audit.play_fetched)} />
                  <StatCard label="App Store reviews" value={String(stats?.appstore_count ?? audit.appstore_fetched)} />
                </section>

                <section className="rounded-xl border border-border bg-surface p-6">
                  <h2 className="font-mono text-sm text-accent">Root causes</h2>
                  <ul className="mt-4 space-y-4">
                    {insights.root_causes?.map((rc, i) => (
                      <li key={i} className="rounded-lg border border-border bg-background p-4">
                        <p className="font-medium text-foreground">{String(rc.theme ?? "Theme")}</p>
                        <p className="mt-1 text-sm text-muted">{String(rc.description ?? "")}</p>
                      </li>
                    ))}
                  </ul>
                </section>

                <section className="rounded-xl border border-border bg-surface p-6">
                  <h2 className="font-mono text-sm text-accent">90-day action plan</h2>
                  <div className="mt-4 overflow-x-auto">
                    <table className="w-full text-left text-sm">
                      <thead>
                        <tr className="border-b border-border text-muted">
                          <th className="py-2 pr-4">Priority</th>
                          <th className="py-2 pr-4">Horizon</th>
                          <th className="py-2 pr-4">Action</th>
                          <th className="py-2">Impact</th>
                        </tr>
                      </thead>
                      <tbody>
                        {insights.action_plan?.map((item, i) => (
                          <tr key={i} className="border-b border-border/60">
                            <td className="py-3 pr-4 font-mono text-accent">{String(item.priority ?? "")}</td>
                            <td className="py-3 pr-4">{String(item.horizon_days ?? "")}</td>
                            <td className="py-3 pr-4">{String(item.action ?? "")}</td>
                            <td className="py-3">{String(item.expected_impact ?? "")}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </section>
              </div>
            )}
          </>
        )}
      </div>
    </ProtectedRoute>
  );
}
