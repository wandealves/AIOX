"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import type { ApiResponse, TokenPair } from "@/lib/types";

interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  userEmail: string | null;
}

export function useAuth() {
  const router = useRouter();
  const [state, setState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    userEmail: null,
  });

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    const email = localStorage.getItem("user_email");
    setState({
      isAuthenticated: !!token,
      isLoading: false,
      userEmail: email,
    });
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await api.post<ApiResponse<TokenPair>>(
        "/api/v1/auth/login",
        { email, password }
      );
      if (res.data) {
        localStorage.setItem("access_token", res.data.access_token);
        localStorage.setItem("refresh_token", res.data.refresh_token);
        localStorage.setItem("user_email", email);
        setState({ isAuthenticated: true, isLoading: false, userEmail: email });
        router.push("/dashboard");
      }
    },
    [router]
  );

  const register = useCallback(
    async (email: string, password: string) => {
      await api.post<ApiResponse<TokenPair>>("/api/v1/auth/register", {
        email,
        password,
      });
      // Auto-login after registration
      await login(email, password);
    },
    [login]
  );

  const logout = useCallback(async () => {
    try {
      await api.post("/api/v1/auth/logout");
    } catch {
      // Ignore logout errors
    }
    api.clearTokens();
    localStorage.removeItem("user_email");
    setState({ isAuthenticated: false, isLoading: false, userEmail: null });
    router.push("/login");
  }, [router]);

  return { ...state, login, register, logout };
}
