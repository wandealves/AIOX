"use client";

import { useState } from "react";
import toast from "react-hot-toast";
import { Plus, Pencil, Trash2 } from "lucide-react";
import {
  useSystemTools,
  useDeleteSystemTool,
  useToggleSystemToolActive,
} from "@/hooks/use-system-tools";
import { SystemToolForm } from "@/components/tools/system-tool-form";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import type { SystemTool } from "@/lib/types";

const categoryVariants: Record<string, "info" | "success" | "warning" | "purple" | "error" | "default"> = {
  search: "info",
  productivity: "success",
  communication: "warning",
  development: "purple",
  "ai-media": "error",
  data: "info",
  utilities: "default",
  custom: "default",
};

const typeVariants: Record<string, "purple" | "info"> = {
  builtin: "purple",
  http: "info",
};

export default function AdminSystemToolsPage() {
  const { data, isLoading } = useSystemTools();
  const deleteTool = useDeleteSystemTool();
  const toggleActive = useToggleSystemToolActive();

  const [formOpen, setFormOpen] = useState(false);
  const [editTool, setEditTool] = useState<SystemTool | undefined>();
  const [pendingDelete, setPendingDelete] = useState<SystemTool | null>(null);

  const tools = data?.data || [];

  const handleCreate = () => {
    setEditTool(undefined);
    setFormOpen(true);
  };

  const handleEdit = (tool: SystemTool) => {
    setEditTool(tool);
    setFormOpen(true);
  };

  const handleCloseForm = () => {
    setFormOpen(false);
    setEditTool(undefined);
  };

  const handleConfirmDelete = async () => {
    if (!pendingDelete) return;
    try {
      await deleteTool.mutateAsync(pendingDelete.id);
      toast.success("System tool deleted");
    } catch {
      toast.error("Failed to delete tool");
    }
    setPendingDelete(null);
  };

  const handleToggleActive = async (tool: SystemTool) => {
    try {
      await toggleActive.mutateAsync({
        toolId: tool.id,
        isActive: !tool.is_active,
      });
      toast.success(tool.is_active ? "Tool deactivated" : "Tool activated");
    } catch {
      toast.error("Failed to toggle tool status");
    }
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-[var(--foreground)]">
          System Tools ({tools.length})
        </h2>
        <button
          onClick={handleCreate}
          className="flex items-center gap-2 rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 px-4 py-2.5 text-sm font-semibold text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98]"
          style={{ boxShadow: "var(--shadow-sm)" }}
        >
          <Plus className="h-4 w-4" />
          Add Tool
        </button>
      </div>

      <div
        className="overflow-x-auto rounded-xl border border-[var(--card-border)] bg-[var(--card)]"
        style={{ boxShadow: "var(--shadow-sm)" }}
      >
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--card-border)] bg-[var(--background-tertiary)]">
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Name
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Category
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Type
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Version
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Status
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {isLoading &&
              Array.from({ length: 5 }).map((_, i) => (
                <tr key={i} className="border-b border-[var(--card-border)]">
                  {Array.from({ length: 6 }).map((_, j) => (
                    <td key={j} className="px-4 py-3">
                      <Skeleton className="h-5 w-full" variant="text" />
                    </td>
                  ))}
                </tr>
              ))}
            {!isLoading && tools.length === 0 && (
              <tr>
                <td
                  colSpan={6}
                  className="px-4 py-12 text-center text-[var(--foreground-muted)]"
                >
                  No system tools yet. Click &quot;Add Tool&quot; to create one.
                </td>
              </tr>
            )}
            {!isLoading &&
              tools.map((tool, i) => (
                <tr
                  key={tool.id}
                  className={`border-b border-[var(--card-border)] last:border-0 transition-colors hover:bg-[var(--card-hover)] ${
                    i % 2 === 1 ? "bg-[var(--background-secondary)]" : ""
                  }`}
                >
                  <td className="px-4 py-3">
                    <div>
                      <span className="font-medium text-[var(--foreground)]">
                        {tool.name}
                      </span>
                      <p className="mt-0.5 line-clamp-1 text-xs text-[var(--foreground-subtle)]">
                        {tool.description}
                      </p>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <Badge variant={categoryVariants[tool.category] || "default"}>
                      {tool.category}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <Badge variant={typeVariants[tool.tool_type] || "info"}>
                      {tool.tool_type}
                    </Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-[var(--foreground-subtle)]">
                    {tool.version || "-"}
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleToggleActive(tool)}
                      className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors ${
                        tool.is_active
                          ? "bg-primary-600"
                          : "bg-[var(--background-tertiary)]"
                      }`}
                      title={tool.is_active ? "Active" : "Inactive"}
                    >
                      <span
                        className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-sm transition-transform ${
                          tool.is_active ? "translate-x-5" : "translate-x-0"
                        }`}
                      />
                    </button>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleEdit(tool)}
                        className="flex items-center gap-1.5 rounded-lg bg-primary-500/10 px-3 py-1.5 text-xs font-medium text-primary-600 transition-colors hover:bg-primary-500/20"
                      >
                        <Pencil className="h-3.5 w-3.5" />
                        Edit
                      </button>
                      <button
                        onClick={() => setPendingDelete(tool)}
                        className="flex items-center justify-center rounded-lg p-1.5 text-[var(--foreground-subtle)] transition-colors hover:bg-red-500/10 hover:text-red-500"
                        title="Delete tool"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>

      {/* Create/Edit Dialog */}
      {formOpen && (
        <SystemToolForm tool={editTool} onClose={handleCloseForm} />
      )}

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={!!pendingDelete}
        title="Delete System Tool"
        message={`Are you sure you want to delete "${pendingDelete?.name}"? This action cannot be undone.`}
        variant="danger"
        confirmLabel="Delete"
        onConfirm={handleConfirmDelete}
        onCancel={() => setPendingDelete(null)}
      />
    </div>
  );
}
