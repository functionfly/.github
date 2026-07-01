/**
 * Contribution Activity Component
 *
 * Main orchestrator that combines the heat ring, streak, summary bar,
 * and grouped activity feed into a unified "Contribution Activity" section.
 */

import { useState } from "react";
import { motion } from "framer-motion";
import { Activity, Flame, LayoutGrid, List, Filter } from "lucide-react";
import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ActivityHeatRing } from "./ActivityHeatRing";
import { ContributionStreak } from "./ContributionStreak";
import { ContributionSummaryBar } from "./ContributionSummaryBar";
import { ActivityGroup } from "./ActivityGroup";
import { ActivityCard } from "./ActivityCard";
import type { UserProfile, ActivityType } from "@/types";

type ViewMode = "grouped" | "list";

export interface ContributionActivityProps {
  profile: UserProfile;
  showFilter?: boolean;
  defaultFilter?: ActivityType | "all";
  maxActivities?: number;
  compact?: boolean;
}

export function ContributionActivity({
  profile,
  showFilter = false,
  defaultFilter = "all",
  maxActivities,
  compact = false,
}: ContributionActivityProps) {
  const [viewMode, setViewMode] = useState<ViewMode>("grouped");
  const [filter, setFilter] = useState<ActivityType | "all">(defaultFilter);

  const { stats, recentActivity } = profile;
  const { contributionStreak, contributionGraph } = stats;

  const totalContributions = contributionGraph.reduce(
    (sum, day) => sum + day.count,
    0
  );

  const activeDays = contributionGraph.filter((d) => d.count > 0).length;

  const filteredActivities =
    filter === "all"
      ? recentActivity
      : recentActivity.filter((a) => a.type === filter);

  const displayedActivities = maxActivities
    ? filteredActivities.slice(0, maxActivities)
    : filteredActivities;

  const filterOptions: { value: ActivityType | "all"; label: string }[] = [
    { value: "all", label: "All Activity" },
    { value: "joined", label: "Joined" },
    { value: "function_published", label: "Functions" },
    { value: "achievement_earned", label: "Achievements" },
    { value: "membership_upgraded", label: "Upgrades" },
    { value: "review_received", label: "Reviews" },
    { value: "milestone_reached", label: "Milestones" },
  ];

  return (
    <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)', overflow: 'hidden' }}>
      <CardHeader className="pb-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
            <div className="w-8 h-8 rounded-[var(--radius)] flex items-center justify-center" style={{ background: 'rgba(143, 255, 208, 0.08)' }}>
              <Activity className="w-4 h-4" style={{ color: 'var(--status-ok)' }} />
            </div>
            Contribution Activity
          </CardTitle>

          <div className="flex items-center gap-2">
            {showFilter && (
              <Select
                value={filter}
                onValueChange={(v) => setFilter(v as ActivityType | "all")}
              >
                <SelectTrigger className="w-[140px] h-8 text-xs">
                  <Filter className="w-3 h-3 mr-1.5" />
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {filterOptions.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}

            {/* View mode toggle */}
            <div className="flex items-center rounded-[var(--radius)] p-0.5" style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)' }}>
              <button
                onClick={() => setViewMode("grouped")}
                className="p-1.5 rounded-[var(--radius-sm)] transition-all text-xs"
                style={{
                  background: viewMode === "grouped" ? 'var(--panel-raised)' : 'transparent',
                  color: viewMode === "grouped" ? 'var(--text)' : 'var(--text-faint)',
                  boxShadow: viewMode === "grouped" ? '0 1px 2px rgba(0,0,0,0.2)' : 'none',
                }}
                title="Grouped view"
              >
                <LayoutGrid className="w-3.5 h-3.5" />
              </button>
              <button
                onClick={() => setViewMode("list")}
                className="p-1.5 rounded-[var(--radius-sm)] transition-all text-xs"
                style={{
                  background: viewMode === "list" ? 'var(--panel-raised)' : 'transparent',
                  color: viewMode === "list" ? 'var(--text)' : 'var(--text-faint)',
                  boxShadow: viewMode === "list" ? '0 1px 2px rgba(0,0,0,0.2)' : 'none',
                }}
                title="List view"
              >
                <List className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-5">
        {/* Visualization + Streak row */}
        <div className="flex flex-col md:flex-row items-center gap-4">
          {/* Heat ring */}
          <div className="flex-shrink-0">
            <ActivityHeatRing data={contributionGraph} size={200} />
          </div>

          {/* Streak + Legend */}
          <div className="flex-1 w-full space-y-3">
            <ContributionStreak
              current={contributionStreak.current}
              longest={contributionStreak.longest}
              lastContribution={contributionStreak.lastContribution}
            />
          </div>
        </div>

        {/* Summary bar */}
        <ContributionSummaryBar
          totalContributions={totalContributions}
          currentStreak={contributionStreak.current}
          longestStreak={contributionStreak.longest}
          activeDays={activeDays}
        />

        {/* Activity feed */}
        {displayedActivities.length > 0 && (
          <div>
            <div className="flex items-center gap-2 mb-3">
              <Flame className="w-4 h-4" style={{ color: 'var(--status-pending)' }} />
              <h3 className="text-sm font-semibold" style={{ color: 'var(--text-dim)' }}>
                Recent Activity
              </h3>
              <span className="text-xs font-mono tabular-nums" style={{ color: 'var(--text-faint)' }}>
                {displayedActivities.length}{" "}
                {displayedActivities.length === 1 ? "event" : "events"}
              </span>
            </div>

            {viewMode === "grouped" ? (
              <ActivityGroup
                activities={displayedActivities}
                compact={compact}
                maxHeight={480}
              />
            ) : (
              <div className="space-y-2 max-h-[480px] overflow-y-auto">
                {displayedActivities.map((activity, index) => (
                  <ActivityCard
                    key={activity.id}
                    activity={activity}
                    index={index}
                    compact={compact}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {displayedActivities.length === 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-8"
          >
            <Activity className="w-10 h-10 mx-auto mb-2" style={{ color: 'var(--text-faint)', opacity: 0.3 }} />
            <p className="text-sm" style={{ color: 'var(--text-faint)' }}>
              No activity to display
            </p>
          </motion.div>
        )}
      </CardContent>
    </Card>
  );
}
