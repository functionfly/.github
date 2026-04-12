/**
 * Constants for the Usage Page
 */

export const USAGE_DAYS = 30;
export const MAX_AGENTS_FOR_USAGE = 25;

export const DATE_RANGES = [
  { label: 'Last 7 days', value: '7d', days: 7 },
  { label: 'Last 30 days', value: '30d', days: 30 },
  { label: 'Last 90 days', value: '90d', days: 90 },
  { label: 'This month', value: 'month', days: 30 },
  { label: 'Custom range', value: 'custom', days: 0 },
] as const;

export type DateRangeValue = (typeof DATE_RANGES)[number]['value'];

// Helper to convert date range to actual dates
export function getDateRangeDates(value: DateRangeValue, customRange?: { from: Date | null; to: Date | null }): { from: Date; to: Date } {
  const to = new Date();
  to.setHours(23, 59, 59, 999);

  if (value === 'custom' && customRange?.from && customRange?.to) {
    const from = new Date(customRange.from);
    from.setHours(0, 0, 0, 0);
    const toDate = new Date(customRange.to);
    toDate.setHours(23, 59, 59, 999);
    return { from, to: toDate };
  }

  const range = DATE_RANGES.find((r) => r.value === value);
  if (!range || range.days === 0) {
    // Default to 30 days
    const from = new Date(to);
    from.setDate(from.getDate() - 30);
    from.setHours(0, 0, 0, 0);
    return { from, to };
  }

  if (value === 'month') {
    // This month
    const from = new Date(to.getFullYear(), to.getMonth(), 1);
    from.setHours(0, 0, 0, 0);
    return { from, to };
  }

  const from = new Date(to);
  from.setDate(from.getDate() - range.days);
  from.setHours(0, 0, 0, 0);
  return { from, to };
}

export const COLORS = {
  execution: '#6366f1',
  compute: '#8b5cf6',
  platform_fee: '#ec4899',
  data_transfer: '#06b6d4',
  success: '#10b981',
  error: '#ef4444',
  cached: '#f59e0b',
};

export const REGION_COLORS = [
  '#6366f1', '#8b5cf6', '#ec4899', '#06b6d4', '#10b981',
  '#f59e0b', '#ef4444', '#84cc16', '#f97316', '#14b8a6',
];
