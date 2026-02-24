"use client";

import { useState } from "react";
import { useAuditLogs } from "@/hooks/use-governance";
import { AuditTable } from "@/components/governance/audit-table";
import { Skeleton } from "@/components/ui/skeleton";

export default function AuditPage() {
  const [page, setPage] = useState(1);
  const [eventType, setEventType] = useState("");
  const [severity, setSeverity] = useState("");
  const { data, isLoading } = useAuditLogs(page, 20, {
    event_type: eventType || undefined,
    severity: severity || undefined,
  });

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-[var(--foreground)]">
        Audit Logs
      </h2>

      <div className="mb-4 flex flex-wrap gap-3">
        <select
          value={eventType}
          onChange={(e) => {
            setEventType(e.target.value);
            setPage(1);
          }}
          className="rounded-lg border border-[var(--input-border)] bg-[var(--input-bg)] px-3 py-2 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
        >
          <option value="">All Events</option>
          <option value="message_routed">Message Routed</option>
          <option value="task_completed">Task Completed</option>
          <option value="task_failed">Task Failed</option>
        </select>

        <select
          value={severity}
          onChange={(e) => {
            setSeverity(e.target.value);
            setPage(1);
          }}
          className="rounded-lg border border-[var(--input-border)] bg-[var(--input-bg)] px-3 py-2 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
        >
          <option value="">All Severities</option>
          <option value="info">Info</option>
          <option value="warn">Warning</option>
          <option value="error">Error</option>
        </select>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          <Skeleton className="h-12 w-full rounded-lg" />
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-14 w-full rounded-lg" />
          ))}
        </div>
      ) : (
        <AuditTable
          logs={data?.data || []}
          page={page}
          totalCount={data?.total_count || 0}
          pageSize={20}
          onPageChange={setPage}
        />
      )}
    </div>
  );
}
