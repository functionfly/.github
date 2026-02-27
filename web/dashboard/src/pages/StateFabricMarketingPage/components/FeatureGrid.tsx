import { Plus, Camera, RotateCcw, UserCheck } from "lucide-react";
import { FeatureCard } from "./FeatureCard";
import { FadeInOnScroll } from "./FadeInOnScroll";

const features = [
  {
    icon: <Plus className="w-6 h-6 text-blue-500" />,
    title: "Append-Only Events",
    description: "Immutable event logs ensure every state change is recorded and auditable."
  },
  {
    icon: <Camera className="w-6 h-6 text-green-500" />,
    title: "Snapshot Acceleration",
    description: "Periodic snapshots provide fast state recovery and efficient storage."
  },
  {
    icon: <RotateCcw className="w-6 h-6 text-purple-500" />,
    title: "Deterministic Replay",
    description: "Replay any execution path with perfect consistency and predictability."
  },
  {
    icon: <UserCheck className="w-6 h-6 text-orange-500" />,
    title: "Built-In Execution Attribution",
    description: "Track and attribute every action to its source for billing and debugging."
  }
];

export function FeatureGrid() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      {features.map((feature, index) => (
        <FadeInOnScroll key={index} delay={index * 0.1}>
          <FeatureCard
            icon={feature.icon}
            title={feature.title}
            description={feature.description}
          />
        </FadeInOnScroll>
      ))}
    </div>
  );
}