"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import ProtectedRoute from "@/components/ProtectedRoute";
import AuditReportView from "@/components/audit/AuditReportView";
import { auditApi } from "@/lib/api";
import type { AuditReport } from "@/types";

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

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-6xl px-6 py-10">
        {!audit && !error && <p className="text-muted">Loading audit…</p>}
        {error && <p className="text-red-500">{error}</p>}

        {audit && (
          <>
            {(audit.status === "running" || audit.status === "queued") && (
              <div className="mb-6 rounded-xl border border-accent/30 bg-accent/5 p-5 text-sm">
                <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
                  {audit.status} · {audit.step}
                </p>
                <p className="mt-2">Crawling and analyzing…</p>
                <p className="mt-2 font-mono text-muted">
                  Play {audit.play_fetched} · App Store {audit.appstore_fetched} · analyzed{" "}
                  {audit.analyzed_count}/{audit.total_reviews}
                </p>
              </div>
            )}

            {audit.status === "failed" && (
              <p className="mb-6 rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-sm text-red-500">
                {audit.error_message || "Audit failed"}
              </p>
            )}

            {audit.status === "completed" && insights && (
              <AuditReportView audit={audit} insights={insights} />
            )}
          </>
        )}
      </div>
    </ProtectedRoute>
  );
}
