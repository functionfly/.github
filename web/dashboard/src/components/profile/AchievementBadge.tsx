/**
 * AchievementBadge Component
 *
 * Displays individual achievement with icon, progress bar for incomplete
 * achievements, tooltip with description, rarity indicator, and unlock date.
 *
 * @example
 * <AchievementBadge
 *   achievement={{
 *     id: "1",
 *     name: "First Function",
 *     description: "Published your first function",
 *     icon: "Zap",
 *     color: "#6366f1",
 *     unlockedAt: "2024-01-15T10:30:00Z",
 *     tier: "bronze",
 *     progress: { current: 1, target: 1 }
 *   }}
 *   size="md"
 * />
 *
 * @example
 * <AchievementBadge achievement={achievement} size="lg" showProgress />
 */

import { motion } from "framer-motion";
import { useTranslation } from "react-i18next";
import { format } from "date-fns";
import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Progress } from "@/components/ui/progress";
import {
  Zap,
  Trophy,
  Star,
  Flame,
  Target,
  Rocket,
  Crown,
  Gem,
  Award,
  Medal,
  Sparkles,
  TrendingUp,
  Users,
  Code2,
  GitBranch,
  Shield,
  Lock,
  type LucideIcon,
} from "lucide-react";
import type { Achievement } from "@/types";

interface AchievementBadgeProps {
  achievement: Achievement;
  size?: "sm" | "md" | "lg";
  showProgress?: boolean;
  showRarity?: boolean;
  className?: string;
}

// Map icon names to Lucide icons
const ICON_MAP: Record<string, LucideIcon> = {
  Zap,
  Trophy,
  Star,
  Flame,
  Target,
  Rocket,
  Crown,
  Gem,
  Award,
  Medal,
  Sparkles,
  TrendingUp,
  Users,
  Code2,
  GitBranch,
  Shield,
  Lock,
};

// Rarity/tier configurations
const RARITY_CONFIG = {
  bronze: {
    label: "Common",
    gradient: "from-amber-700 via-amber-600 to-amber-500",
    bg: "bg-amber-500/10",
    border: "border-amber-500/30",
    text: "text-amber-400",
    glow: "shadow-amber-500/20",
  },
  silver: {
    label: "Rare",
    gradient: "from-slate-400 via-gray-300 to-slate-400",
    bg: "bg-slate-400/10",
    border: "border-slate-400/30",
    text: "text-slate-300",
    glow: "shadow-slate-400/20",
  },
  gold: {
    label: "Epic",
    gradient: "from-yellow-500 via-amber-400 to-yellow-500",
    bg: "bg-yellow-500/10",
    border: "border-yellow-500/30",
    text: "text-yellow-400",
    glow: "shadow-yellow-500/20",
  },
  platinum: {
    label: "Legendary",
    gradient: "from-purple-500 via-pink-500 to-purple-500",
    bg: "bg-purple-500/10",
    border: "border-purple-500/30",
    text: "text-purple-400",
    glow: "shadow-purple-500/30",
  },
};

// Size configurations
const SIZE_CONFIG = {
  sm: {
    container: "w-10 h-10",
    icon: "w-4 h-4",
    progress: "h-1",
  },
  md: {
    container: "w-14 h-14",
    icon: "w-6 h-6",
    progress: "h-1.5",
  },
  lg: {
    container: "w-20 h-20",
    icon: "w-8 h-8",
    progress: "h-2",
  },
};

export function AchievementBadge({
  achievement,
  size = "md",
  showProgress = true,
  showRarity = true,
  className,
}: AchievementBadgeProps) {
  const { t } = useTranslation();
  const Icon = ICON_MAP[achievement.icon] || Trophy;
  const rarityConfig = {
    bronze: { ...RARITY_CONFIG.bronze, label: t('achievement.common') },
    silver: { ...RARITY_CONFIG.silver, label: t('achievement.rare') },
    gold: { ...RARITY_CONFIG.gold, label: t('achievement.epic') },
    platinum: { ...RARITY_CONFIG.platinum, label: t('achievement.legendary') },
  };
  const rarity = rarityConfig[achievement.tier];
  const sizeConfig = SIZE_CONFIG[size];
  const isCompleted = !achievement.progress || achievement.progress.current >= achievement.progress.target;
  const progressPercent = achievement.progress
    ? Math.min(100, (achievement.progress.current / achievement.progress.target) * 100)
    : 100;

  const formattedUnlockDate = achievement.unlockedAt
    ? format(new Date(achievement.unlockedAt), "MMM d, yyyy")
    : null;

  const badgeContent = (
    <motion.div
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      className={cn(
        "relative rounded-xl border-2 transition-all duration-300",
        "flex flex-col items-center justify-center",
        sizeConfig.container,
        rarity.bg,
        rarity.border,
        rarity.glow,
        "shadow-lg hover:shadow-xl",
        !isCompleted && "opacity-70 grayscale-[0.3]",
        className
      )}
    >
      {/* Gradient background for unlocked achievements */}
      {isCompleted && (
        <div
          className={cn(
            "absolute inset-0 rounded-xl bg-gradient-to-br opacity-20",
            rarity.gradient
          )}
        />
      )}

      {/* Icon */}
      <div className="relative z-10">
        <Icon
          className={cn(
            sizeConfig.icon,
            rarity.text,
            !isCompleted && "text-text-muted"
          )}
        />
      </div>

      {/* Progress indicator for incomplete achievements */}
      {!isCompleted && showProgress && achievement.progress && (
        <div className="absolute -bottom-1 left-1/2 -translate-x-1/2 w-[80%]">
          <Progress
            value={progressPercent}
            className={cn(sizeConfig.progress, "bg-black/30")}
          />
        </div>
      )}

      {/* Rarity indicator dot */}
      {showRarity && isCompleted && (
        <div
          className={cn(
            "absolute -top-1 -right-1 w-3 h-3 rounded-full border-2 border-bg-primary",
            rarity.bg.replace("/10", "")
          )}
        />
      )}

      {/* Lock icon for locked achievements */}
      {!isCompleted && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/40 rounded-xl">
          <Lock className="w-4 h-4 text-text-muted" />
        </div>
      )}
    </motion.div>
  );

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>{badgeContent}</TooltipTrigger>
        <TooltipContent side="bottom" className="max-w-xs p-4">
          <div className="space-y-3">
            {/* Header */}
            <div className="flex items-start gap-3">
              <div
                className={cn(
                  "w-10 h-10 rounded-lg flex items-center justify-center shrink-0",
                  rarity.bg,
                  rarity.border,
                  "border"
                )}
              >
                <Icon className={cn("w-5 h-5", rarity.text)} />
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="font-semibold text-text-primary">{achievement.name}</h4>
                {showRarity && (
                  <span
                    className={cn(
                      "text-xs font-medium px-2 py-0.5 rounded-full",
                      rarity.bg,
                      rarity.text
                    )}
                  >
                    {rarity.label}
                  </span>
                )}
              </div>
            </div>

            {/* Description */}
            <p className="text-sm text-text-secondary">{achievement.description}</p>

            {/* Progress */}
            {achievement.progress && (
              <div className="space-y-1">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-text-muted">Progress</span>
                  <span className={rarity.text}>
                    {achievement.progress.current} / {achievement.progress.target}
                  </span>
                </div>
                <Progress value={progressPercent} className="h-2" />
              </div>
            )}

            {/* Unlock date */}
            {formattedUnlockDate && (
              <p className="text-xs text-text-muted flex items-center gap-1">
                <Sparkles className="w-3 h-3" />
                {t('achievement.unlockedOn', { date: formattedUnlockDate })}
              </p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

// Achievement grid component for displaying multiple achievements
interface AchievementGridProps {
  achievements: Achievement[];
  size?: "sm" | "md" | "lg";
  columns?: 3 | 4 | 5 | 6;
  showProgress?: boolean;
  className?: string;
}

export function AchievementGrid({
  achievements,
  size = "md",
  columns = 5,
  showProgress = true,
  className,
}: AchievementGridProps) {
  const columnClass = {
    3: "grid-cols-3",
    4: "grid-cols-4",
    5: "grid-cols-5",
    6: "grid-cols-6",
  };

  return (
    <div className={cn("grid gap-3", columnClass[columns], className)}>
      {achievements.map((achievement) => (
        <div key={achievement.id} className="flex justify-center">
          <AchievementBadge
            achievement={achievement}
            size={size}
            showProgress={showProgress}
          />
        </div>
      ))}
    </div>
  );
}

// Achievement list component for expanded view
interface AchievementListProps {
  achievements: Achievement[];
  className?: string;
}

export function AchievementList({ achievements, className }: AchievementListProps) {
  const { t } = useTranslation();

  return (
    <div className={cn("space-y-3", className)}>
      {achievements.map((achievement) => {
        const Icon = ICON_MAP[achievement.icon] || Trophy;
        const rarityConfig = {
          bronze: { ...RARITY_CONFIG.bronze, label: t('achievement.common') },
          silver: { ...RARITY_CONFIG.silver, label: t('achievement.rare') },
          gold: { ...RARITY_CONFIG.gold, label: t('achievement.epic') },
          platinum: { ...RARITY_CONFIG.platinum, label: t('achievement.legendary') },
        };
        const rarity = rarityConfig[achievement.tier];
        const isCompleted =
          !achievement.progress || achievement.progress.current >= achievement.progress.target;
        const progressPercent = achievement.progress
          ? Math.min(100, (achievement.progress.current / achievement.progress.target) * 100)
          : 100;

        return (
          <div
            key={achievement.id}
            className={cn(
              "flex items-center gap-4 p-3 rounded-lg border transition-all",
              "hover:bg-bg-tertiary",
              rarity.bg,
              rarity.border,
              !isCompleted && "opacity-60"
            )}
          >
            {/* Icon */}
            <div
              className={cn(
                "w-12 h-12 rounded-xl flex items-center justify-center shrink-0",
                rarity.bg,
                "border border-white/10"
              )}
            >
              <Icon className={cn("w-6 h-6", rarity.text)} />
            </div>

            {/* Content */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <h4 className="font-semibold text-text-primary">{achievement.name}</h4>
                <span
                  className={cn(
                    "text-xs font-medium px-2 py-0.5 rounded-full",
                    rarity.bg,
                    rarity.text
                  )}
                >
                  {rarity.label}
                </span>
              </div>
              <p className="text-sm text-text-secondary line-clamp-1">
                {achievement.description}
              </p>
              {achievement.progress && !isCompleted && (
                <div className="mt-2 space-y-1">
                  <div className="flex items-center justify-between text-xs">
                  <span className="text-text-muted">{t('achievement.progress')}</span>
                    <span className={rarity.text}>
                      {achievement.progress.current} / {achievement.progress.target}
                    </span>
                  </div>
                  <Progress value={progressPercent} className="h-1.5" />
                </div>
              )}
            </div>

            {/* Unlock date */}
            {achievement.unlockedAt && (
              <div className="text-right shrink-0">
                <p className="text-xs text-text-muted">
                  {format(new Date(achievement.unlockedAt), "MMM d, yyyy")}
                </p>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

export default AchievementBadge;
