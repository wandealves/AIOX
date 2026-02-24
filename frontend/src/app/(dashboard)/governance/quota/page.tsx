"use client";

import { useQuota } from "@/hooks/use-governance";
import { QuotaChart } from "@/components/governance/quota-chart";
import { Skeleton } from "@/components/ui/skeleton";

export default function QuotaPage() {
  const { data, isLoading, isError } = useQuota();
  const quota = data?.data;

  if (isLoading) {
    return (
      <div>
        <h2 className="mb-6 text-lg font-semibold text-[var(--foreground)]">
          Quota Usage
        </h2>
        <div className="grid gap-4 md:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-[180px] w-full rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  if (isError || !quota) {
    return (
      <div className="py-12 text-center text-[var(--foreground-muted)]">
        {isError ? "Failed to load quota data" : "No quota data"}
      </div>
    );
  }

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-[var(--foreground)]">
        Quota Usage
      </h2>
      <QuotaChart quota={quota} />
    </div>
  );
}
