import { usePageTitle } from '@/hooks';
import { cn } from '@/lib/utils';
import {
  AlertCircle, ChevronDown, LayoutGrid, List,
  RefreshCw, Search, SlidersHorizontal, X, Plus,
} from 'lucide-react';
import { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  PageGrid, Chamber, CornerBrace, TrustSeal,
  SealedButton, FrameButton, StatusPill, AnnotationTag,
} from '@/components/containment';
import { AppCard, AppCardSkeleton } from './components/AppCard';
import { AppsEmptyState } from './components/AppsEmptyState';
import { CreateAppModal } from './components/CreateAppModal';
import { useApps, type SortOption } from './hooks/useApps';

import './apps.css';

type ViewMode = 'grid' | 'list';

export function AppsListPage() {
  usePageTitle('Apps');
  const { t } = useTranslation();
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [sortOpen, setSortOpen] = useState(false);

  const SORT_LABELS: Record<SortOption, string> = {
    'created-desc': t('appsPage.sortNewestFirst'),
    'created-asc': t('appsPage.sortOldestFirst'),
    'name-asc': t('appsPage.sortNameAZ'),
    'name-desc': t('appsPage.sortNameZA'),
  };
  const searchInputRef = useRef<HTMLInputElement>(null);

  const { apps, allApps, isLoading, error, refetch, searchQuery, setSearchQuery, sortOption, setSortOption } = useApps();

  const hasApps = allApps.length > 0;
  const isFiltered = searchQuery.trim().length > 0;
  const showEmpty = !isLoading && !error && apps.length === 0;

  const handleClearSearch = () => {
    setSearchQuery('');
    searchInputRef.current?.focus();
  };

  return (
    <div className="apps-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="apps-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE APP-01" secondary="Applications" position="top-right" />

        <div className="apps-hero__header">
          <div className="apps-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="apps-hero__title">{t('appsPage.appsTitle')}</h1>
            {hasApps && (
              <CreateAppModal
                onSuccess={() => refetch()}
                trigger={
                  <SealedButton size="sm" iconLeft={<Plus className="apps-icon-sm" />}>
                    New App
                  </SealedButton>
                }
              />
            )}
          </div>
          <p className="apps-hero__subtitle">
            {isLoading ? t('appsPage.loadingApps')
              : hasApps ? `${t('appsPage.appCount', { count: allApps.length })}${isFiltered ? ` · ${t('appsPage.shownCount', { count: apps.length })}` : ''}`
              : t('appsPage.manageAppsDescription')}
          </p>
        </div>
      </Chamber>

      {/* Error */}
      {error && (
        <div className="apps-error">
          <AlertCircle className="apps-error__icon" />
          <div className="apps-error__content">
            <p className="apps-error__title">{t('appsPage.failedToLoadApps')}</p>
            <p className="apps-error__message">{error instanceof Error ? error.message : t('appsPage.unexpectedError')}</p>
          </div>
          <FrameButton size="sm" onClick={() => refetch()} iconLeft={<RefreshCw className="apps-icon-xs" />}>
            {t('appsPage.retry')}
          </FrameButton>
        </div>
      )}

      {/* Search & Controls */}
      {(hasApps || isLoading) && !error && (
        <div className="apps-controls">
          <div className="apps-search">
            <Search className="apps-search__icon" />
            <input ref={searchInputRef} className="apps-input apps-search__input" placeholder={t('appsPage.searchPlaceholder')}
              value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} aria-label={t('appsPage.searchApps')} />
            {searchQuery && (
              <button className="apps-search__clear" onClick={handleClearSearch} aria-label={t('appsPage.clearSearch')}>
                <X className="apps-icon-xs" />
              </button>
            )}
          </div>

          {/* Sort Dropdown */}
          <div className="apps-sort">
            <button className="apps-sort__trigger" onClick={() => setSortOpen(!sortOpen)}>
              <SlidersHorizontal className="apps-icon-xs" />
              <span className="apps-sort__label">{SORT_LABELS[sortOption]}</span>
              <ChevronDown className={`apps-icon-xs ${sortOpen ? 'apps-sort__chevron--open' : ''}`} />
            </button>
            {sortOpen && (
              <div className="apps-sort__menu">
                <p className="apps-sort__menu-label">{t('appsPage.sortBy')}</p>
                {(Object.keys(SORT_LABELS) as SortOption[]).map((option) => (
                  <button key={option} className={`apps-sort__item ${sortOption === option ? 'apps-sort__item--active' : ''}`}
                    onClick={() => { setSortOption(option); setSortOpen(false); }}>
                    {SORT_LABELS[option]}
                    {sortOption === option && <span className="apps-sort__check">✓</span>}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* View Toggle */}
          <div className="apps-view-toggle">
            <button className={`apps-view-btn ${viewMode === 'grid' ? 'apps-view-btn--active' : ''}`} onClick={() => setViewMode('grid')} aria-label={t('appsPage.gridView')}>
              <LayoutGrid className="apps-icon-sm" />
            </button>
            <button className={`apps-view-btn ${viewMode === 'list' ? 'apps-view-btn--active' : ''}`} onClick={() => setViewMode('list')} aria-label={t('appsPage.listView')}>
              <List className="apps-icon-sm" />
            </button>
          </div>
        </div>
      )}

      {/* Loading Skeletons */}
      {isLoading && (
        <div className={cn(viewMode === 'grid' ? 'apps-grid' : 'apps-list')} aria-label={t('appsPage.loadingApps')} aria-busy="true">
          {Array.from({ length: 6 }).map((_, i) => <AppCardSkeleton key={i} />)}
        </div>
      )}

      {/* Empty State */}
      {showEmpty && <AppsEmptyState isFiltered={isFiltered} searchQuery={searchQuery} onCreateSuccess={() => refetch()} />}

      {/* Apps Grid/List */}
      {!isLoading && !error && apps.length > 0 && (
        <div className={cn(viewMode === 'grid' ? 'apps-grid' : 'apps-list')} role="list" aria-label={t('appsPage.appsList')}>
          {apps.map((app, index) => (
            <div key={app.id} role="listitem">
              <AppCard app={app} index={index} />
            </div>
          ))}
        </div>
      )}

      {/* Results count */}
      {!isLoading && !error && isFiltered && apps.length > 0 && (
        <p className="apps-results-count">
          {t('appsPage.showingResults', { shown: apps.length, total: allApps.length })}
        </p>
      )}
    </div>
  );
}
