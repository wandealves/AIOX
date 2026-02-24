"use client";

import { useState } from "react";
import Link from "next/link";
import { Bot } from "lucide-react";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { useAuthContext } from "@/providers/auth-provider";

export default function LoginPage() {
  const { login } = useAuthContext();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    try {
      await login(email, password);
      toast.success("Welcome back!");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Login failed";
      setError(msg);
      toast.error(msg);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-gradient-to-br from-primary-600 via-primary-700 to-primary-900">
      {/* Decorative circles */}
      <div className="pointer-events-none absolute -left-32 -top-32 h-96 w-96 rounded-full bg-white/5 blur-3xl" />
      <div className="pointer-events-none absolute -bottom-32 -right-32 h-96 w-96 rounded-full bg-white/5 blur-3xl" />
      <div className="pointer-events-none absolute left-1/2 top-1/3 h-64 w-64 -translate-x-1/2 rounded-full bg-primary-400/10 blur-2xl" />

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="relative z-10 w-full max-w-md px-4"
      >
        <div className="mb-8 text-center">
          <motion.div
            initial={{ scale: 0.8 }}
            animate={{ scale: 1 }}
            transition={{ type: "spring", stiffness: 200, damping: 15, delay: 0.1 }}
            className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-white/10 shadow-lg backdrop-blur"
          >
            <Bot className="h-8 w-8 text-white" />
          </motion.div>
          <h1 className="mt-5 text-3xl font-bold text-white">AIOX</h1>
          <p className="mt-2 text-sm text-primary-200">
            Multi-agent AI platform
          </p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="rounded-2xl border border-white/10 bg-[var(--card)] p-8"
          style={{ boxShadow: "var(--shadow-xl)" }}
        >
          <h2 className="mb-6 text-center text-lg font-semibold text-[var(--foreground)]">
            Sign in to your account
          </h2>

          {error && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              className="mb-4 rounded-lg bg-red-500/10 border border-red-500/20 p-3 text-sm text-red-600"
            >
              {error}
            </motion.div>
          )}

          <div className="mb-4">
            <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-3 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
              placeholder="you@example.com"
            />
          </div>

          <div className="mb-6">
            <label className="mb-1.5 block text-sm font-medium text-[var(--foreground-muted)]">
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full rounded-xl border border-[var(--input-border)] bg-[var(--input-bg)] px-4 py-3 text-sm text-[var(--foreground)] outline-none transition-all focus:border-[var(--input-focus-border)] focus:ring-2 focus:ring-[var(--input-focus-ring)]"
            />
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full rounded-xl bg-gradient-to-r from-primary-600 to-primary-700 py-3 text-sm font-semibold text-white transition-all hover:from-primary-700 hover:to-primary-800 active:scale-[0.98] disabled:opacity-50"
            style={{ boxShadow: "var(--shadow-md)" }}
          >
            {isLoading ? "Signing in..." : "Sign in"}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-primary-200">
          Don&apos;t have an account?{" "}
          <Link
            href="/register"
            className="font-semibold text-white underline-offset-2 hover:underline"
          >
            Register
          </Link>
        </p>
      </motion.div>
    </div>
  );
}
