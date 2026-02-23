"use client";

import { useState } from "react";
import { useAuditLogs } from "@/hooks/use-governance";
import { AuditTable } from "@/components/governance/audit-table";

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
      <h2 className="mb-6 text-lg font-semibold text-gray-900">Audit Logs</h2>

      <div className="mb-4 flex gap-3">
        <select
          value={eventType}
          onChange={(e) => {
            setEventType(e.target.value);
            setPage(1);
          }}
          className="rounded-lg border border-gray-200 px-3 py-2 text-sm"
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
          className="rounded-lg border border-gray-200 px-3 py-2 text-sm"
        >
          <option value="">All Severities</option>
          <option value="info">Info</option>
          <option value="warn">Warning</option>
          <option value="error">Error</option>
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
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
