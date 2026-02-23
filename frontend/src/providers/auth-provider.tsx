"use client";

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import type { ApiResponse, TokenPair } from "@/lib/types";

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  userEmail: string | null;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [userEmail, setUserEmail] = useState<string | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    const email = localStorage.getItem("user_email");
    setIsAuthenticated(!!token);
    setUserEmail(email);
    setIsLoading(false);
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
        setIsAuthenticated(true);
        setUserEmail(email);
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
      await login(email, password);
    },
    [login]
  );

  const logout = useCallback(async () => {
    try {
      await api.post("/api/v1/auth/logout");
    } catch {
      // Ignore
    }
    api.clearTokens();
    localStorage.removeItem("user_email");
    setIsAuthenticated(false);
    setUserEmail(null);
    router.push("/login");
  }, [router]);

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, isLoading, userEmail, login, register, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuthContext() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuthContext must be used within AuthProvider");
  return ctx;
}
