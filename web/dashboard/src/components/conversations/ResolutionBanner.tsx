import { CheckCircle } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ResolutionBannerProps {
  /** Short message, e.g. "This solution improved error rate by 12%" */
  message?: string;
  /** Reputation points awarded (optional; when integrated with flywheel) */
  reputationAwarded?: number;
  resolvedAt: string;
  className?: string;
}

export function ResolutionBanner({
  message,
  reputationAwarded,
  resolvedAt,
  className,
}: ResolutionBannerProps) {
  return (
    <div
      className={cn(
        "flex flex-wrap items-center gap-2 rounded-lg border border-green-500/30 bg-green-500/10 px-3 py-2 text-sm",
        className
      )}
    >
      <CheckCircle className="h-4 w-4 text-green-600 shrink-0" />
      <span className="text-green-800 dark:text-green-200">
        {message ?? "Conversation resolved"}
      </span>
      {resolvedAt && (
        <span className="text-muted-foreground text-xs">
          {new Date(resolvedAt).toLocaleString()}
        </span>
      )}
      {reputationAwarded != null && reputationAwarded > 0 && (
        <span className="font-medium text-green-700 dark:text-green-300">
          +{reputationAwarded} Community Reputation awarded
        </span>
      )}
    </div>
  );
}
