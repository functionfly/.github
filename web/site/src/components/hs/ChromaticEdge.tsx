import React, { forwardRef, type ReactNode } from 'react';

export interface ChromaticEdgeProps extends React.HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
  className?: string;
}

/**
 * ChromaticEdge - thin CSS-only gradient border (red -> clear -> blue
 * hint) for elements that need to suggest dispersion without the cost
 * of a real shader. Use on cards/inputs that don't justify a full
 * <RefractiveObject>.
 */
export const ChromaticEdge = forwardRef<HTMLDivElement, ChromaticEdgeProps>(function ChromaticEdge(
  { children, className = '', ...rest },
  ref
) {
  return (
    <div ref={ref} className={`hs-chromatic ${className}`} {...rest}>
      {children}
    </div>
  );
});
