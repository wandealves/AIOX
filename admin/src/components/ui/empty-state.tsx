import type { LucideIcon } from "lucide-react";

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description: string;
  action?: {
    label: string;
    href?: string;
    onClick?: () => void;
  };
}

export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-[var(--card-border)] bg-[var(--card)] px-8 py-16">
      <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary-500/10">
        <Icon className="h-7 w-7 text-primary-500" />
      </div>
      <h3 className="mt-4 text-base font-semibold text-[var(--foreground)]">
        {title}
      </h3>
      <p className="mt-1 text-sm text-[var(--foreground-muted)]">
        {description}
      </p>
      {action && (
        action.href ? (
          <a
            href={action.href}
            className="mt-5 inline-flex items-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
          >
            {action.label}
          </a>
        ) : (
          <button
            onClick={action.onClick}
            className="mt-5 inline-flex items-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
          >
            {action.label}
          </button>
        )
      )}
    </div>
  );
}
