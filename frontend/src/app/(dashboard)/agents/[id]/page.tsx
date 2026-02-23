"use client";

import { use } from "react";
import Link from "next/link";
import { MessageSquare } from "lucide-react";
import { useAgent, useUpdateAgent } from "@/hooks/use-agents";
import { AgentForm } from "@/components/agents/agent-form";
import type { UpdateAgentRequest } from "@/lib/types";

export default function AgentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data, isLoading } = useAgent(id);
  const updateAgent = useUpdateAgent(id);

  const agent = data?.data;

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  if (!agent) {
    return <div className="text-center text-gray-500">Agent not found</div>;
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900">
          {agent.profile.name}
        </h2>
        <Link
          href={`/agents/${id}/chat`}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
        >
          <MessageSquare className="h-4 w-4" />
          Chat
        </Link>
      </div>
      <AgentForm
        agent={agent}
        onSubmit={async (data) => {
          await updateAgent.mutateAsync(data as UpdateAgentRequest);
        }}
        isLoading={updateAgent.isPending}
      />
    </div>
  );
}
