import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import { Calendar, LayoutGrid, List, Search } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { DATE_RANGE_OPTIONS, STATUS_FILTER_OPTIONS, type DateRangeFilter, type StatusFilter, type ViewMode } from '../constants';

type PurchaseToolbarProps = {
  search: string;
  onSearchChange: (value: string) => void;
  statusFilter: StatusFilter;
  onStatusFilterChange: (value: StatusFilter) => void;
  dateRange: DateRangeFilter;
  onDateRangeChange: (value: DateRangeFilter) => void;
  viewMode: ViewMode;
  onViewModeChange: (value: ViewMode) => void;
  showViewToggle: boolean;
  lastUpdated?: Date | null;
};

export function PurchaseToolbar({
  search,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  dateRange,
  onDateRangeChange,
  viewMode,
  onViewModeChange,
  showViewToggle,
  lastUpdated,
}: PurchaseToolbarProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="relative flex-1 max-w-md">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-aviation-text-dim" />
          <Input
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder={t('purchasesPage.searchPlaceholder')}
            className="pl-9 bg-aviation-bg-instrument/40 border-aviation-border-instrument"
            aria-label={t('purchasesPage.searchPlaceholder')}
          />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Select value={statusFilter} onValueChange={(v) => onStatusFilterChange(v as StatusFilter)}>
            <SelectTrigger className="w-[140px] border-aviation-border-instrument bg-aviation-bg-instrument/40">
              <SelectValue placeholder={t('purchasesPage.statusAll')} />
            </SelectTrigger>
            <SelectContent>
              {STATUS_FILTER_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {t(opt.labelKey)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select value={dateRange} onValueChange={(v) => onDateRangeChange(v as DateRangeFilter)}>
            <SelectTrigger className="w-[150px] border-aviation-border-instrument bg-aviation-bg-instrument/40">
              <Calendar className="mr-2 h-4 w-4 shrink-0" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {DATE_RANGE_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {t(opt.labelKey)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {showViewToggle && (
            <div className="flex rounded-lg border border-aviation-border-instrument p-0.5">
              <Button
                type="button"
                size="sm"
                variant={viewMode === 'cards' ? 'default' : 'ghost'}
                className={cn('h-8 px-2', viewMode === 'cards' && 'shadow-sm')}
                onClick={() => onViewModeChange('cards')}
                aria-pressed={viewMode === 'cards'}
              >
                <LayoutGrid className="h-4 w-4" />
                <span className="sr-only">{t('purchasesPage.viewCards')}</span>
              </Button>
              <Button
                type="button"
                size="sm"
                variant={viewMode === 'table' ? 'default' : 'ghost'}
                className={cn('h-8 px-2', viewMode === 'table' && 'shadow-sm')}
                onClick={() => onViewModeChange('table')}
                aria-pressed={viewMode === 'table'}
              >
                <List className="h-4 w-4" />
                <span className="sr-only">{t('purchasesPage.viewTable')}</span>
              </Button>
            </div>
          )}
        </div>
      </div>
      {lastUpdated && (
        <p className="text-xs text-aviation-text-dim">
          {t('purchasesPage.lastUpdated', {
            time: lastUpdated.toLocaleTimeString(undefined, {
              hour: 'numeric',
              minute: '2-digit',
            }),
          })}
        </p>
      )}
    </div>
  );
}
