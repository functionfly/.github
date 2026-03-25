import { functionsApi } from '@/api/functions';
import { AviationEmptyState } from '@/components/functions/AviationEmptyState';
import { AviationFunctionCard } from '@/components/functions/AviationFunctionCard';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import type { FunctionConfig } from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  AlertTriangle,
  ArrowUpRight,
  Filter,
  LayoutGrid,
  List,
  Loader2,
  Plus,
  Radar,
  Search,
  Zap,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

/**
 * Functions Page - Aviation-themed cockpit interface
 * Theme-aware: Industrial instrumentation aesthetic for both light/dark modes
 */
export function FunctionsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [functionToDelete, setFunctionToDelete] = useState<FunctionConfig | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [mounted, setMounted] = useState(false);

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
      toast.success('Function terminated successfully');
      setDeleteDialogOpen(false);
      setFunctionToDelete(null);
    },
    onError: () => {
      toast.error('Failed to terminate function');
      setDeleteDialogOpen(false);
    },
  });

  const functions = data?.functions ?? [];

  const filteredFunctions = functions.filter(
    (fn) =>
      fn.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      fn.id.toLowerCase().includes(searchQuery.toLowerCase())
  );

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
            SYSTEM ERROR
          </h3>
          <p
            className="font-mono text-sm mb-6"
            style={{ color: 'var(--color-aviation-text-secondary)' }}
          >
            Failed to load function registry. Check connection and retry.
          </p>
          <Button
            onClick={() => queryClient.invalidateQueries({ queryKey: ['functions'] })}
            className="aviation-button gap-2"
          >
            <Radar className="w-4 h-4" />
            RETRY SCAN
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
                  <span className="aviation-label">FLEET MANAGEMENT //</span>
                  <span className="aviation-value-cyan text-xs">ACTIVE</span>
                </div>
                <h1
                  className="text-2xl md:text-3xl font-bold font-mono tracking-tight uppercase"
                  style={{ color: 'var(--color-aviation-text-primary)' }}
                >
                  Functions
                </h1>
                <p
                  className="text-sm font-mono mt-1"
                  style={{ color: 'var(--color-aviation-text-secondary)' }}
                >
                  Deploy and manage edge functions across distributed nodes
                </p>
              </div>
            </div>

            {/* Stats row */}
            <div className="flex items-center gap-6 lg:gap-8">
              <div className="text-right">
                <div className="aviation-label mb-0.5">TOTAL UNITS</div>
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
                <div className="aviation-label mb-0.5">ONLINE</div>
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
                DEPLOY NEW
              </Button>
            </div>
          </div>

          {/* Mobile CTA */}
          <Button
            onClick={() => navigate('/functions/new')}
            className="aviation-button aviation-button-primary gap-2 w-full mt-4 md:hidden"
          >
            <Plus className="w-4 h-4" />
            DEPLOY NEW FUNCTION
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
              className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4"
              style={{ color: 'var(--color-aviation-text-muted)' }}
            />
            <input
              type="text"
              placeholder="SEARCH FUNCTIONS BY NAME OR ID..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="aviation-input w-full pl-14 pr-4"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-4 top-1/2 -translate-y-1/2 transition-colors font-mono text-xs"
                style={{ color: 'var(--color-aviation-text-muted)' }}
              >
                CLEAR
              </button>
            )}
          </div>

          {/* Controls */}
          <div className="flex items-center gap-2">
            {/* View toggle */}
            <div className="aviation-panel flex items-center rounded-lg p-1">
              <button
                onClick={() => setViewMode('grid')}
                className={`p-2 rounded-md transition-all ${viewMode === 'grid' ? '' : ''}`}
                style={{
                  background:
                    viewMode === 'grid' ? 'var(--color-aviation-amber-subtle)' : 'transparent',
                  color:
                    viewMode === 'grid'
                      ? 'var(--color-aviation-amber)'
                      : 'var(--color-aviation-text-muted)',
                }}
              >
                <LayoutGrid className="w-4 h-4" />
              </button>
              <button
                onClick={() => setViewMode('list')}
                className="p-2 rounded-md transition-all"
                style={{
                  background:
                    viewMode === 'list' ? 'var(--color-aviation-amber-subtle)' : 'transparent',
                  color:
                    viewMode === 'list'
                      ? 'var(--color-aviation-amber)'
                      : 'var(--color-aviation-text-muted)',
                }}
              >
                <List className="w-4 h-4" />
              </button>
            </div>

            {/* Filter button */}
            <Button variant="outline" className="aviation-button gap-2">
              <Filter className="w-4 h-4" />
              FILTER
            </Button>
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
                INITIALIZING SYSTEMS...
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
                    <span className="aviation-label">SCAN RESULTS:</span>
                    <span
                      className="font-mono text-sm font-semibold"
                      style={{ color: 'var(--color-aviation-text-primary)' }}
                    >
                      {filteredFunctions.length} UNIT{filteredFunctions.length !== 1 ? 'S' : ''}{' '}
                      DETECTED
                    </span>
                  </div>

                  {searchQuery && (
                    <button
                      onClick={() => setSearchQuery('')}
                      className="font-mono text-xs flex items-center gap-1 transition-colors hover:opacity-80"
                      style={{ color: 'var(--color-aviation-cyan)' }}
                    >
                      CLEAR SCAN
                      <ArrowUpRight className="w-3 h-3" />
                    </button>
                  )}
                </div>

                {/* Functions Grid */}
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
              CONFIRM TERMINATION
            </DialogTitle>
            <DialogDescription
              className="font-mono text-sm"
              style={{ color: 'var(--color-aviation-text-secondary)' }}
            >
              You are about to terminate function &ldquo;{functionToDelete?.name}&rdquo;. This
              action is irreversible and will immediately cease all operations.
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
                WARNING:
              </span>
            </div>
            <ul
              className="text-xs font-mono space-y-1 ml-4"
              style={{ color: 'var(--color-aviation-text-secondary)' }}
            >
              <li>• All active invocations will be interrupted</li>
              <li>• Function will be removed from all providers</li>
              <li>• Associated logs will be archived after 30 days</li>
            </ul>
          </div>

          <DialogFooter className="gap-2">
            <Button
              variant="outline"
              onClick={handleCancelDelete}
              className="aviation-button gap-2"
            >
              ABORT
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
                  TERMINATING...
                </>
              ) : (
                <>
                  <Zap className="w-4 h-4" />
                  CONFIRM TERMINATE
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
