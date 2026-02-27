import { ReactNode } from "react";
import { HoverGlowCard } from "./HoverGlowCard";
import { IconWithTitle } from "./IconWithTitle";

interface FeatureCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  className?: string;
}

export function FeatureCard({ icon, title, description, className = "" }: FeatureCardProps) {
  return (
    <HoverGlowCard className={`h-full ${className}`}>
      <IconWithTitle icon={icon} title={title} />
      <p className="text-slate-600 dark:text-text-secondary mt-4 text-center">{description}</p>
    </HoverGlowCard>
  );
}