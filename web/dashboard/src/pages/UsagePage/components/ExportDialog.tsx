/**
 * Export dialog.
 * Generates a CSV download of cost allocation entries for the selected range.
 */
import { useState } from 'react';
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
import { Label } from '@/components/ui/label';
import { Download, Loader2, FileText } from 'lucide-react';
import { toast } from 'sonner';
import { getCostEntries } from '@/api/usageAnalytics';
import { getDateRangeDates, type DateRangeValue } from '../constants';
import type { DateRangeSelection } from '@/components/ui/date-picker';
import { useMemo } from 'react';

const FORMAT_OPTIONS = [
  { value: 'csv', label: 'CSV' },
  { value: 'json', label: 'JSON' },
];

function formatLocalDate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

export function ExportDialog({
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
  const [format, setFormat] = useState<'csv' | 'json'>('csv');
  const [exporting, setExporting] = useState(false);

  const { from, to } = useMemo(
    () => getDateRangeDates(dateRange, customDateRange),
    [dateRange, customDateRange]
  );

  const handleExport = async () => {
    setExporting(true);
    try {
      const all: unknown[] = [];
      const pageSize = 200;
      let offset = 0;
      let hasMore = true;
      while (hasMore) {
        const res = await getCostEntries({
          startDate: formatLocalDate(from),
          endDate: formatLocalDate(to),
          limit: pageSize,
          offset,
        });
        all.push(...res.entries);
        hasMore = res.has_more;
        offset += pageSize;
        if (all.length > 20000) {
          // Safety cap; warn user
          toast.warning('Export truncated at 20,000 rows');
          break;
        }
      }
      if (all.length === 0) {
        toast.error('No data to export in this range');
        return;
      }
      const blob = serialize(all, format);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `functionfly-usage-${formatLocalDate(from)}-to-${formatLocalDate(to)}.${format}`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
      toast.success(`Exported ${all.length.toLocaleString()} rows`);
      onOpenChange(false);
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Export failed';
      toast.error(msg);
    } finally {
      setExporting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-3">
            <span className="v-icon-brand w-9 h-9">
              <Download className="h-4 w-4" />
            </span>
            <div>
              <span className="text-base font-semibold">Export usage data</span>
              <DialogDescription className="mt-0.5">
                Download raw cost allocation entries for the selected range.
              </DialogDescription>
            </div>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="rounded-lg border border-border-subtle bg-bg-secondary/40 p-3">
            <Label className="text-xs text-text-muted">Date range</Label>
            <p className="text-sm font-medium mt-1">
              {from.toLocaleDateString()} – {to.toLocaleDateString()}
            </p>
          </div>
          <div className="space-y-2">
            <Label className="text-xs text-text-muted">Format</Label>
            <div className="flex gap-2">
              {FORMAT_OPTIONS.map((o) => (
                <Button
                  key={o.value}
                  size="sm"
                  variant={format === o.value ? 'default' : 'outline'}
                  onClick={() => setFormat(o.value as 'csv' | 'json')}
                  disabled={exporting}
                  className="flex-1"
                >
                  {o.label}
                </Button>
              ))}
            </div>
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={exporting}>
            Cancel
          </Button>
          <Button onClick={handleExport} disabled={exporting}>
            {exporting ? (
              <Loader2 className="h-4 w-4 animate-spin mr-1" />
            ) : (
              <FileText className="h-4 w-4 mr-1" />
            )}
            Export
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function serialize(rows: unknown[], format: 'csv' | 'json'): Blob {
  if (format === 'json') {
    return new Blob([JSON.stringify(rows, null, 2)], { type: 'application/json' });
  }
  // CSV
  if (rows.length === 0) return new Blob([''], { type: 'text/csv' });
  const first = rows[0] as Record<string, unknown>;
  const headers = Object.keys(first);
  const escape = (v: unknown): string => {
    if (v === null || v === undefined) return '';
    const s = typeof v === 'object' ? JSON.stringify(v) : String(v);
    if (s.includes(',') || s.includes('"') || s.includes('\n')) {
      return `"${s.replace(/"/g, '""')}"`;
    }
    return s;
  };
  const lines = [headers.join(',')];
  for (const r of rows) {
    const obj = r as Record<string, unknown>;
    lines.push(headers.map((h) => escape(obj[h])).join(','));
  }
  return new Blob([lines.join('\n')], { type: 'text/csv' });
}
