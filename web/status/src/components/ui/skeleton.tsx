import { cn } from '@/lib/utils';

function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "animate-pulse rounded-md bg-bg-tertiary",
        "bg-gradient-to-r from-bg-tertiary via-bg-elevated to-bg-tertiary",
        "bg-[length:200%_100%]",
        "animate-shimmer",
        className
      )}
      {...props}
    />
  );
}

export { Skeleton };
