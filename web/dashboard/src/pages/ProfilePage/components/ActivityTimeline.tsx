/**
 * Activity Timeline Component
 *
 * Displays user's activity feed in a timeline format.
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
  GitBranch,
  Zap,
  Code2,
  UserPlus,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { UserActivity, ActivityType } from "@/types";

export interface ActivityTimelineProps {
  activities: UserActivity[];
  filter?: ActivityType | "all";
}

export function ActivityTimeline({ activities, filter = "all" }: ActivityTimelineProps) {
  const filteredActivities = filter === "all"
    ? activities
    : activities.filter(a => a.type === filter);

  const typeIcons: Record<ActivityType, React.ReactNode> = {
    joined: <UserPlus className="w-4 h-4" />,
    function_published: <Package className="w-4 h-4" />,
    function_updated: <Edit3 className="w-4 h-4" />,
    function_deleted: <AlertCircle className="w-4 h-4" />,
    achievement_earned: <Award className="w-4 h-4" />,
    review_received: <Star className="w-4 h-4" />,
    milestone_reached: <Target className="w-4 h-4" />,
    followed: <Users className="w-4 h-4" />,
    follower_gained: <Heart className="w-4 h-4" />,
    contribution: <GitBranch className="w-4 h-4" />,
    deployment: <Zap className="w-4 h-4" />,
  };

  const typeColors: Record<ActivityType, string> = {
    joined: "bg-brand-500/20 text-brand-400",
    function_published: "bg-blue-500/20 text-blue-400",
    function_updated: "bg-yellow-500/20 text-yellow-400",
    function_deleted: "bg-red-500/20 text-red-400",
    achievement_earned: "bg-amber-500/20 text-amber-400",
    review_received: "bg-purple-500/20 text-purple-400",
    milestone_reached: "bg-emerald-500/20 text-emerald-400",
    followed: "bg-pink-500/20 text-pink-400",
    follower_gained: "bg-pink-500/20 text-pink-400",
    contribution: "bg-brand-500/20 text-brand-400",
    deployment: "bg-green-500/20 text-green-400",
  };

  return (
    <div className="space-y-4">
      {filteredActivities.map((activity, index) => (
        <motion.div
          key={activity.id}
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: index * 0.05 }}
          className="flex gap-4"
        >
          <div className="flex flex-col items-center">
            <div className={cn("w-10 h-10 rounded-full flex items-center justify-center", typeColors[activity.type])}>
              {typeIcons[activity.type]}
            </div>
            {index < filteredActivities.length - 1 && (
              <div className="w-px flex-1 bg-border-subtle my-2" />
            )}
          </div>
          <div className="flex-1 pb-6">
            <div className="flex items-start justify-between gap-2">
              <div>
                <h4 className="font-medium text-text-primary">{activity.title}</h4>
                {activity.description && (
                  <p className="text-sm text-text-muted mt-0.5">{activity.description}</p>
                )}
                {activity.relatedFunction && (
                  <Link
                    to={`/fx/${activity.relatedFunction.author}/${activity.relatedFunction.name}`}
                    className="inline-flex items-center gap-1 text-sm text-brand-400 hover:text-brand-300 mt-1"
                  >
                    <Code2 className="w-3.5 h-3.5" />
                    {activity.relatedFunction.name}
                  </Link>
                )}
              </div>
              <span className="text-xs text-text-muted shrink-0">
                {formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true })}
              </span>
            </div>
          </div>
        </motion.div>
      ))}
    </div>
  );
}
