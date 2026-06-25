import React, { type CSSProperties, type ReactNode, forwardRef } from 'react';

export interface NeumorphicSurfaceProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
  children?: ReactNode;
  /** Inset (pressed-in) variant */
  inset?: boolean;
  /** Border radius size */
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
 * NeumorphicSurface - opaque, same-color-as-background surface with
 * dual box-shadows from the global light source. Use ONLY for small
 * interactive controls (buttons, toggles, inputs). Never for large
 * panels - this is exactly what made 2020-era neumorphism fail
 * accessibility (low contrast at scale).
 */
export const NeumorphicSurface = forwardRef<HTMLDivElement, NeumorphicSurfaceProps>(
  function NeumorphicSurface(
    { children, inset = false, radius = 'md', className = '', style, ...rest },
    ref
  ) {
    const classes = ['hs-neu', inset ? 'hs-neu--inset' : '', className].filter(Boolean).join(' ');

    return (
      <div
        ref={ref}
        className={classes}
        style={{ borderRadius: radiusMap[radius] ?? radius, ...style }}
        {...rest}
      >
        {children}
      </div>
    );
  }
);
