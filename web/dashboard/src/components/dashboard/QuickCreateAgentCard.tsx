import { Plus, Bot } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export interface QuickCreateAgentCardProps {
  title?: string;
  description?: string;
  /** Callback when card is clicked (e.g. navigate to create flow) */
  onCreateClick?: () => void;
  /** Primary action label */
  actionLabel?: string;
  className?: string;
}

export function QuickCreateAgentCard({
  title = "Create agent",
  description = "Deploy a new agent in seconds.",
  onCreateClick,
  actionLabel = "Create agent",
  className,
}: QuickCreateAgentCardProps) {
  return (
    <Card
      className={cn(
        "border-theme bg-card cursor-pointer transition-all duration-200",
        "hover:border-[var(--color-brand-500)]/40 hover:bg-bg-hover/50",
        "border-dashed",
        className
      )}
      onClick={onCreateClick}
      role={onCreateClick ? "button" : undefined}
      tabIndex={onCreateClick ? 0 : undefined}
      onKeyDown={
        onCreateClick
          ? (e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onCreateClick();
              }
            }
          : undefined
      }
    >
      <CardContent className="flex items-center gap-4 p-5">
        <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-brand-500/15 text-brand-400">
          <Bot className="h-6 w-6" />
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold text-text-primary">{title}</h3>
          {description && (
            <p className="mt-0.5 text-sm text-text-secondary">{description}</p>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-1.5 text-sm font-medium text-brand-400">
          <Plus className="h-4 w-4" />
          <span>{actionLabel}</span>
        </div>
      </CardContent>
    </Card>
  );
}
