"use client";

import { use } from "react";
import Link from "next/link";
import { MessageSquare, Wrench } from "lucide-react";
import { useAgent, useUpdateAgent } from "@/hooks/use-agents";
import { AgentForm } from "@/components/agents/agent-form";
import { Skeleton } from "@/components/ui/skeleton";
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
      <div className="mx-auto max-w-2xl space-y-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-7 w-48" variant="text" />
          <Skeleton className="h-10 w-24 rounded-lg" />
        </div>
        <Skeleton className="h-[500px] w-full rounded-xl" />
      </div>
    );
  }

  if (!agent) {
    return (
      <div className="py-12 text-center text-[var(--foreground-muted)]">
        Agent not found
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-[var(--foreground)]">
          {agent.profile.name}
        </h2>
        <div className="flex items-center gap-2">
          <Link
            href={`/agents/${id}/tools`}
            className="flex items-center gap-2 rounded-lg border border-[var(--card-border)] bg-[var(--card)] px-4 py-2 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)]"
          >
            <Wrench className="h-4 w-4" />
            Tools
          </Link>
          <Link
            href={`/agents/${id}/chat`}
            className="flex items-center gap-2 rounded-lg bg-gradient-to-r from-primary-600 to-primary-700 px-4 py-2 text-sm font-medium text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98]"
          >
            <MessageSquare className="h-4 w-4" />
            Chat
          </Link>
        </div>
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
