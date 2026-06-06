import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";
import {
    AlertCircle,
    Bot,
    CheckCircle2,
    Clock,
    XCircle,
    Zap,
} from "lucide-react";

export type AgentActivityType =
  | "invocation"
  | "success"
  | "error"
  | "timeout"
  | "deploy"
  | "info";

export interface AgentActivityItem {
  id: string;
  type: AgentActivityType;
  title: string;
  description?: string;
  timestamp: Date | string;
  agentId?: string;
  agentName?: string;
  metadata?: Record<string, unknown>;
}

export interface AgentActivityFeedProps {
  activities: AgentActivityItem[];
  title?: string;
  maxItems?: number;
  className?: string;
}

const typeConfig: Record<
  AgentActivityType,
  { icon: React.ReactNode; borderClass: string; iconBgClass: string }
> = {
  invocation: {
    icon: <Zap className="h-4 w-4" />,
    borderClass: "border-l-[var(--color-brand-500)]",
    iconBgClass: "bg-brand-500/20 text-brand-400",
  },
  success: {
    icon: <CheckCircle2 className="h-4 w-4" />,
    borderClass: "border-l-[var(--color-success)]",
    iconBgClass: "bg-[var(--color-success)]/20 text-[var(--color-success)]",
  },
  error: {
    icon: <XCircle className="h-4 w-4" />,
    borderClass: "border-l-[var(--color-error)]",
    iconBgClass: "bg-[var(--color-error)]/20 text-[var(--color-error)]",
  },
  timeout: {
    icon: <Clock className="h-4 w-4" />,
    borderClass: "border-l-[var(--color-warning)]",
    iconBgClass: "bg-[var(--color-warning)]/20 text-[var(--color-warning)]",
  },
  deploy: {
    icon: <Bot className="h-4 w-4" />,
    borderClass: "border-l-blue-400",
    iconBgClass: "bg-blue-500/20 text-blue-400",
  },
  info: {
    icon: <AlertCircle className="h-4 w-4" />,
    borderClass: "border-l-[var(--color-info)]",
    iconBgClass: "bg-[var(--color-info)]/20 text-[var(--color-info)]",
  },
};

export function AgentActivityFeed({
  activities,
  title = "Agent activity",
  maxItems = 10,
  className,
}: AgentActivityFeedProps) {
  const displayActivities = activities.slice(0, maxItems);

  return (
    <Card className={cn("", className)}>
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold text-text-primary">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        {displayActivities.length === 0 ? (
          <div className="py-8 text-center text-sm text-text-muted">
            No recent agent activity
          </div>
        ) : (
          <ul className="divide-y divide-[var(--color-border-subtle)]">
            {displayActivities.map((activity, index) => {
              const config = typeConfig[activity.type];
              return (
                <li
                  key={activity.id}
                  className={cn(
                    "flex items-start gap-3 border-l-2 py-3 px-4 transition-colors hover:bg-bg-hover/50",
                    config.borderClass,
                    index === 0 && "pt-0",
                    index === displayActivities.length - 1 && "pb-0"
                  )}
                >
                  <div
                    className={cn(
                      "mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg",
                      config.iconBgClass
                    )}
                  >
                    {config.icon}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-sm font-medium text-text-primary">
                        {activity.title}
                      </span>
                      <span className="shrink-0 text-xs text-text-muted">
                        {formatDistanceToNow(new Date(activity.timestamp), {
                          addSuffix: true,
                        })}
                      </span>
                    </div>
                    {activity.description && (
                      <p className="mt-0.5 text-xs text-text-secondary line-clamp-2">
                        {activity.description}
                      </p>
                    )}
                    {activity.agentName && (
                      <p className="mt-1 text-xs text-text-muted">
                        {activity.agentName}
                        {activity.agentId && ` · ${activity.agentId}`}
                      </p>
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
