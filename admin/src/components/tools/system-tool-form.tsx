"use client";

import { useState } from "react";
import { X } from "lucide-react";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { useCreateSystemTool, useUpdateSystemTool } from "@/hooks/use-system-tools";
import type { SystemTool, CreateSystemToolRequest, UpdateSystemToolRequest } from "@/lib/types";

interface SystemToolFormProps {
  tool?: SystemTool;
  onClose: () => void;
}

const CATEGORIES = [
  { value: "utilities", label: "Utilities" },
  { value: "search", label: "Search" },
  { value: "productivity", label: "Productivity" },
  { value: "communication", label: "Communication" },
  { value: "development", label: "Development" },
  { value: "ai-media", label: "AI & Media" },
  { value: "data", label: "Data" },
  { value: "custom", label: "Custom" },
];

const TOOL_TYPES = [
  { value: "builtin", label: "Built-in" },
  { value: "http", label: "HTTP" },
];

const HTTP_METHODS = ["GET", "POST", "PUT", "DELETE", "PATCH"];

const AUTH_TYPES = [
  { value: "none", label: "None" },
  { value: "bearer", label: "Bearer Token" },
  { value: "api_key", label: "API Key" },
];

export function SystemToolForm({ tool, onClose }: SystemToolFormProps) {
  const isEdit = !!tool;
  const createTool = useCreateSystemTool();
  const updateTool = useUpdateSystemTool(tool?.id || "");

  const [name, setName] = useState(tool?.name || "");
  const [description, setDescription] = useState(tool?.description || "");
  const [category, setCategory] = useState(tool?.category || "utilities");
  const [toolType, setToolType] = useState(tool?.tool_type || "http");
  const [endpointUrl, setEndpointUrl] = useState(tool?.endpoint_url || "");
  const [httpMethod, setHttpMethod] = useState(tool?.http_method || "POST");
  const [parameters, setParameters] = useState(
    tool?.parameters ? JSON.stringify(tool.parameters, null, 2) : ""
  );
  const [outputSchema, setOutputSchema] = useState(
    tool?.output_schema ? JSON.stringify(tool.output_schema, null, 2) : ""
  );
  const [authType, setAuthType] = useState(tool?.auth_type || "none");
  const [authConfig, setAuthConfig] = useState(
    tool?.auth_config ? JSON.stringify(tool.auth_config, null, 2) : ""
  );
  const [timeoutSec, setTimeoutSec] = useState(tool?.timeout_sec ?? 30);
  const [version, setVersion] = useState(tool?.version || "1.0.0");
  const [error, setError] = useState("");

  const isPending = createTool.isPending || updateTool.isPending;

  const parseJSON = (str: string, fieldName: string): Record<string, unknown> | undefined => {
    if (!str.trim()) return undefined;
    try {
      return JSON.parse(str);
    } catch {
      throw new Error(`${fieldName} must be valid JSON`);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name.trim()) {
      setError("Name is required");
      return;
    }
    if (!description.trim()) {
      setError("Description is required");
      return;
    }
    if (toolType === "http" && !endpointUrl.trim()) {
      setError("Endpoint URL is required for HTTP tools");
      return;
    }

    try {
      const parsedParams = parseJSON(parameters, "Parameters");
      const parsedOutputSchema = parseJSON(outputSchema, "Output Schema");
      const parsedAuthConfig = parseJSON(authConfig, "Auth Config");

      if (isEdit) {
        const data: UpdateSystemToolRequest = {
          name: name.trim(),
          description: description.trim(),
          category,
          tool_type: toolType,
          endpoint_url: toolType === "http" ? endpointUrl.trim() : undefined,
          http_method: toolType === "http" ? httpMethod : undefined,
          parameters: parsedParams,
          output_schema: parsedOutputSchema,
          auth_type: authType,
          auth_config: parsedAuthConfig,
          timeout_sec: timeoutSec,
          version: version.trim(),
        };
        await updateTool.mutateAsync(data);
        toast.success("System tool updated!");
      } else {
        const data: CreateSystemToolRequest = {
          name: name.trim(),
          description: description.trim(),
          category,
          tool_type: toolType,
          endpoint_url: toolType === "http" ? endpointUrl.trim() : undefined,
          http_method: toolType === "http" ? httpMethod : undefined,
          parameters: parsedParams,
          output_schema: parsedOutputSchema,
          auth_type: authType,
          auth_config: parsedAuthConfig,
          timeout_sec: timeoutSec,
          version: version.trim(),
        };
        await createTool.mutateAsync(data);
        toast.success("System tool created!");
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
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="absolute inset-0 bg-black/40 backdrop-blur-sm"
        onClick={onClose}
      />
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.15 }}
        className="relative z-10 mx-4 max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl border border-[var(--card-border)] bg-[var(--card)]"
        style={{ boxShadow: "var(--shadow-lg)" }}
      >
        {/* Header */}
        <div className="h-[3px] bg-primary-600" />
        <div className="flex items-center justify-between border-b border-[var(--card-border)] p-5">
          <h3 className="text-lg font-semibold text-[var(--foreground)]">
            {isEdit ? "Edit System Tool" : "Create System Tool"}
          </h3>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)]"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-5">
          <div className="space-y-4">
            {error && (
              <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-3 text-sm text-red-600">
                {error}
              </div>
            )}

            {/* Name */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className={inputClass}
                placeholder="e.g. web_search"
              />
            </div>

            {/* Description */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Description <span className="text-red-500">*</span>
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={2}
                className={inputClass}
                placeholder="What does this tool do?"
              />
            </div>

            {/* Category + Tool Type row */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                  Category
                </label>
                <select
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                  className={inputClass}
                >
                  {CATEGORIES.map((c) => (
                    <option key={c.value} value={c.value}>
                      {c.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                  Type
                </label>
                <select
                  value={toolType}
                  onChange={(e) => setToolType(e.target.value)}
                  className={inputClass}
                >
                  {TOOL_TYPES.map((t) => (
                    <option key={t.value} value={t.value}>
                      {t.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* HTTP fields (conditional) */}
            {toolType === "http" && (
              <>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                    Endpoint URL <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="url"
                    value={endpointUrl}
                    onChange={(e) => setEndpointUrl(e.target.value)}
                    className={inputClass}
                    placeholder="https://api.example.com/v1/action"
                  />
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                    HTTP Method
                  </label>
                  <select
                    value={httpMethod}
                    onChange={(e) => setHttpMethod(e.target.value)}
                    className={inputClass}
                  >
                    {HTTP_METHODS.map((m) => (
                      <option key={m} value={m}>
                        {m}
                      </option>
                    ))}
                  </select>
                </div>
              </>
            )}

            {/* Parameters JSON */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Parameters (JSON Schema)
              </label>
              <textarea
                value={parameters}
                onChange={(e) => setParameters(e.target.value)}
                rows={3}
                className={`${inputClass} font-mono text-xs`}
                placeholder='{"type": "object", "properties": {...}}'
              />
            </div>

            {/* Output Schema JSON */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Output Schema (JSON)
              </label>
              <textarea
                value={outputSchema}
                onChange={(e) => setOutputSchema(e.target.value)}
                rows={3}
                className={`${inputClass} font-mono text-xs`}
                placeholder='{"type": "object", "properties": {...}}'
              />
            </div>

            {/* Auth Type */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Auth Type
              </label>
              <select
                value={authType}
                onChange={(e) => setAuthType(e.target.value)}
                className={inputClass}
              >
                {AUTH_TYPES.map((a) => (
                  <option key={a.value} value={a.value}>
                    {a.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Auth Config (conditional) */}
            {authType !== "none" && (
              <div>
                <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                  Auth Config (JSON)
                </label>
                <textarea
                  value={authConfig}
                  onChange={(e) => setAuthConfig(e.target.value)}
                  rows={3}
                  className={`${inputClass} font-mono text-xs`}
                  placeholder='{"token": "..."}'
                />
              </div>
            )}

            {/* Timeout + Version row */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                  Timeout (sec)
                </label>
                <input
                  type="number"
                  value={timeoutSec}
                  onChange={(e) => setTimeoutSec(Number(e.target.value))}
                  min={1}
                  max={300}
                  className={inputClass}
                />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                  Version
                </label>
                <input
                  type="text"
                  value={version}
                  onChange={(e) => setVersion(e.target.value)}
                  className={inputClass}
                  placeholder="1.0.0"
                />
              </div>
            </div>

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
                    : "Create Tool"}
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
      </motion.div>
    </div>
  );
}
