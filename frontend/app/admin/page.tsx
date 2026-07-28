"use client";

import { useEffect, useState } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { useAuth } from "@/context/AuthContext";
import { adminApi } from "@/lib/api";
import { AnalyzeLogEntry, LLMConfig } from "@/types";
import { useRouter } from "next/navigation";

const defaultLLM: LLMConfig = {
  system_prompt: "",
  temperature: 0,
  max_tokens: 60,
  top_p: 1,
  active_model: "",
  active_adapter: "",
};

export default function AdminPage() {
  const { user } = useAuth();
  const router = useRouter();
  const [llm, setLLM] = useState<LLMConfig>(defaultLLM);
  const [logs, setLogs] = useState<AnalyzeLogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (user && !user.is_admin) {
      router.replace("/dashboard");
    }
  }, [user, router]);

  useEffect(() => {
    if (!user?.is_admin) return;
    void loadAll();
  }, [user]);

  async function loadAll() {
    setLoading(true);
    setError("");
    try {
      const [cfg, entries] = await Promise.all([
        adminApi.getLLMConfig(),
        adminApi.analyzeLogs(50),
      ]);
      setLLM(cfg);
      setLogs(entries);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load admin data");
    } finally {
      setLoading(false);
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setMessage("");
    setError("");
    try {
      const updated = await adminApi.updateLLMConfig(llm);
      setLLM(updated);
      setMessage("LLM settings saved — applies to the next analyze request.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }

  if (!user?.is_admin) {
    return (
      <ProtectedRoute>
        <div className="mx-auto max-w-lg px-6 py-16 text-center text-muted">
          Admin access required.
        </div>
      </ProtectedRoute>
    );
  }

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-5xl px-6 py-10">
        <header className="mb-8">
          <p className="font-mono text-xs uppercase tracking-wider text-accent">
            FINAL BOSS · Admin LLM Modify Panel
          </p>
          <h1 className="mt-2 text-2xl font-semibold text-foreground">
            LLM cockpit
          </h1>
          <p className="mt-2 text-sm text-muted">
            Live prompt and inference limits (no model restart). Adapter hot-swap
            requires MLC convert + engine restart — see status below.
          </p>
        </header>

        {loading ? (
          <p className="text-sm text-muted">Loading…</p>
        ) : (
          <div className="grid gap-6 lg:grid-cols-2">
            <section className="rounded-xl border border-border bg-surface-1 p-5 lg:col-span-2">
              <h2 className="font-mono text-sm font-medium text-foreground">
                Adapter status
              </h2>
              <dl className="mt-4 grid gap-3 sm:grid-cols-2">
                <Stat label="Active model" value={llm.active_model || "—"} />
                <Stat
                  label="LoRA adapter"
                  value={llm.active_adapter || "(merged in model weights)"}
                />
              </dl>
              <p className="mt-4 text-xs text-muted">
                To switch adapters: merge LoRA in Colab or locally, upload to HF,
                update <code className="text-foreground">MLC_MODEL</code> and restart{" "}
                <code className="text-foreground">mlc-engine</code>.
              </p>
            </section>

            <form
              onSubmit={(e) => void handleSave(e)}
              className="space-y-4 rounded-xl border border-border bg-surface-1 p-5 lg:col-span-2"
            >
              <h2 className="font-mono text-sm font-medium text-foreground">
                System prompt
              </h2>
              <textarea
                value={llm.system_prompt}
                onChange={(e) =>
                  setLLM((prev) => ({ ...prev, system_prompt: e.target.value }))
                }
                rows={12}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-xs text-foreground"
              />

              <h2 className="font-mono text-sm font-medium text-foreground">
                Context limits
              </h2>
              <div className="grid gap-4 sm:grid-cols-3">
                <label className="block text-xs text-muted">
                  Temperature
                  <input
                    type="number"
                    step="0.1"
                    min={0}
                    max={2}
                    value={llm.temperature}
                    onChange={(e) =>
                      setLLM((prev) => ({
                        ...prev,
                        temperature: Number(e.target.value),
                      }))
                    }
                    className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm"
                  />
                </label>
                <label className="block text-xs text-muted">
                  Max tokens
                  <input
                    type="number"
                    min={1}
                    max={4096}
                    value={llm.max_tokens}
                    onChange={(e) =>
                      setLLM((prev) => ({
                        ...prev,
                        max_tokens: Number(e.target.value),
                      }))
                    }
                    className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm"
                  />
                </label>
                <label className="block text-xs text-muted">
                  Top-P
                  <input
                    type="number"
                    step="0.05"
                    min={0}
                    max={1}
                    value={llm.top_p}
                    onChange={(e) =>
                      setLLM((prev) => ({
                        ...prev,
                        top_p: Number(e.target.value),
                      }))
                    }
                    className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm"
                  />
                </label>
              </div>

              {message && (
                <p className="text-sm text-emerald-500">{message}</p>
              )}
              {error && <p className="text-sm text-red-500">{error}</p>}

              <button
                type="submit"
                disabled={saving}
                className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-accent-foreground disabled:opacity-50"
              >
                {saving ? "Saving…" : "Save LLM settings"}
              </button>
            </form>

            <section className="rounded-xl border border-border bg-surface-1 p-5 lg:col-span-2">
              <div className="mb-4 flex items-center justify-between">
                <h2 className="font-mono text-sm font-medium text-foreground">
                  Analyze log monitor
                </h2>
                <button
                  type="button"
                  onClick={() => void loadAll()}
                  className="font-mono text-xs text-accent hover:opacity-80"
                >
                  refresh
                </button>
              </div>
              {logs.length === 0 ? (
                <p className="text-sm text-muted">No analyze requests yet.</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead>
                      <tr className="border-b border-border text-muted">
                        <th className="py-2 pr-3">time</th>
                        <th className="py-2 pr-3">status</th>
                        <th className="py-2 pr-3">category</th>
                        <th className="py-2 pr-3">latency</th>
                        <th className="py-2">review</th>
                      </tr>
                    </thead>
                    <tbody>
                      {logs.map((entry, i) => (
                        <tr key={`${entry.time}-${i}`} className="border-b border-border/50">
                          <td className="py-2 pr-3 font-mono text-muted">
                            {new Date(entry.time).toLocaleTimeString()}
                          </td>
                          <td className="py-2 pr-3">{entry.status}</td>
                          <td className="py-2 pr-3">
                            {entry.category
                              ? `${entry.category}/${entry.sentiment}`
                              : "—"}
                          </td>
                          <td className="py-2 pr-3">
                            {entry.latency_ms ? `${entry.latency_ms}ms` : "—"}
                          </td>
                          <td className="py-2 font-mono text-muted">
                            {entry.review_id.slice(0, 8)}…
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </section>
          </div>
        )}
      </div>
    </ProtectedRoute>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-background px-3 py-2">
      <dt className="font-mono text-[10px] uppercase text-muted">{label}</dt>
      <dd className="mt-1 break-all font-mono text-xs text-foreground">{value}</dd>
    </div>
  );
}
