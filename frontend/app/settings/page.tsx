"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import ProtectedRoute from "@/components/ProtectedRoute";
import { useAuth } from "@/context/AuthContext";
import { authApi, orgApi } from "@/lib/api";
import type { OrgInvite, OrgMember } from "@/types";

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

  const [members, setMembers] = useState<OrgMember[]>([]);
  const [invites, setInvites] = useState<OrgInvite[]>([]);
  const [inviteRole, setInviteRole] = useState("company_member");
  const [inviteEmail, setInviteEmail] = useState("");
  const [orgMsg, setOrgMsg] = useState("");
  const [orgError, setOrgError] = useState("");
  const [orgBusy, setOrgBusy] = useState(false);

  const isCompany = user?.account_kind === "company" && user.organization;
  const isCompanyAdmin = isCompany && user?.organization?.role === "company_admin";

  useEffect(() => {
    if (!isCompany) return;
    orgApi
      .listMembers()
      .then(setMembers)
      .catch(() => setMembers([]));
  }, [isCompany]);

  useEffect(() => {
    if (!isCompanyAdmin) return;
    orgApi
      .listInvites()
      .then(setInvites)
      .catch(() => setInvites([]));
  }, [isCompanyAdmin]);

  function inviteURL(path: string) {
    if (typeof window === "undefined") return path;
    return `${window.location.origin}${path}`;
  }

  async function handleCreateInvite(e: React.FormEvent) {
    e.preventDefault();
    if (!isCompanyAdmin) return;
    setOrgBusy(true);
    setOrgMsg("");
    setOrgError("");
    try {
      const inv = await orgApi.createInvite({
        role: inviteRole,
        email: inviteEmail.trim(),
        days: 14,
      });
      setInvites((prev) => [inv, ...prev]);
      setInviteEmail("");
      setOrgMsg("Invite link created — share it with your teammate.");
    } catch (err) {
      setOrgError(err instanceof Error ? err.message : "Could not create invite");
    } finally {
      setOrgBusy(false);
    }
  }

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
              <div>
                <dt className="text-muted">Account type</dt>
                <dd className="font-medium capitalize text-foreground">
                  {user.account_kind ?? "individual"}
                </dd>
              </div>
              {user.platform_role === "platform_admin" && (
                <div>
                  <dt className="text-muted">Platform role</dt>
                  <dd className="font-medium text-accent">platform admin</dd>
                </div>
              )}
            </dl>
          </div>
        )}

        {isCompany && user.organization && (
          <div className="mb-6 rounded-2xl border border-border bg-surface p-6">
            <p className="font-mono text-xs uppercase tracking-wider text-muted">
              organization
            </p>
            <dl className="mt-4 space-y-3 text-sm">
              <div>
                <dt className="text-muted">Company</dt>
                <dd className="font-medium text-foreground">{user.organization.name}</dd>
              </div>
              <div>
                <dt className="text-muted">Slug</dt>
                <dd className="font-mono text-foreground">{user.organization.slug}</dd>
              </div>
              <div>
                <dt className="text-muted">Your role</dt>
                <dd className="font-medium text-foreground">
                  {user.organization.role.replace("_", " ")}
                </dd>
              </div>
            </dl>

            {members.length > 0 && (
              <div className="mt-6 border-t border-border pt-4">
                <p className="font-mono text-xs uppercase tracking-wider text-muted">
                  team ({members.length})
                </p>
                <ul className="mt-3 space-y-2 text-sm">
                  {members.map((m) => (
                    <li
                      key={m.user_id}
                      className="flex items-center justify-between rounded-lg border border-border bg-background px-3 py-2"
                    >
                      <span className="text-foreground">{m.username}</span>
                      <span className="font-mono text-[11px] text-muted">
                        {m.role.replace("_", " ")}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {isCompanyAdmin && (
              <div className="mt-6 border-t border-border pt-4">
                <p className="font-mono text-xs uppercase tracking-wider text-muted">
                  invite teammates
                </p>
                <form onSubmit={(e) => void handleCreateInvite(e)} className="mt-3 space-y-3">
                  <label className="block text-xs text-muted">
                    Role
                    <select
                      value={inviteRole}
                      onChange={(e) => setInviteRole(e.target.value)}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                    >
                      <option value="company_admin">company admin</option>
                      <option value="company_member">company member</option>
                    </select>
                  </label>
                  <label className="block text-xs text-muted">
                    Email (optional lock)
                    <input
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      placeholder="teammate@company.com"
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                    />
                  </label>
                  {orgError && <p className="text-sm text-red-500">{orgError}</p>}
                  {orgMsg && <p className="text-sm text-emerald-500">{orgMsg}</p>}
                  <button
                    type="submit"
                    disabled={orgBusy}
                    className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-accent-fg disabled:opacity-50"
                  >
                    Generate invite link
                  </button>
                </form>

                {invites.length > 0 && (
                  <ul className="mt-4 space-y-2">
                    {invites.map((inv) => (
                      <li
                        key={inv.id}
                        className="rounded-lg border border-border bg-background px-3 py-2 font-mono text-[11px]"
                      >
                        <p className="text-muted">
                          {inv.role} · expires{" "}
                          {new Date(inv.expires_at).toLocaleDateString()}
                          {inv.used_at ? " · used" : ""}
                        </p>
                        <p className="mt-1 break-all text-accent">
                          {inviteURL(inv.invite_path)}
                        </p>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}
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
