/**
 * Contribution Streak Component
 *
 * Displays the user's contribution streak with animated fire particles
 * and progress toward the next milestone.
 */

import { useMemo } from "react";
import { motion } from "framer-motion";
import { Flame, Zap, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ContributionStreakProps {
  current: number;
  longest: number;
  lastContribution: string;
}

const MILESTONES = [7, 14, 30, 60, 90, 180, 365];

function getNextMilestone(current: number): number {
  return MILESTONES.find((m) => m > current) ?? current + 100;
}

function getMilestoneProgress(current: number): number {
  const next = getNextMilestone(current);
  const prev = MILESTONES[MILESTONES.findIndex((m) => m === next) - 1] ?? 0;
  if (next === prev) return 100;
  return Math.min(100, ((current - prev) / (next - prev)) * 100);
}

interface ParticleProps {
  delay: number;
  x: number;
}

function FireParticle({ delay, x }: ParticleProps) {
  return (
    <motion.div
      className="absolute w-1 h-1 rounded-full bg-amber-400 ca-fire-particle"
      style={{ left: `${x}%`, bottom: "80%" }}
      initial={{ opacity: 0, y: 0, scale: 1 }}
      animate={{
        opacity: [0, 1, 0],
        y: [-5, -20, -35],
        scale: [1, 0.8, 0.3],
      }}
      transition={{
        duration: 1.5,
        delay,
        repeat: Infinity,
        repeatDelay: Math.random() * 2,
      }}
    />
  );
}

export function ContributionStreak({
  current,
  longest,
}: ContributionStreakProps) {
  const nextMilestone = useMemo(() => getNextMilestone(current), [current]);
  const progress = useMemo(() => getMilestoneProgress(current), [current]);
  const intensity = Math.min(4, Math.floor(current / 10));

  const particles = useMemo(
    () =>
      Array.from({ length: Math.min(current, 12) }, (_, i) => ({
        delay: i * 0.3 + Math.random() * 0.5,
        x: 20 + Math.random() * 60,
      })),
    [current]
  );

  return (
    <div className="relative overflow-hidden rounded-xl border border-border-subtle bg-gradient-to-br from-surface-primary via-surface-primary to-amber-500/5 p-4 ca-streak-card">
      {/* Background glow */}
      <div
        className="absolute -top-8 -right-8 w-32 h-32 rounded-full blur-3xl pointer-events-none ca-streak-glow"
        style={{
          opacity: current > 0 ? 0.15 + intensity * 0.05 : 0,
          background: `radial-gradient(circle, rgb(245 158 11) 0%, transparent 70%)`,
        }}
      />

      <div className="relative flex items-center gap-4">
        {/* Fire icon with particles */}
        <div className="relative flex-shrink-0">
          <motion.div
            className={cn(
              "w-14 h-14 rounded-2xl flex items-center justify-center relative",
              current > 0
                ? "bg-gradient-to-br from-amber-500/20 to-red-500/20 border border-amber-500/30 ca-streak-icon-active"
                : "bg-surface-secondary border border-border-subtle ca-streak-icon-idle"
            )}
            animate={
              current > 0
                ? { scale: [1, 1.05, 1] }
                : {}
            }
            transition={{ duration: 2, repeat: Infinity }}
          >
            <Flame
              className={cn(
                "w-7 h-7 transition-colors",
                current > 0 ? "text-amber-400" : "text-text-muted"
              )}
            />
          </motion.div>

          {/* Fire particles */}
          {current > 0 &&
            particles.map((p, i) => <FireParticle key={i} {...p} />)}
        </div>

        {/* Streak info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-baseline gap-2">
            <motion.span
              className="text-3xl font-bold font-mono text-text-primary tabular-nums ca-streak-count"
              key={current}
              initial={{ scale: 1.3, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ type: "spring", stiffness: 300 }}
            >
              {current}
            </motion.span>
            <span className="text-sm text-text-muted ca-streak-label">day streak</span>
          </div>

          {/* Progress bar */}
          <div className="mt-2">
            <div className="flex items-center justify-between text-xs mb-1">
              <span className="text-text-muted flex items-center gap-1 ca-streak-milestone-text">
                <Zap className="w-3 h-3 text-amber-400" />
                Next: <span className="font-medium text-text-secondary">{nextMilestone}</span> days
              </span>
              <span className="text-text-muted font-mono tabular-nums">
                {Math.round(progress)}%
              </span>
            </div>
            <div className="h-1.5 bg-surface-secondary rounded-full overflow-hidden ca-streak-progress-track">
              <motion.div
                className="h-full rounded-full bg-gradient-to-r from-amber-500 to-orange-500 ca-streak-progress-fill"
                initial={{ width: 0 }}
                animate={{ width: `${progress}%` }}
                transition={{ duration: 0.8, ease: "easeOut" }}
              />
            </div>
          </div>

          {/* Longest streak badge */}
          {longest > current && (
            <div className="mt-2 flex items-center gap-1 text-xs text-text-muted ca-streak-best">
              <ChevronRight className="w-3 h-3" />
              <span>
                Best: <span className="font-medium text-text-secondary">{longest}</span> days
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
