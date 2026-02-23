"use client";

import { Bot, MessageSquare, Zap, Activity } from "lucide-react";
import { useAgents } from "@/hooks/use-agents";
import { useQuota } from "@/hooks/use-governance";

export default function DashboardPage() {
  const { data: agentsData } = useAgents(1, 1);
  const { data: quotaData } = useQuota();

  const quota = quotaData?.data;
  const totalAgents = agentsData?.total_count || 0;

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-gray-900">Overview</h2>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          icon={Bot}
          label="Total Agents"
          value={totalAgents}
          color="blue"
        />
        <StatCard
          icon={Zap}
          label="Tokens Used Today"
          value={quota?.tokens_used_today || 0}
          color="amber"
        />
        <StatCard
          icon={Activity}
          label="Requests Today"
          value={quota?.requests_today || 0}
          color="green"
        />
        <StatCard
          icon={MessageSquare}
          label="Rate Limit / min"
          value={`${quota?.rate_limit_current_minute || 0}/${quota?.rate_limit_requests_per_minute || 0}`}
          color="purple"
        />
      </div>
    </div>
  );
}

function StatCard({
  icon: Icon,
  label,
  value,
  color,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: number | string;
  color: string;
}) {
  const colors: Record<string, string> = {
    blue: "bg-blue-50 text-blue-600",
    amber: "bg-amber-50 text-amber-600",
    green: "bg-green-50 text-green-600",
    purple: "bg-purple-50 text-purple-600",
  };

  return (
    <div className="rounded-xl border bg-white p-5">
      <div className="flex items-center gap-3">
        <div
          className={`flex h-10 w-10 items-center justify-center rounded-lg ${colors[color]}`}
        >
          <Icon className="h-5 w-5" />
        </div>
        <div>
          <p className="text-sm text-gray-500">{label}</p>
          <p className="text-xl font-bold text-gray-900">
            {typeof value === "number" ? value.toLocaleString() : value}
          </p>
        </div>
      </div>
    </div>
  );
}
