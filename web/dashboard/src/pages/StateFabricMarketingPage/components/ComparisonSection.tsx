import { SectionWrapper } from "./SectionWrapper";
import { FadeInOnScroll } from "./FadeInOnScroll";
import { ComparisonTable } from "./ComparisonTable";

const comparisonData = [
  {
    feature: "Event Logging",
    diy: "Custom implementation required",
    stateFabric: true
  },
  {
    feature: "Snapshots",
    diy: "Manual periodic snapshots",
    stateFabric: true
  },
  {
    feature: "Replay",
    diy: "Complex custom logic",
    stateFabric: true
  },
  {
    feature: "Cost Attribution",
    diy: "Manual tracking required",
    stateFabric: true
  },
  {
    feature: "Object Storage Offload",
    diy: "Custom integration needed",
    stateFabric: true
  },
  {
    feature: "Multi-Agent Support",
    diy: "Complex coordination logic",
    stateFabric: true
  },
  {
    feature: "Automatic Scaling",
    diy: false,
    stateFabric: true
  },
  {
    feature: "Built-in Monitoring",
    diy: false,
    stateFabric: true
  },
  {
    feature: "Enterprise Security",
    diy: "Custom implementation",
    stateFabric: true
  },
  {
    feature: "Deterministic Replay",
    diy: false,
    stateFabric: true
  }
];

export function ComparisonSection() {
  return (
    <SectionWrapper className="comparison-section bg-bg-secondary/50">
      <div className="max-w-6xl mx-auto">
        <FadeInOnScroll>
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-6">
              StateFabric vs. DIY Infrastructure
            </h2>
            <p className="text-xl text-slate-600 dark:text-text-secondary max-w-3xl mx-auto">
              See why building state management infrastructure from scratch is prohibitively complex
              compared to StateFabric's production-ready platform.
            </p>
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.2}>
          <ComparisonTable items={comparisonData} />
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.4}>
          <div className="mt-12 text-center">
            <div className="bg-blue-50 dark:bg-blue-900/10 border border-blue-200 dark:border-blue-800 rounded-lg p-6 max-w-2xl mx-auto">
              <h3 className="text-lg font-semibold text-blue-900 dark:text-blue-100 mb-2">
                Months of Development Time Saved
              </h3>
              <p className="text-blue-700 dark:text-blue-300 text-sm">
                What would take a team of engineers months to build and maintain is now available as a
                production-ready service with enterprise-grade reliability and performance.
              </p>
            </div>
          </div>
        </FadeInOnScroll>
      </div>
    </SectionWrapper>
  );
}