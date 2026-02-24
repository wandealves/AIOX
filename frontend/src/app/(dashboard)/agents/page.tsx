"use client";

import { useState } from "react";
import Link from "next/link";
import { Plus, Bot } from "lucide-react";
import toast from "react-hot-toast";
import { useAgents, useDeleteAgent } from "@/hooks/use-agents";
import { AgentCard } from "@/components/agents/agent-card";
import { Skeleton } from "@/components/ui/skeleton";
import { EmptyState } from "@/components/ui/empty-state";
import { Pagination } from "@/components/ui/pagination";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

export default function AgentsPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useAgents(page, 12);
  const deleteAgent = useDeleteAgent();
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const agents = data?.data || [];
  const totalCount = data?.total_count || 0;
  const totalPages = Math.ceil(totalCount / 12);

  const handleDelete = async () => {
    if (!deleteId) return;
    try {
      await deleteAgent.mutateAsync(deleteId);
      toast.success("Agent deleted successfully");
    } catch {
      toast.error("Failed to delete agent");
    }
    setDeleteId(null);
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-[var(--foreground)]">
          My Agents ({totalCount})
        </h2>
        <Link
          href="/agents/new"
          className="flex items-center gap-2 rounded-lg bg-gradient-to-r from-primary-600 to-primary-700 px-4 py-2 text-sm font-medium text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98]"
          style={{ boxShadow: "var(--shadow-sm)" }}
        >
          <Plus className="h-4 w-4" />
          New Agent
        </Link>
      </div>

      {isLoading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-[200px] w-full rounded-xl" />
          ))}
        </div>
      )}

      {!isLoading && agents.length === 0 && (
        <EmptyState
          icon={Bot}
          title="No agents yet"
          description="Create your first agent to get started with AI automation."
          action={{ label: "Create your first agent", href: "/agents/new" }}
        />
      )}

      {!isLoading && agents.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {agents.map((agent, i) => (
            <AgentCard
              key={agent.id}
              agent={agent}
              index={i}
              onDelete={(id) => setDeleteId(id)}
            />
          ))}
        </div>
      )}

      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />

      <ConfirmDialog
        open={!!deleteId}
        title="Delete Agent"
        message="Are you sure you want to delete this agent? This action cannot be undone."
        variant="danger"
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setDeleteId(null)}
      />
    </div>
  );
}
