"use client";

import { useState } from "react";
import { X, Eye, EyeOff } from "lucide-react";
import toast from "react-hot-toast";
import {
  TOOL_ICONS,
  TOOL_CATEGORIES,
  CATEGORY_COLORS,
  type ToolTemplate,
  type ConfigField,
} from "@/lib/tool-templates";
import { useCreateTool, useUpdateTool } from "@/hooks/use-tools";
import type { ToolDefinition } from "@/lib/types";

interface TemplateConfigDialogProps {
  template: ToolTemplate;
  agentId: string;
  existingTool?: ToolDefinition;
  onClose: () => void;
}

export function TemplateConfigDialog({
  template,
  agentId,
  existingTool,
  onClose,
}: TemplateConfigDialogProps) {
  const isEdit = !!existingTool;
  const createTool = useCreateTool(agentId);
  const updateTool = useUpdateTool(agentId, existingTool?.id || "");

  // Initialize form values from defaults or existing tool
  const [values, setValues] = useState<Record<string, string>>(() => {
    const initial: Record<string, string> = {};
    for (const field of template.configFields) {
      initial[field.key] = field.defaultValue || "";
    }
    return initial;
  });

  const [showPasswords, setShowPasswords] = useState<Record<string, boolean>>({});
  const [error, setError] = useState("");
  const isPending = createTool.isPending || updateTool.isPending;

  const Icon = TOOL_ICONS[template.iconName];
  const categoryInfo = TOOL_CATEGORIES[template.category];
  const colors = CATEGORY_COLORS[categoryInfo.color];

  const setValue = (key: string, value: string) => {
    setValues((prev) => ({ ...prev, [key]: value }));
  };

  const togglePassword = (key: string) => {
    setShowPasswords((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const validateForm = (): boolean => {
    for (const field of template.configFields) {
      if (field.required && !values[field.key]?.trim()) {
        setError(`${field.label} is required`);
        return false;
      }
      if (field.type === "json" && values[field.key]?.trim()) {
        try {
          JSON.parse(values[field.key]);
        } catch {
          setError(`${field.label} must be valid JSON`);
          return false;
        }
      }
      if (field.type === "url" && values[field.key]?.trim()) {
        try {
          new URL(values[field.key]);
        } catch {
          setError(`${field.label} must be a valid URL`);
          return false;
        }
      }
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!validateForm()) return;

    try {
      const request = template.buildRequest(values);

      if (isEdit) {
        await updateTool.mutateAsync(request);
        toast.success("Tool updated!");
      } else {
        await createTool.mutateAsync(request);
        toast.success("Tool added!");
      }
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "An error occurred";
      setError(msg);
      toast.error(msg);
    }
  };

  const inputClass =
    "w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-2.5 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative z-10 mx-4 max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl border border-[var(--card-border)] bg-[var(--card)]" style={{ boxShadow: "var(--shadow-lg)" }}>
        {/* Header */}
        <div className={`h-[3px] ${colors.accent}`} />
        <div className="flex items-center justify-between border-b border-[var(--card-border)] p-5">
          <div className="flex items-center gap-3">
            <div className={`flex h-10 w-10 items-center justify-center rounded-xl ${colors.bg}`}>
              {Icon && <Icon className={`h-5 w-5 ${colors.text}`} />}
            </div>
            <div>
              <h3 className="text-lg font-semibold text-[var(--foreground)]">
                {isEdit ? "Edit" : "Configure"} {template.name}
              </h3>
              <p className="text-xs text-[var(--foreground-muted)]">
                {template.description}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)]"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-5">
          <div className="space-y-5">
            {error && (
              <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-3 text-sm text-red-600">
                {error}
              </div>
            )}

            {template.configFields.length === 0 ? (
              <div className="rounded-lg bg-[var(--background-tertiary)] p-4 text-center text-sm text-[var(--foreground-muted)]">
                No configuration needed. This tool works out of the box!
              </div>
            ) : (
              <>
                <h4 className="text-sm font-semibold uppercase tracking-wide text-[var(--foreground-muted)]">
                  Configuration
                </h4>
                <div className="mt-1 mb-2 h-px bg-[var(--card-border)]" />

                {template.configFields.map((field) => (
                  <FieldRenderer
                    key={field.key}
                    field={field}
                    value={values[field.key] || ""}
                    onChange={(v) => setValue(field.key, v)}
                    inputClass={inputClass}
                    showPassword={showPasswords[field.key] || false}
                    onTogglePassword={() => togglePassword(field.key)}
                  />
                ))}
              </>
            )}

            {/* Actions */}
            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={isPending}
                className="rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 px-5 py-2.5 text-sm font-semibold text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98] disabled:opacity-50"
                style={{ boxShadow: "var(--shadow-sm)" }}
              >
                {isPending
                  ? "Saving..."
                  : isEdit
                    ? "Update Tool"
                    : "Add Tool"}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] px-5 py-2.5 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)]"
              >
                Cancel
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Field Renderer
// ---------------------------------------------------------------------------

function FieldRenderer({
  field,
  value,
  onChange,
  inputClass,
  showPassword,
  onTogglePassword,
}: {
  field: ConfigField;
  value: string;
  onChange: (value: string) => void;
  inputClass: string;
  showPassword: boolean;
  onTogglePassword: () => void;
}) {
  return (
    <div>
      <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
        {field.label}
        {field.required && <span className="ml-1 text-red-500">*</span>}
      </label>

      {field.type === "select" ? (
        <select
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className={inputClass}
        >
          {!field.required && <option value="">None</option>}
          {field.options?.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      ) : field.type === "json" ? (
        <textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={4}
          className={`${inputClass} font-mono text-xs`}
          placeholder={field.placeholder}
        />
      ) : field.type === "password" ? (
        <div className="relative">
          <input
            type={showPassword ? "text" : "password"}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            required={field.required}
            className={`${inputClass} pr-10`}
            placeholder={field.placeholder}
          />
          <button
            type="button"
            onClick={onTogglePassword}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--foreground-subtle)] hover:text-[var(--foreground-muted)]"
          >
            {showPassword ? (
              <EyeOff className="h-4 w-4" />
            ) : (
              <Eye className="h-4 w-4" />
            )}
          </button>
        </div>
      ) : field.type === "number" ? (
        <input
          type="number"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          required={field.required}
          className={inputClass}
          placeholder={field.placeholder}
        />
      ) : (
        <input
          type={field.type === "url" ? "url" : "text"}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          required={field.required}
          className={inputClass}
          placeholder={field.placeholder}
        />
      )}

      {field.helpText && (
        <p className="mt-1 text-xs text-[var(--foreground-subtle)]">
          {field.helpText}
        </p>
      )}
    </div>
  );
}
