/**
 * @functionfly/ui-devops
 * DevOps & Infrastructure Components - Types and Interfaces
 */

// ============================================================================
// Deployment Pipeline
// ============================================================================

export interface PipelineStage {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped' | 'waiting';
  duration?: number;
  startedAt?: number;
  completedAt?: number;
  artifacts?: Array<{
    name: string;
    type: string;
    url: string;
    size?: number;
  }>;
  tasks: PipelineTask[];
}

export interface PipelineTask {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  duration?: number;
  logs?: string[];
  error?: string;
}

export interface Pipeline {
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

export interface DeploymentPipelineProps {
  pipeline: Pipeline;
  selectedStageId?: string | null;
  onStageSelect?: (stage: PipelineStage) => void;
  onStageRetry?: (stageId: string) => void;
  onPipelinePause?: () => void;
  onPipelineResume?: () => void;
  showLogs?: boolean;
  className?: string;
}

// ============================================================================
// Environment Manager
// ============================================================================

export interface Environment {
  id: string;
  name: string;
  type: 'development' | 'staging' | 'production' | 'preview';
  color: string;
  variables: Record<string, string>;
  secrets: Array<{
    key: string;
    masked: boolean;
    lastUpdated: number;
  }>;
  replicas: number;
  autoScale: boolean;
  region?: string;
  createdAt: number;
  updatedAt: number;
}

export interface EnvironmentManagerProps {
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

// ============================================================================
// Cloud Region Selector
// ============================================================================

export interface CloudRegion {
  id: string;
  name: string;
  provider: 'aws' | 'gcp' | 'azure' | 'custom';
  zone: string;
  zoneName: string;
  location: string;
  country: string;
  coordinates: {
    lat: number;
    lng: number;
  };
  isAvailable: boolean;
  isRecommended?: boolean;
  specs?: {
    compute?: number;
    memory?: number;
    storage?: number;
    gpu?: boolean;
  };
}

export interface CloudRegionSelectorProps {
  regions: CloudRegion[];
  selectedRegionId?: string | null;
  selectedProvider?: 'aws' | 'gcp' | 'azure' | 'all';
  onRegionSelect?: (region: CloudRegion) => void;
  onProviderFilter?: (provider: 'aws' | 'gcp' | 'azure' | 'all') => void;
  showRegionStats?: boolean;
  className?: string;
}

// ============================================================================
// Runtime Target Selector
// ============================================================================

export interface RuntimeTarget {
  id: string;
  name: string;
  type: 'nodejs' | 'python' | 'go' | 'rust' | 'kotlin' | 'ruby' | 'deno' | 'bun' | 'sar';
  version: string;
  status: 'stable' | 'beta' | 'deprecated';
  memoryLimit?: number;
  timeout?: number;
  layers?: string[];
}

export interface RuntimeTargetSelectorProps {
  targets: RuntimeTarget[];
  selectedTargetId?: string | null;
  onTargetSelect?: (target: RuntimeTarget) => void;
  showVersionHistory?: boolean;
  className?: string;
}

// ============================================================================
// Kubernetes Topology View
// ============================================================================

export interface K8sNode {
  id: string;
  name: string;
  type: 'control-plane' | 'worker' | 'ingress' | 'storage';
  status: 'ready' | 'not-ready' | 'unknown';
  resources?: {
    cpu?: number;
    memory?: number;
    pods?: number;
  };
  labels?: Record<string, string>;
  taints?: string[];
  conditions?: Array<{
    type: string;
    status: boolean;
    lastTransition: number;
  }>;
}

export interface K8sService {
  id: string;
  name: string;
  type: 'cluster-ip' | 'node-port' | 'load-balancer' | 'external-name';
  selector: Record<string, string>;
  ports?: Array<{
    name: string;
    port: number;
    targetPort: number;
    protocol: 'TCP' | 'UDP';
  }>;
  clusterIp: string;
}

export interface K8sNamespace {
  id: string;
  name: string;
  status: 'active' | 'terminating';
  labels?: Record<string, string>;
  pods: string[];
  services: string[];
}

export interface KubernetesTopologyViewProps {
  nodes: K8sNode[];
  services: K8sService[];
  namespaces: K8sNamespace[];
  selectedNodeId?: string | null;
  onNodeSelect?: (node: K8sNode) => void;
  onServiceSelect?: (service: K8sService) => void;
  layout?: 'hierarchical' | 'force' | 'grid';
  showResourceUtilization?: boolean;
  className?: string;
}

// ============================================================================
// Edge Deployment Map
// ============================================================================

export interface EdgeLocation {
  id: string;
  name: string;
  provider: string;
  region: string;
  city: string;
  country: string;
  coordinates: {
    lat: number;
    lng: number;
  };
  status: 'online' | 'offline' | 'degraded';
  latency?: number;
  capacity?: {
    current: number;
    max: number;
  };
  deployments?: number;
}

export interface EdgeDeploymentMapProps {
  locations: EdgeLocation[];
  selectedLocationId?: string | null;
  onLocationSelect?: (location: EdgeLocation) => void;
  onLocationHover?: (location: EdgeLocation | null) => void;
  showLatency?: boolean;
  showCapacity?: boolean;
  className?: string;
}

// ============================================================================
// Container Lifecycle Panel
// ============================================================================

export interface Container {
  id: string;
  name: string;
  image: string;
  tag: string;
  status: 'running' | 'paused' | 'stopped' | 'restarting' | 'exited';
  createdAt: number;
  startedAt?: number;
  ports?: Array<{
    host: number;
    container: number;
    protocol: 'TCP' | 'UDP';
  }>;
  envVars?: Record<string, string>;
  command?: string[];
  limits?: {
    cpu?: string;
    memory?: string;
  };
  usage?: {
    cpu?: number;
    memory?: number;
    network?: number;
  };
}

export interface ContainerLifecyclePanelProps {
  containers: Container[];
  selectedContainerId?: string | null;
  onContainerSelect?: (container: Container) => void;
  onContainerStart?: (containerId: string) => void;
  onContainerStop?: (containerId: string) => void;
  onContainerRestart?: (containerId: string) => void;
  onContainerDelete?: (containerId: string) => void;
  onLogsView?: (containerId: string) => void;
  className?: string;
}

// ============================================================================
// Secret Vault Manager
// ============================================================================

export interface SecretVault {
  id: string;
  name: string;
  type: 'aws-secrets-manager' | 'gcp-secret-manager' | 'azure-keyvault' | 'hashicorp-vault' | 'local';
  status: 'connected' | 'disconnected' | 'error';
  secrets: Array<{
    key: string;
    masked: boolean;
    lastRotated?: number;
    version?: number;
  }>;
  lastSync?: number;
}

export interface SecretVaultManagerProps {
  vaults: SecretVault[];
  selectedVaultId?: string | null;
  onVaultSelect?: (vault: SecretVault) => void;
  onVaultConnect?: (vaultId: string) => void;
  onVaultDisconnect?: (vaultId: string) => void;
  onSecretCreate?: (vaultId: string, key: string, value: string) => void;
  onSecretUpdate?: (vaultId: string, key: string, value: string) => void;
  onSecretDelete?: (vaultId: string, key: string) => void;
  onSecretRotate?: (vaultId: string, key: string) => void;
  className?: string;
}

// ============================================================================
// Infrastructure Diff Viewer
// ============================================================================

export interface InfraDiffFile {
  id: string;
  path: string;
  oldContent: string;
  newContent: string;
  type: 'yaml' | 'json' | 'terraform' | 'helm';
  status?: 'added' | 'deleted' | 'modified';
}

export interface InfraDiffHunk {
  id: string;
  lines: Array<{
    id: string;
    type: 'add' | 'delete' | 'context';
    content: string;
    oldLineNumber?: number;
    newLineNumber?: number;
  }>;
}

export interface InfrastructureDiffViewerProps {
  files: InfraDiffFile[];
  selectedFileId?: string | null;
  hunks?: InfraDiffHunk[];
  onFileSelect?: (file: InfraDiffFile) => void;
  onAcceptChange?: (fileId: string, changeId: string) => void;
  onRejectChange?: (fileId: string, changeId: string) => void;
  className?: string;
}

// ============================================================================
// Resource Scaler
// ============================================================================

export interface ScalableResource {
  id: string;
  name: string;
  type: 'function' | 'container' | 'service' | 'database';
  currentReplicas: number;
  minReplicas: number;
  maxReplicas: number;
  targetCpuUtilization?: number;
  targetMemoryUtilization?: number;
  scalingHistory?: Array<{
    timestamp: number;
    replicas: number;
    reason: string;
  }>;
  status: 'scaled' | 'scaling' | 'error' | 'paused';
}

export interface ResourceScalerProps {
  resources: ScalableResource[];
  selectedResourceId?: string | null;
  onResourceSelect?: (resource: ScalableResource) => void;
  onReplicasChange?: (resourceId: string, replicas: number) => void;
  onAutoScaleToggle?: (resourceId: string, enabled: boolean) => void;
  onScalingPolicyUpdate?: (resourceId: string, policy: Partial<ScalableResource>) => void;
  className?: string;
}

// ============================================================================
// Traffic Balancer View
// ============================================================================

export interface TrafficTarget {
  id: string;
  name: string;
  type: 'function' | 'container' | 'external';
  url: string;
  weight: number;
  status: 'healthy' | 'unhealthy' | 'unknown';
  latency?: number;
  errorRate?: number;
  requestsPerSecond?: number;
}

export interface TrafficBalancerViewProps {
  balancerName: string;
  targets: TrafficTarget[];
  selectedTargetId?: string | null;
  onTargetSelect?: (target: TrafficTarget) => void;
  onTargetWeightChange?: (targetId: string, weight: number) => void;
  onTargetAdd?: (target: Partial<TrafficTarget>) => void;
  onTargetRemove?: (targetId: string) => void;
  className?: string;
}

// ============================================================================
// Rollback Manager
// ============================================================================

export interface RollbackVersion {
  id: string;
  version: string;
  deployedAt: number;
  deployedBy: string;
  status: 'current' | 'previous' | 'archived';
  changes?: string;
  artifacts?: string[];
  healthScore?: number;
}

export interface RollbackManagerProps {
  resourceId: string;
  resourceName: string;
  versions: RollbackVersion[];
  onRollbackSelect?: (version: RollbackVersion) => void;
  onRollbackConfirm?: (versionId: string) => void;
  className?: string;
}

// ============================================================================
// Deployment Simulation
// ============================================================================

export interface SimulationStep {
  id: string;
  type: 'validate' | 'deploy' | 'test' | 'rollback' | 'scale';
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  message: string;
  details?: string;
  duration?: number;
}

export interface DeploymentSimulationProps {
  simulation: {
    id: string;
    name: string;
    steps: SimulationStep[];
    impactAnalysis?: {
      downtime?: number;
      affectedRequests?: number;
      estimatedTime?: number;
    };
  } | null;
  onAccept?: () => void;
  onReject?: () => void;
  onStepForward?: () => void;
  onStepBackward?: () => void;
  className?: string;
}

// ============================================================================
// Build Artifact Explorer
// ============================================================================

export interface BuildArtifact {
  id: string;
  name: string;
  version: string;
  type: 'binary' | 'image' | 'archive' | 'layer';
  size: number;
  checksum: string;
  createdAt: number;
  buildNumber: number;
  commitSha: string;
  status: 'available' | 'building' | 'failed' | 'expired';
  downloads?: number;
  metadata?: Record<string, string>;
}

export interface BuildArtifactExplorerProps {
  artifacts: BuildArtifact[];
  selectedArtifactId?: string | null;
  onArtifactSelect?: (artifact: BuildArtifact) => void;
  onArtifactDownload?: (artifactId: string) => void;
  onArtifactDelete?: (artifactId: string) => void;
  onArtifactPromote?: (artifactId: string, target: string) => void;
  className?: string;
}

// ============================================================================
// Cluster Health Monitor
// ============================================================================

export interface ClusterHealth {
  id: string;
  name: string;
  provider: 'aws' | 'gcp' | 'azure' | 'custom';
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  score: number;
  components: Array<{
    name: string;
    status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
    message?: string;
    lastChecked: number;
  }>;
  metrics?: {
    cpuUsage?: number;
    memoryUsage?: number;
    diskUsage?: number;
    networkIn?: number;
    networkOut?: number;
  };
  events?: Array<{
    id: string;
    type: 'info' | 'warning' | 'error';
    message: string;
    timestamp: number;
  }>;
}

export interface ClusterHealthMonitorProps {
  clusters: ClusterHealth[];
  selectedClusterId?: string | null;
  onClusterSelect?: (cluster: ClusterHealth) => void;
  onRefresh?: () => void;
  showMetrics?: boolean;
  className?: string;
}

// ============================================================================
// Cold Start Analyzer
// ============================================================================

export interface ColdStartMetric {
  functionId: string;
  functionName: string;
  region: string;
  averageDuration: number;
  p50Duration: number;
  p99Duration: number;
  invocations: number;
  memorySize: number;
  runtime: string;
}

export interface ColdStartAnalyzerProps {
  metrics: ColdStartMetric[];
  selectedFunctionId?: string | null;
  onFunctionSelect?: (metric: ColdStartMetric) => void;
  onOptimize?: (functionId: string) => void;
  sortBy?: 'averageDuration' | 'p99Duration' | 'invocations';
  className?: string;
}

// ============================================================================
// Serverless Execution Map
// ============================================================================

export interface ExecutionFlow {
  id: string;
  functionName: string;
  region: string;
  invocations: number;
  avgDuration: number;
  successRate: number;
  coldStarts: number;
  dependencies: string[];
}

export interface ServerlessExecutionMapProps {
  executions: ExecutionFlow[];
  selectedExecutionId?: string | null;
  onExecutionSelect?: (execution: ExecutionFlow) => void;
  onRegionFilter?: (region: string) => void;
  className?: string;
}
