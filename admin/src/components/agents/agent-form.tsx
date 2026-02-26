"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import toast from "react-hot-toast";
import type { Agent, CreateAgentRequest, UpdateAgentRequest } from "@/lib/types";
import { LLM_PROVIDERS, getProviderConfig } from "@/lib/llm-providers";

interface AgentFormProps {
  agent?: Agent;
  onSubmit: (data: CreateAgentRequest | UpdateAgentRequest) => Promise<void>;
  isLoading?: boolean;
}

const CUSTOM_MODEL_VALUE = "__custom__";

export function AgentForm({ agent, onSubmit, isLoading }: AgentFormProps) {
  const router = useRouter();
  const [name, setName] = useState(agent?.profile.name || "");
  const [description, setDescription] = useState(
    agent?.profile.description || ""
  );
  const [systemPrompt, setSystemPrompt] = useState(
    agent?.profile.system_prompt || ""
  );
  const [visibility, setVisibility] = useState(
    agent?.visibility || "private"
  );
  const [llmProvider, setLlmProvider] = useState(
    (agent?.llm_config as Record<string, string>)?.provider || "openai"
  );
  const [llmModel, setLlmModel] = useState(
    (agent?.llm_config as Record<string, string>)?.model || "gpt-4o-mini"
  );
  const [error, setError] = useState("");

  // Determine whether the current model is a custom (not in the provider's list)
  const providerModels = getProviderConfig(llmProvider)?.models ?? [];
  const isCustom = llmModel !== "" && !providerModels.some((m) => m.value === llmModel);
  const [useCustomModel, setUseCustomModel] = useState(isCustom);

  const handleProviderChange = (newProvider: string) => {
    setLlmProvider(newProvider);
    const config = getProviderConfig(newProvider);
    if (config) {
      setLlmModel(config.defaultModel);
      setUseCustomModel(false);
    }
  };

  const handleModelSelectChange = (value: string) => {
    if (value === CUSTOM_MODEL_VALUE) {
      setUseCustomModel(true);
      setLlmModel("");
    } else {
      setUseCustomModel(false);
      setLlmModel(value);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    try {
      const data: CreateAgentRequest = {
        name,
        description,
        system_prompt: systemPrompt,
        visibility,
        llm_config: { provider: llmProvider, model: llmModel },
      };
      await onSubmit(data);
      toast.success(agent ? "Agent updated!" : "Agent created!");
      router.push("/agents");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "An error occurred";
      setError(msg);
      toast.error(msg);
    }
  };

  const inputClass =
    "w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-2.5 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]";

  return (
    <form onSubmit={handleSubmit} className="mx-auto max-w-2xl">
      <div
        className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-6 md:p-8"
        style={{ boxShadow: "var(--shadow-sm)" }}
      >
        {error && (
          <div className="mb-6 rounded-lg bg-red-500/10 border border-red-500/20 p-3 text-sm text-red-600">
            {error}
          </div>
        )}

        {/* General Info */}
        <div className="mb-8">
          <h3 className="text-sm font-semibold uppercase tracking-wide text-[var(--foreground-muted)]">
            General Info
          </h3>
          <div className="mt-1 mb-5 h-px bg-[var(--card-border)]" />

          <div className="space-y-4">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                className={inputClass}
                placeholder="My Agent"
              />
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Description
              </label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className={inputClass}
                placeholder="A helpful assistant that..."
              />
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                System Prompt
              </label>
              <textarea
                value={systemPrompt}
                onChange={(e) => setSystemPrompt(e.target.value)}
                required
                rows={6}
                className={inputClass}
                placeholder="You are a helpful assistant..."
              />
            </div>
          </div>
        </div>

        {/* LLM Config */}
        <div className="mb-8">
          <h3 className="text-sm font-semibold uppercase tracking-wide text-[var(--foreground-muted)]">
            LLM Configuration
          </h3>
          <div className="mt-1 mb-5 h-px bg-[var(--card-border)]" />

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {/* Provider */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Provider
              </label>
              <select
                value={llmProvider}
                onChange={(e) => handleProviderChange(e.target.value)}
                className={inputClass}
              >
                {LLM_PROVIDERS.map((p) => (
                  <option key={p.value} value={p.value}>
                    {p.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Model */}
            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Model
              </label>
              <select
                value={useCustomModel ? CUSTOM_MODEL_VALUE : llmModel}
                onChange={(e) => handleModelSelectChange(e.target.value)}
                className={inputClass}
              >
                {providerModels.map((m) => (
                  <option key={m.value} value={m.value}>
                    {m.label}
                  </option>
                ))}
                <option value={CUSTOM_MODEL_VALUE}>Custom model...</option>
              </select>
            </div>
          </div>

          {/* Custom model text input — shown only when "Custom model..." is selected */}
          {useCustomModel && (
            <div className="mt-3">
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Custom model name
              </label>
              <input
                type="text"
                value={llmModel}
                onChange={(e) => setLlmModel(e.target.value)}
                required
                className={inputClass}
                placeholder={
                  llmProvider === "ollama"
                    ? "e.g. llama3.2:70b"
                    : "e.g. gpt-4-custom"
                }
                autoFocus
              />
            </div>
          )}
        </div>

        {/* Visibility */}
        <div className="mb-8">
          <h3 className="text-sm font-semibold uppercase tracking-wide text-[var(--foreground-muted)]">
            Visibility
          </h3>
          <div className="mt-1 mb-5 h-px bg-[var(--card-border)]" />

          <select
            value={visibility}
            onChange={(e) => setVisibility(e.target.value)}
            className={inputClass}
          >
            <option value="private">Private</option>
            <option value="public">Public</option>
          </select>
        </div>

        {/* Actions */}
        <div className="flex gap-3">
          <button
            type="submit"
            disabled={isLoading}
            className="rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 px-5 py-2.5 text-sm font-semibold text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98] disabled:opacity-50"
            style={{ boxShadow: "var(--shadow-sm)" }}
          >
            {isLoading ? "Saving..." : agent ? "Update Agent" : "Create Agent"}
          </button>
          <button
            type="button"
            onClick={() => router.back()}
            className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] px-5 py-2.5 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)]"
          >
            Cancel
          </button>
        </div>
      </div>
    </form>
  );
}
