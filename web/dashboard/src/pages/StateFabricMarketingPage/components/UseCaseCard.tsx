import { ReactNode } from "react";
import { HoverGlowCard } from "./HoverGlowCard";

interface UseCaseCardProps {
  title: string;
  description: string;
  icon?: ReactNode;
  className?: string;
}

export function UseCaseCard({ title, description, icon, className = "" }: UseCaseCardProps) {
  return (
    <HoverGlowCard className={`h-full ${className}`}>
      <div className="flex items-start gap-4">
        {icon && (
          <div className="shrink-0 w-10 h-10 rounded-lg glass-light flex items-center justify-center">
            {icon}
          </div>
        )}
        <div className="flex-1 min-w-0">
          <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-2">
            {title}
          </h3>
          <p className="text-slate-600 dark:text-text-secondary text-sm leading-relaxed">
            {description}
          </p>
        </div>
      </div>
    </HoverGlowCard>
  );
}