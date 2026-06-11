/**
 * @functionfly/ui-devops
 * DevOps & Infrastructure Components - Full implementation
 */

import React, { useState, useCallback, useMemo } from 'react';
import { cn } from '@functionfly/ui-core';
import {
  GitBranch,
  GitCommit,
  Play,
  Pause,
  RotateCcw,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Clock,
  Zap,
  Server,
  Container,
  Globe,
  MapPin,
  Boxes,
  Cloud,
  Key,
  Shield,
  Eye,
  EyeOff,
  Copy,
  Check,
  Plus,
  Minus,
  Settings,
  Trash2,
  Edit3,
  Save,
  RefreshCw,
  Download,
  Upload,
  ArrowRight,
  ArrowUpRight,
  ArrowDownRight,
  ChevronRight,
  ChevronDown,
  Search,
  Filter,
  SortAsc,
  Activity,
  Cpu,
  HardDrive,
  MemoryStick,
  Network,
  Database,
  Rocket,
  TrafficCone,
  AlertTriangle,
  Info,
  Bug,
  Gauge,
  Layers,
  Scaling,
  Route,
  RotateCw,
  FlaskConical,
  Package,
  Box,
  FileCode,
  FileText,
  History,
  TrendingUp,
  BarChart3,
  PieChart,
  LineChart,
  Map,
  Users,
  User,
  Lock,
  Unlock,
  DatabaseIcon,
  LayersIcon,
  ServerIcon,
  MonitorIcon,
  Microscope,
  Webhook,
  GitFork,
  Circle,
  Square,
  Hexagon,
  type LucideIcon,
} from 'lucide-react';

// ============================================================================
// Deployment Pipeline
// ============================================================================

interface PipelineStage {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped' | 'waiting';
  duration?: number;
  startedAt?: number;
  completedAt?: number;
  artifacts?: Array<{ name: string; type: string; url: string; size?: number }>;
  tasks: Array<{ id: string; name: string; status: string; duration?: number; logs?: string[]; error?: string }>;
}

interface Pipeline {
  id: string;
  name: string;
  version: string;
  status: 'active' | 'paused' | 'archived';
  stages: PipelineStage[];
  currentStageId?: string;
  triggeredBy: string;
  triggeredAt: number;
  branch: string;
  commitSha: string;
  source: 'manual' | 'webhook' | 'scheduled' | 'api';
}

interface DeploymentPipelineProps {
  pipeline: Pipeline;
  selectedStageId?: string | null;
  onStageSelect?: (stage: PipelineStage) => void;
  onStageRetry?: (stageId: string) => void;
  onPipelinePause?: () => void;
  onPipelineResume?: () => void;
  showLogs?: boolean;
  className?: string;
}

export const DeploymentPipeline: React.FC<DeploymentPipelineProps> = ({
  pipeline,
  selectedStageId,
  onStageSelect,
  onStageRetry,
  onPipelinePause,
  onPipelineResume,
  showLogs = true,
  className,
}) => {
  const [expandedStageId, setExpandedStageId] = useState<string | null>(null);
  const [activeLogTab, setActiveLogTab] = useState<'stdout' | 'stderr'>('stdout');

  const getStageIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircle2 className="w-4 h-4 text-green-400" />;
      case 'failed': return <XCircle className="w-4 h-4 text-red-400" />;
      case 'running': return <RefreshCw className="w-4 h-4 text-cyan-400 animate-spin" />;
      case 'waiting': return <Clock className="w-4 h-4 text-amber-400" />;
      default: return <Circle className="w-4 h-4 text-aviation-text-muted" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-green-400 bg-green-500/20';
      case 'failed': return 'text-red-400 bg-red-500/20';
      case 'running': return 'text-cyan-400 bg-cyan-500/20';
      case 'waiting': return 'text-amber-400 bg-amber-500/20';
      default: return 'text-gray-400 bg-gray-500/20';
    }
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return '--';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}m`;
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-aviation-cyan/20 rounded">
            <Rocket className="w-5 h-5 text-aviation-cyan" />
          </div>
          <div>
            <div className="text-sm font-medium">{pipeline.name}</div>
            <div className="flex items-center gap-2 text-xs text-aviation-text-muted mt-0.5">
              <GitBranch className="w-3 h-3" />
              <span>{pipeline.branch}</span>
              <span>·</span>
              <GitCommit className="w-3 h-3" />
              <span className="font-mono">{pipeline.commitSha?.substring(0, 7)}</span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className={cn('px-2 py-1 rounded text-xs', getStatusColor(pipeline.status))}>
            {pipeline.status}
          </span>
          {pipeline.status === 'active' ? (
            <button onClick={onPipelinePause} className="p-2 hover:bg-aviation-bg-instrument rounded">
              <Pause className="w-4 h-4" />
            </button>
          ) : pipeline.status === 'paused' ? (
            <button onClick={onPipelineResume} className="p-2 hover:bg-aviation-bg-instrument rounded">
              <Play className="w-4 h-4" />
            </button>
          ) : null}
        </div>
      </div>

      {/* Pipeline Stages */}
      <div className="flex-1 overflow-auto p-4">
        <div className="relative">
          {/* Connector Line */}
          <div className="absolute left-6 top-0 bottom-0 w-0.5 bg-aviation-border-panel" />

          <div className="space-y-4">
            {pipeline.stages.map((stage, idx) => {
              const isSelected = stage.id === selectedStageId;
              const isExpanded = expandedStageId === stage.id;
              const isLast = idx === pipeline.stages.length - 1;
              
              return (
                <div key={stage.id} className="relative flex gap-4">
                  {/* Stage Node */}
                  <div
                    className={cn(
                      'relative z-10 flex-shrink-0 w-12 h-12 rounded-full border-2 flex items-center justify-center cursor-pointer transition-all',
                      isSelected ? 'border-aviation-cyan bg-aviation-cyan/20' : 'border-aviation-border-panel bg-aviation-bg-panel hover:border-aviation-text-muted',
                      stage.status === 'running' && 'animate-pulse'
                    )}
                    onClick={() => onStageSelect?.(stage)}
                  >
                    {getStageIcon(stage.status)}
                  </div>

                  {/* Stage Content */}
                  <div className="flex-1 min-w-0">
                    <div
                      className={cn(
                        'flex items-center justify-between p-3 rounded-lg border cursor-pointer transition-colors',
                        isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                      )}
                      onClick={() => setExpandedStageId(isExpanded ? null : stage.id)}
                    >
                      <div className="flex items-center gap-3">
                        <span className="text-sm font-medium">{stage.name}</span>
                        <span className={cn('px-1.5 py-0.5 rounded text-[10px]', getStatusColor(stage.status))}>
                          {stage.status}
                        </span>
                      </div>
                      <div className="flex items-center gap-3 text-xs text-aviation-text-muted">
                        <span>{formatDuration(stage.duration)}</span>
                        {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                      </div>
                    </div>

                    {/* Expanded Tasks */}
                    {isExpanded && stage.tasks.length > 0 && (
                      <div className="mt-2 ml-14 space-y-2">
                        {stage.tasks.map((task) => (
                          <div
                            key={task.id}
                            className="flex items-center gap-3 px-3 py-2 bg-aviation-bg-secondary rounded"
                          >
                            {getStageIcon(task.status)}
                            <span className="flex-1 text-xs">{task.name}</span>
                            <span className="text-xs text-aviation-text-muted">
                              {formatDuration(task.duration)}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Logs Panel */}
      {showLogs && selectedStageId && (
        <div className="h-64 border-t border-aviation-border-panel flex flex-col">
          <div className="flex items-center gap-4 px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
            <span className="text-xs font-medium">Logs</span>
            <div className="flex items-center gap-1">
              {(['stdout', 'stderr'] as const).map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveLogTab(tab)}
                  className={cn(
                    'px-2 py-1 text-xs rounded capitalize',
                    activeLogTab === tab ? 'bg-aviation-cyan/20 text-aviation-cyan' : 'text-aviation-text-muted'
                  )}
                >
                  {tab}
                </button>
              ))}
            </div>
          </div>
          <div className="flex-1 overflow-auto p-3 font-mono text-xs text-aviation-text-primary">
            {pipeline.stages.find(s => s.id === selectedStageId)?.tasks[0]?.logs?.map((log, i) => (
              <div key={i} className="leading-6">{log}</div>
            )) || <span className="text-aviation-text-muted">No logs available</span>}
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Environment Manager
// ============================================================================

interface Environment {
  id: string;
  name: string;
  type: 'development' | 'staging' | 'production' | 'preview';
  color: string;
  variables: Record<string, string>;
  secrets: Array<{ key: string; masked: boolean; lastUpdated: number }>;
  replicas: number;
  autoScale: boolean;
  region?: string;
}

interface EnvironmentManagerProps {
  environments: Environment[];
  activeEnvironmentId?: string | null;
  onEnvironmentSelect?: (env: Environment) => void;
  onEnvironmentCreate?: (env: Partial<Environment>) => void;
  onEnvironmentUpdate?: (envId: string, updates: Partial<Environment>) => void;
  onEnvironmentDelete?: (envId: string) => void;
  onVariableAdd?: (envId: string, key: string, value: string) => void;
  onVariableUpdate?: (envId: string, key: string, value: string) => void;
  onVariableDelete?: (envId: string, key: string) => void;
  onSecretAdd?: (envId: string, key: string) => void;
  onSecretDelete?: (envId: string, key: string) => void;
  className?: string;
}

export const EnvironmentManager: React.FC<EnvironmentManagerProps> = ({
  environments,
  activeEnvironmentId,
  onEnvironmentSelect,
  onEnvironmentCreate,
  onEnvironmentUpdate,
  onEnvironmentDelete,
  onVariableAdd,
  onVariableUpdate,
  onVariableDelete,
  onSecretDelete,
  className,
}) => {
  const [newVarKey, setNewVarKey] = useState('');
  const [newVarValue, setNewVarValue] = useState('');
  const [secretRevealed, setSecretRevealed] = useState<Record<string, boolean>>({});

  const activeEnv = environments.find(e => e.id === activeEnvironmentId);

  const getEnvTypeColor = (type: string) => {
    switch (type) {
      case 'production': return 'text-red-400 bg-red-500/20';
      case 'staging': return 'text-amber-400 bg-amber-500/20';
      case 'preview': return 'text-purple-400 bg-purple-500/20';
      default: return 'text-green-400 bg-green-500/20';
    }
  };

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Environment List */}
      <div className="w-72 flex flex-col border-r border-aviation-border-panel">
        <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
          <span className="text-sm font-medium">Environments</span>
          <button
            onClick={() => onEnvironmentCreate?.({})}
            className="p-1.5 hover:bg-aviation-bg-instrument rounded"
          >
            <Plus className="w-4 h-4" />
          </button>
        </div>
        <div className="flex-1 overflow-auto">
          {environments.map((env) => (
            <div
              key={env.id}
              className={cn(
                'flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-aviation-border-panel transition-colors',
                activeEnvironmentId === env.id ? 'bg-aviation-cyan/10 border-l-2 border-l-aviation-cyan' : 'hover:bg-aviation-bg-secondary'
              )}
              onClick={() => onEnvironmentSelect?.(env)}
            >
              <div className="w-3 h-3 rounded-full" style={{ backgroundColor: env.color }} />
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium">{env.name}</div>
                <span className={cn('px-1.5 py-0.5 rounded text-[10px]', getEnvTypeColor(env.type))}>
                  {env.type}
                </span>
              </div>
              {env.region && (
                <span className="text-[10px] text-aviation-text-muted">{env.region}</span>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Environment Details */}
      <div className="flex-1 flex flex-col">
        {activeEnv ? (
          <>
            <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
              <div className="flex items-center gap-3">
                <div className="w-4 h-4 rounded-full" style={{ backgroundColor: activeEnv.color }} />
                <span className="text-sm font-medium">{activeEnv.name}</span>
                <span className={cn('px-1.5 py-0.5 rounded text-[10px]', getEnvTypeColor(activeEnv.type))}>
                  {activeEnv.type}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => onEnvironmentUpdate?.(activeEnv.id, { autoScale: !activeEnv.autoScale })}
                  className={cn(
                    'p-2 rounded transition-colors',
                    activeEnv.autoScale ? 'text-aviation-cyan bg-aviation-cyan/20' : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
                  )}
                >
                  <Scaling className="w-4 h-4" />
                </button>
                <button
                  onClick={() => onEnvironmentDelete?.(activeEnv.id)}
                  className="p-2 text-red-400 hover:bg-red-500/20 rounded"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Variables & Secrets */}
            <div className="flex-1 overflow-auto p-4">
              <div className="mb-6">
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium text-aviation-text-muted">VARIABLES</span>
                </div>
                <div className="space-y-2">
                  {Object.entries(activeEnv.variables).map(([key, value]) => (
                    <div key={key} className="flex items-center gap-2 px-3 py-2 bg-aviation-bg-secondary rounded">
                      <code className="flex-1 text-xs font-mono text-aviation-cyan">{key}</code>
                      <code className="flex-1 text-xs font-mono text-aviation-text-primary truncate">{value}</code>
                      <button onClick={() => onVariableDelete?.(activeEnv.id, key)} className="p-1 hover:bg-aviation-bg-instrument rounded">
                        <Trash2 className="w-3 h-3" />
                      </button>
                    </div>
                  ))}
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      value={newVarKey}
                      onChange={(e) => setNewVarKey(e.target.value)}
                      placeholder="KEY"
                      className="flex-1 px-2 py-1.5 bg-aviation-bg-secondary border border-aviation-border-panel rounded text-xs font-mono"
                    />
                    <input
                      type="text"
                      value={newVarValue}
                      onChange={(e) => setNewVarValue(e.target.value)}
                      placeholder="value"
                      className="flex-1 px-2 py-1.5 bg-aviation-bg-secondary border border-aviation-border-panel rounded text-xs font-mono"
                    />
                    <button
                      onClick={() => { if (newVarKey && newVarValue) { onVariableAdd?.(activeEnv.id, newVarKey, newVarValue); setNewVarKey(''); setNewVarValue(''); } }}
                      className="p-1.5 hover:bg-aviation-bg-instrument rounded"
                    >
                      <Plus className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium text-aviation-text-muted">SECRETS</span>
                </div>
                <div className="space-y-2">
                  {activeEnv.secrets.map((secret) => (
                    <div key={secret.key} className="flex items-center gap-2 px-3 py-2 bg-aviation-bg-secondary rounded">
                      <Lock className="w-3 h-3 text-aviation-text-muted" />
                      <code className="flex-1 text-xs font-mono text-aviation-amber">{secret.key}</code>
                      <span className="text-xs text-aviation-text-muted">
                        {secretRevealed[secret.key] ? '••••••••' : '••••••••'}
                      </span>
                      <button
                        onClick={() => setSecretRevealed(prev => ({ ...prev, [secret.key]: !prev[secret.key] }))}
                        className="p-1 hover:bg-aviation-bg-instrument rounded"
                      >
                        {secretRevealed[secret.key] ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                      </button>
                      <button onClick={() => onSecretDelete?.(activeEnv.id, secret.key)} className="p-1 hover:bg-aviation-bg-instrument rounded">
                        <Trash2 className="w-3 h-3" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <Server className="w-8 h-8 text-aviation-text-muted mx-auto mb-3" />
              <p className="text-sm text-aviation-text-muted">Select an environment</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Cloud Region Selector
// ============================================================================

interface CloudRegion {
  id: string;
  name: string;
  provider: 'aws' | 'gcp' | 'azure' | 'custom';
  zone: string;
  zoneName: string;
  location: string;
  country: string;
  coordinates: { lat: number; lng: number };
  isAvailable: boolean;
  isRecommended?: boolean;
  specs?: { compute?: number; memory?: number; storage?: number; gpu?: boolean };
}

interface CloudRegionSelectorProps {
  regions: CloudRegion[];
  selectedRegionId?: string | null;
  selectedProvider?: 'aws' | 'gcp' | 'azure' | 'all';
  onRegionSelect?: (region: CloudRegion) => void;
  onProviderFilter?: (provider: 'aws' | 'gcp' | 'azure' | 'all') => void;
  showRegionStats?: boolean;
  className?: string;
}

export const CloudRegionSelector: React.FC<CloudRegionSelectorProps> = ({
  regions,
  selectedRegionId,
  selectedProvider = 'all',
  onRegionSelect,
  onProviderFilter,
  showRegionStats = true,
  className,
}) => {
  const providers = ['all', 'aws', 'gcp', 'azure'] as const;

  const getProviderIcon = (provider: string) => {
    switch (provider) {
      case 'aws': return <span className="text-amber-400 font-bold">AWS</span>;
      case 'gcp': return <span className="text-blue-400 font-bold">GCP</span>;
      case 'azure': return <span className="text-blue-500 font-bold">AZ</span>;
      default: return <Globe className="w-4 h-4" />;
    }
  };

  const filteredRegions = useMemo(() => {
    return selectedProvider === 'all' ? regions : regions.filter(r => r.provider === selectedProvider);
  }, [regions, selectedProvider]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Provider Filter */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        {providers.map((provider) => (
          <button
            key={provider}
            onClick={() => onProviderFilter?.(provider)}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium transition-colors capitalize',
              selectedProvider === provider ? 'bg-aviation-cyan/20 text-aviation-cyan' : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
            )}
          >
            {getProviderIcon(provider)}
            {provider === 'all' ? 'All Providers' : provider}
          </button>
        ))}
      </div>

      {/* Region Grid */}
      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(200px,1fr))] gap-3">
          {filteredRegions.map((region) => {
            const isSelected = region.id === selectedRegionId;
            return (
              <div
                key={region.id}
                className={cn(
                  'p-3 rounded-lg border cursor-pointer transition-all',
                  isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted',
                  !region.isAvailable && 'opacity-50'
                )}
                onClick={() => region.isAvailable && onRegionSelect?.(region)}
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <MapPin className="w-4 h-4 text-aviation-text-muted" />
                    <span className="text-sm font-medium">{region.zoneName}</span>
                  </div>
                  {region.isRecommended && (
                    <span className="px-1.5 py-0.5 bg-aviation-cyan/20 text-aviation-cyan rounded text-[10px]">Recommended</span>
                  )}
                </div>
                <div className="text-xs text-aviation-text-muted mb-2">
                  {region.location}, {region.country}
                </div>
                <div className="flex items-center gap-2 mb-2">
                  {getProviderIcon(region.provider)}
                  <span className="text-[10px] text-aviation-text-dim">{region.zone}</span>
                </div>
                {showRegionStats && region.specs && (
                  <div className="flex items-center gap-3 text-[10px] text-aviation-text-muted">
                    {region.specs.compute && <span>{region.specs.compute} vCPU</span>}
                    {region.specs.memory && <span>{region.specs.memory}GB RAM</span>}
                    {region.specs.gpu && <span className="text-aviation-amber">GPU</span>}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Runtime Target Selector
// ============================================================================

interface RuntimeTarget {
  id: string;
  name: string;
  type: 'nodejs' | 'python' | 'go' | 'rust' | 'kotlin' | 'ruby' | 'deno' | 'bun' | 'sar';
  version: string;
  status: 'stable' | 'beta' | 'deprecated';
  memoryLimit?: number;
  timeout?: number;
}

interface RuntimeTargetSelectorProps {
  targets: RuntimeTarget[];
  selectedTargetId?: string | null;
  onTargetSelect?: (target: RuntimeTarget) => void;
  showVersionHistory?: boolean;
  className?: string;
}

export const RuntimeTargetSelector: React.FC<RuntimeTargetSelectorProps> = ({
  targets,
  selectedTargetId,
  onTargetSelect,
  className,
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'stable': return 'text-green-400 bg-green-500/20';
      case 'beta': return 'text-amber-400 bg-amber-500/20';
      case 'deprecated': return 'text-red-400 bg-red-500/20';
      default: return 'text-gray-400 bg-gray-500/20';
    }
  };

  const getRuntimeIcon = (type: string) => {
    switch (type) {
      case 'nodejs': return <span className="text-green-400 font-mono text-xs">JS</span>;
      case 'python': return <span className="text-blue-400 font-mono text-xs">PY</span>;
      case 'go': return <span className="text-cyan-400 font-mono text-xs">GO</span>;
      case 'rust': return <span className="text-orange-400 font-mono text-xs">RS</span>;
      case 'kotlin': return <span className="text-purple-400 font-mono text-xs">KT</span>;
      case 'ruby': return <span className="text-red-400 font-mono text-xs">RB</span>;
      case 'deno': return <span className="text-gray-400 font-mono text-xs">DN</span>;
      case 'bun': return <span className="text-amber-400 font-mono text-xs">BN</span>;
      case 'sar': return <span className="text-aviation-cyan font-mono text-xs">SAR</span>;
      default: return <Box className="w-4 h-4" />;
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <span className="text-sm font-medium">Runtime Targets</span>
      </div>
      <div className="flex-1 overflow-auto p-3 space-y-2">
        {targets.map((target) => {
          const isSelected = target.id === selectedTargetId;
          return (
            <div
              key={target.id}
              className={cn(
                'flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors',
                isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
              )}
              onClick={() => onTargetSelect?.(target)}
            >
              <div className="p-2 bg-aviation-bg-secondary rounded">
                {getRuntimeIcon(target.type)}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">{target.name}</span>
                  <span className={cn('px-1.5 py-0.5 rounded text-[10px]', getStatusColor(target.status))}>
                    {target.status}
                  </span>
                </div>
                <div className="flex items-center gap-3 text-xs text-aviation-text-muted mt-1">
                  <span>v{target.version}</span>
                  {target.memoryLimit && <span>{target.memoryLimit}MB</span>}
                  {target.timeout && <span>{target.timeout}s timeout</span>}
                </div>
              </div>
              {isSelected && <CheckCircle2 className="w-5 h-5 text-aviation-cyan" />}
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Kubernetes Topology View
// ============================================================================

interface K8sNode {
  id: string;
  name: string;
  type: 'control-plane' | 'worker' | 'ingress' | 'storage';
  status: 'ready' | 'not-ready' | 'unknown';
  resources?: { cpu?: number; memory?: number; pods?: number };
}

interface K8sService {
  id: string;
  name: string;
  type: 'cluster-ip' | 'node-port' | 'load-balancer' | 'external-name';
  selector: Record<string, string>;
  ports?: Array<{ name: string; port: number; targetPort: number }>;
}

interface K8sNamespace {
  id: string;
  name: string;
  status: 'active' | 'terminating';
  pods: string[];
  services: string[];
}

interface ContainerTopologyViewProps {
  nodes: K8sNode[];
  services: K8sService[];
  namespaces: K8sNamespace[];
  selectedNodeId?: string | null;
  onNodeSelect?: (node: K8sNode) => void;
  layout?: 'hierarchical' | 'force' | 'grid';
  className?: string;
}

export const ContainerTopologyView: React.FC<ContainerTopologyViewProps> = ({
  nodes,
  services,
  namespaces,
  selectedNodeId,
  onNodeSelect,
  className,
}) => {
  const getNodeIcon = (type: string) => {
    switch (type) {
      case 'control-plane': return <Hexagon className="w-5 h-5 text-aviation-cyan" />;
      case 'worker': return <ServerIcon className="w-5 h-5 text-green-400" />;
      case 'ingress': return <Route className="w-5 h-5 text-amber-400" />;
      case 'storage': return <DatabaseIcon className="w-5 h-5 text-purple-400" />;
      default: return <Box className="w-5 h-5" />;
    }
  };

  const getNodeStatusColor = (status: string) => {
    switch (status) {
      case 'ready': return 'text-green-400';
      case 'not-ready': return 'text-red-400';
      default: return 'text-gray-400';
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <Cloud className="w-5 h-5 text-aviation-cyan" />
          <span className="text-sm font-medium">Kubernetes Topology</span>
        </div>
        <span className="text-xs text-aviation-text-muted">{nodes.length} nodes</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        {namespaces.map((ns) => (
          <div key={ns.id} className="mb-4">
            <div className="flex items-center gap-2 px-2 py-1 bg-aviation-bg-secondary rounded mb-2">
              <span className="text-xs font-medium">{ns.name}</span>
              <span className="text-[10px] text-aviation-text-muted">{ns.pods.length} pods · {ns.services.length} services</span>
            </div>
            <div className="grid grid-cols-[repeat(auto-fill,minmax(150px,1fr))] gap-2">
              {nodes.filter((_, i) => i % namespaces.length === namespaces.indexOf(ns) || namespaces.length === 1).slice(0, 6).map((node) => {
                const isSelected = node.id === selectedNodeId;
                return (
                  <div
                    key={node.id}
                    className={cn(
                      'flex items-center gap-2 p-2 rounded border cursor-pointer transition-colors',
                      isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                    )}
                    onClick={() => onNodeSelect?.(node)}
                  >
                    {getNodeIcon(node.type)}
                    <div className="flex-1 min-w-0">
                      <div className="text-xs font-medium truncate">{node.name}</div>
                      <div className="flex items-center gap-1 text-[10px] text-aviation-text-muted">
                        <span className={getNodeStatusColor(node.status)}>{node.status}</span>
                        {node.resources?.cpu && <span>{node.resources.cpu}% CPU</span>}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Edge Deployment Map
// ============================================================================

interface EdgeLocation {
  id: string;
  name: string;
  provider: string;
  region: string;
  city: string;
  country: string;
  status: 'online' | 'offline' | 'degraded';
  latency?: number;
  capacity?: { current: number; max: number };
}

interface EdgeDeploymentMapProps {
  locations: EdgeLocation[];
  selectedLocationId?: string | null;
  onLocationSelect?: (location: EdgeLocation) => void;
  className?: string;
}

export const EdgeDeploymentMap: React.FC<EdgeDeploymentMapProps> = ({
  locations,
  selectedLocationId,
  onLocationSelect,
  className,
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'bg-green-500';
      case 'offline': return 'bg-red-500';
      case 'degraded': return 'bg-amber-500';
      default: return 'bg-gray-500';
    }
  };

  const groupedByRegion = useMemo(() => {
    const groups: Record<string, EdgeLocation[]> = {};
    locations.forEach((loc) => {
      if (!groups[loc.region]) groups[loc.region] = [];
      groups[loc.region].push(loc);
    });
    return groups;
  }, [locations]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Globe className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Edge Deployment Map</span>
        <span className="text-xs text-aviation-text-muted ml-auto">{locations.length} locations</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        {Object.entries(groupedByRegion).map(([region, locs]) => (
          <div key={region} className="mb-4">
            <div className="text-xs font-medium text-aviation-text-muted mb-2">{region}</div>
            <div className="grid grid-cols-[repeat(auto-fill,minmax(180px,1fr))] gap-2">
              {locs.map((loc) => {
                const isSelected = loc.id === selectedLocationId;
                return (
                  <div
                    key={loc.id}
                    className={cn(
                      'p-3 rounded-lg border cursor-pointer transition-colors',
                      isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                    )}
                    onClick={() => onLocationSelect?.(loc)}
                  >
                    <div className="flex items-center gap-2 mb-2">
                      <div className={cn('w-2 h-2 rounded-full', getStatusColor(loc.status))} />
                      <span className="text-xs font-medium truncate">{loc.city}</span>
                    </div>
                    <div className="text-[10px] text-aviation-text-muted mb-2">{loc.country} · {loc.provider}</div>
                    <div className="flex items-center justify-between text-[10px]">
                      {loc.latency && <span className="text-aviation-cyan">{loc.latency}ms</span>}
                      {loc.capacity && (
                        <span className="text-aviation-text-muted">{loc.capacity.current}/{loc.capacity.max}</span>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Container Lifecycle Panel
// ============================================================================

interface Container {
  id: string;
  name: string;
  image: string;
  status: 'running' | 'paused' | 'stopped' | 'restarting' | 'exited';
  ports?: Array<{ host: number; container: number }>;
  usage?: { cpu?: number; memory?: number };
}

interface ContainerLifecyclePanelProps {
  containers: Container[];
  selectedContainerId?: string | null;
  onContainerSelect?: (container: Container) => void;
  onContainerStart?: (containerId: string) => void;
  onContainerStop?: (containerId: string) => void;
  onContainerRestart?: (containerId: string) => void;
  className?: string;
}

export const ContainerLifecyclePanel: React.FC<ContainerLifecyclePanelProps> = ({
  containers,
  selectedContainerId,
  onContainerSelect,
  onContainerStart,
  onContainerStop,
  onContainerRestart,
  className,
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'text-green-400 bg-green-500/20';
      case 'paused': return 'text-amber-400 bg-amber-500/20';
      case 'stopped': case 'exited': return 'text-red-400 bg-red-500/20';
      case 'restarting': return 'text-cyan-400 bg-cyan-500/20';
      default: return 'text-gray-400 bg-gray-500/20';
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Container className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Containers</span>
        <span className="text-xs text-aviation-text-muted ml-auto">{containers.length}</span>
      </div>
      <div className="flex-1 overflow-auto">
        {containers.map((container) => {
          const isSelected = container.id === selectedContainerId;
          return (
            <div
              key={container.id}
              className={cn(
                'px-4 py-3 border-b border-aviation-border-panel',
                isSelected ? 'bg-aviation-cyan/10' : 'hover:bg-aviation-bg-secondary cursor-pointer'
              )}
              onClick={() => onContainerSelect?.(container)}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-medium">{container.name}</span>
                <span className={cn('px-1.5 py-0.5 rounded text-[10px]', getStatusColor(container.status))}>
                  {container.status}
                </span>
              </div>
              <div className="text-xs text-aviation-text-muted font-mono truncate">{container.image}</div>
              {container.usage && (
                <div className="flex items-center gap-3 mt-2 text-[10px] text-aviation-text-muted">
                  {container.usage.cpu && <span>CPU: {container.usage.cpu}%</span>}
                  {container.usage.memory && <span>MEM: {container.usage.memory}%</span>}
                </div>
              )}
              <div className="flex items-center gap-1 mt-2">
                {container.status === 'stopped' || container.status === 'exited' ? (
                  <button onClick={(e) => { e.stopPropagation(); onContainerStart?.(container.id); }} className="p-1.5 hover:bg-green-500/20 rounded text-green-400">
                    <Play className="w-3 h-3" />
                  </button>
                ) : (
                  <button onClick={(e) => { e.stopPropagation(); onContainerStop?.(container.id); }} className="p-1.5 hover:bg-red-500/20 rounded text-red-400">
                    <Pause className="w-3 h-3" />
                  </button>
                )}
                <button onClick={(e) => { e.stopPropagation(); onContainerRestart?.(container.id); }} className="p-1.5 hover:bg-aviation-bg-instrument rounded">
                  <RefreshCw className="w-3 h-3" />
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Secret Vault Manager
// ============================================================================

interface SecretVault {
  id: string;
  name: string;
  type: 'aws-secrets-manager' | 'gcp-secret-manager' | 'azure-keyvault' | 'hashicorp-vault' | 'local';
  status: 'connected' | 'disconnected' | 'error';
  secrets: Array<{ key: string; masked: boolean; lastRotated?: number }>;
}

interface SecretVaultManagerProps {
  vaults: SecretVault[];
  selectedVaultId?: string | null;
  onVaultSelect?: (vault: SecretVault) => void;
  onSecretCreate?: (vaultId: string, key: string, value: string) => void;
  onSecretDelete?: (vaultId: string, key: string) => void;
  className?: string;
}

export const SecretVaultManager: React.FC<SecretVaultManagerProps> = ({
  vaults,
  selectedVaultId,
  onVaultSelect,
  onSecretCreate,
  onSecretDelete,
  className,
}) => {
  const [newSecretKey, setNewSecretKey] = useState('');
  const [newSecretValue, setNewSecretValue] = useState('');

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected': return 'text-green-400';
      case 'disconnected': return 'text-gray-400';
      case 'error': return 'text-red-400';
      default: return 'text-gray-400';
    }
  };

  const getVaultIcon = (type: string) => {
    switch (type) {
      case 'aws-secrets-manager': return <span className="text-amber-400 font-bold">AWS</span>;
      case 'gcp-secret-manager': return <span className="text-blue-400 font-bold">GCP</span>;
      case 'azure-keyvault': return <span className="text-blue-500 font-bold">AZ</span>;
      case 'hashicorp-vault': return <span className="text-red-400 font-bold">HV</span>;
      default: return <Lock className="w-4 h-4" />;
    }
  };

  const selectedVault = vaults.find(v => v.id === selectedVaultId);

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Vault List */}
      <div className="w-64 flex flex-col border-r border-aviation-border-panel">
        <div className="px-4 py-3 border-b border-aviation-border-panel">
          <span className="text-sm font-medium">Vaults</span>
        </div>
        <div className="flex-1 overflow-auto">
          {vaults.map((vault) => (
            <div
              key={vault.id}
              className={cn(
                'flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-aviation-border-panel',
                selectedVaultId === vault.id ? 'bg-aviation-cyan/10' : 'hover:bg-aviation-bg-secondary'
              )}
              onClick={() => onVaultSelect?.(vault)}
            >
              {getVaultIcon(vault.type)}
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium truncate">{vault.name}</div>
                <div className="flex items-center gap-1 text-[10px]">
                  <span className={getStatusColor(vault.status)}>{vault.status}</span>
                  <span>·</span>
                  <span>{vault.secrets.length} secrets</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Secrets List */}
      <div className="flex-1 flex flex-col">
        {selectedVault ? (
          <>
            <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
              <div className="flex items-center gap-2">
                {getVaultIcon(selectedVault.type)}
                <span className="text-sm font-medium">{selectedVault.name}</span>
              </div>
            </div>
            <div className="flex-1 overflow-auto p-4">
              <div className="space-y-2">
                {selectedVault.secrets.map((secret) => (
                  <div key={secret.key} className="flex items-center gap-3 px-3 py-2 bg-aviation-bg-secondary rounded">
                    <Lock className="w-4 h-4 text-aviation-amber" />
                    <code className="flex-1 text-xs font-mono text-aviation-text-primary">{secret.key}</code>
                    <span className="text-xs text-aviation-text-muted">••••••••</span>
                    <button onClick={() => onSecretDelete?.(selectedVault.id, secret.key)} className="p-1 hover:bg-red-500/20 rounded">
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                ))}
                <div className="flex items-center gap-2 mt-4">
                  <input
                    type="text"
                    value={newSecretKey}
                    onChange={(e) => setNewSecretKey(e.target.value)}
                    placeholder="SECRET_KEY"
                    className="flex-1 px-2 py-1.5 bg-aviation-bg-secondary border border-aviation-border-panel rounded text-xs font-mono"
                  />
                  <input
                    type="password"
                    value={newSecretValue}
                    onChange={(e) => setNewSecretValue(e.target.value)}
                    placeholder="value"
                    className="flex-1 px-2 py-1.5 bg-aviation-bg-secondary border border-aviation-border-panel rounded text-xs font-mono"
                  />
                  <button
                    onClick={() => { if (newSecretKey) { onSecretCreate?.(selectedVault.id, newSecretKey, newSecretValue); setNewSecretKey(''); setNewSecretValue(''); } }}
                    className="p-1.5 hover:bg-aviation-bg-instrument rounded"
                  >
                    <Plus className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <Shield className="w-8 h-8 text-aviation-text-muted mx-auto mb-3" />
              <p className="text-sm text-aviation-text-muted">Select a vault</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Infrastructure Diff Viewer
// ============================================================================

interface InfraDiffFile {
  id: string;
  path: string;
  type: 'yaml' | 'json' | 'terraform' | 'helm';
  status?: 'added' | 'deleted' | 'modified';
}

interface InfrastructureDiffViewerProps {
  files: InfraDiffFile[];
  selectedFileId?: string | null;
  onFileSelect?: (file: InfraDiffFile) => void;
  className?: string;
}

export const InfrastructureDiffViewer: React.FC<InfrastructureDiffViewerProps> = ({
  files,
  selectedFileId,
  onFileSelect,
  className,
}) => {
  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'added': return 'text-green-400 bg-green-500/20';
      case 'deleted': return 'text-red-400 bg-red-500/20';
      case 'modified': return 'text-amber-400 bg-amber-500/20';
      default: return 'text-gray-400 bg-gray-500/20';
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <FileCode className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Infrastructure Diff</span>
      </div>
      <div className="flex-1 overflow-auto">
        {files.map((file) => {
          const isSelected = file.id === selectedFileId;
          return (
            <div
              key={file.id}
              className={cn(
                'flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-aviation-border-panel',
                isSelected ? 'bg-aviation-cyan/10' : 'hover:bg-aviation-bg-secondary'
              )}
              onClick={() => onFileSelect?.(file)}
            >
              <FileCode className="w-4 h-4 text-aviation-text-muted" />
              <div className="flex-1 min-w-0">
                <code className="text-xs font-mono truncate">{file.path}</code>
              </div>
              <span className="text-[10px] text-aviation-text-muted uppercase">{file.type}</span>
              {file.status && (
                <span className={cn('px-1.5 py-0.5 rounded text-[10px]', getStatusColor(file.status))}>
                  {file.status}
                </span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Resource Scaler
// ============================================================================

interface ScalableResource {
  id: string;
  name: string;
  type: 'function' | 'container' | 'service' | 'database';
  currentReplicas: number;
  minReplicas: number;
  maxReplicas: number;
  targetCpuUtilization?: number;
  status: 'scaled' | 'scaling' | 'error' | 'paused';
}

interface ResourceScalerProps {
  resources: ScalableResource[];
  selectedResourceId?: string | null;
  onResourceSelect?: (resource: ScalableResource) => void;
  onReplicasChange?: (resourceId: string, replicas: number) => void;
  onAutoScaleToggle?: (resourceId: string, enabled: boolean) => void;
  className?: string;
}

export const ResourceScaler: React.FC<ResourceScalerProps> = ({
  resources,
  selectedResourceId,
  onResourceSelect,
  onReplicasChange,
  onAutoScaleToggle,
  className,
}) => {
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Scaling className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Resource Scaler</span>
      </div>
      <div className="flex-1 overflow-auto p-4 space-y-3">
        {resources.map((resource) => {
          const isSelected = resource.id === selectedResourceId;
          return (
            <div
              key={resource.id}
              className={cn(
                'p-4 rounded-lg border cursor-pointer transition-colors',
                isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
              )}
              onClick={() => onResourceSelect?.(resource)}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <Box className="w-4 h-4 text-aviation-text-muted" />
                  <span className="text-sm font-medium">{resource.name}</span>
                </div>
                <span className="text-xs text-aviation-text-muted capitalize">{resource.type}</span>
              </div>
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <button
                    onClick={(e) => { e.stopPropagation(); onReplicasChange?.(resource.id, Math.max(resource.minReplicas, resource.currentReplicas - 1)); }}
                    className="p-1 hover:bg-aviation-bg-instrument rounded"
                  >
                    <Minus className="w-4 h-4" />
                  </button>
                  <span className="text-lg font-mono w-8 text-center">{resource.currentReplicas}</span>
                  <button
                    onClick={(e) => { e.stopPropagation(); onReplicasChange?.(resource.id, Math.min(resource.maxReplicas, resource.currentReplicas + 1)); }}
                    className="p-1 hover:bg-aviation-bg-instrument rounded"
                  >
                    <Plus className="w-4 h-4" />
                  </button>
                </div>
                <div className="flex-1">
                  <div className="flex items-center justify-between text-[10px] text-aviation-text-muted mb-1">
                    <span>Range</span>
                    <span>{resource.minReplicas} - {resource.maxReplicas}</span>
                  </div>
                  <div className="h-1 bg-aviation-bg-secondary rounded-full overflow-hidden">
                    <div
                      className="h-full bg-aviation-cyan transition-all"
                      style={{ width: `${((resource.currentReplicas - resource.minReplicas) / (resource.maxReplicas - resource.minReplicas)) * 100}%` }}
                    />
                  </div>
                </div>
              </div>
              <div className="flex items-center justify-between mt-3">
                <span className="text-[10px] text-aviation-text-muted">
                  Target: {resource.targetCpuUtilization || 70}% CPU
                </span>
                <button
                  onClick={(e) => { e.stopPropagation(); onAutoScaleToggle?.(resource.id, resource.status !== 'paused'); }}
                  className={cn(
                    'px-2 py-1 rounded text-[10px] transition-colors',
                    resource.status === 'paused' ? 'text-aviation-text-muted bg-aviation-bg-secondary' : 'text-aviation-cyan bg-aviation-cyan/20'
                  )}
                >
                  {resource.status === 'paused' ? 'Auto-scale off' : 'Auto-scale on'}
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Traffic Balancer View
// ============================================================================

interface TrafficTarget {
  id: string;
  name: string;
  type: 'function' | 'container' | 'external';
  weight: number;
  status: 'healthy' | 'unhealthy' | 'unknown';
  latency?: number;
  errorRate?: number;
}

interface TrafficBalancerViewProps {
  balancerName: string;
  targets: TrafficTarget[];
  selectedTargetId?: string | null;
  onTargetSelect?: (target: TrafficTarget) => void;
  onTargetWeightChange?: (targetId: string, weight: number) => void;
  className?: string;
}

export const TrafficBalancerView: React.FC<TrafficBalancerViewProps> = ({
  balancerName,
  targets,
  selectedTargetId,
  onTargetSelect,
  onTargetWeightChange,
  className,
}) => {
  const totalWeight = targets.reduce((sum, t) => sum + t.weight, 0);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Route className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">{balancerName}</span>
        <span className="text-xs text-aviation-text-muted ml-auto">{targets.length} targets</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="space-y-3">
          {targets.map((target) => {
            const isSelected = target.id === selectedTargetId;
            const weightPercent = totalWeight > 0 ? (target.weight / totalWeight) * 100 : 0;
            return (
              <div
                key={target.id}
                className={cn(
                  'p-4 rounded-lg border cursor-pointer transition-colors',
                  isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                )}
                onClick={() => onTargetSelect?.(target)}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{target.name}</span>
                    <span className={cn(
                      'px-1.5 py-0.5 rounded text-[10px]',
                      target.status === 'healthy' ? 'text-green-400 bg-green-500/20' :
                      target.status === 'unhealthy' ? 'text-red-400 bg-red-500/20' : 'text-gray-400 bg-gray-500/20'
                    )}>
                      {target.status}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 text-xs text-aviation-text-muted">
                    {target.latency && <span>{target.latency}ms</span>}
                    {target.errorRate !== undefined && <span>{target.errorRate}% errors</span>}
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <input
                    type="range"
                    min="0"
                    max="100"
                    value={target.weight}
                    onChange={(e) => onTargetWeightChange?.(target.id, parseInt(e.target.value))}
                    onClick={(e) => e.stopPropagation()}
                    className="flex-1"
                  />
                  <span className="text-xs font-mono w-12 text-right">{Math.round(weightPercent)}%</span>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Rollback Manager
// ============================================================================

interface RollbackVersion {
  id: string;
  version: string;
  deployedAt: number;
  deployedBy: string;
  status: 'current' | 'previous' | 'archived';
  healthScore?: number;
}

interface RollbackManagerProps {
  resourceId: string;
  resourceName: string;
  versions: RollbackVersion[];
  onRollbackSelect?: (version: RollbackVersion) => void;
  onRollbackConfirm?: (versionId: string) => void;
  className?: string;
}

export const RollbackManager: React.FC<RollbackManagerProps> = ({
  resourceName,
  versions,
  onRollbackSelect,
  onRollbackConfirm,
  className,
}) => {
  const formatDate = (ts: number) => new Date(ts).toLocaleDateString() + ' ' + new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <RotateCw className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Rollback: {resourceName}</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="space-y-2">
          {versions.map((version) => (
            <div
              key={version.id}
              className={cn(
                'flex items-center gap-3 p-3 rounded-lg border transition-colors cursor-pointer',
                version.status === 'current' ? 'border-green-500/50 bg-green-500/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
              )}
              onClick={() => onRollbackSelect?.(version)}
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-mono">{version.version}</span>
                  {version.status === 'current' && (
                    <span className="px-1.5 py-0.5 bg-green-500/20 text-green-400 rounded text-[10px]">current</span>
                  )}
                </div>
                <div className="flex items-center gap-2 text-xs text-aviation-text-muted mt-1">
                  <span>{formatDate(version.deployedAt)}</span>
                  <span>·</span>
                  <span>{version.deployedBy}</span>
                </div>
              </div>
              {version.healthScore !== undefined && (
                <div className={cn(
                  'text-sm font-medium',
                  version.healthScore >= 90 ? 'text-green-400' :
                  version.healthScore >= 70 ? 'text-amber-400' : 'text-red-400'
                )}>
                  {version.healthScore}%
                </div>
              )}
              {version.status !== 'current' && (
                <button
                  onClick={(e) => { e.stopPropagation(); onRollbackConfirm?.(version.id); }}
                  className="px-3 py-1.5 bg-aviation-cyan text-black rounded text-xs font-medium hover:bg-aviation-cyan/80"
                >
                  Rollback
                </button>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Deployment Simulation
// ============================================================================

interface SimulationStep {
  id: string;
  type: 'validate' | 'deploy' | 'test' | 'rollback' | 'scale';
  status: 'pending' | 'running' | 'completed' | 'failed';
  message: string;
}

interface DeploymentSimulationProps {
  simulation: {
    id: string;
    name: string;
    steps: SimulationStep[];
    impactAnalysis?: { downtime?: number; affectedRequests?: number; estimatedTime?: number };
  } | null;
  onAccept?: () => void;
  onReject?: () => void;
  onStepForward?: () => void;
  onStepBackward?: () => void;
  className?: string;
}

export const DeploymentSimulation: React.FC<DeploymentSimulationProps> = ({
  simulation,
  onAccept,
  onReject,
  className,
}) => {
  const getStepIcon = (type: string) => {
    switch (type) {
      case 'validate': return <CheckCircle2 className="w-4 h-4" />;
      case 'deploy': return <Rocket className="w-4 h-4" />;
      case 'test': return <FlaskConical className="w-4 h-4" />;
      case 'rollback': return <RotateCw className="w-4 h-4" />;
      case 'scale': return <Scaling className="w-4 h-4" />;
      default: return <Circle className="w-4 h-4" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-green-400';
      case 'failed': return 'text-red-400';
      case 'running': return 'text-cyan-400';
      default: return 'text-aviation-text-muted';
    }
  };

  if (!simulation) {
    return (
      <div className={cn('flex flex-col h-full items-center justify-center bg-aviation-bg-panel rounded-lg border border-aviation-border-panel', className)}>
        <FlaskConical className="w-8 h-8 text-aviation-text-muted mb-3" />
        <p className="text-sm text-aviation-text-muted">No simulation loaded</p>
      </div>
    );
  }

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <FlaskConical className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">{simulation.name}</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="space-y-2">
          {simulation.steps.map((step, idx) => (
            <div key={step.id} className="flex items-center gap-3 px-3 py-2 rounded bg-aviation-bg-secondary">
              <div className={getStatusColor(step.status)}>{getStepIcon(step.type)}</div>
              <span className="flex-1 text-xs">{step.message}</span>
              <span className={cn('text-[10px]', getStatusColor(step.status))}>{step.status}</span>
            </div>
          ))}
        </div>
        {simulation.impactAnalysis && (
          <div className="mt-4 p-3 bg-aviation-bg-secondary rounded">
            <div className="text-xs text-aviation-text-muted mb-2">Impact Analysis</div>
            <div className="grid grid-cols-3 gap-2 text-xs">
              {simulation.impactAnalysis.downtime !== undefined && (
                <div className="text-center">
                  <div className="text-lg font-mono text-amber-400">{simulation.impactAnalysis.downtime}s</div>
                  <div className="text-aviation-text-dim">Downtime</div>
                </div>
              )}
              {simulation.impactAnalysis.affectedRequests !== undefined && (
                <div className="text-center">
                  <div className="text-lg font-mono text-cyan-400">{simulation.impactAnalysis.affectedRequests}</div>
                  <div className="text-aviation-text-dim">Affected</div>
                </div>
              )}
              {simulation.impactAnalysis.estimatedTime !== undefined && (
                <div className="text-center">
                  <div className="text-lg font-mono text-green-400">{simulation.impactAnalysis.estimatedTime}m</div>
                  <div className="text-aviation-text-dim">Duration</div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
      <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <button onClick={onReject} className="px-4 py-2 text-sm text-aviation-text-muted hover:text-aviation-text-primary">
          Cancel
        </button>
        <button onClick={onAccept} className="px-4 py-2 bg-aviation-cyan text-black rounded text-sm font-medium">
          Deploy
        </button>
      </div>
    </div>
  );
};

// ============================================================================
// Build Artifact Explorer
// ============================================================================

interface BuildArtifact {
  id: string;
  name: string;
  version: string;
  type: 'binary' | 'image' | 'archive' | 'layer';
  size: number;
  buildNumber: number;
  status: 'available' | 'building' | 'failed';
}

interface BuildArtifactExplorerProps {
  artifacts: BuildArtifact[];
  selectedArtifactId?: string | null;
  onArtifactSelect?: (artifact: BuildArtifact) => void;
  onArtifactDownload?: (artifactId: string) => void;
  className?: string;
}

export const BuildArtifactExplorer: React.FC<BuildArtifactExplorerProps> = ({
  artifacts,
  selectedArtifactId,
  onArtifactSelect,
  onArtifactDownload,
  className,
}) => {
  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes}B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'image': return <Container className="w-4 h-4 text-cyan-400" />;
      case 'binary': return <Box className="w-4 h-4 text-green-400" />;
      case 'archive': return <Package className="w-4 h-4 text-amber-400" />;
      default: return <FileCode className="w-4 h-4" />;
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Package className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Build Artifacts</span>
        <span className="text-xs text-aviation-text-muted ml-auto">{artifacts.length}</span>
      </div>
      <div className="flex-1 overflow-auto">
        {artifacts.map((artifact) => {
          const isSelected = artifact.id === selectedArtifactId;
          return (
            <div
              key={artifact.id}
              className={cn(
                'flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-aviation-border-panel',
                isSelected ? 'bg-aviation-cyan/10' : 'hover:bg-aviation-bg-secondary'
              )}
              onClick={() => onArtifactSelect?.(artifact)}
            >
              {getTypeIcon(artifact.type)}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium truncate">{artifact.name}</span>
                  <span className="text-xs font-mono text-aviation-text-muted">v{artifact.version}</span>
                </div>
                <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
                  <span>#{artifact.buildNumber}</span>
                  <span>·</span>
                  <span>{formatSize(artifact.size)}</span>
                </div>
              </div>
              <span className={cn(
                'px-1.5 py-0.5 rounded text-[10px]',
                artifact.status === 'available' ? 'text-green-400 bg-green-500/20' :
                artifact.status === 'building' ? 'text-amber-400 bg-amber-500/20' : 'text-red-400 bg-red-500/20'
              )}>
                {artifact.status}
              </span>
              {artifact.status === 'available' && (
                <button
                  onClick={(e) => { e.stopPropagation(); onArtifactDownload?.(artifact.id); }}
                  className="p-1.5 hover:bg-aviation-bg-instrument rounded"
                >
                  <Download className="w-4 h-4" />
                </button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Cluster Health Monitor
// ============================================================================

interface ClusterHealth {
  id: string;
  name: string;
  provider: string;
  status: 'healthy' | 'degraded' | 'unhealthy';
  score: number;
  components: Array<{ name: string; status: string; message?: string }>;
  metrics?: { cpuUsage?: number; memoryUsage?: number };
}

interface ClusterHealthMonitorProps {
  clusters: ClusterHealth[];
  selectedClusterId?: string | null;
  onClusterSelect?: (cluster: ClusterHealth) => void;
  className?: string;
}

export const ClusterHealthMonitor: React.FC<ClusterHealthMonitorProps> = ({
  clusters,
  selectedClusterId,
  onClusterSelect,
  className,
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy': return 'text-green-400 bg-green-500/20';
      case 'degraded': return 'text-amber-400 bg-amber-500/20';
      case 'unhealthy': return 'text-red-400 bg-red-500/20';
      default: return 'text-gray-400 bg-gray-500/20';
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Activity className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Cluster Health</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3">
          {clusters.map((cluster) => {
            const isSelected = cluster.id === selectedClusterId;
            return (
              <div
                key={cluster.id}
                className={cn(
                  'p-4 rounded-lg border cursor-pointer transition-colors',
                  isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                )}
                onClick={() => onClusterSelect?.(cluster)}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <ServerIcon className="w-4 h-4 text-aviation-text-muted" />
                    <span className="text-sm font-medium">{cluster.name}</span>
                  </div>
                  <span className={cn('px-1.5 py-0.5 rounded text-[10px]', getStatusColor(cluster.status))}>
                    {cluster.status}
                  </span>
                </div>
                <div className="flex items-center justify-center mb-3">
                  <div className="relative w-20 h-20">
                    <svg className="w-full h-full transform -rotate-90">
                      <circle cx="40" cy="40" r="35" stroke="currentColor" strokeWidth="6" fill="none" className="text-aviation-bg-secondary" />
                      <circle
                        cx="40" cy="40" r="35" stroke="currentColor" strokeWidth="6" fill="none"
                        className={cluster.score >= 90 ? 'text-green-400' : cluster.score >= 70 ? 'text-amber-400' : 'text-red-400'}
                        strokeDasharray={`${(cluster.score / 100) * 220} 220`}
                      />
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-lg font-mono">{cluster.score}</span>
                    </div>
                  </div>
                </div>
                {cluster.metrics && (
                  <div className="grid grid-cols-2 gap-2 text-xs">
                    {cluster.metrics.cpuUsage !== undefined && (
                      <div className="flex items-center justify-between">
                        <span className="text-aviation-text-muted">CPU</span>
                        <span className={cluster.metrics.cpuUsage > 80 ? 'text-red-400' : 'text-aviation-text-primary'}>
                          {cluster.metrics.cpuUsage}%
                        </span>
                      </div>
                    )}
                    {cluster.metrics.memoryUsage !== undefined && (
                      <div className="flex items-center justify-between">
                        <span className="text-aviation-text-muted">Memory</span>
                        <span className={cluster.metrics.memoryUsage > 80 ? 'text-red-400' : 'text-aviation-text-primary'}>
                          {cluster.metrics.memoryUsage}%
                        </span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Cold Start Analyzer
// ============================================================================

interface ColdStartMetric {
  functionId: string;
  functionName: string;
  region: string;
  averageDuration: number;
  p99Duration: number;
  invocations: number;
  runtime: string;
}

interface ColdStartAnalyzerProps {
  metrics: ColdStartMetric[];
  selectedFunctionId?: string | null;
  onFunctionSelect?: (metric: ColdStartMetric) => void;
  className?: string;
}

export const ColdStartAnalyzer: React.FC<ColdStartAnalyzerProps> = ({
  metrics,
  selectedFunctionId,
  onFunctionSelect,
  className,
}) => {
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Zap className="w-5 h-5 text-aviation-amber" />
        <span className="text-sm font-medium">Cold Start Analysis</span>
      </div>
      <div className="flex-1 overflow-auto">
        {metrics.map((metric) => {
          const isSelected = metric.functionId === selectedFunctionId;
          return (
            <div
              key={metric.functionId}
              className={cn(
                'flex items-center gap-4 px-4 py-3 cursor-pointer border-b border-aviation-border-panel',
                isSelected ? 'bg-aviation-amber/10' : 'hover:bg-aviation-bg-secondary'
              )}
              onClick={() => onFunctionSelect?.(metric)}
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium truncate">{metric.functionName}</span>
                  <span className="text-[10px] text-aviation-text-muted">{metric.region} · {metric.runtime}</span>
                </div>
                <div className="flex items-center gap-4 text-xs text-aviation-text-muted mt-1">
                  <span>{metric.invocations.toLocaleString()} invocations</span>
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm font-mono text-aviation-cyan">{metric.averageDuration}ms</div>
                <div className="text-[10px] text-aviation-text-muted">avg</div>
              </div>
              <div className="text-right">
                <div className="text-sm font-mono text-red-400">{metric.p99Duration}ms</div>
                <div className="text-[10px] text-aviation-text-muted">p99</div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Serverless Execution Map
// ============================================================================

interface ExecutionFlow {
  id: string;
  functionName: string;
  region: string;
  invocations: number;
  avgDuration: number;
  successRate: number;
  coldStarts: number;
}

interface ServerlessExecutionMapProps {
  executions: ExecutionFlow[];
  selectedExecutionId?: string | null;
  onExecutionSelect?: (execution: ExecutionFlow) => void;
  className?: string;
}

export const ServerlessExecutionMap: React.FC<ServerlessExecutionMapProps> = ({
  executions,
  selectedExecutionId,
  onExecutionSelect,
  className,
}) => {
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-aviation-border-panel">
        <Network className="w-5 h-5 text-aviation-cyan" />
        <span className="text-sm font-medium">Serverless Execution Map</span>
      </div>
      <div className="flex-1 overflow-auto p-4">
        <div className="relative">
          {/* Simple flow visualization */}
          <div className="space-y-4">
            {executions.map((exec) => {
              const isSelected = exec.id === selectedExecutionId;
              return (
                <div
                  key={exec.id}
                  className={cn(
                    'p-4 rounded-lg border cursor-pointer transition-colors',
                    isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                  )}
                  onClick={() => onExecutionSelect?.(exec)}
                >
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">{exec.functionName}</span>
                      <span className="text-[10px] text-aviation-text-muted">{exec.region}</span>
                    </div>
                    <span className={cn(
                      'px-1.5 py-0.5 rounded text-[10px]',
                      exec.successRate >= 99 ? 'text-green-400 bg-green-500/20' :
                      exec.successRate >= 95 ? 'text-amber-400 bg-amber-500/20' : 'text-red-400 bg-red-500/20'
                    )}>
                      {exec.successRate}% success
                    </span>
                  </div>
                  <div className="grid grid-cols-4 gap-3 text-xs">
                    <div>
                      <div className="text-aviation-text-muted">Invocations</div>
                      <div className="text-aviation-text-primary font-mono">{exec.invocations.toLocaleString()}</div>
                    </div>
                    <div>
                      <div className="text-aviation-text-muted">Avg Duration</div>
                      <div className="text-aviation-text-primary font-mono">{exec.avgDuration}ms</div>
                    </div>
                    <div>
                      <div className="text-aviation-text-muted">Cold Starts</div>
                      <div className="text-aviation-amber font-mono">{exec.coldStarts}</div>
                    </div>
                    <div>
                      <div className="text-aviation-text-muted">Success Rate</div>
                      <div className="text-green-400 font-mono">{exec.successRate}%</div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
};
