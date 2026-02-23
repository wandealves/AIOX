"use client";

import { Bot, Users, Activity, BarChart3 } from "lucide-react";
import { useAdminStats } from "@/hooks/use-admin";

export default function AdminStatsPage() {
  const { data, isLoading } = useAdminStats();
  const stats = data?.data;

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-gray-900">
        Platform Statistics
      </h2>
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-xl border bg-white p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50">
              <Users className="h-5 w-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Total Users</p>
              <p className="text-xl font-bold text-gray-900">{stats?.total_users || 0}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl border bg-white p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-green-50">
              <Bot className="h-5 w-5 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Total Agents</p>
              <p className="text-xl font-bold text-gray-900">{stats?.total_agents || 0}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl border bg-white p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-50">
              <Activity className="h-5 w-5 text-amber-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Total Executions</p>
              <p className="text-xl font-bold text-gray-900">{stats?.total_executions || 0}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl border bg-white p-5">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-50">
              <BarChart3 className="h-5 w-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Active Today</p>
              <p className="text-xl font-bold text-gray-900">{stats?.active_users_today || 0}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
