/**
 * FlywheelHubPage - Main community landing page
 */

import { useNavigate } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { MetaTags } from '@/components/seo/MetaTags';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';
import { FlywheelPageLayout, FlywheelSection, FlywheelCard } from '@/components/flywheel/layout/FlywheelLayout';
import { ThreadCard, ThreadCardSkeleton } from '@/components/flywheel/thread/ThreadCard';
import { ChallengeCard, ChallengeCardCompact } from '@/components/flywheel/challenge/ChallengeCard';
import { LeaderboardTable } from '@/components/flywheel/reputation/LeaderboardTable';
import { useThreads, useChallenges, useLeaderboard } from '@/api/flywheel';
import {
  MessageSquare,
  Trophy,
  TrendingUp,
  Users,
  ArrowRight,
  Sparkles,
  Zap,
  Target,
  Code,
  BookOpen,
  Cpu,
} from 'lucide-react';

// Stats for hero section
function AnimatedCounter({ value, suffix = '' }: { value: number; suffix?: string }) {
  return (
    <span className="text-3xl font-bold text-white tabular-nums">
      {value.toLocaleString()}{suffix}
    </span>
  );
}

function HeroSection() {
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
            <span className="flywheel-hero-stat"><AnimatedCounter value={1234} /></span>
            <p className="text-sm text-indigo-200">Active Threads</p>
          </div>
          <div className="rounded-xl bg-white/10 p-4 backdrop-blur-sm">
            <span className="flywheel-hero-stat"><AnimatedCounter value={567} /></span>
            <p className="text-sm text-indigo-200">Verified Solutions</p>
          </div>
          <div className="rounded-xl bg-white/10 p-4 backdrop-blur-sm">
            <span className="flywheel-hero-stat"><AnimatedCounter value={12} /></span>
            <p className="text-sm text-indigo-200">Active Challenges</p>
          </div>
          <div className="rounded-xl bg-white/10 p-4 backdrop-blur-sm">
            <span className="flywheel-hero-stat"><AnimatedCounter value={89000} suffix="+" /></span>
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
  const { data: challengesData, isLoading: challengesLoading } = useChallenges({ status: 'active' }, 2);
  const { data: leaderboardData, isLoading: leaderboardLoading } = useLeaderboard({ type: 'overall', limit: 5 });

  const categories = [
    { name: 'Algorithms', count: 245, icon: Zap, color: 'bg-blue-500' },
    { name: 'Data Structures', count: 189, icon: Target, color: 'bg-violet-500' },
    { name: 'System Design', count: 156, icon: Cpu, color: 'bg-emerald-500' },
    { name: 'Optimization', count: 98, icon: TrendingUp, color: 'bg-amber-500' },
    { name: 'Machine Learning', count: 87, icon: Sparkles, color: 'bg-pink-500' },
    { name: 'Web Development', count: 234, icon: Code, color: 'bg-cyan-500' },
  ];

  return (
    <FlywheelPageLayout>
      {/* SEO Meta Tags */}
      <MetaTags
        title="Flywheel Network - Proof-of-Execution Knowledge Network"
        description="Join the Flywheel Network - a proof-of-execution knowledge network where developers compete, collaborate, and earn rewards through algorithmic challenges and community contributions."
        keywords={["flywheel network", "proof-of-execution", "algorithmic challenges", "developer community", "coding competitions", "reputation system"]}
        url={`${window.location.origin}/flywheel`}
        type="website"
      />

      {/* Public Analytics (Hotjar for user behavior) */}
      <PublicAnalytics />

      <div className="space-y-8">
        {/* Hero */}
        <HeroSection />

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
                <div key={i} className="flywheel-skeleton h-64 animate-pulse rounded-xl bg-bg-tertiary" />
              ))}
            </div>
          ) : challengesData?.challenges.length ? (
            <div className="grid gap-4 md:grid-cols-2">
              {challengesData.challenges.map((challenge) => (
                <ChallengeCard key={challenge.id} challenge={challenge} />
              ))}
            </div>
          ) : (
            <FlywheelCard className="text-center py-8">
              <Trophy className="flywheel-empty-icon mx-auto h-12 w-12 text-text-muted" />
              <p className="flywheel-empty-text mt-2 text-text-secondary">No active challenges right now</p>
            </FlywheelCard>
          )}
        </FlywheelSection>

        {/* Categories */}
        <FlywheelSection
          title="Browse by Category"
          description="Find threads in your area of interest"
        >
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
            {categories.map((category) => (
              <CategoryCard
                key={category.name}
                {...category}
                onClick={() => navigate(`/flywheel/threads?category=${category.name.toLowerCase().replace(' ', '-')}`)}
              />
            ))}
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
                ) : threadsData?.threads.length ? (
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
              ) : leaderboardData?.leaders.length ? (
                <LeaderboardTable
                  entries={leaderboardData.leaders.slice(0, 5)}
                  compact
                />
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
                    <p className="flywheel-getting-title font-medium text-text-primary">Read the Guide</p>
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
                    <p className="flywheel-getting-title font-medium text-text-primary">Join the Community</p>
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
