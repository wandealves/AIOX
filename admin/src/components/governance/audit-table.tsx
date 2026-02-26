"use client";

import { formatDistanceToNow } from "date-fns";
import { Badge } from "@/components/ui/badge";
import { Pagination } from "@/components/ui/pagination";
import type { AuditLog } from "@/lib/types";

interface AuditTableProps {
  logs: AuditLog[];
  page: number;
  totalCount: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}

const severityVariant: Record<string, "info" | "warning" | "error"> = {
  info: "info",
  warn: "warning",
  error: "error",
};

export function AuditTable({
  logs,
  page,
  totalCount,
  pageSize,
  onPageChange,
}: AuditTableProps) {
  const totalPages = Math.ceil(totalCount / pageSize);

  return (
    <div>
      <div
        className="overflow-x-auto rounded-xl border border-[var(--card-border)] bg-[var(--card)]"
        style={{ boxShadow: "var(--shadow-sm)" }}
      >
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--card-border)] bg-[var(--background-tertiary)]">
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Event
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Severity
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Resource
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Details
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Time
              </th>
            </tr>
          </thead>
          <tbody>
            {logs.length === 0 && (
              <tr>
                <td
                  colSpan={5}
                  className="px-4 py-12 text-center text-[var(--foreground-subtle)]"
                >
                  No audit logs found
                </td>
              </tr>
            )}
            {logs.map((log, i) => (
              <tr
                key={log.id}
                className={`border-b border-[var(--card-border)] last:border-0 transition-colors hover:bg-[var(--card-hover)] ${
                  i % 2 === 1 ? "bg-[var(--background-secondary)]" : ""
                }`}
              >
                <td className="px-4 py-3 font-medium text-[var(--foreground)]">
                  {log.event_type}
                </td>
                <td className="px-4 py-3">
                  <Badge variant={severityVariant[log.severity] || "default"}>
                    {log.severity}
                  </Badge>
                </td>
                <td className="px-4 py-3 text-[var(--foreground-muted)]">
                  {log.resource_type}
                  <span className="ml-1 text-xs text-[var(--foreground-subtle)]">
                    {log.resource_id.substring(0, 8)}...
                  </span>
                </td>
                <td className="max-w-xs truncate px-4 py-3 text-[var(--foreground-muted)]">
                  {log.details}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-[var(--foreground-subtle)]">
                  <span
                    title={new Date(log.created_at).toLocaleString()}
                  >
                    {(() => {
                      try {
                        return formatDistanceToNow(new Date(log.created_at), {
                          addSuffix: true,
                        });
                      } catch {
                        return log.created_at;
                      }
                    })()}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <Pagination
        page={page}
        totalPages={totalPages}
        onPageChange={onPageChange}
      />
    </div>
  );
}
