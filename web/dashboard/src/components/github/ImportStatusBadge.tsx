import { Clock, Loader2, CheckCircle, XCircle } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { GitHubImport } from '@/types/github';

type ImportStatus = GitHubImport['status'];

interface StatusConfig {
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  className: string;
  spinning?: boolean;
}

const STATUS_MAP: Record<ImportStatus, StatusConfig> = {
  pending: {
    label: 'Pending',
    icon: Clock,
    className: 'text-gray-400 border-gray-400/30 bg-gray-400/5',
  },
  scanning: {
    label: 'Scanning',
    icon: Loader2,
    className: 'text-blue-400 border-blue-400/30 bg-blue-400/5',
    spinning: true,
  },
  configuring: {
    label: 'Configuring',
    icon: Loader2,
    className: 'text-blue-400 border-blue-400/30 bg-blue-400/5',
    spinning: true,
  },
  fetching: {
    label: 'Fetching',
    icon: Loader2,
    className: 'text-blue-400 border-blue-400/30 bg-blue-400/5',
    spinning: true,
  },
  building: {
    label: 'Building',
    icon: Loader2,
    className: 'text-blue-400 border-blue-400/30 bg-blue-400/5',
    spinning: true,
  },
  publishing: {
    label: 'Publishing',
    icon: Loader2,
    className: 'text-blue-400 border-blue-400/30 bg-blue-400/5',
    spinning: true,
  },
  completed: {
    label: 'Completed',
    icon: CheckCircle,
    className: 'text-emerald-500 border-emerald-500/30 bg-emerald-500/5',
  },
  failed: {
    label: 'Failed',
    icon: XCircle,
    className: 'text-red-500 border-red-500/30 bg-red-500/5',
  },
  cancelled: {
    label: 'Cancelled',
    icon: XCircle,
    className: 'text-gray-500 border-gray-500/30 bg-gray-500/5',
  },
};

interface ImportStatusBadgeProps {
  status: ImportStatus;
  className?: string;
}

export function ImportStatusBadge({ status, className }: ImportStatusBadgeProps) {
  const config = STATUS_MAP[status] ?? STATUS_MAP.pending;
  const Icon = config.icon;

  return (
    <Badge
      variant="outline"
      className={cn('gap-1.5 text-xs', config.className, className)}
    >
      <Icon className={cn('h-3 w-3', config.spinning && 'animate-spin')} />
      {config.label}
    </Badge>
  );
}
