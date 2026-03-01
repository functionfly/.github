import { cn } from "@/lib/utils";

export type DeterminismTier = "full" | "lite" | "partial" | "drifted";

export interface DeterminismBadgeProps {
  /** The determinism tier */
  tier: DeterminismTier;
  /** Show label */
  showLabel?: boolean;
  /** Size variant */
  size?: "sm" | "md" | "lg";
  /** Custom className */
  className?: string;
  /** Click handler */
  onClick?: () => void;
}

const tierConfig = {
  full: {
    label: "FULL",
    className: "bg-green-500/10 text-green-500 border-green-500/20",
    description: "Fully deterministic execution",
  },
  lite: {
    label: "LITE",
    className: "bg-blue-500/10 text-blue-500 border-blue-500/20",
    description: "Light determinism with some non-deterministic operations",
  },
  partial: {
    label: "PARTIAL",
    className: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
    description: "Partially deterministic - some operations may vary",
  },
  drifted: {
    label: "DRIFTED",
    className: "bg-red-500/10 text-red-500 border-red-500/20",
    description: "Execution drift detected",
  },
};

const sizeConfig = {
  sm: "text-xs px-2 py-0.5",
  md: "text-sm px-2.5 py-1",
  lg: "text-base px-3 py-1.5",
};

export function DeterminismBadge({
  tier,
  showLabel = true,
  size = "md",
  className,
  onClick,
}: DeterminismBadgeProps) {
  const config = tierConfig[tier];

  return (
    <span
      className={cn(
        "inline-flex items-center font-semibold rounded-md border",
        config.className,
        sizeConfig[size],
        onClick && "cursor-pointer hover:opacity-80 transition-opacity",
        className
      )}
      onClick={onClick}
      title={config.description}
    >
      {showLabel && config.label}
    </span>
  );
}
