import { motion } from 'framer-motion';
import { Download, FolderSearch } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { EmptyState } from '@/components/ui/empty-state';
import { cn } from '@/lib/utils';

interface NoImportsYetProps {
  onBrowseRepos?: () => void;
  className?: string;
}

export function NoImportsYet({ onBrowseRepos, className }: NoImportsYetProps) {
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
        icon={<Download className="h-8 w-8" />}
        title="No imports yet"
        description="Import your first serverless function from a GitHub repository to get started."
        action={
          <Button
            variant="default"
            size="sm"
            onClick={onBrowseRepos}
            className="gap-2"
            aria-label="Browse repositories"
          >
            <FolderSearch className="h-3.5 w-3.5" />
            Browse Repos
          </Button>
        }
      />
    </motion.div>
  );
}
