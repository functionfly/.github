import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import {
  AlertCircle,
  ChevronDown,
  LayoutGrid,
  List,
  Plus,
  RefreshCw,
  Search,
  SlidersHorizontal,
  X,
} from 'lucide-react';
import { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AppCard, AppCardSkeleton } from './components/AppCard';
import { AppsEmptyState } from './components/AppsEmptyState';
import { CreateAppModal } from './components/CreateAppModal';
import { useApps, type SortOption } from './hooks/useApps';

type ViewMode = 'grid' | 'list';

export function AppsListPage() {
  const { t } = useTranslation();
  const [viewMode, setViewMode] = useState<ViewMode>('grid');

  const SORT_LABELS: Record<SortOption, string> = {
    'created-desc': t('appsPage.sortNewestFirst'),
    'created-asc': t('appsPage.sortOldestFirst'),
    'name-asc': t('appsPage.sortNameAZ'),
    'name-desc': t('appsPage.sortNameZA'),
  };
  const searchInputRef = useRef<HTMLInputElement>(null);

  const {
    apps,
    allApps,
    isLoading,
    error,
    refetch,
    searchQuery,
    setSearchQuery,
    sortOption,
    setSortOption,
  } = useApps();

  const hasApps = allApps.length > 0;
  const isFiltered = searchQuery.trim().length > 0;
  const showEmpty = !isLoading && !error && apps.length === 0;

  const handleClearSearch = () => {
    setSearchQuery('');
    searchInputRef.current?.focus();
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t('appsPage.appsTitle')}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {isLoading
              ? t('appsPage.loadingApps')
              : hasApps
                ? `${t('appsPage.appCount', { count: allApps.length })}${isFiltered ? ` · ${t('appsPage.shownCount', { count: apps.length })}` : ''}`
                : t('appsPage.manageAppsDescription')}
          </p>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <CreateAppModal
            onSuccess={() => refetch()}
            trigger={
              <Button className="gap-2">
                <Plus className="w-4 h-4" />
                {t('appsPage.createApp')}
              </Button>
            }
          />
        </div>
      </div>

      {/* Error State */}
      {error && (
        <div className="flex items-center gap-3 p-4 rounded-xl bg-destructive/10 border border-destructive/20 text-destructive">
          <AlertCircle className="w-5 h-5 flex-shrink-0" />
          <div className="flex-1 min-w-0">
            <p className="font-medium text-sm">{t('appsPage.failedToLoadApps')}</p>
            <p className="text-xs opacity-80 mt-0.5">
              {error instanceof Error ? error.message : t('appsPage.unexpectedError')}
            </p>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            className="shrink-0 border-destructive/30 text-destructive hover:bg-destructive/10"
          >
            <RefreshCw className="w-3.5 h-3.5 mr-1.5" />
            {t('appsPage.retry')}
          </Button>
        </div>
      )}

      {/* Search & Filter Bar - only show when there are apps or loading */}
      {(hasApps || isLoading) && !error && (
        <div className="flex flex-col sm:flex-row gap-3">
          {/* Search */}
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
            <Input
              ref={searchInputRef}
              placeholder={t('appsPage.searchPlaceholder')}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 pr-9"
              aria-label={t('appsPage.searchApps')}
            />
            {searchQuery && (
              <button
                onClick={handleClearSearch}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                aria-label={t('appsPage.clearSearch')}
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>

          {/* Sort */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="gap-2 shrink-0">
                <SlidersHorizontal className="w-4 h-4" />
                <span className="hidden sm:inline">{SORT_LABELS[sortOption]}</span>
                <span className="sm:hidden">{t('appsPage.sort')}</span>
                <ChevronDown className="w-3.5 h-3.5 opacity-60" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-44">
              <DropdownMenuLabel className="text-xs text-muted-foreground font-normal">
                {t('appsPage.sortBy')}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              {(Object.keys(SORT_LABELS) as SortOption[]).map((option) => (
                <DropdownMenuItem
                  key={option}
                  onClick={() => setSortOption(option)}
                  className={cn('text-sm', sortOption === option && 'font-medium text-brand-500')}
                >
                  {SORT_LABELS[option]}
                  {sortOption === option && <span className="ml-auto text-brand-500">✓</span>}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>

          {/* View Mode Toggle */}
          <div className="flex items-center rounded-lg border border-border/60 p-0.5 bg-muted/30 shrink-0">
            <button
              onClick={() => setViewMode('grid')}
              className={cn(
                'p-1.5 rounded-md transition-all',
                viewMode === 'grid'
                  ? 'bg-background shadow-sm text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
              aria-label={t('appsPage.gridView')}
              aria-pressed={viewMode === 'grid'}
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={cn(
                'p-1.5 rounded-md transition-all',
                viewMode === 'list'
                  ? 'bg-background shadow-sm text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
              aria-label={t('appsPage.listView')}
              aria-pressed={viewMode === 'list'}
            >
              <List className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Loading State - Skeleton Cards */}
      {isLoading && (
        <div
          className={cn(
            viewMode === 'grid'
              ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4'
              : 'flex flex-col gap-3'
          )}
          aria-label={t('appsPage.loadingApps')}
          aria-busy="true"
        >
          {Array.from({ length: 6 }).map((_, i) => (
            <AppCardSkeleton key={i} />
          ))}
        </div>
      )}

      {/* Empty State */}
      {showEmpty && <AppsEmptyState isFiltered={isFiltered} searchQuery={searchQuery} />}

      {/* Apps Grid/List */}
      {!isLoading && !error && apps.length > 0 && (
        <div
          className={cn(
            viewMode === 'grid'
              ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4'
              : 'flex flex-col gap-3'
          )}
          role="list"
          aria-label={t('appsPage.appsList')}
        >
          {apps.map((app, index) => (
            <div key={app.id} role="listitem">
              <AppCard app={app} index={index} />
            </div>
          ))}
        </div>
      )}

      {/* Results count when filtered */}
      {!isLoading && !error && isFiltered && apps.length > 0 && (
        <p className="text-xs text-muted-foreground text-center">
          {t('appsPage.showingResults', { shown: apps.length, total: allApps.length })}
        </p>
      )}
    </div>
  );
}
