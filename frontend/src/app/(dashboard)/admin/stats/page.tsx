"use client";

import { Bot, Users, Activity, BarChart3 } from "lucide-react";
import { motion } from "framer-motion";
import { useAdminStats } from "@/hooks/use-admin";
import { Skeleton } from "@/components/ui/skeleton";

const cardVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: i * 0.1, duration: 0.4, ease: "easeOut" as const },
  }),
};

export default function AdminStatsPage() {
  const { data, isLoading } = useAdminStats();
  const stats = data?.data;

  if (isLoading) {
    return (
      <div>
        <h2 className="mb-6 text-lg font-semibold text-[var(--foreground)]">
          Platform Statistics
        </h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-[100px] w-full rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  const cards = [
    {
      icon: Users,
      label: "Total Users",
      value: stats?.total_users || 0,
      gradient: "from-blue-500 to-blue-600",
    },
    {
      icon: Bot,
      label: "Total Agents",
      value: stats?.total_agents || 0,
      gradient: "from-green-500 to-green-600",
    },
    {
      icon: Activity,
      label: "Total Executions",
      value: stats?.total_executions || 0,
      gradient: "from-amber-500 to-amber-600",
    },
    {
      icon: BarChart3,
      label: "Active Today",
      value: stats?.active_users_today || 0,
      gradient: "from-purple-500 to-purple-600",
    },
  ];

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-[var(--foreground)]">
        Platform Statistics
      </h2>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map((card, i) => (
          <motion.div
            key={card.label}
            custom={i}
            initial="hidden"
            animate="visible"
            variants={cardVariants}
            className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-5 transition-shadow hover:shadow-md"
            style={{ boxShadow: "var(--shadow-sm)" }}
          >
            <div className="flex items-center gap-3">
              <div
                className={`flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br ${card.gradient}`}
              >
                <card.icon className="h-5 w-5 text-white" />
              </div>
              <div>
                <p className="text-sm text-[var(--foreground-muted)]">
                  {card.label}
                </p>
                <p className="text-2xl font-bold text-[var(--foreground)]">
                  {card.value.toLocaleString()}
                </p>
              </div>
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  );
}
