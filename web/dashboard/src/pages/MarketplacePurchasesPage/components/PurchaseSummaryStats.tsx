import { cn } from '@/lib/utils';
import { KIND_META } from '../constants';
import { formatUsd } from '../utils';
import { useTranslation } from 'react-i18next';

type PurchaseSummaryStatsProps = {
  counts: {
    functions: number;
    agents: number;
    licenses: number;
    subscriptions: number;
    totalSpend: number;
  };
  totals?: {
    functions: number;
    agents: number;
    licenses: number;
    subscriptions: number;
  };
};

export function PurchaseSummaryStats({ counts, totals }: PurchaseSummaryStatsProps) {
  const { t } = useTranslation();

  const tiles = [
    {
      key: 'function' as const,
      count: counts.functions,
      total: totals?.functions,
      labelKey: 'purchasesPage.statFunctions',
    },
    {
      key: 'agent' as const,
      count: counts.agents,
      total: totals?.agents,
      labelKey: 'purchasesPage.statAgents',
    },
    {
      key: 'license' as const,
      count: counts.licenses,
      total: totals?.licenses,
      labelKey: 'purchasesPage.statLicenses',
    },
    {
      key: 'subscription' as const,
      count: counts.subscriptions,
      total: totals?.subscriptions,
      labelKey: 'purchasesPage.statSubscriptions',
    },
  ];

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
      {tiles.map(({ key, count, total, labelKey }) => {
        const meta = KIND_META[key];
        const Icon = meta.icon;
        return (
          <div
            key={key}
            className="aviation-stat-card rounded-lg border border-aviation-border-instrument bg-aviation-bg-instrument/50 p-4 transition-colors hover:border-aviation-border-panel"
          >
            <div className="flex items-center gap-3">
              <div className={cn('rounded-lg p-2', meta.iconBg)}>
                <Icon className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <p className="truncate text-xs uppercase tracking-wider text-aviation-text-muted">
                  {t(labelKey)}
                </p>
                <p className="text-2xl font-bold text-aviation-text-primary">{count}</p>
                {total != null && total > count && (
                  <p className="text-[10px] text-aviation-text-dim">
                    {t('purchasesPage.ofTotal', { total })}
                  </p>
                )}
              </div>
            </div>
          </div>
        );
      })}
      <div className="aviation-stat-card col-span-2 rounded-lg border border-aviation-border-instrument bg-aviation-bg-instrument/50 p-4 sm:col-span-1 lg:col-span-1">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-aviation-green/10 p-2 text-aviation-green">
            <span className="text-lg font-bold">$</span>
          </div>
          <div className="min-w-0">
            <p className="truncate text-xs uppercase tracking-wider text-aviation-text-muted">
              {t('purchasesPage.statTotalSpend')}
            </p>
            <p className="text-2xl font-bold text-aviation-text-primary">
              {formatUsd(counts.totalSpend)}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
