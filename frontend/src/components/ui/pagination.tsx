import { ChevronLeft, ChevronRight } from "lucide-react";

interface PaginationProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export function Pagination({ page, totalPages, onPageChange }: PaginationProps) {
  if (totalPages <= 1) return null;

  return (
    <div className="mt-6 flex items-center justify-center gap-2">
      <button
        onClick={() => onPageChange(Math.max(1, page - 1))}
        disabled={page <= 1}
        className="flex items-center gap-1 rounded-lg border border-[var(--card-border)] bg-[var(--card)] px-3 py-1.5 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)] disabled:opacity-40 disabled:pointer-events-none"
      >
        <ChevronLeft className="h-4 w-4" />
        Previous
      </button>
      <span className="px-3 py-1.5 text-sm text-[var(--foreground-muted)]">
        Page {page} of {totalPages}
      </span>
      <button
        onClick={() => onPageChange(Math.min(totalPages, page + 1))}
        disabled={page >= totalPages}
        className="flex items-center gap-1 rounded-lg border border-[var(--card-border)] bg-[var(--card)] px-3 py-1.5 text-sm font-medium text-[var(--foreground)] transition-colors hover:bg-[var(--background-tertiary)] disabled:opacity-40 disabled:pointer-events-none"
      >
        Next
        <ChevronRight className="h-4 w-4" />
      </button>
    </div>
  );
}
