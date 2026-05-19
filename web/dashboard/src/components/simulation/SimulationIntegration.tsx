/**
 * Simulation Integration Component
 * Unified panel that wires all R-Sim components together
 */

import React, { useState, useCallback, useMemo } from 'react'
import { cn } from '@functionfly/ui-core'
import {
  Gauge,
  TrendingUp,
  AlertTriangle,
  Timer,
  AlertCircle,
  CheckCircle,
  XCircle,
  Brain,
  Zap,
  Clock,
  Target,
  BarChart3,
  PieChart,
  LineChart,
  AreaChart,
  Hexagon,
  Circle,
  Box,
  Database,
  Cpu,
  Server,
  Network,
  GitFork,
  GitBranch,
  GitMerge,
  Users,
  Bot,
  MessageSquare,
  FileText,
  FastForward,
  Rewind,
  SkipForward,
  SkipBack,
  ChevronRight,
  ChevronDown,
  Play,
  Pause,
  Square,
  RotateCcw,
  RefreshCw,
  Flame,
  DollarSign,
  Activity,
  Layers,
  Workflow,
} from 'lucide-react'

// Import simulation components
import {
  SimulationControlCenter,
  ExecutionForecastPanel,
  FailureProbabilityMap,
  LatencyPredictionGraph,
  HallucinationRiskAnalyzer,
  CostSimulationChart,
  StressTestRunner,
  ScalingForecastMap,
  AgentBehaviorPredictor,
  WorkflowOutcomeSimulator,
  ResourceCollisionDetector,
} from '@functionfly/ui-simulation'

// Panel navigation items
const NAV_ITEMS = [
  { id: 'control', label: 'Control Center', icon: Gauge },
  { id: 'forecast', label: 'Execution Forecast', icon: TrendingUp },
  { id: 'failure', label: 'Failure Map', icon: AlertTriangle },
  { id: 'latency', label: 'Latency Prediction', icon: Timer },
  { id: 'hallucination', label: 'Hallucination Risk', icon: Brain },
  { id: 'cost', label: 'Cost Simulation', icon: DollarSign },
  { id: 'stress', label: 'Stress Test', icon: Flame },
  { id: 'scaling', label: 'Scaling Forecast', icon: Activity },
  { id: 'behavior', label: 'Agent Behavior', icon: Bot },
  { id: 'workflow', label: 'Workflow Outcomes', icon: Workflow },
  { id: 'collision', label: 'Resource Collision', icon: AlertCircle },
] as const

type PanelId = typeof NAV_ITEMS[number]['id']

// Mock data generators
const generateMockConfig = () => ({
  id: 'sim-1',
  name: 'Load Test Simulation',
  type: 'load' as const,
  duration: 120,
  parallelism: 50,
  rampUpTime: 30,
})

const generateMockMetrics = () => ({
  requestsTotal: 15420,
  requestsSuccess: 15180,
  requestsFailed: 240,
  avgLatency: 145,
  p50Latency: 98,
  p95Latency: 312,
  p99Latency: 489,
  maxLatency: 892,
  throughput: 128.5,
  errorRate: 1.56,
  timestamp: Date.now(),
})

const generateMockForecasts = () => {
  const now = Date.now()
  return Array.from({ length: 30 }, (_, i) => ({
    timestamp: now + i * 60000,
    expectedLatency: 100 + Math.random() * 50 + Math.sin(i / 5) * 30,
    confidence: 0.85 + Math.random() * 0.1,
    upperBound: 150 + Math.random() * 40,
    lowerBound: 70 + Math.random() * 30,
  }))
}

const generateMockFailureNodes = () => [
  { id: 'fn-1', name: 'Auth Service', type: 'service' as const, failureProbability: 0.75, affectedRequests: 2340 },
  { id: 'fn-2', name: 'User DB', type: 'database' as const, failureProbability: 0.45, affectedRequests: 890 },
  { id: 'fn-3', name: 'Cache Cluster', type: 'cache' as const, failureProbability: 0.82, affectedRequests: 4500, correlationId: 'fn-1' },
  { id: 'fn-4', name: 'API Gateway', type: 'endpoint' as const, failureProbability: 0.23 },
  { id: 'fn-5', name: 'Message Queue', type: 'queue' as const, failureProbability: 0.56, affectedRequests: 1200 },
  { id: 'fn-6', name: 'Worker Pool', type: 'worker' as const, failureProbability: 0.18 },
]

const generateMockLatencyData = () => {
  const now = Date.now()
  return Array.from({ length: 60 }, (_, i) => ({
    timestamp: now - (60 - i) * 10000,
    predicted: 100 + Math.random() * 40 + Math.sin(i / 10) * 20,
    actual: 95 + Math.random() * 50,
    p50: 85 + Math.random() * 30,
    p95: 180 + Math.random() * 50,
    p99: 280 + Math.random() * 80,
  }))
}

const generateMockHallucinationRisks = () => [
  {
    id: 'hr-1',
    source: 'user-context-v3',
    type: 'context' as const,
    severity: 'critical' as const,
    confidence: 0.89,
    description: 'Conflicting user preferences detected - theme settings may be outdated',
    indicators: ['contradiction', 'temporal mismatch', 'context decay'],
    mitigationSuggestion: 'Refresh user context from authoritative source',
  },
  {
    id: 'hr-2',
    source: 'training-data-2024',
    type: 'training' as const,
    severity: 'high' as const,
    confidence: 0.76,
    description: 'Model may be hallucinating based on outdated training patterns',
    indicators: ['pattern mismatch', 'low confidence'],
  },
  {
    id: 'hr-3',
    source: 'retrieval-system',
    type: 'retrieval' as const,
    severity: 'medium' as const,
    confidence: 0.62,
    description: 'Retrieved facts may not match current system state',
  },
]

const generateMockCostProjections = () => {
  const now = Date.now()
  let cumulative = 0
  return Array.from({ length: 24 }, (_, i) => {
    const computeCost = 0.05 + Math.random() * 0.03
    const memoryCost = 0.02 + Math.random() * 0.01
    const networkCost = 0.01 + Math.random() * 0.005
    const storageCost = 0.005 + Math.random() * 0.002
    const totalCost = computeCost + memoryCost + networkCost + storageCost
    cumulative += totalCost
    return {
      timestamp: now + i * 3600000,
      computeCost,
      memoryCost,
      networkCost,
      storageCost,
      totalCost,
      cumulativeCost: cumulative,
    }
  })
}

const generateMockStressTestResults = () => [
  {
    id: 'stress-1',
    timestamp: Date.now() - 3600000,
    duration: 180,
    peakLoad: 500,
    steadyStateLoad: 350,
    successRate: 97.5,
    avgResponseTime: 124,
    maxResponseTime: 890,
    errors: [{ type: 'timeout', count: 12 }, { type: 'connection_reset', count: 5 }],
    bottlenecks: ['database_pool', 'connection_limit'],
  },
  {
    id: 'stress-2',
    timestamp: Date.now() - 7200000,
    duration: 300,
    peakLoad: 1000,
    successRate: 94.2,
    avgResponseTime: 198,
    maxResponseTime: 1200,
    errors: [{ type: 'memory_exceeded', count: 23 }],
    bottlenecks: ['worker_threads'],
  },
]

const generateMockScalingProjections = () => {
  const now = Date.now()
  return Array.from({ length: 20 }, (_, i) => ({
    timestamp: now + i * 300000,
    currentReplicas: 5 + Math.floor(i / 4),
    predictedReplicas: 5 + Math.floor(i / 4) + (i > 10 ? 3 : 0),
    confidence: 0.85 + Math.random() * 0.1,
    trigger: (['cpu', 'memory', 'requests'] as const)[i % 3],
    estimatedCostPerHour: 0.12 + i * 0.01,
  }))
}

const generateMockAgentPredictions = () => [
  {
    agentId: 'agent-1',
    agentName: 'Auth Agent',
    predictedActions: [
      { action: 'Refresh tokens', probability: 0.92, expectedOutcome: 'Tokens refreshed successfully', confidence: 0.95, timestamp: Date.now() + 30000 },
      { action: 'Validate session', probability: 0.88, expectedOutcome: 'Session valid', confidence: 0.91, timestamp: Date.now() + 60000 },
    ],
    riskScore: 0.15,
    recommendedInterventions: ['Pre-authorize backup tokens'],
  },
  {
    agentId: 'agent-2',
    agentName: 'Billing Agent',
    predictedActions: [
      { action: 'Process pending invoices', probability: 0.78, expectedOutcome: '12 invoices processed', confidence: 0.82, timestamp: Date.now() + 45000 },
    ],
    riskScore: 0.42,
    recommendedInterventions: ['Review rate limits', 'Check API quotas'],
  },
]

const generateMockWorkflowPaths = () => [
  {
    id: 'path-1',
    name: 'Standard Checkout',
    probability: 0.72,
    steps: [
      { name: 'Validate Cart', duration: 50, successProbability: 0.98 },
      { name: 'Check Inventory', duration: 120, successProbability: 0.95 },
      { name: 'Process Payment', duration: 300, successProbability: 0.92 },
      { name: 'Confirm Order', duration: 80, successProbability: 0.99 },
    ],
    totalDuration: 550,
    expectedCost: 0.045,
  },
  {
    id: 'path-2',
    name: 'Express Checkout',
    probability: 0.18,
    steps: [
      { name: 'Validate Cart', duration: 50, successProbability: 0.98 },
      { name: 'Express Payment', duration: 150, successProbability: 0.94 },
      { name: 'Quick Confirm', duration: 40, successProbability: 0.99 },
    ],
    totalDuration: 240,
    expectedCost: 0.028,
  },
  {
    id: 'path-3',
    name: 'Retry Checkout',
    probability: 0.10,
    steps: [
      { name: 'Validate Cart', duration: 50, successProbability: 0.98 },
      { name: 'Check Inventory', duration: 120, successProbability: 0.95 },
      { name: 'Retry Payment', duration: 400, successProbability: 0.88 },
      { name: 'Confirm Order', duration: 80, successProbability: 0.99 },
    ],
    totalDuration: 650,
    expectedCost: 0.055,
  },
]

const generateMockResourceCollisions = () => [
  {
    id: 'rc-1',
    resourceA: 'CPU-Cluster-1',
    resourceB: 'Memory-Node-3',
    type: 'cpu' as const,
    severity: 'high' as const,
    probability: 0.72,
    impact: 'High latency on batch processing jobs',
    mitigation: 'Increase cooldown period between jobs',
  },
  {
    id: 'rc-2',
    resourceA: 'Network-Link-A',
    resourceB: 'Disk-IO-Controller',
    type: 'network' as const,
    severity: 'medium' as const,
    probability: 0.58,
    impact: 'Slight throughput degradation during backup',
  },
]

interface SimulationIntegrationProps {
  className?: string
}

export const SimulationIntegration: React.FC<SimulationIntegrationProps> = ({ className }) => {
  const [activePanel, setActivePanel] = useState<PanelId>('control')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [simulationStatus, setSimulationStatus] = useState<'idle' | 'preparing' | 'running' | 'paused' | 'completed' | 'failed'>('idle')

  // Mock data states
  const [mockConfig] = useState(generateMockConfig)
  const [mockMetrics] = useState(generateMockMetrics)
  const [mockForecasts] = useState(generateMockForecasts)
  const [mockFailureNodes] = useState(generateMockFailureNodes)
  const [mockLatencyData] = useState(generateMockLatencyData)
  const [mockHallucinationRisks] = useState(generateMockHallucinationRisks)
  const [mockCostProjections] = useState(generateMockCostProjections)
  const [mockStressTestResults] = useState(generateMockStressTestResults)
  const [mockScalingProjections] = useState(generateMockScalingProjections)
  const [mockAgentPredictions] = useState(generateMockAgentPredictions)
  const [mockWorkflowPaths] = useState(generateMockWorkflowPaths)
  const [mockResourceCollisions] = useState(generateMockResourceCollisions)

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Navigation Sidebar */}
      <div className={cn(
        'flex flex-col border-r border-aviation-border-panel transition-all duration-300',
        sidebarCollapsed ? 'w-12' : 'w-56'
      )}>
        {/* Collapse Toggle */}
        <div className="flex items-center justify-end px-2 py-2 border-b border-aviation-border-panel">
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="p-1.5 hover:bg-aviation-bg-instrument rounded transition-colors"
          >
            {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
        </div>

        {/* Navigation Items */}
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

        {/* Status Indicator */}
        {!sidebarCollapsed && (
          <div className="px-3 py-2 border-t border-aviation-border-panel">
            <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
              <div className={cn('w-2 h-2 rounded-full', {
                'bg-green-400 animate-pulse': simulationStatus === 'running',
                'bg-amber-400': simulationStatus === 'paused',
                'bg-aviation-cyan': simulationStatus === 'completed',
                'bg-red-400': simulationStatus === 'failed',
                'bg-aviation-text-muted': simulationStatus === 'idle',
              })} />
              <span>R-Sim {simulationStatus}</span>
            </div>
          </div>
        )}
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center gap-2">
            <Activity className="w-5 h-5 text-aviation-cyan" />
            <span className="text-sm font-medium">R-Sim Runtime Simulation</span>
          </div>
          <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
            <Gauge className="w-4 h-4" />
            <span>{NAV_ITEMS.find(i => i.id === activePanel)?.label}</span>
          </div>
        </div>

        {/* Content Panel */}
        <div className="flex-1 overflow-hidden">
          {activePanel === 'control' && (
            <SimulationControlCenter
              config={mockConfig}
              status={simulationStatus}
              metrics={mockMetrics}
              onStart={() => setSimulationStatus('running')}
              onPause={() => setSimulationStatus('paused')}
              onStop={() => setSimulationStatus('idle')}
              onReset={() => setSimulationStatus('idle')}
              className="h-full"
            />
          )}

          {activePanel === 'forecast' && (
            <ExecutionForecastPanel
              forecasts={mockForecasts}
              horizon={60}
              className="h-full"
            />
          )}

          {activePanel === 'failure' && (
            <FailureProbabilityMap
              nodes={mockFailureNodes}
              threshold={0.3}
              className="h-full"
            />
          )}

          {activePanel === 'latency' && (
            <LatencyPredictionGraph
              data={mockLatencyData}
              showActual
              showConfidence
              className="h-full"
            />
          )}

          {activePanel === 'hallucination' && (
            <HallucinationRiskAnalyzer
              risks={mockHallucinationRisks}
              threshold={0.5}
              className="h-full"
            />
          )}

          {activePanel === 'cost' && (
            <CostSimulationChart
              projections={mockCostProjections}
              showBreakdown
              className="h-full"
            />
          )}

          {activePanel === 'stress' && (
            <StressTestRunner
              results={mockStressTestResults}
              isRunning={false}
              className="h-full"
            />
          )}

          {activePanel === 'scaling' && (
            <ScalingForecastMap
              projections={mockScalingProjections}
              className="h-full"
            />
          )}

          {activePanel === 'behavior' && (
            <AgentBehaviorPredictor
              predictions={mockAgentPredictions}
              timeHorizon={60}
              className="h-full"
            />
          )}

          {activePanel === 'workflow' && (
            <WorkflowOutcomeSimulator
              workflowId="checkout-workflow"
              paths={mockWorkflowPaths}
              className="h-full"
            />
          )}

          {activePanel === 'collision' && (
            <ResourceCollisionDetector
              collisions={mockResourceCollisions}
              threshold={0.5}
              className="h-full"
            />
          )}
        </div>
      </div>
    </div>
  )
}

export default SimulationIntegration
