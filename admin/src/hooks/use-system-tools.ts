"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type {
  ApiResponse,
  SystemTool,
  CreateSystemToolRequest,
  UpdateSystemToolRequest,
} from "@/lib/types";

export function useSystemTools() {
  return useQuery({
    queryKey: ["system-tools"],
    queryFn: () =>
      api.get<ApiResponse<SystemTool[]>>("/api/v1/admin/tools"),
  });
}

export function useCreateSystemTool() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSystemToolRequest) =>
      api.post<ApiResponse<SystemTool>>("/api/v1/admin/tools", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["system-tools"] });
    },
  });
}

export function useUpdateSystemTool(toolId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateSystemToolRequest) =>
      api.put<ApiResponse<SystemTool>>(
        `/api/v1/admin/tools/${toolId}`,
        data
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["system-tools"] });
    },
  });
}

export function useDeleteSystemTool() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (toolId: string) =>
      api.delete<ApiResponse<null>>(`/api/v1/admin/tools/${toolId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["system-tools"] });
    },
  });
}

export function useToggleSystemToolActive() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ toolId, isActive }: { toolId: string; isActive: boolean }) =>
      api.put<ApiResponse<SystemTool>>(
        `/api/v1/admin/tools/${toolId}`,
        { is_active: isActive }
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["system-tools"] });
    },
  });
}
