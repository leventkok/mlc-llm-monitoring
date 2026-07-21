"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import ProtectedRoute from "@/components/ProtectedRoute";
import { useAuth } from "@/context/AuthContext";
import { authApi } from "@/lib/api";

export default function SettingsPage() {
  const { user, logout } = useAuth();
  const router = useRouter();
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [passwordMsg, setPasswordMsg] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [deleteError, setDeleteError] = useState("");
  const [deleteLoading, setDeleteLoading] = useState(false);

  async function handleChangePassword(e: React.FormEvent) {
    e.preventDefault();
    setPasswordMsg("");
    setPasswordError("");
    setPasswordLoading(true);
    try {
      await authApi.changePassword(oldPassword, newPassword);
      setPasswordMsg("Password updated. Please sign in again.");
      setOldPassword("");
      setNewPassword("");
      await logout();
      router.push("/login");
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : "Could not update password");
    } finally {
      setPasswordLoading(false);
    }
  }

  async function handleDeleteAccount() {
    const confirmed = window.confirm(
      "Delete your account permanently? All reviews, decisions, and scores will be removed.",
    );
    if (!confirmed) return;

    setDeleteError("");
    setDeleteLoading(true);
    try {
      await authApi.deleteAccount();
      await logout();
      router.push("/register");
    } catch (err) {
      setDeleteError(err instanceof Error ? err.message : "Could not delete account");
    } finally {
      setDeleteLoading(false);
    }
  }

  return (
    <ProtectedRoute>
      <div className="mx-auto max-w-lg px-6 py-10">
        <div className="mb-8">
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">
            settings
          </p>
          <h1 className="mt-2 text-2xl font-medium text-foreground">Account</h1>
          <p className="mt-1 text-sm text-muted">
            Manage your profile, password, and account data.
          </p>
        </div>

        {user && (
          <div className="mb-6 rounded-2xl border border-border bg-surface p-6">
            <p className="font-mono text-xs uppercase tracking-wider text-muted">
              profile
            </p>
            <dl className="mt-4 space-y-3 text-sm">
              <div>
                <dt className="text-muted">Email</dt>
                <dd className="font-medium text-foreground">{user.email}</dd>
              </div>
              <div>
                <dt className="text-muted">Username</dt>
                <dd className="font-medium text-foreground">{user.username}</dd>
              </div>
            </dl>
          </div>
        )}

        <form
          onSubmit={handleChangePassword}
          className="mb-6 space-y-4 rounded-2xl border border-border bg-surface p-6"
        >
          <p className="font-mono text-xs uppercase tracking-wider text-muted">
            change password
          </p>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-foreground">
              Current password
            </label>
            <input
              type="password"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-foreground outline-none transition focus:border-accent"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-foreground">
              New password
            </label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              minLength={12}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-foreground outline-none transition focus:border-accent"
              required
            />
          </div>
          {passwordError && (
            <p className="text-sm text-red-500">{passwordError}</p>
          )}
          {passwordMsg && (
            <p className="text-sm text-emerald-500">{passwordMsg}</p>
          )}
          <button
            type="submit"
            disabled={passwordLoading}
            className="rounded-lg bg-accent px-4 py-2 font-medium text-accent-fg transition hover:opacity-90 disabled:opacity-50"
          >
            {passwordLoading ? "Updating…" : "Update password"}
          </button>
        </form>

        <div className="rounded-2xl border border-red-500/30 bg-red-500/5 p-6">
          <p className="font-mono text-xs uppercase tracking-wider text-red-500">
            danger zone
          </p>
          <p className="mt-2 text-sm text-muted">
            Permanently delete your account and all associated data.
          </p>
          {deleteError && (
            <p className="mt-3 text-sm text-red-500">{deleteError}</p>
          )}
          <button
            type="button"
            onClick={handleDeleteAccount}
            disabled={deleteLoading}
            className="mt-4 rounded-lg border border-red-500/50 px-4 py-2 text-sm font-medium text-red-500 transition hover:bg-red-500/10 disabled:opacity-50"
          >
            {deleteLoading ? "Deleting…" : "Delete account"}
          </button>
        </div>
      </div>
    </ProtectedRoute>
  );
}
