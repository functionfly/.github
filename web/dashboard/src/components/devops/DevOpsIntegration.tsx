/**
 * DevOps Integration Component
 * Unified panel that wires all DevOps & Infrastructure components together
 */

import React, { useState, useCallback, useMemo } from 'react'
import { cn } from '@functionfly/ui-core'
import {
  Rocket,
  Server,
  Cloud,
  Container,
  Globe,
  MapPin,
  Box,
  Key,
  Shield,
  Lock,
  Scaling,
  Route,
  RotateCw,
  FlaskConical,
  Package,
  Activity,
  Zap,
  ServerIcon,
  DatabaseIcon,
  MonitorIcon,
  LayersIcon,
  GitFork,
  ChevronRight,
  ChevronDown,
  Settings,
  Plus,
  Trash2,
  Play,
  Pause,
  RefreshCw,
  Download,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Clock,
  TrafficCone,
  BarChart3,
  PieChart,
  LineChart,
  HardDrive,
  Cpu,
  MemoryStick,
  Network,
  GitBranch,
  GitCommit,
} from 'lucide-react'

// Import components from the devops package
import {
  DeploymentPipeline,
  EnvironmentManager,
  CloudRegionSelector,
  RuntimeTargetSelector,
  ContainerTopologyView,
  EdgeDeploymentMap,
  ContainerLifecyclePanel,
  SecretVaultManager,
  InfrastructureDiffViewer,
  ResourceScaler,
  TrafficBalancerView,
  RollbackManager,
  DeploymentSimulation,
  BuildArtifactExplorer,
  ClusterHealthMonitor,
  ColdStartAnalyzer,
  ServerlessExecutionMap,
} from '@functionfly/ui-devops'

// Panel navigation items
const NAV_ITEMS = [
  { id: 'pipeline', label: 'Pipeline', icon: Rocket },
  { id: 'environments', label: 'Environments', icon: Server },
  { id: 'regions', label: 'Regions', icon: Globe },
  { id: 'runtime', label: 'Runtime', icon: Cpu },
  { id: 'kubernetes', label: 'Container', icon: Container },
  { id: 'edge', label: 'Edge', icon: MapPin },
  { id: 'containers', label: 'Containers', icon: Container },
  { id: 'vaults', label: 'Vaults', icon: Lock },
  { id: 'resources', label: 'Resources', icon: Scaling },
  { id: 'traffic', label: 'Traffic', icon: Route },
  { id: 'rollback', label: 'Rollback', icon: RotateCw },
  { id: 'artifacts', label: 'Artifacts', icon: Package },
  { id: 'clusters', label: 'Clusters', icon: MonitorIcon },
  { id: 'coldstart', label: 'Cold Start', icon: Zap },
  { id: 'executions', label: 'Executions', icon: Activity },
] as const

type PanelId = typeof NAV_ITEMS[number]['id']

interface DevOpsIntegrationProps {
  className?: string
}

export const DevOpsIntegration: React.FC<DevOpsIntegrationProps> = ({
  className,
}) => {
  const [activePanel, setActivePanel] = useState<PanelId>('pipeline')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  // Mock data - Pipeline
  const mockPipeline = useMemo(() => ({
    id: 'pipeline-1',
    name: 'Production Deployment',
    version: 'v2.3.1',
    status: 'active' as const,
    stages: [
      {
        id: 'stage-1',
        name: 'Build',
        status: 'completed' as const,
        duration: 45000,
        tasks: [
          { id: 't1', name: 'Compile TypeScript', status: 'completed', duration: 12000 },
          { id: 't2', name: 'Run Tests', status: 'completed', duration: 25000 },
          { id: 't3', name: 'Build Docker Image', status: 'completed', duration: 8000 },
        ],
      },
      {
        id: 'stage-2',
        name: 'Test',
        status: 'completed' as const,
        duration: 120000,
        tasks: [
          { id: 't4', name: 'Unit Tests', status: 'completed', duration: 30000 },
          { id: 't5', name: 'Integration Tests', status: 'completed', duration: 60000 },
          { id: 't6', name: 'E2E Tests', status: 'completed', duration: 30000 },
        ],
      },
      {
        id: 'stage-3',
        name: 'Deploy to Staging',
        status: 'running' as const,
        tasks: [
          { id: 't7', name: 'Deploy Container', status: 'running' },
          { id: 't8', name: 'Health Check', status: 'pending' },
        ],
      },
      {
        id: 'stage-4',
        name: 'Deploy to Production',
        status: 'pending' as const,
        tasks: [
          { id: 't9', name: 'Blue-Green Switch', status: 'pending' },
          { id: 't10', name: 'Final Verification', status: 'pending' },
        ],
      },
    ],
    triggeredBy: 'Sarah Chen',
    triggeredAt: Date.now() - 300000,
    branch: 'main',
    commitSha: 'a1b2c3d4e5f6',
    source: 'webhook' as const,
  }), [])

  // Mock data - Environments
  const mockEnvironments = useMemo(() => [
    {
      id: 'env-1',
      name: 'Production',
      type: 'production' as const,
      color: '#ef4444',
      variables: { API_URL: 'https://api.production.com', LOG_LEVEL: 'info' },
      secrets: [{ key: 'DATABASE_URL', masked: true, lastUpdated: Date.now() - 86400000 }],
      replicas: 5,
      autoScale: true,
      region: 'us-east-1',
    },
    {
      id: 'env-2',
      name: 'Staging',
      type: 'staging' as const,
      color: '#f59e0b',
      variables: { API_URL: 'https://api.staging.com', LOG_LEVEL: 'debug' },
      secrets: [{ key: 'DATABASE_URL', masked: true, lastUpdated: Date.now() - 172800000 }],
      replicas: 2,
      autoScale: false,
      region: 'us-east-1',
    },
    {
      id: 'env-3',
      name: 'Development',
      type: 'development' as const,
      color: '#22c55e',
      variables: { API_URL: 'http://localhost:3000', LOG_LEVEL: 'debug' },
      secrets: [],
      replicas: 1,
      autoScale: false,
    },
  ], [])

  // Mock data - Regions
  const mockRegions = useMemo(() => [
    { id: 'r1', name: 'US East', provider: 'aws' as const, zone: 'us-east-1', zoneName: 'N. Virginia', location: 'Ashburn', country: 'USA', coordinates: { lat: 39.01, lng: -77.36 }, isAvailable: true, isRecommended: true, specs: { compute: 64, memory: 256, storage: 1000, gpu: true } },
    { id: 'r2', name: 'EU West', provider: 'aws' as const, zone: 'eu-west-1', zoneName: 'Ireland', location: 'Dublin', country: 'Ireland', coordinates: { lat: 53.33, lng: -6.26 }, isAvailable: true, specs: { compute: 48, memory: 192, storage: 750 } },
    { id: 'r3', name: 'AP Southeast', provider: 'aws' as const, zone: 'ap-southeast-1', zoneName: 'Singapore', location: 'Singapore', country: 'Singapore', coordinates: { lat: 1.35, lng: 103.82 }, isAvailable: true, specs: { compute: 32, memory: 128, storage: 500 } },
    { id: 'r4', name: 'US West', provider: 'gcp' as const, zone: 'us-west1', zoneName: 'Oregon', location: 'The Dalles', country: 'USA', coordinates: { lat: 45.72, lng: -121.17 }, isAvailable: true, specs: { compute: 56, memory: 224, storage: 800, gpu: true } },
    { id: 'r5', name: 'EU Central', provider: 'gcp' as const, zone: 'europe-central2', zoneName: 'Warsaw', location: 'Warsaw', country: 'Poland', coordinates: { lat: 52.23, lng: 21.01 }, isAvailable: true, specs: { compute: 40, memory: 160, storage: 600 } },
    { id: 'r6', name: 'East US', provider: 'azure' as const, zone: 'eastus', zoneName: 'Virginia', location: 'Washington D.C.', country: 'USA', coordinates: { lat: 38.95, lng: -77.37 }, isAvailable: true, specs: { compute: 48, memory: 192, storage: 700 } },
  ], [])

  // Mock data - Runtime Targets
  const mockRuntimes = useMemo(() => [
    { id: 'rt-1', name: 'Node.js 20', type: 'nodejs' as const, version: '20.10.0', status: 'stable' as const, memoryLimit: 512, timeout: 30 },
    { id: 'rt-2', name: 'Python 3.12', type: 'python' as const, version: '3.12.0', status: 'stable' as const, memoryLimit: 1024, timeout: 60 },
    { id: 'rt-3', name: 'Go 1.21', type: 'go' as const, version: '1.21.5', status: 'stable' as const, memoryLimit: 256, timeout: 30 },
    { id: 'rt-4', name: 'Rust 1.75', type: 'rust' as const, version: '1.75.0', status: 'beta' as const, memoryLimit: 128, timeout: 10 },
    { id: 'rt-5', name: 'Bun 1.0', type: 'bun' as const, version: '1.0.12', status: 'stable' as const, memoryLimit: 512, timeout: 30 },
    { id: 'rt-6', name: 'Deno 1.40', type: 'deno' as const, version: '1.40.0', status: 'beta' as const, memoryLimit: 512, timeout: 60 },
    { id: 'rt-7', name: 'SAR 2.0', type: 'sar' as const, version: '2.0.1', status: 'stable' as const, memoryLimit: 1024, timeout: 300 },
  ], [])

  // Mock data - Kubernetes
  const mockK8s = useMemo(() => ({
    nodes: [
      { id: 'n1', name: 'control-plane-1', type: 'control-plane' as const, status: 'ready' as const, resources: { cpu: 45, memory: 62, pods: 12 } },
      { id: 'n2', name: 'worker-node-1', type: 'worker' as const, status: 'ready' as const, resources: { cpu: 78, memory: 85, pods: 24 } },
      { id: 'n3', name: 'worker-node-2', type: 'worker' as const, status: 'ready' as const, resources: { cpu: 65, memory: 72, pods: 18 } },
      { id: 'n4', name: 'ingress-nginx', type: 'ingress' as const, status: 'ready' as const },
    ],
    services: [
      { id: 's1', name: 'api-service', type: 'load-balancer' as const, selector: { app: 'api' }, ports: [{ name: 'http', port: 80, targetPort: 8080 }], clusterIp: '10.96.0.1' },
      { id: 's2', name: 'web-frontend', type: 'load-balancer' as const, selector: { app: 'web' }, ports: [{ name: 'http', port: 80, targetPort: 3000 }], clusterIp: '10.96.0.2' },
    ],
    namespaces: [
      { id: 'ns1', name: 'production', status: 'active' as const, pods: ['pod-1', 'pod-2'], services: ['s1'] },
      { id: 'ns2', name: 'staging', status: 'active' as const, pods: ['pod-3'], services: [] },
    ],
  }), [])

  // Mock data - Edge Locations
  const mockEdgeLocations = useMemo(() => [
    { id: 'e1', name: 'NYC Edge', provider: 'Cloudflare', region: 'us-east', city: 'New York', country: 'USA', status: 'online' as const, latency: 12, capacity: { current: 450, max: 1000 } },
    { id: 'e2', name: 'LAX Edge', provider: 'Cloudflare', region: 'us-west', city: 'Los Angeles', country: 'USA', status: 'online' as const, latency: 25, capacity: { current: 320, max: 1000 } },
    { id: 'e3', name: 'LHR Edge', provider: 'Cloudflare', region: 'eu-west', city: 'London', country: 'UK', status: 'degraded' as const, latency: 85, capacity: { current: 780, max: 1000 } },
    { id: 'e4', name: 'TYO Edge', provider: 'Cloudflare', region: 'ap-northeast', city: 'Tokyo', country: 'Japan', status: 'online' as const, latency: 45, capacity: { current: 180, max: 500 } },
    { id: 'e5', name: 'SGP Edge', provider: 'Cloudflare', region: 'ap-southeast', city: 'Singapore', country: 'Singapore', status: 'online' as const, latency: 38, capacity: { current: 290, max: 500 } },
  ], [])

  // Mock data - Containers
  const mockContainers = useMemo(() => [
    { id: 'c1', name: 'api-server', image: 'functionfly/api:latest', status: 'running' as const, ports: [{ host: 8080, container: 3000 }], usage: { cpu: 45, memory: 62 } },
    { id: 'c2', name: 'worker-1', image: 'functionfly/worker:latest', status: 'running' as const, usage: { cpu: 78, memory: 85 } },
    { id: 'c3', name: 'postgres-db', image: 'postgres:15', status: 'running' as const, ports: [{ host: 5432, container: 5432 }], usage: { cpu: 15, memory: 45 } },
    { id: 'c4', name: 'redis-cache', image: 'redis:7-alpine', status: 'paused' as const, usage: { cpu: 5, memory: 22 } },
    { id: 'c5', name: 'nginx-proxy', image: 'nginx:latest', status: 'stopped' as const },
  ], [])

  // Mock data - Vaults
  const mockVaults = useMemo(() => [
    { id: 'v1', name: 'AWS Secrets Manager', type: 'aws-secrets-manager' as const, status: 'connected' as const, secrets: [{ key: 'DATABASE_PASSWORD', masked: true, lastRotated: Date.now() - 86400000 }, { key: 'API_KEY', masked: true }] },
    { id: 'v2', name: 'GCP Secret Manager', type: 'gcp-secret-manager' as const, status: 'connected' as const, secrets: [{ key: 'JWT_SECRET', masked: true, lastRotated: Date.now() - 604800000 }] },
    { id: 'v3', name: 'HashiCorp Vault', type: 'hashicorp-vault' as const, status: 'disconnected' as const, secrets: [] },
  ], [])

  // Mock data - Resources
  const mockResources = useMemo(() => [
    { id: 'res-1', name: 'API Service', type: 'container' as const, currentReplicas: 3, minReplicas: 1, maxReplicas: 10, targetCpuUtilization: 70, status: 'scaled' as const },
    { id: 'res-2', name: 'Background Worker', type: 'function' as const, currentReplicas: 5, minReplicas: 1, maxReplicas: 20, targetCpuUtilization: 60, status: 'scaled' as const },
    { id: 'res-3', name: 'PostgreSQL', type: 'database' as const, currentReplicas: 1, minReplicas: 1, maxReplicas: 3, targetCpuUtilization: 80, status: 'paused' as const },
  ], [])

  // Mock data - Traffic
  const mockTrafficTargets = useMemo(() => [
    { id: 'tt-1', name: 'API v2.3', type: 'function' as const, weight: 80, status: 'healthy' as const, latency: 45, errorRate: 0.1 },
    { id: 'tt-2', name: 'API v2.2', type: 'function' as const, weight: 20, status: 'healthy' as const, latency: 52, errorRate: 0.3 },
  ], [])

  // Mock data - Rollback
  const mockRollbackVersions = useMemo(() => [
    { id: 'rv-1', version: 'v2.3.1', deployedAt: Date.now() - 86400000, deployedBy: 'Sarah Chen', status: 'current' as const, healthScore: 98 },
    { id: 'rv-2', version: 'v2.3.0', deployedAt: Date.now() - 604800000, deployedBy: 'Mike Johnson', status: 'previous' as const, healthScore: 95 },
    { id: 'rv-3', version: 'v2.2.5', deployedAt: Date.now() - 1209600000, deployedBy: 'Alex Rivera', status: 'archived' as const, healthScore: 92 },
  ], [])

  // Mock data - Artifacts
  const mockArtifacts = useMemo(() => [
    { id: 'a1', name: 'api-server', version: '2.3.1', type: 'image' as const, size: 245760000, buildNumber: 142, status: 'available' as const },
    { id: 'a2', name: 'worker', version: '2.3.1', type: 'image' as const, size: 182400000, buildNumber: 142, status: 'available' as const },
    { id: 'a3', name: 'deployment-yaml', version: '2.3.1', type: 'archive' as const, size: 4096, buildNumber: 142, status: 'available' as const },
  ], [])

  // Mock data - Clusters
  const mockClusters = useMemo(() => [
    { id: 'cl-1', name: 'Production (AWS)', provider: 'aws', status: 'healthy' as const, score: 96, components: [{ name: 'EC2 Instances', status: 'healthy' as const }, { name: 'RDS Database', status: 'healthy' as const }], metrics: { cpuUsage: 62, memoryUsage: 74 } },
    { id: 'cl-2', name: 'Staging (GCP)', provider: 'gcp', status: 'degraded' as const, score: 78, components: [{ name: 'GKE Cluster', status: 'degraded' as const, message: 'High memory usage' }], metrics: { cpuUsage: 85, memoryUsage: 91 } },
  ], [])

  // Mock data - Cold Start
  const mockColdStartMetrics = useMemo(() => [
    { functionId: 'fn-1', functionName: 'processImage', region: 'us-east-1', averageDuration: 250, p99Duration: 890, invocations: 15420, runtime: 'Node.js 20' },
    { functionId: 'fn-2', functionName: 'generateReport', region: 'us-east-1', averageDuration: 180, p99Duration: 420, invocations: 8920, runtime: 'Python 3.12' },
    { functionId: 'fn-3', functionName: 'sendEmail', region: 'eu-west-1', averageDuration: 95, p99Duration: 280, invocations: 45200, runtime: 'Node.js 20' },
    { functionId: 'fn-4', functionName: 'resizeImage', region: 'us-west-1', averageDuration: 520, p99Duration: 1200, invocations: 3200, runtime: 'Rust 1.75' },
  ], [])

  // Mock data - Execution Map
  const mockExecutionFlows = useMemo(() => [
    { id: 'ef-1', functionName: 'User Authentication', region: 'us-east-1', invocations: 125000, avgDuration: 45, successRate: 99.8, coldStarts: 120 },
    { id: 'ef-2', functionName: 'Payment Processing', region: 'us-east-1', invocations: 8500, avgDuration: 380, successRate: 99.2, coldStarts: 45 },
    { id: 'ef-3', functionName: 'Email Delivery', region: 'eu-west-1', invocations: 45000, avgDuration: 120, successRate: 97.5, coldStarts: 89 },
    { id: 'ef-4', functionName: 'Data Analytics', region: 'ap-southeast-1', invocations: 1200, avgDuration: 2500, successRate: 95.8, coldStarts: 12 },
  ], [])

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Navigation Sidebar */}
      <div className={cn(
        'flex flex-col border-r border-aviation-border-panel transition-all duration-300',
        sidebarCollapsed ? 'w-12' : 'w-56'
      )}>
        <div className="flex items-center justify-end px-2 py-2 border-b border-aviation-border-panel">
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="p-1.5 hover:bg-aviation-bg-instrument rounded"
          >
            {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
        </div>
        <nav className="flex-1 overflow-auto py-2">
          {NAV_ITEMS.map((item) => {
            const Icon = item.icon
            const isActive = activePanel === item.id
            return (
              <button
                key={item.id}
                onClick={() => setActivePanel(item.id)}
                className={cn(
                  'flex items-center gap-3 w-full px-3 py-2 text-left transition-colors',
                  isActive ? 'bg-aviation-cyan/20 text-aviation-cyan border-l-2 border-aviation-cyan' : 'text-aviation-text-muted hover:text-aviation-text-primary hover:bg-aviation-bg-secondary',
                  sidebarCollapsed && 'justify-center px-0'
                )}
                title={sidebarCollapsed ? item.label : undefined}
              >
                <Icon className="w-4 h-4 flex-shrink-0" />
                {!sidebarCollapsed && <span className="text-sm truncate">{item.label}</span>}
              </button>
            )
          })}
        </nav>
        {!sidebarCollapsed && (
          <div className="px-3 py-2 border-t border-aviation-border-panel">
            <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
              <div className="w-2 h-2 rounded-full bg-green-400" />
              <span>Systems Normal</span>
            </div>
          </div>
        )}
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center gap-2">
            <LayersIcon className="w-5 h-5 text-aviation-cyan" />
            <span className="text-sm font-medium">DevOps & Infrastructure</span>
          </div>
          <span className="text-xs text-aviation-text-muted">
            {NAV_ITEMS.find(i => i.id === activePanel)?.label}
          </span>
        </div>

        <div className="flex-1 overflow-hidden">
          {activePanel === 'pipeline' && (
            <DeploymentPipeline
              pipeline={mockPipeline}
              className="h-full"
            />
          )}
          {activePanel === 'environments' && (
            <EnvironmentManager
              environments={mockEnvironments}
              className="h-full"
            />
          )}
          {activePanel === 'regions' && (
            <CloudRegionSelector
              regions={mockRegions}
              className="h-full"
            />
          )}
          {activePanel === 'runtime' && (
            <RuntimeTargetSelector
              targets={mockRuntimes}
              className="h-full"
            />
          )}
          {activePanel === 'kubernetes' && (
            <ContainerTopologyView
              nodes={mockK8s.nodes}
              services={mockK8s.services}
              namespaces={mockK8s.namespaces}
              className="h-full"
            />
          )}
          {activePanel === 'edge' && (
            <EdgeDeploymentMap
              locations={mockEdgeLocations}
              className="h-full"
            />
          )}
          {activePanel === 'containers' && (
            <ContainerLifecyclePanel
              containers={mockContainers}
              className="h-full"
            />
          )}
          {activePanel === 'vaults' && (
            <SecretVaultManager
              vaults={mockVaults}
              className="h-full"
            />
          )}
          {activePanel === 'resources' && (
            <ResourceScaler
              resources={mockResources}
              className="h-full"
            />
          )}
          {activePanel === 'traffic' && (
            <TrafficBalancerView
              balancerName="Main Load Balancer"
              targets={mockTrafficTargets}
              className="h-full"
            />
          )}
          {activePanel === 'rollback' && (
            <RollbackManager
              resourceId="api-server"
              resourceName="API Server"
              versions={mockRollbackVersions}
              className="h-full"
            />
          )}
          {activePanel === 'artifacts' && (
            <BuildArtifactExplorer
              artifacts={mockArtifacts}
              className="h-full"
            />
          )}
          {activePanel === 'clusters' && (
            <ClusterHealthMonitor
              clusters={mockClusters}
              className="h-full"
            />
          )}
          {activePanel === 'coldstart' && (
            <ColdStartAnalyzer
              metrics={mockColdStartMetrics}
              className="h-full"
            />
          )}
          {activePanel === 'executions' && (
            <ServerlessExecutionMap
              executions={mockExecutionFlows}
              className="h-full"
            />
          )}
        </div>
      </div>
    </div>
  )
}

export default DevOpsIntegration
