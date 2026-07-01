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
    bgStyle: React.CSSProperties;
    iconStyle: React.CSSProperties;
  }
> = {
  joined: {
    icon: <UserPlus className="w-4 h-4" />,
    bgStyle: { background: 'rgba(143, 255, 208, 0.08)' },
    iconStyle: { color: 'var(--status-ok)' },
  },
  function_published: {
    icon: <Package className="w-4 h-4" />,
    bgStyle: { background: 'rgba(159, 216, 255, 0.08)' },
    iconStyle: { color: 'var(--foil-a)' },
  },
  function_updated: {
    icon: <Edit3 className="w-4 h-4" />,
    bgStyle: { background: 'rgba(232, 196, 104, 0.08)' },
    iconStyle: { color: 'var(--status-pending)' },
  },
  function_deleted: {
    icon: <AlertCircle className="w-4 h-4" />,
    bgStyle: { background: 'rgba(255, 107, 107, 0.08)' },
    iconStyle: { color: 'var(--status-revoked)' },
  },
  achievement_earned: {
    icon: <Award className="w-4 h-4" />,
    bgStyle: { background: 'rgba(232, 196, 104, 0.08)' },
    iconStyle: { color: 'var(--status-pending)' },
  },
  review_received: {
    icon: <Star className="w-4 h-4" />,
    bgStyle: { background: 'rgba(217, 196, 255, 0.08)' },
    iconStyle: { color: 'var(--foil-b)' },
  },
  milestone_reached: {
    icon: <Target className="w-4 h-4" />,
    bgStyle: { background: 'rgba(143, 255, 208, 0.08)' },
    iconStyle: { color: 'var(--status-ok)' },
  },
  followed: {
    icon: <Users className="w-4 h-4" />,
    bgStyle: { background: 'rgba(255, 217, 240, 0.08)' },
    iconStyle: { color: 'var(--foil-d)' },
  },
  follower_gained: {
    icon: <Heart className="w-4 h-4" />,
    bgStyle: { background: 'rgba(255, 217, 240, 0.08)' },
    iconStyle: { color: 'var(--foil-d)' },
  },
  contribution: {
    icon: <GitCommit className="w-4 h-4" />,
    bgStyle: { background: 'rgba(143, 255, 208, 0.08)' },
    iconStyle: { color: 'var(--status-ok)' },
  },
  deployment: {
    icon: <Rocket className="w-4 h-4" />,
    bgStyle: { background: 'rgba(143, 255, 208, 0.08)' },
    iconStyle: { color: 'var(--status-ok)' },
  },
  membership_upgraded: {
    icon: <Crown className="w-4 h-4" />,
    bgStyle: { background: 'rgba(232, 196, 104, 0.1)' },
    iconStyle: { color: 'var(--status-pending)' },
  },
};

const PLAN_STYLES: Record<string, React.CSSProperties> = {
  free: { background: 'rgba(74, 86, 95, 0.15)', color: 'var(--text-faint)', borderColor: 'rgba(74, 86, 95, 0.3)' },
  starter: { background: 'rgba(159, 216, 255, 0.1)', color: 'var(--foil-a)', borderColor: 'rgba(159, 216, 255, 0.3)' },
  professional: { background: 'rgba(217, 196, 255, 0.1)', color: 'var(--foil-b)', borderColor: 'rgba(217, 196, 255, 0.3)' },
  enterprise: { background: 'rgba(232, 196, 104, 0.1)', color: 'var(--status-pending)', borderColor: 'rgba(232, 196, 104, 0.3)' },
};

function PlanBadge({ plan }: { plan?: string }) {
  if (!plan) return null;

  return (
    <span
      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium border"
      style={PLAN_STYLES[plan.toLowerCase()] || PLAN_STYLES.free}
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
        "group relative rounded-[var(--radius-lg)] transition-all duration-200",
        compact ? "p-3" : "p-4"
      )}
      style={{
        border: '1px solid var(--panel-edge)',
        background: 'var(--panel-raised)',
      }}
    >
      {/* Accent line */}
      <div
        className="absolute left-0 top-3 bottom-3 w-0.5 rounded-full"
        style={{ ...config.bgStyle, opacity: 0.5 }}
      />

      <div className="flex gap-3 pl-2">
        {/* Icon */}
        <div
          className="flex-shrink-0 w-9 h-9 rounded-[var(--radius-lg)] flex items-center justify-center transition-transform duration-200 group-hover:scale-110"
          style={{ ...config.bgStyle, ...config.iconStyle }}
        >
          {config.icon}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2">
            <h4
              className={cn(
                "font-medium leading-snug",
                compact ? "text-sm" : "text-[15px]"
              )}
              style={{ color: 'var(--text)' }}
            >
              {activity.title}
            </h4>
            <span className="text-xs shrink-0 font-mono tabular-nums mt-0.5" style={{ color: 'var(--text-faint)' }}>
              {activity.timestamp ? formatDistanceToNow(new Date(activity.timestamp), {
                addSuffix: true,
              }) : ''}
            </span>
          </div>

          {activity.description && !compact && (
            <p className="text-sm mt-1 leading-relaxed" style={{ color: 'var(--text-faint)' }}>
              {activity.description}
            </p>
          )}

          {/* Type-specific inline context */}
          {activity.relatedFunction?.author && activity.relatedFunction?.name && (
            <Link
              to={`/fx/${activity.relatedFunction.author}/${activity.relatedFunction.name}`}
              className="inline-flex items-center gap-1.5 mt-2 px-2.5 py-1 rounded-[var(--radius)] text-xs font-medium transition-colors"
              style={{ background: 'var(--panel)', color: 'var(--text-dim)', border: '1px solid var(--panel-edge)' }}
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
                  <span className="text-xs" style={{ color: 'var(--text-faint)' }}>
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
                    className="w-4 h-4 rounded-full border"
                    style={{ background: 'rgba(232, 196, 104, 0.15)', borderColor: 'rgba(232, 196, 104, 0.3)', zIndex: 3 - i }}
                  />
                ))}
              </div>
              <span className="text-xs" style={{ color: 'var(--status-pending)', opacity: 0.8 }}>Achievement unlocked</span>
            </div>
          )}
        </div>
      </div>
    </motion.div>
  );
}
