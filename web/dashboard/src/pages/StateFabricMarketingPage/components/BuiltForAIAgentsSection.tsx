import { SectionWrapper } from "./SectionWrapper";
import { FadeInOnScroll } from "./FadeInOnScroll";
import { UseCaseCard } from "./UseCaseCard";
import { UseCaseGrid } from "./UseCaseGrid";
import { IntegrationBadge } from "./IntegrationBadge";
import { TagPill } from "./TagPill";
import { Cpu, Shield, DollarSign, Network, Users, Settings } from "lucide-react";

const integrations = [
  {
    title: "Agent Execution Runtimes",
    description: "Direct integration with popular agent frameworks",
    icon: <Cpu className="w-5 h-5 text-blue-500" />
  },
  {
    title: "Tool Discovery Systems",
    description: "Automatic tool registration and discovery",
    icon: <Settings className="w-5 h-5 text-green-500" />
  },
  {
    title: "Quota Enforcement",
    description: "Built-in rate limiting and budget controls",
    icon: <Shield className="w-5 h-5 text-red-500" />
  },
  {
    title: "Memory Indexing",
    description: "Intelligent memory organization and retrieval",
    icon: <Network className="w-5 h-5 text-purple-500" />
  },
  {
    title: "Cost Attribution",
    description: "Per-agent cost tracking and billing",
    icon: <DollarSign className="w-5 h-5 text-yellow-500" />
  }
];

const benefits = [
  {
    title: "Handles Bursty Concurrency",
    description: "Automatically scales to handle sudden spikes in agent activity without performance degradation.",
    icon: <Network className="w-5 h-5 text-blue-500" />
  },
  {
    title: "Prevents Runaway State Growth",
    description: "Intelligent garbage collection and state compaction prevent unbounded memory growth.",
    icon: <Shield className="w-5 h-5 text-green-500" />
  },
  {
    title: "Enforces Per-Agent Budgets",
    description: "Set spending limits per agent with automatic throttling when limits are reached.",
    icon: <DollarSign className="w-5 h-5 text-red-500" />
  },
  {
    title: "Supports Multi-Agent Systems",
    description: "Coordination primitives for complex multi-agent workflows and agent-to-agent communication.",
    icon: <Users className="w-5 h-5 text-purple-500" />
  }
];

const agentTypes = [
  "Conversational Agents",
  "Task Automation",
  "Data Analysis",
  "Code Generation",
  "Decision Support",
  "Workflow Orchestration"
];

export function BuiltForAIAgentsSection() {
  return (
    <SectionWrapper className="built-for-ai-agents-section bg-bg-secondary/50">
      <div className="max-w-6xl mx-auto">
        <FadeInOnScroll>
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-6">
              Purpose-Built for Autonomous Systems
            </h2>
            <p className="text-xl text-slate-600 dark:text-text-secondary max-w-3xl mx-auto mb-8">
              StateFabric integrates directly with AI agent architectures, providing the state management layer
              that autonomous systems need to be reliable, auditable, and cost-effective.
            </p>
            <div className="flex flex-wrap justify-center gap-2">
              {agentTypes.map((type, index) => (
                <TagPill key={index} text={type} variant="primary" />
              ))}
            </div>
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.2}>
          <div className="mb-16">
            <h3 className="text-2xl font-bold text-center text-slate-900 dark:text-white mb-8">
              Direct Integration Points
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {integrations.map((integration, index) => (
                <FadeInOnScroll key={index} delay={index * 0.1}>
                  <IntegrationBadge
                    title={integration.title}
                    description={integration.description}
                    icon={integration.icon}
                  />
                </FadeInOnScroll>
              ))}
            </div>
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.4}>
          <div>
            <h3 className="text-2xl font-bold text-center text-slate-900 dark:text-white mb-8">
              Enterprise-Grade Agent Management
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