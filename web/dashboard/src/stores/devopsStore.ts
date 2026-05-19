/**
 * DevOps Store
 * Global state management for DevOps & Infrastructure components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types
// ============================================================================

export interface PipelineStage {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped' | 'waiting'
  duration?: number
  startedAt?: number
  completedAt?: number
  artifacts?: Array<{ name: string; type: string; url: string; size?: number }>
  tasks: Array<{ id: string; name: string; status: string; duration?: number; logs?: string[]; error?: string }>
}

export interface Pipeline {
  id: string
  name: string
  version: string
  status: 'active' | 'paused' | 'archived'
  stages: PipelineStage[]
  currentStageId?: string
  triggeredBy: string
  triggeredAt: number
  branch: string
  commitSha: string
  source: 'manual' | 'webhook' | 'scheduled' | 'api'
}

export interface Environment {
  id: string
  name: string
  type: 'development' | 'staging' | 'production' | 'preview'
  color: string
  variables: Record<string, string>
  secrets: Array<{ key: string; masked: boolean; lastUpdated: number }>
  replicas: number
  autoScale: boolean
  region?: string
}

export interface CloudRegion {
  id: string
  name: string
  provider: 'aws' | 'gcp' | 'azure' | 'custom'
  zone: string
  zoneName: string
  location: string
  country: string
  coordinates: { lat: number; lng: number }
  isAvailable: boolean
  isRecommended?: boolean
  specs?: { compute?: number; memory?: number; storage?: number; gpu?: boolean }
}

export interface RuntimeTarget {
  id: string
  name: string
  type: 'nodejs' | 'python' | 'go' | 'rust' | 'kotlin' | 'ruby' | 'deno' | 'bun' | 'sar'
  version: string
  status: 'stable' | 'beta' | 'deprecated'
  memoryLimit?: number
  timeout?: number
}

export interface K8sNode {
  id: string
  name: string
  type: 'control-plane' | 'worker' | 'ingress' | 'storage'
  status: 'ready' | 'not-ready' | 'unknown'
  resources?: { cpu?: number; memory?: number; pods?: number }
}

export interface K8sService {
  id: string
  name: string
  type: 'cluster-ip' | 'node-port' | 'load-balancer' | 'external-name'
  selector: Record<string, string>
  ports?: Array<{ name: string; port: number; targetPort: number }>
}

export interface K8sNamespace {
  id: string
  name: string
  status: 'active' | 'terminating'
  pods: string[]
  services: string[]
}

export interface EdgeLocation {
  id: string
  name: string
  provider: string
  region: string
  city: string
  country: string
  status: 'online' | 'offline' | 'degraded'
  latency?: number
  capacity?: { current: number; max: number }
}

export interface Container {
  id: string
  name: string
  image: string
  tag: string
  status: 'running' | 'paused' | 'stopped' | 'restarting' | 'exited'
  ports?: Array<{ host: number; container: number }>
  usage?: { cpu?: number; memory?: number }
}

export interface SecretVault {
  id: string
  name: string
  type: 'aws-secrets-manager' | 'gcp-secret-manager' | 'azure-keyvault' | 'hashicorp-vault' | 'local'
  status: 'connected' | 'disconnected' | 'error'
  secrets: Array<{ key: string; masked: boolean; lastRotated?: number }>
}

export interface ScalableResource {
  id: string
  name: string
  type: 'function' | 'container' | 'service' | 'database'
  currentReplicas: number
  minReplicas: number
  maxReplicas: number
  targetCpuUtilization?: number
  status: 'scaled' | 'scaling' | 'error' | 'paused'
}

export interface TrafficTarget {
  id: string
  name: string
  type: 'function' | 'container' | 'external'
  weight: number
  status: 'healthy' | 'unhealthy' | 'unknown'
  latency?: number
  errorRate?: number
}

export interface RollbackVersion {
  id: string
  version: string
  deployedAt: number
  deployedBy: string
  status: 'current' | 'previous' | 'archived'
  healthScore?: number
}

export interface BuildArtifact {
  id: string
  name: string
  version: string
  type: 'binary' | 'image' | 'archive' | 'layer'
  size: number
  buildNumber: number
  status: 'available' | 'building' | 'failed'
}

export interface ClusterHealth {
  id: string
  name: string
  provider: string
  status: 'healthy' | 'degraded' | 'unhealthy'
  score: number
  components: Array<{ name: string; status: string; message?: string }>
  metrics?: { cpuUsage?: number; memoryUsage?: number }
}

export interface ColdStartMetric {
  functionId: string
  functionName: string
  region: string
  averageDuration: number
  p99Duration: number
  invocations: number
  runtime: string
}

export interface ExecutionFlow {
  id: string
  functionName: string
  region: string
  invocations: number
  avgDuration: number
  successRate: number
  coldStarts: number
}

export interface DevOpsState {
  // Pipeline
  pipeline: Pipeline | null
  selectedStageId: string | null

  // Environments
  environments: Environment[]
  activeEnvironmentId: string | null

  // Cloud Regions
  regions: CloudRegion[]
  selectedRegionId: string | null
  regionProviderFilter: 'aws' | 'gcp' | 'azure' | 'all'

  // Runtime Targets
  runtimeTargets: RuntimeTarget[]
  selectedRuntimeTargetId: string | null

  // Kubernetes
  k8sNodes: K8sNode[]
  k8sServices: K8sService[]
  k8sNamespaces: K8sNamespace[]
  selectedK8sNodeId: string | null

  // Edge
  edgeLocations: EdgeLocation[]
  selectedEdgeLocationId: string | null

  // Containers
  containers: Container[]
  selectedContainerId: string | null

  // Vaults
  vaults: SecretVault[]
  selectedVaultId: string | null

  // Resources
  scalableResources: ScalableResource[]
  selectedResourceId: string | null

  // Traffic
  trafficTargets: TrafficTarget[]
  selectedTrafficTargetId: string | null

  // Rollback
  rollbackVersions: RollbackVersion[]
  selectedRollbackVersionId: string | null

  // Artifacts
  artifacts: BuildArtifact[]
  selectedArtifactId: string | null

  // Clusters
  clusters: ClusterHealth[]
  selectedClusterId: string | null

  // Cold Start
  coldStartMetrics: ColdStartMetric[]
  selectedColdStartFunctionId: string | null

  // Execution Map
  executionFlows: ExecutionFlow[]
  selectedExecutionId: string | null

  // UI State
  activePanel: 'pipeline' | 'environments' | 'regions' | 'runtime' | 'kubernetes' | 'edge' | 'containers' | 'vaults' | 'resources' | 'traffic' | 'rollback' | 'artifacts' | 'clusters' | 'coldstart' | 'executions'
  sidebarCollapsed: boolean
}

export const useDevOpsStore = create<DevOpsState>()(
  immer((set) => ({
    // Pipeline
    pipeline: null,
    selectedStageId: null,
    setPipeline: (pipeline) => set((state) => { state.pipeline = pipeline }),
    selectStage: (stageId) => set((state) => { state.selectedStageId = stageId }),

    // Environments
    environments: [],
    activeEnvironmentId: null,
    setEnvironments: (envs) => set((state) => { state.environments = envs }),
    selectEnvironment: (envId) => set((state) => { state.activeEnvironmentId = envId }),
    addEnvironment: (env) => set((state) => { state.environments.push(env) }),
    updateEnvironment: (envId, updates) => set((state) => {
      const env = state.environments.find(e => e.id === envId)
      if (env) Object.assign(env, updates)
    }),
    deleteEnvironment: (envId) => set((state) => {
      state.environments = state.environments.filter(e => e.id !== envId)
    }),

    // Cloud Regions
    regions: [],
    selectedRegionId: null,
    regionProviderFilter: 'all',
    setRegions: (regions) => set((state) => { state.regions = regions }),
    selectRegion: (regionId) => set((state) => { state.selectedRegionId = regionId }),
    setRegionProviderFilter: (filter) => set((state) => { state.regionProviderFilter = filter }),

    // Runtime Targets
    runtimeTargets: [],
    selectedRuntimeTargetId: null,
    setRuntimeTargets: (targets) => set((state) => { state.runtimeTargets = targets }),
    selectRuntimeTarget: (targetId) => set((state) => { state.selectedRuntimeTargetId = targetId }),

    // Kubernetes
    k8sNodes: [],
    k8sServices: [],
    k8sNamespaces: [],
    selectedK8sNodeId: null,
    setKubernetes: (nodes, services, namespaces) => set((state) => {
      state.k8sNodes = nodes
      state.k8sServices = services
      state.k8sNamespaces = namespaces
    }),
    selectK8sNode: (nodeId) => set((state) => { state.selectedK8sNodeId = nodeId }),

    // Edge
    edgeLocations: [],
    selectedEdgeLocationId: null,
    setEdgeLocations: (locations) => set((state) => { state.edgeLocations = locations }),
    selectEdgeLocation: (locationId) => set((state) => { state.selectedEdgeLocationId = locationId }),

    // Containers
    containers: [],
    selectedContainerId: null,
    setContainers: (containers) => set((state) => { state.containers = containers }),
    selectContainer: (containerId) => set((state) => { state.selectedContainerId = containerId }),
    updateContainerStatus: (containerId, status) => set((state) => {
      const container = state.containers.find(c => c.id === containerId)
      if (container) container.status = status as Container['status']
    }),

    // Vaults
    vaults: [],
    selectedVaultId: null,
    setVaults: (vaults) => set((state) => { state.vaults = vaults }),
    selectVault: (vaultId) => set((state) => { state.selectedVaultId = vaultId }),
    addSecret: (vaultId, key) => set((state) => {
      const vault = state.vaults.find(v => v.id === vaultId)
      if (vault) vault.secrets.push({ key, masked: true, lastRotated: Date.now() })
    }),
    deleteSecret: (vaultId, key) => set((state) => {
      const vault = state.vaults.find(v => v.id === vaultId)
      if (vault) vault.secrets = vault.secrets.filter(s => s.key !== key)
    }),

    // Resources
    scalableResources: [],
    selectedResourceId: null,
    setScalableResources: (resources) => set((state) => { state.scalableResources = resources }),
    selectResource: (resourceId) => set((state) => { state.selectedResourceId = resourceId }),
    updateReplicas: (resourceId, replicas) => set((state) => {
      const resource = state.scalableResources.find(r => r.id === resourceId)
      if (resource) resource.currentReplicas = replicas
    }),

    // Traffic
    trafficTargets: [],
    selectedTrafficTargetId: null,
    setTrafficTargets: (targets) => set((state) => { state.trafficTargets = targets }),
    selectTrafficTarget: (targetId) => set((state) => { state.selectedTrafficTargetId = targetId }),
    updateTrafficWeight: (targetId, weight) => set((state) => {
      const target = state.trafficTargets.find(t => t.id === targetId)
      if (target) target.weight = weight
    }),

    // Rollback
    rollbackVersions: [],
    selectedRollbackVersionId: null,
    setRollbackVersions: (versions) => set((state) => { state.rollbackVersions = versions }),
    selectRollbackVersion: (versionId) => set((state) => { state.selectedRollbackVersionId = versionId }),

    // Artifacts
    artifacts: [],
    selectedArtifactId: null,
    setArtifacts: (artifacts) => set((state) => { state.artifacts = artifacts }),
    selectArtifact: (artifactId) => set((state) => { state.selectedArtifactId = artifactId }),

    // Clusters
    clusters: [],
    selectedClusterId: null,
    setClusters: (clusters) => set((state) => { state.clusters = clusters }),
    selectCluster: (clusterId) => set((state) => { state.selectedClusterId = clusterId }),

    // Cold Start
    coldStartMetrics: [],
    selectedColdStartFunctionId: null,
    setColdStartMetrics: (metrics) => set((state) => { state.coldStartMetrics = metrics }),
    selectColdStartFunction: (functionId) => set((state) => { state.selectedColdStartFunctionId = functionId }),

    // Execution Map
    executionFlows: [],
    selectedExecutionId: null,
    setExecutionFlows: (flows) => set((state) => { state.executionFlows = flows }),
    selectExecution: (executionId) => set((state) => { state.selectedExecutionId = executionId }),

    // UI State
    activePanel: 'pipeline',
    sidebarCollapsed: false,
    setActivePanel: (panel) => set((state) => { state.activePanel = panel }),
    toggleSidebar: () => set((state) => { state.sidebarCollapsed = !state.sidebarCollapsed }),
  }))
)

export const usePipeline = () => useDevOpsStore((state) => ({
  pipeline: state.pipeline,
  selectedStageId: state.selectedStageId,
}))

export const useEnvironments = () => useDevOpsStore((state) => ({
  environments: state.environments,
  activeId: state.activeEnvironmentId,
}))

export const useCloudRegions = () => useDevOpsStore((state) => ({
  regions: state.regions,
  selectedId: state.selectedRegionId,
  providerFilter: state.regionProviderFilter,
}))

export const useRuntimeTargets = () => useDevOpsStore((state) => ({
  targets: state.runtimeTargets,
  selectedId: state.selectedRuntimeTargetId,
}))

export const useKubernetes = () => useDevOpsStore((state) => ({
  nodes: state.k8sNodes,
  services: state.k8sServices,
  namespaces: state.k8sNamespaces,
  selectedNodeId: state.selectedK8sNodeId,
}))

export const useEdgeLocations = () => useDevOpsStore((state) => ({
  locations: state.edgeLocations,
  selectedId: state.selectedEdgeLocationId,
}))

export const useContainers = () => useDevOpsStore((state) => ({
  containers: state.containers,
  selectedId: state.selectedContainerId,
}))

export const useVaults = () => useDevOpsStore((state) => ({
  vaults: state.vaults,
  selectedId: state.selectedVaultId,
}))

export const useScalableResources = () => useDevOpsStore((state) => ({
  resources: state.scalableResources,
  selectedId: state.selectedResourceId,
}))

export const useTrafficBalancer = () => useDevOpsStore((state) => ({
  targets: state.trafficTargets,
  selectedId: state.selectedTrafficTargetId,
}))

export const useRollbackManager = () => useDevOpsStore((state) => ({
  versions: state.rollbackVersions,
  selectedId: state.selectedRollbackVersionId,
}))

export const useBuildArtifacts = () => useDevOpsStore((state) => ({
  artifacts: state.artifacts,
  selectedId: state.selectedArtifactId,
}))

export const useClusterHealth = () => useDevOpsStore((state) => ({
  clusters: state.clusters,
  selectedId: state.selectedClusterId,
}))

export const useColdStartAnalyzer = () => useDevOpsStore((state) => ({
  metrics: state.coldStartMetrics,
  selectedId: state.selectedColdStartFunctionId,
}))

export const useServerlessExecutionMap = () => useDevOpsStore((state) => ({
  flows: state.executionFlows,
  selectedId: state.selectedExecutionId,
}))

export const useDevOpsUI = () => useDevOpsStore((state) => ({
  activePanel: state.activePanel,
  sidebarCollapsed: state.sidebarCollapsed,
}))
