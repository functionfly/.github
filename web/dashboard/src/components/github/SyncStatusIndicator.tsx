import { motion } from 'framer-motion';
import {
  RefreshCw,
  Clock,
  GitBranch,
  History,
  CheckCircle,
  XCircle,
  Loader2,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { cn, formatDistanceToNow } from '@/lib/utils';
import { useUpdateSync, useSyncLogs } from '@/hooks/useGitHubSync';
import type { GitHubImport, GitHubSyncLog } from '@/types/github';

function getSyncStatusIcon(status: GitHubSyncLog['status']) {
  switch (status) {
    case 'completed':
      return <CheckCircle className="h-3.5 w-3.5 text-emerald-500" />;
    case 'failed':
      return <XCircle className="h-3.5 w-3.5 text-red-500" />;
    case 'building':
    case 'deploying':
      return <Loader2 className="h-3.5 w-3.5 text-blue-400 animate-spin" />;
    default:
      return <Clock className="h-3.5 w-3.5 text-text-muted" />;
  }
}

interface SyncStatusIndicatorProps {
  importId: string;
  autoSyncEnabled: boolean;
  syncBranches: string[] | null;
  lastSyncedAt?: string | null;
  onToggleSync?: (enabled: boolean) => void;
  onViewHistory?: () => void;
  className?: string;
}

export function SyncStatusIndicator({
  importId,
  autoSyncEnabled,
  syncBranches,
  lastSyncedAt,
  onToggleSync,
  onViewHistory,
  className,
}: SyncStatusIndicatorProps) {
  const updateSync = useUpdateSync(importId);
  const { data: syncLogs } = useSyncLogs(importId, { per_page: 5 });

  const handleToggle = (enabled: boolean) => {
    updateSync.mutate({ auto_sync_enabled: enabled });
    onToggleSync?.(enabled);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={className}
    >
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm flex items-center gap-2">
              <RefreshCw className={cn('h-4 w-4 text-text-muted', autoSyncEnabled && 'text-[#00D4FF]')} />
              Auto-Sync
            </CardTitle>
            <Switch
              checked={autoSyncEnabled}
              onCheckedChange={handleToggle}
              aria-label="Toggle auto-sync"
            />
          </div>
        </CardHeader>

        <CardContent className="space-y-3">
          {syncBranches && syncBranches.length > 0 && (
            <div className="flex items-center gap-1.5 flex-wrap">
              <GitBranch className="h-3 w-3 text-text-muted shrink-0" />
              {syncBranches.map((branch) => (
                <Badge key={branch} variant="secondary" className="text-[10px] font-mono">
                  {branch}
                </Badge>
              ))}
            </div>
          )}

          {lastSyncedAt && (
            <div className="flex items-center gap-2 text-xs text-text-muted">
              <Clock className="h-3 w-3" />
              <span>Last synced {formatDistanceToNow(lastSyncedAt, { addSuffix: true })}</span>
            </div>
          )}

          {syncLogs && syncLogs.logs.length > 0 && (
            <div className="space-y-1.5 pt-2 border-t border-border-subtle">
              {syncLogs.logs.slice(0, 3).map((log) => (
                <div key={log.id} className="flex items-center gap-2 text-xs">
                  {getSyncStatusIcon(log.status)}
                  <span className="text-text-secondary truncate flex-1">
                    {log.trigger_type === 'push' ? 'Push' : log.trigger_type.startsWith('pr_') ? 'PR' : log.trigger_type}
                    {log.trigger_branch && ` to ${log.trigger_branch}`}
                  </span>
                  {log.version_published && (
                    <Badge variant="outline" className="text-[9px] font-mono shrink-0">
                      v{log.version_published}
                    </Badge>
                  )}
                  <span className="text-text-muted shrink-0">
                    {formatDistanceToNow(log.created_at, { addSuffix: true })}
                  </span>
                </div>
              ))}
            </div>
          )}

          <Button
            variant="ghost"
            size="sm"
            className="w-full text-xs"
            onClick={onViewHistory}
            aria-label="View sync history"
          >
            <History className="h-3.5 w-3.5 mr-1.5" />
            View History
          </Button>
        </CardContent>
      </Card>
    </motion.div>
  );
}
