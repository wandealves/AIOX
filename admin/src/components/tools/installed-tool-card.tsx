"use client";

import { Pencil, Trash2, Globe } from "lucide-react";
import { motion } from "framer-motion";
import {
  TOOL_ICONS,
  TOOL_TEMPLATES,
  TOOL_CATEGORIES,
  CATEGORY_COLORS,
} from "@/lib/tool-templates";
import type { ToolDefinition } from "@/lib/types";

interface InstalledToolCardProps {
  tool: ToolDefinition;
  index?: number;
  onEdit?: (tool: ToolDefinition) => void;
  onDelete?: (tool: ToolDefinition) => void;
  onToggleActive?: (tool: ToolDefinition) => void;
}

const methodColors: Record<string, string> = {
  GET: "bg-green-500/10 text-green-600",
  POST: "bg-blue-500/10 text-blue-600",
  PUT: "bg-yellow-500/10 text-yellow-600",
  DELETE: "bg-red-500/10 text-red-600",
  PATCH: "bg-purple-500/10 text-purple-600",
};

export function InstalledToolCard({
  tool,
  index = 0,
  onEdit,
  onDelete,
  onToggleActive,
}: InstalledToolCardProps) {
  // Try to match tool name to a template for icon/color resolution
  const matchedTemplate = TOOL_TEMPLATES.find((t) => t.id === tool.name);
  const Icon = matchedTemplate
    ? TOOL_ICONS[matchedTemplate.iconName]
    : Globe;
  const categoryInfo = matchedTemplate
    ? TOOL_CATEGORIES[matchedTemplate.category]
    : null;
  const colors = categoryInfo
    ? CATEGORY_COLORS[categoryInfo.color]
    : { bg: "bg-primary-500/10", text: "text-primary-500", accent: "bg-primary-500" };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.3 }}
      className="group relative overflow-hidden rounded-xl border border-[var(--card-border)] bg-[var(--card)] transition-all hover:-translate-y-0.5 hover:shadow-lg"
      style={{ boxShadow: "var(--shadow-sm)" }}
    >
      <div className={`h-[3px] ${colors.accent}`} />

      <div className="p-5">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div
              className={`flex h-10 w-10 items-center justify-center rounded-xl ${colors.bg}`}
            >
              {Icon && <Icon className={`h-5 w-5 ${colors.text}`} />}
            </div>
            <div>
              <h3 className="font-semibold text-[var(--foreground)]">
                {matchedTemplate?.name || tool.name}
              </h3>
              <div className="flex items-center gap-2">
                <span
                  className={`inline-block rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${
                    methodColors[tool.http_method] ||
                    "bg-[var(--background-tertiary)] text-[var(--foreground-muted)]"
                  }`}
                >
                  {tool.http_method}
                </span>
                {tool.auth_type && tool.auth_type !== "none" && (
                  <span className="rounded-full bg-[var(--background-tertiary)] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-[var(--foreground-muted)]">
                    {tool.auth_type}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Active toggle */}
          <button
            onClick={() => onToggleActive?.(tool)}
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
        </div>

        <p className="mt-3 line-clamp-2 text-sm text-[var(--foreground-muted)]">
          {tool.description || "No description"}
        </p>

        <p className="mt-2 truncate text-xs text-[var(--foreground-subtle)]">
          {tool.endpoint_url}
        </p>

        <div className="mt-4 flex items-center gap-2">
          {onEdit && (
            <button
              onClick={() => onEdit(tool)}
              className="flex items-center gap-1.5 rounded-lg bg-primary-500/10 px-3 py-1.5 text-xs font-medium text-primary-600 transition-colors hover:bg-primary-500/20"
            >
              <Pencil className="h-3.5 w-3.5" />
              Edit
            </button>
          )}
          {onDelete && (
            <button
              onClick={() => onDelete(tool)}
              className="ml-auto flex items-center justify-center rounded-lg p-1.5 text-[var(--foreground-subtle)] transition-colors hover:bg-red-500/10 hover:text-red-500"
              title="Delete tool"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          )}
        </div>
      </div>
    </motion.div>
  );
}
