import React, { type CSSProperties, type ReactNode, forwardRef } from 'react';

export interface StaticGlassPanelProps extends React.HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
  radius?: 'sm' | 'md' | 'lg' | 'pill' | string;
  className?: string;
  style?: CSSProperties;
}

const radiusMap: Record<string, string> = {
  sm: 'var(--hs-radius-sm)',
  md: 'var(--hs-radius-md)',
  lg: 'var(--hs-radius-lg)',
  pill: 'var(--hs-radius-pill)',
};

/**
 * Static fallback for GlassPanel. Same look, no expensive
 * backdrop-filter blur, no animations. Used by ReducedMotionGate
 * on low-GPU devices or when reduced motion is requested.
 */
export const StaticGlassPanel = forwardRef<HTMLDivElement, StaticGlassPanelProps>(
  function StaticGlassPanel(
    { children, radius = 'md', className = '', style, ...rest },
    ref
  ) {
    return (
      <div
        ref={ref}
        className={className}
        style={{
          background: 'var(--hs-glass-bg-deep)',
          border: '1px solid var(--hs-border)',
          borderRadius: radiusMap[radius] ?? radius,
          boxShadow: 'var(--hs-shadow-soft)',
          ...style,
        }}
        {...rest}
      >
        {children}
      </div>
    );
  }
);
