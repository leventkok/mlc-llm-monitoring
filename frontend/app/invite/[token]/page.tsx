"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { authApi, inviteApi } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { InvitePreview } from "@/types";
import ThemeToggle from "@/components/ThemeToggle";

function emailsMatch(a: string, b: string) {
  return a.trim().toLowerCase() === b.trim().toLowerCase();
}

type AuthMode = "signin" | "register";

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

  const [authMode, setAuthMode] = useState<AuthMode>("signin");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [authBusy, setAuthBusy] = useState(false);

  useEffect(() => {
    if (!token) return;
    inviteApi
      .preview(token)
      .then((p) => {
        setPreview(p);
        if (p.email) setEmail(p.email);
      })
      .catch(() => setError("Could not load invite"));
  }, [token]);

  const joinOrganization = useCallback(async () => {
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
      if (msg.includes("already belong") || msg.includes("already used")) {
        await login();
        router.replace("/home");
        return;
      }
      setError(msg);
      autoAcceptStarted.current = false;
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
    void joinOrganization();
  }, [
    authLoading,
    user,
    preview,
    done,
    accepting,
    emailMismatch,
    joinOrganization,
    router,
  ]);

  async function handleSignIn(e: React.FormEvent) {
    e.preventDefault();
    setAuthBusy(true);
    setError("");
    autoAcceptStarted.current = true;
    try {
      await authApi.login({ email, password });
      await login();
      await joinOrganization();
    } catch (err) {
      autoAcceptStarted.current = false;
      setError(err instanceof Error ? err.message : "Sign in failed");
    } finally {
      setAuthBusy(false);
    }
  }

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault();
    setAuthBusy(true);
    setError("");
    autoAcceptStarted.current = true;
    try {
      await authApi.register({ email, username, password });
      await authApi.login({ email, password });
      await login();
      await joinOrganization();
    } catch (err) {
      autoAcceptStarted.current = false;
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setAuthBusy(false);
    }
  }

  const showGuestAuth =
    preview?.valid && !done && !user && !authLoading && !accepting && !authBusy;

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

          {(accepting || authBusy) && (
            <p className="text-sm text-muted">
              {authBusy ? "Signing you in…" : "Joining organization…"}
            </p>
          )}

          {showGuestAuth && (
            <div className="space-y-4 border-t border-border pt-4">
              <p className="text-sm text-muted">
                Sign in or create an account on this page — you&apos;ll join{" "}
                {preview?.org_name} automatically.
              </p>

              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => setAuthMode("signin")}
                  className={`flex-1 rounded-lg px-3 py-2 text-sm ${
                    authMode === "signin"
                      ? "bg-accent text-accent-fg"
                      : "border border-border text-muted"
                  }`}
                >
                  Sign in
                </button>
                <button
                  type="button"
                  onClick={() => setAuthMode("register")}
                  className={`flex-1 rounded-lg px-3 py-2 text-sm ${
                    authMode === "register"
                      ? "bg-accent text-accent-fg"
                      : "border border-border text-muted"
                  }`}
                >
                  Create account
                </button>
              </div>

              {authMode === "signin" ? (
                <form onSubmit={(e) => void handleSignIn(e)} className="space-y-3">
                  <label className="block text-xs text-muted">
                    Email
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      readOnly={!!lockedEmail}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                      required
                    />
                  </label>
                  <label className="block text-xs text-muted">
                    Password
                    <input
                      type="password"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                      required
                    />
                  </label>
                  <button
                    type="submit"
                    disabled={authBusy}
                    className="w-full rounded-lg bg-accent py-2.5 font-medium text-accent-fg disabled:opacity-50"
                  >
                    Sign in and join
                  </button>
                </form>
              ) : (
                <form onSubmit={(e) => void handleRegister(e)} className="space-y-3">
                  <label className="block text-xs text-muted">
                    Email
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      readOnly={!!lockedEmail}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                      required
                    />
                  </label>
                  <label className="block text-xs text-muted">
                    Username
                    <input
                      type="text"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                      required
                    />
                  </label>
                  <label className="block text-xs text-muted">
                    Password
                    <input
                      type="password"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      minLength={12}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                      required
                    />
                  </label>
                  <button
                    type="submit"
                    disabled={authBusy}
                    className="w-full rounded-lg bg-accent py-2.5 font-medium text-accent-fg disabled:opacity-50"
                  >
                    Create account and join
                  </button>
                </form>
              )}
            </div>
          )}

          {preview?.valid && !done && user && emailMismatch && !accepting && (
            <div className="space-y-2">
              <p className="rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-300">
                Signed in as <span className="font-mono">{user.email}</span>, but
                this invite is for{" "}
                <span className="font-mono">{lockedEmail}</span>.
              </p>
              <button
                type="button"
                onClick={() =>
                  void logout().then(() => {
                    setEmail(lockedEmail ?? "");
                    setPassword("");
                  })
                }
                className="block w-full rounded-lg border border-border py-2.5 text-center text-sm text-foreground"
              >
                Sign out and use invited email
              </button>
            </div>
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
