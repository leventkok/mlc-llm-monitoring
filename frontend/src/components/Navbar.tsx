"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import ThemeToggle from "@/components/ThemeToggle";

export default function Navbar() {
  const { user, logout } = useAuth();
  const pathname = usePathname();

  if (!user || pathname === "/login" || pathname === "/register") {
    return null;
  }

  const links = [
    { href: "/", label: "Home" },
    { href: "/dashboard", label: "Dashboard" },
    { href: "/monitoring", label: "Monitoring" },
    { href: "/settings", label: "Settings" },
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
