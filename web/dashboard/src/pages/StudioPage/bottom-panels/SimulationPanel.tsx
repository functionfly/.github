import React from "react";
import {
  SimulationControlCenter,
  ExecutionForecastPanel,
  FailureProbabilityMap,
  LatencyPredictionGraph,
  CostSimulationChart,
  HallucinationRiskAnalyzer,
  StressTestRunner,
  ScalingForecastMap,
  AgentBehaviorPredictor,
} from "@functionfly/ui-simulation";
import { useStudioAgents } from "@/hooks/useStudio";
import { Play, Pause, RefreshCw } from "lucide-react";

interface SimulationConfig {
  name: string;
  iterations: number;
  duration: number;
  stressLevel?: "low" | "medium" | "high" | "extreme";
}

interface SimulationResult {
  id: string;
  status: "running" | "completed" | "failed";
  metrics?: {
    throughput?: number;
    p50LatencyMs?: number;
    p95LatencyMs?: number;
    p99LatencyMs?: number;
  };
}

interface TokenUsage {
  totalTokens: number;
  promptTokens: number;
  completionTokens: number;
  costUsd: number;
}

interface SimulationPanelProps {
  config: SimulationConfig;
  result?: SimulationResult;
  isRunning: boolean;
  tokenUsage: TokenUsage;
  onConfigChange: (config: SimulationConfig) => void;
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

  const agents = rawAgents.map((agent) => ({
    id: agent.id,
    name: agent.name,
    role: agent.agentId,
    status:
      agent.status === "active"
        ? "running"
        : agent.status === "pending"
          ? "idle"
          : agent.status === "terminating" || agent.status === "terminated"
            ? "stopped"
            : (agent.status as "running" | "idle" | "paused" | "stopped" | "error"),
    memoryUsage: 0,
    memoryLimit: 512 * 1024 * 1024,
    executionBudget: 10.0,
    executionBudgetUsed: 0,
    permissions: [],
    tools: [],
    runtime: "wasm",
    model: "gpt-4o",
    uptime: 0,
    tasksCompleted: 0,
    tasksFailed: 0,
    avgLatency: 0,
    lastHeartbeat: agent.lastActivity || new Date().toISOString(),
    createdAt: new Date().toISOString(),
    description: `Agent ${agent.name}`,
    tags: [],
  }));

  return (
    <div className="p-3 space-y-4 overflow-y-auto">
      <SimulationControlCenter
        config={config}
        result={result}
        onConfigChange={onConfigChange}
        onToggle={onToggle}
        isRunning={isRunning}
      />

      {result?.metrics && (
        <div className="grid grid-cols-2 gap-3">
          <ExecutionForecastPanel
            data={[
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
            ]}
          />
          <FailureProbabilityMap
            data={[
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
            ]}
          />
        </div>
      )}

      {result?.metrics && (
        <div className="grid grid-cols-2 gap-3">
          <LatencyPredictionGraph
            data={[
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
            ]}
          />
          <CostSimulationChart
            estimate={{
              totalTokens: tokenUsage.totalTokens,
              inputTokens: tokenUsage.promptTokens,
              outputTokens: tokenUsage.completionTokens,
              computeCost: 0,
              apiCost: 0,
              totalCost: tokenUsage.costUsd,
              hourlyBreakdown: [],
            }}
          />
        </div>
      )}

      <HallucinationRiskAnalyzer
        risks={[
          {
            score: 0.15,
            factors: [
              { name: "Temperature", weight: 0.3, description: "Temperature set to 0.7" },
              { name: "Token Limit", weight: 0.2, description: "Sufficient context" },
            ],
            model: "gpt-4o",
            confidence: 0.92,
            recommendations: [
              "Lower temperature for deterministic output",
              "Add validation layer",
            ],
          },
        ]}
      />

      <div className="grid grid-cols-3 gap-3">
        <StressTestRunner
          results={[]}
          isRunning={isRunning}
          onTestStart={() =>
            onConfigChange({ ...config, iterations: 100, stressLevel: "high" })
          }
          onTestStop={onToggle}
        />
        <ScalingForecastMap
          projections={[
            {
              component: "API Gateway",
              currentReplicas: 180,
              predictedReplicas: 250,
              timestamp: Date.now(),
              trigger: "requests" as const,
            },
            {
              component: "Database Pool",
              currentReplicas: 85,
              predictedReplicas: 150,
              timestamp: Date.now(),
              trigger: "memory" as const,
            },
          ]}
        />
        <AgentBehaviorPredictor
          predictions={agents.slice(0, 3).map((agent) => ({
            agentId: agent.id,
            likelyActions: [
              { action: "Process batch", probability: 0.85, expectedOutcome: "Batch processed" },
              { action: "Wait for input", probability: 0.1, expectedOutcome: "Idle" },
              { action: "Retry failed", probability: 0.05, expectedOutcome: "Retry success" },
            ],
            confidence: 0.92,
            basedOnSamples: 150,
          }))}
        />
      </div>
    </div>
  );
}