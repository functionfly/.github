import { functionsApi } from '@/api/functions';
import { AviationEmptyState } from '@/components/functions/AviationEmptyState';
import { AviationFunctionCard } from '@/components/functions/AviationFunctionCard';
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

/**
 * Functions Page - Aviation-themed cockpit interface
 * Theme-aware: Industrial instrumentation aesthetic for both light/dark modes
 */
export function FunctionsPage() {
  usePageTitle('Functions');
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [functionToDelete, setFunctionToDelete] = useState<FunctionConfig | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [mounted, setMounted] = useState(false);
  const [filterOpen, setFilterOpen] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string[]>([]);
  const [regionFilter, setRegionFilter] = useState<string[]>([]);

  useEffect(() => {
    setMounted(true);
  }, []);

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

  // Bulk delete mutation
  const bulkDeleteMutation = useMutation({
    mutationFn: async (ids: string[]) => {
      const results = await Promise.allSettled(ids.map((id) => functionsApi.delete(id)));
      const failed = results.filter((r) => r.status === 'rejected');
      if (failed.length > 0) {
        throw new Error(`Failed to delete ${failed.length} functions`);
      }
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

  // Define table columns for list view
  const columns = useMemo<ColumnDef<FunctionConfig>[]>(
    () => [
      {
        accessorKey: 'name',
        header: t('functionsPage.columnName'),
        size: 200,
        cell: ({ row }) => (
          <div className="flex flex-col">
            <span className="font-medium">{row.original.name}</span>
            <span className="text-xs text-muted-foreground font-mono">{row.original.id}</span>
          </div>
        ),
      },
      {
        accessorKey: 'status',
        header: t('functionsPage.columnStatus'),
        size: 120,
        cell: ({ row }) => {
          const status = row.original.status;
          const statusColors: Record<string, string> = {
            deployed: 'bg-green-500/20 text-green-600 border-green-500/30',
            draft: 'bg-amber-500/20 text-amber-600 border-amber-500/30',
            deploying: 'bg-blue-500/20 text-blue-600 border-blue-500/30',
            failed: 'bg-red-500/20 text-red-600 border-red-500/30',
          };
          return (
            <Badge
              variant="outline"
              className={`${statusColors[status] || 'bg-gray-500/20 text-gray-600'} font-mono text-xs`}
            >
              {status}
            </Badge>
          );
        },
      },
      {
        accessorKey: 'region',
        header: t('functionsPage.columnRegion'),
        size: 120,
        cell: ({ row }) => (
          <span className="text-xs font-mono uppercase">{row.original.region}</span>
        ),
      },
      {
        accessorKey: 'providers',
        header: t('functionsPage.columnProviders'),
        size: 150,
        cell: ({ row }) => (
          <div className="flex flex-wrap gap-1">
            {row.original.providers?.slice(0, 3).map((provider) => (
              <Badge key={provider} variant="secondary" className="text-xs">
                {provider}
              </Badge>
            ))}
            {row.original.providers && row.original.providers.length > 3 && (
              <Badge variant="outline" className="text-xs">
                +{row.original.providers.length - 3}
              </Badge>
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
          if (score === undefined) return <span className="text-muted-foreground">-</span>;
          const color =
            score >= 80 ? 'text-green-500' : score >= 60 ? 'text-amber-500' : 'text-red-500';
          return (
            <div className="flex items-center gap-2">
              <span className={`font-mono font-semibold ${color}`}>{score}</span>
              <span className="text-xs text-muted-foreground">/100</span>
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
            <span className="text-sm text-muted-foreground">
              {date.toLocaleDateString()}{' '}
              {date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
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
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate(`/functions/${row.original.id}`)}
              className="h-8 w-8"
            >
              <Eye className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate(`/functions/${row.original.id}/edit`)}
              className="h-8 w-8"
            >
              <Edit3 className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => handleDeleteClick(row.original)}
              className="h-8 w-8 text-red-500 hover:text-red-600"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ),
      },
    ],
    [navigate, t]
  );

  // Handle bulk actions
  const handleBulkAction = (action: string, selectedRows: FunctionConfig[]) => {
    if (action === 'delete') {
      const ids = selectedRows.map((row) => row.id);
      bulkDeleteMutation.mutate(ids);
    }
  };

  // Get unique regions for filter
  const availableRegions = Array.from(new Set(functions.map((fn) => fn.region).filter(Boolean)));

  const filteredFunctions = functions.filter((fn) => {
    // Search filter
    const matchesSearch =
      fn.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      fn.id.toLowerCase().includes(searchQuery.toLowerCase());

    // Status filter
    const matchesStatus = statusFilter.length === 0 || statusFilter.includes(fn.status);

    // Region filter
    const matchesRegion = regionFilter.length === 0 || regionFilter.includes(fn.region);

    return matchesSearch && matchesStatus && matchesRegion;
  });

  const activeFiltersCount = statusFilter.length + regionFilter.length;

  const toggleStatusFilter = (status: string) => {
    setStatusFilter((prev) =>
      prev.includes(status) ? prev.filter((s) => s !== status) : [...prev, status]
    );
  };

  const toggleRegionFilter = (region: string) => {
    setRegionFilter((prev) =>
      prev.includes(region) ? prev.filter((r) => r !== region) : [...prev, region]
    );
  };

  const clearFilters = () => {
    setStatusFilter([]);
    setRegionFilter([]);
  };

  const handleDeleteClick = (fn: FunctionConfig) => {
    setFunctionToDelete(fn);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = () => {
    if (functionToDelete) {
      deleteMutation.mutate(functionToDelete.id);
    }
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setFunctionToDelete(null);
  };

  // Error state - theme aware
  if (error) {
    return (
      <div className="min-h-[calc(100vh-4rem)] aviation-radar-bg p-6 md:p-8">
        <div className="aviation-panel aviation-panel-glow p-12 text-center max-w-lg mx-auto mt-20">
          <div
            className="w-16 h-16 mx-auto mb-6 rounded-full flex items-center justify-center border"
            style={{
              background: 'var(--color-aviation-red-subtle, rgba(239,68,68,0.1))',
              borderColor: 'var(--color-aviation-red-dim, rgba(239,68,68,0.3))',
            }}
          >
            <AlertTriangle className="w-8 h-8" style={{ color: 'var(--color-aviation-red)' }} />
          </div>
          <h3
            className="text-lg font-mono font-semibold mb-2"
            style={{ color: 'var(--color-aviation-text-primary)' }}
          >
            {t('functionsPage.systemError')}
          </h3>
          <p
            className="font-mono text-sm mb-6"
            style={{ color: 'var(--color-aviation-text-secondary)' }}
          >
            {t('functionsPage.failedToLoadRegistry')}
          </p>
          <Button
            onClick={() => queryClient.invalidateQueries({ queryKey: ['functions'] })}
            className="aviation-button gap-2"
          >
            <Radar className="w-4 h-4" />
            {t('functionsPage.retryScan')}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-[calc(100vh-4rem)] aviation-radar-bg aviation-scroll overflow-y-auto">
      <div className="p-6 md:p-8 space-y-6">
        {/* Header Section */}
        <div
          className={`aviation-panel aviation-panel-glow p-6 transition-all duration-700 ${
            mounted ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}
        >
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
            {/* Title section */}
            <div className="flex items-start gap-4">
              {/* Icon badge - amber themed */}
              <div
                className="w-12 h-12 rounded-lg flex items-center justify-center shrink-0 border"
                style={{
                  background: 'var(--color-aviation-amber-subtle)',
                  borderColor: 'var(--color-aviation-amber-dim)',
                }}
              >
                <Radar className="w-6 h-6" style={{ color: 'var(--color-aviation-amber)' }} />
              </div>

              <div>
                <div className="flex items-center gap-2 mb-1">
                  <span className="aviation-label">{t('functionsPage.fleetManagement')}</span>
                  <span className="aviation-value-cyan text-xs">{t('functionsPage.active')}</span>
                </div>
                <h1
                  className="text-2xl md:text-3xl font-bold font-mono tracking-tight uppercase"
                  style={{ color: 'var(--color-aviation-text-primary)' }}
                >
                  {t('functionsPage.title')}
                </h1>
                <p
                  className="text-sm font-mono mt-1"
                  style={{ color: 'var(--color-aviation-text-secondary)' }}
                >
                  {t('functionsPage.subtitle')}
                </p>
              </div>
            </div>

            {/* Stats row */}
            <div className="flex items-center gap-6 lg:gap-8">
              <div className="text-right">
                <div className="aviation-label mb-0.5">{t('functionsPage.totalUnits')}</div>
                <div
                  className="text-2xl font-mono font-bold"
                  style={{ color: 'var(--color-aviation-amber)' }}
                >
                  {functions.length.toString().padStart(2, '0')}
                </div>
              </div>

              <div
                className="h-10 w-px"
                style={{ background: 'var(--color-aviation-border-panel)' }}
              />

              <div className="text-right">
                <div className="aviation-label mb-0.5">{t('functionsPage.online')}</div>
                <div
                  className="text-2xl font-mono font-bold"
                  style={{ color: 'var(--color-aviation-green)' }}
                >
                  {functions
                    .filter((f) => f.providers?.length > 0)
                    .length.toString()
                    .padStart(2, '0')}
                </div>
              </div>

              <div
                className="h-10 w-px"
                style={{ background: 'var(--color-aviation-border-panel)' }}
              />

              <Button
                onClick={() => navigate('/functions/new')}
                className="aviation-button aviation-button-primary gap-2 hidden md:flex"
              >
                <Plus className="w-4 h-4" />
                {t('functionsPage.deployNew')}
              </Button>

              <Button
                onClick={() => navigate(ROUTES.FRG)}
                variant="outline"
                className="aviation-button gap-2 hidden md:flex border-aviation-amber text-aviation-amber hover:bg-aviation-amber/10"
              >
                <Network className="w-4 h-4" />
                {t('functionsPage.graphEditor')}
              </Button>
            </div>
          </div>

          {/* Mobile CTA */}
          <Button
            onClick={() => navigate('/functions/new')}
            className="aviation-button aviation-button-primary gap-2 w-full mt-4 md:hidden"
          >
            <Plus className="w-4 h-4" />
            {t('functionsPage.deployNewFunction')}
          </Button>
        </div>

        {/* Toolbar */}
        <div
          className={`flex flex-col sm:flex-row gap-4 transition-all duration-700 delay-100 ${
            mounted ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}
        >
          {/* Search */}
          <div className="relative flex-1 max-w-xl">
            <Search
              className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 pointer-events-none"
              style={{ color: 'var(--color-aviation-text-muted)' }}
            />
            <input
              type="text"
              placeholder={t('functionsPage.searchPlaceholder')}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="aviation-input w-full pl-12 pr-4"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-4 top-1/2 -translate-y-1/2 transition-colors font-mono text-xs"
                style={{ color: 'var(--color-aviation-text-muted)' }}
              >
                {t('functionsPage.clear')}
              </button>
            )}
          </div>

          {/* Controls */}
          <div className="flex items-center gap-2">
            {/* View toggle using standardized ToggleButtonGroup */}
            <ToggleButtonGroup
              value={viewMode}
              onValueChange={(v) => setViewMode(v as 'grid' | 'list')}
              options={[
                {
                  value: 'grid',
                  label: t('functionsPage.grid'),
                  icon: <LayoutGrid className="h-4 w-4" />,
                },
                {
                  value: 'list',
                  label: t('functionsPage.list'),
                  icon: <List className="h-4 w-4" />,
                },
              ]}
              variant="outline"
              size="sm"
            />

            {/* Filter button */}
            <Popover open={filterOpen} onOpenChange={setFilterOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className="aviation-button gap-2 h-11 relative"
                  style={{
                    borderColor: activeFiltersCount > 0 ? 'var(--color-aviation-amber)' : undefined,
                  }}
                >
                  <Filter className="w-4 h-4" />
                  {t('functionsPage.filter')}
                  {activeFiltersCount > 0 && (
                    <span
                      className="absolute -top-1 -right-1 w-5 h-5 rounded-full text-xs flex items-center justify-center font-mono"
                      style={{
                        background: 'var(--color-aviation-amber)',
                        color: 'var(--color-aviation-bg-primary)',
                      }}
                    >
                      {activeFiltersCount}
                    </span>
                  )}
                </Button>
              </PopoverTrigger>
              <PopoverContent
                className="aviation-panel w-64 p-4"
                style={{
                  background: 'var(--color-aviation-bg-secondary)',
                  borderColor: 'var(--color-aviation-border-panel)',
                }}
              >
                <div className="flex items-center justify-between mb-4">
                  <span
                    className="font-mono text-sm font-semibold"
                    style={{ color: 'var(--color-aviation-text-primary)' }}
                  >
                    {t('functionsPage.filterOptions')}
                  </span>
                  {activeFiltersCount > 0 && (
                    <button
                      onClick={clearFilters}
                      className="font-mono text-xs flex items-center gap-1 transition-opacity hover:opacity-80"
                      style={{ color: 'var(--color-aviation-cyan)' }}
                    >
                      <X className="w-3 h-3" />
                      {t('functionsPage.clear')}
                    </button>
                  )}
                </div>

                {/* Status Filter */}
                <div className="mb-4">
                  <span
                    className="aviation-label block mb-2"
                    style={{ color: 'var(--color-aviation-text-muted)' }}
                  >
                    {t('functionsPage.status')}
                  </span>
                  <div className="space-y-2">
                    {['draft', 'deployed', 'deploying', 'failed'].map((status) => (
                      <button
                        key={status}
                        onClick={() => toggleStatusFilter(status)}
                        className="w-full flex items-center gap-2 p-2 rounded transition-colors"
                        style={{
                          background: statusFilter.includes(status)
                            ? 'var(--color-aviation-amber-subtle)'
                            : 'transparent',
                        }}
                      >
                        <div
                          className="w-4 h-4 rounded border flex items-center justify-center"
                          style={{
                            borderColor: statusFilter.includes(status)
                              ? 'var(--color-aviation-amber)'
                              : 'var(--color-aviation-border-instrument)',
                            background: statusFilter.includes(status)
                              ? 'var(--color-aviation-amber)'
                              : 'transparent',
                          }}
                        >
                          {statusFilter.includes(status) && (
                            <Check
                              className="w-3 h-3"
                              style={{ color: 'var(--color-aviation-bg-primary)' }}
                            />
                          )}
                        </div>
                        <span
                          className="font-mono text-xs uppercase"
                          style={{
                            color: statusFilter.includes(status)
                              ? 'var(--color-aviation-amber)'
                              : 'var(--color-aviation-text-secondary)',
                          }}
                        >
                          {status}
                        </span>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Region Filter */}
                {availableRegions.length > 0 && (
                  <div>
                    <span
                      className="aviation-label block mb-2"
                      style={{ color: 'var(--color-aviation-text-muted)' }}
                    >
                      {t('functionsPage.region')}
                    </span>
                    <div className="space-y-2 max-h-32 overflow-y-auto">
                      {availableRegions.map((region) => (
                        <button
                          key={region}
                          onClick={() => toggleRegionFilter(region)}
                          className="w-full flex items-center gap-2 p-2 rounded transition-colors"
                          style={{
                            background: regionFilter.includes(region)
                              ? 'var(--color-aviation-amber-subtle)'
                              : 'transparent',
                          }}
                        >
                          <div
                            className="w-4 h-4 rounded border flex items-center justify-center"
                            style={{
                              borderColor: regionFilter.includes(region)
                                ? 'var(--color-aviation-amber)'
                                : 'var(--color-aviation-border-instrument)',
                              background: regionFilter.includes(region)
                                ? 'var(--color-aviation-amber)'
                                : 'transparent',
                            }}
                          >
                            {regionFilter.includes(region) && (
                              <Check
                                className="w-3 h-3"
                                style={{ color: 'var(--color-aviation-bg-primary)' }}
                              />
                            )}
                          </div>
                          <span
                            className="font-mono text-xs uppercase"
                            style={{
                              color: regionFilter.includes(region)
                                ? 'var(--color-aviation-amber)'
                                : 'var(--color-aviation-text-secondary)',
                            }}
                          >
                            {region}
                          </span>
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </PopoverContent>
            </Popover>
          </div>
        </div>

        {/* Loading State */}
        {isLoading && (
          <div className="flex items-center justify-center py-20">
            <div className="text-center">
              <div className="relative w-16 h-16 mx-auto mb-4">
                <div
                  className="absolute inset-0 rounded-full border-2"
                  style={{ borderColor: 'var(--color-aviation-amber-dim)' }}
                />
                <div
                  className="absolute inset-0 rounded-full border-2 border-transparent animate-spin"
                  style={{ borderTopColor: 'var(--color-aviation-amber)' }}
                />
                <Loader2
                  className="absolute inset-0 m-auto w-6 h-6 animate-spin"
                  style={{ color: 'var(--color-aviation-amber)' }}
                />
              </div>
              <p className="aviation-label" style={{ color: 'var(--color-aviation-text-muted)' }}>
                {t('functionsPage.initializing')}
              </p>
            </div>
          </div>
        )}

        {/* Content */}
        {!isLoading && (
          <>
            {filteredFunctions.length === 0 ? (
              <AviationEmptyState
                onDeploy={() => navigate('/functions/new')}
                searchQuery={searchQuery}
              />
            ) : (
              <div
                className={`transition-all duration-700 delay-200 ${
                  mounted ? 'opacity-100' : 'opacity-0'
                }`}
              >
                {/* Results count */}
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <span className="aviation-label">{t('functionsPage.scanResults')}</span>
                    <span
                      className="font-mono text-sm font-semibold"
                      style={{ color: 'var(--color-aviation-text-primary)' }}
                    >
                      {t('functionsPage.unitsDetected', {
                        count: filteredFunctions.length,
                        plural: filteredFunctions.length !== 1 ? 'S' : '',
                      })}
                    </span>
                  </div>

                  {(searchQuery || activeFiltersCount > 0) && (
                    <button
                      onClick={() => {
                        setSearchQuery('');
                        clearFilters();
                      }}
                      className="font-mono text-xs flex items-center gap-1 transition-colors hover:opacity-80"
                      style={{ color: 'var(--color-aviation-cyan)' }}
                    >
                      {t('functionsPage.clearScan')}
                      <ArrowUpRight className="w-3 h-3" />
                    </button>
                  )}
                </div>

                {/* Functions Display - Grid or List View */}
                {viewMode === 'grid' ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                    {filteredFunctions.map((fn, index) => (
                      <AviationFunctionCard
                        key={fn.id}
                        fn={fn}
                        index={index}
                        onView={(id) => navigate(`/functions/${id}`)}
                        onEdit={(id) => navigate(`/functions/${id}/edit`)}
                        onDelete={handleDeleteClick}
                      />
                    ))}
                  </div>
                ) : (
                  <DataTable
                    data={filteredFunctions}
                    columns={columns}
                    enableRowSelection={true}
                    enableColumnResize={true}
                    enableColumnVisibility={true}
                    enableExport={true}
                    enableGlobalFilter={true}
                    enableColumnFilters={true}
                    onBulkAction={handleBulkAction}
                    bulkActions={[
                      {
                        label: t('functionsPage.deleteSelected'),
                        value: 'delete',
                        variant: 'destructive',
                      },
                    ]}
                    exportFileName={`functions-${new Date().toISOString().split('T')[0]}`}
                    isLoading={isLoading}
                    emptyState={
                      <AviationEmptyState
                        onDeploy={() => navigate('/functions/new')}
                        searchQuery={searchQuery}
                      />
                    }
                  />
                )}
              </div>
            )}
          </>
        )}
      </div>

      {/* Delete Confirmation Dialog - Theme Aware */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent
          className="aviation-panel max-w-md border"
          style={{
            background: 'var(--color-aviation-bg-secondary)',
            borderColor: 'var(--color-aviation-red-dim, rgba(239,68,68,0.3))',
          }}
        >
          <DialogHeader>
            <DialogTitle
              className="flex items-center gap-2 font-mono"
              style={{ color: 'var(--color-aviation-text-primary)' }}
            >
              <AlertTriangle className="w-5 h-5" style={{ color: 'var(--color-aviation-red)' }} />
              {t('functionsPage.confirmTermination')}
            </DialogTitle>
            <DialogDescription
              className="font-mono text-sm"
              style={{ color: 'var(--color-aviation-text-secondary)' }}
            >
              {t('functionsPage.confirmTerminationDesc', { name: functionToDelete?.name })}
            </DialogDescription>
          </DialogHeader>

          <div
            className="my-4 p-3 border rounded"
            style={{
              background: 'var(--color-aviation-red-subtle, rgba(239,68,68,0.05))',
              borderColor: 'var(--color-aviation-red-dim, rgba(239,68,68,0.2))',
            }}
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="aviation-label" style={{ color: 'var(--color-aviation-red)' }}>
                {t('functionsPage.warning')}
              </span>
            </div>
            <ul
              className="text-xs font-mono space-y-1 ml-4"
              style={{ color: 'var(--color-aviation-text-secondary)' }}
            >
              <li>• {t('functionsPage.warningInvocations')}</li>
              <li>• {t('functionsPage.warningProviders')}</li>
              <li>• {t('functionsPage.warningLogs')}</li>
            </ul>
          </div>

          <DialogFooter className="gap-2">
            <Button
              variant="outline"
              onClick={handleCancelDelete}
              className="aviation-button gap-2"
            >
              {t('functionsPage.abort')}
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleteMutation.isPending}
              className="font-mono text-sm px-4 py-2 rounded-md transition-all flex items-center gap-2"
              style={{
                background: 'var(--color-aviation-red)',
                color: 'white',
              }}
            >
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  {t('functionsPage.terminating')}
                </>
              ) : (
                <>
                  <Zap className="w-4 h-4" />
                  {t('functionsPage.confirmTerminate')}
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Global styles for animations */}
      <style>{`
        @keyframes aviation-blip {
          0%, 100% { opacity: 0; transform: scale(0); }
          50% { opacity: 1; transform: scale(1); }
        }

        @keyframes aviation-scan {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(100%); }
        }
      `}</style>
    </div>
  );
}
