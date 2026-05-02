import { motion } from 'framer-motion';
import { Search, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { EmptyState } from '@/components/ui/empty-state';
import { cn } from '@/lib/utils';
import { useRefreshGitHubRepos } from '@/hooks/useGitHubRepos';

interface NoReposFoundProps {
  onRefresh?: () => void;
  className?: string;
}

export function NoReposFound({ onRefresh, className }: NoReposFoundProps) {
  const refreshMutation = useRefreshGitHubRepos();

  const handleRefresh = () => {
    if (onRefresh) {
      onRefresh();
    } else {
      refreshMutation.mutate();
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={className}
    >
      <EmptyState
        variant="ghost"
        size="default"
        icon={<Search className="h-8 w-8" />}
        title="No repositories found"
        description="We couldn't find any repositories matching your criteria. Try adjusting your search or sync your repos."
        action={
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={refreshMutation.isPending}
            className="gap-2"
            aria-label="Refresh repositories"
          >
            <RefreshCw className={cn('h-3.5 w-3.5', refreshMutation.isPending && 'animate-spin')} />
            {refreshMutation.isPending ? 'Syncing...' : 'Refresh Repos'}
          </Button>
        }
      />
    </motion.div>
  );
}
