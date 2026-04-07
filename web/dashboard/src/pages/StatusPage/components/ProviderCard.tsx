import type { ComponentStatusType, ProviderStatus, RegionStatus } from '@/api/status';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import { AnimatePresence, motion } from 'framer-motion';
import { CheckCircle, ChevronDown, Clock, Cloud, Globe, Server, Zap } from 'lucide-react';
import { useState } from 'react';

interface ProviderCardProps {
  provider: ProviderStatus;
  index?: number;
}

interface ProviderCardSkeletonProps {
  index?: number;
}

const statusConfig: Record<ComponentStatusType, { color: string; label: string }> = {
  operational: { color: 'bg-emerald-500', label: 'Operational' },
  degraded: { color: 'bg-amber-500', label: 'Degraded' },
  down: { color: 'bg-red-500', label: 'Down' },
};

const providerIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  cloudflare: Cloud,
  vercel: Zap,
  fly: Server,
  deno: Globe,
  functionfly_edge: Cloud,
};

const providerColors: Record<string, string> = {
  cloudflare: 'from-orange-500 to-amber-500',
  vercel: 'from-gray-200 to-white',
  fly: 'from-purple-500 to-violet-500',
  deno: 'from-blue-500 to-cyan-500',
  functionfly_edge: 'from-brand-500 to-purple-500',
};

function RegionDot({ status, region }: { status: ComponentStatusType; region: string }) {
  const config = statusConfig[status];

  return (
    <div className="group/region relative" title={`${region}: ${config.label}`}>
      <span
        className={cn(
          'block h-3 w-3 rounded-full transition-transform duration-200',
          'hover:scale-125 cursor-pointer',
          config.color
        )}
      />
      {/* Tooltip */}
      <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 opacity-0 group-hover/region:opacity-100 transition-opacity pointer-events-none z-10">
        <div className="bg-bg-secondary border border-border-subtle rounded-lg px-3 py-1.5 shadow-lg whitespace-nowrap">
          <p className="text-xs font-medium text-text-primary">{region}</p>
          <p
            className={cn(
              'text-xs',
              status === 'operational'
                ? 'text-emerald-400'
                : status === 'degraded'
                  ? 'text-amber-400'
                  : 'text-red-400'
            )}
          >
            {config.label}
          </p>
        </div>
      </div>
    </div>
  );
}

function RegionRow({ region }: { region: RegionStatus }) {
  return (
    <div className="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-bg-tertiary/50 transition-colors">
      <div className="flex items-center gap-3">
        <RegionDot status={region.status} region={region.region} />
        <span className="text-sm font-medium text-text-primary">{region.region}</span>
      </div>
      <div className="flex items-center gap-4 text-xs text-text-muted">
        <span className="flex items-center gap-1">
          <Clock className="h-3 w-3" />
          {region.latency_ms}ms
        </span>
        <span className="flex items-center gap-1">
          <CheckCircle className="h-3 w-3" />
          {(region.success_rate * 100).toFixed(1)}%
        </span>
      </div>
    </div>
  );
}

export function ProviderCard({ provider, index = 0 }: ProviderCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const Icon = providerIcons[provider.id] || Cloud;
  const status = statusConfig[provider.status];
  const gradientClass = providerColors[provider.id] || 'from-gray-500 to-gray-400';

  const regions = Array.isArray(provider.regions) ? provider.regions : [];
  const operationalRegions = regions.filter((r) => r.status === 'operational').length;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.1, duration: 0.3 }}
    >
      <Card
        className={cn(
          'overflow-hidden transition-all duration-300',
          'hover:border-border-default',
          provider.status === 'degraded' && 'border-amber-500/30',
          provider.status === 'down' && 'border-red-500/30'
        )}
      >
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <div
                className={cn(
                  'flex h-10 w-10 items-center justify-center rounded-lg',
                  'bg-gradient-to-br',
                  gradientClass,
                  provider.id === 'vercel' ? 'text-gray-900' : 'text-white'
                )}
              >
                <Icon className="h-5 w-5" />
              </div>
              <div>
                <h3 className="font-semibold text-text-primary">{provider.name}</h3>
                <div className="flex items-center gap-2 mt-0.5">
                  <span className={cn('flex h-2 w-2 rounded-full', status.color)} />
                  <span className="text-xs text-text-secondary">{status.label}</span>
                </div>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
              className="h-8 w-8 p-0"
              aria-label={isExpanded ? 'Collapse regions' : 'Expand regions'}
            >
              <motion.div animate={{ rotate: isExpanded ? 180 : 0 }} transition={{ duration: 0.2 }}>
                <ChevronDown className="h-4 w-4 text-text-muted" />
              </motion.div>
            </Button>
          </div>
        </CardHeader>

        <CardContent className="pt-0">
          {/* Summary stats */}
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div className="rounded-lg bg-bg-tertiary/50 p-3">
              <p className="text-xs text-text-muted">Avg Latency</p>
              <p className="text-lg font-semibold text-text-primary">{provider.avg_latency_ms}ms</p>
            </div>
            <div className="rounded-lg bg-bg-tertiary/50 p-3">
              <p className="text-xs text-text-muted">Success Rate</p>
              <p className="text-lg font-semibold text-text-primary">
                {(provider.avg_success_rate * 100).toFixed(2)}%
              </p>
            </div>
          </div>

          {/* Region status dots */}
          <div className="mb-3">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs text-text-muted">Regions</span>
              <span className="text-xs text-text-secondary">
                {operationalRegions}/{regions.length} operational
              </span>
            </div>
            <div className="flex flex-wrap gap-2">
              {regions.map((region) => (
                <RegionDot key={region.region} status={region.status} region={region.region} />
              ))}
            </div>
          </div>

          {/* Expandable region details */}
          <AnimatePresence>
            {isExpanded && (
              <motion.div
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: 'auto', opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="overflow-hidden"
              >
                <div className="pt-3 border-t border-border-subtle">
                  <div className="space-y-1">
                    {regions.map((region) => (
                      <RegionRow key={region.region} region={region} />
                    ))}
                  </div>
                </div>
              </motion.div>
            )}
          </AnimatePresence>

          {/* Last updated */}
          <p className="mt-3 text-xs text-text-muted">
            Updated {new Date(provider.last_updated).toLocaleTimeString()}
          </p>
        </CardContent>
      </Card>
    </motion.div>
  );
}

export function ProviderCardSkeleton({ index = 0 }: ProviderCardSkeletonProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.1, duration: 0.3 }}
    >
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <Skeleton className="h-10 w-10 rounded-lg" />
              <div className="space-y-1.5">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-3 w-16" />
              </div>
            </div>
            <Skeleton className="h-8 w-8 rounded" />
          </div>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div className="rounded-lg bg-bg-tertiary/50 p-3 space-y-1.5">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-6 w-20" />
            </div>
            <div className="rounded-lg bg-bg-tertiary/50 p-3 space-y-1.5">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-6 w-20" />
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} className="h-3 w-3 rounded-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

interface ProviderGridProps {
  providers: ProviderStatus[];
  isLoading?: boolean;
}

export function ProviderGrid({ providers, isLoading }: ProviderGridProps) {
  const list = Array.isArray(providers) ? providers : [];
  return (
    <section aria-label="Provider Status">
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-text-primary">Provider Status</h2>
        <p className="mt-1 text-sm text-text-secondary">
          Health status across all edge providers and regions
        </p>
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5].map((i) => (
            <ProviderCardSkeleton key={i} index={i} />
          ))}
        </div>
      ) : list.length === 0 ? (
        <div className="rounded-xl border border-border-subtle bg-bg-secondary p-8 text-center">
          <p className="text-text-secondary">No provider data available</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {list.map((provider, index) => (
            <ProviderCard key={provider.id} provider={provider} index={index} />
          ))}
        </div>
      )}
    </section>
  );
}
