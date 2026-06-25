import React, { type CSSProperties, type ReactNode, forwardRef } from 'react';
import { useLightSource } from './LightSourceProvider';

export type GlassDepth = 0 | 1 | 2 | 3;

export interface GlassPanelProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
  children?: ReactNode;
  /** Stack depth. Higher = more blur. Cap at 3. */
  depth?: GlassDepth;
  /** Add subtle grain texture overlay (sells the "physical glass" feel). */
  noise?: boolean;
  /** Border radius size */
  radius?: 'sm' | 'md' | 'lg' | 'pill' | string;
  /** Inline style overrides */
  style?: CSSProperties;
  /** Forwarded className */
  className?: string;
}

const radiusMap: Record<string, string> = {
  sm: 'var(--hs-radius-sm)',
  md: 'var(--hs-radius-md)',
  lg: 'var(--hs-radius-lg)',
  pill: 'var(--hs-radius-pill)',
};

/**
 * GlassPanel - the base translucent container for hs/* system.
 * Compose higher-level surfaces from this. Never hand-roll
 * backdrop-filter styling in component code.
 */
export const GlassPanel = forwardRef<HTMLDivElement, GlassPanelProps>(function GlassPanel(
  {
    children,
    depth = 1,
    noise = false,
    radius = 'md',
    className = '',
    style,
    ...rest
  },
  ref
) {
  const light = useLightSource();
  const computedRadius = radiusMap[radius] ?? radius;

  const classes = [
    'hs-glass',
    `hs-glass--depth-${depth}`,
    noise ? 'hs-glass--noise' : '',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  const inlineStyle: CSSProperties = {
    borderRadius: computedRadius,
    // Pass the light direction to CSS so the highlight gradient can align
    ['--hs-highlight-direction' as string]: light.highlightDirection,
    ...style,
  };

  return (
    <div ref={ref} className={classes} style={inlineStyle} {...rest}>
      {children}
    </div>
  );
});
