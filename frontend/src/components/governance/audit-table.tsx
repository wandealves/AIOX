"use client";

import type { AuditLog } from "@/lib/types";

interface AuditTableProps {
  logs: AuditLog[];
  page: number;
  totalCount: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}

function SeverityBadge({ severity }: { severity: string }) {
  const colors: Record<string, string> = {
    info: "bg-blue-50 text-blue-700",
    warn: "bg-amber-50 text-amber-700",
    error: "bg-red-50 text-red-700",
  };
  return (
    <span
      className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
        colors[severity] || "bg-gray-50 text-gray-700"
      }`}
    >
      {severity}
    </span>
  );
}

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
      <div className="overflow-x-auto rounded-xl border bg-white">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-gray-50">
              <th className="px-4 py-3 text-left font-medium text-gray-600">
                Event
              </th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">
                Severity
              </th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">
                Resource
              </th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">
                Details
              </th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">
                Time
              </th>
            </tr>
          </thead>
          <tbody>
            {logs.length === 0 && (
              <tr>
                <td
                  colSpan={5}
                  className="px-4 py-8 text-center text-gray-400"
                >
                  No audit logs found
                </td>
              </tr>
            )}
            {logs.map((log) => (
              <tr key={log.id} className="border-b last:border-0">
                <td className="px-4 py-3 font-medium text-gray-900">
                  {log.event_type}
                </td>
                <td className="px-4 py-3">
                  <SeverityBadge severity={log.severity} />
                </td>
                <td className="px-4 py-3 text-gray-600">
                  {log.resource_type}
                  <span className="ml-1 text-xs text-gray-400">
                    {log.resource_id.substring(0, 8)}...
                  </span>
                </td>
                <td className="max-w-xs truncate px-4 py-3 text-gray-600">
                  {log.details}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-gray-500">
                  {new Date(log.created_at).toLocaleString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between">
          <span className="text-sm text-gray-500">
            Page {page} of {totalPages} ({totalCount} total)
          </span>
          <div className="flex gap-2">
            <button
              onClick={() => onPageChange(page - 1)}
              disabled={page <= 1}
              className="rounded-lg border px-3 py-1.5 text-sm transition-colors hover:bg-gray-50 disabled:opacity-50"
            >
              Previous
            </button>
            <button
              onClick={() => onPageChange(page + 1)}
              disabled={page >= totalPages}
              className="rounded-lg border px-3 py-1.5 text-sm transition-colors hover:bg-gray-50 disabled:opacity-50"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
