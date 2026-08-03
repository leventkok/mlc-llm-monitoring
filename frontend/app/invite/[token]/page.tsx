"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { inviteApi } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { InvitePreview } from "@/types";
import ThemeToggle from "@/components/ThemeToggle";

export default function InvitePage() {
  const params = useParams();
  const token = String(params.token ?? "");
  const router = useRouter();
  const { user, login, loading: authLoading } = useAuth();
  const [preview, setPreview] = useState<InvitePreview | null>(null);
  const [error, setError] = useState("");
  const [accepting, setAccepting] = useState(false);
  const [done, setDone] = useState(false);

  useEffect(() => {
    if (!token) return;
    inviteApi
      .preview(token)
      .then(setPreview)
      .catch(() => setError("Could not load invite"));
  }, [token]);

  async function handleAccept() {
    setAccepting(true);
    setError("");
    try {
      await inviteApi.accept(token);
      await login();
      setDone(true);
      setTimeout(() => router.push("/home"), 800);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not accept invite");
    } finally {
      setAccepting(false);
    }
  }

  const next = encodeURIComponent(`/invite/${token}`);

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-background px-4">
      <div className="absolute right-4 top-4">
        <ThemeToggle />
      </div>

      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
            company invite
          </p>
          <h1 className="mt-3 text-2xl font-medium text-foreground">
            Join your team
          </h1>
        </div>

        <div className="space-y-4 rounded-2xl border border-border bg-surface p-6">
          {!preview && !error && (
            <p className="text-sm text-muted">Loading invite…</p>
          )}

          {preview && (
            <>
              <div>
                <p className="text-xs text-muted">Organization</p>
                <p className="text-lg font-medium text-foreground">
                  {preview.org_name}
                </p>
              </div>
              <div className="flex gap-4 text-sm">
                <div>
                  <p className="text-xs text-muted">Role</p>
                  <p className="font-mono text-foreground">
                    {preview.role.replace("_", " ")}
                  </p>
                </div>
                {preview.email && (
                  <div>
                    <p className="text-xs text-muted">For email</p>
                    <p className="font-mono text-foreground">{preview.email}</p>
                  </div>
                )}
              </div>
              {!preview.valid && (
                <p className="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-500">
                  Invite {preview.reason ?? "invalid"}
                </p>
              )}
            </>
          )}

          {error && (
            <p className="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-500">
              {error}
            </p>
          )}

          {done && (
            <p className="text-sm text-accent">Joined! Redirecting…</p>
          )}

          {preview?.valid && !done && (
            <>
              {authLoading ? (
                <p className="text-sm text-muted">Checking session…</p>
              ) : user ? (
                <button
                  type="button"
                  onClick={() => void handleAccept()}
                  disabled={accepting}
                  className="w-full rounded-lg bg-accent py-2.5 font-medium text-accent-fg disabled:opacity-50"
                >
                  {accepting ? "Joining…" : `Join ${preview.org_name}`}
                </button>
              ) : (
                <div className="space-y-2">
                  <p className="text-sm text-muted">
                    Sign in or create an account to accept this invite.
                  </p>
                  <Link
                    href={`/login?next=${next}`}
                    className="block w-full rounded-lg bg-accent py-2.5 text-center font-medium text-accent-fg"
                  >
                    Sign in
                  </Link>
                  <Link
                    href={`/register?next=${next}`}
                    className="block w-full rounded-lg border border-border py-2.5 text-center text-sm text-foreground"
                  >
                    Create account
                  </Link>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
