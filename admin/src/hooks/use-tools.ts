"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type {
  ApiResponse,
  ToolDefinition,
  CreateToolRequest,
  UpdateToolRequest,
} from "@/lib/types";

export function useTools(agentId: string) {
  return useQuery({
    queryKey: ["tools", agentId],
    queryFn: () =>
      api.get<ApiResponse<ToolDefinition[]>>(
        `/api/v1/agents/${agentId}/tools`
      ),
    enabled: !!agentId,
  });
}

export function useCreateTool(agentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateToolRequest) =>
      api.post<ApiResponse<ToolDefinition>>(
        `/api/v1/agents/${agentId}/tools`,
        data
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tools", agentId] });
    },
  });
}

export function useUpdateTool(agentId: string, toolId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateToolRequest) =>
      api.put<ApiResponse<ToolDefinition>>(
        `/api/v1/agents/${agentId}/tools/${toolId}`,
        data
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tools", agentId] });
    },
  });
}

export function useDeleteTool(agentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (toolId: string) =>
      api.delete<ApiResponse<null>>(
        `/api/v1/agents/${agentId}/tools/${toolId}`
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tools", agentId] });
    },
  });
}

export function useToggleToolActive(agentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ toolId, isActive }: { toolId: string; isActive: boolean }) =>
      api.put<ApiResponse<ToolDefinition>>(
        `/api/v1/agents/${agentId}/tools/${toolId}`,
        { is_active: isActive }
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tools", agentId] });
    },
  });
}
