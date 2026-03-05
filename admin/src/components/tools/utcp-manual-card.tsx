"use client";

import { Link2, RefreshCw, Trash2, Loader2 } from "lucide-react";
import { motion } from "framer-motion";
import type { UTCPManual } from "@/lib/types";

interface UTCPManualCardProps {
  manual: UTCPManual;
  index?: number;
  onSync: (manual: UTCPManual) => void;
  onDelete: (manual: UTCPManual) => void;
  isSyncing?: boolean;
}

export function UTCPManualCard({
  manual,
  index = 0,
  onSync,
  onDelete,
  isSyncing = false,
}: UTCPManualCardProps) {
  const lastSynced = manual.last_synced_at
    ? new Date(manual.last_synced_at).toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      })
    : "Never";

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.3 }}
      className="group relative overflow-hidden rounded-xl border border-[var(--card-border)] bg-[var(--card)] transition-all hover:-translate-y-0.5 hover:shadow-lg"
      style={{ boxShadow: "var(--shadow-sm)" }}
    >
      <div className="h-[3px] bg-indigo-500" />

      <div className="p-5">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-500/10">
              <Link2 className="h-5 w-5 text-indigo-600" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="font-semibold text-[var(--foreground)]">
                  {manual.name}
                </h3>
                <span className="inline-flex items-center rounded-full bg-indigo-500/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-indigo-600">
                  UTCP
                </span>
              </div>
              <p className="text-xs text-[var(--foreground-muted)]">
                {manual.tools_count} tool{manual.tools_count !== 1 ? "s" : ""}
                {manual.utcp_version ? ` — v${manual.utcp_version}` : ""}
              </p>
            </div>
          </div>
        </div>

        {manual.description && (
          <p className="mt-3 line-clamp-2 text-sm text-[var(--foreground-muted)]">
            {manual.description}
          </p>
        )}

        <p className="mt-2 truncate text-xs text-[var(--foreground-subtle)]">
          {manual.manual_url}
        </p>

        <div className="mt-3 flex items-center gap-2 text-xs text-[var(--foreground-subtle)]">
          <span>Last synced: {lastSynced}</span>
        </div>

        <div className="mt-4 flex items-center gap-2">
          <button
            onClick={() => onSync(manual)}
            disabled={isSyncing}
            className="flex items-center gap-1.5 rounded-lg bg-indigo-500/10 px-3 py-1.5 text-xs font-medium text-indigo-600 transition-colors hover:bg-indigo-500/20 disabled:opacity-50"
          >
            {isSyncing ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <RefreshCw className="h-3.5 w-3.5" />
            )}
            Sync
          </button>
          <button
            onClick={() => onDelete(manual)}
            className="ml-auto flex items-center justify-center rounded-lg p-1.5 text-[var(--foreground-subtle)] transition-colors hover:bg-red-500/10 hover:text-red-500"
            title="Disconnect"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </motion.div>
  );
}
