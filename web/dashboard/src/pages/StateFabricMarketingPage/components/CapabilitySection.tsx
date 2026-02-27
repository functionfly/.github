import { ReactNode } from "react";
import { HoverGlowCard } from "./HoverGlowCard";
import { FadeInOnScroll } from "./FadeInOnScroll";

interface CapabilitySectionProps {
  title: string;
  description: string;
  icon: ReactNode;
  children?: ReactNode;
  className?: string;
  index?: number;
}

export function CapabilitySection({
  title,
  description,
  icon,
  children,
  className = "",
  index = 0
}: CapabilitySectionProps) {
  return (
    <FadeInOnScroll delay={index * 0.1} className={className}>
      <HoverGlowCard className="group">
        <div className="flex items-start gap-6">
          <div className="shrink-0">
            <div className="w-12 h-12 rounded-lg glass-light flex items-center justify-center mb-4 animate-float">
              {icon}
            </div>
          </div>

          <div className="flex-1 min-w-0">
            <h3 className="text-xl font-semibold text-slate-900 dark:text-white mb-3 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
              {title}
            </h3>
            <p className="text-slate-600 dark:text-text-secondary leading-relaxed mb-6">
              {description}
            </p>

            {children}
          </div>
        </div>
      </HoverGlowCard>
    </FadeInOnScroll>
  );
}