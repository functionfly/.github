import { Suspense, lazy, useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import {
  ArrowLeft,
  Calendar,
  ChevronRight,
  ExternalLink,
  GitFork,
  Heart,
  Loader2,
  Play,
  Share2,
  Sparkles,
  Star,
  Tag,
} from 'lucide-react';
import { toast } from 'sonner';
import { galleryApi, RUNTIME_MONACO_LANG } from '@/api/composer';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { FunctionTile } from './components/FunctionTile';
import { RemixDialog } from './components/RemixDialog';
import { TrustGauge } from './components/TrustGauge';
import {
  CATEGORY_META,
  RUNTIME_COLORS,
  RUNTIME_ICONS,
} from './constants';
import { galleryFunctionPath, useGalleryFunction, useRelatedFunctions } from './useGalleryFunction';
import { useFunctionSource } from './useFunctionSource';
import './gallery.css';

const MonacoEditor = lazy(() => import('@monaco-editor/react'));

export default function GalleryFunctionPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: fn, isLoading, isError, error } = useGalleryFunction(author, name);
  const { data: related = [] } = useRelatedFunctions(fn);
  const { data: source, isLoading: sourceLoading } = useFunctionSource(author, name);

  const [tab, setTab] = useState<'overview' | 'code'>('overview');
  const [remixDialogOpen, setRemixDialogOpen] = useState(false);
  const [customization, setCustomization] = useState('');
  const [remixCost, setRemixCost] = useState(0.5);
  const [walletBalance, setWalletBalance] = useState(0);
  const [canRemix, setCanRemix] = useState(false);
  const [isOwnFunction, setIsOwnFunction] = useState(false);

  useEffect(() => {
    if (remixDialogOpen && fn) {
      galleryApi
        .getRemixCost(fn.author, fn.name)
        .then((data) => {
          setRemixCost(data.cost_usd);
          setWalletBalance(data.balance_usd);
          setCanRemix(data.can_remix || data.is_own_function);
          setIsOwnFunction(data.is_own_function);
        })
        .catch(() => setCanRemix(true));
    }
  }, [remixDialogOpen, fn]);

  const remixMutation = useMutation({
    mutationFn: (data: { author: string; name: string; customization?: string }) =>
      galleryApi.remix(data.author, data.name, { customization: data.customization }),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['gallery'] });
      toast.success(`Remixed! Created "${data.new_name}"`);
      setRemixDialogOpen(false);
      setCustomization('');
    },
    onError: (err: Error) => toast.error(`Failed to remix: ${err.message}`),
  });

  const likeMutation = useMutation({
    mutationFn: (data: { author: string; name: string }) => galleryApi.like(data.author, data.name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['gallery'] });
      toast.success('Liked!');
    },
  });

  if (isLoading) {
    return (
      <div className="flyway-gallery -m-4 lg:-m-6 rounded-xl min-h-[60vh] flex items-center justify-center">
        <Loader2 className="w-10 h-10 animate-spin text-[var(--flyway-flame)]" />
      </div>
    );
  }

  if (isError || !fn) {
    return (
      <div className="flyway-gallery -m-4 lg:-m-6 rounded-xl min-h-[60vh] flex flex-col items-center justify-center gap-4 p-8">
        <p className="text-lg font-medium">Function not found</p>
        <p className="text-sm text-muted-foreground">{(error as Error)?.message || 'Unknown error'}</p>
        <Button variant="outline" onClick={() => navigate('/gallery')}>
          <ArrowLeft className="w-4 h-4 mr-2" /> Back to Gallery
        </Button>
      </div>
    );
  }

  const runtime = fn.runtime || 'python';
  const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;
  const categoryMeta = CATEGORY_META[fn.category || 'default'] || CATEGORY_META.default;
  const monacoLang = RUNTIME_MONACO_LANG[runtime] || 'plaintext';
  const code = sourceLoading ? '// Loading source code...' : source || '// No source code available';

  return (
    <div className="flyway-gallery -m-4 lg:-m-6 rounded-xl overflow-hidden">
      {/* Breadcrumb */}
      <nav className="flyway-detail-nav">
        <button type="button" onClick={() => navigate('/gallery')} className="flyway-back-link">
          <ArrowLeft className="w-4 h-4" />
          Gallery
        </button>
        <ChevronRight className="w-3.5 h-3.5 text-muted-foreground" />
        <span className="text-muted-foreground truncate">@{fn.author}</span>
        <ChevronRight className="w-3.5 h-3.5 text-muted-foreground" />
        <span className="font-medium truncate">{fn.name}</span>
      </nav>

      {/* Hero */}
      <motion.header
        className="flyway-detail-hero"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        style={{ '--detail-accent': colors.primary, '--detail-glow': colors.glow } as React.CSSProperties}
      >
        <div
          className="flyway-detail-hero-glow"
          style={{ background: `radial-gradient(circle, ${colors.primary}40, transparent 70%)` }}
        />

        <div className="flex flex-col lg:flex-row lg:items-start gap-6">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-3">
              <span className="text-4xl">{RUNTIME_ICONS[runtime] || '⚡'}</span>
              <div>
                <div className="flex flex-wrap items-center gap-2 mb-1">
                  <Badge style={{ backgroundColor: `${categoryMeta.color}25`, color: categoryMeta.color, border: 'none' }}>
                    {categoryMeta.label}
                  </Badge>
                  <Badge variant="outline" className="font-mono capitalize">{runtime}</Badge>
                </div>
                <h1 className="flyway-detail-title">{fn.title || fn.name}</h1>
                <p className="text-muted-foreground mt-1">
                  by <Link to={`/u/${fn.author}`} className="text-[var(--flyway-cyan)] hover:underline">@{fn.author}</Link>
                </p>
              </div>
            </div>

            <p className="text-base text-muted-foreground leading-relaxed max-w-2xl">{fn.description}</p>

            {fn.tags && fn.tags.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-4">
                {fn.tags.map((tag) => (
                  <span key={tag} className="flyway-tag">
                    <Tag className="w-3 h-3" /> {tag}
                  </span>
                ))}
              </div>
            )}
          </div>

          <div className="flex flex-col items-center gap-4 shrink-0">
            <TrustGauge score={fn.trust_score || 0} runtime={runtime} size={80} />
            <p className="text-xs text-muted-foreground text-center">Trust Score</p>
          </div>
        </div>

        {/* Stats row */}
        <div className="flyway-detail-stats">
          <div className="flyway-detail-stat">
            <Star className="w-4 h-4 text-amber-400" />
            <span className="flyway-detail-stat-value">{Math.round(fn.trust_score || 0)}</span>
            <span className="flyway-detail-stat-label">Trust</span>
          </div>
          <div className="flyway-detail-stat">
            <GitFork className="w-4 h-4 text-blue-400" />
            <span className="flyway-detail-stat-value">{fn.remix_count || 0}</span>
            <span className="flyway-detail-stat-label">Remixes</span>
          </div>
          <div className="flyway-detail-stat">
            <Heart className="w-4 h-4 text-pink-400" />
            <span className="flyway-detail-stat-value">{fn.like_count || 0}</span>
            <span className="flyway-detail-stat-label">Likes</span>
          </div>
          {fn.created_at && (
            <div className="flyway-detail-stat">
              <Calendar className="w-4 h-4 text-muted-foreground" />
              <span className="flyway-detail-stat-value text-sm">
                {new Date(fn.created_at).toLocaleDateString()}
              </span>
              <span className="flyway-detail-stat-label">Created</span>
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="flex flex-wrap gap-3 mt-6">
          <Button
            size="lg"
            className="bg-gradient-to-r from-[var(--flyway-flame)] to-[var(--flyway-afterburner)]"
            onClick={() => setRemixDialogOpen(true)}
          >
            <GitFork className="w-4 h-4 mr-2" /> Remix Function
          </Button>
          <Button size="lg" variant="outline" onClick={() => likeMutation.mutate({ author: fn.author, name: fn.name })}>
            <Heart className="w-4 h-4 mr-2" /> Like
          </Button>
          <Button size="lg" variant="outline" onClick={() => navigate(`/run/${fn.author}/${fn.name}`)}>
            <Play className="w-4 h-4 mr-2" /> Try in Playground
          </Button>
          <Button size="lg" variant="ghost" onClick={() => navigate(`/ai/composer?remix=${fn.author}/${fn.name}`)}>
            <Sparkles className="w-4 h-4 mr-2" /> AI Customize
          </Button>
          <Button size="lg" variant="ghost" onClick={() => navigate(`/registry/${fn.author}/${fn.name}`)}>
            <ExternalLink className="w-4 h-4 mr-2" /> Registry
          </Button>
          <Button
            size="lg"
            variant="ghost"
            onClick={() => {
              navigator.clipboard.writeText(window.location.href);
              toast.success('Link copied!');
            }}
          >
            <Share2 className="w-4 h-4 mr-2" /> Share
          </Button>
        </div>
      </motion.header>

      {/* Tabs */}
      <div className="flyway-detail-tabs">
        {(['overview', 'code'] as const).map((t) => (
          <button
            key={t}
            type="button"
            className={cn('flyway-detail-tab', tab === t && 'active')}
            onClick={() => setTab(t)}
          >
            {t}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="flyway-detail-content">
        {tab === 'overview' ? (
          <div className="grid lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 space-y-6">
              <section className="flyway-detail-card">
                <h2 className="text-lg font-semibold mb-3">About this function</h2>
                <p className="text-muted-foreground leading-relaxed">{fn.description}</p>
                <dl className="grid grid-cols-2 gap-4 mt-6 text-sm">
                  <div>
                    <dt className="text-muted-foreground">Runtime</dt>
                    <dd className="font-mono capitalize mt-1">{runtime}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Category</dt>
                    <dd className="capitalize mt-1">{categoryMeta.label}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Author</dt>
                    <dd className="mt-1">@{fn.author}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Popularity</dt>
                    <dd className="mt-1">{fn.popularity_score || 0}</dd>
                  </div>
                </dl>
              </section>

              <section className="flyway-detail-card">
                <h2 className="text-lg font-semibold mb-3">Quick start</h2>
                <pre className="flyway-code-preview">
                  {(sourceLoading ? '// Loading...' : code).split('\n').slice(0, 8).join('\n')}...
                </pre>
                <Button variant="link" className="mt-2 px-0" onClick={() => setTab('code')}>
                  View full source →
                </Button>
              </section>
            </div>

            <aside className="space-y-4">
              <section className="flyway-detail-card">
                <h3 className="font-semibold mb-3">Trust attestation</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  This function has a trust score of {Math.round(fn.trust_score || 0)}/100 based on
                  execution reliability, community remixes, and verification checks.
                </p>
                <div className="h-2 rounded-full bg-white/5 overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all"
                    style={{
                      width: `${fn.trust_score || 0}%`,
                      background: `linear-gradient(90deg, ${colors.primary}, ${colors.glow})`,
                    }}
                  />
                </div>
              </section>
            </aside>
          </div>
        ) : (
          <div className="flyway-detail-code-panel">
            <Suspense fallback={<div className="flex justify-center p-16"><Loader2 className="w-8 h-8 animate-spin" /></div>}>
              <MonacoEditor
                height="520px"
                language={monacoLang}
                value={code}
                theme="vs-dark"
                options={{
                  readOnly: true,
                  minimap: { enabled: true },
                  fontSize: 14,
                  lineNumbers: 'on',
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  padding: { top: 16 },
                }}
              />
            </Suspense>
          </div>
        )}
      </div>

      {/* Related */}
      {related.length > 0 && (
        <section className="flyway-detail-related">
          <h2 className="text-lg font-semibold mb-4">Related functions</h2>
          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {related.map((rel, i) => (
              <FunctionTile
                key={rel.id}
                fn={rel}
                index={i}
                onClick={() => navigate(galleryFunctionPath(rel.author, rel.name))}
              />
            ))}
          </div>
        </section>
      )}

      <RemixDialog
        open={remixDialogOpen}
        onOpenChange={setRemixDialogOpen}
        fn={fn}
        customization={customization}
        onCustomizationChange={setCustomization}
        remixCost={remixCost}
        walletBalance={walletBalance}
        canRemix={canRemix}
        isOwnFunction={isOwnFunction}
        isPending={remixMutation.isPending}
        onConfirm={() =>
          remixMutation.mutate({ author: fn.author, name: fn.name, customization })
        }
        onAddFunds={() => navigate('/wallet')}
      />
    </div>
  );
}
