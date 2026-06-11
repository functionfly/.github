import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import {
  AlertCircle,
  ArrowUpRight,
  Box,
  Globe,
  Key,
  Layers,
  Server,
  Users,
  Zap,
} from 'lucide-react';

export interface QuotaItem {
  name: string;
  used: number;
  limit: number;
  icon: React.ReactNode;
  warningThreshold?: number;
  criticalThreshold?: number;
}

export interface QuotaUsageWidgetProps {
  quotas: QuotaItem[];
  className?: string;
  onUpgradeClick?: () => void;
}

function QuotaRow({
  name,
  used,
  limit,
  icon,
  warningThreshold = 70,
  criticalThreshold = 90,
  index,
}: QuotaItem & { index: number }) {
  // Handle unlimited (-1) or Infinity cases
  const isUnlimited = limit === -1 || limit === Infinity;
  const percentage = isUnlimited ? 0 : Math.min(100, Math.round((used / limit) * 100));
  const isWarning = percentage >= warningThreshold && percentage < criticalThreshold;
  const isCritical = !isUnlimited && percentage >= criticalThreshold;

  const progressColor = isCritical
    ? 'bg-[var(--color-error)]'
    : isWarning
      ? 'bg-[var(--color-aviation-amber)]'
      : 'bg-[var(--color-success)]';

  const displayLimit = isUnlimited ? '∞' : limit.toLocaleString();
  const displayUsed = used.toLocaleString();

  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.4, delay: index * 0.1 }}
      className="space-y-2"
    >
      <div className="flex items-center justify-between text-sm">
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-bg-tertiary text-text-muted">
            {icon}
          </div>
          <span className="font-medium text-text-primary">{name}</span>
        </div>
        <div className="flex items-center gap-2">
          {isCritical && <AlertCircle className="w-4 h-4 text-(--color-error)" />}
          <span
            className={cn(
              'tabular-nums',
              isCritical && 'text-(--color-error) font-semibold',
              isWarning && 'text-(--color-aviation-amber)'
            )}
          >
            {displayUsed}
            <span className="text-text-muted"> / {displayLimit}</span>
          </span>
        </div>
      </div>
      <div className="relative">
        <Progress value={percentage} className="h-2 bg-bg-tertiary" />
        <motion.div
          className={cn(
            'absolute top-0 left-0 h-2 rounded-full transition-all duration-500',
            progressColor
          )}
          initial={{ width: 0 }}
          animate={{ width: isUnlimited ? '100%' : `${percentage}%` }}
          transition={{ duration: 0.8, delay: 0.2 + index * 0.1 }}
        />
        {/* Threshold markers */}
        <div
          className="absolute top-0 w-0.5 h-2 bg-white/30"
          style={{ left: `${warningThreshold}%` }}
        />
        <div
          className="absolute top-0 w-0.5 h-2 bg-white/50"
          style={{ left: `${criticalThreshold}%` }}
        />
      </div>
      <p className="text-xs text-text-muted">
        {isUnlimited ? 'Unlimited' : `${percentage}% used`}
        {isCritical && ' - Limit nearly reached'}
        {isWarning && !isCritical && ' - Approaching limit'}
      </p>
    </motion.div>
  );
}

export function QuotaUsageWidget({ quotas, className, onUpgradeClick }: QuotaUsageWidgetProps) {
  const hasCritical = quotas.some((q) => {
    if (q.limit === -1 || q.limit === Infinity) return false;
    return (q.used / q.limit) * 100 >= 90;
  });

  return (
    <Card className={cn('overflow-hidden', className)}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium text-text-secondary">Plan Usage</CardTitle>
          {hasCritical && (
            <motion.div
              initial={{ scale: 0.8 }}
              animate={{ scale: 1 }}
              className="flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-(--color-error)/10 text-(--color-error)"
            >
              <AlertCircle className="w-3 h-3" />
              <span>Limits reached</span>
            </motion.div>
          )}
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="space-y-5">
          {quotas.map((quota, index) => (
            <QuotaRow key={quota.name} {...quota} index={index} />
          ))}
        </div>

        {onUpgradeClick && hasCritical && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
          >
            <button
              onClick={onUpgradeClick}
              className="mt-5 w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-(--color-aviation-amber)/10 border border-(--color-aviation-amber)/30 text-(--color-aviation-amber) text-sm font-medium hover:bg-(--color-aviation-amber)/20 transition-colors"
            >
              Upgrade plan
              <ArrowUpRight className="w-4 h-4" />
            </button>
          </motion.div>
        )}
      </CardContent>
    </Card>
  );
}

// Helper to build quota items from plan limits
export const QUOTA_ICONS = {
  functions: <Box className="w-4 h-4" />,
  requests: <Zap className="w-4 h-4" />,
  secrets: <Key className="w-4 h-4" />,
  providers: <Server className="w-4 h-4" />,
  stateFabrics: <Layers className="w-4 h-4" />,
  agents: <Users className="w-4 h-4" />,
  customDomains: <Globe className="w-4 h-4" />,
} as const;

export type QuotaKey = keyof typeof QUOTA_ICONS;
