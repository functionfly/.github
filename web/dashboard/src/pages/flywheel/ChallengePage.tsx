/**
 * ChallengePage - Browse challenges
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
import { ChallengeCard } from '@/components/flywheel/challenge/ChallengeCard';
import { useChallenges } from '@/api/flywheel';
import type { ChallengeFilters } from '@/components/flywheel/types';
import {
  Trophy,
  Clock,
  CheckCircle2,
  Calendar,
} from 'lucide-react';

type TabType = 'active' | 'upcoming' | 'completed';

export default function ChallengePage() {
  const [activeTab, setActiveTab] = useState<TabType>('active');
  const [typeFilter, setTypeFilter] = useState<string>('all');

  const filters: ChallengeFilters = {
    status: activeTab === 'upcoming' ? 'upcoming' : activeTab === 'completed' ? 'completed' : 'active',
    type: typeFilter === 'all' ? undefined : typeFilter as ChallengeFilters['type'],
  };

  const { data: challengesData, isLoading } = useChallenges(filters, 20);

  const tabs = [
    { id: 'active' as TabType, label: 'Active', icon: Clock },
    { id: 'upcoming' as TabType, label: 'Upcoming', icon: Calendar },
    { id: 'completed' as TabType, label: 'Completed', icon: CheckCircle2 },
  ];

  return (
    <FlywheelPageLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white">Challenges</h1>
            <p className="text-slate-400">Compete in coding challenges and win prizes</p>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex items-center gap-2 border-b border-slate-800">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={cn(
                  'flex items-center gap-2 border-b-2 px-4 py-3 text-sm font-medium transition-colors',
                  activeTab === tab.id
                    ? 'border-indigo-500 text-indigo-400'
                    : 'border-transparent text-slate-500 hover:text-slate-300'
                )}
              >
                <Icon className="h-4 w-4" />
                {tab.label}
              </button>
            );
          })}
        </div>

        {/* Filters */}
        <div className="flex flex-wrap items-center gap-4">
          <Select value={typeFilter} onValueChange={setTypeFilter}>
            <SelectTrigger className="w-48 border-slate-800 bg-slate-900 text-slate-200">
              <SelectValue placeholder="Challenge Type" />
            </SelectTrigger>
            <SelectContent className="bg-slate-900 border-slate-800">
              <SelectItem value="all" className="text-slate-200 focus:bg-slate-800">All Types</SelectItem>
              <SelectItem value="speed" className="text-slate-200 focus:bg-slate-800">Speed</SelectItem>
              <SelectItem value="efficiency" className="text-slate-200 focus:bg-slate-800">Efficiency</SelectItem>
              <SelectItem value="accuracy" className="text-slate-200 focus:bg-slate-800">Accuracy</SelectItem>
              <SelectItem value="creative" className="text-slate-200 focus:bg-slate-800">Creative</SelectItem>
              <SelectItem value="optimization" className="text-slate-200 focus:bg-slate-800">Optimization</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Challenges Grid */}
        {isLoading ? (
          <div className="grid gap-6 md:grid-cols-2">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="h-80 animate-pulse rounded-xl bg-slate-900" />
            ))}
          </div>
        ) : challengesData?.challenges.length ? (
          <div className="grid gap-6 md:grid-cols-2">
            {challengesData.challenges.map((challenge) => (
              <ChallengeCard key={challenge.id} challenge={challenge} />
            ))}
          </div>
        ) : (
          <div className="rounded-xl border border-slate-800 bg-slate-900 py-16 text-center">
            <Trophy className="mx-auto h-16 w-16 text-slate-600" />
            <h3 className="mt-4 text-xl font-medium text-white">
              No {activeTab} challenges
            </h3>
            <p className="mt-2 text-slate-400">
              Check back later for new challenges
            </p>
          </div>
        )}
      </div>
    </FlywheelPageLayout>
  );
}
