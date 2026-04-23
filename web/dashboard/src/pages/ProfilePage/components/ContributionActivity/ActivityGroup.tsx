/**
 * Activity Group Component
 *
 * Groups activities by time period (Today, Yesterday, This Week, etc.)
 * and renders them with a section header and staggered animations.
 */

import { useMemo } from "react";
import { motion } from "framer-motion";
import { isToday, isYesterday, isThisWeek, isThisMonth, format } from "date-fns";
import { Calendar } from "lucide-react";
import { cn } from "@/lib/utils";
import { ActivityCard } from "./ActivityCard";
import type { UserActivity } from "@/types";

type TimeGroup = {
  label: string;
  activities: UserActivity[];
};

function groupByTimePeriod(activities: UserActivity[]): TimeGroup[] {
  const groups = new Map<string, UserActivity[]>();

  for (const activity of activities) {
    const date = new Date(activity.timestamp);
    let key: string;

    if (isToday(date)) {
      key = "Today";
    } else if (isYesterday(date)) {
      key = "Yesterday";
    } else if (isThisWeek(date)) {
      key = "This Week";
    } else if (isThisMonth(date)) {
      key = "This Month";
    } else {
      key = format(date, "MMMM yyyy");
    }

    if (!groups.has(key)) {
      groups.set(key, []);
    }
    groups.get(key)!.push(activity);
  }

  const order = ["Today", "Yesterday", "This Week", "This Month"];
  const sorted: TimeGroup[] = [];

  for (const label of order) {
    const items = groups.get(label);
    if (items && items.length > 0) {
      sorted.push({ label, activities: items });
      groups.delete(label);
    }
  }

  // Remaining groups sorted by most recent first
  const remaining = Array.from(groups.entries())
    .map(([label, activities]) => ({ label, activities }))
    .sort((a, b) => {
      const dateA = new Date(a.activities[0].timestamp);
      const dateB = new Date(b.activities[0].timestamp);
      return dateB.getTime() - dateA.getTime();
    });

  sorted.push(...remaining);

  return sorted;
}

const ACTIVE_GROUPS = new Set(["Today", "Yesterday"]);

export interface ActivityGroupProps {
  activities: UserActivity[];
  compact?: boolean;
  maxHeight?: number;
}

export function ActivityGroup({
  activities,
  compact = false,
  maxHeight,
}: ActivityGroupProps) {
  const groups = useMemo(() => groupByTimePeriod(activities), [activities]);

  return (
    <div
      className="space-y-6"
      style={maxHeight ? { maxHeight, overflowY: "auto" } : undefined}
    >
      {groups.map((group, groupIndex) => (
        <motion.div
          key={group.label}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: groupIndex * 0.1 }}
        >
          {/* Group header */}
          <div className="flex items-center gap-2 mb-3">
            <Calendar
              className={cn(
                "w-4 h-4",
                ACTIVE_GROUPS.has(group.label)
                  ? "text-brand-400 ca-group-header-icon-active"
                  : "text-text-muted ca-group-header-icon"
              )}
            />
            <h3 className="text-sm font-semibold text-text-secondary ca-group-header-text">
              {group.label}
            </h3>
            <span className="text-xs text-text-muted font-mono tabular-nums ca-group-header-count">
              {group.activities.length}{" "}
              {group.activities.length === 1 ? "event" : "events"}
            </span>
            <div className="flex-1 h-px bg-border-subtle/50 ml-2 ca-group-divider" />
          </div>

          {/* Activity cards */}
          <div
            className={cn(
              "space-y-2",
              !compact && "ml-6"
            )}
          >
            {group.activities.map((activity, index) => (
              <ActivityCard
                key={activity.id}
                activity={activity}
                index={index}
                compact={compact}
              />
            ))}
          </div>
        </motion.div>
      ))}
    </div>
  );
}
