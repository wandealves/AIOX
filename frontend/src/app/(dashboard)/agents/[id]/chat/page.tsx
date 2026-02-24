"use client";

import { use } from "react";
import { useAgent } from "@/hooks/use-agents";
import { ChatWindow } from "@/components/chat/chat-window";
import { Skeleton } from "@/components/ui/skeleton";

export default function AgentChatPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data, isLoading } = useAgent(id);

  const agent = data?.data;

  if (isLoading) {
    return (
      <div className="flex h-[calc(100vh-8rem)] flex-col rounded-xl border border-[var(--card-border)] bg-[var(--card)]">
        <div className="flex items-center gap-3 border-b border-[var(--card-border)] px-4 py-3">
          <Skeleton className="h-9 w-9 rounded-full" variant="circular" />
          <div>
            <Skeleton className="h-4 w-32 mb-1" variant="text" />
            <Skeleton className="h-3 w-20" variant="text" />
          </div>
        </div>
        <div className="flex-1 space-y-4 p-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className={`flex gap-3 ${i % 2 === 0 ? "" : "flex-row-reverse"}`}>
              <Skeleton className="h-8 w-8 shrink-0 rounded-full" variant="circular" />
              <Skeleton className={`h-16 rounded-2xl ${i % 2 === 0 ? "w-2/3" : "w-1/2"}`} />
            </div>
          ))}
        </div>
        <div className="border-t border-[var(--card-border)] p-4">
          <Skeleton className="h-11 w-full rounded-xl" />
        </div>
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

  return <ChatWindow agentId={id} agentName={agent.profile.name} />;
}
