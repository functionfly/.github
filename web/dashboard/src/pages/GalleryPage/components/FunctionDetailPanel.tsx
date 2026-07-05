import { Suspense, lazy, useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import { AnimatePresence, motion } from 'framer-motion';
import { Code2, ExternalLink, GitFork, Heart, Loader2, Maximize2, Star, X } from 'lucide-react';
import type { GalleryFunction } from '@/api/composer';
import { RUNTIME_MONACO_LANG } from '@/api/composer';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { LazyMonacoEditor } from '@/components/LazyMonacoEditor';
import { RUNTIME_COLORS } from '../constants';
import { useFunctionSource, useFunctionStats } from '../useFunctionSource';
import { TrustGauge } from './TrustGauge';

const MonacoEditor = lazy(() => import('@monaco-editor/react').then((m) => ({ default: m.Editor as React.ComponentType<unknown> })));

interface FunctionDetailPanelProps {
  fn: GalleryFunction;
  onClose: () => void;
  onRemix: () => void;
  onLike: () => void;
  onOpenFullPage: () => void;
}

export function FunctionDetailPanel({
  fn,
  onClose,
  onRemix,
  onLike,
  onOpenFullPage,
}: FunctionDetailPanelProps) {
  const [activeTab, setActiveTab] = useState<'info' | 'code' | 'execution'>('info');
  const { data: source, isLoading: sourceLoading, isError: sourceError } = useFunctionSource(
    fn.author,
    fn.name
  );
  const { data: stats, isLoading: statsLoading } = useFunctionStats(fn.author, fn.name);

  useEffect(() => {
    setActiveTab('info');
  }, [fn]);

  const runtime = fn.runtime || 'python';
  const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;
  const monacoLang = RUNTIME_MONACO_LANG[runtime] || 'plaintext';
  const code = sourceLoading
    ? '// Loading source code...'
    : sourceError
      ? '// Source code unavailable'
      : source || '// No source code available';

  return createPortal(
    <AnimatePresence>
      <>
        <motion.div
          key="backdrop"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
          className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[200]"
          onClick={onClose}
        >
          <div className="absolute bottom-8 left-1/2 -translate-x-1/2 text-white/60 text-sm pointer-events-none">
            Click anywhere to close
          </div>
        </motion.div>

        <motion.aside
          key="panel"
          initial={{ x: '100%' }}
          animate={{ x: 0 }}
          exit={{ x: '100%' }}
          transition={{ type: 'spring', damping: 30, stiffness: 300 }}
          className="fixed inset-y-0 right-0 w-full max-w-2xl bg-background border-l-2 border-primary/30 z-[210] shadow-2xl shadow-black/50 overflow-hidden flex flex-col"
        >
          {/* Header */}
          <div
            className="flex items-center justify-between p-6 border-b-2"
            style={{
              background: `linear-gradient(135deg, ${colors.primary}20 0%, transparent 100%)`,
              borderColor: `${colors.primary}40`,
            }}
          >
            <div className="flex items-center gap-3 min-w-0">
              <div
                className="w-12 h-12 rounded-xl flex items-center justify-center border-2 shrink-0"
                style={{
                  backgroundColor: `${colors.primary}30`,
                  borderColor: `${colors.glow}50`,
                }}
              >
                <Code2 className="w-6 h-6" style={{ color: colors.glow }} />
              </div>
              <div className="min-w-0">
                <h2 className="text-xl font-bold text-foreground truncate">{fn.title || fn.name}</h2>
                <p className="text-sm text-muted-foreground flex items-center gap-1">
                  <span className="text-primary">@</span>{fn.author}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-1 shrink-0">
              <Button variant="ghost" size="sm" onClick={onOpenFullPage} title="Open full page">
                <Maximize2 className="w-4 h-4 mr-1" />
                Full page
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={onClose}
                className="text-muted-foreground hover:text-foreground hover:bg-destructive/20"
              >
                <X className="w-5 h-5" />
              </Button>
            </div>
          </div>

          {/* Tabs */}
          <div className="flex gap-1 p-2 border-b border-border">
            {(['info', 'code', 'execution'] as const).map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors capitalize ${
                  activeTab === tab
                    ? 'bg-accent text-accent-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
                }`}
              >
                {tab}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto min-h-0">
            {activeTab === 'info' && (
              <div className="p-6 space-y-6">
                <div className="flex items-center gap-4">
                  <TrustGauge score={fn.trust_score || 0} runtime={runtime} size={56} />
                  <div>
                    <p className="text-sm font-medium text-foreground">Trust Score</p>
                    <p className="text-xs text-muted-foreground">Verified by FunctionFly attestation</p>
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-medium text-foreground/80 mb-2">Description</h3>
                  <p className="text-sm text-muted-foreground leading-relaxed">
                    {fn.description || 'No description available'}
                  </p>
                </div>

                <div className="grid grid-cols-3 gap-4">
                  <div className="p-4 rounded-lg bg-muted/50 text-center">
                    <Star className="w-5 h-5 mx-auto mb-2 text-yellow-500" />
                    <div className="text-2xl font-bold text-foreground">{Math.round(fn.trust_score || 0)}</div>
                    <div className="text-xs text-muted-foreground">Trust Score</div>
                  </div>
                  <div className="p-4 rounded-lg bg-muted/50 text-center">
                    <GitFork className="w-5 h-5 mx-auto mb-2 text-blue-500" />
                    <div className="text-2xl font-bold text-foreground">{fn.remix_count || 0}</div>
                    <div className="text-xs text-muted-foreground">Remixes</div>
                  </div>
                  <div className="p-4 rounded-lg bg-muted/50 text-center">
                    <Heart className="w-5 h-5 mx-auto mb-2 text-pink-500" />
                    <div className="text-2xl font-bold text-foreground">{fn.like_count || 0}</div>
                    <div className="text-xs text-muted-foreground">Likes</div>
                  </div>
                </div>

                <div className="space-y-3">
                  <h3 className="text-sm font-medium text-foreground/80">Details</h3>
                  <div className="grid grid-cols-2 gap-3 text-sm">
                    <div className="flex justify-between py-2 border-b border-border">
                      <span className="text-muted-foreground">Runtime</span>
                      <Badge variant="outline" className="font-mono capitalize">{runtime}</Badge>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border">
                      <span className="text-muted-foreground">Category</span>
                      <span className="text-foreground capitalize">{fn.category || 'General'}</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border">
                      <span className="text-muted-foreground">Created</span>
                      <span className="text-foreground">
                        {fn.created_at ? new Date(fn.created_at).toLocaleDateString() : 'Unknown'}
                      </span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border">
                      <span className="text-muted-foreground">Updated</span>
                      <span className="text-foreground">
                        {fn.updated_at ? new Date(fn.updated_at).toLocaleDateString() : 'Unknown'}
                      </span>
                    </div>
                  </div>
                </div>

                {fn.tags && fn.tags.length > 0 && (
                  <div>
                    <h3 className="text-sm font-medium text-foreground/80 mb-2">Tags</h3>
                    <div className="flex flex-wrap gap-2">
                      {fn.tags.map((tag) => (
                        <Badge key={tag} variant="secondary" className="bg-muted text-muted-foreground">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}

                <div className="flex gap-3 pt-4">
                  <Button
                    className="flex-1"
                    style={{ backgroundColor: colors.primary }}
                    onClick={onRemix}
                  >
                    <GitFork className="w-4 h-4 mr-2" />
                    Remix
                  </Button>
                  <Button variant="outline" className="flex-1" onClick={onLike}>
                    <Heart className="w-4 h-4 mr-2" />
                    Like
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => window.open(`/registry/${fn.author}/${fn.name}`, '_blank')}
                  >
                    <ExternalLink className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            )}

            {activeTab === 'code' && (
              <div className="h-full min-h-[400px]">
                <Suspense
                  fallback={
                    <div className="flex items-center justify-center h-64">
                      <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
                    </div>
                  }
                >
                  <LazyMonacoEditor
                    height="100%"
                    language={monacoLang}
                    value={code}
                    theme="vs-dark"
                    options={{
                      readOnly: true,
                      minimap: { enabled: false },
                      fontSize: 14,
                      lineNumbers: 'on',
                      roundedSelection: false,
                      scrollBeyondLastLine: false,
                      automaticLayout: true,
                      padding: { top: 20 },
                    }}
                  />
                </Suspense>
              </div>
            )}

            {activeTab === 'execution' && (
              <div className="p-6 space-y-4">
                <div className="p-4 rounded-lg bg-muted/50 border border-border">
                  <h3 className="font-medium text-foreground mb-4">Execution Metrics</h3>
                  {statsLoading ? (
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Loading stats...
                    </div>
                  ) : stats ? (
                    <div className="space-y-3 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Avg. Execution Time</span>
                        <span className="text-emerald-500 font-mono">{stats.avg_latency_ms}ms</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Success Rate</span>
                        <span className="text-emerald-500 font-mono">
                          {(stats.success_rate <= 1 ? stats.success_rate * 100 : stats.success_rate).toFixed(1)}%
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Total Calls</span>
                        <span className="text-primary font-mono">
                          {stats.total_calls.toLocaleString()}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">P95 Latency</span>
                        <span className="text-warning font-mono">{stats.p95_latency_ms}ms</span>
                      </div>
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">No execution stats available yet.</p>
                  )}
                </div>
              </div>
            )}
          </div>
        </motion.aside>
      </>
    </AnimatePresence>,
    document.body
  );
}
