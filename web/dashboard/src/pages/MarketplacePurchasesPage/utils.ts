import type {
  AgentHiringRow,
  BuyerLicenseRow,
  BuyerSubscriptionRow,
  FunctionPurchaseRow,
  MarketplacePurchasesResponse,
} from '@/hooks/useMarketplacePurchases';
import type { DateRangeFilter, PurchaseKind, StatusFilter } from './constants';

export type UnifiedPurchase = {
  id: string;
  kind: PurchaseKind;
  title: string;
  subtitle: string;
  dateMs: number;
  amount?: number;
  status: string;
  functionRow?: FunctionPurchaseRow;
  agentRow?: AgentHiringRow;
  licenseRow?: BuyerLicenseRow;
  subscriptionRow?: BuyerSubscriptionRow;
};

export function formatDate(ms: number): string {
  if (!ms) return '—';
  return new Date(ms).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function formatUsd(amount: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD' }).format(amount);
}

export function formatRelativeTime(
  ms: number,
  t: (key: string, opts?: Record<string, unknown>) => string
): string {
  if (!ms) return '—';
  const diffMs = Date.now() - ms;
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return t('purchasesPage.relativeJustNow');
  if (minutes < 60) return t('purchasesPage.relativeMinutes', { count: minutes });
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return t('purchasesPage.relativeHours', { count: hours });
  const days = Math.floor(hours / 24);
  if (days < 30) return t('purchasesPage.relativeDays', { count: days });
  const months = Math.floor(days / 30);
  if (months < 12) return t('purchasesPage.relativeMonths', { count: months });
  const years = Math.floor(months / 12);
  return t('purchasesPage.relativeYears', { count: years });
}

export function normalizeStatus(status: string): string {
  return status.trim().toLowerCase();
}

export function isActiveStatus(status: string): boolean {
  const s = normalizeStatus(status);
  return ['active', 'completed', 'paid', 'valid'].includes(s);
}

export function isRevokedStatus(status: string, revoked?: boolean): boolean {
  if (revoked) return true;
  const s = normalizeStatus(status);
  return ['revoked', 'cancelled', 'canceled', 'failed', 'expired'].includes(s);
}

export function buildUnifiedPurchases(data: MarketplacePurchasesResponse | undefined): UnifiedPurchase[] {
  if (!data) return [];

  const functions: UnifiedPurchase[] = (data.functions ?? []).map((row) => ({
    id: `fn-${row.id}`,
    kind: 'function' as const,
    title: `${row.functionAuthor}/${row.functionName}`,
    subtitle: row.agentId,
    dateMs: row.createdAt,
    amount: row.pricePaidUsd,
    status: row.status,
    functionRow: row,
  }));

  const agents: UnifiedPurchase[] = (data.agents ?? []).map((row) => ({
    id: `ag-${row.id}`,
    kind: 'agent' as const,
    title: row.taskType,
    subtitle: row.agentId,
    dateMs: row.createdAt,
    amount: row.budgetUsd,
    status: row.status,
    agentRow: row,
  }));

  const licenses: UnifiedPurchase[] = (data.licenses ?? []).map((row) => ({
    id: `lic-${row.id}`,
    kind: 'license' as const,
    title: row.functionName || row.functionId,
    subtitle: row.purchaserName,
    dateMs: row.issuedAt,
    status: row.revoked ? 'revoked' : row.type,
    licenseRow: row,
  }));

  const subscriptions: UnifiedPurchase[] = (data.subscriptions ?? []).map((row) => ({
    id: `sub-${row.id}`,
    kind: 'subscription' as const,
    title: row.planName,
    subtitle: row.billingCycle,
    dateMs: row.currentPeriodStart,
    amount: row.amount,
    status: row.status,
    subscriptionRow: row,
  }));

  return [...functions, ...agents, ...licenses, ...subscriptions].sort(
    (a, b) => b.dateMs - a.dateMs
  );
}

export function computeTotalSpend(data: MarketplacePurchasesResponse | undefined): number {
  if (!data) return 0;
  const fn = (data.functions ?? []).reduce((sum, r) => sum + (r.pricePaidUsd || 0), 0);
  const ag = (data.agents ?? []).reduce((sum, r) => sum + (r.budgetUsd || 0), 0);
  const sub = (data.subscriptions ?? []).reduce((sum, r) => sum + (r.amount || 0), 0);
  return fn + ag + sub;
}

function matchesDateRange(dateMs: number, range: DateRangeFilter): boolean {
  if (range === 'all' || !dateMs) return true;
  const days = range === '7d' ? 7 : range === '30d' ? 30 : 90;
  const cutoff = Date.now() - days * 24 * 60 * 60 * 1000;
  return dateMs >= cutoff;
}

function matchesStatus(entry: UnifiedPurchase, filter: StatusFilter): boolean {
  if (filter === 'all') return true;
  if (filter === 'revoked') {
    return isRevokedStatus(entry.status, entry.licenseRow?.revoked);
  }
  if (filter === 'active') {
    return isActiveStatus(entry.status) && !entry.licenseRow?.revoked;
  }
  return !isActiveStatus(entry.status) && !isRevokedStatus(entry.status, entry.licenseRow?.revoked);
}

export function filterPurchases(
  entries: UnifiedPurchase[],
  opts: {
    search: string;
    statusFilter: StatusFilter;
    dateRange: DateRangeFilter;
    kind?: PurchaseKind | 'all' | 'timeline';
  }
): UnifiedPurchase[] {
  const q = opts.search.trim().toLowerCase();
  return entries.filter((entry) => {
    if (opts.kind && opts.kind !== 'all' && opts.kind !== 'timeline' && entry.kind !== opts.kind) {
      return false;
    }
    if (!matchesDateRange(entry.dateMs, opts.dateRange)) return false;
    if (!matchesStatus(entry, opts.statusFilter)) return false;
    if (!q) return true;
    const haystack = [
      entry.title,
      entry.subtitle,
      entry.status,
      entry.functionRow?.functionAuthor,
      entry.functionRow?.functionName,
      entry.agentRow?.agentId,
      entry.agentRow?.taskType,
      entry.licenseRow?.functionId,
      entry.licenseRow?.keyPrefix,
      entry.subscriptionRow?.planName,
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase();
    return haystack.includes(q);
  });
}

export function hasMorePages(data: MarketplacePurchasesResponse | undefined, offset: number): boolean {
  if (!data) return false;
  const loaded =
    (data.functions?.length ?? 0) +
    (data.agents?.length ?? 0) +
    (data.licenses?.length ?? 0) +
    (data.subscriptions?.length ?? 0);
  const total =
    (data.totalFunctions ?? 0) +
    (data.totalAgents ?? 0) +
    (data.totalLicenses ?? 0) +
    (data.totalSubscriptions ?? 0);
  return loaded + offset < total;
}
