"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";

export default function HomePage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && user) router.replace("/home");
  }, [user, loading, router]);

  return (
    <div className="min-h-screen bg-background">
      <header className="mx-auto flex max-w-6xl items-center justify-between px-6 py-5">
        <p className="font-mono text-sm font-medium tracking-tight text-foreground">
          Infer<span className="text-accent">Review</span>
        </p>
        <div className="flex items-center gap-3">
          <Link
            href="/login"
            className="rounded-lg px-3 py-1.5 text-sm text-muted transition hover:text-foreground"
          >
            Sign in
          </Link>
          <Link
            href="/register"
            className="rounded-lg border border-accent/40 bg-accent/10 px-3 py-1.5 font-mono text-sm text-accent transition hover:bg-accent/20"
          >
            Get started
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 pb-20 pt-8">
        <section className="max-w-2xl">
          <p className="font-mono text-xs uppercase tracking-[0.25em] text-accent">
            MLC · app store reviews · monitoring
          </p>
          <h1 className="mt-4 text-4xl font-medium leading-tight text-foreground sm:text-5xl">
            Classify app reviews with local MLC — monitor quality over time
          </h1>
          <p className="mt-5 text-lg text-muted">
            InferReview combines hybrid GPU inference, Hugging Face datasets,
            and rich monitoring for Play Store and App Store feedback.
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Link
              href="/register"
              className="rounded-xl bg-accent px-5 py-2.5 font-medium text-accent-fg transition hover:opacity-90"
            >
              Start free — individual
            </Link>
            <Link
              href="/login"
              className="rounded-xl border border-border px-5 py-2.5 font-medium text-foreground transition hover:bg-surface-2"
            >
              Sign in
            </Link>
          </div>
        </section>

        <section className="mt-16 grid gap-4 sm:grid-cols-3">
          {[
            {
              title: "Analyze",
              desc: "Server-side MLC classification with LoRA v2 and structured JSON output.",
            },
            {
              title: "Dataset",
              desc: "Browse labeled reviews from Hugging Face Hub and batch-evaluate accuracy.",
            },
            {
              title: "Monitor",
              desc: "Compliance KPIs, quality scores, and decision history in one place.",
            },
          ].map((item) => (
            <article
              key={item.title}
              className="rounded-2xl border border-border bg-surface p-6"
            >
              <h2 className="font-mono text-sm text-accent">{item.title}</h2>
              <p className="mt-2 text-sm text-muted">{item.desc}</p>
            </article>
          ))}
        </section>

        <section className="mt-16 rounded-2xl border border-accent/30 bg-accent/5 p-8">
          <p className="font-mono text-xs uppercase tracking-wider text-accent">
            For teams &amp; companies
          </p>
          <h2 className="mt-2 text-2xl font-medium text-foreground">
            Company access is invite-only
          </h2>
          <p className="mt-3 max-w-xl text-sm text-muted">
            When your organization wants to use InferReview, our team creates
            your company workspace and sends a secure invite link. Open the link,
            sign in or register, and join your team — no self-serve company
            signup required.
          </p>
          <p className="mt-4 font-mono text-xs text-muted">
            Have an invite? Open the link you received (e.g.{" "}
            <span className="text-foreground">/invite/…</span>).
          </p>
        </section>
      </main>
    </div>
  );
}
