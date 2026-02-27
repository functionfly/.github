import { SectionWrapper } from "./SectionWrapper";
import { ProblemGrid } from "./ProblemGrid";
import { FadeInOnScroll } from "./FadeInOnScroll";

export function ProblemSection() {
  return (
    <SectionWrapper className="bg-bg-primary problem-section-enhanced">
      <div className="max-w-6xl mx-auto">
        <FadeInOnScroll>
          <h2 className="text-3xl lg:text-4xl font-bold text-center mb-12 problem-section-header">
            Modern AI Systems Break Because State Breaks.
          </h2>
        </FadeInOnScroll>

        <ProblemGrid />

        <FadeInOnScroll delay={0.3}>
          <div className="text-center">
            <p className="text-xl max-w-2xl mx-auto problem-section-subtitle">
              Autonomous systems need <span className="font-semibold problem-section-highlight">durable, structured, replayable state.</span>
            </p>
          </div>
        </FadeInOnScroll>
      </div>
    </SectionWrapper>
  );
}