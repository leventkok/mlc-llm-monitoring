"use client";

import ProtectedRoute from "@/components/ProtectedRoute";

export default function DashboardPage() {
  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50 p-8">
        <div className="mx-auto max-w-5xl">
          <h1 className="mb-2 text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="mb-8 text-sm text-gray-600">
            Interact with the raw LLM and run prompts.
          </p>

          <div className="rounded-xl bg-white p-6 shadow-md">
            <p className="text-gray-500">
              LLM interaction will be built here once the model integration is
              ready.
            </p>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
