import { formatDistanceToNow } from "date-fns";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { ProviderIcon } from "./ProviderIcon";

interface ActivityItem {
  id: string;
  type: "deployment" | "error" | "success" | "info" | "warning";
  title: string;
  description?: string;
  timestamp: Date | string;
  provider?: string;
  function?: string;
  metadata?: Record<string, any>;
}

interface ActivityFeedProps {
  activities: ActivityItem[];
  title?: string;
  maxItems?: number;
  className?: string;
}

const activityIcons = {
  deployment: "🚀",
  error: "❌",
  success: "✅",
  info: "ℹ️",
  warning: "⚠️",
};

const activityColors = {
  deployment: "text-blue-400",
  error: "text-red-400",
  success: "text-green-400",
  info: "text-blue-400",
  warning: "text-yellow-400",
};

export function ActivityFeed({
  activities,
  title = "Recent Activity",
  maxItems = 10,
  className,
}: ActivityFeedProps) {
  const displayActivities = activities.slice(0, maxItems);

  if (displayActivities.length === 0) {
    return (
      <Card className={cn("", className)}>
        <CardHeader>
          <CardTitle className="text-lg">{title}</CardTitle>
        </CardHeader>
          <CardContent>
            <div className="text-center py-8 text-text-secondary">
              <p>No recent activity</p>
            </div>
          </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn("", className)}>
      <CardHeader>
        <CardTitle className="text-lg">{title}</CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <div className="space-y-0">
          {displayActivities.map((activity, index) => {
            const isLast = index === displayActivities.length - 1;

            return (
              <div
                key={activity.id}
                className={cn(
                  "flex items-start gap-4 p-4 border-l-2 transition-colors hover:bg-[#1a1a1a]/50",
                  activity.type === "error" && "border-l-red-400",
                  activity.type === "success" && "border-l-green-400",
                  activity.type === "warning" && "border-l-yellow-400",
                  activity.type === "deployment" && "border-l-blue-400",
                  activity.type === "info" && "border-l-blue-400",
                  !isLast && "border-b border-[#2a2a2a]"
                )}
              >
                {/* Timeline dot */}
                <div className="flex flex-col items-center">
                  <div
                    className={cn(
                      "w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium",
                      activity.type === "error" && "bg-red-400/20 text-red-400",
                      activity.type === "success" && "bg-green-400/20 text-green-400",
                      activity.type === "warning" && "bg-yellow-400/20 text-yellow-400",
                      activity.type === "deployment" && "bg-blue-400/20 text-blue-400",
                      activity.type === "info" && "bg-blue-400/20 text-blue-400"
                    )}
                  >
                    {activity.provider ? (
                      <ProviderIcon provider={activity.provider} size="sm" />
                    ) : (
                      activityIcons[activity.type]
                    )}
                  </div>
                  {!isLast && (
                    <div className="w-px h-8 bg-[#2a2a2a] mt-2" />
                  )}
                </div>

                {/* Activity content */}
                <div className="flex-1 min-w-0 space-y-1">
                  <div className="flex items-center justify-between">
                    <h4 className="text-sm font-medium text-white truncate">
                      {activity.title}
                    </h4>
                    <span className="text-xs text-text-secondary ml-2 whitespace-nowrap">
                      {formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true })}
                    </span>
                  </div>

                  {activity.description && (
                    <p className="text-sm text-text-secondary line-clamp-2">
                      {activity.description}
                    </p>
                  )}

                  {activity.function && (
                    <div className="flex items-center gap-2 mt-2">
                      <span className="text-xs px-2 py-1 bg-[#2a2a2a] rounded-md text-text-secondary">
                        {activity.function}
                      </span>
                      {activity.provider && (
                        <ProviderIcon provider={activity.provider} size="sm" />
                      )}
                    </div>
                  )}

                  {activity.metadata && Object.keys(activity.metadata).length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {Object.entries(activity.metadata).map(([key, value]) => (
                        <span
                          key={key}
                          className="text-xs px-2 py-1 bg-[#1a1a1a] rounded text-text-muted"
                        >
                          {key}: {String(value)}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}