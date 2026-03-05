"use client";

import { useState } from "react";
import { X, Search, Link2, Loader2, Check, Key } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { useUTCPDiscover, useUTCPConnect } from "@/hooks/use-utcp";
import type { UTCPDiscoverResult } from "@/lib/types";

interface UTCPConnectDialogProps {
  agentId: string;
  onClose: () => void;
}

type Step = "url" | "preview" | "auth" | "success";

export function UTCPConnectDialog({ agentId, onClose }: UTCPConnectDialogProps) {
  const [step, setStep] = useState<Step>("url");
  const [manualUrl, setManualUrl] = useState("");
  const [manualName, setManualName] = useState("");
  const [preview, setPreview] = useState<UTCPDiscoverResult | null>(null);
  const [authValues, setAuthValues] = useState<Record<string, string>>({});
  const [error, setError] = useState("");

  const discover = useUTCPDiscover(agentId);
  const connect = useUTCPConnect(agentId);

  const handleDiscover = async () => {
    setError("");
    try {
      const result = await discover.mutateAsync(manualUrl);
      const data = result.data || (result as unknown as UTCPDiscoverResult);
      setPreview(data);
      setManualName(data.name);

      if (data.auth_vars && data.auth_vars.length > 0) {
        setStep("auth");
      } else {
        setStep("preview");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to discover tools");
    }
  };

  const handleConnect = async () => {
    setError("");
    try {
      await connect.mutateAsync({
        name: manualName,
        manual_url: manualUrl,
        auth_values: Object.keys(authValues).length > 0 ? authValues : undefined,
      });
      setStep("success");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to connect");
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        className="relative z-10 mx-4 w-full max-w-lg"
      >
        <div
          className="overflow-hidden rounded-2xl border border-[var(--card-border)] bg-[var(--card)]"
          style={{ boxShadow: "var(--shadow-lg)" }}
        >
          {/* Header */}
          <div className="flex items-center justify-between border-b border-[var(--card-border)] px-6 py-4">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-indigo-500/10">
                <Link2 className="h-4.5 w-4.5 text-indigo-600" />
              </div>
              <div>
                <h3 className="font-semibold text-[var(--foreground)]">
                  Connect UTCP Provider
                </h3>
                <p className="text-xs text-[var(--foreground-muted)]">
                  {step === "url" && "Enter the UTCP manual URL"}
                  {step === "auth" && "Provide authentication credentials"}
                  {step === "preview" && "Review discovered tools"}
                  {step === "success" && "Connected successfully"}
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="rounded-lg p-1.5 text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)]"
            >
              <X className="h-4 w-4" />
            </button>
          </div>

          {/* Body */}
          <div className="px-6 py-5">
            <AnimatePresence mode="wait">
              {/* Step 1: URL Input */}
              {step === "url" && (
                <motion.div
                  key="url"
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  className="space-y-4"
                >
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-[var(--foreground)]">
                      Manual URL
                    </label>
                    <input
                      type="url"
                      value={manualUrl}
                      onChange={(e) => setManualUrl(e.target.value)}
                      placeholder="https://api.example.com/utcp"
                      className="w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-2.5 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
                    />
                    <p className="mt-1.5 text-xs text-[var(--foreground-subtle)]">
                      The URL to the UTCP manual JSON endpoint
                    </p>
                  </div>

                  {error && (
                    <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-600">
                      {error}
                    </p>
                  )}
                </motion.div>
              )}

              {/* Step 2: Auth Values */}
              {step === "auth" && preview && (
                <motion.div
                  key="auth"
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  className="space-y-4"
                >
                  <div className="rounded-xl bg-[var(--background-tertiary)] px-4 py-3">
                    <p className="text-sm font-medium text-[var(--foreground)]">
                      {preview.name}
                    </p>
                    <p className="text-xs text-[var(--foreground-muted)]">
                      {preview.description} — {preview.tools.length} tool{preview.tools.length !== 1 ? "s" : ""}
                    </p>
                  </div>

                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-[var(--foreground)]">
                      Name
                    </label>
                    <input
                      type="text"
                      value={manualName}
                      onChange={(e) => setManualName(e.target.value)}
                      className="w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-2.5 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
                    />
                  </div>

                  {preview.auth_vars?.map((varName) => (
                    <div key={varName}>
                      <label className="mb-1.5 flex items-center gap-1.5 text-sm font-medium text-[var(--foreground)]">
                        <Key className="h-3.5 w-3.5 text-[var(--foreground-muted)]" />
                        {varName}
                      </label>
                      <input
                        type="password"
                        value={authValues[varName] || ""}
                        onChange={(e) =>
                          setAuthValues((prev) => ({
                            ...prev,
                            [varName]: e.target.value,
                          }))
                        }
                        placeholder={`Enter ${varName}`}
                        className="w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-2.5 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
                      />
                    </div>
                  ))}

                  {error && (
                    <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-600">
                      {error}
                    </p>
                  )}
                </motion.div>
              )}

              {/* Step 2b: Preview (no auth needed) */}
              {step === "preview" && preview && (
                <motion.div
                  key="preview"
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  className="space-y-4"
                >
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-[var(--foreground)]">
                      Name
                    </label>
                    <input
                      type="text"
                      value={manualName}
                      onChange={(e) => setManualName(e.target.value)}
                      className="w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-2.5 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
                    />
                  </div>

                  <div>
                    <p className="mb-2 text-sm font-medium text-[var(--foreground)]">
                      Discovered Tools ({preview.tools.length})
                    </p>
                    <div className="max-h-48 space-y-1.5 overflow-y-auto rounded-xl border border-[var(--card-border)] bg-[var(--background-tertiary)] p-3">
                      {preview.tools.map((tool) => (
                        <div
                          key={tool.name}
                          className="flex items-center justify-between rounded-lg bg-[var(--card)] px-3 py-2"
                        >
                          <div>
                            <p className="text-sm font-medium text-[var(--foreground)]">
                              {tool.name}
                            </p>
                            <p className="text-xs text-[var(--foreground-muted)]">
                              {tool.description}
                            </p>
                          </div>
                          <span className="rounded-full bg-blue-500/10 px-2 py-0.5 text-[10px] font-semibold uppercase text-blue-600">
                            {tool.method}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {error && (
                    <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-600">
                      {error}
                    </p>
                  )}
                </motion.div>
              )}

              {/* Step 3: Success */}
              {step === "success" && (
                <motion.div
                  key="success"
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  className="py-6 text-center"
                >
                  <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-green-500/10">
                    <Check className="h-7 w-7 text-green-600" />
                  </div>
                  <h4 className="mt-4 text-base font-semibold text-[var(--foreground)]">
                    UTCP Provider Connected
                  </h4>
                  <p className="mt-1 text-sm text-[var(--foreground-muted)]">
                    {preview?.tools.length} tool{preview?.tools.length !== 1 ? "s" : ""} added to your agent.
                  </p>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Footer */}
          <div className="flex justify-end gap-3 border-t border-[var(--card-border)] px-6 py-4">
            {step === "success" ? (
              <button
                onClick={onClose}
                className="rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 px-5 py-2 text-sm font-medium text-white transition-all hover:from-primary-700 hover:to-primary-800"
              >
                Done
              </button>
            ) : (
              <>
                <button
                  onClick={onClose}
                  className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] px-4 py-2 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)]"
                >
                  Cancel
                </button>

                {step === "url" && (
                  <button
                    onClick={handleDiscover}
                    disabled={!manualUrl.trim() || discover.isPending}
                    className="flex items-center gap-2 rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 px-5 py-2 text-sm font-medium text-white transition-all hover:from-primary-700 hover:to-primary-800 disabled:opacity-50"
                  >
                    {discover.isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Search className="h-4 w-4" />
                    )}
                    Discover
                  </button>
                )}

                {(step === "auth" || step === "preview") && (
                  <button
                    onClick={handleConnect}
                    disabled={!manualName.trim() || connect.isPending}
                    className="flex items-center gap-2 rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 px-5 py-2 text-sm font-medium text-white transition-all hover:from-primary-700 hover:to-primary-800 disabled:opacity-50"
                  >
                    {connect.isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Link2 className="h-4 w-4" />
                    )}
                    Connect
                  </button>
                )}
              </>
            )}
          </div>
        </div>
      </motion.div>
    </div>
  );
}
