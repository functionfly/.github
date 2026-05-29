import type { ModelCatalogItem, ModelSelection } from '@/api/aiModels';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useResponsivePageSize } from '@/hooks/useResponsivePageSize';
import { cn } from '@/lib/utils';
import { AlertCircle, ArrowDown, ArrowUp, Loader2, RefreshCw, Search } from 'lucide-react';
import { useEffect, useMemo, useState, type MouseEvent } from 'react';
import {
  enabledModelCount,
  isModelEnabled,
  sortCatalog,
  type CatalogSortField,
  type CatalogSortOrder,
} from './utils';

const SORT_OPTIONS: { value: CatalogSortField; label: string }[] = [
  { value: 'name', label: 'Model name' },
  { value: 'provider', label: 'Provider' },
  { value: 'tier', label: 'Tier' },
  { value: 'cost', label: 'Cost' },
  { value: 'enabled', label: 'Enabled status' },
  { value: 'availability', label: 'Availability' },
];

type Props = {
  catalog: ModelCatalogItem[];
  enabledModels: ModelSelection[];
  isLoading?: boolean;
  error?: string | null;
  onToggle: (provider: string, modelId: string, enabled: boolean) => void;
  onEnableAll: () => void;
  onAllowAll: () => void;
  onRefresh?: () => void;
  isRefreshing?: boolean;
  disabled?: boolean;
};

const CAPABILITY_COLORS: Record<string, string> = {
  code: 'border-sky-500/30 text-sky-400',
  chat: 'border-violet-500/30 text-violet-400',
  tools: 'border-amber-500/30 text-amber-400',
  embedding: 'border-emerald-500/30 text-emerald-400',
};

const TIER_LABELS: Record<string, string> = {
  frontier: 'Frontier',
  fast: 'Fast',
  reasoning: 'Reasoning',
  code: 'Code',
  embedding: 'Embedding',
  local: 'Local',
  balanced: 'Other',
};

const TIER_COLORS: Record<string, string> = {
  frontier: 'border-purple-500/40 text-purple-300',
  fast: 'border-cyan-500/40 text-cyan-300',
  reasoning: 'border-orange-500/40 text-orange-300',
  code: 'border-sky-500/40 text-sky-300',
  embedding: 'border-emerald-500/40 text-emerald-300',
  local: 'border-zinc-500/40 text-zinc-300',
  balanced: 'border-border-subtle text-text-muted',
};

export function ModelCatalogTable({
  catalog,
  enabledModels,
  isLoading = false,
  error = null,
  onToggle,
  onEnableAll,
  onAllowAll,
  onRefresh,
  isRefreshing = false,
  disabled = false,
}: Props) {
  const [search, setSearch] = useState('');
  const [sortField, setSortField] = useState<CatalogSortField>('name');
  const [sortOrder, setSortOrder] = useState<CatalogSortOrder>('asc');
  const [page, setPage] = useState(1);
  const pageSize = useResponsivePageSize({ min: 6, max: 18, rowHeight: 52 });

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    const matches = !q
      ? catalog
      : catalog.filter(
          (m) =>
            m.display_name.toLowerCase().includes(q) ||
            m.id.toLowerCase().includes(q) ||
            m.provider.toLowerCase().includes(q) ||
            m.capabilities?.some((c) => c.toLowerCase().includes(q))
        );
    return sortCatalog(matches, sortField, sortOrder, enabledModels);
  }, [catalog, search, sortField, sortOrder, enabledModels]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize));

  useEffect(() => {
    setPage(1);
  }, [search, sortField, sortOrder, pageSize, catalog.length]);

  useEffect(() => {
    if (page > totalPages) {
      setPage(totalPages);
    }
  }, [page, totalPages]);

  const paginated = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filtered.slice(start, start + pageSize);
  }, [filtered, page, pageSize]);

  const rangeStart = filtered.length === 0 ? 0 : (page - 1) * pageSize + 1;
  const rangeEnd = Math.min(page * pageSize, filtered.length);

  const goToPage = (next: number) => (e: MouseEvent) => {
    e.preventDefault();
    if (next >= 1 && next <= totalPages && next !== page) {
      setPage(next);
    }
  };

  const enabledCount = enabledModelCount(enabledModels, catalog);
  const allEnabled = enabledModels.length === 0;

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-sm font-medium text-text-primary">Available models</p>
          <p className="text-sm text-text-muted">
            {allEnabled
              ? `All ${catalog.length} models enabled for your organization`
              : `${enabledCount} of ${catalog.length} models enabled`}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onAllowAll}
            disabled={disabled || allEnabled || catalog.length === 0}
          >
            Allow all
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onEnableAll}
            disabled={disabled || catalog.length === 0}
          >
            Restrict to all listed
          </Button>
          {onRefresh && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onRefresh}
              disabled={isRefreshing}
              className="gap-2"
            >
              {isRefreshing ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="h-4 w-4" />
              )}
              Refresh
            </Button>
          )}
        </div>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search models, providers, capabilities…"
            className="pl-9"
          />
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <Select value={sortField} onValueChange={(v) => setSortField(v as CatalogSortField)}>
            <SelectTrigger className="w-[180px]" aria-label="Sort models by">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              {SORT_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="shrink-0"
            onClick={() => setSortOrder((o) => (o === 'asc' ? 'desc' : 'asc'))}
            aria-label={sortOrder === 'asc' ? 'Sort ascending' : 'Sort descending'}
            title={
              sortOrder === 'asc'
                ? 'Ascending — click for descending'
                : 'Descending — click for ascending'
            }
          >
            {sortOrder === 'asc' ? (
              <ArrowUp className="h-4 w-4" />
            ) : (
              <ArrowDown className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center gap-2 py-12 text-sm text-text-muted">
          <Loader2 className="h-5 w-5 animate-spin" />
          Loading model catalog…
        </div>
      ) : error ? (
        <div className="flex items-start gap-3 rounded-lg border border-error/30 bg-error/5 p-4 text-sm text-error">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <div>
            <p className="font-medium">Could not load models</p>
            <p className="mt-1 text-text-muted">{error}</p>
          </div>
        </div>
      ) : filtered.length === 0 ? (
        <div className="rounded-lg border border-dashed border-border-subtle px-4 py-10 text-center">
          <p className="text-sm font-medium text-text-primary">No models found</p>
          <p className="mt-1 text-sm text-text-muted">
            {catalog.length === 0
              ? 'FlyMind returned an empty catalog. Ensure ai-service is running and providers are configured, then refresh.'
              : 'Try a different search term.'}
          </p>
        </div>
      ) : (
        <div className="rounded-xl border border-border-subtle overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[72px]">Enabled</TableHead>
                <TableHead>Model</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead>Tier</TableHead>
                <TableHead>Capabilities</TableHead>
                <TableHead className="hidden md:table-cell">Cost hint</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {paginated.map((model) => {
                const enabled = isModelEnabled(enabledModels, model.provider, model.id);
                const providerReady = model.provider_available !== false;
                return (
                  <TableRow
                    key={`${model.provider}:${model.id}`}
                    className={providerReady ? undefined : 'opacity-70'}
                  >
                    <TableCell>
                      <Switch
                        checked={enabled}
                        onCheckedChange={(checked) => onToggle(model.provider, model.id, checked)}
                        disabled={disabled || !providerReady}
                        aria-label={`Toggle ${model.display_name}`}
                      />
                    </TableCell>
                    <TableCell>
                      <div className="font-medium text-text-primary">{model.display_name}</div>
                      <div className="font-mono text-xs text-text-muted">{model.id}</div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap items-center gap-1.5">
                        <Badge variant="outline" className="font-normal capitalize">
                          {model.provider}
                        </Badge>
                        {!providerReady && (
                          <Badge
                            variant="outline"
                            className="text-[10px] text-warning border-warning/40"
                          >
                            API key required
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={`text-[10px] uppercase ${TIER_COLORS[model.tier ?? 'balanced'] ?? TIER_COLORS.balanced}`}
                      >
                        {TIER_LABELS[model.tier ?? 'balanced'] ?? model.tier ?? 'Other'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {(model.capabilities ?? []).map((cap) => (
                          <Badge
                            key={cap}
                            variant="outline"
                            className={`text-[10px] uppercase ${CAPABILITY_COLORS[cap] ?? ''}`}
                          >
                            {cap}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="hidden md:table-cell text-text-muted">
                      {model.cost_hint || '—'}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}

      {!isLoading && !error && filtered.length > 0 && totalPages > 1 && (
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-text-muted">
            Showing{' '}
            <span className="font-medium text-text-primary">
              {rangeStart}–{rangeEnd}
            </span>{' '}
            of <span className="font-medium text-text-primary">{filtered.length}</span> models
            <span className="hidden sm:inline"> · {pageSize} per page</span>
          </p>
          <Pagination>
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious
                  href="#"
                  onClick={goToPage(page - 1)}
                  aria-disabled={page <= 1}
                  className={cn(page <= 1 && 'pointer-events-none opacity-40')}
                />
              </PaginationItem>
              {Array.from({ length: totalPages }, (_, i) => i + 1)
                .filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
                .reduce<(number | 'ellipsis')[]>((acc, p, i, arr) => {
                  if (i > 0 && p - (arr[i - 1] as number) > 1) acc.push('ellipsis');
                  acc.push(p);
                  return acc;
                }, [])
                .map((p, i) =>
                  p === 'ellipsis' ? (
                    <PaginationItem key={`ellipsis-${i}`}>
                      <span className="flex h-9 w-9 items-center justify-center text-text-muted">
                        …
                      </span>
                    </PaginationItem>
                  ) : (
                    <PaginationItem key={p}>
                      <PaginationLink href="#" isActive={p === page} onClick={goToPage(p)}>
                        {p}
                      </PaginationLink>
                    </PaginationItem>
                  )
                )}
              <PaginationItem>
                <PaginationNext
                  href="#"
                  onClick={goToPage(page + 1)}
                  aria-disabled={page >= totalPages}
                  className={cn(page >= totalPages && 'pointer-events-none opacity-40')}
                />
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      )}
    </div>
  );
}
