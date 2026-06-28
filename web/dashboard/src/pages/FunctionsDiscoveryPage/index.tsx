import { functionsApi } from '@/api/functions';
import { favoritesApi } from '@/api/favorites';
import { registryApi, RegistryFunction } from '@/api/registry';
import { AviationFunctionCard } from '@/components/functions/AviationFunctionCard';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { FunctionConfig } from '@/types';
import { Flame, Loader2, Sparkles, Star, TrendingUp, Trash2, User, Zap } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import {
  PageGrid, Chamber, CornerBrace, TrustSeal,
  SealedButton, FrameButton, StatusPill, AnnotationTag,
} from '@/components/containment';
import { usePageTitle } from '@/hooks';

import './discovery.css';

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
  region?: string;
  code?: string;
  env_vars?: Array<{ key: string; value: string; isSecret?: boolean }>;
  tenant_id?: string;
  version?: string;
  status?: 'draft' | 'deploying' | 'deployed' | 'failed' | 'active' | 'suspended';
}

type FilterType = 'hot' | 'trending' | 'new' | 'popular' | 'favorites' | 'my';

const FILTER_CONFIG: Record<FilterType, { title: string; icon: React.ReactNode; description: string }> = {
  hot: { title: 'Hot Functions', icon: <Flame className="disc-icon-sm disc-icon-orange" />, description: 'Functions trending right now with high activity' },
  trending: { title: 'Trending', icon: <TrendingUp className="disc-icon-sm disc-icon-green" />, description: 'Functions gaining popularity this week' },
  new: { title: 'New Arrivals', icon: <Sparkles className="disc-icon-sm disc-icon-purple" />, description: 'Recently created and published functions' },
  popular: { title: 'Most Popular', icon: <Zap className="disc-icon-sm disc-icon-yellow" />, description: 'Most used functions of all time' },
  favorites: { title: 'Your Favorites', icon: <Star className="disc-icon-sm disc-icon-amber" />, description: 'Functions you have starred' },
  my: { title: 'My Functions', icon: <User className="disc-icon-sm disc-icon-blue" />, description: 'Functions you have created' },
};

function getFilterFromPath(pathname: string): FilterType {
  const lastSegment = pathname.split('/').pop();
  const validFilters: FilterType[] = ['hot', 'trending', 'new', 'popular', 'favorites', 'my'];
  return validFilters.includes(lastSegment as FilterType) ? (lastSegment as FilterType) : 'hot';
}

export function FunctionsDiscoveryPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [functions, setFunctions] = useState<FunctionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [functionToDelete, setFunctionToDelete] = useState<FunctionSummary | null>(null);

  const filter = getFilterFromPath(location.pathname);
  const config = FILTER_CONFIG[filter] || FILTER_CONFIG.hot;
  const isMyFunctionsPage = filter === 'my';

  usePageTitle(config.title);

  const deleteMutation = useMutation({
    mutationFn: (id: string) => functionsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['functions'] });
      toast.success('Function deleted successfully');
      setDeleteDialogOpen(false);
      setFunctionToDelete(null);
      loadFunctions();
    },
    onError: () => { toast.error('Failed to delete function'); setDeleteDialogOpen(false); },
  });

  useEffect(() => { loadFunctions(); }, [filter]);

  const loadFunctions = async () => {
    setLoading(true);
    try {
      let data: FunctionSummary[] = [];
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

      const userFunctions: FunctionSummary[] = userFunctionsRaw.map((f) => ({
        id: f.id, name: f.name, description: f.code?.substring(0, 100) || 'No description',
        runtime: f.version || 'unknown', created_at: f.createdAt, updated_at: f.updatedAt,
        execution_count: 0, author: undefined, isPublic: false,
      }));

      const mappedPublicFunctions: FunctionSummary[] = publicFunctions.map((f) => ({
        id: f.id, name: `${f.author}/${f.name}`, description: f.description,
        runtime: f.latest_version?.split('@')[0] || 'unknown', created_at: f.created_at, updated_at: f.created_at,
        execution_count: Math.floor(f.popularity_score * 100), author: f.author, isPublic: true,
      }));

      const mappedMyRegistryFunctions: FunctionSummary[] = myRegistryFunctions.map((f) => ({
        id: f.id, name: `${f.author}/${f.name}`, description: f.description,
        runtime: f.latest_version?.split('@')[0] || 'unknown', created_at: f.created_at, updated_at: f.created_at,
        execution_count: Math.floor(f.popularity_score * 100), author: f.author, isPublic: f.visibility === 'public',
      }));

      const favoriteIds = new Set(userFavorites.map((fav) => fav.function_id));
      const userFunctionNames = new Set(userFunctions.map((f) => f.name));
      const uniquePublicFunctions = mappedPublicFunctions.filter((f) => !userFunctionNames.has(f.name.split('/')[1] || f.name));
      const allFunctions = [...userFunctions, ...uniquePublicFunctions];

      switch (filter) {
        case 'hot': case 'trending':
          data = allFunctions.sort((a, b) => (b.execution_count || 0) - (a.execution_count || 0)).slice(0, 50); break;
        case 'new':
          data = allFunctions.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).slice(0, 50); break;
        case 'popular':
          data = allFunctions.sort((a, b) => (b.execution_count || 0) - (a.execution_count || 0)).slice(0, 50); break;
        case 'favorites':
          data = allFunctions.filter((f) => favoriteIds.has(f.id)); break;
        case 'my': {
          const seenNames = new Set(userFunctions.map((f) => f.name.toLowerCase()));
          const merged = [...userFunctions];
          for (const fn of mappedMyRegistryFunctions) {
            const baseName = (fn.name.split('/')[1] || fn.name).toLowerCase();
            if (!seenNames.has(baseName)) { seenNames.add(baseName); merged.push(fn); }
          }
          data = merged; break;
        }
        default: data = allFunctions;
      }
      setFunctions(data);
    } catch { toast.error('Failed to load functions'); } finally { setLoading(false); }
  };

  const handleCreateFunction = () => navigate('/functions/new');
  const handleDeleteClick = (fn: FunctionConfig) => { setFunctionToDelete({ ...fn, created_at: fn.createdAt, updated_at: fn.updatedAt }); setDeleteDialogOpen(true); };
  const handleConfirmDelete = () => { if (functionToDelete) deleteMutation.mutate(functionToDelete.id); };

  return (
    <div className="disc-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="disc-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary={`MODULE FN-${filter.toUpperCase()}`} secondary={config.title} position="top-right" />

        <div className="disc-hero__header">
          <div className="disc-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="disc-hero__title">{config.title}</h1>
          </div>
          <p className="disc-hero__subtitle">{config.description}</p>
          <div className="disc-hero__actions">
            <SealedButton onClick={handleCreateFunction}>Create Function</SealedButton>
          </div>
        </div>
      </Chamber>

      {/* Filter Tabs */}
      <div className="disc-filters">
        {(Object.keys(FILTER_CONFIG) as FilterType[]).map((f) => (
          <button key={f} className={`disc-filter-btn ${filter === f ? 'disc-filter-btn--active' : ''}`}
            onClick={() => navigate(`/functions/discovery/${f}`)}>
            {FILTER_CONFIG[f].icon}
            {FILTER_CONFIG[f].title}
          </button>
        ))}
      </div>

      {/* Content */}
      {loading ? (
        <div className="disc-grid">
          {Array.from({ length: 6 }).map((_, i) => <div key={i} className="disc-skeleton" />)}
        </div>
      ) : functions.length === 0 ? (
        <Chamber className="disc-empty">
          <div className="disc-empty__icon">{config.icon}</div>
          <h2 className="disc-empty__title">No functions found</h2>
          <p className="disc-empty__desc">No {filter} functions available right now. Be the first to create one!</p>
          <SealedButton onClick={handleCreateFunction}>Create Function</SealedButton>
        </Chamber>
      ) : (
        <div className="disc-grid">
          {functions.map((fn, index) => (
            <AviationFunctionCard
              key={fn.id} index={index}
              fn={{
                id: fn.id, name: fn.name, providers: ['registry'], region: fn.region || 'auto',
                code: fn.code || '', envVars: (fn.env_vars || []).map(ev => ({ ...ev, isSecret: ev.isSecret ?? false })),
                tenantId: fn.tenant_id || '', createdAt: fn.created_at, updatedAt: fn.updated_at,
                version: fn.version || '1.0', status: fn.status || 'draft',
                executionCount: fn.execution_count || 0, avgDurationMs: fn.avg_duration_ms,
              }}
              onView={(id) => { if (fn.author) { navigate(`/fx/${fn.author}/${fn.name.split('/')[1] || fn.name}`); } else { navigate(`/functions/${id}`); } }}
              onEdit={(id) => { if (fn.author) { navigate(`/fx/${fn.author}/${fn.name.split('/')[1] || fn.name}`); } else { navigate(`/functions/${id}/edit`); } }}
              onDelete={isMyFunctionsPage ? handleDeleteClick : undefined}
            />
          ))}
        </div>
      )}

      {/* Delete Dialog */}
      {deleteDialogOpen && (
        <div className="disc-modal-overlay" onClick={() => setDeleteDialogOpen(false)}>
          <div className="disc-modal" onClick={(e) => e.stopPropagation()}>
            <div className="disc-modal__header">
              <Trash2 className="disc-icon-sm disc-icon-danger" />
              <h2 className="disc-modal__title">Delete Function</h2>
            </div>
            <p className="disc-modal__desc">
              Are you sure you want to delete "{functionToDelete?.name}"? This action cannot be undone.
            </p>
            <div className="disc-modal__actions">
              <FrameButton onClick={() => setDeleteDialogOpen(false)}>Cancel</FrameButton>
              <SealedButton onClick={handleConfirmDelete} disabled={deleteMutation.isPending} loading={deleteMutation.isPending}>
                Delete
              </SealedButton>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default FunctionsDiscoveryPage;
