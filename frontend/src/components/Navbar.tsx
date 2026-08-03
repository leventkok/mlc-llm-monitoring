"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import ThemeToggle from "@/components/ThemeToggle";

export default function Navbar() {
  const { user, logout } = useAuth();
  const pathname = usePathname();

  if (!user || pathname === "/" || pathname === "/login" || pathname === "/register" || pathname.startsWith("/invite/")) {
    return null;
  }

  const links = [
    { href: "/home", label: "Home" },
    { href: "/dashboard", label: "Dashboard" },
    ...(user.is_admin ? [{ href: "/dataset", label: "Dataset" }] : []),
    { href: "/deepkwiki", label: "DeepKwiki" },
    { href: "/monitoring", label: "Monitoring" },
    { href: "/settings", label: "Settings" },
    ...(user.is_admin ? [{ href: "/admin", label: "Admin" }] : []),
  ];

  return (
    <nav className="sticky top-0 z-10 border-b border-border bg-background/80 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
        <div className="flex items-center gap-6">
          <Link
            href="/"
            className="font-mono text-sm font-medium tracking-tight text-foreground"
          >
            app<span className="text-accent">·</span>review
          </Link>
          <div className="flex gap-1">
            {links.map((link) => {
              const active = pathname === link.href;
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={`rounded-lg px-3 py-1.5 text-sm font-medium transition ${
                    active
                      ? "bg-surface-2 text-foreground"
                      : "text-muted hover:text-foreground"
                  }`}
                >
                  {link.label}
                </Link>
              );
            })}
          </div>
        </div>

        <div className="flex items-center gap-3">
          {user.organization && (
            <span className="hidden font-mono text-[10px] text-muted sm:inline">
              {user.organization.name}
            </span>
          )}
          <span className="font-mono text-xs text-muted">{user.username}</span>
          <ThemeToggle />
          <button
            onClick={() => void logout()}
            className="rounded-lg border border-border px-3 py-1.5 text-sm font-medium text-muted transition hover:text-foreground"
          >
            Sign out
          </button>
        </div>
      </div>
    </nav>
  );
}
