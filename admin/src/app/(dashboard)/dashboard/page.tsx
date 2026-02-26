"use client";

import { Bot, MessageSquare, Zap, Activity } from "lucide-react";
import { motion } from "framer-motion";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { useAgents } from "@/hooks/use-agents";
import { useQuota } from "@/hooks/use-governance";
import { Skeleton } from "@/components/ui/skeleton";

const mockChartData = [
  { day: "Mon", tokens: 0, requests: 0 },
  { day: "Tue", tokens: 0, requests: 0 },
  { day: "Wed", tokens: 0, requests: 0 },
  { day: "Thu", tokens: 0, requests: 0 },
  { day: "Fri", tokens: 0, requests: 0 },
  { day: "Sat", tokens: 0, requests: 0 },
  { day: "Sun", tokens: 0, requests: 0 },
];

const cardVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: i * 0.1, duration: 0.4, ease: "easeOut" as const },
  }),
};

export default function DashboardPage() {
  const { data: agentsData, isLoading: agentsLoading } = useAgents(1, 1);
  const { data: quotaData, isLoading: quotaLoading } = useQuota();

  const isLoading = agentsLoading || quotaLoading;
  const quota = quotaData?.data;
  const totalAgents = agentsData?.total_count || 0;

  const stats = [
    {
      icon: Bot,
      label: "Total Agents",
      value: totalAgents,
      gradient: "from-blue-500 to-blue-600",
      bg: "bg-blue-500/10",
    },
    {
      icon: Zap,
      label: "Tokens Used Today",
      value: quota?.tokens_used_today || 0,
      gradient: "from-amber-500 to-amber-600",
      bg: "bg-amber-500/10",
    },
    {
      icon: Activity,
      label: "Requests Today",
      value: quota?.requests_today || 0,
      gradient: "from-green-500 to-green-600",
      bg: "bg-green-500/10",
    },
    {
      icon: MessageSquare,
      label: "Rate Limit / min",
      value: `${quota?.rate_limit_current_minute || 0}/${quota?.rate_limit_requests_per_minute || 0}`,
      gradient: "from-purple-500 to-purple-600",
      bg: "bg-purple-500/10",
    },
  ];

  if (isLoading) {
    return (
      <div>
        <h2 className="mb-6 text-lg font-semibold text-[var(--foreground)]">
          Overview
        </h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-[100px] w-full rounded-xl" />
          ))}
        </div>
        <Skeleton className="mt-6 h-[320px] w-full rounded-xl" />
      </div>
    );
  }

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-[var(--foreground)]">
        Overview
      </h2>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat, i) => (
          <motion.div
            key={stat.label}
            custom={i}
            initial="hidden"
            animate="visible"
            variants={cardVariants}
            className="group rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-5 transition-shadow hover:shadow-md"
            style={{ boxShadow: "var(--shadow-sm)" }}
          >
            <div className="flex items-center gap-3">
              <div
                className={`flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br ${stat.gradient}`}
              >
                <stat.icon className="h-5 w-5 text-white" />
              </div>
              <div>
                <p className="text-sm text-[var(--foreground-muted)]">
                  {stat.label}
                </p>
                <p className="text-2xl font-bold text-[var(--foreground)]">
                  {typeof stat.value === "number"
                    ? stat.value.toLocaleString()
                    : stat.value}
                </p>
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Usage chart */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4, duration: 0.4 }}
        className="mt-6 rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-5"
        style={{ boxShadow: "var(--shadow-sm)" }}
      >
        <h3 className="mb-4 text-sm font-semibold text-[var(--foreground)]">
          Usage Over Time
        </h3>
        <div className="h-[260px]">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={mockChartData}>
              <defs>
                <linearGradient id="tokenGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="var(--primary-500)" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="var(--primary-500)" stopOpacity={0.05} />
                </linearGradient>
              </defs>
              <CartesianGrid
                strokeDasharray="3 3"
                stroke="var(--card-border)"
                vertical={false}
              />
              <XAxis
                dataKey="day"
                tick={{ fontSize: 12, fill: "var(--foreground-muted)" }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis
                tick={{ fontSize: 12, fill: "var(--foreground-muted)" }}
                axisLine={false}
                tickLine={false}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: "var(--card)",
                  border: "1px solid var(--card-border)",
                  borderRadius: "var(--radius-lg)",
                  boxShadow: "var(--shadow-md)",
                  color: "var(--foreground)",
                }}
              />
              <Area
                type="monotone"
                dataKey="tokens"
                stroke="var(--primary-500)"
                strokeWidth={2}
                fill="url(#tokenGradient)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </motion.div>
    </div>
  );
}
