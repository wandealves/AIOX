"use client";

import { use } from "react";
import { useAgent } from "@/hooks/use-agents";
import { ChatWindow } from "@/components/chat/chat-window";

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
      <div className="flex justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  if (!agent) {
    return <div className="text-center text-gray-500">Agent not found</div>;
  }

  return <ChatWindow agentId={id} agentName={agent.profile.name} />;
}
