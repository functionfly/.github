/**
 * LeaderboardPage - Global rankings
 */

import { useState } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { FlywheelPageLayout, FlywheelSection } from '@/components/flywheel/layout/FlywheelLayout';
import { LeaderboardTable, LeaderboardPodium } from '@/components/flywheel/reputation/LeaderboardTable';
import { useLeaderboard } from '@/api/flywheel';
import type { ReputationType } from '@/components/flywheel/types';
import {
  Trophy,
  BarChart3,
  Zap,
  Target,
  Users,
  Sparkles,
} from 'lucide-react';

type LeaderboardType = 'overall' | ReputationType;
type Timeframe = 'daily' | 'weekly' | 'monthly' | 'all_time';

const leaderboardTypes: { id: LeaderboardType; label: string; icon: React.ComponentType<{ className?: string }> }[] = [
  { id: 'overall', label: 'Overall', icon: Trophy },
  { id: 'builder', label: 'Builder', icon: Zap },
  { id: 'optimizer', label: 'Optimizer', icon: Target },
  { id: 'mentor', label: 'Mentor', icon: Users },
  { id: 'agent_whisperer', label: 'Agent Whisperer', icon: Sparkles },
];

const timeframes: { id: Timeframe; label: string }[] = [
  { id: 'all_time', label: 'All Time' },
  { id: 'monthly', label: 'This Month' },
  { id: 'weekly', label: 'This Week' },
  { id: 'daily', label: 'Today' },
];

export default function LeaderboardPage() {
  const [selectedType, setSelectedType] = useState<LeaderboardType>('overall');
  const [timeframe, setTimeframe] = useState<Timeframe>('all_time');

  const { data: leaderboardData, isLoading } = useLeaderboard({
    type: selectedType,
    timeframe,
    limit: 100,
  });

  return (
    <FlywheelPageLayout>
      <div className="space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-3xl font-bold text-white">Leaderboards</h1>
          <p className="mt-2 text-slate-400">
            See who's leading the Flywheel Network
          </p>
        </div>

        {/* Type Tabs */}
        <div className="flex flex-wrap justify-center gap-2">
          {leaderboardTypes.map((type) => {
            const Icon = type.icon;
            return (
              <button
                key={type.id}
                onClick={() => setSelectedType(type.id)}
                className={cn(
                  'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors',
                  selectedType === type.id
                    ? 'bg-indigo-600 text-white'
                    : 'bg-slate-900 text-slate-400 hover:bg-slate-800 hover:text-slate-200'
                )}
              >
                <Icon className="h-4 w-4" />
                {type.label}
              </button>
            );
          })}
        </div>

        {/* Timeframe Filter */}
        <div className="flex justify-center">
          <div className="inline-flex rounded-lg border border-slate-800 bg-slate-900 p-1">
            {timeframes.map((tf) => (
              <button
                key={tf.id}
                onClick={() => setTimeframe(tf.id)}
                className={cn(
                  'rounded-md px-4 py-1.5 text-sm font-medium transition-colors',
                  timeframe === tf.id
                    ? 'bg-slate-800 text-white'
                    : 'text-slate-400 hover:text-slate-200'
                )}
              >
                {tf.label}
              </button>
            ))}
          </div>
        </div>

        {/* Podium */}
        {!isLoading && leaderboardData?.leaders && leaderboardData.leaders.length >= 3 && (
          <div className="rounded-xl border border-slate-800 bg-slate-900 py-8">
            <LeaderboardPodium entries={leaderboardData.leaders.slice(0, 3)} />
          </div>
        )}

        {/* Full Rankings */}
        <FlywheelSection
          title="Full Rankings"
          action={
            <div className="flex items-center gap-2">
              <BarChart3 className="h-4 w-4 text-slate-500" />
              <span className="text-sm text-slate-400">
                {leaderboardData?.pagination.total.toLocaleString()} users
              </span>
            </div>
          }
        >
          {isLoading ? (
            <div className="h-96 animate-pulse rounded-xl bg-slate-900" />
          ) : leaderboardData?.leaders.length ? (
            <LeaderboardTable
              entries={leaderboardData.leaders}
              myRank={leaderboardData.myRank}
              showTrend={timeframe !== 'all_time'}
            />
          ) : (
            <div className="rounded-xl border border-slate-800 bg-slate-900 py-16 text-center">
              <Trophy className="mx-auto h-16 w-16 text-slate-600" />
              <h3 className="mt-4 text-xl font-medium text-white">No rankings yet</h3>
              <p className="mt-2 text-slate-400">
                Start participating to appear on the leaderboard
              </p>
            </div>
          )}
        </FlywheelSection>
      </div>
    </FlywheelPageLayout>
  );
}
