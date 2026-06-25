import { type CSSProperties, type ReactNode } from 'react';

export interface StaticHolographicBackdropProps {
  children?: ReactNode;
  className?: string;
  style?: CSSProperties;
  /** Intensity scales the gradient opacity (0-1). Default 1. */
  intensity?: 'subtle' | 'medium' | 'strong';
}

const intensityMap = {
  subtle: 0.4,
  medium: 0.7,
  strong: 1,
} as const;

/**
 * Static fallback for HolographicBackdrop. Renders a pre-baked
 * gradient + grain image. Zero GPU cost.
 */
export function StaticHolographicBackdrop({
  children,
  className = '',
  style,
  intensity = 'strong',
}: StaticHolographicBackdropProps): React.ReactElement {
  const opacity = intensityMap[intensity];
  return (
    <div
      className={className}
      style={{
        position: 'absolute',
        inset: 0,
        overflow: 'hidden',
        pointerEvents: 'none',
        zIndex: 'var(--hs-z-backdrop)',
        opacity,
        background:
          'radial-gradient(ellipse 60% 40% at 30% 30%, rgba(95, 208, 255, 0.12) 0%, transparent 50%),' +
          'radial-gradient(ellipse 50% 30% at 70% 60%, rgba(255, 157, 108, 0.08) 0%, transparent 40%),' +
          'radial-gradient(ellipse 40% 25% at 40% 80%, rgba(0, 255, 157, 0.05) 0%, transparent 35%)',
        ...style,
      }}
      aria-hidden="true"
    >
      {children}
    </div>
  );
}
