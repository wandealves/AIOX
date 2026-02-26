"use client";

import {
  TOOL_CATEGORIES,
  TOOL_ICONS,
  CATEGORY_COLORS,
  type ToolCategory,
} from "@/lib/tool-templates";

interface CategoryFilterProps {
  selected: ToolCategory | null;
  onSelect: (category: ToolCategory | null) => void;
}

export function CategoryFilter({ selected, onSelect }: CategoryFilterProps) {
  return (
    <div className="flex flex-wrap gap-2">
      {/* All pill */}
      <button
        onClick={() => onSelect(null)}
        className={`flex items-center gap-1.5 rounded-full px-3.5 py-1.5 text-xs font-medium transition-all ${
          selected === null
            ? "bg-primary-500/10 text-primary-600"
            : "bg-[var(--background-tertiary)] text-[var(--foreground-muted)] hover:bg-[var(--card-hover)]"
        }`}
      >
        All
      </button>

      {/* Category pills */}
      {(Object.entries(TOOL_CATEGORIES) as [ToolCategory, typeof TOOL_CATEGORIES[ToolCategory]][]).map(
        ([key, cat]) => {
          const Icon = TOOL_ICONS[cat.iconName];
          const colors = CATEGORY_COLORS[cat.color];
          const isActive = selected === key;

          return (
            <button
              key={key}
              onClick={() => onSelect(isActive ? null : key)}
              className={`flex items-center gap-1.5 rounded-full px-3.5 py-1.5 text-xs font-medium transition-all ${
                isActive
                  ? `${colors.bg} ${colors.text}`
                  : "bg-[var(--background-tertiary)] text-[var(--foreground-muted)] hover:bg-[var(--card-hover)]"
              }`}
            >
              {Icon && <Icon className="h-3.5 w-3.5" />}
              {cat.label}
            </button>
          );
        }
      )}
    </div>
  );
}
