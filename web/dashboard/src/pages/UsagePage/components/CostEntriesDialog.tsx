/**
 * Cost Entries drill-down dialog.
 * Shows paginated raw cost allocation entries for the selected date range.
 */
import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Loader2, ChevronLeft, ChevronRight, AlertCircle, ListTree } from 'lucide-react';
import { getCostEntries } from '@/api/usageAnalytics';
import { formatCostUsd, type CostAllocationEntry } from '@/api/usageAnalytics';
import { getDateRangeDates, type DateRangeValue } from '../constants';
import type { DateRangeSelection } from '@/components/ui/date-picker';

const PAGE_SIZE = 50;

function formatLocalDate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

export function CostEntriesDialog({
  open,
  onOpenChange,
  dateRange,
  customDateRange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  dateRange: DateRangeValue;
  customDateRange: DateRangeSelection;
}) {
  const [page, setPage] = useState(0);
  const [outcome, setOutcome] = useState<'all' | 'success' | 'error'>('all');

  const { from, to } = useMemo(
    () => getDateRangeDates(dateRange, customDateRange),
    [dateRange, customDateRange]
  );

  const { data, isLoading, error } = useQuery({
    queryKey: ['cost-entries', from.toISOString(), to.toISOString(), page, outcome],
    queryFn: () =>
      getCostEntries({
        startDate: formatLocalDate(from),
        endDate: formatLocalDate(to),
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
        outcome: outcome === 'all' ? undefined : outcome,
      }),
    enabled: open,
  });

  const entries = data?.entries ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[80vh] flex flex-col gap-0 p-0">
        <DialogHeader className="px-6 pt-5 pb-4 border-b border-border-subtle">
          <DialogTitle className="flex items-center gap-3">
            <span className="v-icon-brand w-9 h-9">
              <ListTree className="h-4 w-4" />
            </span>
            <div>
              <span className="text-base font-semibold">Cost allocation entries</span>
              <DialogDescription className="mt-0.5">
                {from.toLocaleDateString()} – {to.toLocaleDateString()} •{' '}
                {total.toLocaleString()} record{total === 1 ? '' : 's'}
              </DialogDescription>
            </div>
          </DialogTitle>
        </DialogHeader>

        <div className="flex items-center gap-2 px-6 pt-4">
          <span className="text-xs text-text-secondary font-medium">Outcome:</span>
          {(['all', 'success', 'error'] as const).map((o) => (
            <Button
              key={o}
              size="sm"
              variant={outcome === o ? 'default' : 'outline'}
              onClick={() => {
                setOutcome(o);
                setPage(0);
              }}
            >
              {o[0].toUpperCase() + o.slice(1)}
            </Button>
          ))}
        </div>

        <div className="flex-1 overflow-auto mx-6 my-3 border border-border-subtle rounded-lg">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-bg-secondary text-text-secondary">
              <tr>
                <th className="text-left p-2.5 font-medium">Time</th>
                <th className="text-left p-2.5 font-medium">Function</th>
                <th className="text-left p-2.5 font-medium">Region</th>
                <th className="text-left p-2.5 font-medium">Outcome</th>
                <th className="text-right p-2.5 font-medium">Cost</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 6 }).map((_, i) => (
                  <tr key={i} className="border-t border-border-subtle">
                    <td className="p-2.5">
                      <Skeleton className="h-4 w-32" />
                    </td>
                    <td className="p-2.5">
                      <Skeleton className="h-4 w-28" />
                    </td>
                    <td className="p-2.5">
                      <Skeleton className="h-4 w-20" />
                    </td>
                    <td className="p-2.5">
                      <Skeleton className="h-4 w-16" />
                    </td>
                    <td className="p-2.5 text-right">
                      <Skeleton className="h-4 w-16 ml-auto" />
                    </td>
                  </tr>
                ))
              ) : error ? (
                <tr>
                  <td colSpan={5} className="p-6 text-center text-text-muted">
                    <AlertCircle className="h-5 w-5 mx-auto mb-2 text-destructive" />
                    Failed to load entries
                  </td>
                </tr>
              ) : entries.length === 0 ? (
                <tr>
                  <td colSpan={5} className="p-8 text-center text-text-muted">
                    No cost entries in this range
                  </td>
                </tr>
              ) : (
                entries.map((e: CostAllocationEntry) => (
                  <tr
                    key={e.id}
                    className="border-t border-border-subtle hover:bg-bg-hover/50 transition-colors"
                  >
                    <td className="p-2.5 text-text-secondary whitespace-nowrap">
                      {new Date(e.timestamp).toLocaleString([], {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </td>
                    <td className="p-2.5 font-medium">{e.function_name}</td>
                    <td className="p-2.5 text-text-secondary">{e.region}</td>
                    <td className="p-2.5">
                      <span
                        className={
                          e.execution_outcome === 'success'
                            ? 'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/30'
                            : 'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-red-500/10 text-red-500 border border-red-500/30'
                        }
                      >
                        <span
                          className={`w-1.5 h-1.5 rounded-full ${
                            e.execution_outcome === 'success' ? 'bg-emerald-500' : 'bg-red-500'
                          }`}
                        />
                        {e.execution_outcome}
                      </span>
                    </td>
                    <td className="p-2.5 text-right font-mono font-medium">
                      {formatCostUsd(e.total_cost_cents)}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <DialogFooter className="flex items-center justify-between sm:justify-between px-6 py-4 border-t border-border-subtle">
          <div className="text-xs text-text-muted">
            Page {page + 1} of {totalPages}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page === 0 || isLoading}
              onClick={() => setPage((p) => Math.max(0, p - 1))}
            >
              <ChevronLeft className="h-4 w-4 mr-1" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page + 1 >= totalPages || isLoading}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
              <ChevronRight className="h-4 w-4 ml-1" />
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
