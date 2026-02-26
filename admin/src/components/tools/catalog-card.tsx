"use client";

import { Plus, Check } from "lucide-react";
import { motion } from "framer-motion";
import {
  TOOL_ICONS,
  TOOL_CATEGORIES,
  CATEGORY_COLORS,
  type ToolTemplate,
} from "@/lib/tool-templates";

interface CatalogCardProps {
  template: ToolTemplate;
  index?: number;
  onAdd: (template: ToolTemplate) => void;
  isInstalled?: boolean;
}

export function CatalogCard({
  template,
  index = 0,
  onAdd,
  isInstalled = false,
}: CatalogCardProps) {
  const Icon = TOOL_ICONS[template.iconName];
  const categoryInfo = TOOL_CATEGORIES[template.category];
  const colors = CATEGORY_COLORS[categoryInfo.color];

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.3 }}
      className="group relative overflow-hidden rounded-xl border border-[var(--card-border)] bg-[var(--card)] transition-all hover:-translate-y-0.5 hover:shadow-lg"
      style={{ boxShadow: "var(--shadow-sm)" }}
    >
      {/* Top accent bar with category color */}
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
                {template.name}
              </h3>
              <span
                className={`inline-block rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${colors.bg} ${colors.text}`}
              >
                {categoryInfo.label}
              </span>
            </div>
          </div>
        </div>

        <p className="mt-3 line-clamp-2 text-sm text-[var(--foreground-muted)]">
          {template.description}
        </p>

        {/* Auth indicator */}
        {template.authType !== "none" && (
          <div className="mt-2">
            <span className="rounded-full bg-[var(--background-tertiary)] px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-[var(--foreground-muted)]">
              {template.authType === "bearer" ? "OAuth / Bearer" : "API Key"}
            </span>
          </div>
        )}

        <div className="mt-4">
          {isInstalled ? (
            <span className="inline-flex items-center gap-1.5 rounded-lg bg-green-500/10 px-3 py-1.5 text-xs font-medium text-green-600">
              <Check className="h-3.5 w-3.5" />
              Installed
            </span>
          ) : (
            <button
              onClick={() => onAdd(template)}
              className="flex items-center gap-1.5 rounded-lg bg-gradient-to-r from-primary-600 to-primary-700 px-3 py-1.5 text-xs font-medium text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98]"
            >
              <Plus className="h-3.5 w-3.5" />
              Add
            </button>
          )}
        </div>
      </div>
    </motion.div>
  );
}
