"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { ApiResponse, PaginatedResponse, User, Agent, AuditLog, AdminStats } from "@/lib/types";

export function useAdminStats() {
  return useQuery({
    queryKey: ["admin-stats"],
    queryFn: () => api.get<ApiResponse<AdminStats>>("/api/v1/admin/stats"),
  });
}

export function useAdminUsers(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["admin-users", page, pageSize],
    queryFn: () => api.get<PaginatedResponse<User>>(`/api/v1/admin/users?page=${page}&page_size=${pageSize}`),
  });
}

export function useAdminAgents(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["admin-agents", page, pageSize],
    queryFn: () => api.get<PaginatedResponse<Agent>>(`/api/v1/admin/agents?page=${page}&page_size=${pageSize}`),
  });
}

export function useAdminUpdateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      api.put<ApiResponse<null>>(`/api/v1/admin/users/${userId}/role`, { role }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin-users"] }),
  });
}

export function useAdminDisableUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) =>
      api.post<ApiResponse<null>>(`/api/v1/admin/users/${userId}/disable`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin-users"] }),
  });
}

export function useAdminEnableUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) =>
      api.post<ApiResponse<null>>(`/api/v1/admin/users/${userId}/enable`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin-users"] }),
  });
}

export function useAdminAuditLogs(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["admin-audit-logs", page, pageSize],
    queryFn: () => api.get<PaginatedResponse<AuditLog>>(`/api/v1/admin/audit?page=${page}&page_size=${pageSize}`),
  });
}
