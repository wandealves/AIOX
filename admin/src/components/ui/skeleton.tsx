interface SkeletonProps {
  className?: string;
  variant?: "text" | "circular" | "rectangular";
}

export function Skeleton({ className = "", variant = "rectangular" }: SkeletonProps) {
  const variantClasses = {
    text: "h-4 w-full rounded",
    circular: "rounded-full",
    rectangular: "",
  };

  return (
    <div
      className={`skeleton ${variantClasses[variant]} ${className}`}
      aria-hidden="true"
    />
  );
}
