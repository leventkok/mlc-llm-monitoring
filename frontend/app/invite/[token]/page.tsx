"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { inviteApi } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { InvitePreview } from "@/types";
import ThemeToggle from "@/components/ThemeToggle";

function emailsMatch(a: string, b: string) {
  return a.trim().toLowerCase() === b.trim().toLowerCase();
}

export default function InvitePage() {
  const params = useParams();
  const token = String(params.token ?? "");
  const router = useRouter();
  const { user, login, logout, loading: authLoading } = useAuth();
  const [preview, setPreview] = useState<InvitePreview | null>(null);
  const [error, setError] = useState("");
  const [accepting, setAccepting] = useState(false);
  const [done, setDone] = useState(false);
  const autoAcceptStarted = useRef(false);

  useEffect(() => {
    if (!token) return;
    inviteApi
      .preview(token)
      .then(setPreview)
      .catch(() => setError("Could not load invite"));
  }, [token]);

  const handleAccept = useCallback(async () => {
    if (!token || accepting || done) return;
    setAccepting(true);
    setError("");
    try {
      await inviteApi.accept(token);
      await login();
      setDone(true);
      router.replace("/home");
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Could not accept invite";
      if (
        msg.includes("already belong") ||
        msg.includes("already used")
      ) {
        await login();
        router.replace("/home");
        return;
      }
      setError(msg);
    } finally {
      setAccepting(false);
    }
  }, [token, accepting, done, login, router]);

  const lockedEmail = preview?.email?.trim();
  const emailMismatch =
    !!lockedEmail && !!user && !emailsMatch(user.email, lockedEmail);

  useEffect(() => {
    if (authLoading || !user || !preview?.valid || done || accepting) return;
    if (emailMismatch) return;

    if (user.account_kind === "company" && user.organization) {
      router.replace("/home");
      return;
    }

    if (autoAcceptStarted.current) return;
    autoAcceptStarted.current = true;
    void handleAccept();
  }, [
    authLoading,
    user,
    preview,
    done,
    accepting,
    emailMismatch,
    handleAccept,
    router,
  ]);

  const next = encodeURIComponent(`/invite/${token}`);

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-background px-4">
      <div className="absolute right-4 top-4">
        <ThemeToggle />
      </div>
      {user && (
        <Link
          href="/home"
          className="absolute left-4 top-4 text-sm text-muted transition hover:text-foreground"
        >
          Go to app →
        </Link>
      )}

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
                  <div className="sm:col-span-2">
                    <p className="text-xs text-muted">Locked to email</p>
                    <p className="font-mono text-foreground">{preview.email}</p>
                    <p className="mt-1 text-xs text-muted">
                      You must sign in or register with this exact address.
                    </p>
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

          {(accepting || (user && preview?.valid && !done && !emailMismatch && !error)) && (
            <p className="text-sm text-muted">
              {accepting ? "Joining organization…" : "Preparing your workspace…"}
            </p>
          )}

          {preview?.valid && !done && !accepting && (
            <>
              {authLoading ? (
                <p className="text-sm text-muted">Checking session…</p>
              ) : user ? (
                emailMismatch ? (
                  <div className="space-y-2">
                    <p className="rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-300">
                      Signed in as <span className="font-mono">{user.email}</span>, but
                      this invite is for{" "}
                      <span className="font-mono">{lockedEmail}</span>.
                    </p>
                    <Link
                      href={`/login?next=${next}`}
                      className="block w-full rounded-lg bg-accent py-2.5 text-center font-medium text-accent-fg"
                    >
                      Sign in with invited email
                    </Link>
                    <button
                      type="button"
                      onClick={() =>
                        void logout().then(() =>
                          router.push(`/login?next=${next}`),
                        )
                      }
                      className="block w-full rounded-lg border border-border py-2.5 text-center text-sm text-foreground"
                    >
                      Sign out and use another account
                    </button>
                  </div>
                ) : null
              ) : (
                <div className="space-y-2">
                  <p className="text-sm text-muted">
                    Sign in or create an account — we&apos;ll join you to the team
                    automatically.
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

          {preview && !preview.valid && user && (
            <Link
              href="/home"
              className="block w-full rounded-lg border border-border py-2.5 text-center text-sm text-foreground"
            >
              Go to app
            </Link>
          )}
        </div>
      </div>
    </div>
  );
}
