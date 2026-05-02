import { registryApi, type RegistryFunction, type RegistrySearchParams } from '@/api/registry';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useRealtime } from '@/hooks/useRealtime';
import { useAuthStore } from '@/stores/authStore';
import {
  Activity,
  Boxes,
  Brain,
  ChevronRight,
  Code,
  CreditCard,
  Database,
  Download,
  FileJson,
  FileText,
  Filter,
  GitBranch,
  Globe,
  Grid,
  Image,
  Layers,
  List,
  LogIn,
  Mail,
  Play,
  Search,
  Shield,
  SortAsc,
  SortDesc,
  Sparkles,
  Star,
  TrendingUp,
  Wrench,
  X,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

type BrowseFunctionsViewVariant = 'public' | 'dashboard';

/* ────────────────────────────────────────────────────────────── */
/*  Constants                                                      */
/* ────────────────────────────────────────────────────────────── */

const CATEGORY_ALIASES: Record<string, string> = {
  ai: 'AI & ML',
  'ai-ml': 'AI & ML',
  ml: 'AI & ML',
  machinelearning: 'AI & ML',
  analytics: 'Analytics',
  community_pain_point: 'Community',
  encoding: 'Encoding',
  devops: 'DevOps',
  ecommerce: 'Ecommerce',
  finance: 'Finance',
  format: 'Format',
  network: 'Network',
  social: 'Social',
  crypto: 'Security',
  arrays: 'Arrays',
  datetime: 'datetime',
  http: 'http',
  math: 'math',
  media: 'media',
  security: 'Security',
  textprocessing: 'text-processing',
  utilities: 'Utility',
};

const PREFERRED_CATEGORY_ORDER: string[] = [
  'AI & ML',
  'Utility',
  'Encoding',
  'DevOps',
  'Security',
  'Ecommerce',
  'Finance',
  'Format',
  'Analytics',
  'Community',
  'Network',
  'Social',
  'datetime',
  'http',
  'math',
  'media',
  'text-processing',
  'arrays',
];

const SORT_OPTIONS = [
  { value: 'popularity', label: 'Popularity' },
  { value: 'rating', label: 'Rating' },
  { value: 'reliability', label: 'Reliability' },
  { value: 'trust_score', label: 'Trust Score' },
  { value: 'newest', label: 'Newest' },
  { value: 'name', label: 'Name' },
];

/** Trust tier filter options */
const TRUST_TIER_OPTIONS = [
  { value: 'all', label: 'All Trust Levels' },
  { value: 'highly_trusted', label: 'Highly Trusted' },
  { value: 'verified', label: 'Verified' },
  { value: 'basic', label: 'Basic' },
  { value: 'unverified', label: 'Unverified' },
];

/** Minimum trust score filter options */
const MIN_TRUST_SCORE_OPTIONS = [
  { value: 0, label: 'Any Score' },
  { value: 50, label: '50%+' },
  { value: 70, label: '70%+' },
  { value: 80, label: '80%+' },
  { value: 90, label: '90%+' },
];

/** Map each category to a lucide icon + a CSS color class */
const CATEGORY_META: Record<string, { icon: React.ElementType; colorClass: string }> = {
  'All Categories': { icon: Boxes, colorClass: 'registry-cat-color-indigo' },
  api: { icon: Zap, colorClass: 'registry-cat-color-indigo' },
  'API Tools': { icon: Zap, colorClass: 'registry-cat-color-indigo' },
  arrays: { icon: Layers, colorClass: 'registry-cat-color-slate' },
  automation: { icon: GitBranch, colorClass: 'registry-cat-color-orange' },
  Authentication: { icon: Shield, colorClass: 'registry-cat-color-violet' },
  crypto: { icon: Shield, colorClass: 'registry-cat-color-red' },
  'Data Format': { icon: FileJson, colorClass: 'registry-cat-color-teal' },
  'data-formatting': { icon: FileJson, colorClass: 'registry-cat-color-teal' },
  Database: { icon: Database, colorClass: 'registry-cat-color-sky' },
  datetime: { icon: Activity, colorClass: 'registry-cat-color-amber' },
  Email: { icon: Mail, colorClass: 'registry-cat-color-amber' },
  encoding: { icon: Code, colorClass: 'registry-cat-color-gray' },
  'File Processing': { icon: FileText, colorClass: 'registry-cat-color-teal' },
  formatting: { icon: FileText, colorClass: 'registry-cat-color-blue' },
  'getting-started': { icon: Sparkles, colorClass: 'registry-cat-color-yellow' },
  http: { icon: Globe, colorClass: 'registry-cat-color-green' },
  'Image Processing': { icon: Image, colorClass: 'registry-cat-color-fuchsia' },
  integrations: { icon: Boxes, colorClass: 'registry-cat-color-purple' },
  'Machine Learning': { icon: Brain, colorClass: 'registry-cat-color-violet' },
  math: { icon: TrendingUp, colorClass: 'registry-cat-color-cyan' },
  media: { icon: Image, colorClass: 'registry-cat-color-fuchsia' },
  Payment: { icon: CreditCard, colorClass: 'registry-cat-color-emerald' },
  security: { icon: Shield, colorClass: 'registry-cat-color-violet' },
  text: { icon: FileText, colorClass: 'registry-cat-color-teal' },
  'text-processing': { icon: FileText, colorClass: 'registry-cat-color-teal' },
  utilities: { icon: Wrench, colorClass: 'registry-cat-color-cyan' },
  Utility: { icon: Wrench, colorClass: 'registry-cat-color-cyan' },
  'Web Scraping': { icon: Globe, colorClass: 'registry-cat-color-rose' },
  Workflow: { icon: GitBranch, colorClass: 'registry-cat-color-orange' },
  'AI & ML': { icon: Brain, colorClass: 'registry-cat-color-violet' },
  'Community': { icon: Sparkles, colorClass: 'registry-cat-color-yellow' },
  'Encoding': { icon: Code, colorClass: 'registry-cat-color-gray' },
  'DevOps': { icon: GitBranch, colorClass: 'registry-cat-color-orange' },
  'Ecommerce': { icon: CreditCard, colorClass: 'registry-cat-color-emerald' },
  'Finance': { icon: CreditCard, colorClass: 'registry-cat-color-teal' },
  'Format': { icon: FileJson, colorClass: 'registry-cat-color-blue' },
  'Network': { icon: Globe, colorClass: 'registry-cat-color-green' },
  'Social': { icon: Activity, colorClass: 'registry-cat-color-pink' },
  'Analytics': { icon: TrendingUp, colorClass: 'registry-cat-color-cyan' },
  'Arrays': { icon: Layers, colorClass: 'registry-cat-color-slate' },
};

/* ────────────────────────────────────────────────────────────── */
/*  Helpers                                                        */
/* ────────────────────────────────────────────────────────────── */

function formatScore(score: number | undefined | null): string {
  return typeof score === 'number' && !Number.isNaN(score) ? score.toFixed(1) : '0.0';
}

function reliabilityLabel(score: number): { label: string; cls: string } {
  if (score >= 0.9) return { label: 'High', cls: 'registry-reliability-high' };
  if (score >= 0.6) return { label: 'Med', cls: 'registry-reliability-mid' };
  return { label: 'Low', cls: 'registry-reliability-low' };
}

function getColorClass(category: string | null | undefined): string {
  if (!category) return 'registry-cat-color-indigo';
  return CATEGORY_META[category]?.colorClass ?? 'registry-cat-color-indigo';
}

function normalizeCategory(raw: string | null | undefined): string {
  const s = String(raw ?? '').trim();
  if (!s) return '';
  const key = s
    .toLowerCase()
    .replace(/[_\s]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
  return CATEGORY_ALIASES[key] ?? s;
}

function getCategoryList(functions: RegistryFunction[]): string[] {
  const unique = new Set<string>();
  for (const fn of functions) {
    if (fn.category != null) {
      const c = normalizeCategory(fn.category);
      if (c) unique.add(c);
    }
  }

  const present = Array.from(unique);
  const preferred = PREFERRED_CATEGORY_ORDER.filter((c) => unique.has(c));
  const remaining = present
    .filter((c) => !preferred.includes(c))
    .sort((a, b) => a.localeCompare(b));

  return ['All Categories', ...preferred, ...remaining];
}

/* ────────────────────────────────────────────────────────────── */
/*  Sub-components                                                 */
/* ────────────────────────────────────────────────────────────── */

function StatItem({
  value,
  label,
  icon: Icon,
}: {
  value: string | number;
  label: string;
  icon: React.ElementType;
}) {
  const IconComponent = Icon as React.ComponentType<{ className?: string }>;
  return (
    <div className="registry-stat-item">
      <div className="flex items-center gap-1.5">
        <IconComponent className="h-3.5 w-3.5 text-brand-500 opacity-70 shrink-0" />
        <span className="registry-stat-value">{value}</span>
      </div>
      <span className="registry-stat-label">{label}</span>
    </div>
  );
}

function CategoryItem({
  category,
  count,
  active,
  onClick,
}: {
  category: string;
  count: number;
  active: boolean;
  onClick: () => void;
}) {
  const meta = CATEGORY_META[category] ?? { icon: Code, colorClass: 'registry-cat-color-indigo' };
  const Icon = meta.icon;
  const IconComponent = Icon as React.ComponentType<{ className?: string }>;
  return (
    <button
      type="button"
      className={`registry-cat-item ${meta.colorClass} ${active ? 'active' : ''}`}
      onClick={onClick}
    >
      <span className="registry-cat-icon">
        <IconComponent className="h-3.5 w-3.5" />
      </span>
      <span className="registry-cat-label">{category}</span>
      <span className="registry-cat-count">{count}</span>
    </button>
  );
}

function MobileChip({
  category,
  active,
  onClick,
}: {
  category: string;
  active: boolean;
  onClick: () => void;
}) {
  const meta = CATEGORY_META[category] ?? { icon: Code, colorClass: 'registry-cat-color-indigo' };
  const Icon = meta.icon;
  const IconComponent = Icon as React.ComponentType<{ className?: string }>;
  return (
    <button
      type="button"
      className={`registry-mobile-chip ${active ? 'active' : ''}`}
      onClick={onClick}
    >
      <IconComponent className="h-3 w-3 shrink-0" />
      {category === 'All Categories' ? 'All' : category}
    </button>
  );
}

function GridCard({
  fn,
  isAuthenticated,
  onView,
  onDeploy,
  onTry,
  flashId,
}: {
  fn: RegistryFunction;
  isAuthenticated: boolean;
  onView: () => void;
  onDeploy: () => void;
  onTry: () => void;
  flashId: string | null;
}) {
  const colorClass = getColorClass(fn.category);
  const rating = Number(fn.overall_score ?? 0);
  const ratingPct = Math.min(100, (rating / 5) * 100);
  const reliability = fn.reliability_score ?? 0;
  const rel = reliabilityLabel(reliability);
  const isFlashing = flashId === fn.id;

  return (
    <div className={`registry-card ${isFlashing ? 'registry-card-flash' : ''}`}>
      <div className="registry-card-accent" />
      <div className="p-4 flex flex-col h-full gap-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2.5 min-w-0">
            <div
              className={`registry-card-icon ${colorClass}`}
              style={{
                background: 'var(--cat-bg)',
                border: '1px solid color-mix(in srgb, var(--cat-color) 25%, transparent)',
              }}
            >
              <Code className="h-5 w-5" style={{ color: 'var(--cat-color)' }} />
            </div>
            <div className="min-w-0">
              <h3 className="text-sm font-bold text-text-primary truncate leading-tight">
                {fn.name}
              </h3>
              <p className="text-xs text-text-muted truncate">@{fn.author}</p>
            </div>
          </div>

          {fn.category && (
            <span className={`registry-cat-badge ${colorClass} shrink-0`}>
              <span className="registry-cat-dot" />
              {fn.category}
            </span>
          )}
        </div>

        <p className="text-xs text-text-secondary line-clamp-2 leading-relaxed flex-grow">
          {fn.description || fn.title || 'No description available.'}
        </p>

        {fn.tags && fn.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {fn.tags.slice(0, 3).map((tag, i) => (
              <span
                key={i}
                className="inline-flex items-center px-1.5 py-0.5 rounded-md text-[0.62rem] font-medium
                           bg-bg-tertiary/60 border border-border-subtle text-text-secondary"
              >
                {tag}
              </span>
            ))}
            {fn.tags.length > 3 && (
              <span
                className="inline-flex items-center px-1.5 py-0.5 rounded-md text-[0.62rem] font-medium
                           bg-bg-tertiary/60 border border-border-subtle text-text-secondary"
              >
                +{fn.tags.length - 3}
              </span>
            )}
          </div>
        )}

        <div className="flex items-center justify-between pt-1 border-t border-border-subtle/70">
          <div className="flex items-center gap-1.5">
            <Star className="h-3.5 w-3.5 fill-amber-400 text-amber-400 shrink-0" />
            <span className="text-xs font-bold text-text-primary">
              {formatScore(fn.overall_score)}
            </span>
            <div className="registry-rating-bar">
              <div className="registry-rating-fill" style={{ width: `${ratingPct}%` }} />
            </div>
            <span className="text-[0.62rem] text-text-muted">({fn.total_ratings ?? 0})</span>
          </div>

          <div className="flex items-center gap-2">
            <span className={`registry-reliability-badge ${rel.cls}`}>
              <Activity className="h-2.5 w-2.5" />
              {rel.label}
            </span>
            <span className="flex items-center gap-1 text-[0.65rem] text-text-muted">
              <Download className="h-3 w-3" />
              {Math.floor(Number(fn.popularity_score) || 0)}
            </span>
            {Number(fn.price_per_call) > 0 && (
              <span className="registry-price-badge">${fn.price_per_call}/call</span>
            )}
          </div>
        </div>

        <div className="flex gap-1.5 mt-auto pt-1">
          <button
            type="button"
            onClick={onView}
            className="flex-1 h-8 rounded-lg text-xs font-semibold border border-border-subtle card-action-btn
                       bg-bg-tertiary/60 text-text-secondary hover:bg-bg-hover hover:text-text-primary
                       transition-all duration-150 flex items-center justify-center gap-1"
          >
            <ChevronRight className="h-3 w-3" />
            View
          </button>
          <button
            type="button"
            onClick={onDeploy}
            className="flex-1 h-8 rounded-lg text-xs font-semibold card-action-btn btn-deploy
                       bg-gradient-to-r from-brand-500 via-purple-500 to-pink-500
                       text-white hover:brightness-110 hover:scale-[1.02]
                       transition-all duration-150 flex items-center justify-center gap-1
                       shadow-lg shadow-brand-500/20"
          >
            {!isAuthenticated && <LogIn className="h-3 w-3" />}
            <Layers className="h-3 w-3" />
            Deploy
          </button>
          <button
            type="button"
            onClick={onTry}
            className="flex-1 h-8 rounded-lg text-xs font-semibold border border-brand-500/25 card-action-btn btn-try
                       bg-brand-500/10 text-brand-400 hover:bg-brand-500/[0.18] hover:border-brand-500/40
                       transition-all duration-150 flex items-center justify-center gap-1"
          >
            {isAuthenticated ? <Play className="h-3 w-3" /> : <LogIn className="h-3 w-3" />}
            Try
          </button>
        </div>
      </div>
    </div>
  );
}

function ListCard({
  fn,
  isAuthenticated,
  onView,
  onDeploy,
  onTry,
  flashId,
}: {
  fn: RegistryFunction;
  isAuthenticated: boolean;
  onView: () => void;
  onDeploy: () => void;
  onTry: () => void;
  flashId: string | null;
}) {
  const colorClass = getColorClass(fn.category);
  const isFlashing = flashId === fn.id;

  return (
    <div className={`registry-list-card ${colorClass} ${isFlashing ? 'registry-card-flash' : ''}`}>
      <div
        className="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
        style={{
          background: 'var(--cat-bg)',
          border: '1px solid color-mix(in srgb, var(--cat-color) 25%, transparent)',
        }}
      >
        <Code className="h-4 w-4" style={{ color: 'var(--cat-color)' }} />
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-sm font-bold text-text-primary truncate">
            {fn.author}/{fn.name}
          </span>
          {fn.category && (
            <span className={`registry-cat-badge ${colorClass}`}>
              <span className="registry-cat-dot" />
              {fn.category}
            </span>
          )}
        </div>
        <p className="text-xs text-text-secondary line-clamp-1 mt-0.5">
          {fn.title || fn.description || 'No description'}
        </p>
      </div>

      <div className="hidden sm:flex items-center gap-4 shrink-0">
        <div className="flex items-center gap-1 text-xs">
          <Star className="h-3.5 w-3.5 fill-amber-400 text-amber-400" />
          <span className="font-bold text-text-primary">{formatScore(fn.overall_score)}</span>
          <span className="text-text-muted">({fn.total_ratings ?? 0})</span>
        </div>
        <div className="flex items-center gap-1 text-xs text-text-muted">
          <Download className="h-3.5 w-3.5" />
          <span>{Math.floor(Number(fn.popularity_score) || 0)}</span>
        </div>
        {Number(fn.price_per_call) > 0 && (
          <span className="registry-price-badge">${fn.price_per_call}/call</span>
        )}
      </div>

      <div className="flex items-center gap-1.5 shrink-0">
        <button
          type="button"
          onClick={onView}
          className="h-7 px-3 rounded-lg text-xs font-semibold border border-border-subtle
                     bg-bg-tertiary/60 text-text-secondary hover:bg-bg-hover hover:text-text-primary
                     transition-all duration-150 flex items-center gap-1"
        >
          View
        </button>
        <button
          type="button"
          onClick={onDeploy}
          className="h-7 px-3 rounded-lg text-xs font-semibold
                     bg-gradient-to-r from-brand-500 to-purple-500 text-white
                     hover:brightness-110 transition-all duration-150 flex items-center gap-1"
        >
          {!isAuthenticated && <LogIn className="h-3 w-3" />}
          Deploy
        </button>
        <button
          type="button"
          onClick={onTry}
          className="h-7 px-3 rounded-lg text-xs font-semibold border border-brand-500/25
                     bg-brand-500/10 text-brand-400 hover:bg-brand-500/[0.18]
                     transition-all duration-150 flex items-center gap-1"
        >
          {isAuthenticated ? <Play className="h-3 w-3" /> : <LogIn className="h-3 w-3" />}
          Try
        </button>
      </div>
    </div>
  );
}

function SkeletonGrid({ count = 6, mode }: { count?: number; mode: 'grid' | 'list' }) {
  if (mode === 'list') {
    return (
      <div className="space-y-2">
        {Array.from({ length: count }).map((_, i) => (
          <div
            key={i}
            className="registry-list-card animate-fade-in-up"
            style={{ animationDelay: `${i * 60}ms`, opacity: 0 }}
          >
            <div className="w-10 h-10 rounded-xl bg-bg-tertiary/60 shrink-0" />
            <div className="flex-1 space-y-2">
              <div className="h-3.5 w-40 rounded-md bg-bg-tertiary/60" />
              <div className="h-2.5 w-64 rounded-md bg-bg-tertiary/40" />
            </div>
            <div className="hidden sm:flex gap-3">
              <div className="h-4 w-16 rounded-md bg-bg-tertiary/60" />
              <div className="h-4 w-12 rounded-md bg-bg-tertiary/40" />
            </div>
            <div className="flex gap-1.5">
              <div className="h-7 w-14 rounded-lg bg-bg-tertiary/60" />
              <div className="h-7 w-16 rounded-lg bg-bg-tertiary/60" />
              <div className="h-7 w-12 rounded-lg bg-bg-tertiary/60" />
            </div>
          </div>
        ))}
      </div>
    );
  }
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="registry-skeleton-card animate-fade-in-up"
          style={{ animationDelay: `${i * 80}ms`, opacity: 0 }}
        />
      ))}
    </div>
  );
}

const PAGE_SIZE = 50;

export function BrowseFunctionsView({ variant }: { variant: BrowseFunctionsViewVariant }) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const authorFromUrl = searchParams.get('author')?.trim() ?? '';
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const { subscribe, unsubscribe } = useRealtime();

  const [functions, setFunctions] = useState<RegistryFunction[]>([]);
  const [filteredFunctions, setFilteredFunctions] = useState<RegistryFunction[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [offset, setOffset] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const [searchQuery, setSearchQuery] = useState(authorFromUrl);
  const [selectedCategory, setSelectedCategory] = useState('All Categories');
  const [sortBy, setSortBy] = useState('popularity');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [showMobileFilters, setShowMobileFilters] = useState(false);
  const [flashId, setFlashId] = useState<string | null>(null);
  const [trustTierFilter, setTrustTierFilter] = useState('all');
  const [minTrustScore, setMinTrustScore] = useState(0);

  // Realtime updates
  useEffect(() => {
    const handleRegistryUpdate = (data: {
      event?: string;
      update_type?: string;
      function_id?: string;
      overall_score?: number;
      total_ratings?: number;
    }) => {
      if (data.event === 'registry_update' && data.update_type === 'rating') {
        setFunctions((prev) =>
          prev.map((fn) =>
            fn.id === data.function_id
              ? {
                  ...fn,
                  overall_score: data.overall_score ?? fn.overall_score,
                  total_ratings: data.total_ratings ?? fn.total_ratings,
                }
              : fn
          )
        );
        if (data.function_id) {
          setFlashId(data.function_id);
          setTimeout(() => setFlashId(null), 700);
        }
      }
    };
    subscribe('registry_updates', handleRegistryUpdate);
    return () => unsubscribe('registry_updates', handleRegistryUpdate);
  }, [subscribe, unsubscribe]);

  useEffect(() => {
    setSearchQuery(authorFromUrl);
  }, [authorFromUrl]);

  // Build API params from current filters
  const getApiParams = useCallback((pageOffset: number) => {
    const params: RegistrySearchParams = {
      limit: PAGE_SIZE,
      offset: pageOffset,
      visibility: 'public',
    };
    if (authorFromUrl) params.author = authorFromUrl;
    if (searchQuery) params.query = searchQuery;
    if (selectedCategory && selectedCategory !== 'All Categories') {
      params.category = selectedCategory.toLowerCase().replace(/\s+/g, '-');
    }
    return params;
  }, [authorFromUrl, searchQuery, selectedCategory]);

  // Initial load
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        setLoading(true);
        setFunctions([]);
        setOffset(0);
        setHasMore(true);
        const res = await registryApi.getFunctions(getApiParams(0));
        if (!cancelled) {
          setFunctions(res.functions ?? []);
          setTotalCount((res as any).total ?? res.functions?.length ?? 0);
          setHasMore((res.functions?.length ?? 0) === PAGE_SIZE);
        }
      } catch (e) {
        console.error('Failed to load registry functions:', e);
        if (!cancelled) setFunctions([]);
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [authorFromUrl, searchQuery, selectedCategory, getApiParams]);

  // Load more handler
  const loadMore = useCallback(async () => {
    if (loadingMore || !hasMore) return;
    try {
      setLoadingMore(true);
      const nextOffset = offset + PAGE_SIZE;
      const res = await registryApi.getFunctions(getApiParams(nextOffset));
      const newFunctions = res.functions ?? [];
      setFunctions((prev) => [...prev, ...newFunctions]);
      setOffset(nextOffset);
      setHasMore(newFunctions.length === PAGE_SIZE);
    } catch (e) {
      console.error('Failed to load more functions:', e);
    } finally {
      setLoadingMore(false);
    }
  }, [offset, hasMore, loadingMore, getApiParams]);

  // Client-side filtering for trust filters (not supported by API)
  useEffect(() => {
    let filtered = functions.filter((fn) => {
      const trustScore = fn.trust_score ?? fn.overall_score;
      const trustTier = fn.trust_tier ?? fn.verification_status ?? 'unverified';
      const matchesTrustTier =
        trustTierFilter === 'all' ||
        (trustTierFilter === 'highly_trusted' && trustTier === 'critical') ||
        (trustTierFilter === 'verified' && (trustTier === 'high' || trustTier === 'verified')) ||
        (trustTierFilter === 'basic' && (trustTier === 'medium' || trustTier === 'low')) ||
        (trustTierFilter === 'unverified' &&
          (trustTier === 'untrusted' || trustTier === 'unverified'));

      const matchesMinTrust = (trustScore ?? 0) >= minTrustScore;

      return matchesTrustTier && matchesMinTrust;
    });

    // Client-side sorting
    filtered.sort((a, b) => {
      let aVal: number | string, bVal: number | string;
      switch (sortBy) {
        case 'popularity':
          aVal = a.popularity_score ?? 0;
          bVal = b.popularity_score ?? 0;
          break;
        case 'rating':
          aVal = a.overall_score ?? 0;
          bVal = b.overall_score ?? 0;
          break;
        case 'reliability':
          aVal = a.reliability_score ?? 0;
          bVal = b.reliability_score ?? 0;
          break;
        case 'trust_score':
          aVal = a.trust_score ?? a.overall_score ?? 0;
          bVal = b.trust_score ?? b.overall_score ?? 0;
          break;
        case 'newest':
          aVal = a.created_at ? new Date(a.created_at).getTime() : 0;
          bVal = b.created_at ? new Date(b.created_at).getTime() : 0;
          break;
        case 'name':
          aVal = (a.name ?? '').toLowerCase();
          bVal = (b.name ?? '').toLowerCase();
          break;
        default:
          return 0;
      }
      if (sortOrder === 'asc') return aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
      return aVal > bVal ? -1 : aVal < bVal ? 1 : 0;
    });

    setFilteredFunctions(filtered);
  }, [functions, sortBy, sortOrder, trustTierFilter, minTrustScore]);

  const categoryCounts = useCallback((): Record<string, number> => {
    const counts: Record<string, number> = { 'All Categories': totalCount || functions.length };
    for (const fn of functions) {
      const c = normalizeCategory(fn.category);
      if (c) counts[c] = (counts[c] ?? 0) + 1;
    }
    return counts;
  }, [functions, totalCount]);
  const counts = categoryCounts();

  const categories = getCategoryList(functions);

  const getLoginRedirect = (path: string) => `/login?redirect=${encodeURIComponent(path)}`;

  const handleDeploy = (fn: RegistryFunction) => {
    const path = `/functions/deploy?registry=${fn.author}/${fn.name}`;
    if (!isAuthenticated) {
      navigate(getLoginRedirect(path));
      return;
    }
    navigate(path);
  };

  const handleTry = (fn: RegistryFunction) => {
    const path = `/run/${fn.author}/${fn.name}`;
    if (!isAuthenticated) {
      navigate(getLoginRedirect(path));
      return;
    }
    navigate(path);
  };

  const handleView = (fn: RegistryFunction) => navigate(`/fx/${fn.author}/${fn.name}`);

  const loadedCount = functions.length;
  const categoryCount = Math.max(0, categories.length - 1);
  const hasActiveFilters =
    searchQuery !== '' ||
    selectedCategory !== 'All Categories' ||
    trustTierFilter !== 'all' ||
    minTrustScore > 0;

  const topPaddingClass = variant === 'public' ? 'pt-16' : 'pt-6';

  return (
    <main className={`flex-1 ${topPaddingClass}`}>
      <section className="registry-hero section-padding">
        <div className="registry-orb registry-orb-1" aria-hidden="true" />
        <div className="registry-orb registry-orb-2" aria-hidden="true" />
        <div className="registry-orb registry-orb-3" aria-hidden="true" />

        <div className="container-wide px-4 lg:px-6">
          <div className="max-w-3xl animate-stagger">
            <div className="flex items-center gap-3 mb-5">
              <span className="registry-label-pill">
                <Sparkles className="h-3 w-3" />
                Registry
              </span>
              <span className="registry-live-badge">
                <span className="registry-live-dot" />
                Live
              </span>
            </div>

            <h1 className="registry-hero-title mb-4">
              Discover
              <br />
              <span style={{ opacity: 0.7 }}>Functions</span>
            </h1>

            <p className="text-base md:text-lg text-text-secondary mb-7 max-w-xl leading-relaxed">
              Browse community-built serverless functions. View docs and examples freely — sign in
              to deploy or run them live.
            </p>

            {!isAuthenticated && (
              <div className="registry-auth-banner mb-8">
                <LogIn className="h-4 w-4 text-amber-400 shrink-0" />
                <span className="text-sm text-amber-400/90">
                  Sign in to deploy or run functions.
                </span>
                <div className="flex items-center gap-2 ml-auto">
                  <Button
                    size="sm"
                    variant="outline"
                    className="h-7 text-xs border-amber-500/30 text-amber-400 hover:bg-amber-500/10"
                    onClick={() => navigate('/login')}
                  >
                    Sign in
                  </Button>
                  <Button
                    size="sm"
                    className="h-7 text-xs bg-amber-500 hover:bg-amber-400 text-black font-semibold"
                    onClick={() => navigate('/signup')}
                  >
                    Sign up free
                  </Button>
                </div>
              </div>
            )}

            <div className="registry-stats-bar max-w-md">
              <StatItem
                value={loading ? '—' : (totalCount || loadedCount).toLocaleString()}
                label="Functions"
                icon={Code}
              />
              <StatItem value={loading ? '—' : categoryCount} label="Categories" icon={Layers} />
              <StatItem
                value={loading ? '—' : `${filteredFunctions.length.toLocaleString()}${hasMore ? '+' : ''}`}
                label="Visible"
                icon={TrendingUp}
              />
            </div>
          </div>
        </div>
      </section>

      <section className="py-8">
        <div className="container-wide px-4 lg:px-6">
          <div className="registry-layout">
            <aside className="hidden lg:block">
              <div className="registry-sidebar">
                <div className="registry-sidebar-header">
                  <Filter className="h-3.5 w-3.5 text-text-muted" />
                  <span className="registry-sidebar-title">Categories</span>
                </div>
                <div className="registry-sidebar-list">
                  {categories.map((cat) => (
                    <CategoryItem
                      key={cat}
                      category={cat}
                      count={counts[cat] ?? 0}
                      active={selectedCategory === cat}
                      onClick={() => setSelectedCategory(cat)}
                    />
                  ))}
                </div>
              </div>
            </aside>

            <div className="min-w-0 space-y-4">
              <div className="lg:hidden registry-mobile-cats">
                {categories.map((cat) => (
                  <MobileChip
                    key={cat}
                    category={cat}
                    active={selectedCategory === cat}
                    onClick={() => setSelectedCategory(cat)}
                  />
                ))}
              </div>

              <div className="space-y-3">
                <div className="flex gap-3">
                  <div className="registry-search-wrapper flex-1">
                    <div className="relative registry-search-input">
                      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted pointer-events-none" />
                      <Input
                        placeholder="Search by name, author, or description…"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-10 pr-16 bg-bg-secondary border-border-subtle h-10 rounded-xl focus-visible:ring-0"
                      />
                      <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center pointer-events-none">
                        <span className="registry-search-shortcut">⌘K</span>
                      </div>
                    </div>
                  </div>

                  <button
                    type="button"
                    className="lg:hidden flex items-center gap-1.5 h-10 px-3 rounded-xl text-sm font-medium
                               border border-border-subtle bg-bg-secondary text-text-secondary
                               hover:text-text-primary hover:border-brand-500/30 transition-all"
                    onClick={() => setShowMobileFilters(!showMobileFilters)}
                  >
                    <Filter className="h-4 w-4" />
                    Sort
                  </button>

                  <div className="registry-view-toggle">
                    <button
                      type="button"
                      className={`registry-view-btn ${viewMode === 'grid' ? 'active' : ''}`}
                      onClick={() => setViewMode('grid')}
                      aria-label="Grid view"
                    >
                      <Grid className="h-4 w-4" />
                    </button>
                    <button
                      type="button"
                      className={`registry-view-btn ${viewMode === 'list' ? 'active' : ''}`}
                      onClick={() => setViewMode('list')}
                      aria-label="List view"
                    >
                      <List className="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <div className={`${showMobileFilters ? 'block' : 'hidden'} lg:block`}>
                  <div className="registry-controls">
                    <span className="text-xs font-semibold text-text-muted whitespace-nowrap">
                      Sort by
                    </span>
                    <Select value={sortBy} onValueChange={setSortBy}>
                      <SelectTrigger className="flex-1 h-8 text-xs bg-transparent border-border-subtle min-w-[120px] max-w-[160px]">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {SORT_OPTIONS.map((o) => (
                          <SelectItem key={o.value} value={o.value} className="text-xs">
                            {o.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <button
                      type="button"
                      onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
                      className="h-8 w-8 flex items-center justify-center rounded-lg border border-border-subtle
                                 bg-bg-tertiary/50 text-text-muted hover:text-text-primary hover:bg-bg-hover
                                 transition-all"
                      aria-label={sortOrder === 'asc' ? 'Ascending' : 'Descending'}
                    >
                      {sortOrder === 'asc' ? (
                        <SortAsc className="h-3.5 w-3.5" />
                      ) : (
                        <SortDesc className="h-3.5 w-3.5" />
                      )}
                    </button>
                    <div className="flex-1" />
                    <div className="hidden xl:flex items-center gap-2">
                      <Shield className="h-3.5 w-3.5 text-text-muted" aria-hidden="true" />
                      <Select value={trustTierFilter} onValueChange={setTrustTierFilter}>
                        <SelectTrigger className="h-8 text-xs bg-transparent border-border-subtle min-w-[120px] max-w-[140px]">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {TRUST_TIER_OPTIONS.map((o) => (
                            <SelectItem key={o.value} value={o.value} className="text-xs">
                              {o.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <Select
                        value={String(minTrustScore)}
                        onValueChange={(v) => setMinTrustScore(Number(v))}
                      >
                        <SelectTrigger className="h-8 text-xs bg-transparent border-border-subtle min-w-[80px] max-w-[100px]">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {MIN_TRUST_SCORE_OPTIONS.map((o) => (
                            <SelectItem key={o.value} value={String(o.value)} className="text-xs">
                              {o.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="registry-results-count">
                      <strong>{loading ? '…' : filteredFunctions.length}</strong>&nbsp;
                      {filteredFunctions.length === 1 ? 'function' : 'functions'}
                    </div>
                  </div>
                </div>

                <div className={`xl:hidden ${showMobileFilters ? 'block' : 'hidden'}`}>
                  <div className="registry-controls">
                    <Shield className="h-3.5 w-3.5 text-text-muted" aria-hidden="true" />
                    <span className="text-xs font-semibold text-text-muted whitespace-nowrap">
                      Trust
                    </span>
                    <Select value={trustTierFilter} onValueChange={setTrustTierFilter}>
                      <SelectTrigger className="flex-1 h-8 text-xs bg-transparent border-border-subtle min-w-[100px]">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {TRUST_TIER_OPTIONS.map((o) => (
                          <SelectItem key={o.value} value={o.value} className="text-xs">
                            {o.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Select
                      value={String(minTrustScore)}
                      onValueChange={(v) => setMinTrustScore(Number(v))}
                    >
                      <SelectTrigger className="flex-1 h-8 text-xs bg-transparent border-border-subtle min-w-[80px]">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {MIN_TRUST_SCORE_OPTIONS.map((o) => (
                          <SelectItem key={o.value} value={String(o.value)} className="text-xs">
                            {o.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                {hasActiveFilters && (
                  <div className="flex flex-wrap gap-2 items-center">
                    <span className="text-xs text-text-muted font-medium">Filters:</span>
                    {searchQuery && (
                      <span className="registry-filter-chip">
                        <Search className="h-2.5 w-2.5" />"
                        {searchQuery.length > 16 ? searchQuery.slice(0, 16) + '…' : searchQuery}"
                        <button
                          type="button"
                          className="registry-filter-chip-close"
                          onClick={() => setSearchQuery('')}
                          aria-label="Clear search"
                        >
                          <X className="h-2 w-2" />
                        </button>
                      </span>
                    )}
                    {selectedCategory !== 'All Categories' && (
                      <span className="registry-filter-chip">
                        <Layers className="h-2.5 w-2.5" />
                        {selectedCategory}
                        <button
                          type="button"
                          className="registry-filter-chip-close"
                          onClick={() => setSelectedCategory('All Categories')}
                          aria-label="Clear category"
                        >
                          <X className="h-2 w-2" />
                        </button>
                      </span>
                    )}
                    {trustTierFilter !== 'all' && (
                      <span className="registry-filter-chip">
                        <Shield className="h-2.5 w-2.5" />
                        {TRUST_TIER_OPTIONS.find((o) => o.value === trustTierFilter)?.label}
                        <button
                          type="button"
                          className="registry-filter-chip-close"
                          onClick={() => setTrustTierFilter('all')}
                          aria-label="Clear trust tier"
                        >
                          <X className="h-2 w-2" />
                        </button>
                      </span>
                    )}
                    {minTrustScore > 0 && (
                      <span className="registry-filter-chip">
                        <Shield className="h-2.5 w-2.5" />
                        {minTrustScore}%+
                        <button
                          type="button"
                          className="registry-filter-chip-close"
                          onClick={() => setMinTrustScore(0)}
                          aria-label="Clear minimum trust score"
                        >
                          <X className="h-2 w-2" />
                        </button>
                      </span>
                    )}
                    <button
                      type="button"
                      onClick={() => {
                        setSearchQuery('');
                        setSelectedCategory('All Categories');
                        setTrustTierFilter('all');
                        setMinTrustScore(0);
                      }}
                      className="text-xs text-text-muted hover:text-text-secondary transition-colors
                                 underline underline-offset-2"
                    >
                      Clear all
                    </button>
                  </div>
                )}
              </div>

              {loading ? (
                <SkeletonGrid mode={viewMode} count={6} />
              ) : filteredFunctions.length === 0 ? (
                <div className="registry-empty-state">
                  <div className="registry-empty-icon">
                    <Code className="h-8 w-8 text-brand-500" />
                  </div>
                  <h3 className="text-xl font-bold text-text-primary mb-2">No functions found</h3>
                  <p className="text-sm text-text-secondary max-w-sm mb-6">
                    Try adjusting your search query or category filter to find what you're looking
                    for.
                  </p>
                  {hasActiveFilters && (
                    <button
                      type="button"
                      onClick={() => {
                        setSearchQuery('');
                        setSelectedCategory('All Categories');
                        setTrustTierFilter('all');
                        setMinTrustScore(0);
                      }}
                      className="h-9 px-5 rounded-xl text-sm font-semibold
                                 bg-gradient-to-r from-brand-500 to-purple-500 text-white
                                 hover:brightness-110 transition-all shadow-lg shadow-brand-500/20"
                    >
                      Reset filters
                    </button>
                  )}
                </div>
              ) : viewMode === 'grid' ? (
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4 animate-stagger">
                  {filteredFunctions.map((fn) => (
                    <GridCard
                      key={fn.id}
                      fn={fn}
                      isAuthenticated={isAuthenticated}
                      onView={() => handleView(fn)}
                      onDeploy={() => handleDeploy(fn)}
                      onTry={() => handleTry(fn)}
                      flashId={flashId}
                    />
                  ))}
                </div>
              ) : (
                <div className="space-y-2 animate-stagger">
                  {filteredFunctions.map((fn) => (
                    <ListCard
                      key={fn.id}
                      fn={fn}
                      isAuthenticated={isAuthenticated}
                      onView={() => handleView(fn)}
                      onDeploy={() => handleDeploy(fn)}
                      onTry={() => handleTry(fn)}
                      flashId={flashId}
                    />
                  ))}
                </div>
              )}

              {/* Load More */}
              {!loading && hasMore && !hasActiveFilters && (
                <div className="flex justify-center pt-6">
                  <button
                    type="button"
                    onClick={loadMore}
                    disabled={loadingMore}
                    className="h-10 px-6 rounded-xl text-sm font-semibold
                               bg-bg-secondary border border-border-subtle text-text-secondary
                               hover:bg-bg-hover hover:text-text-primary hover:border-brand-500/30
                               disabled:opacity-50 disabled:cursor-not-allowed
                               transition-all flex items-center gap-2"
                  >
                    {loadingMore ? (
                      <>
                        <div className="h-4 w-4 border-2 border-brand-500/30 border-t-brand-500 rounded-full animate-spin" />
                        Loading...
                      </>
                    ) : (
                      <>
                        <Download className="h-4 w-4" />
                        Load More ({loadedCount.toLocaleString()} of {totalCount.toLocaleString()})
                      </>
                    )}
                  </button>
                </div>
              )}

              {/* Loading more indicator (infinite scroll style) */}
              {loadingMore && (
                <div className="flex justify-center py-4">
                  <div className="flex items-center gap-2 text-text-muted text-sm">
                    <div className="h-4 w-4 border-2 border-brand-500/30 border-t-brand-500 rounded-full animate-spin" />
                    Loading more functions...
                  </div>
                </div>
              )}

              {/* End of results */}
              {!loading && !hasMore && loadedCount > 0 && !hasActiveFilters && (
                <div className="text-center py-4 text-text-muted text-sm">
                  Showing all {loadedCount.toLocaleString()} functions
                </div>
              )}

              {/* Trust filter note when paginating */}
              {!loading && hasMore && (trustTierFilter !== 'all' || minTrustScore > 0) && (
                <div className="text-center py-4 text-amber-400/80 text-sm">
                  <Shield className="h-4 w-4 inline mr-1" />
                  Trust filters applied to loaded functions only. Load more to search deeper.
                </div>
              )}
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
