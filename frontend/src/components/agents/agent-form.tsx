"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { Agent, CreateAgentRequest, UpdateAgentRequest } from "@/lib/types";

interface AgentFormProps {
  agent?: Agent;
  onSubmit: (data: CreateAgentRequest | UpdateAgentRequest) => Promise<void>;
  isLoading?: boolean;
}

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
    (agent?.llm_config as Record<string, string>)?.model || "gpt-4"
  );
  const [error, setError] = useState("");

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
      router.push("/agents");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "An error occurred");
    }
  };

  return (
    <form onSubmit={handleSubmit} className="mx-auto max-w-2xl space-y-6">
      {error && (
        <div className="rounded-lg bg-red-50 p-3 text-sm text-red-700">
          {error}
        </div>
      )}

      <div>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">
          Name
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-300 focus:ring-1 focus:ring-blue-300"
          placeholder="My Agent"
        />
      </div>

      <div>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">
          Description
        </label>
        <input
          type="text"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-300 focus:ring-1 focus:ring-blue-300"
          placeholder="A helpful assistant that..."
        />
      </div>

      <div>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">
          System Prompt
        </label>
        <textarea
          value={systemPrompt}
          onChange={(e) => setSystemPrompt(e.target.value)}
          required
          rows={6}
          className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-300 focus:ring-1 focus:ring-blue-300"
          placeholder="You are a helpful assistant..."
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="mb-1.5 block text-sm font-medium text-gray-700">
            LLM Provider
          </label>
          <select
            value={llmProvider}
            onChange={(e) => setLlmProvider(e.target.value)}
            className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-300"
          >
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
          </select>
        </div>

        <div>
          <label className="mb-1.5 block text-sm font-medium text-gray-700">
            Model
          </label>
          <input
            type="text"
            value={llmModel}
            onChange={(e) => setLlmModel(e.target.value)}
            className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-300 focus:ring-1 focus:ring-blue-300"
            placeholder="gpt-4"
          />
        </div>
      </div>

      <div>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">
          Visibility
        </label>
        <select
          value={visibility}
          onChange={(e) => setVisibility(e.target.value)}
          className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-300"
        >
          <option value="private">Private</option>
          <option value="public">Public</option>
        </select>
      </div>

      <div className="flex gap-3">
        <button
          type="submit"
          disabled={isLoading}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
        >
          {isLoading ? "Saving..." : agent ? "Update Agent" : "Create Agent"}
        </button>
        <button
          type="button"
          onClick={() => router.back()}
          className="rounded-lg border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
