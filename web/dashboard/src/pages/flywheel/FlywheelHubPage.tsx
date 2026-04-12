/**
 * FlywheelHubPage - Main community landing page
 */

import {
  useCategories,
  useChallenges,
  useFlywheelStats,
  useLeaderboard,
  useThreads,
} from '@/api/flywheel';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';
import { ChallengeCard } from '@/components/flywheel/challenge/ChallengeCard';
import {
  FlywheelCard,
  FlywheelPageLayout,
  FlywheelSection,
} from '@/components/flywheel/layout/FlywheelLayout';
import { LeaderboardTable } from '@/components/flywheel/reputation/LeaderboardTable';
import { ThreadCard, ThreadCardSkeleton } from '@/components/flywheel/thread/ThreadCard';
import { MetaTags } from '@/components/seo/MetaTags';
import { Button } from '@/components/ui/button';
import { useWebVitals } from '@/hooks/useWebVitals';
import { cn } from '@/lib/utils';
import {
  ArrowRight,
  BookOpen,
  Code,
  Cpu,
  MessageSquare,
  Sparkles,
  Target,
  TrendingUp,
  Trophy,
  Users,
  Zap,
} from 'lucide-react';
import type React from 'react';
import { useNavigate } from 'react-router-dom';

// Stats for hero section
function AnimatedCounter({ value, suffix = '' }: { value: number; suffix?: string }) {
  return (
    <span className="text-3xl font-bold text-white tabular-nums">
      {value.toLocaleString()}
      {suffix}
    </span>
  );
}

interface HeroSectionProps {
  stats?: {
    activeThreads: number;
    verifiedSolutions: number;
    activeChallenges: number;
    totalReputationPoints: number;
  };
  isLoading?: boolean;
}

function HeroSection({ stats, isLoading }: HeroSectionProps) {
  const navigate = useNavigate();

  return (
    <section className="flywheel-hero relative overflow-hidden rounded-2xl bg-gradient-to-br from-indigo-600 via-violet-600 to-pink-600 px-6 py-12 sm:px-12 sm:py-16">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,_white_1px,_transparent_1px)] bg-[length:24px_24px]" />
      </div>

      <div className="relative z-10 mx-auto max-w-3xl text-center">
        <h1 className="flywheel-hero-title text-3xl font-bold text-white sm:text-4xl lg:text-5xl">
          Proof-of-Execution
          <br />
          Knowledge Network
        </h1>
        <p className="flywheel-hero-subtitle mt-4 text-lg text-indigo-100">
          Solve problems. Build reputation. Earn rewards.
        </p>

        <div className="mt-8 flex flex-wrap justify-center gap-4">
          <Button
            size="lg"
            onClick={() => navigate('/flywheel/threads')}
            className="bg-white text-indigo-600 hover:bg-indigo-50"
          >
            <MessageSquare className="mr-2 h-5 w-5" />
            Browse Threads
          </Button>
          <Button
            size="lg"
            variant="outline"
            onClick={() => navigate('/flywheel/challenges')}
            className="border-white/30 bg-white/10 text-white hover:bg-white/20"
          >
            <Trophy className="mr-2 h-5 w-5" />
            Start a Challenge
          </Button>
        </div>

        {/* Stats */}
        <div className="mt-10 grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div className="rounded-xl bg-white/10 p-4 backdrop-blur-sm">
            <span className="flywheel-hero-stat">
              {isLoading ? (
                <span className="text-3xl font-bold text-white/50">--</span>
              ) : (
                <AnimatedCounter value={stats?.activeThreads ?? 0} />
              )}
            </span>
            <p className="text-sm text-indigo-200">Active Threads</p>
          </div>
          <div className="rounded-xl bg-white/10 p-4 backdrop-blur-sm">
            <span className="flywheel-hero-stat">
              {isLoading ? (
                <span className="text-3xl font-bold text-white/50">--</span>
              ) : (
                <AnimatedCounter value={stats?.verifiedSolutions ?? 0} />
              )}
            </span>
            <p className="text-sm text-indigo-200">Verified Solutions</p>
          </div>
          <div className="rounded-xl bg-white/10 p-4 backdrop-blur-sm">
            <span className="flywheel-hero-stat">
              {isLoading ? (
                <span className="text-3xl font-bold text-white/50">--</span>
              ) : (
                <AnimatedCounter value={stats?.activeChallenges ?? 0} />
              )}
            </span>
            <p className="text-sm text-indigo-200">Active Challenges</p>
          </div>
          <div className="rounded-xl bg-white/10 p-4 backdrop-blur-sm">
            <span className="flywheel-hero-stat">
              {isLoading ? (
                <span className="text-3xl font-bold text-white/50">--</span>
              ) : (
                <AnimatedCounter value={stats?.totalReputationPoints ?? 0} suffix="+" />
              )}
            </span>
            <p className="text-sm text-indigo-200">Reputation Points</p>
          </div>
        </div>
      </div>
    </section>
  );
}

// Category Card
function CategoryCard({
  name,
  count,
  icon: Icon,
  color,
  onClick,
}: {
  name: string;
  count: number;
  icon: React.ComponentType<{ className?: string }>;
  color: string;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className="flywheel-category-card group flex flex-col items-center rounded-xl border border-border-default bg-bg-tertiary p-5 transition-all hover:border-border-strong hover:bg-bg-hover"
    >
      <div className={cn('rounded-lg p-3', color)}>
        <Icon className="h-6 w-6 text-white" />
      </div>
      <h3 className="flywheel-category-title mt-3 font-medium text-text-primary group-hover:text-indigo-400">
        {name}
      </h3>
      <p className="flywheel-category-count text-sm text-text-muted">{count} threads</p>
    </button>
  );
}

export default function FlywheelHubPage() {
  const navigate = useNavigate();

  // Monitor Core Web Vitals
  useWebVitals((metrics) => {
    // Optional: Send to your analytics service
    console.log('Web Vitals:', metrics);
  });

  // Fetch data
  const { data: threadsData, isLoading: threadsLoading } = useThreads({ sortBy: 'popular' }, 3);
  const { data: challengesData, isLoading: challengesLoading } = useChallenges(
    { status: 'active' },
    2
  );
  const { data: leaderboardData, isLoading: leaderboardLoading } = useLeaderboard({
    type: 'overall',
    limit: 5,
  });
  const { data: statsData, isLoading: statsLoading } = useFlywheelStats();
  const { data: categoriesData, isLoading: categoriesLoading } = useCategories();

  // Map category icons based on name
  const getCategoryIcon = (name: string) => {
    const iconMap: Record<string, React.ComponentType<{ className?: string }>> = {
      Algorithms: Zap,
      'Data Structures': Target,
      'System Design': Cpu,
      Optimization: TrendingUp,
      'Machine Learning': Sparkles,
      'Web Development': Code,
    };
    return iconMap[name] || Code;
  };

  // Map category colors based on name
  const getCategoryColor = (name: string) => {
    const colorMap: Record<string, string> = {
      Algorithms: 'bg-blue-500',
      'Data Structures': 'bg-violet-500',
      'System Design': 'bg-emerald-500',
      Optimization: 'bg-amber-500',
      'Machine Learning': 'bg-pink-500',
      'Web Development': 'bg-cyan-500',
    };
    return colorMap[name] || 'bg-slate-500';
  };

  const categories =
    categoriesData?.categories?.map((cat) => ({
      name: cat.name,
      count: cat.threadCount ?? 0,
      icon: getCategoryIcon(cat.name),
      color: cat.color || getCategoryColor(cat.name),
    })) ?? [];

  return (
    <FlywheelPageLayout>
      {/* SEO Meta Tags */}
      <MetaTags
        title="Flywheel Network - Proof-of-Execution Knowledge Network"
        description="Join the Flywheel Network - a proof-of-execution knowledge network where developers compete, collaborate, and earn rewards through algorithmic challenges and community contributions."
        keywords={[
          'flywheel network',
          'proof-of-execution',
          'algorithmic challenges',
          'developer community',
          'coding competitions',
          'reputation system',
        ]}
        url={`${window.location.origin}/flywheel`}
        type="website"
      />

      {/* Public Analytics (Hotjar for user behavior) */}
      <PublicAnalytics />

      <div className="space-y-8">
        {/* Hero */}
        <HeroSection stats={statsData?.stats} isLoading={statsLoading} />

        {/* Featured Challenges */}
        <FlywheelSection
          title="Featured Challenges"
          description="Compete for prizes and reputation"
          action={
            <Button
              variant="ghost"
              onClick={() => navigate('/flywheel/challenges')}
              className="flywheel-section-action text-indigo-400 hover:text-indigo-300"
            >
              View All
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          }
        >
          {challengesLoading ? (
            <div className="grid gap-4 md:grid-cols-2">
              {[1, 2].map((i) => (
                <div
                  key={i}
                  className="flywheel-skeleton h-64 animate-pulse rounded-xl bg-bg-tertiary"
                />
              ))}
            </div>
          ) : challengesData?.challenges?.length ? (
            <div className="grid gap-4 md:grid-cols-2">
              {challengesData.challenges.map((challenge) => (
                <ChallengeCard key={challenge.id} challenge={challenge} />
              ))}
            </div>
          ) : (
            <FlywheelCard className="text-center py-8">
              <Trophy className="flywheel-empty-icon mx-auto h-12 w-12 text-text-muted" />
              <p className="flywheel-empty-text mt-2 text-text-secondary">
                No active challenges right now
              </p>
            </FlywheelCard>
          )}
        </FlywheelSection>

        {/* Categories */}
        <FlywheelSection
          title="Browse by Category"
          description="Find threads in your area of interest"
        >
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
            {categoriesLoading ? (
              // Loading skeletons for categories
              Array.from({ length: 6 }).map((_, i) => (
                <div
                  key={i}
                  className="flywheel-skeleton h-28 animate-pulse rounded-xl bg-bg-tertiary"
                />
              ))
            ) : categories.length ? (
              categories.map((category) => (
                <CategoryCard
                  key={category.name}
                  {...category}
                  onClick={() =>
                    navigate(
                      `/flywheel/threads?category=${category.name.toLowerCase().replace(' ', '-')}`
                    )
                  }
                />
              ))
            ) : (
              <div className="col-span-full text-center py-8">
                <p className="text-text-secondary">No categories available</p>
              </div>
            )}
          </div>
        </FlywheelSection>

        {/* Two Column Layout */}
        <div className="grid gap-8 lg:grid-cols-3">
          {/* Trending Problems */}
          <div className="lg:col-span-2">
            <FlywheelSection
              title="Trending Problems"
              description="Most discussed threads this week"
              action={
                <Button
                  variant="ghost"
                  onClick={() => navigate('/flywheel/threads')}
                  className="flywheel-section-action text-indigo-400 hover:text-indigo-300"
                >
                  View All
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
              }
            >
              <div className="space-y-4">
                {threadsLoading ? (
                  <>
                    <ThreadCardSkeleton />
                    <ThreadCardSkeleton />
                    <ThreadCardSkeleton />
                  </>
                ) : threadsData?.threads?.length ? (
                  threadsData.threads.map((thread) => (
                    <ThreadCard key={thread.id} thread={thread} />
                  ))
                ) : (
                  <FlywheelCard className="text-center py-8">
                    <MessageSquare className="flywheel-empty-icon mx-auto h-12 w-12 text-text-muted" />
                    <p className="flywheel-empty-text mt-2 text-text-secondary">No threads yet</p>
                  </FlywheelCard>
                )}
              </div>
            </FlywheelSection>
          </div>

          {/* Sidebar */}
          <div className="space-y-8">
            {/* Leaderboard Preview */}
            <FlywheelSection
              title="Top Builders"
              action={
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => navigate('/flywheel/leaderboards')}
                  className="flywheel-section-action text-indigo-400 hover:text-indigo-300"
                >
                  Full Rankings
                </Button>
              }
            >
              {leaderboardLoading ? (
                <div className="flywheel-skeleton h-48 animate-pulse rounded-xl bg-bg-tertiary" />
              ) : leaderboardData?.leaders?.length ? (
                <LeaderboardTable entries={leaderboardData.leaders.slice(0, 5)} compact />
              ) : null}
            </FlywheelSection>

            {/* Getting Started */}
            <FlywheelSection title="Getting Started">
              <FlywheelCard className="space-y-3">
                <div className="flex items-start gap-3">
                  <div className="rounded-lg bg-indigo-500/10 p-2">
                    <BookOpen className="h-4 w-4 text-indigo-400" />
                  </div>
                  <div>
                    <p className="flywheel-getting-title font-medium text-text-primary">
                      Read the Guide
                    </p>
                    <p className="flywheel-getting-desc text-sm text-text-secondary">
                      Learn how to earn reputation and participate in challenges
                    </p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <div className="rounded-lg bg-emerald-500/10 p-2">
                    <Users className="h-4 w-4 text-emerald-400" />
                  </div>
                  <div>
                    <p className="flywheel-getting-title font-medium text-text-primary">
                      Join the Community
                    </p>
                    <p className="flywheel-getting-desc text-sm text-text-secondary">
                      Connect with other builders and mentors
                    </p>
                  </div>
                </div>
              </FlywheelCard>
            </FlywheelSection>
          </div>
        </div>
      </div>
    </FlywheelPageLayout>
  );
}
