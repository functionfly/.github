/**
 * Activity Card Component
 *
 * Rich activity card with type-specific inline context.
 * Renders different layouts based on activity type (function link,
 * achievement badge, plan upgrade, etc.).
 */

import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { formatDistanceToNow } from "date-fns";
import {
  Package,
  Edit3,
  AlertCircle,
  Award,
  Star,
  Target,
  Users,
  Heart,
  GitCommit,
  Rocket,
  Crown,
  Sparkles,
  UserPlus,
  Code2,
  ExternalLink,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { UserActivity, ActivityType } from "@/types";

const TYPE_CONFIG: Record<
  ActivityType,
  {
    icon: React.ReactNode;
    bgClass: string;
    iconClass: string;
  }
> = {
  joined: {
    icon: <UserPlus className="w-4 h-4" />,
    bgClass: "bg-brand-500/10",
    iconClass: "text-brand-400 ca-icon-joined",
  },
  function_published: {
    icon: <Package className="w-4 h-4" />,
    bgClass: "bg-blue-500/10",
    iconClass: "text-blue-400 ca-icon-function_published",
  },
  function_updated: {
    icon: <Edit3 className="w-4 h-4" />,
    bgClass: "bg-yellow-500/10",
    iconClass: "text-yellow-400 ca-icon-function_updated",
  },
  function_deleted: {
    icon: <AlertCircle className="w-4 h-4" />,
    bgClass: "bg-red-500/10",
    iconClass: "text-red-400 ca-icon-function_deleted",
  },
  achievement_earned: {
    icon: <Award className="w-4 h-4" />,
    bgClass: "bg-amber-500/10",
    iconClass: "text-amber-400 ca-icon-achievement_earned",
  },
  review_received: {
    icon: <Star className="w-4 h-4" />,
    bgClass: "bg-purple-500/10",
    iconClass: "text-purple-400 ca-icon-review_received",
  },
  milestone_reached: {
    icon: <Target className="w-4 h-4" />,
    bgClass: "bg-emerald-500/10",
    iconClass: "text-emerald-400 ca-icon-milestone_reached",
  },
  followed: {
    icon: <Users className="w-4 h-4" />,
    bgClass: "bg-pink-500/10",
    iconClass: "text-pink-400 ca-icon-followed",
  },
  follower_gained: {
    icon: <Heart className="w-4 h-4" />,
    bgClass: "bg-pink-500/10",
    iconClass: "text-pink-400 ca-icon-follower_gained",
  },
  contribution: {
    icon: <GitCommit className="w-4 h-4" />,
    bgClass: "bg-brand-500/10",
    iconClass: "text-brand-400 ca-icon-contribution",
  },
  deployment: {
    icon: <Rocket className="w-4 h-4" />,
    bgClass: "bg-green-500/10",
    iconClass: "text-green-400 ca-icon-deployment",
  },
  membership_upgraded: {
    icon: <Crown className="w-4 h-4" />,
    bgClass: "bg-amber-500/15",
    iconClass: "text-amber-300 ca-icon-membership_upgraded",
  },
};

const PLAN_STYLES: Record<string, string> = {
  free: "bg-slate-500/20 text-slate-400 border-slate-500/30 ca-plan-free",
  starter: "bg-blue-500/20 text-blue-400 border-blue-500/30 ca-plan-starter",
  professional: "bg-purple-500/20 text-purple-400 border-purple-500/30 ca-plan-professional",
  enterprise: "bg-amber-500/20 text-amber-400 border-amber-500/30 ca-plan-enterprise",
};

function PlanBadge({ plan }: { plan?: string }) {
  if (!plan) return null;

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium border",
        PLAN_STYLES[plan.toLowerCase()] || PLAN_STYLES.free
      )}
    >
      <Sparkles className="w-3 h-3" />
      {plan.charAt(0).toUpperCase() + plan.slice(1)}
    </span>
  );
}

export interface ActivityCardProps {
  activity: UserActivity;
  index?: number;
  compact?: boolean;
}

export function ActivityCard({
  activity,
  index = 0,
  compact = false,
}: ActivityCardProps) {
  const config = TYPE_CONFIG[activity.type];

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.04, duration: 0.3 }}
      className={cn(
        "group relative rounded-xl border transition-all duration-200",
        "border-border-subtle",
        "bg-surface-primary hover:bg-surface-secondary/50",
        "hover:shadow-lg",
        `ca-activity-card ca-accent-${activity.type}`,
        compact ? "p-3" : "p-4"
      )}
    >
      {/* Accent line */}
      <div
        className={cn(
          "absolute left-0 top-3 bottom-3 w-0.5 rounded-full",
          config.bgClass.replace("/10", "/40").replace("/15", "/40"),
          `ca-accent-${activity.type}`
        )}
      />

      <div className="flex gap-3 pl-2">
        {/* Icon */}
        <div
          className={cn(
            "flex-shrink-0 w-9 h-9 rounded-xl flex items-center justify-center transition-transform duration-200 group-hover:scale-110",
            config.bgClass,
            config.iconClass,
            `ca-icon-${activity.type}`
          )}
        >
          {config.icon}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2">
            <h4
              className={cn(
                "font-medium text-text-primary leading-snug ca-activity-title",
                compact ? "text-sm" : "text-[15px]"
              )}
            >
              {activity.title}
            </h4>
            <span className="text-xs text-text-muted shrink-0 font-mono tabular-nums mt-0.5 ca-activity-time">
              {formatDistanceToNow(new Date(activity.timestamp), {
                addSuffix: true,
              })}
            </span>
          </div>

          {activity.description && !compact && (
            <p className="text-sm text-text-muted mt-1 leading-relaxed ca-activity-desc">
              {activity.description}
            </p>
          )}

          {/* Type-specific inline context */}
          {activity.relatedFunction && (
            <Link
              to={`/fx/${activity.relatedFunction.author}/${activity.relatedFunction.name}`}
              className={cn(
                "inline-flex items-center gap-1.5 mt-2 px-2.5 py-1 rounded-lg text-xs font-medium transition-colors",
                "bg-surface-secondary hover:bg-surface-tertiary",
                "text-text-secondary hover:text-text-primary",
                "border border-border-subtle",
                "ca-function-link"
              )}
            >
              <Code2 className="w-3 h-3" />
              {activity.relatedFunction.name}
              <ExternalLink className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
            </Link>
          )}

          {activity.type === "membership_upgraded" &&
            activity.metadata?.plan && (
              <div className="flex items-center gap-2 mt-2">
                <PlanBadge plan={activity.metadata.plan as string} />
                {activity.metadata.previousPlan && (
                  <span className="text-xs text-text-muted ca-activity-desc">
                    from{" "}
                    {String(activity.metadata.previousPlan).charAt(0).toUpperCase() +
                      String(activity.metadata.previousPlan).slice(1)}
                  </span>
                )}
              </div>
            )}

          {activity.type === "achievement_earned" && (
            <div className="mt-2 flex items-center gap-1.5">
              <div className="flex -space-x-1">
                {[...Array(3)].map((_, i) => (
                  <div
                    key={i}
                    className="w-4 h-4 rounded-full bg-amber-500/20 border border-amber-500/30 ca-achievement-dots"
                    style={{ zIndex: 3 - i }}
                  />
                ))}
              </div>
              <span className="text-xs text-amber-400/80 ca-achievement-text">Achievement unlocked</span>
            </div>
          )}
        </div>
      </div>
    </motion.div>
  );
}
