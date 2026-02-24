interface BadgeProps {
  children: React.ReactNode;
  variant?: "info" | "warning" | "error" | "success" | "default" | "purple";
  className?: string;
}

const variantStyles: Record<string, string> = {
  info: "bg-blue-500/10 text-blue-600 dark:text-blue-400",
  warning: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
  error: "bg-red-500/10 text-red-600 dark:text-red-400",
  success: "bg-green-500/10 text-green-600 dark:text-green-400",
  purple: "bg-purple-500/10 text-purple-600 dark:text-purple-400",
  default: "bg-[var(--background-tertiary)] text-[var(--foreground-muted)]",
};

export function Badge({ children, variant = "default", className = "" }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${variantStyles[variant]} ${className}`}
    >
      {children}
    </span>
  );
}
