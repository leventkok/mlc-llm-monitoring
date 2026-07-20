"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";

export default function HomePage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) router.push("/login");
  }, [user, loading, router]);

  if (loading || !user) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center font-mono text-sm text-muted">
        Loading…
      </div>
    );
  }

  const views = [
    {
      href: "/dashboard",
      title: "Dashboard",
      desc: "Add reviews and run them through the model",
      tag: "analyze",
    },
    {
      href: "/monitoring",
      title: "Monitoring",
      desc: "Track decisions, score quality, watch accuracy",
      tag: "observe",
    },
  ];

  return (
    <div className="mx-auto max-w-6xl px-6 py-12">
      <div className="mb-10">
        <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
          welcome back, {user.username}
        </p>
        <h1 className="mt-3 text-3xl font-medium text-foreground">
          Raw LLM monitoring
        </h1>
        <p className="mt-2 max-w-lg text-muted">
          Classify app-store reviews with a raw language model, then observe and
          score how well it decides.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {views.map((v) => (
          <Link
            key={v.href}
            href={v.href}
            className="group rounded-2xl border border-border bg-surface p-6 transition hover:border-accent"
          >
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-medium text-foreground">{v.title}</h2>
              <span className="font-mono text-xs text-muted">{v.tag}</span>
            </div>
            <p className="mt-2 text-sm text-muted">{v.desc}</p>
            <span className="mt-4 inline-block font-mono text-sm text-accent opacity-0 transition group-hover:opacity-100">
              open →
            </span>
          </Link>
        ))}
      </div>
    </div>
  );
}
