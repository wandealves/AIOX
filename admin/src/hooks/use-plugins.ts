"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { ApiResponse, Plugin, RegisterPluginRequest } from "@/lib/types";

export function usePlugins() {
  return useQuery({
    queryKey: ["plugins"],
    queryFn: () => api.get<ApiResponse<Plugin[]>>("/api/v1/plugins"),
  });
}

export function useRegisterPlugin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: RegisterPluginRequest) =>
      api.post<ApiResponse<Plugin>>("/api/v1/plugins/register", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["plugins"] });
    },
  });
}

export function useDeletePlugin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (pluginId: string) =>
      api.delete<ApiResponse<null>>(`/api/v1/plugins/${pluginId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["plugins"] });
    },
  });
}

export function useSyncPlugin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (pluginId: string) =>
      api.post<ApiResponse<Plugin>>(`/api/v1/plugins/${pluginId}/sync`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["plugins"] });
    },
  });
}
