/**
 * Contribution Summary Bar Component
 *
 * A compact horizontal strip showing key contribution metrics
 * with animated counters and trend indicators.
 */

import { motion } from "framer-motion";
import {
  Activity,
  TrendingUp,
  Flame,
  Calendar,
} from "lucide-react";
import { cn, formatNumber } from "@/lib/utils";

export interface ContributionSummaryBarProps {
  totalContributions: number;
  currentStreak: number;
  longestStreak: number;
  activeDays: number;
  totalDays?: number;
}

interface StatItemProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  iconClass?: string;
  delay?: number;
}

function StatItem({ icon, label, value, iconClass = "text-brand-400 ca-summary-icon-brand", delay = 0 }: StatItemProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.3 }}
      className="flex items-center gap-2 px-3 py-2"
    >
      <div className={cn("flex-shrink-0", iconClass)}>{icon}</div>
      <div className="min-w-0">
        <div className="text-sm font-bold font-mono text-text-primary tabular-nums leading-tight ca-summary-value">
          {typeof value === "number" ? formatNumber(value) : value}
        </div>
        <div className="text-[10px] text-text-muted uppercase tracking-wider ca-summary-label">
          {label}
        </div>
      </div>
    </motion.div>
  );
}

export function ContributionSummaryBar({
  totalContributions,
  currentStreak,
  longestStreak,
  activeDays,
  totalDays = 365,
}: ContributionSummaryBarProps) {
  const activityRate = totalDays > 0 ? Math.round((activeDays / totalDays) * 100) : 0;

  return (
    <div className="flex items-center gap-1 p-1.5 rounded-xl bg-surface-secondary/50 border border-border-subtle overflow-x-auto ca-summary-bar">
      <StatItem
        icon={<Activity className="w-4 h-4" />}
        label="Total"
        value={totalContributions}
        iconClass="text-brand-400 ca-summary-icon-brand"
        delay={0}
      />

      <div className="w-px h-8 bg-border-subtle/50 flex-shrink-0 ca-summary-divider" />

      <StatItem
        icon={<Flame className="w-4 h-4" />}
        label="Streak"
        value={`${currentStreak}d`}
        iconClass="text-amber-400 ca-summary-icon-amber"
        delay={0.05}
      />

      <div className="w-px h-8 bg-border-subtle/50 flex-shrink-0 ca-summary-divider" />

      <StatItem
        icon={<TrendingUp className="w-4 h-4" />}
        label="Best"
        value={`${longestStreak}d`}
        iconClass="text-emerald-400 ca-summary-icon-emerald"
        delay={0.1}
      />

      <div className="w-px h-8 bg-border-subtle/50 flex-shrink-0 ca-summary-divider" />

      <StatItem
        icon={<Calendar className="w-4 h-4" />}
        label="Active"
        value={`${activityRate}%`}
        iconClass="text-blue-400 ca-summary-icon-blue"
        delay={0.15}
      />
    </div>
  );
}
