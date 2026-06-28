import { FunctionDetailNotFound } from './FunctionDetailPage/FunctionDetailNotFound';
import { apiClient } from '@/api/client';
import { BarChart } from '@/components/common/BarChart';
import { LineChart } from '@/components/common/LineChart';
import { PieChart } from '@/components/common/PieChart';
import { ProviderIcon } from '@/components/common/ProviderIcon';
import { StatCard } from '@/components/common/StatCard';
import { DNAHelix, DNATrustBadge } from '@/components/dna';
import { FunctionCodeViewer, FunctionHeader } from '@/components/functions';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  useDNAProfile,
  useToggleDNAEvolution,
  useTriggerDNAAnalysis,
} from '@/hooks/useFunctionDNA';
import '@/styles/components.css';
import type { FunctionHeaderData, TrustTier } from '@/types';
import { TraceList } from '@/components/atlas';
import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Clock,
  Code2,
  Dna,
  Globe,
  Layers,
  RotateCcw,
  XCircle,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';

interface FunctionData {
  id: string;
  name: string;
  author?: string;
  status: 'online' | 'offline' | 'degraded';
  providers: string[];
  region: string;
  lastDeployed: string;
  createdAt: string;
  version: string;
  runtime: string;
  code: string;
  requests: number;
  avgLatency: number;
  errorRate: number;
  uptime: number;
}

interface Deployment {
  id: string;
  version: string;
  status: 'success' | 'failed' | 'pending';
  timestamp: string;
  duration: number;
  triggeredBy: string;
  commit?: string;
}

interface LogEntry {
  id: string;
  timestamp: string;
  level: 'info' | 'warn' | 'error';
  message: string;
  source: string;
}

function asRecord(v: unknown): Record<string, unknown> | null {
  return v !== null && typeof v === 'object' && !Array.isArray(v)
    ? (v as Record<string, unknown>)
    : null;
}

function asStr(v: unknown, fallback = ''): string {
  return typeof v === 'string' ? v : fallback;
}

function asNum(v: unknown, fallback = 0): number {
  if (typeof v === 'number' && !Number.isNaN(v)) return v;
  if (typeof v === 'string') {
    const n = parseFloat(v);
    return Number.isNaN(n) ? fallback : n;
  }
  return fallback;
}

function asStrArray(v: unknown): string[] {
  if (!Array.isArray(v)) return [];
  return v.filter((x): x is string => typeof x === 'string');
}

function formatTs(v: unknown): string {
  if (v == null) return '—';
  const d = new Date(v as string | number | Date);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString();
}

function mapApiStatusToUi(status: string): FunctionData['status'] {
  const s = status.toLowerCase();
  if (s === 'deployed') return 'online';
  if (s === 'deploying') return 'degraded';
  if (s === 'failed') return 'offline';
  if (s === 'draft') return 'offline';
  return 'offline';
}

/** GET /v1/functions/:id returns storage.FunctionConfig (snake_case); map to UI model with safe defaults. */
function mapApiFunctionToFunctionData(raw: unknown): FunctionData | null {
  const api = asRecord(raw);
  if (!api) return null;
  const id = asStr(api.id);
  const name = asStr(api.name);
  if (!id || !name) return null;

  const providers = asStrArray(api.providers);
  const region = asStr(api.region, 'global');
  const version = asStr(api.version, '0.0.0');
  const statusRaw = asStr(api.status, 'draft');
  const createdAt = formatTs(api.created_at ?? api.createdAt);
  const updatedRaw = api.updated_at ?? api.updatedAt;

  return {
    id,
    name,
    status: mapApiStatusToUi(statusRaw),
    providers: providers.length > 0 ? providers : ['functionfly'],
    region,
    lastDeployed: formatTs(updatedRaw),
    createdAt,
    version,
    runtime: asStr(api.runtime, '—'),
    code: asStr(api.code, ''),
    requests: asNum(api.requests, 0),
    avgLatency: asNum(api.avg_latency ?? api.avgLatency, 0),
    errorRate: asNum(api.error_rate ?? api.errorRate, 0),
    uptime: asNum(api.uptime, 100),
  };
}

function mapApiDeploymentsToDeployments(raw: unknown): Deployment[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => {
    const d = asRecord(item) ?? {};
    const st = asStr(d.status, 'pending');
    const status: Deployment['status'] =
      st === 'success' || st === 'failed' || st === 'pending' ? st : 'pending';
    return {
      id: asStr(d.id),
      version: asStr(d.version, '—'),
      status,
      timestamp: formatTs(d.created_at ?? d.timestamp),
      duration: asNum(d.duration, 0),
      triggeredBy: asStr(d.triggered_by ?? d.triggeredBy, 'dashboard'),
      commit: typeof d.commit === 'string' ? d.commit : undefined,
    };
  });
}

function mapApiLogsToLogEntries(raw: unknown): LogEntry[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => {
    const l = asRecord(item) ?? {};
    const lev = asStr(l.level, 'info');
    const level: LogEntry['level'] =
      lev === 'warn' || lev === 'error' || lev === 'info' ? lev : 'info';
    return {
      id: asStr(l.id),
      timestamp: formatTs(l.timestamp),
      level,
      message: asStr(l.message, ''),
      source: asStr(l.source, 'runtime'),
    };
  });
}

function formatLogLineTime(ts: string): string {
  const d = new Date(ts);
  if (!Number.isNaN(d.getTime())) {
    return d.toLocaleTimeString(undefined, {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  }
  const space = ts.indexOf(' ');
  if (space > 0) return ts.slice(space + 1).trim() || ts;
  const t = ts.includes('T') ? ts.split('T')[1] : ts;
  return (t ?? ts).replace(/\.\d+Z?$/, '').replace('Z', '') || '—';
}

const requestData = [
  { name: '00:00', requests: 120, errors: 1 },
  { name: '04:00', requests: 98, errors: 0 },
  { name: '08:00', requests: 245, errors: 2 },
  { name: '12:00', requests: 189, errors: 1 },
  { name: '16:00', requests: 312, errors: 3 },
  { name: '20:00', requests: 156, errors: 1 },
];

const latencyData = [
  { name: 'Workers', latency: 45, color: '#f48120' },
  { name: 'Vercel', latency: 62, color: '#000000' },
];

const errorData = [
  { name: '4xx', value: 65, color: '#f59e0b' },
  { name: '5xx', value: 25, color: '#ef4444' },
  { name: 'Timeout', value: 10, color: '#8b5cf6' },
];

/**
 * Map function data to FunctionHeaderData format
 * Uses mock data for fields not available from the API
 */
function mapToFunctionHeaderData(
  data: FunctionData,
  trustTier: TrustTier = 'high',
  economicScore: number = 87
): FunctionHeaderData {
  // Generate a mock execution root hash based on function id
  const executionRootHash = `0x${data.id
    .split('')
    .map((c) => c.charCodeAt(0).toString(16))
    .join('')
    .padEnd(64, '0')
    .slice(0, 64)}`;

  // Generate a mock resource signature
  const resourceSignature = `res_sig_${data.id.replace(/[^a-zA-Z0-9]/g, '').slice(0, 16)}`;

  return {
    name: data.name,
    id: data.id,
    executionRootHash,
    trustTier,
    economicScore,
    runtime: data.runtime,
    resourceSignature,
    fxcert: {
      verified: true,
      issuedAt: data.createdAt,
      issuer: 'FunctionFly Registry',
    },
    status: data.status,
    version: data.version,
    description: `Function deployed across ${data.providers.join(', ')} in ${data.region.toUpperCase()} region`,
  };
}

export function FunctionDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('overview');
  const [isRedeploying, setIsRedeploying] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showReportDialog, setShowReportDialog] = useState(false);
  const [reportDescription, setReportDescription] = useState('');

  // State for API data
  const [functionData, setFunctionData] = useState<FunctionData | null>(null);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // DNA hooks
  const { data: dnaProfile, isLoading: dnaLoading } = useDNAProfile(id || '');
  const toggleEvolution = useToggleDNAEvolution(id || '');
  const triggerAnalysis = useTriggerDNAAnalysis(id || '');

  // Fetch function data
  useEffect(() => {
    const fetchFunctionData = async () => {
      if (!id) return;

      try {
        setLoading(true);
        setError(null);

        // Fetch function details (API returns FunctionConfig JSON, not the legacy mock shape)
        const functionResponse = await apiClient.get<unknown>(`/v1/functions/${id}`);
        const mapped = mapApiFunctionToFunctionData(functionResponse);
        if (!mapped) {
          setError(t('functionDetail.invalidFunctionResponse'));
          setFunctionData(null);
          toast.error(t('functionDetail.couldNotLoadFunctionDetails'));
          return;
        }
        setFunctionData(mapped);

        // Fetch deployments
        const deploymentsResponse = await apiClient.get<{ deployments?: unknown }>(
          `/v1/functions/${id}/deployments`
        );
        setDeployments(mapApiDeploymentsToDeployments(deploymentsResponse.deployments));

        // Fetch logs
        const logsResponse = await apiClient.get<{ logs?: unknown }>(`/v1/functions/${id}/logs`);
        setLogs(mapApiLogsToLogEntries(logsResponse.logs));
      } catch (err) {
        console.error('Failed to load function data:', err);
        setError(t('functionDetail.failedToLoadFunctionData'));
        toast.error(t('functionDetail.failedToLoadFunctionData'));
      } finally {
        setLoading(false);
      }
    };

    fetchFunctionData();
  }, [id, t]);

  const handleRedeploy = async () => {
    if (!id) return;
    setIsRedeploying(true);
    try {
      await apiClient.post(`/v1/functions/${id}/redeploy`);
      toast.success(t('functionDetail.redeployedSuccessfully'));
    } catch (error) {
      toast.error(t('functionDetail.failedToRedeploy'));
    } finally {
      setIsRedeploying(false);
    }
  };

  const handleDelete = () => {
    setShowDeleteDialog(true);
  };

  const confirmDelete = async () => {
    if (!id || !functionData) return;

    try {
      setIsDeleting(true);
      await apiClient.delete(`/v1/functions/${id}`);
      toast.success(t('functionDetail.deletedSuccessfully', { name: functionData.name }));
      navigate('/functions');
    } catch (error) {
      console.error('Failed to delete function:', error);
      toast.error(t('functionDetail.failedToDelete'));
    } finally {
      setIsDeleting(false);
      setShowDeleteDialog(false);
    }
  };

  const handleReportIssue = () => {
    setShowReportDialog(true);
  };

  const submitReport = async () => {
    if (!id || !functionData) return;

    try {
      // Submit issue report to the API
      await apiClient.post(`/v1/functions/${id}/report-issue`, {
        description: reportDescription,
        functionName: functionData.name,
        author: functionData.author,
      });
      toast.success(t('functionDetail.issueReported'));
      setShowReportDialog(false);
      setReportDescription('');
    } catch (error) {
      console.error('Failed to report issue:', error);
      toast.error(t('functionDetail.failedToReportIssue'));
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 bg-muted rounded animate-pulse" />
          <div className="space-y-2">
            <div className="w-48 h-6 bg-muted rounded animate-pulse" />
            <div className="w-32 h-4 bg-muted rounded animate-pulse" />
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="p-6 border rounded-lg">
              <div className="w-16 h-16 bg-muted rounded animate-pulse mb-4" />
              <div className="w-20 h-4 bg-muted rounded animate-pulse mb-2" />
              <div className="w-16 h-6 bg-muted rounded animate-pulse" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!functionData) {
    return <FunctionDetailNotFound id={id} errorMessage={error ?? undefined} />;
  }

  const stats = [
    {
      title: t('functionDetail.totalRequests'),
      value: (functionData.requests ?? 0).toLocaleString(),
      change: { value: 12, label: t('functionDetail.fromLastWeek') },
      icon: <Globe className="w-5 h-5" />,
      trend: 'up' as const,
    },
    {
      title: t('functionDetail.avgLatency'),
      value: `${functionData.avgLatency ?? 0}ms`,
      change: { value: -8, label: t('functionDetail.fromLastWeek') },
      icon: <Clock className="w-5 h-5" />,
      trend: 'up' as const,
    },
    {
      title: t('functionDetail.errorRate'),
      value: `${functionData.errorRate ?? 0}%`,
      change: { value: -0.2, label: t('functionDetail.fromLastWeek') },
      icon: <AlertTriangle className="w-5 h-5" />,
      trend: 'up' as const,
    },
    {
      title: t('functionDetail.uptime'),
      value: `${functionData.uptime ?? 0}%`,
      change: { value: 0.1, label: t('functionDetail.fromLastWeek') },
      icon: <Activity className="w-5 h-5" />,
      trend: 'up' as const,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Function Header */}
      <FunctionHeader
        data={mapToFunctionHeaderData(functionData)}
        onBack={() => navigate('/functions')}
        onEdit={() => navigate(`/functions/${id}/edit`)}
        onDeploy={handleRedeploy}
        onTest={() => toast.info(t('functionDetail.testComingSoon'))}
        onShare={() => toast.info(t('functionDetail.shareComingSoon'))}
        onReportIssue={handleReportIssue}
      />

      {/* Function Info Card */}
      <Card className="card">
        <CardContent className="card-content p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">
                {t('functionDetail.providers')}
              </h3>
              <div className="flex items-center gap-2">
                {functionData.providers.map((provider) => (
                  <div key={provider} className="flex items-center gap-2">
                    <ProviderIcon provider={provider} size="sm" />
                    <span className="text-sm text-text-primary capitalize">{provider}</span>
                  </div>
                ))}
              </div>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">
                {t('functionDetail.region')}
              </h3>
              <p className="text-text-primary">{functionData.region.toUpperCase()}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">
                {t('functionDetail.runtime')}
              </h3>
              <p className="text-text-primary">{functionData.runtime}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">
                {t('functionDetail.version')}
              </h3>
              <Badge variant="secondary">{functionData.version}</Badge>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">
                {t('functionDetail.lastDeployed')}
              </h3>
              <p className="text-text-primary">{functionData.lastDeployed}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-text-secondary mb-2">
                {t('functionDetail.created')}
              </h3>
              <p className="text-text-primary">{functionData.createdAt}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Row */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <StatCard key={stat.title} {...stat} />
        ))}
      </div>

      {/* Main Content Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-7">
          <TabsTrigger value="overview">{t('functionDetail.overview')}</TabsTrigger>
          <TabsTrigger value="code" className="gap-1.5">
            <Code2 className="h-3.5 w-3.5" />
            {t('functionDetail.code')}
          </TabsTrigger>
          <TabsTrigger value="deployments">{t('functionDetail.deployments')}</TabsTrigger>
          <TabsTrigger value="logs">{t('functionDetail.logs')}</TabsTrigger>
          <TabsTrigger value="analytics">{t('functionDetail.analytics')}</TabsTrigger>
          <TabsTrigger value="dna" className="gap-1.5">
            <Dna className="h-3.5 w-3.5" />
            DNA
          </TabsTrigger>
          <TabsTrigger value="traces" className="gap-1.5">
            <Layers className="h-3.5 w-3.5" />
            Traces
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <LineChart
              title={t('functionDetail.requestsOverTime')}
              data={requestData}
              series={[
                { key: 'requests', name: t('functionDetail.requests'), color: '#6366f1' },
                { key: 'errors', name: t('functionDetail.errors'), color: '#ef4444' },
              ]}
              height={300}
            />

            <BarChart
              title={t('functionDetail.latencyByProvider')}
              data={latencyData}
              series={[{ key: 'latency', name: t('functionDetail.latencyMs'), color: '#10b981' }]}
              height={300}
            />
          </div>
        </TabsContent>

        <TabsContent value="code" className="space-y-4">
          <FunctionCodeViewer
            code={functionData.code || '// No source code available for this function.'}
            runtime={functionData.runtime}
            functionName={functionData.name}
            version={functionData.version}
            lastModified={functionData.lastDeployed}
            onEdit={() => navigate(`/functions/${id}/edit`)}
          />
        </TabsContent>

        <TabsContent value="deployments" className="space-y-4">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">{t('functionDetail.deploymentHistory')}</CardTitle>
            </CardHeader>
            <CardContent className="card-content">
              <div className="space-y-4">
                {deployments.map((deployment) => (
                  <div
                    key={deployment.id}
                    className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary"
                  >
                    <div className="flex items-center gap-4">
                      {deployment.status === 'success' && (
                        <CheckCircle2 className="w-5 h-5 text-green-400" />
                      )}
                      {deployment.status === 'failed' && (
                        <XCircle className="w-5 h-5 text-red-400" />
                      )}
                      {deployment.status === 'pending' && (
                        <RotateCcw className="w-5 h-5 text-yellow-400 animate-spin" />
                      )}

                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-text-primary">
                            v{deployment.version}
                          </span>
                          <Badge
                            variant={deployment.status === 'success' ? 'default' : 'destructive'}
                            className="text-xs"
                          >
                            {deployment.status}
                          </Badge>
                        </div>
                        <p className="text-sm text-text-secondary">
                          {deployment.timestamp} • {deployment.duration}s • by{' '}
                          {deployment.triggeredBy}
                        </p>
                        {deployment.commit && (
                          <p className="text-xs text-text-muted font-mono">{deployment.commit}</p>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="logs" className="space-y-4">
          <Card className="card h-[600px]">
            <CardHeader className="card-header">
              <CardTitle className="card-title">{t('functionDetail.recentLogs')}</CardTitle>
            </CardHeader>
            <CardContent className="card-content p-0">
              <ScrollArea className="h-[520px] p-6">
                <div className="space-y-3">
                  {logs.map((log) => (
                    <div
                      key={log.id}
                      className="flex items-start gap-3 p-3 rounded-lg bg-bg-tertiary"
                    >
                      <div className="text-text-muted font-mono text-xs w-28 shrink-0">
                        {formatLogLineTime(log.timestamp)}
                      </div>
                      <div className="flex items-center gap-2 flex-1">
                        {log.level === 'error' && <XCircle className="w-4 h-4 text-red-400" />}
                        {log.level === 'warn' && (
                          <AlertTriangle className="w-4 h-4 text-yellow-400" />
                        )}
                        {log.level === 'info' && <Activity className="w-4 h-4 text-blue-400" />}
                        <div className="flex-1">
                          <span className="text-sm text-text-primary">{log.message}</span>
                          <span className="text-xs text-text-muted ml-2">({log.source})</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="analytics" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <LineChart
              title={t('functionDetail.errorRateOverTime')}
              data={requestData}
              series={[{ key: 'errors', name: t('functionDetail.errors'), color: '#ef4444' }]}
              height={300}
            />

            <Card className="card">
              <CardHeader className="card-header">
                <CardTitle className="card-title">
                  {t('functionDetail.errorDistribution')}
                </CardTitle>
              </CardHeader>
              <CardContent className="card-content">
                <div className="h-[300px]">
                  <PieChart data={errorData} height={300} />
                </div>
                <div className="flex justify-center gap-6 mt-4">
                  {errorData.map((item) => (
                    <div key={item.name} className="flex items-center gap-2">
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: item.color }}
                      />
                      <span className="text-sm text-text-secondary">
                        {item.name} ({item.value}%)
                      </span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="dna" className="space-y-6">
          {dnaLoading ? (
            <div className="flex items-center justify-center h-64">
              <LoadingSpinner />
            </div>
          ) : dnaProfile ? (
            <>
              {dnaProfile.generation > 1 && (
                <DNATrustBadge
                  generation={dnaProfile.generation}
                  fitnessScore={dnaProfile.fitness_score}
                  totalMutations={dnaProfile.total_mutations}
                  totalExecutions={dnaProfile.total_executions}
                  variant="full"
                />
              )}
              <DNAHelix
                profile={dnaProfile}
                onToggleEvolution={(enabled) => toggleEvolution.mutate(enabled)}
                onTriggerAnalysis={() => triggerAnalysis.mutate()}
                isToggling={toggleEvolution.isPending}
                isAnalyzing={triggerAnalysis.isPending}
              />
              <div className="flex justify-end">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => navigate(`/functions/${id}/dna`)}
                  className="gap-1.5"
                >
                  <Dna className="h-3.5 w-3.5" />
                  Full DNA View
                </Button>
              </div>
            </>
          ) : (
            <Card className="card">
              <CardContent className="card-content flex flex-col items-center justify-center py-12">
                <Dna className="h-10 w-10 text-text-muted mb-4" />
                <h3 className="text-lg font-semibold text-text-primary mb-2">DNA Not Enabled</h3>
                <p className="text-sm text-text-secondary mb-4 text-center max-w-md">
                  Enable Function DNA to track execution patterns and receive AI-powered evolution
                  suggestions.
                </p>
                <Button onClick={() => navigate(`/functions/${id}/dna`)} className="gap-1.5">
                  <Dna className="h-4 w-4" />
                  Enable DNA
                </Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="traces" className="space-y-4">
          {functionData?.author ? (
            <TraceList functionFilter={{ author: functionData.author, name: functionData.name }} />
          ) : (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Layers className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
                <h3 className="text-lg font-semibold mb-2">Atlas Traces</h3>
                <p className="text-sm text-muted-foreground text-center max-w-md">
                  Execution traces are recorded by the Atlas Memory Engine when ATLAS_URL is configured.
                </p>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('functionDetail.deleteFunction')}</DialogTitle>
            <DialogDescription>
              {t('functionDetail.deleteFunctionConfirm', { name: functionData?.name })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDeleteDialog(false)}
              disabled={isDeleting}
            >
              {t('functionDetail.cancel')}
            </Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={isDeleting}>
              {isDeleting ? t('functionDetail.deleting') : t('functionDetail.deleteFunction')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Report Issue Dialog */}
      <Dialog open={showReportDialog} onOpenChange={setShowReportDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-amber-500" />
              {t('functionDetail.reportIssue')}
            </DialogTitle>
            <DialogDescription>
              {t('functionDetail.reportIssueDescription', { name: functionData?.name })}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <label
                htmlFor="report-description"
                className="block text-sm font-medium text-text-primary mb-2"
              >
                {t('functionDetail.whatIsTheIssue')}
              </label>
              <textarea
                id="report-description"
                value={reportDescription}
                onChange={(e) => setReportDescription(e.target.value)}
                placeholder={t('functionDetail.issuePlaceholder')}
                className="w-full min-h-[120px] px-3 py-2 rounded-lg border border-border-subtle bg-bg-primary text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-amber-500/50 resize-none"
              />
            </div>
            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-3">
              <p className="text-xs text-amber-400">
                {t('functionDetail.reportDisclaimer', {
                  functionName: functionData?.name,
                  author: functionData?.author,
                })}
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowReportDialog(false)}>
              {t('functionDetail.cancel')}
            </Button>
            <Button
              variant="default"
              onClick={submitReport}
              disabled={!reportDescription.trim()}
              className="bg-amber-500 hover:bg-amber-600 text-white border-0"
            >
              {t('functionDetail.submitReport')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
