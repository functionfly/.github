/**
 * ChallengeCard - Challenge preview card
 */

import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import {
  Clock,
  Users,
  Trophy,
  Zap,
  Target,
  Sparkles,
  Gauge,
  Palette,
} from 'lucide-react';
import type { Challenge, ChallengeType, ChallengeStatus } from '../types';

interface ChallengeCardProps {
  challenge: Challenge;
  className?: string;
}

const typeConfig: Record<ChallengeType, {
  icon: React.ComponentType<{ className?: string }>;
  gradient: string;
  label: string;
}> = {
  speed: {
    icon: Zap,
    gradient: 'from-red-500 to-orange-500',
    label: 'Speed',
  },
  efficiency: {
    icon: Gauge,
    gradient: 'from-emerald-500 to-teal-500',
    label: 'Efficiency',
  },
  accuracy: {
    icon: Target,
    gradient: 'from-blue-500 to-indigo-500',
    label: 'Accuracy',
  },
  creative: {
    icon: Palette,
    gradient: 'from-amber-500 to-yellow-500',
    label: 'Creative',
  },
  optimization: {
    icon: Sparkles,
    gradient: 'from-violet-500 to-purple-500',
    label: 'Optimization',
  },
};

const statusConfig: Record<ChallengeStatus, string> = {
  upcoming: 'Upcoming',
  active: 'Active',
  judging: 'Judging',
  completed: 'Completed',
  cancelled: 'Cancelled',
};

function CountdownTimer({ endTime }: { endTime: string }) {
  const [timeLeft, setTimeLeft] = useState(getTimeLeft(endTime));

  useEffect(() => {
    const timer = setInterval(() => {
      setTimeLeft(getTimeLeft(endTime));
    }, 1000);

    return () => clearInterval(timer);
  }, [endTime]);

  function getTimeLeft(end: string) {
    const endDate = new Date(end);
    const now = new Date();
    const diff = endDate.getTime() - now.getTime();

    if (diff <= 0) return { days: 0, hours: 0, minutes: 0, seconds: 0 };

    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((diff % (1000 * 60)) / 1000);

    return { days, hours, minutes, seconds };
  }

  const isUrgent = timeLeft.days === 0 && timeLeft.hours < 24;
  const isCritical = timeLeft.days === 0 && timeLeft.hours === 0 && timeLeft.minutes < 10;

  return (
    <div className={cn(
      'flex items-center gap-1 font-mono text-lg font-bold',
      isCritical ? 'text-red-400 animate-pulse' : isUrgent ? 'text-amber-400' : 'text-white'
    )}>
      {timeLeft.days > 0 && (
        <>
          <TimeUnit value={timeLeft.days} label="d" />
          <span>:</span>
        </>
      )}
      <TimeUnit value={timeLeft.hours} label="h" />
      <span>:</span>
      <TimeUnit value={timeLeft.minutes} label="m" />
      {timeLeft.days === 0 && (
        <>
          <span>:</span>
          <TimeUnit value={timeLeft.seconds} label="s" />
        </>
      )}
    </div>
  );
}

function TimeUnit({ value, label }: { value: number; label: string }) {
  return (
    <span className="flex items-baseline gap-0.5">
      <span className="tabular-nums">{String(value).padStart(2, '0')}</span>
      <span className="text-xs font-normal text-text-muted">{label}</span>
    </span>
  );
}

export function ChallengeCard({ challenge, className }: ChallengeCardProps) {
  const type = typeConfig[challenge.challengeType];
  const TypeIcon = type.icon;

  const isActive = challenge.status === 'active';
  const isUpcoming = challenge.status === 'upcoming';
  const isCompleted = challenge.status === 'completed';

  // Calculate progress for active challenges
  const startTime = new Date(challenge.schedule.startTime).getTime();
  const endTime = new Date(challenge.schedule.endTime).getTime();
  const now = Date.now();
  const progress = isActive
    ? Math.min(100, Math.max(0, ((now - startTime) / (endTime - startTime)) * 100))
    : 0;

  return (
    <div
      className={cn(
        'flywheel-card relative overflow-hidden rounded-xl border border-border-default bg-bg-tertiary transition-all duration-300 hover:border-border-strong hover:shadow-lg',
        className
      )}
    >
      {/* Type Gradient Bar */}
      <div className={cn('h-1.5 bg-gradient-to-r', type.gradient)} />

      <div className="p-5">
        {/* Header */}
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className={cn(
              'flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br',
              type.gradient
            )}>
              <TypeIcon className="h-5 w-5 text-white" />
            </div>
            <div>
              <Badge variant="outline" className={cn(
                'mb-1 border-border-default text-xs',
                challenge.status === 'active' ? 'bg-emerald-500/10 text-emerald-400' : 'text-text-muted'
              )}>
                {statusConfig[challenge.status]}
              </Badge>
              <p className={cn('text-xs font-medium', 'text-text-muted')}>
                {type.label} Challenge
              </p>
            </div>
          </div>

          {/* Prize Pool */}
          <div className="text-right">
            <p className="text-xs text-text-muted">Prize Pool</p>
            <p className="text-lg font-bold text-white">
              ${challenge.rewards.totalPool.toLocaleString()}
            </p>
          </div>
        </div>

        {/* Title & Description */}
        <h3 className="mt-3 text-lg font-semibold text-white line-clamp-1">
          {challenge.title}
        </h3>
        <p className="mt-1 line-clamp-2 text-sm text-text-secondary">
          {challenge.description}
        </p>

        {/* Countdown Timer */}
        {(isActive || isUpcoming) && (
          <div className="mt-4">
            <p className="mb-2 text-xs text-text-muted">
              {isActive ? 'Ends in:' : 'Starts in:'}
            </p>
            <CountdownTimer endTime={challenge.schedule.endTime} />

            {isActive && (
              <div className="mt-3">
                <Progress value={progress} className="h-1 bg-bg-hover" />
                <p className="mt-1 text-xs text-text-muted">
                  {progress.toFixed(0)}% complete
                </p>
              </div>
            )}
          </div>
        )}

        {/* Prize Breakdown */}
        <div className="mt-4 flex items-center gap-3">
          {challenge.rewards.breakdown.slice(0, 3).map((prize, index) => (
            <div key={prize.rank} className="flex items-center gap-1">
              <span className="text-lg">
                {index === 0 ? '🥇' : index === 1 ? '🥈' : '🥉'}
              </span>
              <span className="text-sm font-medium text-text-secondary">
                ${prize.amount.toLocaleString()}
              </span>
            </div>
          ))}
        </div>

        {/* Stats & Actions */}
        <div className="mt-4 flex items-center justify-between border-t border-border-default pt-4">
          <div className="flex items-center gap-4 text-sm text-text-muted">
            <span className="flex items-center gap-1">
              <Users className="h-4 w-4" />
              {challenge.statistics.participantCount}
            </span>
            <span className="flex items-center gap-1">
              <Trophy className="h-4 w-4" />
              {challenge.statistics.submissionCount}
            </span>
          </div>

          <div className="flex items-center gap-2">
            {challenge.mySubmission && (
              <Badge variant="outline" className={cn(
                'text-xs',
                challenge.mySubmission.status === 'scored'
                  ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                  : 'bg-amber-500/10 text-amber-400 border-amber-500/30'
              )}>
                {challenge.mySubmission.status === 'scored'
                  ? `#${challenge.mySubmission.rank}`
                  : 'Entered'}
              </Badge>
            )}

            <Link to={`/flywheel/challenges/${challenge.id}`}>
              <Button
                size="sm"
                className={cn(
                  'bg-gradient-to-r text-white',
                  type.gradient
                )}
              >
                {isActive ? 'Enter' : isUpcoming ? 'Remind Me' : 'View Results'}
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}

/**
 * Compact challenge card for sidebars
 */
export function ChallengeCardCompact({ challenge }: { challenge: Challenge }) {
  const type = typeConfig[challenge.challengeType];
  const TypeIcon = type.icon;

  return (
    <Link
      to={`/flywheel/challenges/${challenge.id}`}
      className="flex items-center gap-3 rounded-lg border border-border-default bg-bg-tertiary p-3 transition-colors hover:border-border-strong"
    >
      <div className={cn(
        'flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br',
        type.gradient
      )}>
        <TypeIcon className="h-5 w-5 text-white" />
      </div>
      <div className="min-w-0 flex-1">
        <h4 className="truncate text-sm font-medium text-white">
          {challenge.title}
        </h4>
        <p className="text-xs text-text-muted">
          ${challenge.rewards.totalPool.toLocaleString()} • {challenge.statistics.participantCount} participants
        </p>
      </div>
      <Badge variant="outline" className="shrink-0 border-border-default text-xs text-text-muted">
        {statusConfig[challenge.status]}
      </Badge>
    </Link>
  );
}
