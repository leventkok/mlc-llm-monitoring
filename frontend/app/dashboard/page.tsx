"use client";

import { useEffect, useState } from "react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { reviewApi } from "@/lib/api";
import { Review, Decision } from "@/types";
import { categoryBadge, sentimentBadge } from "@/lib/badges";

export default function DashboardPage() {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [text, setText] = useState("");
  const [appName, setAppName] = useState("");
  const [store, setStore] = useState("play");
  const [rating, setRating] = useState(3);
  const [submitting, setSubmitting] = useState(false);
  const [analyzing, setAnalyzing] = useState<string | null>(null);
  const [decisions, setDecisions] = useState<Record<string, Decision>>({});
  const [error, setError] = useState("");

  // Load existing reviews on mount
  useEffect(() => {
    reviewApi
      .list()
      .then(setReviews)
      .catch(() => {});
  }, []);

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    if (!text || !appName) return;
    setSubmitting(true);
    setError("");
    try {
      const review = await reviewApi.create({
        app_name: appName,
        store,
        rating,
        text,
      });
      setReviews((prev) => [review, ...prev]);
      setText("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add review");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleAnalyze(reviewId: string) {
    setAnalyzing(reviewId);
    try {
      const decision = await reviewApi.analyze(reviewId);
      setDecisions((prev) => ({ ...prev, [reviewId]: decision }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Analysis failed");
    } finally {
      setAnalyzing(null);
    }
  }

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-6xl px-6 py-10">
        <div className="mb-8">
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
            dashboard
          </p>
          <h1 className="mt-2 text-2xl font-medium text-foreground">
            Analyze reviews
          </h1>
          <p className="mt-1 text-sm text-muted">
            Add an app-store review and let the raw model classify it.
          </p>
        </div>

        {/* Add review form */}
        <form
          onSubmit={handleAdd}
          className="mb-8 space-y-4 rounded-2xl border border-border bg-surface p-6"
        >
          <div className="grid gap-4 sm:grid-cols-3">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-foreground">
                App name
              </label>
              <input
                value={appName}
                onChange={(e) => setAppName(e.target.value)}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-foreground outline-none transition focus:border-accent"
                placeholder="MyApp"
                required
              />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-foreground">
                Store
              </label>
              <select
                value={store}
                onChange={(e) => setStore(e.target.value)}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-foreground outline-none transition focus:border-accent"
              >
                <option value="play">Google Play</option>
                <option value="appstore">App Store</option>
              </select>
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-foreground">
                Rating
              </label>
              <select
                value={rating}
                onChange={(e) => setRating(Number(e.target.value))}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-foreground outline-none transition focus:border-accent"
              >
                {[1, 2, 3, 4, 5].map((r) => (
                  <option key={r} value={r}>
                    {r} ★
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-foreground">
              Review text
            </label>
            <textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              rows={3}
              className="w-full resize-none rounded-lg border border-border bg-background px-3 py-2 text-foreground outline-none transition focus:border-accent"
              placeholder="App keeps crashing on startup…"
              required
            />
          </div>

          {error && <p className="text-sm text-red-500">{error}</p>}

          <button
            type="submit"
            disabled={submitting}
            className="rounded-lg bg-accent px-4 py-2 font-medium text-accent-fg transition hover:opacity-90 disabled:opacity-50"
          >
            {submitting ? "Adding…" : "Add review"}
          </button>
        </form>

        {/* Reviews list */}
        <div className="space-y-3">
          {reviews.length === 0 && (
            <p className="rounded-2xl border border-dashed border-border p-8 text-center text-sm text-muted">
              No reviews yet. Add one above to get started.
            </p>
          )}

          {reviews.map((review) => {
            const decision = decisions[review.id];
            return (
              <div
                key={review.id}
                className="rounded-2xl border border-border bg-surface p-5"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="min-w-0">
                    <div className="mb-1 flex items-center gap-2 font-mono text-xs text-muted">
                      <span>{review.app_name}</span>
                      <span>·</span>
                      <span>{review.store}</span>
                      <span>·</span>
                      <span>{review.rating}★</span>
                    </div>
                    <p className="text-sm text-foreground">{review.text}</p>
                  </div>
                  <button
                    onClick={() => handleAnalyze(review.id)}
                    disabled={analyzing === review.id}
                    className="shrink-0 rounded-lg border border-accent px-3 py-1.5 font-mono text-xs text-accent transition hover:bg-accent hover:text-accent-fg disabled:opacity-50"
                  >
                    {analyzing === review.id ? "analyzing…" : "analyze"}
                  </button>
                </div>

                {/* Verdict card — the signature element */}
                {decision && (
                  <div className="mt-4 rounded-xl border border-border bg-background p-4">
                    <div className="mb-2 flex flex-wrap items-center gap-2">
                      <span className="font-mono text-xs text-muted">
                        verdict
                      </span>
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
                      <span className="ml-auto font-mono text-xs text-muted">
                        {decision.latency_ms}ms
                      </span>
                    </div>
                    <p className="font-mono text-xs text-muted">
                      {decision.raw_output}
                    </p>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </ProtectedRoute>
  );
}
