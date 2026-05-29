import { PageHeader } from '@/components/layout/PageHeader';
import { PageLayout } from '@/components/layout/PageLayout';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useMarketplacePurchases } from '@/hooks/useMarketplacePurchases';
import { isMarketplacePurchasesEnabled } from '@/lib/marketplace-purchases';
import { Clock, LayoutList, Loader2, RefreshCw, ShoppingBag } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import {
  KIND_META,
  PAGE_SIZE,
  TAB_ORDER,
  type DateRangeFilter,
  type PurchaseKind,
  type PurchaseTab,
  type StatusFilter,
  type ViewMode,
} from './constants';
import { PurchaseSection, UnifiedPurchaseCard } from './components/PurchaseCards';
import { PurchaseCrossLinks } from './components/PurchaseCrossLinks';
import { PurchaseEmptyState } from './components/PurchaseEmptyState';
import { PurchaseSkeleton } from './components/PurchaseSkeleton';
import { PurchaseSummaryStats } from './components/PurchaseSummaryStats';
import { PurchaseTable } from './components/PurchaseTable';
import { PurchaseTimeline } from './components/PurchaseTimeline';
import { PurchaseToolbar } from './components/PurchaseToolbar';
import {
  buildUnifiedPurchases,
  computeTotalSpend,
  filterPurchases,
} from './utils';

const VALID_TABS = new Set<string>(TAB_ORDER);

function parseTab(value: string | null): PurchaseTab {
  if (value && VALID_TABS.has(value)) return value as PurchaseTab;
  return 'all';
}

export function MarketplacePurchasesPage() {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const featureEnabled = isMarketplacePurchasesEnabled();

  const [limit, setLimit] = useState(PAGE_SIZE);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [dateRange, setDateRange] = useState<DateRangeFilter>('all');
  const [viewMode, setViewMode] = useState<ViewMode>('cards');
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const activeTab = parseTab(searchParams.get('tab'));

  const { data, isLoading, isError, error, refetch, isFetching, dataUpdatedAt, isSuccess } =
    useMarketplacePurchases(limit, 0);

  useEffect(() => {
    if (isSuccess && !isFetching) {
      setLastUpdated(new Date(dataUpdatedAt));
    }
  }, [isSuccess, isFetching, dataUpdatedAt]);

  const apiEnabled = data?.enabled !== false;
  const allEntries = useMemo(() => buildUnifiedPurchases(data), [data]);

  const rawCounts = useMemo(
    () => ({
      functions: data?.functions?.length ?? 0,
      agents: data?.agents?.length ?? 0,
      licenses: data?.licenses?.length ?? 0,
      subscriptions: data?.subscriptions?.length ?? 0,
    }),
    [data]
  );

  const totals = useMemo(
    () => ({
      functions: data?.totalFunctions ?? rawCounts.functions,
      agents: data?.totalAgents ?? rawCounts.agents,
      licenses: data?.totalLicenses ?? rawCounts.licenses,
      subscriptions: data?.totalSubscriptions ?? rawCounts.subscriptions,
    }),
    [data, rawCounts]
  );

  const totalCount = rawCounts.functions + rawCounts.agents + rawCounts.licenses + rawCounts.subscriptions;
  const totalSpend = useMemo(() => computeTotalSpend(data), [data]);
  const showLoadMore =
    rawCounts.functions < totals.functions ||
    rawCounts.agents < totals.agents ||
    rawCounts.licenses < totals.licenses ||
    rawCounts.subscriptions < totals.subscriptions;

  const getFilteredForTab = (tab: PurchaseTab) => {
    const kind = tab === 'all' || tab === 'timeline' ? tab : tab;
    return filterPurchases(allEntries, {
      search,
      statusFilter,
      dateRange,
      kind,
    });
  };

  const renderTabContent = (tab: PurchaseTab) => {
    const entries = getFilteredForTab(tab);

    if (entries.length === 0) {
      return (
        <p className="rounded-lg border border-dashed border-aviation-border-panel px-6 py-10 text-center text-sm text-aviation-text-secondary">
          {t('purchasesPage.noResults')}
        </p>
      );
    }

    if (tab === 'timeline') {
      return <PurchaseTimeline entries={entries} t={t} />;
    }

    if (viewMode === 'table') {
      return <PurchaseTable entries={entries} t={t} />;
    }

    if (tab === 'all') {
      const kinds: PurchaseKind[] = ['function', 'agent', 'license', 'subscription'];
      return (
        <div className="space-y-6">
          {kinds.map((kind) => {
            const sectionEntries = entries.filter((e) => e.kind === kind);
            if (sectionEntries.length === 0) return null;
            return (
              <PurchaseSection key={kind} title={t(KIND_META[kind].tabKey)}>
                {sectionEntries.map((entry, i) => (
                  <UnifiedPurchaseCard key={entry.id} entry={entry} t={t} index={i} />
                ))}
              </PurchaseSection>
            );
          })}
        </div>
      );
    }

    return (
      <div className="space-y-3">
        {entries.map((entry, i) => (
          <UnifiedPurchaseCard key={entry.id} entry={entry} t={t} index={i} />
        ))}
      </div>
    );
  };
  const isEmptyRaw = !isLoading && !isError && apiEnabled && totalCount === 0;

  const setActiveTab = (tab: PurchaseTab) => {
    const next = new URLSearchParams(searchParams);
    if (tab === 'all') {
      next.delete('tab');
    } else {
      next.set('tab', tab);
    }
    setSearchParams(next, { replace: true });
  };

  const tabCount = (tab: PurchaseTab): number => {
    if (tab === 'all' || tab === 'timeline') return totalCount;
    const kindMap: Record<PurchaseKind, keyof typeof rawCounts> = {
      function: 'functions',
      agent: 'agents',
      license: 'licenses',
      subscription: 'subscriptions',
    };
    return rawCounts[kindMap[tab as PurchaseKind]] ?? 0;
  };

  if (!featureEnabled) {
    return (
      <PageLayout>
        <PageHeader title={t('purchasesPage.title')} subtitle={t('purchasesPage.subtitle')} />
        <div className="rounded-xl border border-dashed border-aviation-border-panel bg-aviation-bg-panel/40 px-6 py-16 text-center">
          <ShoppingBag className="mx-auto mb-4 h-10 w-10 text-aviation-cyan" />
          <h2 className="text-lg font-semibold text-aviation-text-primary">
            {t('purchasesPage.comingSoonTitle')}
          </h2>
          <p className="mx-auto mt-2 max-w-md text-sm text-aviation-text-secondary">
            {t('purchasesPage.comingSoonBody')}
          </p>
        </div>
      </PageLayout>
    );
  }

  return (
    <PageLayout>
      <PageHeader
        title={t('purchasesPage.title')}
        subtitle={t('purchasesPage.subtitle')}
        badges={[{ label: t('purchasesPage.badgeLedger'), variant: 'beta' }]}
        actions={[
          {
            label: t('purchasesPage.refresh'),
            onClick: () => refetch(),
            variant: 'outline',
            size: 'sm',
            icon: isFetching ? Loader2 : RefreshCw,
            disabled: isFetching,
          },
        ]}
      />

      <div className="space-y-6">
        {!isLoading && !isError && apiEnabled && (
          <PurchaseSummaryStats
            counts={{
              ...rawCounts,
              totalSpend,
            }}
            totals={totals}
          />
        )}

        {isLoading && <PurchaseSkeleton />}

        {isError && (
          <div
            className="rounded-xl border border-destructive/30 bg-destructive/5 px-6 py-10 text-center"
            role="alert"
          >
            <p className="text-sm text-destructive">
              {t('purchasesPage.error')} {error instanceof Error ? error.message : ''}
            </p>
            <Button className="mt-4" variant="outline" onClick={() => refetch()}>
              {t('purchasesPage.tryAgain')}
            </Button>
          </div>
        )}

        {!isLoading && !isError && !apiEnabled && (
          <div className="rounded-xl border border-dashed border-aviation-border-panel px-6 py-16 text-center">
            <p className="text-sm text-aviation-text-secondary">{t('purchasesPage.disabled')}</p>
          </div>
        )}

        {isEmptyRaw && <PurchaseEmptyState />}

        {!isLoading && !isError && apiEnabled && totalCount > 0 && (
          <>
            <PurchaseToolbar
              search={search}
              onSearchChange={setSearch}
              statusFilter={statusFilter}
              onStatusFilterChange={setStatusFilter}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
              viewMode={viewMode}
              onViewModeChange={setViewMode}
              showViewToggle={activeTab !== 'timeline'}
              lastUpdated={lastUpdated}
            />

            <Tabs value={activeTab} onValueChange={(v) => setActiveTab(parseTab(v))} className="w-full">
              <TabsList className="mb-6 flex h-auto w-full flex-nowrap justify-start gap-1 overflow-x-auto pb-1">
                {TAB_ORDER.map((tab) => {
                    const count = tabCount(tab);
                    const hidden = tab !== 'all' && tab !== 'timeline' && count === 0;
                    if (hidden) return null;

                    const labelKey =
                      tab === 'all'
                        ? 'purchasesPage.tabAll'
                        : tab === 'timeline'
                          ? 'purchasesPage.tabTimeline'
                          : KIND_META[tab].tabKey;
                    const Icon =
                      tab === 'timeline'
                        ? Clock
                        : tab === 'all'
                          ? LayoutList
                          : KIND_META[tab as PurchaseKind].icon;

                    return (
                      <TabsTrigger
                        key={tab}
                        value={tab}
                        className="shrink-0 gap-1.5 data-[state=active]:bg-aviation-bg-instrument"
                      >
                        <Icon className="h-3.5 w-3.5" />
                        {t(labelKey)}
                        <span className="text-[10px] text-aviation-text-dim">({count})</span>
                      </TabsTrigger>
                    );
                  })}
              </TabsList>

              {TAB_ORDER.map((tab) => (
                <TabsContent key={tab} value={tab} className="space-y-4">
                  {renderTabContent(tab)}
                </TabsContent>
              ))}
            </Tabs>

            {showLoadMore && (
              <div className="flex justify-center pt-2">
                <Button
                  variant="outline"
                  onClick={() => setLimit((l) => l + PAGE_SIZE)}
                  disabled={isFetching}
                >
                  {isFetching ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : null}
                  {t('purchasesPage.loadMore')}
                </Button>
              </div>
            )}

            <PurchaseCrossLinks />
          </>
        )}

        {!isLoading && !isError && apiEnabled && totalCount === 0 && <PurchaseCrossLinks />}
      </div>
    </PageLayout>
  );
}

export default MarketplacePurchasesPage;
