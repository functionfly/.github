import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export interface TrustIndicatorBadgeProps {
  tier?: 1 | 2 | 3 | 4 | 5 | 6;
  label?: string;
  deterministic?: boolean;
  className?: string;
}

const tierLabels: Record<number, string> = {
  1: "Bronze",
  2: "Silver",
  3: "Gold",
  4: "Platinum",
  5: "Diamond",
  6: "Legend",
};

export function TrustIndicatorBadge({
  tier = 1,
  label,
  deterministic,
  className,
}: TrustIndicatorBadgeProps) {
  const text = label ?? tierLabels[tier] ?? "Builder";
  return (
    <span className={cn("inline-flex items-center gap-1", className)}>
      <Badge variant="secondary" className="text-xs">
        {text}
      </Badge>
      {deterministic && (
        <Badge variant="outline" className="text-xs border-green-500/50 text-green-700 dark:text-green-400">
          Deterministic
        </Badge>
      )}
    </span>
  );
}
