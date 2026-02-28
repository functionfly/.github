import { SectionWrapper } from "./SectionWrapper";
import { FadeInOnScroll } from "./FadeInOnScroll";
import { PricingPreviewCard } from "./PricingPreviewCard";
import { CTASection } from "./CTASection";

const pricingPlans = [
  {
    title: "Agent Execution Plan",
    price: "Included",
    period: "",
    description: "StateFabric is included in every Agent Execution Plan at no additional cost.",
    features: [
      "Full StateFabric platform access",
      "Deterministic replay included",
      "Enterprise-grade reliability",
      "Built-in cost attribution",
      "Automatic scaling",
      "24/7 enterprise support"
    ],
    highlighted: true
  },
  {
    title: "Standalone StateFabric",
    price: "$49",
    period: "month",
    description: "Use StateFabric independently of agent execution.",
    features: [
      "Full platform access",
      "Deterministic replay",
      "Cost attribution",
      "Object storage offload",
      "Multi-agent coordination",
      "API access only"
    ],
    highlighted: false
  }
];

export function PricingCTASection() {
  return (
    <SectionWrapper className="pricing-cta-section bg-linear-to-br from-bg-primary via-bg-secondary/30 to-bg-primary">
      <div className="max-w-6xl mx-auto">
        <FadeInOnScroll>
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-6">
              Simple, Transparent Pricing
            </h2>
            <p className="text-xl text-slate-600 dark:text-text-secondary max-w-3xl mx-auto">
              StateFabric is included in every Agent Execution Plan, or available standalone
              for teams building their own agent infrastructure.
            </p>
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.2}>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-16">
            {pricingPlans.map((plan, index) => (
              <PricingPreviewCard
                key={index}
                title={plan.title}
                price={plan.price}
                period={plan.period}
                description={plan.description}
                features={plan.features}
                highlighted={plan.highlighted}
              />
            ))}
          </div>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.4}>
          <CTASection
            title="Ready to Build Reliable AI Agents?"
            subtitle="Get started with StateFabric today and see the difference deterministic state management makes."
            primaryButton={{
              text: "Start with StateFabric",
              href: "/signup"
            }}
            secondaryButton={{
              text: "View Pricing Details",
              href: "/pricing#state-fabric"
            }}
          />
        </FadeInOnScroll>
      </div>
    </SectionWrapper>
  );
}
