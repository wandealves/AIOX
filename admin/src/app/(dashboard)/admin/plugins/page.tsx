"use client";

import { useState } from "react";
import toast from "react-hot-toast";
import { Plus, Trash2, RefreshCw, ExternalLink, X } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import {
  usePlugins,
  useRegisterPlugin,
  useDeletePlugin,
  useSyncPlugin,
} from "@/hooks/use-plugins";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import type { Plugin } from "@/lib/types";

function timeAgo(dateStr?: string): string {
  if (!dateStr) return "Never";
  const diff = Date.now() - new Date(dateStr).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "Just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export default function AdminPluginsPage() {
  const { data, isLoading } = usePlugins();
  const registerPlugin = useRegisterPlugin();
  const deletePlugin = useDeletePlugin();
  const syncPlugin = useSyncPlugin();

  const [registerOpen, setRegisterOpen] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<Plugin | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);

  const plugins = data?.data || [];

  const handleSync = async (plugin: Plugin) => {
    setSyncingId(plugin.id);
    try {
      await syncPlugin.mutateAsync(plugin.id);
      toast.success(`Plugin "${plugin.name}" synced`);
    } catch {
      toast.error("Failed to sync plugin");
    }
    setSyncingId(null);
  };

  const handleConfirmDelete = async () => {
    if (!pendingDelete) return;
    try {
      await deletePlugin.mutateAsync(pendingDelete.id);
      toast.success("Plugin deleted");
    } catch {
      toast.error("Failed to delete plugin");
    }
    setPendingDelete(null);
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-[var(--foreground)]">
          Plugins ({plugins.length})
        </h2>
        <button
          onClick={() => setRegisterOpen(true)}
          className="flex items-center gap-2 rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 px-4 py-2.5 text-sm font-semibold text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98]"
          style={{ boxShadow: "var(--shadow-sm)" }}
        >
          <Plus className="h-4 w-4" />
          Register Plugin
        </button>
      </div>

      {/* Loading skeleton */}
      {isLoading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div
              key={i}
              className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-5"
            >
              <Skeleton className="mb-3 h-5 w-3/4" variant="text" />
              <Skeleton className="mb-2 h-4 w-full" variant="text" />
              <Skeleton className="mb-4 h-4 w-2/3" variant="text" />
              <Skeleton className="h-8 w-1/3" variant="rectangular" />
            </div>
          ))}
        </div>
      )}

      {/* Empty state */}
      {!isLoading && plugins.length === 0 && (
        <div
          className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] px-6 py-12 text-center"
          style={{ boxShadow: "var(--shadow-sm)" }}
        >
          <p className="text-[var(--foreground-muted)]">
            No plugins registered yet. Click &quot;Register Plugin&quot; to add one.
          </p>
        </div>
      )}

      {/* Plugin cards */}
      {!isLoading && plugins.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {plugins.map((plugin, i) => (
            <motion.div
              key={plugin.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.05, duration: 0.3 }}
              className="group relative overflow-hidden rounded-xl border border-[var(--card-border)] bg-[var(--card)] transition-all hover:-translate-y-0.5 hover:shadow-lg"
              style={{ boxShadow: "var(--shadow-sm)" }}
            >
              <div className="h-[3px] bg-primary-600" />
              <div className="p-5">
                <div className="flex items-start justify-between">
                  <h3 className="font-semibold text-[var(--foreground)]">
                    {plugin.name}
                  </h3>
                  <Badge variant={plugin.is_active ? "success" : "default"}>
                    {plugin.is_active ? "Active" : "Inactive"}
                  </Badge>
                </div>

                <p className="mt-2 line-clamp-2 text-sm text-[var(--foreground-muted)]">
                  {plugin.description || "No description"}
                </p>

                <div className="mt-3 flex items-center gap-1.5 text-xs text-[var(--foreground-subtle)]">
                  <ExternalLink className="h-3 w-3" />
                  <span className="truncate">{plugin.manifest_url}</span>
                </div>

                <div className="mt-2 flex items-center gap-3 text-xs text-[var(--foreground-subtle)]">
                  <span>
                    Synced: {timeAgo(plugin.last_synced_at)}
                  </span>
                  {plugin.auto_sync && (
                    <Badge variant="info">Auto-sync</Badge>
                  )}
                </div>

                <div className="mt-4 flex items-center gap-2">
                  <button
                    onClick={() => handleSync(plugin)}
                    disabled={syncingId === plugin.id}
                    className="flex items-center gap-1.5 rounded-lg bg-primary-500/10 px-3 py-1.5 text-xs font-medium text-primary-600 transition-colors hover:bg-primary-500/20 disabled:opacity-50"
                  >
                    <RefreshCw
                      className={`h-3.5 w-3.5 ${syncingId === plugin.id ? "animate-spin" : ""}`}
                    />
                    Sync
                  </button>
                  <button
                    onClick={() => setPendingDelete(plugin)}
                    className="ml-auto flex items-center justify-center rounded-lg p-1.5 text-[var(--foreground-subtle)] transition-colors hover:bg-red-500/10 hover:text-red-500"
                    title="Delete plugin"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      )}

      {/* Register Dialog */}
      <AnimatePresence>
        {registerOpen && (
          <RegisterPluginDialog onClose={() => setRegisterOpen(false)} />
        )}
      </AnimatePresence>

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={!!pendingDelete}
        title="Delete Plugin"
        message={`Are you sure you want to delete "${pendingDelete?.name}"? All associated system tools will also be removed.`}
        variant="danger"
        confirmLabel="Delete"
        onConfirm={handleConfirmDelete}
        onCancel={() => setPendingDelete(null)}
      />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Register Plugin Dialog
// ---------------------------------------------------------------------------

function RegisterPluginDialog({ onClose }: { onClose: () => void }) {
  const registerPlugin = useRegisterPlugin();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [manifestUrl, setManifestUrl] = useState("");
  const [autoSync, setAutoSync] = useState(true);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!name.trim()) {
      setError("Name is required");
      return;
    }
    if (!manifestUrl.trim()) {
      setError("Manifest URL is required");
      return;
    }
    try {
      new URL(manifestUrl);
    } catch {
      setError("Manifest URL must be a valid URL");
      return;
    }

    try {
      await registerPlugin.mutateAsync({
        name: name.trim(),
        description: description.trim(),
        manifest_url: manifestUrl.trim(),
        auto_sync: autoSync,
      });
      toast.success("Plugin registered!");
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "An error occurred";
      setError(msg);
      toast.error(msg);
    }
  };

  const inputClass =
    "w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-2.5 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="absolute inset-0 bg-black/40 backdrop-blur-sm"
        onClick={onClose}
      />
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        transition={{ duration: 0.15 }}
        className="relative z-10 mx-4 w-full max-w-lg rounded-xl border border-[var(--card-border)] bg-[var(--card)]"
        style={{ boxShadow: "var(--shadow-lg)" }}
      >
        <div className="h-[3px] bg-primary-600" />
        <div className="flex items-center justify-between border-b border-[var(--card-border)] p-5">
          <h3 className="text-lg font-semibold text-[var(--foreground)]">
            Register Plugin
          </h3>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)]"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5">
          <div className="space-y-4">
            {error && (
              <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-3 text-sm text-red-600">
                {error}
              </div>
            )}

            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className={inputClass}
                placeholder="e.g. weather-tools"
              />
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Description
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={2}
                className={inputClass}
                placeholder="What does this plugin provide?"
              />
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
                Manifest URL <span className="text-red-500">*</span>
              </label>
              <input
                type="url"
                value={manifestUrl}
                onChange={(e) => setManifestUrl(e.target.value)}
                className={inputClass}
                placeholder="https://example.com/plugin/manifest.json"
              />
              <p className="mt-1 text-xs text-[var(--foreground-subtle)]">
                URL to the plugin manifest JSON that defines available tools
              </p>
            </div>

            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={() => setAutoSync(!autoSync)}
                className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors ${
                  autoSync
                    ? "bg-primary-600"
                    : "bg-[var(--background-tertiary)]"
                }`}
              >
                <span
                  className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-sm transition-transform ${
                    autoSync ? "translate-x-5" : "translate-x-0"
                  }`}
                />
              </button>
              <label className="text-sm font-medium text-[var(--foreground-muted)]">
                Auto-sync tools from manifest
              </label>
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={registerPlugin.isPending}
                className="rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 px-5 py-2.5 text-sm font-semibold text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98] disabled:opacity-50"
                style={{ boxShadow: "var(--shadow-sm)" }}
              >
                {registerPlugin.isPending ? "Registering..." : "Register Plugin"}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] px-5 py-2.5 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)]"
              >
                Cancel
              </button>
            </div>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
