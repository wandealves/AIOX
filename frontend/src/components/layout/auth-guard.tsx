"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthContext } from "@/providers/auth-provider";
import { Skeleton } from "@/components/ui/skeleton";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuthContext();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="flex h-screen">
        {/* Sidebar skeleton */}
        <div className="hidden w-64 flex-col border-r border-[var(--card-border)] bg-[var(--background-secondary)] p-4 md:flex">
          <div className="flex items-center gap-2 px-2 py-3">
            <Skeleton className="h-8 w-8 rounded-lg" />
            <Skeleton className="h-5 w-20" variant="text" />
          </div>
          <div className="mt-6 space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-9 w-full rounded-lg" />
            ))}
          </div>
        </div>
        {/* Content skeleton */}
        <div className="flex flex-1 flex-col">
          <div className="flex h-14 items-center border-b border-[var(--card-border)] bg-[var(--card)] px-6">
            <Skeleton className="h-5 w-32" variant="text" />
          </div>
          <div className="flex-1 bg-[var(--background-secondary)] p-6">
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-24 w-full rounded-xl" />
              ))}
            </div>
            <Skeleton className="mt-6 h-64 w-full rounded-xl" />
          </div>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) return null;

  return <>{children}</>;
}
