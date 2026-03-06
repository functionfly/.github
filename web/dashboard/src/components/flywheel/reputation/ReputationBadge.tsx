/**
 * ReputationBadge - Visual indicator of user's reputation score and tier
 */

import { useEffect, useState } from 'react';
import { cn } from '@/lib/utils';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import type { ReputationType } from '../types';

interface ReputationBadgeProps {
  score: number;
  type: ReputationType | 'overall';
  tier?: number;
  showScore?: boolean;
  showTier?: boolean;
  size?: 'xs' | 'sm' | 'md' | 'lg';
  animated?: boolean;
}

const typeColors: Record<string, { bg: string; text: string; ring: string }> = {
  builder: { bg: 'bg-blue-500/10', text: 'text-blue-400', ring: 'stroke-blue-500' },
  optimizer: { bg: 'bg-violet-500/10', text: 'text-violet-400', ring: 'stroke-violet-500' },
  mentor: { bg: 'bg-emerald-500/10', text: 'text-emerald-400', ring: 'stroke-emerald-500' },
  agent_whisperer: { bg: 'bg-amber-500/10', text: 'text-amber-400', ring: 'stroke-amber-500' },
  overall: { bg: 'bg-pink-500/10', text: 'text-pink-400', ring: 'stroke-pink-500' },
};

const typeLabels: Record<string, string> = {
  builder: 'Builder',
  optimizer: 'Optimizer',
  mentor: 'Mentor',
  agent_whisperer: 'Agent Whisperer',
  overall: 'Overall',
};

const tierNames: Record<number, string> = {
  1: 'Novice',
  2: 'Apprentice',
  3: 'Expert',
  4: 'Master',
  5: 'Legend',
};

export function ReputationBadge({
  score,
  type,
  tier = 1,
  showScore = false,
  showTier = false,
  size = 'sm',
  animated = true,
}: ReputationBadgeProps) {
  const [displayScore, setDisplayScore] = useState(0);
  const colors = typeColors[type] || typeColors.overall;

  const sizeConfig = {
    xs: { ring: 16, stroke: 2, font: 'text-[10px]', badge: 'w-4 h-4' },
    sm: { ring: 24, stroke: 2.5, font: 'text-xs', badge: 'w-6 h-6' },
    md: { ring: 32, stroke: 3, font: 'text-sm', badge: 'w-8 h-8' },
    lg: { ring: 48, stroke: 4, font: 'text-base', badge: 'w-12 h-12' },
  };

  const config = sizeConfig[size];
  const radius = (config.ring - config.stroke) / 2;
  const circumference = 2 * Math.PI * radius;
  const progress = Math.min(score / 10000, 1);
  const strokeDashoffset = circumference - progress * circumference;

  useEffect(() => {
    if (!animated) {
      setDisplayScore(score);
      return;
    }

    let startTime: number;
    const duration = 1500;

    const animate = (timestamp: number) => {
      if (!startTime) startTime = timestamp;
      const elapsed = timestamp - startTime;
      const progress = Math.min(elapsed / duration, 1);

      // Ease out cubic
      const easeOut = 1 - Math.pow(1 - progress, 3);
      setDisplayScore(Math.floor(easeOut * score));

      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }, [score, animated]);

  const formatScore = (num: number): string => {
    if (num >= 1000) {
      return `${(num / 1000).toFixed(1)}K`;
    }
    return String(num);
  };

  const badgeContent = (
    <div
      className={cn(
        'relative flex items-center justify-center rounded-full',
        config.badge,
        colors.bg
      )}
    >
      {/* Progress Ring */}
      <svg
        className="absolute inset-0 -rotate-90"
        width={config.ring}
        height={config.ring}
      >
        {/* Background ring */}
        <circle
          cx={config.ring / 2}
          cy={config.ring / 2}
          r={radius}
          fill="none"
          className="stroke-slate-800"
          strokeWidth={config.stroke}
        />
        {/* Progress ring */}
        <circle
          cx={config.ring / 2}
          cy={config.ring / 2}
          r={radius}
          fill="none"
          className={cn(colors.ring, animated && 'transition-all duration-1000 ease-out')}
          strokeWidth={config.stroke}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={animated ? undefined : strokeDashoffset}
          style={animated ? { strokeDashoffset } : undefined}
        />
      </svg>

      {/* Score text */}
      <span className={cn('relative z-10 font-semibold', config.font, colors.text)}>
        {formatScore(displayScore)}
      </span>
    </div>
  );

  const tooltipContent = (
    <div className="space-y-1">
      <p className="font-medium text-slate-200">{typeLabels[type]} Score</p>
      <p className="text-2xl font-bold text-white">{score.toLocaleString()}</p>
      <p className="text-sm text-slate-400">
        Tier: {tierNames[tier]} ({tier}/5)
      </p>
      <p className="text-xs text-slate-400">
        Next tier: {(tier * 1000 + 1000).toLocaleString()} points
      </p>
    </div>
  );

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="flex items-center gap-2">
            {badgeContent}
            {(showScore || showTier) && (
              <div className="flex flex-col">
                {showScore && (
                  <span className={cn('font-medium', colors.text)}>
                    {score.toLocaleString()}
                  </span>
                )}
                {showTier && (
                  <span className="text-xs text-slate-400">
                    {tierNames[tier]}
                  </span>
                )}
              </div>
            )}
          </div>
        </TooltipTrigger>
        <TooltipContent
          side="bottom"
          className="bg-slate-900 border-slate-800 p-3"
        >
          {tooltipContent}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Compact reputation badge for inline display
 */
export function ReputationBadgeInline({
  score,
  type,
  tier,
  size = 'xs',
}: {
  score: number;
  type: ReputationType | 'overall';
  tier?: number;
  size?: 'xs' | 'sm';
}) {
  const colors = typeColors[type] || typeColors.overall;

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium',
        colors.bg,
        colors.text,
        size === 'sm' && 'px-2.5 py-1'
      )}
    >
      <span>⭐</span>
      {score.toLocaleString()}
      {tier && tier > 0 && (
        <span className="opacity-70">({tierNames[tier]})</span>
      )}
    </span>
  );
}
