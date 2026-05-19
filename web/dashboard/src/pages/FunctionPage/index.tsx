import { useEffect, useState } from 'react';
import { FunctionNotFound } from './FunctionNotFound';
import { registryApi } from '@/api/registry';
import { favoritesApi } from '@/api/favorites';
import { ErrorMessage } from '@/components/common/ErrorMessage';
import { Navbar } from '@/components/common/Navbar';
import { FollowFunctionButton } from '@/components/follow';
import { FunctionCard, FunctionEmbedSection, FunctionHeader, TrustScoreBadge } from '@/components/functions';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { CodeBlock } from '@/components/common/CodeBlock';
import { Footer } from '@/pages/LandingPage/components';
import { useAuthStore } from '@/stores/authStore';
import { Icon } from '@iconify/react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Helmet } from 'react-helmet-async';
import { motion } from 'framer-motion';
import { Activity, BarChart3, ChevronRight, Clock, FileJson, Package, Play, Shield, Star, TrendingUp, Zap } from 'lucide-react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { ApiReferenceSection } from './ApiReferenceSection';
import { CodeExamplesSection } from './CodeExamplesSection';
import type { FunctionInfo } from './types';
import { mapToFunctionCardData, mapToFunctionHeaderData, mapToTrustMetrics, mapToDNATrustData } from './utils';
import { ReviewsSection } from './ReviewsSection';
import { ShareButton } from './ShareButton';
import { SourceSection } from './SourceSection';
import { TableOfContents, type TocItem } from './TableOfContents';
import { DNATrustBadge } from '@/components/dna/DNATrustBadge';
import { TrustDashboardWidget } from '@/components/trust/TrustDashboardWidget';
import { TrustHistory } from '@/components/trust/TrustHistory';
import { ActivityFeed } from '@/components/common/ActivityFeed';
import { ExecutionTimeline } from '@/components/execution/ExecutionTimeline';
import type { TrustHistoryDataPoint } from '@/components/trust/TrustHistory';
import { MarkdownRenderer } from './MarkdownRenderer';
import { formatNumber } from '@/lib/utils';

export default function FunctionPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const { data: functionInfo, isLoading, error } = useQuery<FunctionInfo>({
    queryKey: ['function', author, name],
    queryFn: async () => {
      const response = await fetch(`/v1/functions/${author}/${name}?expand=manifest`);
      if (response.status === 404) throw new Error('Function not found');
      if (!response.ok) throw new Error('Failed to fetch function');
      return response.json();
    },
    enabled: !!author && !!name,
  });

  const { data: favoritesResp } = useQuery({
    queryKey: ['favorites-list'],
    queryFn: async () => favoritesApi.list(1, 100),
    enabled: isAuthenticated,
  });

  const isFavorited = favoritesResp?.favorites.some((f) => f.function_id === functionInfo?.id) ?? false;

  const { data: sourceCode, isLoading: isLoadingSource, error: sourceError, refetch: refetchSource } = useQuery({
    queryKey: ['function-source', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      return registryApi.getFunctionVersionSource(author, name, 'latest');
    },
    enabled: !!author && !!name,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  const { data: trustHistoryData } = useQuery({
    queryKey: ['function-trust-history', functionInfo?.id],
    queryFn: async () => {
      if (!functionInfo?.id) return null;
      const response = await fetch(`/v1/platform/functions/${functionInfo.id}/trust/history?page_size=30`);
      if (!response.ok) return null;
      return response.json();
    },
    enabled: !!functionInfo?.id,
    staleTime: 5 * 60 * 1000,
  });

  const { data: executionsData } = useQuery({
    queryKey: ['function-executions', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      const response = await fetch(`/functions/${author}/${name}/executions?limit=10`);
      if (!response.ok) return null;
      return response.json();
    },
    enabled: !!author && !!name,
    staleTime: 30 * 1000,
  });

  const { data: versionsData } = useQuery({
    queryKey: ['function-versions', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      const response = await fetch(`/functions/${author}/${name}/versions`);
      if (!response.ok) return null;
      return response.json();
    },
    enabled: !!author && !!name,
    staleTime: 5 * 60 * 1000,
  });

  const { data: statsData } = useQuery({
    queryKey: ['function-stats', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      const response = await fetch(`/functions/${author}/${name}/stats`);
      if (!response.ok) return null;
      return response.json();
    },
    enabled: !!author && !!name,
    staleTime: 30 * 1000,
  });

  const latestExecution = executionsData?.executions?.[0];

  const getExecutionPhases = (exec: { execution_id: string; created_at: string; version: string; replay_verified: boolean; roots_match?: boolean; determinism_tier?: string }, avgLatencyMs: number | undefined) => {
    const execMs = avgLatencyMs ?? 120;
    const phase2End = 150;
    const phase3End = phase2End + 80;
    const phase4End = phase3End + execMs;
    const phase5End = phase4End + 40;
    const phase6End = phase5End + 15;
    return [
      { id: '1', name: 'Container Start', durationMs: 150, startTime: 0, endTime: 150, status: 'completed' as const },
      { id: '2', name: 'Runtime Init', durationMs: 80, startTime: 150, endTime: phase3End, status: 'completed' as const },
      { id: '3', name: 'Function Execution', durationMs: execMs, startTime: phase3End, endTime: phase4End, status: 'completed' as const, metadata: { 'avg_latency_ms': execMs, determinism_tier: exec.determinism_tier ?? 'full' } },
      { id: '4', name: 'Verification', durationMs: 40, startTime: phase4End, endTime: phase5End, status: exec.replay_verified ? 'completed' as const : 'failed' as const, metadata: { verified: exec.replay_verified ? 1 : 0, deterministic: (exec.roots_match ?? false) ? 1 : 0 } },
      { id: '5', name: 'Response', durationMs: 15, startTime: phase5End, endTime: phase6End, status: 'completed' as const },
    ];
  };

  if (isLoading) {
    const isNotFound = error != null && (error as any) instanceof Error && (error as Error).message === 'Function not found';
    if (isNotFound || !functionInfo) {
      return (
        <div className="function-page">
          <Navbar variant="landing" />
          <main className="flex-1 pt-16">
            <FunctionNotFound author={author} name={name} />
          </main>
          <Footer />
        </div>
      );
    }
    return (
      <div className="function-page">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16 flex items-center justify-center">
          <ErrorMessage error={error as Error} />
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
    <>
      <Helmet>
        <title>{functionInfo.title || functionInfo.name} by {functionInfo.author} | FunctionFly</title>
        <meta name="description" content={functionInfo.description || `Explore ${functionInfo.name} on FunctionFly`} />
        <meta property="og:title" content={`${functionInfo.title || functionInfo.name} | FunctionFly`} />
        <meta property="og:description" content={functionInfo.description || `Function ${functionInfo.name} by ${functionInfo.author}`} />
        <meta property="og:type" content="website" />
        <meta name="twitter:card" content="summary" />
        <meta name="twitter:title" content={`${functionInfo.title || functionInfo.name} | FunctionFly`} />
        <meta name="twitter:description" content={functionInfo.description || `Function ${functionInfo.name} by ${functionInfo.author}`} />
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
              />

              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4 }}
                className="function-page-cta"
              >
                <Link to={`/run/${functionInfo.author}/${functionInfo.name}`}>
                  <Button size="lg" className="function-page-cta-button function-page-cta-button--primary gap-2 px-8">
                    <Play className="w-4 h-4" />
                    Try it Now
                  </Button>
                </Link>
                <Link to={`/registry/${functionInfo.author}/${functionInfo.name}/executions`}>
                  <Button variant="outline" size="lg" className="function-page-cta-button function-page-cta-button--secondary gap-2">
                    <Activity className="w-4 h-4" />
                    Executions
                  </Button>
                </Link>
                <Link to={`/registry/${functionInfo.author}/${functionInfo.name}/executions?tab=certificates`}>
                  <Button variant="outline" size="lg" className="function-page-cta-button function-page-cta-button--secondary gap-2">
                    <Shield className="w-4 h-4" />
                    Certificates
                  </Button>
                </Link>
                {functionInfo.repo_url ? (
                  <a href={functionInfo.repo_url} target="_blank" rel="noopener noreferrer">
                    <Button variant="outline" size="lg" className="function-page-cta-button function-page-cta-button--secondary gap-2">
                      <Icon icon="simple-icons:github" className="w-4 h-4" />
                      View on GitHub
                    </Button>
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
                  document.getElementById('fp-api')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
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
                <BarChart3 className="h-6 w-6 text-brand-500" />
                <h2 className="text-2xl font-bold">Function Statistics</h2>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <Card className="bg-gradient-to-br from-brand-500/10 to-brand-600/5 border-brand-500/20">
                  <CardContent className="p-4 text-center">
                    <Activity className="h-8 w-8 text-brand-500 mx-auto mb-2" />
                    <p className="text-2xl font-bold text-text-primary">
                      {statsData?.total_calls ? formatNumber(statsData.total_calls) : formatNumber(functionInfo.executions || 0)}
                    </p>
                    <p className="text-xs text-text-muted">Total Executions</p>
                  </CardContent>
                </Card>
                <Card className="bg-gradient-to-br from-emerald-500/10 to-emerald-600/5 border-emerald-500/20">
                  <CardContent className="p-4 text-center">
                    <TrendingUp className="h-8 w-8 text-emerald-500 mx-auto mb-2" />
                    <p className="text-2xl font-bold text-text-primary">
                      {statsData?.success_rate ? `${statsData.success_rate.toFixed(1)}%` : '99.9%'}
                    </p>
                    <p className="text-xs text-text-muted">Success Rate</p>
                  </CardContent>
                </Card>
                <Card className="bg-gradient-to-br from-amber-500/10 to-amber-600/5 border-amber-500/20">
                  <CardContent className="p-4 text-center">
                    <Zap className="h-8 w-8 text-amber-500 mx-auto mb-2" />
                    <p className="text-2xl font-bold text-text-primary">
                      {statsData?.avg_latency_ms ? `${Math.round(statsData.avg_latency_ms)}ms` : '<50ms'}
                    </p>
                    <p className="text-xs text-text-muted">Avg Latency</p>
                  </CardContent>
                </Card>
                <Card className="bg-gradient-to-br from-violet-500/10 to-violet-600/5 border-violet-500/20">
                  <CardContent className="p-4 text-center">
                    <Shield className="h-8 w-8 text-violet-500 mx-auto mb-2" />
                    <p className="text-2xl font-bold text-text-primary">
                      {functionInfo.trust_score ? `${functionInfo.trust_score}%` : '95%'}
                    </p>
                    <p className="text-xs text-text-muted">Trust Score</p>
                  </CardContent>
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
                  <h2 className="text-2xl font-bold flex items-center gap-2">
                    <span className="text-velocity-500">⚡</span>
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
                <Shield className="h-6 w-6 text-brand-500" />
                <h2 className="text-2xl font-bold">Trust & Verification</h2>
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
                    data={trustHistoryData.history.map((h: { calculated_at: string; trust_score: number; reliability_score?: number; user_rating_score?: number }) => ({
                      date: h.calculated_at,
                      score: h.trust_score,
                      reliability: h.reliability_score,
                      community: h.user_rating_score,
                    }))}
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
                <Activity className="h-6 w-6 text-brand-500" />
                <h2 className="text-2xl font-bold">Recent Activity</h2>
              </div>
              <ActivityFeed
                activities={(executionsData?.executions || []).map((exec: { execution_id: string; created_at: string; version: string; replay_verified: boolean; roots_match?: boolean }) => ({
                  id: exec.execution_id,
                  type: exec.replay_verified ? 'success' : 'info',
                  title: `Execution ${exec.version}`,
                  description: exec.roots_match ? 'Verified deterministic' : 'Execution completed',
                  timestamp: exec.created_at,
                }))}
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
                  <Zap className="h-6 w-6 text-brand-500" />
                  <h2 className="text-2xl font-bold">Latest Execution</h2>
                </div>
                <ExecutionTimeline
                  phases={getExecutionPhases(latestExecution, statsData?.avg_latency_ms)}
                  totalDurationMs={statsData?.avg_latency_ms ? 150 + 80 + (statsData.avg_latency_ms as number) + 40 + 15 : 405}
                  coldStart={true}
                  showDetails={true}
                />
              </motion.div>
            )}

            {versionsData?.versions && versionsData.versions.length > 0 && (
              <motion.div
                id="fp-versions"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="function-page-section function-page-section--delayed-4"
              >
                <div className="flex items-center gap-2 mb-4">
                  <Clock className="h-6 w-6 text-brand-500" />
                  <h2 className="text-2xl font-bold">Version History</h2>
                </div>
                <div className="space-y-3">
                  {versionsData.versions.slice(0, 5).map((version: { version: string; published_at: string; changelog?: string }, idx: number) => (
                    <Card key={version.version} className={idx === 0 ? 'border-brand-500/50 bg-brand-500/5' : ''}>
                      <CardContent className="p-4">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ${idx === 0 ? 'bg-brand-500 text-white' : 'bg-bg-secondary text-text-muted'}`}>
                              v{version.version}
                            </div>
                            <div>
                              <p className="text-sm font-medium text-text-primary">
                                Version {version.version}
                                {idx === 0 && <span className="ml-2 text-xs text-brand-400 font-normal">(current)</span>}
                              </p>
                              <p className="text-xs text-text-muted">
                                Published {new Date(version.published_at).toLocaleDateString()}
                              </p>
                            </div>
                          </div>
                          {version.changelog && (
                            <p className="text-xs text-text-secondary line-clamp-1 max-w-xs">{version.changelog}</p>
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
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
                  <FileJson className="h-6 w-6 text-brand-500" />
                  <h2 className="text-2xl font-bold">README</h2>
                </div>
                <Card>
                  <CardContent className="p-6">
                    <MarkdownRenderer content={functionInfo.readme} />
                  </CardContent>
                </Card>
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
              <h2 className="text-2xl font-bold mb-4 flex items-center gap-2">
                <FileJson className="w-6 h-6 text-brand-500" />
                Input / Output Schema
              </h2>
              <div className="function-page-schema-grid">
                {functionInfo.manifest?.input ? (
                  <Card className="function-page-schema-card">
                    <CardHeader className="pb-3">
                      <CardTitle className="text-lg">Input</CardTitle>
                      <CardDescription>Expected input structure</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <CodeBlock code={JSON.stringify(functionInfo.manifest.input, null, 2)} language="json" />
                    </CardContent>
                  </Card>
                ) : (
                  <Card className="function-page-schema-card border-dashed border-border-muted">
                    <CardHeader className="pb-3">
                      <CardTitle className="text-lg">Input</CardTitle>
                      <CardDescription>No input schema defined</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <p className="text-sm text-text-muted">
                        This function does not declare an input schema. Add an <code>input</code> field to your function manifest to enable typed inputs.
                      </p>
                    </CardContent>
                  </Card>
                )}
                {functionInfo.manifest?.output ? (
                  <Card className="function-page-schema-card">
                    <CardHeader className="pb-3">
                      <CardTitle className="text-lg">Output</CardTitle>
                      <CardDescription>Expected output structure</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <CodeBlock code={JSON.stringify(functionInfo.manifest.output, null, 2)} language="json" />
                    </CardContent>
                  </Card>
                ) : (
                  <Card className="function-page-schema-card border-dashed border-border-muted">
                    <CardHeader className="pb-3">
                      <CardTitle className="text-lg">Output</CardTitle>
                      <CardDescription>No output schema defined</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <p className="text-sm text-text-muted">
                        This function does not declare an output schema. Add an <code>output</code> field to your function manifest to enable typed outputs.
                      </p>
                    </CardContent>
                  </Card>
                )}
              </div>
            </div>

            <Card className="function-page-related-cta">
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-xl bg-brand-500/10 flex items-center justify-center">
                    <Package className="w-6 h-6 text-brand-500" />
                  </div>
                  <div className="flex-1">
                    <h3 className="font-semibold text-lg">Explore More Functions</h3>
                    <p className="text-muted-foreground text-sm">
                      Discover related functions in the registry to build powerful workflows
                    </p>
                  </div>
                  <Link to="/registry">
                    <Button variant="outline">
                      Browse Registry
                      <ChevronRight className="w-4 h-4 ml-1" />
                    </Button>
                  </Link>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
      <Footer />
      </div>
    </>
  );
}