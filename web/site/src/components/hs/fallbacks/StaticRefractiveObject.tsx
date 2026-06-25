import { type CSSProperties, type ReactNode } from 'react';

export interface StaticRefractiveObjectProps {
  children?: ReactNode;
  className?: string;
  style?: CSSProperties;
  /** Approximate color of the refractive object */
  color?: string;
}

/**
 * Static fallback for RefractiveObject. Renders a CSS gradient
 * that suggests refraction without any GPU work. Used on low-end
 * devices or when reduced motion is requested.
 */
export function StaticRefractiveObject({
  children,
  className = '',
  style,
  color = 'var(--hs-accent)',
}: StaticRefractiveObjectProps): React.ReactElement {
  return (
    <div
      className={className}
      style={{
        width: '240px',
        height: '240px',
        borderRadius: '28px',
        background: `linear-gradient(135deg, ${color}40, transparent 60%, ${color}20)`,
        border: `1px solid ${color}55`,
        boxShadow: `
          0 0 60px ${color}22,
          inset 0 1px 0 rgba(255, 255, 255, 0.1),
          inset 0 -1px 0 rgba(0, 0, 0, 0.2)
        `,
        position: 'relative',
        ...style,
      }}
      aria-hidden="true"
    >
      {children}
    </div>
  );
}
