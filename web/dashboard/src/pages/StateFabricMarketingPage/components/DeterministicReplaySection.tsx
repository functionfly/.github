import { SectionWrapper } from "./SectionWrapper";
import { FadeInOnScroll } from "./FadeInOnScroll";
import { CodeBlock } from "./CodeBlock";
import { ReplayVisualizer } from "./ReplayVisualizer";
import { TimelineSlider } from "./TimelineSlider";
import { InteractiveDemoMock } from "./InteractiveDemoMock";
import { UseCaseCard } from "./UseCaseCard";
import { UseCaseGrid } from "./UseCaseGrid";
import { RotateCcw, Bug, FileCheck, TrendingUp } from "lucide-react";

const replaySteps = [
  {
    id: "1",
    timestamp: "2026-02-10T14:00:00Z",
    action: "Initialize Agent Memory",
    state: { memory: "empty", tools: [] },
    cost: 0.0001
  },
  {
    id: "2",
    timestamp: "2026-02-10T14:00:01Z",
    action: "Load Customer Data",
    state: { memory: "customer_data_loaded", tools: ["search"] },
    cost: 0.002
  },
  {
    id: "3",
    timestamp: "2026-02-10T14:00:02Z",
    action: "Analyze Sentiment",
    state: { memory: "sentiment_analyzed", tools: ["search", "analyze"], sentiment: "positive" },
    cost: 0.005
  },
  {
    id: "4",
    timestamp: "2026-02-10T14:00:03Z",
    action: "Generate Response",
    state: { memory: "response_generated", tools: ["search", "analyze", "generate"], response: "Thank you for your feedback!" },
    cost: 0.003
  }
];

const timelineEvents = [
  {
    id: "init",
    timestamp: "2026-02-10T14:00:00Z",
    title: "Agent Initialization",
    description: "Memory and tool setup",
    type: "action"
  },
  {
    id: "data-load",
    timestamp: "2026-02-10T14:00:01Z",
    title: "Data Loading",
    description: "Customer data ingestion",
    type: "action"
  },
  {
    id: "snapshot",
    timestamp: "2026-02-10T14:00:01.5Z",
    title: "Automatic Snapshot",
    description: "State checkpoint created",
    type: "snapshot"
  },
  {
    id: "analysis",
    timestamp: "2026-02-10T14:00:02Z",
    title: "Sentiment Analysis",
    description: "AI processing and inference",
    type: "action"
  },
  {
    id: "response",
    timestamp: "2026-02-10T14:00:03Z",
    title: "Response Generation",
    description: "Final output creation",
    type: "action"
  },
  {
    id: "cost-tracking",
    timestamp: "2026-02-10T14:00:03.1Z",
    title: "Cost Attribution",
    description: "Usage tracking and billing",
    type: "cost"
  }
];

const benefits = [
  {
    title: "Debug AI Hallucinations",
    description: "Step through agent decisions to identify when and why hallucinations occur.",
    icon: <Bug className="w-5 h-5 text-red-500" />
  },
  {
    title: "Audit Decisions",
    description: "Complete audit trail of all agent actions, inputs, and state changes.",
    icon: <FileCheck className="w-5 h-5 text-green-500" />
  },
  {
    title: "Validate Compliance",
    description: "Prove agent behavior meets regulatory and ethical requirements.",
    icon: <RotateCcw className="w-5 h-5 text-blue-500" />
  },
  {
    title: "Improve Agent Reliability",
    description: "Identify failure patterns and optimize agent performance over time.",
    icon: <TrendingUp className="w-5 h-5 text-purple-500" />
  }
];

export function DeterministicReplaySection() {
  return (
    <SectionWrapper className="deterministic-replay-section bg-bg-primary">
      <div className="max-w-6xl mx-auto">
        <FadeInOnScroll>
          <div className="text-center mb-16">
            <h2 className="text-4xl lg:text-5xl font-bold text-slate-900 dark:text-white mb-6">
              Rewind. Replay. Debug. Prove.
            </h2>
            <p className="text-xl text-slate-600 dark:text-text-secondary max-w-3xl mx-auto mb-8">
              StateFabric guarantees reproducible state by replaying events from the last snapshot forward.
              Every execution can be precisely reconstructed, making AI agents truly auditable and debuggable.
            </p>
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.2}>
          <div className="mb-16">
            <CodeBlock
              title="Replay API Example"
              code={`// Replay execution from specific timestamp
const replay = await state.replay("agent-123", {
  timestamp: "2026-02-10T14:00:00Z",
  includeCosts: true,
  stepByStep: true
});

// Results are identical every time
console.log("Replay completed:", replay.executionId);
console.log("Total cost:", replay.totalCost);
console.log("Duration:", replay.duration);`}
              language="javascript"
            />
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.4}>
          <div className="mb-16">
            <h3 className="text-2xl font-bold text-center text-slate-900 dark:text-white mb-8">
              Interactive Execution Replay
            </h3>
            <ReplayVisualizer steps={replaySteps} />
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.6}>
          <div className="mb-16">
            <h3 className="text-2xl font-bold text-center text-slate-900 dark:text-white mb-8">
              Execution Timeline
            </h3>
            <TimelineSlider events={timelineEvents} />
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.8}>
          <div className="mb-16">
            <h3 className="text-2xl font-bold text-center text-slate-900 dark:text-white mb-8">
              Live Demo
            </h3>
            <InteractiveDemoMock />
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={1.0}>
          <div>
            <h3 className="text-2xl font-bold text-center text-slate-900 dark:text-white mb-8">
              Why Deterministic Replay Matters
            </h3>
            <UseCaseGrid>
              {benefits.map((benefit, index) => (
                <UseCaseCard
                  key={index}
                  title={benefit.title}
                  description={benefit.description}
                  icon={benefit.icon}
                />
              ))}
            </UseCaseGrid>
          </div>
        </FadeInOnScroll>
      </div>
    </SectionWrapper>
  );
}