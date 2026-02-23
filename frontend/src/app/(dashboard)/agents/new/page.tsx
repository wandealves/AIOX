"use client";

import { useCreateAgent } from "@/hooks/use-agents";
import { AgentForm } from "@/components/agents/agent-form";
import type { CreateAgentRequest } from "@/lib/types";

export default function NewAgentPage() {
  const createAgent = useCreateAgent();

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-gray-900">
        Create New Agent
      </h2>
      <AgentForm
        onSubmit={async (data) => {
          await createAgent.mutateAsync(data as CreateAgentRequest);
        }}
        isLoading={createAgent.isPending}
      />
    </div>
  );
}
