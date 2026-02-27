import { ReactNode } from "react";
import { FadeInOnScroll } from "./FadeInOnScroll";

interface UseCaseGridProps {
  children: ReactNode;
  className?: string;
}

export function UseCaseGrid({ children, className = "" }: UseCaseGridProps) {
  return (
    <div className={`grid grid-cols-1 md:grid-cols-2 gap-6 ${className}`}>
      {children}
    </div>
  );
}