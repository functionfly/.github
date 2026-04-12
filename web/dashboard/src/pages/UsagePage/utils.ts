import type { DateRangeValue } from './constants';

export function getDateRange(range: DateRangeValue): { start: string; end: string } {
  const end = new Date();
  const start = new Date();
  switch (range) {
    case '7d': start.setDate(end.getDate() - 7); break;
    case '30d': start.setDate(end.getDate() - 30); break;
    case '90d': start.setDate(end.getDate() - 90); break;
    case 'month': start.setDate(1); break;
  }
  return {
    start: start.toISOString().split('T')[0],
    end: end.toISOString().split('T')[0],
  };
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

export function formatLimit(limit: number): string | number {
  return limit === 0 || limit === Number.MAX_SAFE_INTEGER ? 'Unlimited' : limit;
}
