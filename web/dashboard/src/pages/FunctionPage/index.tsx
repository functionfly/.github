import { useState } from 'react';
import { apiClient } from '@/api/client';
import { FunctionNotFound } from './FunctionNotFound';
import { FunctionPageSkeleton } from './FunctionPageSkeleton';
import { registryApi } from '@/api/registry';
import { favoritesApi } from '@/api/favorites';
import { Navbar } from '@/components/common/Navbar';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';
import { FollowFunctionButton } from '@/components/follow';
import {
  FunctionCard,
  FunctionEmbedSection,
  FunctionHeader,
} from '@/components/functions';
import { CodeBlock } from '@/components/common/CodeBlock';
import {
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  StatusPill,
  Card,
  Modal,
  PageGrid,
} from '@/components/containment';
import { Footer } from '@/pages/LandingPage/components';
import { useAuthStore } from '@/stores/authStore';
import { useSubmitRegistryRating } from '@/hooks/useRegistry';
import { usePageTitle } from '@/hooks';
import { Icon } from '@iconify/react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Helmet } from 'react-helmet-async';
import { motion } from 'framer-motion';
import {
  Activity,
  BarChart3,
  Clock,
  FileJson,
  Layers,
  Play,
  Shield,
  TrendingUp,
  Zap,
  AlertTriangle,
  AlertCircle,
  AlertOctagon,
  Gauge,
} from 'lucide-react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { ApiReferenceSection } from './ApiReferenceSection';
import { CodeExamplesSection } from './CodeExamplesSection';
import type { FunctionInfo } from './types';
import {
  mapToFunctionCardData,
  mapToFunctionHeaderData,
  mapToTrustMetrics,
  mapToDNATrustData,
} from './utils';
import { ReviewsSection } from './ReviewsSection';
import { ShareButton } from './ShareButton';
import { SourceSection } from './SourceSection';
import { TableOfContents, type TocItem } from './TableOfContents';
import { DNATrustBadge } from '@/components/dna/DNATrustBadge';
import { TrustDashboardWidget } from '@/components/trust/TrustDashboardWidget';
import { TrustHistory } from '@/components/trust/TrustHistory';
import { ActivityFeed } from '@/components/common/ActivityFeed';
import { ExecutionTimeline } from '@/components/execution/ExecutionTimeline';
import { TraceList } from '@/components/atlas';
import { MarkdownRenderer } from './MarkdownRenderer';
import { RelatedFunctionsSection } from './RelatedFunctionsSection';
import { formatNumber } from '@/lib/utils';
import './FunctionPage.css';

export default function FunctionPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const submitRating = useSubmitRegistryRating();

  const {
    data: functionInfo,
    isLoading,
    error,
    isError,
  } = useQuery<FunctionInfo>({
    queryKey: ['function', author, name],
    queryFn: async () => {
      try {
        return await apiClient.get<FunctionInfo>(`/v1/functions/${author}/${name}?expand=manifest`);
      } catch (err: unknown) {
        const status = (err as { response?: { status?: number } })?.response?.status;
        if (status === 404) throw new Error('Function not found');
        throw err;
      }
    },
    enabled: !!author && !!name,
    retry: (failureCount, error) => {
      if ((error as Error)?.message === 'Function not found') return false;
      return failureCount < 2;
    },
  });

  const pageTitle = functionInfo?.title || functionInfo?.name || `${author}/${name}`;
  usePageTitle(`${pageTitle} by ${author}`);

  const handleRate = (functionId: string, stars: number) => {
    if (!isAuthenticated) {
      toast.info('Sign in to rate this function');
      navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
      return;
    }
    if (!author || !name) return;
    submitRating.mutate({
      author,
      name,
      rating: {
        overall_score: stars,
        reliability_score: 0,
        latency_score: 0,
        documentation_score: 0,
      },
    });
  };

  const [showReportDialog, setShowReportDialog] = useState(false);
  const [reportDescription, setReportDescription] = useState('');

  const { data: favoritesResp } = useQuery({
    queryKey: ['favorites-list'],
    queryFn: async () => favoritesApi.list(1, 100),
    enabled: isAuthenticated,
  });

  const handleReportIssue = () => {
    setShowReportDialog(true);
  };

  const submitReport = async () => {
    if (!functionInfo) return;
    try {
      await apiClient.post(`/v1/functions/${functionInfo.id}/report-issue`, {
        description: reportDescription,
        functionName: functionInfo.name,
        author: functionInfo.author,
      });
      toast.success('Issue reported successfully');
      setShowReportDialog(false);
      setReportDescription('');
    } catch (error) {
      console.error('Failed to report issue:', error);
      toast.error('Failed to report issue');
    }
  };

  const isFavorited =
    favoritesResp?.favorites.some((f) => f.function_id === functionInfo?.id) ?? false;

  const {
    data: sourceCode,
    isLoading: isLoadingSource,
    error: sourceError,
    refetch: refetchSource,
  } = useQuery({
    queryKey: ['function-source', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      return registryApi.getFunctionVersionSource(author, name, 'latest');
    },
    enabled: !!author && !!name,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  const { data: trustHistoryData } = useQuery<{ history?: Array<{ calculated_at: string; trust_score: number; reliability_score?: number; user_rating_score?: number }> }>({
    queryKey: ['function-trust-history', functionInfo?.id],
    queryFn: async () => {
      if (!functionInfo?.id) return null;
      try {
        return await apiClient.get(`/v1/functions/${functionInfo.id}/trust/history`, { params: { page_size: 30 } });
      } catch {
        return null;
      }
    },
    enabled: !!functionInfo?.id,
    staleTime: 5 * 60 * 1000,
  });

  const { data: executionsData } = useQuery<{ executions?: Array<{ execution_id: string; created_at: string; version: string; replay_verified: boolean; roots_match?: boolean }> }>({
    queryKey: ['function-executions', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      try {
        return await apiClient.get(`/functions/${author}/${name}/executions`, { params: { limit: 10 } });
      } catch {
        return null;
      }
    },
    enabled: !!author && !!name,
    staleTime: 30 * 1000,
  });

  const { data: versionsData } = useQuery<{ versions?: Array<{ version: string; published_at: string; changelog?: string }> }>({
    queryKey: ['function-versions', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      try {
        return await apiClient.get(`/functions/${author}/${name}/versions`);
      } catch {
        return null;
      }
    },
    enabled: !!author && !!name,
    staleTime: 5 * 60 * 1000,
  });

  const { data: statsData } = useQuery<{ total_calls?: number; success_rate?: number; avg_latency_ms?: number }>({
    queryKey: ['function-stats', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      try {
        return await apiClient.get(`/functions/${author}/${name}/stats`);
      } catch {
        return null;
      }
    },
    enabled: !!author && !!name,
    staleTime: 30 * 1000,
  });

  const latestExecution = executionsData?.executions?.[0];

  const getExecutionPhases = (
    exec: {
      execution_id: string;
      created_at: string;
      version: string;
      replay_verified: boolean;
      roots_match?: boolean;
      determinism_tier?: string;
    },
    avgLatencyMs: number | undefined
  ) => {
    const execMs = avgLatencyMs ?? 120;
    const phase2End = 150;
    const phase3End = phase2End + 80;
    const phase4End = phase3End + execMs;
    const phase5End = phase4End + 40;
    const phase6End = phase5End + 15;
    return [
      {
        id: '1',
        name: 'Container Start',
        durationMs: 150,
        startTime: 0,
        endTime: 150,
        status: 'completed' as const,
      },
      {
        id: '2',
        name: 'Runtime Init',
        durationMs: 80,
        startTime: 150,
        endTime: phase3End,
        status: 'completed' as const,
      },
      {
        id: '3',
        name: 'Function Execution',
        durationMs: execMs,
        startTime: phase3End,
        endTime: phase4End,
        status: 'completed' as const,
        metadata: { avg_latency_ms: execMs, determinism_tier: exec.determinism_tier ?? 'full' },
      },
      {
        id: '4',
        name: 'Verification',
        durationMs: 40,
        startTime: phase4End,
        endTime: phase5End,
        status: exec.replay_verified ? ('completed' as const) : ('failed' as const),
        metadata: {
          verified: exec.replay_verified ? 1 : 0,
          deterministic: (exec.roots_match ?? false) ? 1 : 0,
        },
      },
      {
        id: '5',
        name: 'Response',
        durationMs: 15,
        startTime: phase5End,
        endTime: phase6End,
        status: 'completed' as const,
      },
    ];
  };

  if (isLoading) {
    return (
      <div className="function-page">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16">
          <FunctionPageSkeleton />
        </main>
        <Footer />
      </div>
    );
  }

  if (isError || !functionInfo) {
    const isNotFound =
      error instanceof Error && error.message === 'Function not found';
    return (
      <div className="function-page">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16">
          {isNotFound ? (
            <FunctionNotFound author={author} name={name} />
          ) : (
            <div className="flex items-center justify-center min-h-[40vh]">
              <div className="text-center space-y-4">
                <AlertTriangle className="h-10 w-10 mx-auto" style={{ color: 'var(--status-pending)' }} />
                <p className="text-lg font-medium" style={{ color: 'var(--text)' }}>Failed to load function</p>
                <p className="text-sm" style={{ color: 'var(--text-faint)' }}>
                  {error instanceof Error ? error.message : 'An unexpected error occurred'}
                </p>
                <FrameButton onClick={() => window.location.reload()}>
                  Retry
                </FrameButton>
              </div>
            </div>
          )}
        </main>
        <Footer />
      </div>
    );
  }

  const tocItems: TocItem[] = [
    { id: 'fp-header', label: 'Header' },
    { id: 'fp-overview', label: 'Overview' },
    { id: 'fp-stats', label: 'Stats' },
    { id: 'fp-dna', label: 'DNA' },
    { id: 'fp-trust', label: 'Trust' },
    { id: 'fp-activity', label: 'Activity' },
    { id: 'fp-execution', label: 'Execution' },
    { id: 'fp-traces', label: 'Traces' },
    { id: 'fp-versions', label: 'Versions' },
    { id: 'fp-readme', label: 'README' },
    { id: 'fp-reviews', label: 'Reviews' },
    { id: 'fp-api', label: 'API Reference' },
    { id: 'fp-source', label: 'Source' },
    { id: 'fp-examples', label: 'Code Examples' },
    { id: 'fp-embed', label: 'Embed' },
    { id: 'fp-schema', label: 'Schema' },
  ];

  return (
    <ErrorBoundary>
    <>
      <Helmet>
        <title>{`${pageTitle} by ${author} | FunctionFly`}</title>
        <meta
          name="description"
          content={functionInfo.description || `Explore ${functionInfo.name} on FunctionFly`}
        />
        <meta name="robots" content="index, follow" />
        <link
          rel="canonical"
          href={`https://functionfly.com/fx/${functionInfo.author}/${functionInfo.name}`}
        />
        <meta
          property="og:title"
          content={`${functionInfo.title || functionInfo.name} | FunctionFly`}
        />
        <meta
          property="og:description"
          content={
            functionInfo.description || `Function ${functionInfo.name} by ${functionInfo.author}`
          }
        />
        <meta property="og:type" content="website" />
        <meta
          property="og:image"
          content={`https://functionfly.com/api/og/function?author=${encodeURIComponent(functionInfo.author)}&name=${encodeURIComponent(functionInfo.name)}`}
        />
        <meta property="og:url" content={`https://functionfly.com/fx/${functionInfo.author}/${functionInfo.name}`} />
        <meta name="twitter:card" content="summary_large_image" />
        <meta
          name="twitter:title"
          content={`${functionInfo.title || functionInfo.name} | FunctionFly`}
        />
        <meta
          name="twitter:description"
          content={
            functionInfo.description || `Function ${functionInfo.name} by ${functionInfo.author}`
          }
        />
        <meta
          name="twitter:image"
          content={`https://functionfly.com/api/og/function?author=${encodeURIComponent(functionInfo.author)}&name=${encodeURIComponent(functionInfo.name)}`}
        />
        <script type="application/ld+json">
          {JSON.stringify({
            '@context': 'https://schema.org',
            '@type': 'SoftwareSourceCode',
            name: functionInfo.title || functionInfo.name,
            description: functionInfo.description || `Serverless function ${functionInfo.name} by ${functionInfo.author}`,
            url: `https://functionfly.com/fx/${functionInfo.author}/${functionInfo.name}`,
            codeRepository: functionInfo.repo_url || undefined,
            programmingLanguage: functionInfo.runtime,
            author: {
              '@type': 'Organization',
              name: functionInfo.author,
            },
            offers: {
              '@type': 'Offer',
              price: functionInfo.price_per_call ?? 0,
              priceCurrency: 'USD',
            },
            aggregateRating: functionInfo.stars
              ? {
                  '@type': 'AggregateRating',
                  ratingValue: Math.min(functionInfo.stars / 20, 5).toFixed(1),
                  bestRating: '5',
                  ratingCount: functionInfo.executions || 0,
                }
              : undefined,
            dateCreated: functionInfo.created_at,
            dateModified: functionInfo.updated_at,
            version: functionInfo.version,
            applicationCategory: functionInfo.category || 'DeveloperApplication',
            isAccessibleForFree: (functionInfo.price_per_call ?? 0) === 0,
          })}
        </script>
      </Helmet>
      <div className="function-page">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16">
          <div className="function-page-layout">
            <aside className="function-page-toc-wrapper">
              <TableOfContents items={tocItems} />
            </aside>
            <div className="function-page-content">
              <motion.div
                id="fp-header"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, ease: 'easeOut' }}
                className="function-page-section"
              >
                <FunctionHeader
                  data={mapToFunctionHeaderData(functionInfo)}
                  onBack={() => navigate('/registry')}
                  onReportIssue={handleReportIssue}
                />

                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.4 }}
                  className="function-page-cta"
                >
                  <Link
                    to={`/run/${functionInfo.author}/${functionInfo.name}`}
                    className="function-page-cta-button function-page-cta-button--primary"
                  >
                    <Play className="w-4 h-4" />
                    Try it Now
                  </Link>
                  <Link
                    to={`/registry/${functionInfo.author}/${functionInfo.name}/executions`}
                    className="function-page-cta-button function-page-cta-button--secondary"
                  >
                    <Activity className="w-4 h-4" />
                    Executions
                  </Link>
                  <Link
                    to={`/registry/${functionInfo.author}/${functionInfo.name}/executions?tab=certificates`}
                    className="function-page-cta-button function-page-cta-button--secondary"
                  >
                    <Shield className="w-4 h-4" />
                    Certificates
                  </Link>
                  {functionInfo.repo_url ? (
                    <a
                      href={functionInfo.repo_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="function-page-cta-button function-page-cta-button--secondary"
                    >
                      <Icon icon="simple-icons:github" className="w-4 h-4" />
                      View on GitHub
                    </a>
                  ) : null}
                  <ShareButton functionInfo={functionInfo} />
                  <FollowFunctionButton
                    functionId={functionInfo.id}
                    functionName={functionInfo.name}
                    size="lg"
                  />
                </motion.div>
              </motion.div>

              {functionInfo.manifest?.deprecated && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3 }}
                  className="rounded-[var(--radius-lg)] p-4 flex items-start gap-3"
                  style={{
                    background: 'rgba(232, 196, 104, 0.06)',
                    border: '1px solid rgba(232, 196, 104, 0.2)',
                    marginBottom: '1.5rem',
                  }}
                >
                  <AlertOctagon className="w-5 h-5 mt-0.5 shrink-0" style={{ color: 'var(--status-pending)' }} />
                  <div>
                    <p className="text-sm font-semibold" style={{ color: 'var(--status-pending)' }}>
                      This function is deprecated
                    </p>
                    <p className="text-xs mt-1" style={{ color: 'var(--text-faint)' }}>
                      {functionInfo.manifest.successor
                        ? <>This function has been replaced. Consider using the successor instead.</>
                        : <>This function may be removed in a future release. Use with caution.</>
                      }
                    </p>
                  </div>
                </motion.div>
              )}

              {functionInfo.manifest?.rate_limit != null && functionInfo.manifest.rate_limit > 0 && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3, delay: 0.1 }}
                  className="rounded-[var(--radius-lg)] p-3 flex items-center gap-3"
                  style={{
                    background: 'var(--panel-raised)',
                    border: '1px solid var(--panel-edge)',
                    marginBottom: '1.5rem',
                  }}
                >
                  <Gauge className="w-4 h-4 shrink-0" style={{ color: 'var(--foil-b)' }} />
                  <p className="text-xs" style={{ color: 'var(--text-dim)' }}>
                    Rate limited to <strong style={{ color: 'var(--text)' }}>{functionInfo.manifest.rate_limit}</strong> requests per minute
                  </p>
                </motion.div>
              )}

              <motion.div
                id="fp-overview"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="function-page-section function-page-section--delayed"
              >
                <FunctionCard
                  data={{ ...mapToFunctionCardData(functionInfo), isFavorite: isFavorited }}
                  variant="expanded"
                  onView={() => {
                    document
                      .getElementById('fp-api')
                      ?.scrollIntoView({ behavior: 'smooth', block: 'start' });
                  }}
                  onExecute={() => navigate(`/run/${functionInfo.author}/${functionInfo.name}`)}
                  onFavorite={async () => {
                    if (!isAuthenticated) {
                      toast.info('Sign in to add favorites');
                      navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
                      return;
                    }
                    try {
                      if (isFavorited) {
                        await favoritesApi.remove(functionInfo.id);
                        toast.success('Removed from favorites');
                      } else {
                        await favoritesApi.add(functionInfo.id);
                        toast.success('Added to favorites');
                      }
                      queryClient.invalidateQueries({ queryKey: ['favorites-list'] });
                    } catch {
                      toast.error('Failed to update favorite');
                    }
                  }}
                  onRate={handleRate}
                />
              </motion.div>

              <motion.div
                id="fp-stats"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="function-page-section function-page-section--delayed"
              >
                <div className="flex items-center gap-2 mb-4">
                  <BarChart3 className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
                  <h2 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)' }}>Function Statistics</h2>
                </div>
                <div className="sc-fp__stats-grid">
                  <Card className="sc-fp__stat-card">
                    <Activity size={28} style={{ color: 'var(--status-ok)' }} />
                    <p className="sc-fp__stat-value">
                      {statsData?.total_calls ? formatNumber(statsData.total_calls) : formatNumber(functionInfo.executions || 0)}
                    </p>
                    <p className="sc-fp__stat-label">Total Executions</p>
                  </Card>
                  <Card className="sc-fp__stat-card">
                    <TrendingUp size={28} style={{ color: 'var(--status-ok)' }} />
                    <p className="sc-fp__stat-value">
                      {statsData?.success_rate != null ? `${statsData.success_rate.toFixed(1)}%` : '—'}
                    </p>
                    <p className="sc-fp__stat-label">Success Rate</p>
                  </Card>
                  <Card className="sc-fp__stat-card">
                    <Zap size={28} style={{ color: 'var(--status-pending)' }} />
                    <p className="sc-fp__stat-value">
                      {statsData?.avg_latency_ms != null ? `${Math.round(statsData.avg_latency_ms)}ms` : '—'}
                    </p>
                    <p className="sc-fp__stat-label">Avg Latency</p>
                  </Card>
                  <Card className="sc-fp__stat-card">
                    <Shield size={28} style={{ color: 'var(--foil-b)' }} />
                    <p className="sc-fp__stat-value">
                      {functionInfo.trust_score ? `${functionInfo.trust_score}%` : 'N/A'}
                    </p>
                    <p className="sc-fp__stat-label">Trust Score</p>
                  </Card>
                </div>
              </motion.div>

              {mapToDNATrustData(functionInfo) && (
                <motion.div
                  id="fp-dna"
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5 }}
                  className="function-page-section function-page-section--delayed"
                >
                  <div className="flex items-center gap-2 mb-4">
                    <h2 className="text-2xl font-bold flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
                      <span style={{ color: 'var(--status-pending)' }}>⚡</span>
                      Function DNA
                    </h2>
                  </div>
                  <DNATrustBadge
                    generation={mapToDNATrustData(functionInfo)!.generation}
                    fitnessScore={mapToDNATrustData(functionInfo)!.fitnessScore}
                    totalMutations={mapToDNATrustData(functionInfo)!.totalMutations}
                    totalExecutions={mapToDNATrustData(functionInfo)!.totalExecutions}
                    variant="full"
                    className="max-w-md"
                  />
                </motion.div>
              )}

              <motion.div
                id="fp-trust"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="function-page-section function-page-section--delayed-2"
              >
                <div className="flex items-center gap-2 mb-4">
                  <Shield className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
                  <h2 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)' }}>Trust & Verification</h2>
                </div>
                <TrustDashboardWidget
                  functionId={functionInfo.id}
                  functionName={`${functionInfo.author}/${functionInfo.name}`}
                  metrics={{
                    overallScore: mapToTrustMetrics(functionInfo).overallScore,
                    reliability: mapToTrustMetrics(functionInfo).reliability,
                    determinism: mapToTrustMetrics(functionInfo).determinism,
                    communityReputation: mapToTrustMetrics(functionInfo).communityReputation,
                  }}
                  variant="standard"
                  showVerification={false}
                />
                {trustHistoryData?.history && trustHistoryData.history.length > 0 && (
                  <div className="mt-6">
                    <TrustHistory
                      data={trustHistoryData.history.map(
                        (h: {
                          calculated_at: string;
                          trust_score: number;
                          reliability_score?: number;
                          user_rating_score?: number;
                        }) => ({
                          date: h.calculated_at,
                          score: h.trust_score,
                          reliability: h.reliability_score,
                          community: h.user_rating_score,
                        })
                      )}
                      title="Trust Score History"
                      variant="area"
                      showTrend
                      height={200}
                    />
                  </div>
                )}
              </motion.div>

              <motion.div
                id="fp-activity"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="function-page-section function-page-section--delayed-3"
              >
                <div className="flex items-center gap-2 mb-4">
                  <Activity className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
                  <h2 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)' }}>Recent Activity</h2>
                </div>
                <ActivityFeed
                  activities={(executionsData?.executions || []).map(
                    (exec: {
                      execution_id: string;
                      created_at: string;
                      version: string;
                      replay_verified: boolean;
                      roots_match?: boolean;
                    }) => ({
                      id: exec.execution_id,
                      type: exec.replay_verified ? 'success' : 'info',
                      title: `Execution ${exec.version}`,
                      description: exec.roots_match
                        ? 'Verified deterministic'
                        : 'Execution completed',
                      timestamp: exec.created_at,
                    })
                  )}
                  title="Recent Executions"
                  maxItems={5}
                />
              </motion.div>

              {latestExecution && (
                <motion.div
                  id="fp-execution"
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5 }}
                  className="function-page-section function-page-section--delayed-3"
                >
                  <div className="flex items-center gap-2 mb-4">
                    <Zap className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
                    <h2 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)' }}>Latest Execution</h2>
                  </div>
                  <ExecutionTimeline
                    phases={getExecutionPhases(latestExecution, statsData?.avg_latency_ms)}
                    totalDurationMs={
                      statsData?.avg_latency_ms
                        ? 150 + 80 + (statsData.avg_latency_ms as number) + 40 + 15
                        : 405
                    }
                    coldStart={true}
                    showDetails={true}
                  />
                </motion.div>
              )}

              <motion.div
                id="fp-traces"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="function-page-section function-page-section--delayed-3"
              >
                <div className="flex items-center gap-2 mb-4">
                  <Layers className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
                  <h2 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)' }}>Execution Traces</h2>
                </div>
                <TraceList functionFilter={{ author: functionInfo.author, name: functionInfo.name }} />
              </motion.div>

              {versionsData?.versions && versionsData.versions.length > 0 && (
                <motion.div
                  id="fp-versions"
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5 }}
                  className="function-page-section function-page-section--delayed-4"
                >
                  <div className="flex items-center gap-2 mb-4">
                    <Clock className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
                    <h2 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)' }}>Version History</h2>
                  </div>
                  <div className="space-y-3">
                    {versionsData.versions
                      .slice(0, 5)
                      .map(
                        (version, idx) => (
                          <Card
                            key={version.version}
                            className={idx === 0 ? 'sc-fp__version-card--current' : ''}
                          >
                            <div className="sc-fp__version-row">
                              <div className={`sc-fp__version-badge ${idx === 0 ? 'current' : ''}`}>
                                v{version.version}
                              </div>
                              <div className="sc-fp__version-info">
                                <p className="sc-fp__version-name">
                                  Version {version.version}
                                  {idx === 0 && <span className="sc-fp__version-current">(current)</span>}
                                </p>
                                <p className="sc-fp__version-date">
                                  Published {new Date(version.published_at).toLocaleDateString()}
                                </p>
                              </div>
                              {version.changelog && (
                                <p className="sc-fp__version-changelog">{version.changelog}</p>
                              )}
                            </div>
                          </Card>
                        )
                      )}
                  </div>
                </motion.div>
              )}

              {functionInfo.readme && (
                <motion.div
                  id="fp-readme"
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5 }}
                  className="function-page-section function-page-section--delayed-4"
                >
                  <div className="flex items-center gap-2 mb-4">
                    <FileJson className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
                    <h2 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)' }}>README</h2>
                  </div>
                  <Chamber nested>
                      <MarkdownRenderer content={functionInfo.readme} />
                  </Chamber>
                </motion.div>
              )}

              <div id="fp-reviews">
                <ReviewsSection />
              </div>

              <div id="fp-api">
                <ApiReferenceSection functionInfo={functionInfo} />
              </div>

              <div id="fp-source">
                <SourceSection
                  functionInfo={functionInfo}
                  sourceCode={sourceCode}
                  isLoadingSource={isLoadingSource}
                  sourceError={sourceError as Error | null}
                  onRetrySource={refetchSource}
                />
              </div>

              <div id="fp-examples">
                <CodeExamplesSection functionInfo={functionInfo} />
              </div>

              <div id="fp-embed">
                <FunctionEmbedSection
                  author={functionInfo.author}
                  name={functionInfo.name}
                  version={functionInfo.version}
                />
              </div>

              <div id="fp-schema" className="function-page-section">
                <h2 className="text-2xl font-bold mb-4 flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
                  <FileJson className="w-6 h-6" style={{ color: 'var(--status-ok)' }} />
                  Input / Output Schema
                </h2>
                <div className="sc-fp__schema-grid">
                  {functionInfo.manifest?.input ? (
                    <Chamber nested className="sc-fp__schema-card">
                      <h3 className="sc-fp__schema-title">Input</h3>
                      <p className="sc-fp__schema-desc">Expected input structure</p>
                      <CodeBlock code={JSON.stringify(functionInfo.manifest.input, null, 2)} language="json" />
                    </Chamber>
                  ) : (
                    <Chamber nested className="sc-fp__schema-card sc-fp__schema-card--empty">
                      <h3 className="sc-fp__schema-title">Input</h3>
                      <p className="sc-fp__schema-desc">No input schema defined</p>
                      <p className="sc-fp__schema-hint">This function does not declare an input schema. Add an <code>input</code> field to your function manifest to enable typed inputs.</p>
                    </Chamber>
                  )}
                  {functionInfo.manifest?.output ? (
                    <Chamber nested className="sc-fp__schema-card">
                      <h3 className="sc-fp__schema-title">Output</h3>
                      <p className="sc-fp__schema-desc">Expected output structure</p>
                      <CodeBlock code={JSON.stringify(functionInfo.manifest.output, null, 2)} language="json" />
                    </Chamber>
                  ) : (
                    <Chamber nested className="sc-fp__schema-card sc-fp__schema-card--empty">
                      <h3 className="sc-fp__schema-title">Output</h3>
                      <p className="sc-fp__schema-desc">No output schema defined</p>
                      <p className="sc-fp__schema-hint">This function does not declare an output schema. Add an <code>output</code> field to your function manifest to enable typed outputs.</p>
                    </Chamber>
                  )}
                </div>
              </div>

              <RelatedFunctionsSection
                author={functionInfo.author}
                name={functionInfo.name}
                category={functionInfo.category}
                tags={functionInfo.tags}
              />
            </div>
          </div>
        </main>
        <Footer />
      </div>

      {/* Report Issue Dialog */}
      <Modal open={showReportDialog} onClose={() => setShowReportDialog(false)} title="Report a Function Issue">
        <p className="sc-fp__dialog-desc">
          Help @{functionInfo?.author} fix problems with {functionInfo?.name}
        </p>
        <div className="sc-fp__report-field">
          <label htmlFor="report-description-fp" className="sc-fp__report-label">Describe the issue</label>
          <textarea
            id="report-description-fp"
            className="sc-community-textarea"
            value={reportDescription}
            onChange={(e) => setReportDescription(e.target.value)}
            placeholder="The function returns an error when I send X input, or it doesn't work as described..."
            rows={4}
          />
        </div>
        <div className="sc-fp__report-info">
          <p className="sc-fp__report-info-title">What happens next</p>
          <p className="sc-fp__report-info-desc">The author will be notified and can investigate. You'll see updates in the function's activity feed.</p>
        </div>
        <div className="sc-fp__dialog-actions">
          <FrameButton onClick={() => setShowReportDialog(false)}>Cancel</FrameButton>
          <SealedButton onClick={submitReport} disabled={!reportDescription.trim()}>Submit Report</SealedButton>
        </div>
      </Modal>
    </>
    </ErrorBoundary>
  );
}
