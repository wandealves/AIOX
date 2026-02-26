"use client";

import { motion } from "framer-motion";
import { Zap, Clock, Activity } from "lucide-react";
import type { QuotaInfo } from "@/lib/types";

interface QuotaChartProps {
  quota: QuotaInfo;
}

function ProgressBar({
  value,
  max,
  label,
}: {
  value: number;
  max: number;
  label: string;
}) {
  const pct = max > 0 ? Math.min((value / max) * 100, 100) : 0;

  let color = "from-green-400 to-green-500";
  if (pct >= 80) color = "from-red-400 to-red-500";
  else if (pct >= 60) color = "from-amber-400 to-amber-500";

  return (
    <div>
      <div className="flex items-center justify-between text-xs text-[var(--foreground-muted)]">
        <span>{label}</span>
        <span>{Math.round(pct)}%</span>
      </div>
      <div className="mt-1.5 h-4 w-full overflow-hidden rounded-full bg-[var(--background-tertiary)]">
        <motion.div
          initial={{ width: 0 }}
          animate={{ width: `${pct}%` }}
          transition={{ duration: 0.8, ease: "easeOut" }}
          className={`h-full rounded-full bg-gradient-to-r ${color}`}
        />
      </div>
    </div>
  );
}

export function QuotaChart({ quota }: QuotaChartProps) {
  const cards = [
    {
      icon: Zap,
      title: "Tokens Today",
      value: quota.tokens_used_today ?? 0,
      max: quota.max_tokens_per_day ?? 0,
      gradient: "from-blue-500 to-blue-600",
    },
    {
      icon: Activity,
      title: "Requests Today",
      value: quota.requests_today ?? 0,
      max: quota.max_requests_per_day ?? 0,
      gradient: "from-green-500 to-green-600",
    },
    {
      icon: Clock,
      title: "Rate Limit (per minute)",
      value: quota.rate_limit_current_minute ?? 0,
      max: quota.rate_limit_requests_per_minute ?? 0,
      gradient: "from-amber-500 to-amber-600",
    },
  ];

  return (
    <div className="grid gap-4 md:grid-cols-3">
      {cards.map((card, i) => (
        <motion.div
          key={card.title}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: i * 0.1, duration: 0.4 }}
          className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-5"
          style={{ boxShadow: "var(--shadow-sm)" }}
        >
          <div className="flex items-center gap-2">
            <div className={`flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br ${card.gradient}`}>
              <card.icon className="h-4 w-4 text-white" />
            </div>
            <span className="text-sm font-medium text-[var(--foreground-muted)]">
              {card.title}
            </span>
          </div>

          <div className="mt-3">
            <span className="text-2xl font-bold text-[var(--foreground)]">
              {card.value.toLocaleString()}
            </span>
            <span className="ml-1.5 text-xs text-[var(--foreground-subtle)]">
              / {card.max.toLocaleString()}
            </span>
          </div>

          <div className="mt-3">
            <ProgressBar
              value={card.value}
              max={card.max}
              label={`${card.value.toLocaleString()} of ${card.max.toLocaleString()}`}
            />
          </div>
        </motion.div>
      ))}
    </div>
  );
}
