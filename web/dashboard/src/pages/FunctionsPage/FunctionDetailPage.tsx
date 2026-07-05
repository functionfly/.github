import { FunctionDetailNotFound } from './FunctionDetailPage/FunctionDetailNotFound';
import { apiClient } from '@/api/client';
import { BarChart } from '@/components/common/BarChart';
import { LineChart } from '@/components/common/LineChart';
import { PieChart } from '@/components/common/PieChart';
import { ProviderIcon } from '@/components/common/ProviderIcon';
import { StatCard } from '@/components/common/StatCard';
import { DNAHelix, DNATrustBadge } from '@/components/dna';
import { FunctionCodeViewer, FunctionHeader } from '@/components/functions';
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
import {
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
  Card,
  Modal,
  PageGrid,
  Input,
} from '@/components/containment';
import './FunctionDetailPage.css';

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
    return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' });
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

function mapToFunctionHeaderData(
  data: FunctionData,
  trustTier: TrustTier = 'high',
  economicScore: number = 87
): FunctionHeaderData {
  const executionRootHash = `0x${data.id.split('').map((c) => c.charCodeAt(0).toString(16)).join('').padEnd(64, '0').slice(0, 64)}`;
  const resourceSignature = `res_sig_${data.id.replace(/[^a-zA-Z0-9]/g, '').slice(0, 16)}`;

  return {
    name: data.name,
    id: data.id,
    executionRootHash,
    trustTier,
    economicScore,
    runtime: data.runtime,
    resourceSignature,
    fxcert: { verified: true, issuedAt: data.createdAt, issuer: 'FunctionFly Registry' },
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

  const [functionData, setFunctionData] = useState<FunctionData | null>(null);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const { data: dnaProfile, isLoading: dnaLoading } = useDNAProfile(id || '');
  const toggleEvolution = useToggleDNAEvolution(id || '');
  const triggerAnalysis = useTriggerDNAAnalysis(id || '');

  useEffect(() => {
    const fetchFunctionData = async () => {
      if (!id) return;
      try {
        setLoading(true);
        setError(null);
        const functionResponse = await apiClient.get<unknown>(`/v1/functions/${id}`);
        const mapped = mapApiFunctionToFunctionData(functionResponse);
        if (!mapped) {
          setError(t('functionDetail.invalidFunctionResponse'));
          setFunctionData(null);
          toast.error(t('functionDetail.couldNotLoadFunctionDetails'));
          return;
        }
        setFunctionData(mapped);
        const deploymentsResponse = await apiClient.get<{ deployments?: unknown }>(`/v1/functions/${id}/deployments`);
        setDeployments(mapApiDeploymentsToDeployments(deploymentsResponse.deployments));
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
    } catch {
      toast.error(t('functionDetail.failedToRedeploy'));
    } finally {
      setIsRedeploying(false);
    }
  };

  const confirmDelete = async () => {
    if (!id || !functionData) return;
    try {
      setIsDeleting(true);
      await apiClient.delete(`/v1/functions/${id}`);
      toast.success(t('functionDetail.deletedSuccessfully', { name: functionData.name }));
      navigate('/functions');
    } catch {
      toast.error(t('functionDetail.failedToDelete'));
    } finally {
      setIsDeleting(false);
      setShowDeleteDialog(false);
    }
  };

  const submitReport = async () => {
    if (!id || !functionData) return;
    try {
      await apiClient.post(`/v1/functions/${id}/report-issue`, {
        description: reportDescription,
        functionName: functionData.name,
        author: functionData.author,
      });
      toast.success(t('functionDetail.issueReported'));
      setShowReportDialog(false);
      setReportDescription('');
    } catch {
      toast.error(t('functionDetail.failedToReportIssue'));
    }
  };

  if (loading) {
    return (
      <div className="sc-fn-detail__page">
        <PageGrid />
        <Chamber nested className="sc-fn-detail__loading">
          <LoadingSpinner />
          <span className="sc-fn-detail__loading-text">Loading function...</span>
        </Chamber>
      </div>
    );
  }

  if (!functionData) {
    return <FunctionDetailNotFound id={id} errorMessage={error ?? undefined} />;
  }

  const statusPill = functionData.status === 'online' ? 'live' : functionData.status === 'offline' ? 'revoked' : 'pending';

  const stats = [
    { title: t('functionDetail.totalRequests'), value: (functionData.requests ?? 0).toLocaleString(), change: { value: 12, label: t('functionDetail.fromLastWeek') }, icon: <Globe size={18} />, trend: 'up' as const },
    { title: t('functionDetail.avgLatency'), value: `${functionData.avgLatency ?? 0}ms`, change: { value: -8, label: t('functionDetail.fromLastWeek') }, icon: <Clock size={18} />, trend: 'up' as const },
    { title: t('functionDetail.errorRate'), value: `${functionData.errorRate ?? 0}%`, change: { value: -0.2, label: t('functionDetail.fromLastWeek') }, icon: <AlertTriangle size={18} />, trend: 'up' as const },
    { title: t('functionDetail.uptime'), value: `${functionData.uptime ?? 0}%`, change: { value: 0.1, label: t('functionDetail.fromLastWeek') }, icon: <Activity size={18} />, trend: 'up' as const },
  ];

  return (
    <div className="sc-fn-detail__page">
      <PageGrid />

      {/* Function Header */}
      <FunctionHeader
        data={mapToFunctionHeaderData(functionData)}
        onBack={() => navigate('/functions')}
        onEdit={() => navigate(`/functions/${id}/edit`)}
        onDeploy={handleRedeploy}
        onTest={() => toast.info(t('functionDetail.testComingSoon'))}
        onShare={() => toast.info(t('functionDetail.shareComingSoon'))}
        onReportIssue={() => setShowReportDialog(true)}
      />

      {/* Function Info */}
      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="FUNCTION · CONFIG" secondary={functionData.version} />
        <div className="sc-fn-detail__info-grid">
          <div className="sc-fn-detail__info-item">
            <span className="sc-fn-detail__info-label">{t('functionDetail.providers')}</span>
            <div className="sc-fn-detail__info-providers">
              {functionData.providers.map((p) => (
                <div key={p} className="sc-fn-detail__provider">
                  <ProviderIcon provider={p} size="sm" />
                  <span>{p}</span>
                </div>
              ))}
            </div>
          </div>
          <div className="sc-fn-detail__info-item">
            <span className="sc-fn-detail__info-label">{t('functionDetail.region')}</span>
            <span className="sc-fn-detail__info-value">{functionData.region.toUpperCase()}</span>
          </div>
          <div className="sc-fn-detail__info-item">
            <span className="sc-fn-detail__info-label">{t('functionDetail.runtime')}</span>
            <span className="sc-fn-detail__info-value">{functionData.runtime}</span>
          </div>
          <div className="sc-fn-detail__info-item">
            <span className="sc-fn-detail__info-label">{t('functionDetail.version')}</span>
            <span className="sc-fn-detail__info-tag">{functionData.version}</span>
          </div>
          <div className="sc-fn-detail__info-item">
            <span className="sc-fn-detail__info-label">{t('functionDetail.lastDeployed')}</span>
            <span className="sc-fn-detail__info-value">{functionData.lastDeployed}</span>
          </div>
          <div className="sc-fn-detail__info-item">
            <span className="sc-fn-detail__info-label">{t('functionDetail.created')}</span>
            <span className="sc-fn-detail__info-value">{functionData.createdAt}</span>
          </div>
        </div>
        <GaugeStrip>
          <Gauge data={{ value: (functionData.requests ?? 0).toLocaleString(), label: 'Requests' }} isFirst />
          <Gauge data={{ value: `${functionData.avgLatency ?? 0}ms`, label: 'Latency' }} />
          <Gauge data={{ value: `${functionData.uptime ?? 0}%`, label: 'Uptime' }} />
          <Gauge data={{ value: `${functionData.errorRate ?? 0}%`, label: 'Errors' }} />
        </GaugeStrip>
      </Chamber>

      {/* Stats Row */}
      <div className="sc-fn-detail__stats-grid">
        {stats.map((stat) => (
          <StatCard key={stat.title} {...stat} />
        ))}
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="sc-fn-detail__tabs-list">
          <TabsTrigger value="overview">{t('functionDetail.overview')}</TabsTrigger>
          <TabsTrigger value="code"><Code2 size={14} /> {t('functionDetail.code')}</TabsTrigger>
          <TabsTrigger value="deployments">{t('functionDetail.deployments')}</TabsTrigger>
          <TabsTrigger value="logs">{t('functionDetail.logs')}</TabsTrigger>
          <TabsTrigger value="analytics">{t('functionDetail.analytics')}</TabsTrigger>
          <TabsTrigger value="dna"><Dna size={14} /> DNA</TabsTrigger>
          <TabsTrigger value="traces"><Layers size={14} /> Traces</TabsTrigger>
        </TabsList>

        {/* Overview */}
        <TabsContent value="overview">
          <div className="sc-fn-detail__charts-grid">
            <LineChart title={t('functionDetail.requestsOverTime')} data={requestData} series={[{ key: 'requests', name: t('functionDetail.requests'), color: '#6366f1' }, { key: 'errors', name: t('functionDetail.errors'), color: '#ef4444' }]} height={300} />
            <BarChart title={t('functionDetail.latencyByProvider')} data={latencyData} series={[{ key: 'latency', name: t('functionDetail.latencyMs'), color: '#10b981' }]} height={300} />
          </div>
        </TabsContent>

        {/* Code */}
        <TabsContent value="code">
          <FunctionCodeViewer code={functionData.code || '// No source code available for this function.'} runtime={functionData.runtime} functionName={functionData.name} version={functionData.version} lastModified={functionData.lastDeployed} onEdit={() => navigate(`/functions/${id}/edit`)} />
        </TabsContent>

        {/* Deployments */}
        <TabsContent value="deployments">
          <Chamber>
            <h2 className="sc-fn-detail__section-title">{t('functionDetail.deploymentHistory')}</h2>
            <div className="sc-fn-detail__deploy-list">
              {deployments.map((dep) => (
                <div key={dep.id} className="sc-fn-detail__deploy-row">
                  {dep.status === 'success' && <CheckCircle2 size={18} style={{ color: 'var(--status-ok)' }} />}
                  {dep.status === 'failed' && <XCircle size={18} style={{ color: 'var(--status-revoked)' }} />}
                  {dep.status === 'pending' && <RotateCcw size={18} className="sc-community-spinner" style={{ color: 'var(--status-pending)' }} />}
                  <div className="sc-fn-detail__deploy-info">
                    <div className="sc-fn-detail__deploy-header">
                      <span className="sc-fn-detail__deploy-version">v{dep.version}</span>
                      <StatusPill status={dep.status === 'success' ? 'live' : dep.status === 'failed' ? 'revoked' : 'pending'} label={dep.status} />
                    </div>
                    <p className="sc-fn-detail__deploy-meta">{dep.timestamp} · {dep.duration}s · by {dep.triggeredBy}</p>
                    {dep.commit && <p className="sc-fn-detail__deploy-commit">{dep.commit}</p>}
                  </div>
                </div>
              ))}
            </div>
          </Chamber>
        </TabsContent>

        {/* Logs */}
        <TabsContent value="logs">
          <Chamber className="sc-fn-detail__logs-chamber">
            <h2 className="sc-fn-detail__section-title">{t('functionDetail.recentLogs')}</h2>
            <ScrollArea className="sc-fn-detail__logs-scroll">
              <div className="sc-fn-detail__logs-list">
                {logs.map((log) => (
                  <div key={log.id} className="sc-fn-detail__log-row">
                    <span className="sc-fn-detail__log-time">{formatLogLineTime(log.timestamp)}</span>
                    <div className="sc-fn-detail__log-content">
                      {log.level === 'error' && <XCircle size={14} style={{ color: 'var(--status-revoked)' }} />}
                      {log.level === 'warn' && <AlertTriangle size={14} style={{ color: 'var(--status-pending)' }} />}
                      {log.level === 'info' && <Activity size={14} style={{ color: 'var(--foil-a)' }} />}
                      <span className="sc-fn-detail__log-msg">{log.message}</span>
                      <span className="sc-fn-detail__log-src">({log.source})</span>
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </Chamber>
        </TabsContent>

        {/* Analytics */}
        <TabsContent value="analytics">
          <div className="sc-fn-detail__charts-grid">
            <LineChart title={t('functionDetail.errorRateOverTime')} data={requestData} series={[{ key: 'errors', name: t('functionDetail.errors'), color: '#ef4444' }]} height={300} />
            <Chamber>
              <h2 className="sc-fn-detail__section-title">{t('functionDetail.errorDistribution')}</h2>
              <div style={{ height: 300 }}><PieChart data={errorData} height={300} /></div>
              <div className="sc-fn-detail__pie-legend">
                {errorData.map((item) => (
                  <div key={item.name} className="sc-fn-detail__pie-legend-item">
                    <div className="sc-fn-detail__pie-dot" style={{ background: item.color }} />
                    <span>{item.name} ({item.value}%)</span>
                  </div>
                ))}
              </div>
            </Chamber>
          </div>
        </TabsContent>

        {/* DNA */}
        <TabsContent value="dna">
          {dnaLoading ? (
            <Chamber nested className="sc-fn-detail__loading"><LoadingSpinner /></Chamber>
          ) : dnaProfile ? (
            <>
              {dnaProfile.generation > 1 && (
                <DNATrustBadge generation={dnaProfile.generation} fitnessScore={dnaProfile.fitness_score} totalMutations={dnaProfile.total_mutations} totalExecutions={dnaProfile.total_executions} variant="full" />
              )}
              <DNAHelix profile={dnaProfile} onToggleEvolution={(enabled) => toggleEvolution.mutate(enabled)} onTriggerAnalysis={() => triggerAnalysis.mutate()} isToggling={toggleEvolution.isPending} isAnalyzing={triggerAnalysis.isPending} />
              <div className="sc-fn-detail__dna-link">
                <FrameButton size="sm" iconLeft={<Dna size={14} />} onClick={() => navigate(`/functions/${id}/dna`)}>Full DNA View</FrameButton>
              </div>
            </>
          ) : (
            <Chamber className="sc-fn-detail__dna-empty">
              <Dna size={36} className="sc-fn-detail__dna-empty-icon" />
              <h3 className="sc-fn-detail__dna-empty-title">DNA Not Enabled</h3>
              <p className="sc-fn-detail__dna-empty-desc">Enable Function DNA to track execution patterns and receive AI-powered evolution suggestions.</p>
              <SealedButton iconLeft={<Dna size={14} />} onClick={() => navigate(`/functions/${id}/dna`)}>Enable DNA</SealedButton>
            </Chamber>
          )}
        </TabsContent>

        {/* Traces */}
        <TabsContent value="traces">
          {functionData?.author ? (
            <TraceList functionFilter={{ author: functionData.author, name: functionData.name }} />
          ) : (
            <Chamber className="sc-fn-detail__dna-empty">
              <Layers size={36} className="sc-fn-detail__dna-empty-icon" />
              <h3 className="sc-fn-detail__dna-empty-title">Atlas Traces</h3>
              <p className="sc-fn-detail__dna-empty-desc">Execution traces are recorded by the Atlas Memory Engine when ATLAS_URL is configured.</p>
            </Chamber>
          )}
        </TabsContent>
      </Tabs>

      {/* Delete Dialog */}
      <Modal open={showDeleteDialog} onClose={() => setShowDeleteDialog(false)} title={t('functionDetail.deleteFunction')}>
        <p className="sc-fn-detail__dialog-desc">{t('functionDetail.deleteFunctionConfirm', { name: functionData?.name })}</p>
        <div className="sc-fn-detail__dialog-actions">
          <FrameButton onClick={() => setShowDeleteDialog(false)} disabled={isDeleting}>{t('functionDetail.cancel')}</FrameButton>
          <SealedButton loading={isDeleting} onClick={confirmDelete} className="sc-fn-detail__delete-confirm">{t('functionDetail.deleteFunction')}</SealedButton>
        </div>
      </Modal>

      {/* Report Dialog */}
      <Modal open={showReportDialog} onClose={() => setShowReportDialog(false)} title={t('functionDetail.reportIssue')}>
        <p className="sc-fn-detail__dialog-desc">{t('functionDetail.reportIssueDescription', { name: functionData?.name })}</p>
        <div className="sc-fn-detail__report-field">
          <label htmlFor="report-description" className="sc-fn-detail__report-label">{t('functionDetail.whatIsTheIssue')}</label>
          <textarea id="report-description" className="sc-community-textarea" value={reportDescription} onChange={(e) => setReportDescription(e.target.value)} placeholder={t('functionDetail.issuePlaceholder')} rows={4} />
        </div>
        <div className="sc-fn-detail__report-disclaimer">
          <p>{t('functionDetail.reportDisclaimer', { functionName: functionData?.name, author: functionData?.author })}</p>
        </div>
        <div className="sc-fn-detail__dialog-actions">
          <FrameButton onClick={() => setShowReportDialog(false)}>{t('functionDetail.cancel')}</FrameButton>
          <SealedButton onClick={submitReport} disabled={!reportDescription.trim()}>{t('functionDetail.submitReport')}</SealedButton>
        </div>
      </Modal>
    </div>
  );
}

export default FunctionDetailPage;
