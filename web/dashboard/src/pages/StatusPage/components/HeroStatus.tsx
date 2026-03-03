import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { CheckCircle, AlertTriangle, XCircle, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import type { PlatformStatus } from '@/api/status';

interface HeroStatusProps {
  status: PlatformStatus | null;
  lastUpdated: string | null;
  isLoading?: boolean;
  onRefresh?: () => void;
}

const statusConfig = {
  operational: {
    icon: CheckCircle,
    label: 'All Systems Operational',
    message: 'All services are running normally.',
    gradient: 'from-emerald-500/20 via-emerald-500/10 to-transparent',
    pulseColor: 'bg-emerald-500',
    textColor: 'text-emerald-400',
    borderColor: 'border-emerald-500/30',
  },
  degraded: {
    icon: AlertTriangle,
    label: 'Degraded Performance',
    message: 'Some services are experiencing issues.',
    gradient: 'from-amber-500/20 via-amber-500/10 to-transparent',
    pulseColor: 'bg-amber-500',
    textColor: 'text-amber-400',
    borderColor: 'border-amber-500/30',
  },
  major_outage: {
    icon: XCircle,
    label: 'Major Service Outage',
    message: 'Multiple services are currently unavailable.',
    gradient: 'from-red-500/20 via-red-500/10 to-transparent',
    pulseColor: 'bg-red-500',
    textColor: 'text-red-400',
    borderColor: 'border-red-500/30',
  },
};

export function HeroStatus({
  status,
  lastUpdated,
  isLoading,
  onRefresh,
}: HeroStatusProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const statusType = status?.status || 'operational';
  const config = statusConfig[statusType];
  const StatusIcon = config.icon;

  const formatLastUpdated = (date: string | null) => {
    if (!date) return 'Never';
    const d = new Date(date);
    return d.toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <section
      className={cn(
        'relative overflow-hidden rounded-2xl border backdrop-blur-sm',
        'bg-gradient-to-b',
        config.gradient,
        config.borderColor
      )}
      aria-label="Platform Status"
    >
      {/* Animated background pulse for operational status */}
      {statusType === 'operational' && mounted && (
        <div className="absolute inset-0 overflow-hidden">
          <motion.div
            className={cn(
              'absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2',
              'w-[600px] h-[600px] rounded-full opacity-20',
              config.pulseColor
            )}
            animate={{
              scale: [1, 1.2, 1],
              opacity: [0.1, 0.2, 0.1],
            }}
            transition={{
              duration: 4,
              repeat: Infinity,
              ease: 'easeInOut',
            }}
          />
        </div>
      )}

      <div className="relative px-6 py-12 sm:px-12 sm:py-16">
        <div className="flex flex-col items-center text-center">
          {/* Status Icon with Pulse */}
          <motion.div
            initial={{ scale: 0.8, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ duration: 0.5 }}
            className={cn(
              'relative mb-6 flex h-24 w-24 items-center justify-center rounded-full',
              'border-2 bg-bg-primary/50 backdrop-blur-sm',
              config.borderColor
            )}
          >
            {/* Pulse ring for non-operational statuses */}
            {statusType !== 'operational' && (
              <motion.span
                className={cn(
                  'absolute inline-flex h-full w-full rounded-full opacity-75',
                  config.pulseColor
                )}
                animate={{
                  scale: [1, 1.3, 1],
                  opacity: [0.5, 0, 0.5],
                }}
                transition={{
                  duration: 2,
                  repeat: Infinity,
                  ease: 'easeInOut',
                }}
              />
            )}
            <StatusIcon
              className={cn('relative h-12 w-12', config.textColor)}
              aria-hidden="true"
            />
          </motion.div>

          {/* Status Label */}
          <motion.h1
            initial={{ y: 20, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ delay: 0.1, duration: 0.5 }}
            className={cn(
              'text-3xl font-bold tracking-tight sm:text-4xl',
              config.textColor
            )}
          >
            {config.label}
          </motion.h1>

          {/* Status Message */}
          <motion.p
            initial={{ y: 20, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ delay: 0.2, duration: 0.5 }}
            className="mt-3 max-w-lg text-lg text-text-secondary"
          >
            {status?.message || config.message}
          </motion.p>

          {/* Last Updated & Refresh */}
          <motion.div
            initial={{ y: 20, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ delay: 0.3, duration: 0.5 }}
            className="mt-6 flex items-center gap-4"
          >
            <span className="text-sm text-text-muted">
              Last updated: {formatLastUpdated(lastUpdated)}
            </span>
            {onRefresh && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onRefresh}
                disabled={isLoading}
                className="h-8 px-2 text-text-muted hover:text-text-primary"
                aria-label="Refresh status"
              >
                <RefreshCw
                  className={cn('h-4 w-4', isLoading && 'animate-spin')}
                />
              </Button>
            )}
          </motion.div>

          {/* Component Summary */}
          {status?.components && status.components.length > 0 && (
            <motion.div
              initial={{ y: 20, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              transition={{ delay: 0.4, duration: 0.5 }}
              className="mt-8 flex flex-wrap items-center justify-center gap-4"
            >
              <div className="flex items-center gap-2 text-sm text-text-secondary">
                <span className="flex h-2 w-2 rounded-full bg-emerald-500" />
                <span>
                  {status.components.filter((c) => c.status === 'operational').length} Operational
                </span>
              </div>
              {status.components.some((c) => c.status === 'degraded') && (
                <div className="flex items-center gap-2 text-sm text-text-secondary">
                  <span className="flex h-2 w-2 rounded-full bg-amber-500" />
                  <span>
                    {status.components.filter((c) => c.status === 'degraded').length} Degraded
                  </span>
                </div>
              )}
              {status.components.some((c) => c.status === 'down') && (
                <div className="flex items-center gap-2 text-sm text-text-secondary">
                  <span className="flex h-2 w-2 rounded-full bg-red-500" />
                  <span>
                    {status.components.filter((c) => c.status === 'down').length} Down
                  </span>
                </div>
              )}
            </motion.div>
          )}
        </div>
      </div>
    </section>
  );
}
