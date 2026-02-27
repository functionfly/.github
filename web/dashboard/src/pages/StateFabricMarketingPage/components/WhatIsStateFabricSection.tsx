import { SectionWrapper } from "./SectionWrapper";
import { FeatureGrid } from "./FeatureGrid";
import { FadeInOnScroll } from "./FadeInOnScroll";

export function WhatIsStateFabricSection() {
  return (
    <SectionWrapper className="what-is-state-fabric-section bg-bg-primary">
      <div className="max-w-6xl mx-auto">
        <FadeInOnScroll>
          <h2 className="text-3xl lg:text-4xl font-bold text-center text-slate-900 dark:text-white mb-6">
            A Durable State Layer Built for AI Agents
          </h2>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.2}>
          <p className="text-xl text-slate-600 dark:text-text-secondary text-center max-w-3xl mx-auto mb-12">
            StateFabric combines append-only event logs, periodic snapshots, and deterministic replay to create a reliable state foundation for AI-driven systems.
          </p>
        </FadeInOnScroll>

        <FeatureGrid />
      </div>
    </SectionWrapper>
  );
}