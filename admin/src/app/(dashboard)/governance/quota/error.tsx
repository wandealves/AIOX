"use client";

export default function QuotaError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="py-12 text-center">
      <h2 className="mb-2 text-lg font-semibold text-[var(--foreground)]">
        Something went wrong
      </h2>
      <p className="mb-4 text-sm text-[var(--foreground-muted)]">
        {error.message}
      </p>
      <button
        onClick={reset}
        className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
      >
        Try again
      </button>
    </div>
  );
}
