/**
 * ObservabilityIntegration
 * Main container that wires all 17 Live Observability components together
 */

import * as React from "react"
import {
  LiveTelemetryPanel,
  TokenUsageStream,
  CostHeatmap,
  LatencyGraph,
  ExecutionProfiler,
  InferenceTraceViewer,
  ErrorCascadeVisualizer,
  MemoryPressureMonitor,
  GPUUsageMonitor,
  BandwidthUsagePanel,
  DistributedTracingViewer,
  RequestReplayPanel,
  LiveLogConsole,
  RuntimeMetricsGrid,
  ObservabilityRadar,
  IncidentTimeline,
  AnomalyDetectionPanel,
  CostPredictionPanel,
  ResourceForecastGraph,
  type TelemetryMetric,
  type TokenUsage,
  type InferenceSpan,
  type ErrorNode,
  type TraceSpan,
  type ReplayRequest,
  type LogEntry,
  type Anomaly,
  type Incident,
  type ResourceForecastPoint,
} from "@functionfly/ui-observability"
import { useObservabilityStore } from "@/stores/observabilityStore"
import { cn } from "@functionfly/ui-core"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@functionfly/ui-core"

// ============================================================================
// Types
// ============================================================================

interface ObservabilityIntegrationProps {
  className?: string
}

// ============================================================================
// Sample Data Helpers
// ============================================================================

const generateSampleLatencyData = (points: number = 50) =>
  Array.from({ length: points }, (_, i) => ({
    timestamp: Date.now() - (points - i) * 5 * 60 * 1000,
    p50: 20 + Math.random() * 30,
    p95: 50 + Math.random() * 50,
    p99: 80 + Math.random() * 100,
  }))

const sampleCostHeatmapData = (() => {
  const days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
  const data: Array<{ hour: number; day: string; cost: number; requests: number }> = []
  days.forEach((day) => {
    for (let hour = 0; hour < 24; hour++) {
      data.push({
        hour,
        day,
        cost: Math.random() * 5 + 0.1,
        requests: Math.floor(Math.random() * 1000) + 100,
      })
    }
  })
  return data
})()

const sampleTokenUsage: TokenUsage = {
  inputTokens: 1_250_000,
  outputTokens: 850_000,
  totalTokens: 2_100_000,
  costUSD: 42.35,
  models: {
    "gpt-4o": { calls: 1250, tokens: 1_500_000, cost: 28.50 },
    "claude-3": { calls: 890, tokens: 600_000, cost: 13.85 },
  },
  timeRange: { start: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), end: new Date().toISOString() },
}

const sampleSpans: InferenceSpan[] = [
  {
    id: "span-1",
    name: "llm.inference",
    startTime: Date.now() - 2000,
    endTime: Date.now() - 500,
    duration: 1500,
    status: "completed",
    model: "gpt-4o",
    inputTokens: 500,
    outputTokens: 300,
    children: [
      {
        id: "span-1-1",
        parentId: "span-1",
        name: "tokenize",
        startTime: Date.now() - 2000,
        endTime: Date.now() - 1900,
        duration: 100,
        status: "completed",
      },
      {
        id: "span-1-2",
        parentId: "span-1",
        name: "model.forward",
        startTime: Date.now() - 1900,
        endTime: Date.now() - 600,
        duration: 1300,
        status: "completed",
        children: [
          {
            id: "span-1-2-1",
            parentId: "span-1-2",
            name: "attention",
            startTime: Date.now() - 1800,
            endTime: Date.now() - 1000,
            duration: 800,
            status: "completed",
          },
        ],
      },
      {
        id: "span-1-3",
        parentId: "span-1",
        name: "decode",
        startTime: Date.now() - 600,
        endTime: Date.now() - 500,
        duration: 100,
        status: "completed",
      },
    ],
  },
  {
    id: "span-2",
    name: "postprocess",
    startTime: Date.now() - 500,
    duration: 500,
    status: "running",
  },
]

const sampleErrorCascade: ErrorNode = {
  id: "root",
  name: "ProcessOrder",
  type: "function",
  error: "Database connection timeout after 30s",
  children: [
    {
      id: "db-1",
      name: "PostgreSQL Pool",
      type: "runtime",
      error: "Connection refused",
      children: [
        {
          id: "db-1-1",
          name: "pg.driver.connect",
          type: "external",
          error: "ETIMEDOUT 10.0.0.5:5432",
        },
      ],
    },
    {
      id: "cache-1",
      name: "Redis Cache",
      type: "runtime",
      children: [
        {
          id: "cache-1-1",
          name: "Cache fallback",
          type: "function",
        },
      ],
    },
  ],
}

const sampleMemoryData = [
  { runtime: "nodejs-18", usedMB: 256, limitMB: 512, pressure: "medium" as const },
  { runtime: "python-3.11", usedMB: 180, limitMB: 256, pressure: "low" as const },
  { runtime: "go-1.21", usedMB: 45, limitMB: 256, pressure: "low" as const },
  { runtime: "rust-1.70", usedMB: 12, limitMB: 128, pressure: "low" as const },
]

const sampleGpuData = [
  {
    gpuId: "gpu-0",
    name: "NVIDIA A100",
    utilizationPct: 78,
    memoryUsedMB: 60_000,
    memoryTotalMB: 80_000,
    temperatureC: 72,
    powerW: 250,
  },
  {
    gpuId: "gpu-1",
    name: "NVIDIA A100",
    utilizationPct: 45,
    memoryUsedMB: 40_000,
    memoryTotalMB: 80_000,
    temperatureC: 65,
    powerW: 180,
  },
]

const sampleBandwidthData = [
  { interface: "eth0", rxMB: 125.5, txMB: 45.2, rxPackets: 150_000, txPackets: 120_000 },
  { interface: "eth1", rxMB: 0.5, txMB: 0.2, rxPackets: 500, txPackets: 450 },
]

const sampleDistributedSpans: TraceSpan[] = [
  { id: "d1", traceId: "abc123", serviceName: "api-gateway", operationName: "POST /orders", startTime: Date.now() - 500, duration: 120, status: "ok" },
  { id: "d2", traceId: "abc123", parentId: "d1", serviceName: "order-service", operationName: "CreateOrder", startTime: Date.now() - 450, duration: 100, status: "ok" },
  { id: "d3", traceId: "abc123", parentId: "d2", serviceName: "payment-service", operationName: "ProcessPayment", startTime: Date.now() - 400, duration: 80, status: "ok" },
  { id: "d4", traceId: "abc123", parentId: "d3", serviceName: "stripe", operationName: "stripe.charges.create", startTime: Date.now() - 380, duration: 60, status: "ok" },
  { id: "d5", traceId: "abc123", parentId: "d2", serviceName: "inventory-service", operationName: "ReserveStock", startTime: Date.now() - 350, duration: 70, status: "error", tags: { error: "insufficient_stock" } },
]

const sampleReplayRequests: ReplayRequest[] = [
  { id: "r1", timestamp: "2024-05-17T14:30:00Z", method: "POST", path: "/api/v1/functions/execute", status: 200, duration: 45, requestSize: 1024, responseSize: 2048 },
  { id: "r2", timestamp: "2024-05-17T14:29:00Z", method: "GET", path: "/api/v1/functions", status: 200, duration: 12, requestSize: 0, responseSize: 4096 },
  { id: "r3", timestamp: "2024-05-17T14:28:00Z", method: "POST", path: "/api/v1/deploy", status: 500, duration: 1500, requestSize: 5120, responseSize: 256 },
  { id: "r4", timestamp: "2024-05-17T14:27:00Z", method: "DELETE", path: "/api/v1/functions/fn-123", status: 204, duration: 8, requestSize: 0, responseSize: 0 },
]

const sampleLogs: LogEntry[] = [
  { id: "l1", timestamp: Date.now() - 1000, level: "info", message: "Server started on port 8080", service: "api" },
  { id: "l2", timestamp: Date.now() - 900, level: "debug", message: "Health check passed", service: "monitor" },
  { id: "l3", timestamp: Date.now() - 800, level: "info", message: "New connection from 10.0.0.5", service: "api" },
  { id: "l4", timestamp: Date.now() - 700, level: "warn", message: "High memory usage detected: 85%", service: "monitor" },
  { id: "l5", timestamp: Date.now() - 600, level: "error", message: "Failed to process request: timeout", service: "worker" },
  { id: "l6", timestamp: Date.now() - 500, level: "info", message: "Retrying operation (attempt 2/3)", service: "worker" },
  { id: "l7", timestamp: Date.now() - 400, level: "info", message: "Operation succeeded after retry", service: "worker" },
  { id: "l8", timestamp: Date.now() - 300, level: "debug", message: "Cache hit for key: user_123", service: "cache" },
]

const sampleRuntimeMetrics = [
  { id: "rt-1", name: "Node.js 18", status: "healthy" as const, metrics: { requests: 12500, latencyP50: 25, latencyP99: 120, errorRate: 0.3, cpu: 45, memory: 62 } },
  { id: "rt-2", name: "Python 3.11", status: "healthy" as const, metrics: { requests: 8300, latencyP50: 45, latencyP99: 200, errorRate: 0.8, cpu: 38, memory: 55 } },
  { id: "rt-3", name: "Go 1.21", status: "degraded" as const, metrics: { requests: 5200, latencyP50: 18, latencyP99: 95, errorRate: 2.1, cpu: 72, memory: 48 } },
  { id: "rt-4", name: "Rust 1.70", status: "healthy" as const, metrics: { requests: 15000, latencyP50: 8, latencyP99: 35, errorRate: 0.1, cpu: 25, memory: 30 } },
]

const sampleRadarMetrics = [
  { label: "Requests", value: 85, max: 100 },
  { label: "Latency", value: 65, max: 100 },
  { label: "Errors", value: 20, max: 100 },
  { label: "CPU", value: 55, max: 100 },
  { label: "Memory", value: 70, max: 100 },
  { label: "Throughput", value: 80, max: 100 },
]

const sampleIncidents: Incident[] = [
  { id: "inc-1", severity: "critical" as const, title: "Database connection pool exhausted", description: "PostgreSQL connection pool has reached maximum capacity", startTime: Date.now() - 3600000, status: "active" as const, affectedServices: ["api", "worker"] },
  { id: "inc-2", severity: "high" as const, title: "High memory usage on order-service", startTime: Date.now() - 7200000, endTime: Date.now() - 3600000, status: "resolved" as const, affectedServices: ["order-service"] },
  { id: "inc-3", severity: "medium" as const, title: "Elevated error rate on Python runtime", startTime: Date.now() - 1800000, status: "active" as const, affectedServices: ["python-runtime"] },
]

const sampleAnomalies: Anomaly[] = [
  { id: "an-1", metric: "p99_latency", description: "p99 latency has increased by 150% in the last hour", detectedAt: Date.now() - 300000, severity: "high" as const, value: 450, expectedRange: [80, 180], confidence: 0.92 },
  { id: "an-2", metric: "error_rate", description: "Error rate exceeded threshold of 1%", detectedAt: Date.now() - 600000, severity: "critical" as const, value: 3.2, expectedRange: [0, 1], confidence: 0.98 },
  { id: "an-3", metric: "throughput", description: "Request throughput dropped by 40%", detectedAt: Date.now() - 900000, severity: "medium" as const, value: 60, expectedRange: [150, 250], confidence: 0.78 },
]

// ============================================================================
// Component
// ============================================================================

export function ObservabilityIntegration({ className }: ObservabilityIntegrationProps) {
  const {
    timeRange,
    setTimeRange,
    latencyData,
    costHeatmapData,
    tokenUsage,
    logs,
    memoryData,
    gpuData,
    bandwidthData,
    distributedTraceId,
    distributedSpans,
    replayRequests,
    runtimeMetrics,
    radarMetrics,
    incidents,
    anomalies,
    currentCost,
    predictedCost,
    costTrend,
    costBasis,
    costProjections,
    resourceForecastData,
    selectedSpanId,
    setSelectedSpanId,
    rootError,
    setRootError,
    selectedLogId,
    setSelectedLogId,
    selectedIncidentId,
    setSelectedIncidentId,
    selectedAnomalyId,
    setSelectedAnomalyId,
    visiblePanels,
    togglePanel,
  } = useObservabilityStore()

  const [activeTab, setActiveTab] = React.useState("overview")

  // Initialize with sample data
  React.useEffect(() => {
    if (latencyData.length === 0) {
      // Store will initialize with sample data in real usage
    }
  }, [])

  return (
    <div className={cn("space-y-6", className)}>
      {/* Header with time range selector */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-text-primary flex items-center gap-2">
          <Activity className="size-5 text-brand-400" />
          Live Observability
        </h3>
        <div className="flex items-center gap-2">
          {["1h", "6h", "24h", "7d", "30d"].map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range as typeof timeRange)}
              className={cn(
                "px-3 py-1.5 text-xs rounded-lg transition-colors",
                timeRange === range
                  ? "bg-brand-500/20 text-brand-400 border border-brand-500/30"
                  : "text-text-muted hover:text-text-primary hover:bg-bg-secondary"
              )}
            >
              {range}
            </button>
          ))}
        </div>
      </div>

      {/* Main tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="performance">Performance</TabsTrigger>
          <TabsTrigger value="logs">Logs & Traces</TabsTrigger>
          <TabsTrigger value="incidents">Incidents</TabsTrigger>
          <TabsTrigger value="costs">Costs</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Live Telemetry */}
            <LiveTelemetryPanel
              metrics={[]}
              tokenUsage={sampleTokenUsage}
              showTokenUsage={true}
              showAlerts={true}
              timeRange={timeRange}
              onTimeRangeChange={setTimeRange}
            />

            {/* Health Radar */}
            <ObservabilityRadar metrics={sampleRadarMetrics} />

            {/* Runtime Metrics Grid */}
            <RuntimeMetricsGrid
              runtimes={sampleRuntimeMetrics}
              onRuntimeClick={(id) => console.log("Clicked runtime:", id)}
            />

            {/* Memory Pressure */}
            <MemoryPressureMonitor data={sampleMemoryData} />

            {/* GPU Usage */}
            <GPUUsageMonitor data={sampleGpuData} />

            {/* Bandwidth */}
            <BandwidthUsagePanel data={sampleBandwidthData} />
          </div>
        </TabsContent>

        {/* Performance Tab */}
        <TabsContent value="performance">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Latency Graph */}
            <LatencyGraph
              data={generateSampleLatencyData()}
              timeRange={timeRange}
              height={200}
            />

            {/* Execution Profiler */}
            <ExecutionProfiler
              executions={[
                { id: "e1", name: "processOrder", duration: 145, tokens: 250, cost: 0.0042, timestamp: new Date().toISOString(), status: "success" as const, retries: 0 },
                { id: "e2", name: "sendNotification", duration: 32, tokens: 45, cost: 0.0008, timestamp: new Date(Date.now() - 60000).toISOString(), status: "success" as const, retries: 0 },
                { id: "e3", name: "processPayment", duration: 890, tokens: 120, cost: 0.0150, timestamp: new Date(Date.now() - 120000).toISOString(), status: "error" as const, retries: 2 },
                { id: "e4", name: "generateReport", duration: 2340, tokens: 890, cost: 0.0230, timestamp: new Date(Date.now() - 180000).toISOString(), status: "timeout" as const, retries: 1 },
              ]}
            />

            {/* Resource Forecast */}
            <ResourceForecastGraph
              metric="Request Volume"
              unit="req/min"
              data={Array.from({ length: 30 }, (_, i) => ({
                timestamp: Date.now() - (30 - i) * 60 * 60 * 1000,
                actual: i < 20 ? 100 + Math.random() * 50 : undefined,
                predicted: i >= 15 ? 110 + Math.random() * 50 : undefined,
                lower: i >= 15 ? 80 + Math.random() * 30 : undefined,
                upper: i >= 15 ? 140 + Math.random() * 50 : undefined,
              }))}
            />

            {/* Cost Prediction */}
            <CostPredictionPanel
              currentCost={1250.42}
              predictedCost={1580.75}
              trend="increasing"
              predictionBasis="Based on 30-day rolling average with seasonal adjustment"
              projections={[
                { period: "Day 1", cost: 52, confidence: 0.95 },
                { period: "Day 7", cost: 380, confidence: 0.88 },
                { period: "Day 14", cost: 780, confidence: 0.82 },
                { period: "Day 30", cost: 1580, confidence: 0.75 },
              ]}
            />
          </div>
        </TabsContent>

        {/* Logs & Traces Tab */}
        <TabsContent value="logs">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Live Log Console */}
            <LiveLogConsole
              logs={sampleLogs}
              selectedEntryId={selectedLogId}
              onEntryClick={(entry) => setSelectedLogId(entry.id)}
              autoScroll={true}
            />

            {/* Inference Trace */}
            <InferenceTraceViewer
              traceId="trace-abc123"
              spans={sampleSpans}
              selectedSpanId={selectedSpanId}
              onSpanClick={(span) => setSelectedSpanId(span.id)}
            />

            {/* Distributed Tracing */}
            <DistributedTracingViewer
              traceId="abc123"
              spans={sampleDistributedSpans}
              onSpanClick={(span) => console.log("Span clicked:", span)}
            />

            {/* Request Replay */}
            <RequestReplayPanel
              requests={sampleReplayRequests}
              onReplay={(request) => console.log("Replaying:", request)}
            />

            {/* Error Cascade */}
            <ErrorCascadeVisualizer
              rootError={sampleErrorCascade}
              onNodeClick={(node) => console.log("Error node clicked:", node)}
            />
          </div>
        </TabsContent>

        {/* Incidents Tab */}
        <TabsContent value="incidents">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Incident Timeline */}
            <IncidentTimeline
              incidents={sampleIncidents}
              selectedIncidentId={selectedIncidentId}
              onIncidentClick={(incident) => setSelectedIncidentId(incident.id)}
            />

            {/* Anomaly Detection */}
            <AnomalyDetectionPanel
              anomalies={sampleAnomalies}
              onAnomalyClick={(anomaly) => setSelectedAnomalyId(anomaly.id)}
            />
          </div>
        </TabsContent>

        {/* Costs Tab */}
        <TabsContent value="costs">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Token Usage */}
            <TokenUsageStream
              usage={sampleTokenUsage}
              onTimeRangeChange={setTimeRange}
            />

            {/* Cost Heatmap */}
            <CostHeatmap data={sampleCostHeatmapData} />

            {/* Cost Prediction */}
            <CostPredictionPanel
              currentCost={1250.42}
              predictedCost={1580.75}
              trend="increasing"
              predictionBasis="Based on 30-day rolling average with seasonal adjustment"
              projections={[
                { period: "Day 1", cost: 52, confidence: 0.95 },
                { period: "Day 7", cost: 380, confidence: 0.88 },
                { period: "Day 14", cost: 780, confidence: 0.82 },
                { period: "Day 30", cost: 1580, confidence: 0.75 },
              ]}
            />

            {/* Bandwidth costs */}
            <BandwidthUsagePanel data={sampleBandwidthData} />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}

// ============================================================================
// Export all observability components
// ============================================================================

export {
  LiveTelemetryPanel,
  TokenUsageStream,
  CostHeatmap,
  LatencyGraph,
  ExecutionProfiler,
  InferenceTraceViewer,
  ErrorCascadeVisualizer,
  MemoryPressureMonitor,
  GPUUsageMonitor,
  BandwidthUsagePanel,
  DistributedTracingViewer,
  RequestReplayPanel,
  LiveLogConsole,
  RuntimeMetricsGrid,
  ObservabilityRadar,
  IncidentTimeline,
  AnomalyDetectionPanel,
  CostPredictionPanel,
  ResourceForecastGraph,
} from "@functionfly/ui-observability"

export type {
  TelemetryMetric,
  TokenUsage,
  InferenceSpan,
  ErrorNode,
  TraceSpan,
  ReplayRequest,
  LogEntry,
  Anomaly,
  Incident,
  ResourceForecastPoint,
} from "@functionfly/ui-observability"
