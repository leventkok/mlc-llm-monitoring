"use client";

import { useState } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { mcpDeepKwikiSearch, MCPDeepKwikiResult } from "@/lib/webmcp";

export default function DeepKwikiPage() {
  const [repo, setRepo] = useState("mlc-ai/mlc-llm");
  const [query, setQuery] = useState(
    "How do I serve a quantized model with MLC-LLM?",
  );
  const [mode, setMode] = useState<"ask" | "structure" | "contents">("ask");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<MCPDeepKwikiResult | null>(null);

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    setResult(null);
    try {
      const res = await mcpDeepKwikiSearch(repo.trim(), query.trim(), mode);
      setResult(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "DeepKwiki search failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-4xl px-6 py-10">
        <header className="mb-8">
          <p className="font-mono text-xs uppercase tracking-wider text-accent">
            FINAL BOSS · DeepKwiki
          </p>
          <h1 className="mt-2 text-2xl font-semibold text-foreground">
            Repository research
          </h1>
          <p className="mt-2 text-sm text-muted">
            WebMCP packages your query → Go backend → DeepWiki MCP (
            <code className="text-foreground">mcp.deepwiki.com</code>). Use public{" "}
            <code className="text-foreground">owner/repo</code> slugs.
          </p>
        </header>

        <form
          onSubmit={(e) => void handleSearch(e)}
          className="space-y-4 rounded-xl border border-border bg-surface-1 p-5"
        >
          <label className="block text-xs text-muted">
            GitHub repository
            <input
              value={repo}
              onChange={(e) => setRepo(e.target.value)}
              placeholder="mlc-ai/mlc-llm"
              className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm"
              required
            />
          </label>

          <label className="block text-xs text-muted">
            Mode
            <select
              value={mode}
              onChange={(e) =>
                setMode(e.target.value as "ask" | "structure" | "contents")
              }
              className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm"
            >
              <option value="ask">ask — AI answer</option>
              <option value="structure">structure — wiki TOC</option>
              <option value="contents">contents — full wiki doc</option>
            </select>
          </label>

          {mode === "ask" && (
            <label className="block text-xs text-muted">
              Question
              <textarea
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                rows={3}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                required
              />
            </label>
          )}

          {error && <p className="text-sm text-red-500">{error}</p>}

          <button
            type="submit"
            disabled={loading}
            className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-accent-foreground disabled:opacity-50"
          >
            {loading ? "Searching…" : "Search via WebMCP"}
          </button>
        </form>

        {result && (
          <section className="mt-6 rounded-xl border border-border bg-surface-1 p-5">
            <div className="mb-3 flex flex-wrap items-center gap-2 font-mono text-xs text-muted">
              <span>{result.repo_name}</span>
              <span>·</span>
              <span>{result.mode}</span>
              <span>·</span>
              <span>{result.latency_ms}ms</span>
            </div>
            <pre className="whitespace-pre-wrap font-mono text-xs leading-relaxed text-foreground">
              {result.answer}
            </pre>
          </section>
        )}
      </div>
    </ProtectedRoute>
  );
}
