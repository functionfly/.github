import { functionsApi } from '@/api/functions';
import { favoritesApi } from '@/api/favorites';
import { registryApi, RegistryFunction } from '@/api/registry';
import { AviationEmptyState } from '@/components/functions/AviationEmptyState';
import { AviationFunctionCard } from '@/components/functions/AviationFunctionCard';
import { PageHeader } from '@/components/layout/PageHeader';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { FunctionConfig } from '@/types';
import { Flame, Loader2, Sparkles, Star, TrendingUp, Trash2, User, Zap } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

interface FunctionSummary {
  id: string;
  name: string;
  description?: string;
  runtime?: string;
  created_at: string;
  updated_at: string;
  execution_count?: number;
  avg_duration_ms?: number;
  author?: string;
  isPublic?: boolean;
  // Optional fields for FunctionConfig compatibility
  region?: string;
  code?: string;
  env_vars?: Array<{ key: string; value: string; isSecret?: boolean }>;
  tenant_id?: string;
  version?: string;
  status?: 'draft' | 'deploying' | 'deployed' | 'failed' | 'active' | 'suspended';
}

type FilterType = 'hot' | 'trending' | 'new' | 'popular' | 'favorites' | 'my';

const FILTER_CONFIG: Record<FilterType, { title: string; icon: React.ReactNode; description: string }> = {
  hot: {
    title: 'Hot Functions',
    icon: <Flame className="w-5 h-5 text-orange-500" />,
    description: 'Functions trending right now with high activity',
  },
  trending: {
    title: 'Trending',
    icon: <TrendingUp className="w-5 h-5 text-green-500" />,
    description: 'Functions gaining popularity this week',
  },
  new: {
    title: 'New Arrivals',
    icon: <Sparkles className="w-5 h-5 text-purple-500" />,
    description: 'Recently created and published functions',
  },
  popular: {
    title: 'Most Popular',
    icon: <Zap className="w-5 h-5 text-yellow-500" />,
    description: 'Most used functions of all time',
  },
  favorites: {
    title: 'Your Favorites',
    icon: <Star className="w-5 h-5 text-amber-500" />,
    description: 'Functions you have starred',
  },
  my: {
    title: 'My Functions',
    icon: <User className="w-5 h-5 text-blue-500" />,
    description: 'Functions you have created',
  },
};

function getFilterFromPath(pathname: string): FilterType {
  const pathParts = pathname.split('/');
  const lastSegment = pathParts[pathParts.length - 1];
  const validFilters: FilterType[] = ['hot', 'trending', 'new', 'popular', 'favorites', 'my'];
  if (validFilters.includes(lastSegment as FilterType)) {
    return lastSegment as FilterType;
  }
  return 'hot';
}

export function FunctionsDiscoveryPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [functions, setFunctions] = useState<FunctionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [functionToDelete, setFunctionToDelete] = useState<FunctionSummary | null>(null);

  const deleteMutation = useMutation({
    mutationFn: (id: string) => functionsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['functions'] });
      toast.success('Function deleted successfully');
      setDeleteDialogOpen(false);
      setFunctionToDelete(null);
      loadFunctions();
    },
    onError: () => {
      toast.error('Failed to delete function');
      setDeleteDialogOpen(false);
    },
  });

  const filter = getFilterFromPath(location.pathname);
  const config = FILTER_CONFIG[filter] || FILTER_CONFIG.hot;

  useEffect(() => {
    loadFunctions();
  }, [filter]);

  const loadFunctions = async () => {
    setLoading(true);
    try {
      let data: FunctionSummary[] = [];

      // Fetch user's functions and public registry functions (except for 'favorites' and 'my' which have different sources)
      const needsRegistry = filter !== 'my' && filter !== 'favorites';
      const [userRes, registryRes, favoritesRes, myRegistryRes] = await Promise.allSettled([
        functionsApi.list(),
        needsRegistry ? registryApi.getFunctions({ visibility: 'public', limit: 100 }) : Promise.resolve({ functions: [] }),
        filter === 'favorites' ? favoritesApi.list(1, 50) : Promise.resolve({ favorites: [], total: 0 }),
        filter === 'my' ? registryApi.getMyFunctions({ limit: 100 }) : Promise.resolve({ functions: [] }),
      ]);

      const userFunctionsRaw = userRes.status === 'fulfilled' ? (userRes.value.functions || []) : [];
      const publicFunctions: RegistryFunction[] = registryRes.status === 'fulfilled' ? (registryRes.value.functions || []) : [];
      const userFavorites = favoritesRes.status === 'fulfilled' ? (favoritesRes.value.favorites || []) : [];
      const myRegistryFunctions: RegistryFunction[] = myRegistryRes.status === 'fulfilled' ? (myRegistryRes.value.functions || []) : [];

      // Convert user functions to FunctionSummary format (camelCase -> snake_case)
      const userFunctions: FunctionSummary[] = userFunctionsRaw.map((f) => ({
        id: f.id,
        name: f.name,
        description: f.code?.substring(0, 100) || 'No description',
        runtime: f.version || 'unknown',
        created_at: f.createdAt,
        updated_at: f.updatedAt,
        execution_count: 0,
        author: undefined,
        isPublic: false,
      }));

      // Convert registry functions to FunctionSummary format
      const mappedPublicFunctions: FunctionSummary[] = publicFunctions.map((f) => ({
        id: f.id,
        name: `${f.author}/${f.name}`,
        description: f.description,
        runtime: f.latest_version?.split('@')[0] || 'unknown',
        created_at: f.created_at,
        updated_at: f.created_at,
        execution_count: Math.floor(f.popularity_score * 100),
        author: f.author,
        isPublic: true,
      }));

      // Convert user's own registry functions (from GitHub import etc.)
      const mappedMyRegistryFunctions: FunctionSummary[] = myRegistryFunctions.map((f) => ({
        id: f.id,
        name: `${f.author}/${f.name}`,
        description: f.description,
        runtime: f.latest_version?.split('@')[0] || 'unknown',
        created_at: f.created_at,
        updated_at: f.created_at,
        execution_count: Math.floor(f.popularity_score * 100),
        author: f.author,
        isPublic: f.visibility === 'public',
      }));

      // Build a map of favorite function IDs for quick lookup
      const favoriteIds = new Set(userFavorites.map((fav) => fav.function_id));

      // Combine both sources (exclude duplicates based on name)
      const userFunctionNames = new Set(userFunctions.map((f) => f.name));
      const uniquePublicFunctions = mappedPublicFunctions.filter((f) => !userFunctionNames.has(f.name.split('/')[1] || f.name));

      const allFunctions = [...userFunctions, ...uniquePublicFunctions];

      switch (filter) {
        case 'hot':
        case 'trending':
          data = allFunctions.sort((a, b) => (b.execution_count || 0) - (a.execution_count || 0)).slice(0, 50);
          break;
        case 'new':
          data = allFunctions.sort(
            (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
          ).slice(0, 50);
          break;
        case 'popular':
          data = allFunctions.sort((a, b) => (b.execution_count || 0) - (a.execution_count || 0)).slice(0, 50);
          break;
        case 'favorites':
          data = allFunctions.filter((f) => favoriteIds.has(f.id));
          break;
        case 'my': {
          // Merge platform functions with user's registry functions (deduplicate by name)
          // Registry functions use "author/name" format, platform functions use just "name"
          const seenNames = new Set(userFunctions.map((f) => f.name.toLowerCase()));
          const merged = [...userFunctions];
          for (const fn of mappedMyRegistryFunctions) {
            // Extract the base name from "author/name" format
            const baseName = (fn.name.split('/')[1] || fn.name).toLowerCase();
            if (!seenNames.has(baseName)) {
              seenNames.add(baseName);
              merged.push(fn);
            }
          }
          data = merged;
          break;
        }
        default:
          data = allFunctions;
      }

      setFunctions(data);
    } catch (error) {
      toast.error('Failed to load functions');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateFunction = () => {
    navigate('/functions/new');
  };

  const handleDeleteClick = (fn: FunctionConfig) => {
    const summary: FunctionSummary = {
      ...fn,
      created_at: fn.createdAt,
      updated_at: fn.updatedAt,
    };
    setFunctionToDelete(summary);
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

  const isMyFunctionsPage = filter === 'my';

  const isNewDiscoveryPage = location.pathname === '/functions/discovery/new';

  return (
    <div className={isNewDiscoveryPage ? 'discovery-new-page' : 'discovery-page'}>
      <div className="discovery-page-content">
        <header className={isNewDiscoveryPage ? 'discovery-new-header' : 'discovery-header'}>
          <div className="flex items-center justify-between">
            <div>
              <h1 className={isNewDiscoveryPage ? 'discovery-new-title' : 'discovery-header-title'}>
                <span className="icon-wrapper">{config.icon}</span>
                {config.title}
                {isNewDiscoveryPage && filter === 'new' && (
                  <span className="discovery-new-badge">Fresh</span>
                )}
              </h1>
              <p className={isNewDiscoveryPage ? 'discovery-new-subtitle' : 'discovery-header-subtitle'}>
                {config.description}
              </p>
            </div>
            <Button
              onClick={handleCreateFunction}
              className="bg-aviation-amber hover:bg-aviation-amber-glow text-aviation-bg-primary"
            >
              Create Function
            </Button>
          </div>
        </header>

        <nav className={isNewDiscoveryPage ? 'discovery-new-filters' : 'discovery-filters'}>
          {(Object.keys(FILTER_CONFIG) as FilterType[]).map((f) => (
            <button
              key={f}
              className={`${isNewDiscoveryPage ? 'discovery-new-filter-btn' : 'discovery-filter-btn'} ${filter === f ? 'active' : ''}`}
              onClick={() => navigate(`/functions/discovery/${f}`)}
            >
              {FILTER_CONFIG[f].icon}
              {FILTER_CONFIG[f].title}
            </button>
          ))}
        </nav>

        <main className="discovery-page-content">
          {loading ? (
            <div className={isNewDiscoveryPage ? 'discovery-new-loading' : 'discovery-loading'}>
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className={isNewDiscoveryPage ? 'discovery-new-skeleton' : 'discovery-skeleton'} />
              ))}
            </div>
          ) : functions.length === 0 ? (
            <div className={isNewDiscoveryPage ? 'discovery-new-empty' : 'discovery-empty'}>
              <div className={isNewDiscoveryPage ? 'discovery-new-empty-icon' : 'discovery-empty-icon'}>{config.icon}</div>
              <h2 className={isNewDiscoveryPage ? 'discovery-new-empty-title' : 'discovery-empty-title'}>
                No functions found
              </h2>
              <p className={isNewDiscoveryPage ? 'discovery-new-empty-description' : 'discovery-empty-description'}>
                No {filter} functions available right now. Be the first to create one!
              </p>
              <Button
                onClick={handleCreateFunction}
                className={isNewDiscoveryPage ? 'discovery-new-empty-btn' : ''}
                variant={isNewDiscoveryPage ? undefined : 'default'}
              >
                Create Function
              </Button>
            </div>
          ) : (
            <div className={isNewDiscoveryPage ? 'discovery-new-grid' : 'discovery-grid'}>
              {functions.map((fn, index) => (
              <AviationFunctionCard
                key={fn.id}
                index={index}
                isNewStyle={isNewDiscoveryPage}
                fn={{
                  id: fn.id,
                  name: fn.name,
                  providers: ['registry'],
                  region: fn.region || 'auto',
                  code: fn.code || '',
                  envVars: (fn.env_vars || []).map(ev => ({ ...ev, isSecret: ev.isSecret ?? false })),
                  tenantId: fn.tenant_id || '',
                  createdAt: fn.created_at,
                  updatedAt: fn.updated_at,
                  version: fn.version || '1.0',
                  status: fn.status || 'draft',
                  executionCount: fn.execution_count || 0,
                  avgDurationMs: fn.avg_duration_ms,
                }}
                onView={(id) => {
                  if (fn.author) {
                    // Navigate to registry function (public or private)
                    const funcName = fn.name.split('/')[1] || fn.name;
                    navigate(`/fx/${fn.author}/${funcName}`);
                  } else {
                    // Navigate to user's platform function
                    navigate(`/functions/${id}`);
                  }
                }}
                onEdit={(id) => {
                  if (fn.author) {
                    const funcName = fn.name.split('/')[1] || fn.name;
                    navigate(`/fx/${fn.author}/${funcName}`);
                  } else {
                    navigate(`/functions/${id}/edit`);
                  }
                }}
                onDelete={isMyFunctionsPage ? handleDeleteClick : undefined}
              />
            ))}
          </div>
          )}
        </main>
      </div>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent
          style={{
            background: 'var(--color-aviation-bg-secondary)',
            borderColor: 'var(--color-aviation-red-dim, rgba(239,68,68,0.3))',
          }}
        >
          <DialogHeader>
            <DialogTitle
              className="flex items-center gap-2"
              style={{ color: 'var(--color-aviation-text-primary)' }}
            >
              <Trash2 className="w-5 h-5" style={{ color: 'var(--color-aviation-red)' }} />
              Delete Function
            </DialogTitle>
            <DialogDescription
              style={{ color: 'var(--color-aviation-text-secondary)' }}
            >
              Are you sure you want to delete "{functionToDelete?.name}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <DialogFooter className="gap-2">
            <Button
              variant="outline"
              onClick={handleCancelDelete}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleteMutation.isPending}
              style={{
                background: 'var(--color-aviation-red)',
                color: 'white',
              }}
            >
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                <>
                  <Trash2 className="w-4 h-4" />
                  Delete
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
