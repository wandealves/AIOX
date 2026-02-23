"use client";

import { useQuota } from "@/hooks/use-governance";
import { QuotaChart } from "@/components/governance/quota-chart";

export default function QuotaPage() {
  const { data, isLoading } = useQuota();
  const quota = data?.data;

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  if (!quota) {
    return <div className="text-center text-gray-500">No quota data</div>;
  }

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-gray-900">
        Quota Usage
      </h2>
      <QuotaChart quota={quota} />
    </div>
  );
}
