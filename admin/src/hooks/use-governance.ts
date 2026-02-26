"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { ApiResponse, PaginatedResponse, QuotaInfo, AuditLog } from "@/lib/types";

export function useQuota() {
  return useQuery({
    queryKey: ["quota"],
    queryFn: () => api.get<ApiResponse<QuotaInfo>>("/api/v1/governance/quota"),
    refetchInterval: 30000, // refresh every 30s
  });
}

export function useAuditLogs(page = 1, pageSize = 20, filters?: { event_type?: string; severity?: string }) {
  const params = new URLSearchParams();
  params.set("page", page.toString());
  params.set("page_size", pageSize.toString());
  if (filters?.event_type) params.set("event_type", filters.event_type);
  if (filters?.severity) params.set("severity", filters.severity);

  return useQuery({
    queryKey: ["audit-logs", page, pageSize, filters],
    queryFn: () => api.get<PaginatedResponse<AuditLog>>(`/api/v1/governance/audit?${params}`),
  });
}

export function useAgentAuditLogs(agentId: string, page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["agent-audit-logs", agentId, page, pageSize],
    queryFn: () => api.get<PaginatedResponse<AuditLog>>(`/api/v1/agents/${agentId}/audit?page=${page}&page_size=${pageSize}`),
    enabled: !!agentId,
  });
}
