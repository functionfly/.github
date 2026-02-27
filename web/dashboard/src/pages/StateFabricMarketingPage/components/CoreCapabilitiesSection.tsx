import { RotateCcw, Shield, BarChart3, Database } from "lucide-react";
import { SectionWrapper } from "./SectionWrapper";
import { CapabilitySection } from "./CapabilitySection";
import { SideBySideLayout } from "./SideBySideLayout";
import { MetricBadge } from "./MetricBadge";
import { AnimatedCounter } from "./AnimatedCounter";
import { FadeInOnScroll } from "./FadeInOnScroll";

const capabilities = [
  {
    icon: <RotateCcw className="w-6 h-6 text-blue-500" />,
    title: "Deterministic Replay",
    description: "Re-run any execution path exactly as it occurred.",
    metrics: [
      { value: "100%", label: "Accuracy", variant: "success" as const },
      { value: "0.1ms", label: "Replay Latency", variant: "info" as const }
    ],
    stats: {
      executions: 1000000,
      success: 99.9
    }
  },
  {
    icon: <Shield className="w-6 h-6 text-green-500" />,
    title: "Agent Memory Durability",
    description: "Short-term and long-term memory stored safely.",
    metrics: [
      { value: "99.999%", label: "Durability", variant: "success" as const },
      { value: "∞", label: "Retention", variant: "info" as const }
    ],
    stats: {
      sessions: 5000000,
      uptime: 99.9
    }
  },
  {
    icon: <BarChart3 className="w-6 h-6 text-purple-500" />,
    title: "Execution Attribution",
    description: "Track every tool call with cost and state changes.",
    metrics: [
      { value: "100%", label: "Coverage", variant: "success" as const },
      { value: "$0.001", label: "Per Call", variant: "warning" as const }
    ],
    stats: {
      calls: 50000000,
      cost: 0.001
    }
  },
  {
    icon: <Database className="w-6 h-6 text-orange-500" />,
    title: "Cost-Controlled Storage",
    description: "Object storage backed. Predictable cost at scale.",
    metrics: [
      { value: "$0.02", label: "Per GB/Month", variant: "warning" as const },
      { value: "∞", label: "Scale", variant: "info" as const }
    ],
    stats: {
      storage: 1000000,
      cost: 0.02
    }
  }
];

export function CoreCapabilitiesSection() {
  return (
    <SectionWrapper className="core-capabilities-section bg-bg-secondary/50">
      <div className="max-w-6xl mx-auto">
        <FadeInOnScroll>
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-6">
              Core Capabilities
            </h2>
            <p className="text-xl text-slate-600 dark:text-text-secondary max-w-3xl mx-auto">
              Four essential capabilities that make State Fabric the foundation for reliable AI agent systems.
            </p>
          </div>
        </FadeInOnScroll>

        <div className="space-y-8">
          {capabilities.map((capability, index) => (
            <CapabilitySection
              key={index}
              title={capability.title}
              description={capability.description}
              icon={capability.icon}
              index={index}
            >
              <div className="flex flex-wrap gap-4 mb-6">
                {capability.metrics.map((metric, metricIndex) => (
                  <MetricBadge
                    key={metricIndex}
                    value={metric.value}
                    label={metric.label}
                    variant={metric.variant}
                  />
                ))}
              </div>

              {index === 0 && (
                <SideBySideLayout
                  left={
                    <div className="space-y-4">
                      <div className="text-sm text-slate-500 dark:text-text-secondary uppercase tracking-wide font-medium">
                        Executions Replayed
                      </div>
                      <div className="text-3xl font-bold text-slate-900 dark:text-white">
                        <AnimatedCounter value={capability.stats.executions} />+
                      </div>
                      <div className="text-sm text-slate-600 dark:text-text-secondary">
                        with <AnimatedCounter value={capability.stats.success} decimals={1} />% success rate
                      </div>
                    </div>
                  }
                  right={
                    <div className="bg-bg-secondary p-4 rounded-lg border border-border">
                      <div className="text-sm text-slate-500 dark:text-text-secondary mb-2 font-medium">
                        Sample Replay Sequence
                      </div>
                      <pre className="text-xs text-slate-700 dark:text-slate-300 bg-bg-primary p-3 rounded border overflow-x-auto">
{`// Replay execution path
const replay = await agent.replay({
  executionId: "exec_123",
  fromStep: 0,
  toStep: 10
});

// Results match exactly
assert.deepEqual(replay.result, originalResult);`}
                      </pre>
                    </div>
                  }
                />
              )}

              {index === 1 && (
                <SideBySideLayout
                  left={
                    <div className="space-y-4">
                      <div className="text-sm text-slate-500 dark:text-text-secondary uppercase tracking-wide font-medium">
                        Active Sessions
                      </div>
                      <div className="text-3xl font-bold text-slate-900 dark:text-white">
                        <AnimatedCounter value={capability.stats.sessions} />+
                      </div>
                      <div className="text-sm text-slate-600 dark:text-text-secondary">
                        with <AnimatedCounter value={capability.stats.uptime} decimals={1} />% uptime
                      </div>
                    </div>
                  }
                  right={
                    <div className="bg-bg-secondary p-4 rounded-lg border border-border">
                      <div className="text-sm text-slate-500 dark:text-text-secondary mb-2 font-medium">
                        Memory Persistence
                      </div>
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <span>Short-term:</span>
                          <span className="text-green-600 dark:text-green-400 font-medium">Active</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Long-term:</span>
                          <span className="text-green-600 dark:text-green-400 font-medium">Persisted</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Backup:</span>
                          <span className="text-green-600 dark:text-green-400 font-medium">Redundant</span>
                        </div>
                      </div>
                    </div>
                  }
                />
              )}

              {index === 2 && (
                <SideBySideLayout
                  left={
                    <div className="space-y-4">
                      <div className="text-sm text-slate-500 dark:text-text-secondary uppercase tracking-wide font-medium">
                        Tool Calls Tracked
                      </div>
                      <div className="text-3xl font-bold text-slate-900 dark:text-white">
                        <AnimatedCounter value={capability.stats.calls} />+
                      </div>
                      <div className="text-sm text-slate-600 dark:text-text-secondary">
                        at $<AnimatedCounter value={capability.stats.cost} decimals={3} /> per call
                      </div>
                    </div>
                  }
                  right={
                    <div className="bg-bg-secondary p-4 rounded-lg border border-border">
                      <div className="text-sm text-slate-500 dark:text-text-secondary mb-2 font-medium">
                        Attribution Example
                      </div>
                      <div className="space-y-1 text-sm">
                        <div>🔍 search(query) → $0.002</div>
                        <div>💾 save(data) → $0.001</div>
                        <div>🤖 generate(prompt) → $0.005</div>
                        <div className="text-slate-500 dark:text-text-secondary mt-2 text-xs">
                          Total: $0.008 | Agent: user_123
                        </div>
                      </div>
                    </div>
                  }
                />
              )}

              {index === 3 && (
                <SideBySideLayout
                  left={
                    <div className="space-y-4">
                      <div className="text-sm text-slate-500 dark:text-text-secondary uppercase tracking-wide font-medium">
                        Storage Capacity
                      </div>
                      <div className="text-3xl font-bold text-slate-900 dark:text-white">
                        <AnimatedCounter value={capability.stats.storage} /> GB
                      </div>
                      <div className="text-sm text-slate-600 dark:text-text-secondary">
                        at $<AnimatedCounter value={capability.stats.cost} decimals={2} /> per GB/month
                      </div>
                    </div>
                  }
                  right={
                    <div className="bg-bg-secondary p-4 rounded-lg border border-border">
                      <div className="text-sm text-slate-500 dark:text-text-secondary mb-2 font-medium">
                        Storage Tiers
                      </div>
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between items-center">
                          <span>Hot:</span>
                          <MetricBadge value="$0.05" label="/GB" variant="warning" />
                        </div>
                        <div className="flex justify-between items-center">
                          <span>Warm:</span>
                          <MetricBadge value="$0.02" label="/GB" variant="info" />
                        </div>
                        <div className="flex justify-between items-center">
                          <span>Cold:</span>
                          <MetricBadge value="$0.01" label="/GB" variant="default" />
                        </div>
                      </div>
                    </div>
                  }
                />
              )}
            </CapabilitySection>
          ))}
        </div>
      </div>
    </SectionWrapper>
  );
}