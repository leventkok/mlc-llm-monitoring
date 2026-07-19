"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";

export default function HomePage() {
  const { user, loading, logout } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.push("/login");
    }
  }, [user, loading, router]);

  if (loading || !user) {
    return (
      <div className="flex min-h-screen items-center justify-center text-gray-500">
        Loading...
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="mx-auto max-w-3xl">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              MLC LLM Monitoring
            </h1>
            <p className="text-sm text-gray-600">Welcome, {user.username}</p>
          </div>
          <button
            onClick={logout}
            className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
          >
            Sign Out
          </button>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <Link
            href="/dashboard"
            className="rounded-xl bg-white p-6 shadow-md transition hover:shadow-lg"
          >
            <h2 className="mb-1 text-lg font-semibold text-gray-900">
              Dashboard
            </h2>
            <p className="text-sm text-gray-600">
              Interact with the LLM and run prompts
            </p>
          </Link>

          <Link
            href="/monitoring"
            className="rounded-xl bg-white p-6 shadow-md transition hover:shadow-lg"
          >
            <h2 className="mb-1 text-lg font-semibold text-gray-900">
              Monitoring
            </h2>
            <p className="text-sm text-gray-600">
              View logged runs and decision scores
            </p>
          </Link>
        </div>
      </div>
    </div>
  );
}
