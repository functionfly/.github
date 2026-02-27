import { X } from "lucide-react";
import { IconCard } from "./IconCard";
import { FadeInOnScroll } from "./FadeInOnScroll";

const problems = [
  "Tool calls are stateless",
  "Memory is fragile and inconsistent",
  "Replays are impossible",
  "Debugging distributed agents is chaos",
  "Billing attribution is unreliable"
];

export function ProblemGrid() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12 problem-grid">
      {problems.map((problem, index) => (
        <FadeInOnScroll key={index} delay={index * 0.1}>
          <IconCard
            icon={<X className="w-6 h-6 text-red-500" />}
            title={problem}
          />
        </FadeInOnScroll>
      ))}
    </div>
  );
}