import { motion } from 'framer-motion';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { GitHubRepoCard } from '@/components/github/GitHubRepoCard';
import { NoReposFound } from '@/components/github/EmptyStates/NoReposFound';
import type { GitHubRepo } from '@/types/github';

function SkeletonCard() {
  return (
    <Card className="h-full">
      <CardHeader className="pb-3 space-y-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Skeleton className="h-4 w-4 rounded" />
            <Skeleton className="h-4 w-40" />
          </div>
          <Skeleton className="h-5 w-14 rounded-full" />
        </div>
        <Skeleton className="h-3 w-full" />
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex gap-3">
          <Skeleton className="h-3 w-12" />
          <Skeleton className="h-3 w-12" />
          <Skeleton className="h-3 w-20" />
        </div>
        <div className="flex gap-1.5">
          <Skeleton className="h-5 w-16 rounded-full" />
          <Skeleton className="h-5 w-16 rounded-full" />
        </div>
      </CardContent>
      <CardFooter className="pt-3 border-t border-border-subtle gap-2">
        <Skeleton className="h-8 flex-1 rounded-lg" />
        <Skeleton className="h-8 w-20 rounded-lg" />
      </CardFooter>
    </Card>
  );
}

interface GitHubRepoGridProps {
  repos?: GitHubRepo[];
  isLoading?: boolean;
  onImport?: (repo: GitHubRepo) => void;
  onScan?: (repo: GitHubRepo) => void;
  className?: string;
}

export function GitHubRepoGrid({ repos, isLoading, onImport, onScan, className }: GitHubRepoGridProps) {
  if (isLoading) {
    return (
      <div className={cn('grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4', className)}>
        {Array.from({ length: 6 }).map((_, i) => (
          <SkeletonCard key={i} />
        ))}
      </div>
    );
  }

  if (!repos || repos.length === 0) {
    return <NoReposFound className={className} />;
  }

  return (
    <div className={cn('grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4', className)}>
      {repos.map((repo, index) => (
        <motion.div
          key={repo.id}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: index * 0.05 }}
        >
          <GitHubRepoCard repo={repo} onImport={onImport} onScan={onScan} />
        </motion.div>
      ))}
    </div>
  );
}
