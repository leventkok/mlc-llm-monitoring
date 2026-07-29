"use client";

import { useEffect, useState, useCallback } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { useAuth } from "@/context/AuthContext";
import { adminApi } from "@/lib/api";
import { AnalyzeLogEntry, LLMConfig, ModelProfile, ModelSwitchRequest } from "@/types";
import { useRouter } from "next/navigation";

const GRAFANA_URL =
  process.env.NEXT_PUBLIC_GRAFANA_URL ?? "https://grafana.inferreview.com";

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
  const [profiles, setProfiles] = useState<ModelProfile[]>([]);
  const [selectedProfile, setSelectedProfile] = useState("");
  const [switchStatus, setSwitchStatus] = useState<ModelSwitchRequest | null>(null);
  const [switching, setSwitching] = useState(false);

  const loadSwitchStatus = useCallback(async () => {
    try {
      const status = await adminApi.modelSwitchStatus();
      setSwitchStatus(status);
    } catch {
      /* optional poll */
    }
  }, []);

  useEffect(() => {
    if (user && !user.is_admin) {
      router.replace("/dashboard");
    }
  }, [user, router]);

  useEffect(() => {
    if (!user?.is_admin) return;
    void loadAll();
  }, [user]);

  useEffect(() => {
    if (!user?.is_admin) return;
    const id = setInterval(() => void loadSwitchStatus(), 5000);
    return () => clearInterval(id);
  }, [user, loadSwitchStatus]);

  async function loadAll() {
    setLoading(true);
    setError("");
    try {
      const [cfg, entries, profs, sw] = await Promise.all([
        adminApi.getLLMConfig(),
        adminApi.analyzeLogs(50),
        adminApi.modelProfiles(),
        adminApi.modelSwitchStatus(),
      ]);
      setLLM(cfg);
      setLogs(entries);
      setProfiles(profs);
      setSwitchStatus(sw);
      if (!selectedProfile && profs.length > 0) {
        const active = profs.find((p) => p.request_model === cfg.active_model);
        setSelectedProfile(active?.id ?? profs[0].id);
      }
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

  async function handleSwitchModel() {
    if (!selectedProfile) return;
    setSwitching(true);
    setMessage("");
    setError("");
    try {
      const req = await adminApi.switchModel(selectedProfile);
      setSwitchStatus(req);
      const cfg = await adminApi.getLLMConfig();
      setLLM(cfg);
      setMessage(
        "Model switch queued — mlc-agent will restart the local engine (2–15 min JIT). Prompt/temperature already live."
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Switch failed");
    } finally {
      setSwitching(false);
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
        <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="font-mono text-xs uppercase tracking-wider text-accent">
              FINAL BOSS · Admin LLM Modify Panel
            </p>
            <h1 className="mt-2 text-2xl font-semibold text-foreground">
              LLM cockpit
            </h1>
            <p className="mt-2 max-w-xl text-sm text-muted">
              Live prompt and inference limits apply instantly. Model adapter swap
              queues a local engine restart via mlc-agent (MLC 0.20 merges LoRA at
              convert time).
            </p>
          </div>
          <a
            href={GRAFANA_URL}
            target="_blank"
            rel="noopener noreferrer"
            title="Requires Cloudflare Access (if configured) and Grafana admin login"
            className="rounded-lg border border-accent/40 bg-accent/10 px-3 py-1.5 font-mono text-xs text-accent transition hover:bg-accent/20"
          >
            open grafana ↗
          </a>
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

              <div className="mt-6 space-y-3 border-t border-border pt-5">
                <h3 className="font-mono text-xs font-medium uppercase text-muted">
                  Hot-swap model profile
                </h3>
                <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
                  <label className="block flex-1 text-xs text-muted">
                    Profile
                    <select
                      value={selectedProfile}
                      onChange={(e) => setSelectedProfile(e.target.value)}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm text-foreground"
                    >
                      {profiles.map((p) => (
                        <option key={p.id} value={p.id}>
                          {p.label}
                        </option>
                      ))}
                    </select>
                  </label>
                  <button
                    type="button"
                    disabled={switching || !selectedProfile}
                    onClick={() => void handleSwitchModel()}
                    className="rounded-lg border border-accent px-4 py-2 text-sm font-medium text-accent disabled:opacity-50"
                  >
                    {switching ? "Queuing…" : "Switch model"}
                  </button>
                </div>
                {profiles.find((p) => p.id === selectedProfile)?.description && (
                  <p className="text-xs text-muted">
                    {profiles.find((p) => p.id === selectedProfile)?.description}
                  </p>
                )}
                {switchStatus && (
                  <div className="rounded-lg border border-border bg-background px-3 py-2 font-mono text-xs">
                    <span className="text-muted">Last switch: </span>
                    <span
                      className={
                        switchStatus.status === "completed"
                          ? "text-emerald-500"
                          : switchStatus.status === "failed"
                            ? "text-red-500"
                            : "text-amber-500"
                      }
                    >
                      {switchStatus.status}
                    </span>
                    <span className="text-muted"> · {switchStatus.profile_id}</span>
                    {switchStatus.error_message && (
                      <p className="mt-1 text-red-500">{switchStatus.error_message}</p>
                    )}
                  </div>
                )}
                <p className="text-xs text-muted">
                  Requires <code className="text-foreground">mlc-agent</code> in the
                  real-mlc hybrid stack. API model id updates immediately; GPU
                  weights reload after engine restart.
                </p>
              </div>
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
