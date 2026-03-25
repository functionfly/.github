import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { formatDistanceToNow } from 'date-fns';
import { AlertTriangle, CheckCircle, Clock, XCircle } from 'lucide-react';
import { ProviderIcon } from './ProviderIcon';

interface ProviderStatusData {
  /** Stable unique key (e.g. DB row id). Prefer over `provider` when names may repeat. */
  id: string;
  provider: string;
  status: 'connected' | 'disconnected' | 'connecting' | 'error';
  lastChecked: Date | string;
  latency?: number;
  uptime?: number;
  errorMessage?: string;
  functionsCount?: number;
  deploymentsCount?: number;
}

interface ProviderStatusProps {
  providers: ProviderStatusData[];
  className?: string;
}

const statusConfig = {
  connected: {
    icon: CheckCircle,
    color: 'text-emerald-700 dark:text-green-400',
    bgColor: 'bg-emerald-500/10 dark:bg-green-400/10',
    borderColor: 'border-emerald-500/25 dark:border-green-400/20',
    label: 'Connected',
  },
  disconnected: {
    icon: XCircle,
    color: 'text-red-700 dark:text-red-400',
    bgColor: 'bg-red-500/10 dark:bg-red-400/10',
    borderColor: 'border-red-500/25 dark:border-red-400/20',
    label: 'Disconnected',
  },
  connecting: {
    icon: Clock,
    color: 'text-amber-700 dark:text-yellow-400',
    bgColor: 'bg-amber-500/10 dark:bg-yellow-400/10',
    borderColor: 'border-amber-500/25 dark:border-yellow-400/20',
    label: 'Connecting',
  },
  error: {
    icon: AlertTriangle,
    color: 'text-red-700 dark:text-red-400',
    bgColor: 'bg-red-500/10 dark:bg-red-400/10',
    borderColor: 'border-red-500/25 dark:border-red-400/20',
    label: 'Error',
  },
};

export function ProviderStatus({ providers, className }: ProviderStatusProps) {
  return (
    <div className={cn('grid gap-4 md:grid-cols-2 lg:grid-cols-3', className)}>
      {providers.map((providerData) => {
        const status = statusConfig[providerData.status];
        const StatusIcon = status.icon;

        return (
          <Card
            key={providerData.id}
            className={cn(
              'border transition-all duration-200 hover:shadow-lg',
              status.bgColor,
              status.borderColor
            )}
          >
            <CardContent className="p-4">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <ProviderIcon provider={providerData.provider} size="md" />
                  <div>
                    <h3 className="font-medium capitalize text-text-primary">
                      {providerData.provider}
                    </h3>
                    <div className="flex items-center gap-2 mt-1">
                      <StatusIcon className={cn('w-4 h-4', status.color)} />
                      <Badge
                        variant="secondary"
                        className={cn('text-xs', status.color, status.bgColor)}
                      >
                        {status.label}
                      </Badge>
                    </div>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex justify-between gap-3 text-sm">
                  <span className="text-muted-foreground shrink-0">Last checked</span>
                  <span className="text-right font-medium text-text-primary tabular-nums">
                    {formatDistanceToNow(new Date(providerData.lastChecked), { addSuffix: true })}
                  </span>
                </div>

                {providerData.latency !== undefined && (
                  <div className="flex justify-between gap-3 text-sm">
                    <span className="text-muted-foreground">Latency</span>
                    <span className="font-medium text-text-primary tabular-nums">
                      {providerData.latency}ms
                    </span>
                  </div>
                )}

                {providerData.uptime !== undefined && (
                  <div className="flex justify-between gap-3 text-sm">
                    <span className="text-muted-foreground">Uptime</span>
                    <span className="font-medium text-text-primary tabular-nums">
                      {providerData.uptime.toFixed(1)}%
                    </span>
                  </div>
                )}

                {providerData.functionsCount !== undefined && (
                  <div className="flex justify-between gap-3 text-sm">
                    <span className="text-muted-foreground">Functions</span>
                    <span className="font-medium text-text-primary">
                      {providerData.functionsCount}
                    </span>
                  </div>
                )}

                {providerData.deploymentsCount !== undefined && (
                  <div className="flex justify-between gap-3 text-sm">
                    <span className="text-muted-foreground">Deployments</span>
                    <span className="font-medium text-text-primary">
                      {providerData.deploymentsCount}
                    </span>
                  </div>
                )}

                {providerData.errorMessage && providerData.status === 'error' && (
                  <div className="mt-3 rounded-md border border-red-500/25 bg-red-500/10 p-2 dark:border-red-400/20 dark:bg-red-400/10">
                    <p className="text-xs text-red-800 dark:text-red-400">
                      {providerData.errorMessage}
                    </p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
