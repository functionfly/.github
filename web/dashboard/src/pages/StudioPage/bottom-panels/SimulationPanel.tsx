import { useStudioAgents } from "@/hooks/useStudio";
import {
    AgentBehaviorPredictor,
    CostSimulationChart,
    ExecutionForecastPanel,
    FailureProbabilityMap,
    HallucinationRiskAnalyzer,
    LatencyPredictionGraph,
    ScalingForecastMap,
    SimulationControlCenter,
    StressTestRunner,
} from "@functionfly/ui-simulation";
import {
  adaptSimulationConfig,
  adaptSimulationMetrics,
  adaptSimulationStatus,
  adaptForecastData,
  adaptFailureNodes,
  adaptLatencyData,
  adaptCostProjections,
  adaptScalingProjections,
  adaptBehaviorPredictions,
  createDefaultStressTestConfig,
  type UIHallucinationRisk,
} from "@/adapters";
import type { SimulationConfig as StudioSimulationConfig, SimulationResult as StudioSimulationResult } from "@/hooks/useStudio";

interface TokenUsage {
  totalTokens: number;
  promptTokens: number;
  completionTokens: number;
  costUsd: number;
}

interface SimulationPanelProps {
  config: StudioSimulationConfig;
  result?: StudioSimulationResult;
  isRunning: boolean;
  tokenUsage: TokenUsage;
  onConfigChange: (config: StudioSimulationConfig) => void;
  onToggle: () => void;
  onRefresh: () => void;
}

export function SimulationPanel({
  config,
  result,
  isRunning,
  tokenUsage,
  onConfigChange,
  onToggle,
  onRefresh,
}: SimulationPanelProps) {
  const { agents: rawAgents } = useStudioAgents();

  const agentNames = Object.fromEntries(rawAgents.map(a => [a.id, a.name]));

  const adaptedConfig = adaptSimulationConfig(config);
  const adaptedStatus = adaptSimulationStatus(result);
  const adaptedMetrics = adaptSimulationMetrics(result);

  const forecastData = result?.metrics ? adaptForecastData([
    {
      timestamp: Date.now() - 3600000,
      value: result.metrics.throughput || 100,
      lower: 90,
      upper: 110,
      confidence: 0.95,
    },
    {
      timestamp: Date.now() - 2400000,
      value: (result.metrics.throughput || 100) * 1.1,
      lower: 95,
      upper: 125,
      confidence: 0.93,
    },
    {
      timestamp: Date.now() - 1200000,
      value: (result.metrics.throughput || 100) * 1.05,
      lower: 95,
      upper: 115,
      confidence: 0.94,
    },
    {
      timestamp: Date.now(),
      value: result.metrics.throughput || 100,
      lower: 90,
      upper: 110,
      confidence: 0.92,
    },
  ]) : [];

  const failureNodes = adaptFailureNodes([
    {
      nodeId: "node-1",
      nodeName: "Parse Input",
      probability: 0.05,
      factors: [{ name: "Network", contribution: 0.03 }],
      trend: "stable",
    },
    {
      nodeId: "node-2",
      nodeName: "Enrich Context",
      probability: 0.12,
      factors: [{ name: "Database", contribution: 0.08 }],
      trend: "increasing",
    },
  ]);

  const latencyData = result?.metrics ? adaptLatencyData([
    {
      timestamp: Date.now() - 3600000,
      p50: result.metrics.p50LatencyMs || 42,
      p95: result.metrics.p95LatencyMs || 78,
      p99: result.metrics.p99LatencyMs || 120,
      samples: 1000,
    },
    {
      timestamp: Date.now() - 2400000,
      p50: (result.metrics.p50LatencyMs || 42) * 1.05,
      p95: (result.metrics.p95LatencyMs || 78) * 1.05,
      p99: (result.metrics.p99LatencyMs || 120) * 1.05,
      samples: 1200,
    },
    {
      timestamp: Date.now() - 1200000,
      p50: result.metrics.p50LatencyMs || 42,
      p95: result.metrics.p95LatencyMs || 78,
      p99: result.metrics.p99LatencyMs || 120,
      samples: 1100,
    },
    {
      timestamp: Date.now(),
      p50: result.metrics.p50LatencyMs || 42,
      p95: result.metrics.p95LatencyMs || 78,
      p99: result.metrics.p99LatencyMs || 120,
      samples: 1300,
    },
  ]) : [];

  const costProjections = adaptCostProjections({
    totalTokens: tokenUsage.totalTokens,
    inputTokens: tokenUsage.promptTokens,
    outputTokens: tokenUsage.completionTokens,
    computeCost: 0,
    apiCost: 0,
    totalCost: tokenUsage.costUsd,
    hourlyBreakdown: [],
  });

  const hallucinationRisks: UIHallucinationRisk[] = [
    {
      id: "risk-1",
      source: "Temperature",
      type: "reasoning",
      severity: "medium",
      confidence: 0.92,
      description: "Temperature set to 0.7 may cause inconsistent outputs",
      indicators: ["Temperature", "Token Limit"],
      mitigationSuggestion: "Lower temperature for deterministic output. Add validation layer.",
    },
  ];

  const scalingProjections = adaptScalingProjections([
    {
      component: "API Gateway",
      currentReplicas: 180,
      predictedReplicas: 250,
      timestamp: Date.now(),
      trigger: "requests",
    },
    {
      component: "Database Pool",
      currentReplicas: 85,
      predictedReplicas: 150,
      timestamp: Date.now(),
      trigger: "memory",
    },
  ]);

  const behaviorPredictions = adaptBehaviorPredictions(
    rawAgents.slice(0, 3).map((agent) => ({
      agentId: agent.id,
      likelyActions: [
        { action: "Process batch", probability: 0.85, expectedOutcome: "Batch processed" },
        { action: "Wait for input", probability: 0.1, expectedOutcome: "Idle" },
        { action: "Retry failed", probability: 0.05, expectedOutcome: "Retry success" },
      ],
      confidence: 0.92,
      basedOnSamples: 150,
    })),
    agentNames
  );

  const handleStressTestStart = () => {
    const newConfig: StudioSimulationConfig = {
      ...config,
      iterations: 100,
      stressLevel: "high",
    };
    onConfigChange(newConfig);
  };

  return (
    <div className="p-3 space-y-4 overflow-y-auto">
      <SimulationControlCenter
        config={adaptedConfig}
        status={adaptedStatus}
        metrics={adaptedMetrics}
        onStart={isRunning ? undefined : onToggle}
        onPause={isRunning ? onToggle : undefined}
        onStop={isRunning ? onToggle : undefined}
        onReset={onRefresh}
        onConfigChange={(newConfig) => {
          onConfigChange({
            ...config,
            id: newConfig.id,
            name: newConfig.name,
            iterations: newConfig.parallelism || config.iterations,
            duration: newConfig.duration,
          });
        }}
      />

      {result?.metrics && (
        <div className="grid grid-cols-2 gap-3">
          <ExecutionForecastPanel forecasts={forecastData} />
          <FailureProbabilityMap nodes={failureNodes} />
        </div>
      )}

      {result?.metrics && (
        <div className="grid grid-cols-2 gap-3">
          <LatencyPredictionGraph data={latencyData} />
          <CostSimulationChart projections={costProjections} />
        </div>
      )}

      <HallucinationRiskAnalyzer risks={hallucinationRisks} />

      <div className="grid grid-cols-3 gap-3">
        <StressTestRunner
          results={[]}
          isRunning={isRunning}
          onTestStart={handleStressTestStart}
          onTestStop={onToggle}
        />
        <ScalingForecastMap projections={scalingProjections} />
        <AgentBehaviorPredictor predictions={behaviorPredictions} />
      </div>
    </div>
  );
}
