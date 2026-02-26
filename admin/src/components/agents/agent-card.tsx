"use client";

import Link from "next/link";
import { Bot, MessageSquare, Settings, Trash2, Wrench } from "lucide-react";
import { motion } from "framer-motion";
import type { Agent } from "@/lib/types";
import { getProviderLabel, providerBadgeStyle } from "@/lib/llm-providers";

interface AgentCardProps {
  agent: Agent;
  index?: number;
  onDelete?: (id: string) => void;
}

export function AgentCard({ agent, index = 0, onDelete }: AgentCardProps) {
  const provider =
    (agent.llm_config as Record<string, string>)?.provider || "";

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.3 }}
      className="group relative overflow-hidden rounded-xl border border-[var(--card-border)] bg-[var(--card)] transition-all hover:-translate-y-0.5 hover:shadow-lg"
      style={{ boxShadow: "var(--shadow-sm)" }}
    >
      {/* Top accent bar */}
      <div className="h-[3px] bg-gradient-to-r from-primary-400 to-primary-600" />

      <div className="p-5">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-500/10">
              <Bot className="h-5 w-5 text-primary-500" />
            </div>
            <div>
              <h3 className="font-semibold text-[var(--foreground)]">
                {agent.profile.name}
              </h3>
              <p className="text-xs text-[var(--foreground-subtle)]">
                {agent.visibility === "public" ? "Public" : "Private"}
              </p>
            </div>
          </div>

          {/* Provider badge */}
          {provider && (
            <span
              className={`rounded-full px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${providerBadgeStyle(provider)}`}
            >
              {getProviderLabel(provider)}
            </span>
          )}
        </div>

        <p className="mt-3 line-clamp-2 text-sm text-[var(--foreground-muted)]">
          {agent.profile.description || "No description"}
        </p>

        {/* Status dot */}
        <div className="mt-3 flex items-center gap-1.5">
          <span className="relative flex h-2 w-2">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
            <span className="relative inline-flex h-2 w-2 rounded-full bg-green-500" />
          </span>
          <span className="text-xs text-[var(--foreground-subtle)]">Active</span>
        </div>

        <div className="mt-4 flex items-center gap-2">
          <Link
            href={`/agents/${agent.id}/chat`}
            className="flex items-center gap-1.5 rounded-lg bg-primary-500/10 px-3 py-1.5 text-xs font-medium text-primary-600 transition-colors hover:bg-primary-500/20"
          >
            <MessageSquare className="h-3.5 w-3.5" />
            Chat
          </Link>
          <Link
            href={`/agents/${agent.id}/tools`}
            className="flex items-center gap-1.5 rounded-lg bg-[var(--background-tertiary)] px-3 py-1.5 text-xs font-medium text-[var(--foreground-muted)] transition-colors hover:bg-[var(--card-hover)]"
          >
            <Wrench className="h-3.5 w-3.5" />
            Tools
          </Link>
          <Link
            href={`/agents/${agent.id}`}
            className="flex items-center gap-1.5 rounded-lg bg-[var(--background-tertiary)] px-3 py-1.5 text-xs font-medium text-[var(--foreground-muted)] transition-colors hover:bg-[var(--card-hover)]"
          >
            <Settings className="h-3.5 w-3.5" />
            Configure
          </Link>
          {onDelete && (
            <button
              onClick={() => onDelete(agent.id)}
              title="Delete agent"
              className="ml-auto flex items-center justify-center rounded-lg p-1.5 text-[var(--foreground-subtle)] transition-colors hover:bg-red-500/10 hover:text-red-500"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          )}
        </div>
      </div>
    </motion.div>
  );
}
