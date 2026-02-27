import { SectionWrapper } from "./SectionWrapper";
import { FadeInOnScroll } from "./FadeInOnScroll";
import { FAQAccordion } from "./FAQAccordion";

const faqItems = [
  {
    question: "How is this different from a database?",
    answer: "StateFabric is purpose-built for AI agent state management, not general data storage. It combines event logging, automatic snapshots, deterministic replay, and cost attribution specifically for autonomous systems. Traditional databases lack the agent-specific features like replay guarantees and built-in cost tracking."
  },
  {
    question: "Is this event sourcing?",
    answer: "StateFabric uses event sourcing principles but adds significant enhancements for AI agents: automatic snapshots for performance, deterministic replay guarantees, cost attribution per tool call, and object storage offload for scale. It's event sourcing designed specifically for the unique requirements of autonomous systems."
  },
  {
    question: "Can I export my state?",
    answer: "Yes, StateFabric provides full data export capabilities. You can export state snapshots, event logs, and execution histories in standard formats. All data remains under your control and can be migrated to other systems if needed."
  },
  {
    question: "Is it multi-region?",
    answer: "Yes, StateFabric is designed for global deployment with automatic cross-region replication and failover. Your agent state is automatically synchronized across regions for maximum availability and performance."
  },
  {
    question: "What happens if storage grows?",
    answer: "StateFabric automatically manages storage growth through intelligent compaction, archival to cost-effective object storage, and configurable retention policies. You only pay for what you use, with predictable costs at scale."
  },
  {
    question: "How is replay guaranteed?",
    answer: "Replay is guaranteed through deterministic execution and comprehensive event logging. Every state change is recorded with timestamps, and the system ensures identical replay results. This is crucial for debugging AI behavior and ensuring compliance."
  }
];

export function FAQSection() {
  return (
    <SectionWrapper className="faq-section bg-bg-secondary/50">
      <div className="max-w-4xl mx-auto">
        <FadeInOnScroll>
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-6">
              Frequently Asked Questions
            </h2>
            <p className="text-xl text-slate-600 dark:text-text-secondary max-w-2xl mx-auto">
              Common questions about StateFabric and how it works with AI agents.
            </p>
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.2}>
          <FAQAccordion items={faqItems} />
        </FadeInOnScroll>
      </div>
    </SectionWrapper>
  );
}