"use client";

import { use, useState, useMemo } from "react";
import Link from "next/link";
import { ArrowLeft, Search, Wrench, Package, Link2 } from "lucide-react";
import { useAgent } from "@/hooks/use-agents";
import {
  useTools,
  useDeleteTool,
  useToggleToolActive,
} from "@/hooks/use-tools";
import { useUTCPManuals, useUTCPSync, useUTCPDelete } from "@/hooks/use-utcp";
import { CatalogCard } from "@/components/tools/catalog-card";
import { InstalledToolCard } from "@/components/tools/installed-tool-card";
import { TemplateConfigDialog } from "@/components/tools/template-config-dialog";
import { CategoryFilter } from "@/components/tools/category-filter";
import { UTCPConnectDialog } from "@/components/tools/utcp-connect-dialog";
import { UTCPManualCard } from "@/components/tools/utcp-manual-card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  TOOL_TEMPLATES,
  getTemplateById,
  searchTemplates,
  getTemplatesByCategory,
  type ToolTemplate,
  type ToolCategory,
} from "@/lib/tool-templates";
import type { ToolDefinition, UTCPManual } from "@/lib/types";

type ActiveTab = "catalog" | "installed" | "utcp";

export default function AgentToolsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data: agentData, isLoading: agentLoading } = useAgent(id);
  const { data: toolsData, isLoading: toolsLoading } = useTools(id);
  const deleteTool = useDeleteTool(id);
  const toggleActive = useToggleToolActive(id);
  const { data: manualsData } = useUTCPManuals(id);
  const syncManual = useUTCPSync(id);
  const deleteManual = useUTCPDelete(id);

  const agent = agentData?.data;
  const tools = toolsData?.data || [];
  const manuals = manualsData?.data || [];
  const isLoading = agentLoading || toolsLoading;

  const [activeTab, setActiveTab] = useState<ActiveTab>("catalog");
  const [selectedCategory, setSelectedCategory] = useState<ToolCategory | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedTemplate, setSelectedTemplate] = useState<ToolTemplate | null>(null);
  const [editingTool, setEditingTool] = useState<{ tool: ToolDefinition; template: ToolTemplate } | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<ToolDefinition | null>(null);
  const [showUTCPConnect, setShowUTCPConnect] = useState(false);
  const [deleteManualConfirm, setDeleteManualConfirm] = useState<UTCPManual | null>(null);
  const [syncingManualId, setSyncingManualId] = useState<string | null>(null);

  // Set of installed template IDs
  const installedTemplateIds = useMemo(
    () => new Set(tools.map((t) => t.name)),
    [tools]
  );

  // Filtered templates for catalog
  const filteredTemplates = useMemo(() => {
    let results = TOOL_TEMPLATES;

    if (searchQuery.trim()) {
      results = searchTemplates(searchQuery);
    }

    if (selectedCategory) {
      results = results.filter((t) => t.category === selectedCategory);
    }

    return results;
  }, [searchQuery, selectedCategory]);

  // Handle edit: match tool to template, fall back to custom-tool
  const handleEdit = (tool: ToolDefinition) => {
    const template =
      getTemplateById(tool.name) ||
      getTemplateById("custom-tool")!;
    setEditingTool({ tool, template });
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-7 w-48" variant="text" />
          <Skeleton className="h-10 w-64 rounded-lg" />
        </div>
        <Skeleton className="h-10 w-full rounded-lg" />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-52 w-full rounded-xl" />
          ))}
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

  return (
    <div>
      {/* Header */}
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <Link
            href={`/agents/${id}`}
            className="rounded-lg p-1.5 text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)]"
          >
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div>
            <h2 className="text-lg font-semibold text-[var(--foreground)]">
              {agent.profile.name} — Tool Store
            </h2>
            <p className="text-sm text-[var(--foreground-muted)]">
              Browse and configure tools for your agent
            </p>
          </div>
        </div>

        {/* Search */}
        <div className="relative w-full sm:w-64">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--foreground-subtle)]" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search tools..."
            className="w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] py-2 pl-9 pr-4 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
          />
        </div>
      </div>

      {/* Tabs */}
      <div className="mb-5 flex gap-1 rounded-xl bg-[var(--background-tertiary)] p-1">
        <button
          onClick={() => setActiveTab("catalog")}
          className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all ${
            activeTab === "catalog"
              ? "bg-[var(--card)] text-[var(--foreground)] shadow-sm"
              : "text-[var(--foreground-muted)] hover:text-[var(--foreground)]"
          }`}
        >
          <Package className="h-4 w-4" />
          Catalog
        </button>
        <button
          onClick={() => setActiveTab("installed")}
          className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all ${
            activeTab === "installed"
              ? "bg-[var(--card)] text-[var(--foreground)] shadow-sm"
              : "text-[var(--foreground-muted)] hover:text-[var(--foreground)]"
          }`}
        >
          <Wrench className="h-4 w-4" />
          Installed
          {tools.length > 0 && (
            <span className="rounded-full bg-primary-500/10 px-2 py-0.5 text-xs font-semibold text-primary-600">
              {tools.length}
            </span>
          )}
        </button>
        <button
          onClick={() => setActiveTab("utcp")}
          className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all ${
            activeTab === "utcp"
              ? "bg-[var(--card)] text-[var(--foreground)] shadow-sm"
              : "text-[var(--foreground-muted)] hover:text-[var(--foreground)]"
          }`}
        >
          <Link2 className="h-4 w-4" />
          UTCP
          {manuals.length > 0 && (
            <span className="rounded-full bg-indigo-500/10 px-2 py-0.5 text-xs font-semibold text-indigo-600">
              {manuals.length}
            </span>
          )}
        </button>
      </div>

      {/* Catalog Tab */}
      {activeTab === "catalog" && (
        <>
          {/* Category filter */}
          <div className="mb-5">
            <CategoryFilter
              selected={selectedCategory}
              onSelect={setSelectedCategory}
            />
          </div>

          {/* Templates grid */}
          {filteredTemplates.length === 0 ? (
            <div
              className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] py-16 text-center"
              style={{ boxShadow: "var(--shadow-sm)" }}
            >
              <Search className="mx-auto h-12 w-12 text-[var(--foreground-subtle)]" />
              <h3 className="mt-4 text-base font-semibold text-[var(--foreground)]">
                No tools found
              </h3>
              <p className="mt-1 text-sm text-[var(--foreground-muted)]">
                Try a different search term or category.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {filteredTemplates.map((template, i) => (
                <CatalogCard
                  key={template.id}
                  template={template}
                  index={i}
                  onAdd={(t) => setSelectedTemplate(t)}
                  isInstalled={installedTemplateIds.has(template.id)}
                />
              ))}
            </div>
          )}
        </>
      )}

      {/* Installed Tab */}
      {activeTab === "installed" && (
        <>
          {tools.length === 0 ? (
            <div
              className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] py-16 text-center"
              style={{ boxShadow: "var(--shadow-sm)" }}
            >
              <Wrench className="mx-auto h-12 w-12 text-[var(--foreground-subtle)]" />
              <h3 className="mt-4 text-base font-semibold text-[var(--foreground)]">
                No tools installed
              </h3>
              <p className="mt-1 text-sm text-[var(--foreground-muted)]">
                Browse the catalog to add tools to your agent.
              </p>
              <button
                onClick={() => setActiveTab("catalog")}
                className="mt-4 inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-primary-600 to-primary-700 px-4 py-2 text-sm font-medium text-white transition-all hover:from-primary-700 hover:to-primary-800"
              >
                <Package className="h-4 w-4" />
                Browse Catalog
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              {tools.map((tool, i) => (
                <InstalledToolCard
                  key={tool.id}
                  tool={tool}
                  index={i}
                  onEdit={handleEdit}
                  onDelete={(t) => setDeleteConfirm(t)}
                  onToggleActive={(t) =>
                    toggleActive.mutate({
                      toolId: t.id,
                      isActive: !t.is_active,
                    })
                  }
                />
              ))}
            </div>
          )}
        </>
      )}

      {/* UTCP Tab */}
      {activeTab === "utcp" && (
        <>
          <div className="mb-5 flex items-center justify-between">
            <p className="text-sm text-[var(--foreground-muted)]">
              Connect external tool providers via UTCP protocol
            </p>
            <button
              onClick={() => setShowUTCPConnect(true)}
              className="flex items-center gap-2 rounded-lg bg-gradient-to-r from-indigo-600 to-indigo-700 px-4 py-2 text-sm font-medium text-white transition-all hover:from-indigo-700 hover:to-indigo-800"
            >
              <Link2 className="h-4 w-4" />
              Connect UTCP
            </button>
          </div>

          {manuals.length === 0 ? (
            <div
              className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] py-16 text-center"
              style={{ boxShadow: "var(--shadow-sm)" }}
            >
              <Link2 className="mx-auto h-12 w-12 text-[var(--foreground-subtle)]" />
              <h3 className="mt-4 text-base font-semibold text-[var(--foreground)]">
                No UTCP providers connected
              </h3>
              <p className="mt-1 text-sm text-[var(--foreground-muted)]">
                Connect an external UTCP provider to discover and use its tools.
              </p>
              <button
                onClick={() => setShowUTCPConnect(true)}
                className="mt-4 inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-indigo-600 to-indigo-700 px-4 py-2 text-sm font-medium text-white transition-all hover:from-indigo-700 hover:to-indigo-800"
              >
                <Link2 className="h-4 w-4" />
                Connect UTCP Provider
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              {manuals.map((manual, i) => (
                <UTCPManualCard
                  key={manual.id}
                  manual={manual}
                  index={i}
                  isSyncing={syncingManualId === manual.id}
                  onSync={async (m) => {
                    setSyncingManualId(m.id);
                    try {
                      await syncManual.mutateAsync(m.id);
                    } finally {
                      setSyncingManualId(null);
                    }
                  }}
                  onDelete={(m) => setDeleteManualConfirm(m)}
                />
              ))}
            </div>
          )}
        </>
      )}

      {/* UTCP Connect Dialog */}
      {showUTCPConnect && (
        <UTCPConnectDialog
          agentId={id}
          onClose={() => setShowUTCPConnect(false)}
        />
      )}

      {/* UTCP Delete Confirmation */}
      {deleteManualConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            className="absolute inset-0 bg-black/50"
            onClick={() => setDeleteManualConfirm(null)}
          />
          <div className="relative z-10 mx-4">
            <div
              className="w-full max-w-sm rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-6"
              style={{ boxShadow: "var(--shadow-lg)" }}
            >
              <h3 className="text-lg font-semibold text-[var(--foreground)]">
                Disconnect UTCP Provider
              </h3>
              <p className="mt-2 text-sm text-[var(--foreground-muted)]">
                Disconnect <strong>{deleteManualConfirm.name}</strong>? This will
                remove all {deleteManualConfirm.tools_count} associated tools.
              </p>
              <div className="mt-5 flex gap-3">
                <button
                  onClick={async () => {
                    await deleteManual.mutateAsync(deleteManualConfirm.id);
                    setDeleteManualConfirm(null);
                  }}
                  disabled={deleteManual.isPending}
                  className="rounded-xl bg-red-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-red-700 disabled:opacity-50"
                >
                  {deleteManual.isPending ? "Removing..." : "Disconnect"}
                </button>
                <button
                  onClick={() => setDeleteManualConfirm(null)}
                  className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] px-4 py-2 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)]"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Template Config Dialog (create) */}
      {selectedTemplate && (
        <TemplateConfigDialog
          template={selectedTemplate}
          agentId={id}
          onClose={() => setSelectedTemplate(null)}
        />
      )}

      {/* Template Config Dialog (edit) */}
      {editingTool && (
        <TemplateConfigDialog
          template={editingTool.template}
          agentId={id}
          existingTool={editingTool.tool}
          onClose={() => setEditingTool(null)}
        />
      )}

      {/* Delete confirmation dialog */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            className="absolute inset-0 bg-black/50"
            onClick={() => setDeleteConfirm(null)}
          />
          <div className="relative z-10 mx-4">
            <div
              className="w-full max-w-sm rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-6"
              style={{ boxShadow: "var(--shadow-lg)" }}
            >
              <h3 className="text-lg font-semibold text-[var(--foreground)]">
                Delete Tool
              </h3>
              <p className="mt-2 text-sm text-[var(--foreground-muted)]">
                Are you sure you want to delete{" "}
                <strong>{deleteConfirm.name}</strong>? This action cannot be
                undone.
              </p>
              <div className="mt-5 flex gap-3">
                <button
                  onClick={async () => {
                    await deleteTool.mutateAsync(deleteConfirm.id);
                    setDeleteConfirm(null);
                  }}
                  disabled={deleteTool.isPending}
                  className="rounded-xl bg-red-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-red-700 disabled:opacity-50"
                >
                  {deleteTool.isPending ? "Deleting..." : "Delete"}
                </button>
                <button
                  onClick={() => setDeleteConfirm(null)}
                  className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] px-4 py-2 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)]"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
