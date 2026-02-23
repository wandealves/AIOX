"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type {
  ApiResponse,
  PaginatedResponse,
  Agent,
  CreateAgentRequest,
  UpdateAgentRequest,
} from "@/lib/types";

export function useAgents(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["agents", page, pageSize],
    queryFn: () =>
      api.get<PaginatedResponse<Agent>>(
        `/api/v1/agents?page=${page}&page_size=${pageSize}`
      ),
  });
}

export function useAgent(id: string) {
  return useQuery({
    queryKey: ["agent", id],
    queryFn: () => api.get<ApiResponse<Agent>>(`/api/v1/agents/${id}`),
    enabled: !!id,
  });
}

export function useCreateAgent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateAgentRequest) =>
      api.post<ApiResponse<Agent>>("/api/v1/agents", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agents"] });
    },
  });
}

export function useUpdateAgent(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateAgentRequest) =>
      api.put<ApiResponse<Agent>>(`/api/v1/agents/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agents"] });
      queryClient.invalidateQueries({ queryKey: ["agent", id] });
    },
  });
}

export function useDeleteAgent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.delete<ApiResponse<null>>(`/api/v1/agents/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agents"] });
    },
  });
}
