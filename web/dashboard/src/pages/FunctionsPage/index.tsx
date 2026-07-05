import { functionsApi } from '@/api/functions';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Badge } from '@/components/ui/badge';
import type { FunctionConfig } from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ColumnDef } from '@tanstack/react-table';
import {
  AlertTriangle,
  ArrowUpRight,
  Network,
  Check,
  Edit3,
  Eye,
  Filter,
  LayoutGrid,
  List,
  Loader2,
  Plus,
  Radar,
  Search,
  Trash2,
  X,
  Zap,
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { usePageTitle } from '@/hooks';
import { ROUTES } from '@/lib/constants';
import { toast } from 'sonner';
import { ToggleButtonGroup } from '@/components/ui';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
  Card,
} from '@/components/containment';
import './styles.css';

export function FunctionsPage() {
  usePageTitle('Functions');
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [functionToDelete, setFunctionToDelete] = useState<FunctionConfig | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [filterOpen, setFilterOpen] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string[]>([]);
  const [regionFilter, setRegionFilter] = useState<string[]>([]);

  const { data, isLoading, error } = useQuery({
    queryKey: ['functions'],
    queryFn: () => functionsApi.list(),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => functionsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['functions'] });
      toast.success(t('functionsPage.toastFunctionTerminated'));
      setDeleteDialogOpen(false);
      setFunctionToDelete(null);
    },
    onError: () => {
      toast.error(t('functionsPage.toastFailedToTerminate'));
      setDeleteDialogOpen(false);
    },
  });

  const bulkDeleteMutation = useMutation({
    mutationFn: async (ids: string[]) => {
      const results = await Promise.allSettled(ids.map((id) => functionsApi.delete(id)));
      const failed = results.filter((r) => r.status === 'rejected');
      if (failed.length > 0) throw new Error(`Failed to delete ${failed.length} functions`);
      return ids;
    },
    onSuccess: (_, ids) => {
      queryClient.invalidateQueries({ queryKey: ['functions'] });
      toast.success(t('functionsPage.toastFunctionsTerminated', { count: ids.length }));
    },
    onError: (_, ids) => {
      toast.error(t('functionsPage.toastFailedToTerminateSome', { count: ids.length }));
    },
  });

  const functions = (data?.functions ?? []).filter(
    (fn, index, self) => self.findIndex((f) => f.id === fn.id) === index
  );

  const columns = useMemo<ColumnDef<FunctionConfig>[]>(
    () => [
      {
        accessorKey: 'name',
        header: t('functionsPage.columnName'),
        size: 200,
        cell: ({ row }) => (
          <div className="sc-fns__cell-name">
            <span className="sc-fns__cell-name-text">{row.original.name}</span>
            <span className="sc-fns__cell-name-id">{row.original.id}</span>
          </div>
        ),
      },
      {
        accessorKey: 'status',
        header: t('functionsPage.columnStatus'),
        size: 120,
        cell: ({ row }) => {
          const s = row.original.status;
          const mapped = s === 'deployed' ? 'live' : s === 'failed' ? 'revoked' : 'pending';
          return <StatusPill status={mapped} label={s} />;
        },
      },
      {
        accessorKey: 'region',
        header: t('functionsPage.columnRegion'),
        size: 120,
        cell: ({ row }) => (
          <span className="sc-fns__cell-mono">{row.original.region}</span>
        ),
      },
      {
        accessorKey: 'providers',
        header: t('functionsPage.columnProviders'),
        size: 150,
        cell: ({ row }) => (
          <div className="sc-fns__cell-providers">
            {row.original.providers?.slice(0, 3).map((p) => (
              <span key={p} className="sc-fns__tag">{p}</span>
            ))}
            {row.original.providers && row.original.providers.length > 3 && (
              <span className="sc-fns__tag">+{row.original.providers.length - 3}</span>
            )}
          </div>
        ),
      },
      {
        accessorKey: 'trustScore',
        header: t('functionsPage.columnTrustScore'),
        size: 120,
        cell: ({ row }) => {
          const score = row.original.trustScore;
          if (score === undefined) return <span className="sc-fns__cell-dash">-</span>;
          const color = score >= 80 ? 'var(--status-ok)' : score >= 60 ? 'var(--status-pending)' : 'var(--status-revoked)';
          return (
            <div className="sc-fns__cell-score">
              <span style={{ color, fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{score}</span>
              <span className="sc-fns__cell-dash">/100</span>
            </div>
          );
        },
      },
      {
        accessorKey: 'createdAt',
        header: t('functionsPage.columnCreated'),
        size: 150,
        cell: ({ row }) => {
          const date = new Date(row.original.createdAt);
          return (
            <span className="sc-fns__cell-date">
              {date.toLocaleDateString()} {date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </span>
          );
        },
      },
      {
        id: 'actions',
        header: t('functionsPage.columnActions'),
        size: 150,
        enableSorting: false,
        cell: ({ row }) => (
          <div className="sc-fns__cell-actions">
            <button className="sc-fns__icon-btn" onClick={() => navigate(`/functions/${row.original.id}`)}><Eye size={14} /></button>
            <button className="sc-fns__icon-btn" onClick={() => navigate(`/functions/${row.original.id}/edit`)}><Edit3 size={14} /></button>
            <button className="sc-fns__icon-btn sc-fns__icon-btn--danger" onClick={() => handleDeleteClick(row.original)}><Trash2 size={14} /></button>
          </div>
        ),
      },
    ],
    [navigate, t]
  );

  const handleBulkAction = (action: string, selectedRows: FunctionConfig[]) => {
    if (action === 'delete') bulkDeleteMutation.mutate(selectedRows.map((r) => r.id));
  };

  const availableRegions = Array.from(new Set(functions.map((fn) => fn.region).filter(Boolean)));

  const filteredFunctions = functions.filter((fn) => {
    const matchesSearch = fn.name.toLowerCase().includes(searchQuery.toLowerCase()) || fn.id.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = statusFilter.length === 0 || statusFilter.includes(fn.status);
    const matchesRegion = regionFilter.length === 0 || regionFilter.includes(fn.region);
    return matchesSearch && matchesStatus && matchesRegion;
  });

  const activeFiltersCount = statusFilter.length + regionFilter.length;

  const toggleStatusFilter = (status: string) => setStatusFilter((prev) => prev.includes(status) ? prev.filter((s) => s !== status) : [...prev, status]);
  const toggleRegionFilter = (region: string) => setRegionFilter((prev) => prev.includes(region) ? prev.filter((r) => r !== region) : [...prev, region]);
  const clearFilters = () => { setStatusFilter([]); setRegionFilter([]); };
  const handleDeleteClick = (fn: FunctionConfig) => { setFunctionToDelete(fn); setDeleteDialogOpen(true); };
  const handleConfirmDelete = () => { if (functionToDelete) deleteMutation.mutate(functionToDelete.id); };

  const deployedCount = functions.filter((f) => f.providers?.length > 0).length;

  if (error) {
    return (
      <div className="sc-fns__page">
        <PageGrid />
        <Chamber className="sc-fns__error">
          <div className="sc-fns__error-icon"><AlertTriangle size={32} /></div>
          <h3 className="sc-fns__error-title">{t('functionsPage.systemError')}</h3>
          <p className="sc-fns__error-desc">{t('functionsPage.failedToLoadRegistry')}</p>
          <FrameButton iconLeft={<Radar size={14} />} onClick={() => queryClient.invalidateQueries({ queryKey: ['functions'] })}>
            {t('functionsPage.retryScan')}
          </FrameButton>
        </Chamber>
      </div>
    );
  }

  return (
    <div className="sc-fns__page">
      <PageGrid />

      {/* Header */}
      <div className="sc-fns__header">
        <Chamber ribs>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag primary="FLEET MANAGEMENT" secondary="ACTIVE" />
          <div className="sc-fns__header-top">
            <div className="sc-fns__header-info">
              <h1 className="sc-fns__title">{t('functionsPage.title')}</h1>
              <p className="sc-fns__subtitle">{t('functionsPage.subtitle')}</p>
            </div>
            <div className="sc-fns__header-actions">
              <SealedButton iconLeft={<Plus size={14} />} onClick={() => navigate('/functions/new')}>
                {t('functionsPage.deployNew')}
              </SealedButton>
              <FrameButton iconLeft={<Network size={14} />} onClick={() => navigate(ROUTES.FRG)}>
                {t('functionsPage.graphEditor')}
              </FrameButton>
            </div>
          </div>
          <GaugeStrip>
            <Gauge data={{ value: functions.length, label: 'Total' }} isFirst />
            <Gauge data={{ value: deployedCount, label: 'Online' }} />
            <Gauge data={{ value: filteredFunctions.length, label: 'Filtered' }} />
          </GaugeStrip>
        </Chamber>
      </div>

      {/* Toolbar */}
      <div className="sc-fns__toolbar">
        <div className="sc-fns__search">
          <Search size={14} className="sc-fns__search-icon" />
          <input
            type="text"
            className="sc-fns__search-input"
            placeholder={t('functionsPage.searchPlaceholder')}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchQuery && (
            <button className="sc-fns__search-clear" onClick={() => setSearchQuery('')}>
              <X size={12} />
            </button>
          )}
        </div>
        <div className="sc-fns__toolbar-controls">
          <ToggleButtonGroup
            value={viewMode}
            onValueChange={(v) => setViewMode(v as 'grid' | 'list')}
            options={[
              { value: 'grid', label: t('functionsPage.grid'), icon: <LayoutGrid size={14} /> },
              { value: 'list', label: t('functionsPage.list'), icon: <List size={14} /> },
            ]}
            variant="outline"
            size="sm"
          />
          <Popover open={filterOpen} onOpenChange={setFilterOpen}>
            <PopoverTrigger asChild>
              <Button variant="outline" className="sc-fns__filter-btn">
                <Filter size={14} />
                {t('functionsPage.filter')}
                {activeFiltersCount > 0 && <span className="sc-fns__filter-badge">{activeFiltersCount}</span>}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="sc-fns__filter-popover">
              <div className="sc-fns__filter-header">
                <span className="sc-fns__filter-title">{t('functionsPage.filterOptions')}</span>
                {activeFiltersCount > 0 && (
                  <button className="sc-fns__filter-clear" onClick={clearFilters}>
                    <X size={10} /> {t('functionsPage.clear')}
                  </button>
                )}
              </div>
              <div className="sc-fns__filter-section">
                <span className="sc-fns__filter-label">{t('functionsPage.status')}</span>
                {['draft', 'deployed', 'deploying', 'failed'].map((status) => (
                  <button key={status} className={`sc-fns__filter-option ${statusFilter.includes(status) ? 'active' : ''}`} onClick={() => toggleStatusFilter(status)}>
                    <div className={`sc-fns__filter-check ${statusFilter.includes(status) ? 'checked' : ''}`}>
                      {statusFilter.includes(status) && <Check size={10} />}
                    </div>
                    <span>{status}</span>
                  </button>
                ))}
              </div>
              {availableRegions.length > 0 && (
                <div className="sc-fns__filter-section">
                  <span className="sc-fns__filter-label">{t('functionsPage.region')}</span>
                  {availableRegions.map((region) => (
                    <button key={region} className={`sc-fns__filter-option ${regionFilter.includes(region) ? 'active' : ''}`} onClick={() => toggleRegionFilter(region)}>
                      <div className={`sc-fns__filter-check ${regionFilter.includes(region) ? 'checked' : ''}`}>
                        {regionFilter.includes(region) && <Check size={10} />}
                      </div>
                      <span>{region}</span>
                    </button>
                  ))}
                </div>
              )}
            </PopoverContent>
          </Popover>
        </div>
      </div>

      {/* Loading */}
      {isLoading && (
        <Chamber nested className="sc-fns__loading">
          <Loader2 size={24} className="sc-community-spinner" />
          <span className="sc-fns__loading-text">{t('functionsPage.initializing')}</span>
        </Chamber>
      )}

      {/* Content */}
      {!isLoading && (
        <>
          {filteredFunctions.length === 0 ? (
            <Chamber className="sc-fns__empty">
              <Radar size={40} className="sc-fns__empty-icon" />
              <p className="sc-fns__empty-title">
                {searchQuery ? `No results for "${searchQuery}"` : 'No functions deployed'}
              </p>
              <p className="sc-fns__empty-desc">
                {searchQuery ? 'Try a different search term or clear filters.' : 'Deploy your first function to get started.'}
              </p>
              {!searchQuery && (
                <SealedButton iconLeft={<Plus size={14} />} onClick={() => navigate('/functions/new')}>
                  {t('functionsPage.deployNew')}
                </SealedButton>
              )}
            </Chamber>
          ) : (
            <>
              <div className="sc-fns__results-bar">
                <span className="sc-fns__results-label">
                  {t('functionsPage.unitsDetected', { count: filteredFunctions.length, plural: filteredFunctions.length !== 1 ? 'S' : '' })}
                </span>
                {(searchQuery || activeFiltersCount > 0) && (
                  <button className="sc-fns__results-clear" onClick={() => { setSearchQuery(''); clearFilters(); }}>
                    {t('functionsPage.clearScan')} <ArrowUpRight size={12} />
                  </button>
                )}
              </div>

              {viewMode === 'grid' ? (
                <div className="sc-fns__grid">
                  {filteredFunctions.map((fn) => (
                    <Card key={fn.id} className="sc-fns__card">
                      <div className="sc-fns__card-header">
                        <div className="sc-fns__card-info">
                          <span className="sc-fns__card-name">{fn.name}</span>
                          <span className="sc-fns__card-id">{fn.id}</span>
                        </div>
                        <StatusPill status={fn.status === 'deployed' ? 'live' : fn.status === 'failed' ? 'revoked' : 'pending'} label={fn.status} />
                      </div>
                      <div className="sc-fns__card-meta">
                        {fn.region && <span className="sc-fns__tag">{fn.region}</span>}
                        {fn.providers?.slice(0, 2).map((p) => <span key={p} className="sc-fns__tag">{p}</span>)}
                      </div>
                      <div className="sc-fns__card-actions">
                        <button className="sc-fns__icon-btn" onClick={() => navigate(`/functions/${fn.id}`)}><Eye size={14} /></button>
                        <button className="sc-fns__icon-btn" onClick={() => navigate(`/functions/${fn.id}/edit`)}><Edit3 size={14} /></button>
                        {fn.trustScore !== undefined && (
                          <span className="sc-fns__card-score" style={{ color: fn.trustScore >= 80 ? 'var(--status-ok)' : fn.trustScore >= 60 ? 'var(--status-pending)' : 'var(--status-revoked)' }}>
                            {fn.trustScore}
                          </span>
                        )}
                      </div>
                    </Card>
                  ))}
                </div>
              ) : (
                <DataTable
                  data={filteredFunctions}
                  columns={columns}
                  enableRowSelection
                  enableColumnResize
                  enableColumnVisibility
                  enableExport
                  enableGlobalFilter
                  enableColumnFilters
                  onBulkAction={handleBulkAction}
                  bulkActions={[{ label: t('functionsPage.deleteSelected'), value: 'delete', variant: 'destructive' }]}
                  exportFileName={`functions-${new Date().toISOString().split('T')[0]}`}
                  isLoading={isLoading}
                />
              )}
            </>
          )}
        </>
      )}

      {/* Delete Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sc-fns__delete-dialog">
          <DialogHeader>
            <DialogTitle className="sc-fns__delete-title">
              <AlertTriangle size={18} />
              {t('functionsPage.confirmTermination')}
            </DialogTitle>
            <DialogDescription className="sc-fns__delete-desc">
              {t('functionsPage.confirmTerminationDesc', { name: functionToDelete?.name })}
            </DialogDescription>
          </DialogHeader>
          <div className="sc-fns__delete-warning">
            <span className="sc-fns__delete-warning-label">{t('functionsPage.warning')}</span>
            <ul>
              <li>{t('functionsPage.warningInvocations')}</li>
              <li>{t('functionsPage.warningProviders')}</li>
              <li>{t('functionsPage.warningLogs')}</li>
            </ul>
          </div>
          <DialogFooter className="sc-fns__delete-footer">
            <FrameButton onClick={() => setDeleteDialogOpen(false)}>
              {t('functionsPage.abort')}
            </FrameButton>
            <SealedButton
              loading={deleteMutation.isPending}
              iconLeft={<Zap size={14} />}
              onClick={handleConfirmDelete}
              className="sc-fns__delete-confirm"
            >
              {t('functionsPage.confirmTerminate')}
            </SealedButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default FunctionsPage;
