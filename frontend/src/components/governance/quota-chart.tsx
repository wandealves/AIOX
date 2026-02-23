"use client";

import { Zap, Clock, Activity } from "lucide-react";
import type { QuotaInfo } from "@/lib/types";

interface QuotaChartProps {
  quota: QuotaInfo;
}

function ProgressBar({
  value,
  max,
  color,
}: {
  value: number;
  max: number;
  color: string;
}) {
  const pct = max > 0 ? Math.min((value / max) * 100, 100) : 0;
  return (
    <div className="h-3 w-full overflow-hidden rounded-full bg-gray-100">
      <div
        className={`h-full rounded-full transition-all ${color}`}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}

export function QuotaChart({ quota }: QuotaChartProps) {
  const tokenPct =
    quota.max_tokens_per_day > 0
      ? Math.round((quota.tokens_used_today / quota.max_tokens_per_day) * 100)
      : 0;
  const requestPct =
    quota.max_requests_per_day > 0
      ? Math.round((quota.requests_today / quota.max_requests_per_day) * 100)
      : 0;

  return (
    <div className="grid gap-6 md:grid-cols-3">
      {/* Tokens */}
      <div className="rounded-xl border bg-white p-5">
        <div className="flex items-center gap-2 text-sm font-medium text-gray-600">
          <Zap className="h-4 w-4" />
          Tokens Today
        </div>
        <div className="mt-2 text-2xl font-bold text-gray-900">
          {quota.tokens_used_today.toLocaleString()}
        </div>
        <div className="mt-1 text-xs text-gray-500">
          of {quota.max_tokens_per_day.toLocaleString()} ({tokenPct}%)
        </div>
        <div className="mt-3">
          <ProgressBar
            value={quota.tokens_used_today}
            max={quota.max_tokens_per_day}
            color={tokenPct > 80 ? "bg-red-500" : "bg-blue-500"}
          />
        </div>
      </div>

      {/* Requests */}
      <div className="rounded-xl border bg-white p-5">
        <div className="flex items-center gap-2 text-sm font-medium text-gray-600">
          <Activity className="h-4 w-4" />
          Requests Today
        </div>
        <div className="mt-2 text-2xl font-bold text-gray-900">
          {quota.requests_today.toLocaleString()}
        </div>
        <div className="mt-1 text-xs text-gray-500">
          of {quota.max_requests_per_day.toLocaleString()} ({requestPct}%)
        </div>
        <div className="mt-3">
          <ProgressBar
            value={quota.requests_today}
            max={quota.max_requests_per_day}
            color={requestPct > 80 ? "bg-red-500" : "bg-green-500"}
          />
        </div>
      </div>

      {/* Rate Limit */}
      <div className="rounded-xl border bg-white p-5">
        <div className="flex items-center gap-2 text-sm font-medium text-gray-600">
          <Clock className="h-4 w-4" />
          Rate Limit (per minute)
        </div>
        <div className="mt-2 text-2xl font-bold text-gray-900">
          {quota.rate_limit_current_minute}
        </div>
        <div className="mt-1 text-xs text-gray-500">
          of {quota.rate_limit_requests_per_minute} requests/min
        </div>
        <div className="mt-3">
          <ProgressBar
            value={quota.rate_limit_current_minute}
            max={quota.rate_limit_requests_per_minute}
            color="bg-amber-500"
          />
        </div>
      </div>
    </div>
  );
}
