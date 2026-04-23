import { motion } from 'framer-motion';
import {
  Server,
  Database,
  Zap,
  Shield,
  Cloud,
  HardDrive,
  Layers,
  Activity,
  Brain,
  CreditCard,
  Mail,
  Lock,
  Archive,
  Cpu,
  Globe,
  Workflow,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { SkeletonCard as DomainSkeletonCard, EmptyState } from '@/components/ui';
import type { ComponentHealth } from '@/api/status';

interface ComponentStatusProps {
  components: ComponentHealth[];
  isLoading?: boolean;
}

const categoryConfig = {
  core: {
    label: 'coreServices',
    icon: Server,
    color: 'text-brand-400',
  },
  provider: {
    label: 'providers',
    icon: Cloud,
    color: 'text-purple-400',
  },
  infrastructure: {
    label: 'infrastructure',
    icon: HardDrive,
    color: 'text-blue-400',
  },
};

// Order for displaying categories
const CATEGORY_ORDER = ['core', 'provider', 'infrastructure'];

const statusConfig = {
  operational: {
    dot: 'bg-emerald-500',
    text: 'text-emerald-400',
    label: 'operational',
    pulse: true,
  },
  degraded: {
    dot: 'bg-amber-500',
    text: 'text-amber-400',
    label: 'degraded',
    pulse: false,
  },
  down: {
    dot: 'bg-red-500',
    text: 'text-red-400',
    label: 'down',
    pulse: false,
  },
};

// Map component IDs to icons
const componentIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  // Core services
  api: Layers,
  orchestrator: Layers,
  database: Database,
  cache: Zap,
  'health-monitor': Activity,
  health_monitor: Activity,
  'ai-service': Brain,
  ai: Brain,
  embeddings: Brain,
  recommendations: Brain,
  // Infrastructure
  caddy: Shield,
  'state-fabric': Server,
  microvm: Cpu,
  queue: Workflow,
  'function-backup': Archive,
  email: Mail,
  billing: CreditCard,
  storage: HardDrive,
  cdn: Cloud,
  pgbouncer: Database,
  verification: Lock,
  'trust-api': Shield,
  support: Server,
  registry: Server,
  // Providers
  cloudflare: Globe,
  vercel: Cloud,
  fly: Cloud,
  deno: Cloud,
  functionfly_edge: Cloud,
};

function ComponentCard({
  component,
  index,
}: {
  component: ComponentHealth;
  index: number;
}) {
  const { t } = useTranslation();
  const status =
    statusConfig[component.status as keyof typeof statusConfig] ??
    statusConfig.operational;
  const Icon = componentIcons[component.id] || Server;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.3 }}
    >
      <Card
        className={cn(
          'group relative overflow-hidden transition-all duration-300',
          'hover:border-border-default hover:shadow-lg',
          component.status === 'degraded' && 'border-amber-500/30',
          component.status === 'down' && 'border-red-500/30'
        )}
      >
        <CardContent className="p-4">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <div
                className={cn(
                  'flex h-10 w-10 items-center justify-center rounded-lg',
                  'bg-bg-tertiary text-text-secondary',
                  'group-hover:text-text-primary transition-colors'
                )}
              >
                <Icon className="h-5 w-5" />
              </div>
              <div>
                <h3 className="font-medium text-text-primary">{component.name}</h3>
                <p className="text-xs text-text-muted">
                  {component.latency_ms > 0
                    ? t('statusPage.latencyMs', { ms: component.latency_ms })
                    : t('statusPage.noLatencyData')}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span className="relative flex h-2.5 w-2.5">
                {status.pulse && (
                  <span
                    className={cn(
                      'animate-ping absolute inline-flex h-full w-full rounded-full opacity-75',
                      status.dot
                    )}
                  />
                )}
                <span
                  className={cn(
                    'relative inline-flex rounded-full h-2.5 w-2.5',
                    status.dot
                  )}
                />
              </span>
              <span className={cn('text-xs font-medium', status.text)}>
                {t(`statusPage.${status.label}`)}
              </span>
            </div>
          </div>

          {/* Uptime indicator */}
          <div className="mt-4">
            <div className="flex items-center justify-between text-xs">
              <span className="text-text-muted">{t('statusPage.uptime')}</span>
              <span className={cn('font-medium', status.text)}>
                {component.uptime_percent.toFixed(2)}%
              </span>
            </div>
            <div className="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-bg-tertiary">
              <motion.div
                className={cn('h-full rounded-full', status.dot)}
                initial={{ width: 0 }}
                animate={{ width: `${component.uptime_percent}%` }}
                transition={{ duration: 0.5, delay: index * 0.05 }}
              />
            </div>
          </div>

          {component.message && (
            <p className="mt-3 text-xs text-text-secondary line-clamp-2">
              {component.message}
            </p>
          )}
        </CardContent>
      </Card>
    </motion.div>
  );
}

// Using standardized SkeletonCard component from ui library

export function ComponentStatus({ components, isLoading }: ComponentStatusProps) {
  const { t } = useTranslation();
  const list = Array.isArray(components) ? components : [];
  // Group components by category
  const groupedComponents = list.reduce((acc, component) => {
    const category = component.category;
    if (!acc[category]) {
      acc[category] = [];
    }
    acc[category].push(component);
    return acc;
  }, {} as Record<string, ComponentHealth[]>);

  const categories = Object.keys(groupedComponents).sort((a, b) => {
    return CATEGORY_ORDER.indexOf(a) - CATEGORY_ORDER.indexOf(b);
  });

  return (
    <section aria-label={t('statusPage.componentStatus')}>
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-text-primary">{t('statusPage.componentHealth')}</h2>
        <p className="mt-1 text-sm text-text-secondary">
          {t('statusPage.realtimeComponentStatus')}
        </p>
      </div>

      {isLoading ? (
        <div className="space-y-8">
          {[1, 2, 3].map((categoryIndex) => (
            <div key={categoryIndex}>
              <Skeleton className="mb-4 h-6 w-32" />
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {[1, 2, 3].map((i) => (
                  <DomainSkeletonCard key={i} variant="default" showImage={false} contentLines={2} footer={false} />
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : list.length === 0 ? (
        <EmptyState
          icon={<Server className="h-8 w-8" />}
          title={t('statusPage.noComponentData')}
          description={t('statusPage.noComponentDataDescription')}
          variant="card"
          size="sm"
        />
      ) : (
        <div className="space-y-8">
          {categories.map((category) => {
            const config = categoryConfig[category as keyof typeof categoryConfig] || {
              label: category,
              icon: Server,
              color: 'text-text-secondary',
            };
            const CategoryIcon = config.icon;

            return (
              <div key={category}>
                <div className="mb-4 flex items-center gap-2">
                  <CategoryIcon className={cn('h-5 w-5', config.color)} />
                  <h3 className="font-medium text-text-primary">{t(`statusPage.${config.label}`)}</h3>
                  <span className="text-sm text-text-muted">
                    ({groupedComponents[category].length})
                  </span>
                </div>
                <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                  {groupedComponents[category].map((component, index) => (
                    <ComponentCard
                      key={component.id}
                      component={component}
                      index={index}
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}
