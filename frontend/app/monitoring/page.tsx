"use client";

import ProtectedRoute from "@/components/ProtectedRoute";

export default function MonitoringPage() {
  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50 p-8">
        <div className="mx-auto max-w-5xl">
          <h1 className="mb-2 text-2xl font-bold text-gray-900">Monitoring</h1>
          <p className="mb-8 text-sm text-gray-600">
            Logged LLM runs and decision scores.
          </p>

          <div className="rounded-xl bg-white p-6 shadow-md">
            <p className="text-gray-500">
              Run logs and scoring will appear here once monitoring endpoints
              are ready.
            </p>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
