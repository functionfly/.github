/**
 * Stats Overview Component
 *
 * Displays a grid of user statistics (functions, executions, trust score, followers, etc.)
 */

import { motion } from "framer-motion";
import {
  Package,
  Activity,
  Shield,
  Users,
  Heart,
  Flame,
} from "lucide-react";
import { cn, formatNumber } from "@/lib/utils";
import { Card, CardContent } from "@/components/ui/card";
import { containerVariants, itemVariants } from "../animations";
import type { UserProfile } from "@/types";

export interface StatsOverviewProps {
  stats: UserProfile["stats"];
}

export function StatsOverview({ stats }: StatsOverviewProps) {
  const statItems = [
    {
      label: "Functions",
      value: stats.functionsPublished,
      trend: stats.functionsTrend,
      icon: Package,
      color: "text-blue-500",
      bgColor: "bg-blue-500/10",
    },
    {
      label: "Executions",
      value: stats.totalExecutions,
      trend: stats.executionsTrend,
      icon: Activity,
      color: "text-green-500",
      bgColor: "bg-green-500/10",
    },
    {
      label: "Trust Score",
      value: stats.trustScore,
      suffix: "%",
      trend: null,
      icon: Shield,
      color: stats.trustScore >= 80 ? "text-emerald-500" : stats.trustScore >= 60 ? "text-yellow-500" : "text-orange-500",
      bgColor: stats.trustScore >= 80 ? "bg-emerald-500/10" : stats.trustScore >= 60 ? "bg-yellow-500/10" : "bg-orange-500/10",
    },
    {
      label: "Followers",
      value: stats.followersCount,
      trend: stats.followersTrend,
      icon: Users,
      color: "text-purple-500",
      bgColor: "bg-purple-500/10",
    },
    {
      label: "Following",
      value: stats.followingCount,
      trend: null,
      icon: Heart,
      color: "text-pink-500",
      bgColor: "bg-pink-500/10",
    },
    {
      label: "Streak",
      value: stats.contributionStreak.current,
      suffix: " days",
      trend: null,
      icon: Flame,
      color: "text-orange-500",
      bgColor: "bg-orange-500/10",
      highlight: stats.contributionStreak.current >= 30,
    },
  ];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 px-4 md:px-8 py-4"
    >
      {statItems.map((item, index) => (
        <motion.div key={item.label} variants={itemVariants}>
          <Card
            className={cn(
              "group hover:shadow-lg transition-all duration-300 border-border-subtle hover:border-brand-500/30",
              item.highlight && "ring-2 ring-orange-500/30"
            )}
          >
            <CardContent className="p-4">
              <div className="flex items-center gap-2 mb-2">
                <div className={cn("p-1.5 rounded-md", item.bgColor, item.color)}>
                  <item.icon className="w-4 h-4" />
                </div>
                <span className="text-xs text-text-muted">{item.label}</span>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold font-mono tabular-nums text-text-primary">
                  {formatNumber(item.value)}{item.suffix || ""}
                </span>
                {item.trend !== null && item.trend !== undefined && (
                  <span
                    className={cn(
                      "text-xs font-medium",
                      item.trend >= 0 ? "text-green-500" : "text-red-500"
                    )}
                  >
                    {item.trend >= 0 ? "+" : ""}{item.trend}%
                  </span>
                )}
              </div>
            </CardContent>
          </Card>
        </motion.div>
      ))}
    </motion.div>
  );
}
