import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import { KIND_META } from '../constants';
import type { UnifiedPurchase } from '../utils';
import { formatDate, formatRelativeTime, formatUsd } from '../utils';
import { StatusLed } from './PurchaseCards';

type TFn = (key: string, opts?: Record<string, unknown>) => string;

export function PurchaseTimeline({ entries, t }: { entries: UnifiedPurchase[]; t: TFn }) {
  if (entries.length === 0) {
    return <p className="text-sm text-aviation-text-secondary">{t('purchasesPage.noResults')}</p>;
  }

  return (
    <div className="relative space-y-0">
      <div className="absolute bottom-2 left-[19px] top-2 w-px bg-aviation-border-instrument" aria-hidden />
      {entries.map((entry, index) => {
        const meta = KIND_META[entry.kind];
        const Icon = meta.icon;
        return (
          <motion.div
            key={entry.id}
            initial={{ opacity: 0, x: -8 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: Math.min(index * 0.03, 0.2) }}
            className="relative flex gap-4 pb-6 last:pb-0"
          >
            <div
              className={cn(
                'relative z-10 flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-aviation-border-instrument bg-aviation-bg-panel',
                meta.iconBg
              )}
            >
              <Icon className="h-4 w-4" />
            </div>
            <div
              className={cn(
                'min-w-0 flex-1 rounded-lg border border-aviation-border-instrument/80 bg-aviation-bg-instrument/30 p-4 border-l-4',
                meta.accent
              )}
            >
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div className="min-w-0">
                  <p className="font-medium text-aviation-text-primary">{entry.title}</p>
                  <p className="mt-0.5 text-xs text-aviation-text-dim">{entry.subtitle}</p>
                </div>
                <div className="flex flex-col items-end gap-1">
                  {entry.amount != null && (
                    <span className="font-mono text-sm text-aviation-text-primary">
                      {formatUsd(entry.amount)}
                    </span>
                  )}
                  <Badge variant="outline" className="text-[10px] capitalize">
                    {entry.kind}
                  </Badge>
                </div>
              </div>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-aviation-text-secondary">
                <StatusLed status={entry.status} revoked={entry.licenseRow?.revoked} />
                <span className="capitalize">
                  {entry.licenseRow?.revoked ? t('purchasesPage.revoked') : entry.status}
                </span>
                <span>·</span>
                <span title={formatDate(entry.dateMs)}>{formatRelativeTime(entry.dateMs, t)}</span>
              </div>
            </div>
          </motion.div>
        );
      })}
    </div>
  );
}
