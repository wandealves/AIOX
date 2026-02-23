"use client";

import Link from "next/link";
import { Bot, MessageSquare, Settings, Trash2 } from "lucide-react";
import type { Agent } from "@/lib/types";

interface AgentCardProps {
  agent: Agent;
  onDelete?: (id: string) => void;
}

export function AgentCard({ agent, onDelete }: AgentCardProps) {
  return (
    <div className="rounded-xl border bg-white p-5 transition-shadow hover:shadow-md">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
            <Bot className="h-5 w-5 text-blue-600" />
          </div>
          <div>
            <h3 className="font-medium text-gray-900">
              {agent.profile.name}
            </h3>
            <p className="text-xs text-gray-500">
              {agent.visibility === "public" ? "Public" : "Private"}
            </p>
          </div>
        </div>
      </div>

      <p className="mt-3 line-clamp-2 text-sm text-gray-600">
        {agent.profile.description || "No description"}
      </p>

      <div className="mt-4 flex items-center gap-2">
        <Link
          href={`/agents/${agent.id}/chat`}
          className="flex items-center gap-1.5 rounded-lg bg-blue-50 px-3 py-1.5 text-xs font-medium text-blue-700 transition-colors hover:bg-blue-100"
        >
          <MessageSquare className="h-3.5 w-3.5" />
          Chat
        </Link>
        <Link
          href={`/agents/${agent.id}`}
          className="flex items-center gap-1.5 rounded-lg bg-gray-50 px-3 py-1.5 text-xs font-medium text-gray-700 transition-colors hover:bg-gray-100"
        >
          <Settings className="h-3.5 w-3.5" />
          Configure
        </Link>
        {onDelete && (
          <button
            onClick={() => onDelete(agent.id)}
            className="ml-auto flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium text-red-600 transition-colors hover:bg-red-50"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        )}
      </div>
    </div>
  );
}
