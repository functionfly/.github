import { useEffect, useRef, useState } from 'react';
import { motion } from 'framer-motion';
import { Package, CheckCircle, RefreshCw, TrendingUp } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';

function useAnimatedCount(target: number, duration = 800): number {
  const [count, setCount] = useState(0);
  const frameRef = useRef<number | undefined>(null);

  useEffect(() => {
    if (target === 0) {
      setCount(0);
      return;
    }

    const startTime = performance.now();
    const startVal = 0;

    function tick(now: number) {
      const elapsed = now - startTime;
      const progress = Math.min(elapsed / duration, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      setCount(Math.round(startVal + (target - startVal) * eased));

      if (progress < 1) {
        frameRef.current = requestAnimationFrame(tick);
      }
    }

    frameRef.current = requestAnimationFrame(tick);

    return () => {
      if (frameRef.current) cancelAnimationFrame(frameRef.current);
    };
  }, [target, duration]);

  return count;
}

interface StatCardProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  count: number;
  color: string;
  bgColor: string;
  delay?: number;
}

function StatCard({ icon: Icon, label, count, color, bgColor, delay = 0 }: StatCardProps) {
  const animatedCount = useAnimatedCount(count);

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay }}
    >
      <Card className="overflow-hidden">
        <CardContent className="p-4">
          <div className="flex items-center gap-3">
            <div
              className={cn('flex h-10 w-10 items-center justify-center rounded-lg shrink-0', bgColor)}
            >
              <Icon className={cn('h-5 w-5', color)} />
            </div>
            <div>
              <p className="text-2xl font-bold text-text-primary tabular-nums">
                {animatedCount.toLocaleString()}
              </p>
              <p className="text-xs text-text-muted">{label}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

interface GitHubStatsBarProps {
  totalRepos?: number;
  imported?: number;
  activeSyncs?: number;
  thisWeek?: number;
  className?: string;
}

export function GitHubStatsBar({
  totalRepos = 0,
  imported = 0,
  activeSyncs = 0,
  thisWeek = 0,
  className,
}: GitHubStatsBarProps) {
  return (
    <div className={cn('grid grid-cols-2 lg:grid-cols-4 gap-3', className)}>
      <StatCard
        icon={Package}
        label="Total Repos"
        count={totalRepos}
        color="text-blue-500"
        bgColor="bg-blue-500/10"
        delay={0}
      />
      <StatCard
        icon={CheckCircle}
        label="Imported"
        count={imported}
        color="text-emerald-500"
        bgColor="bg-emerald-500/10"
        delay={0.05}
      />
      <StatCard
        icon={RefreshCw}
        label="Active Syncs"
        count={activeSyncs}
        color="text-[#00D4FF]"
        bgColor="bg-[#00D4FF]/10"
        delay={0.1}
      />
      <StatCard
        icon={TrendingUp}
        label="This Week"
        count={thisWeek}
        color="text-[#FF6B35]"
        bgColor="bg-[#FF6B35]/10"
        delay={0.15}
      />
    </div>
  );
}
