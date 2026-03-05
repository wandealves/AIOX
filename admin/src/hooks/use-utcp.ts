"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type {
  ApiResponse,
  UTCPManual,
  UTCPDiscoverResult,
  ConnectManualRequest,
} from "@/lib/types";

export function useUTCPManuals(agentId: string) {
  return useQuery({
    queryKey: ["utcp-manuals", agentId],
    queryFn: () =>
      api.get<ApiResponse<UTCPManual[]>>(
        `/api/v1/agents/${agentId}/utcp/manuals`
      ),
    enabled: !!agentId,
  });
}

export function useUTCPDiscover(agentId: string) {
  return useMutation({
    mutationFn: (manualUrl: string) =>
      api.post<ApiResponse<UTCPDiscoverResult>>(
        `/api/v1/agents/${agentId}/utcp/discover`,
        { manual_url: manualUrl }
      ),
  });
}

export function useUTCPConnect(agentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: ConnectManualRequest) =>
      api.post<ApiResponse<UTCPManual>>(
        `/api/v1/agents/${agentId}/utcp/manuals`,
        data
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["utcp-manuals", agentId] });
      queryClient.invalidateQueries({ queryKey: ["tools", agentId] });
    },
  });
}

export function useUTCPSync(agentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (manualId: string) =>
      api.post<ApiResponse<UTCPManual>>(
        `/api/v1/agents/${agentId}/utcp/manuals/${manualId}/sync`
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["utcp-manuals", agentId] });
      queryClient.invalidateQueries({ queryKey: ["tools", agentId] });
    },
  });
}

export function useUTCPDelete(agentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (manualId: string) =>
      api.delete<ApiResponse<null>>(
        `/api/v1/agents/${agentId}/utcp/manuals/${manualId}`
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["utcp-manuals", agentId] });
      queryClient.invalidateQueries({ queryKey: ["tools", agentId] });
    },
  });
}
