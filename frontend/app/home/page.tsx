"use client";

import Link from "next/link";
import ProtectedRoute from "@/components/ProtectedRoute";
import { useAuth } from "@/context/AuthContext";

export default function AppHomePage() {
  const { user } = useAuth();

  const views = [
    {
      href: "/dashboard",
      title: "Dashboard",
      desc: "Add reviews, analyze via MLC, and inspect rich classification results",
      tag: "analyze",
    },
    {
      href: "/dataset",
      title: "Dataset",
      desc: "Browse Hugging Face review corpus, export CSV, and batch-evaluate MLC accuracy",
      tag: "hf hub",
    },
    {
      href: "/monitoring",
      title: "Monitoring",
      desc: "Compliance KPIs, quality scores, distributions, and raw LLM output",
      tag: "observe",
    },
    {
      href: "/settings",
      title: "Settings",
      desc: "Update password or delete your account",
      tag: "account",
    },
  ];

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-6xl px-6 py-12">
        <div className="mb-10">
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
            welcome back, {user?.username}
          </p>
          <h1 className="mt-3 text-3xl font-medium text-foreground">
            App Review Monitoring
          </h1>
          <p className="mt-2 max-w-lg text-muted">
            Classify app-store reviews with MLC, view structured rich results on
            the dashboard, and monitor output quality over time.
          </p>
          {user?.account_kind === "company" && user.organization && (
            <p className="mt-3 inline-block rounded-lg border border-accent/30 bg-accent/5 px-3 py-1.5 font-mono text-xs text-accent">
              {user.organization.name} · {user.organization.role.replace("_", " ")}
            </p>
          )}
          {user?.account_kind === "individual" && (
            <p className="mt-3 font-mono text-xs text-muted">
              Individual account
            </p>
          )}
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
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
    </ProtectedRoute>
  );
}
