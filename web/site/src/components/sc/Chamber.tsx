import React from 'react';

interface ChamberProps {
  children: React.ReactNode;
  ribs?: boolean;
  nested?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

export const Chamber: React.FC<ChamberProps> = ({
  children,
  ribs = false,
  nested = false,
  className = '',
  style,
}) => {
  const bgColor = nested ? 'var(--panel-raised)' : 'var(--panel)';
  const radialGradient = nested
    ? undefined
    : 'radial-gradient(140% 100% at 15% 0%, var(--glass-tint), transparent 55%)';

  return (
    <div
      className={`relative ${className}`}
      style={{
        background: radialGradient
          ? `${radialGradient}, ${bgColor}`
          : bgColor,
        boxShadow: nested ? undefined : 'var(--shadow-chamber)',
        borderRadius: 'var(--radius-lg)',
        border: `1px solid var(--panel-edge)`,
        padding: 'var(--space-7)',
      }}
    >
      <style>{`
        @media (max-width: 479px) {
          .chamber-padding {
            padding: var(--space-5) !important;
          }
        }
      `}</style>
      <div className="chamber-padding">
        {ribs && !nested && (
          <div
            className="absolute inset-0 pointer-events-none select-none"
            style={{
              opacity: 0.025,
              backgroundImage:
                'repeating-linear-gradient(90deg, transparent 0px, transparent 119px, rgba(255,255,255,0.025) 120px)',
              borderRadius: 'inherit',
            }}
            aria-hidden="true"
          />
        )}
        {children}
      </div>
    </div>
  );
};
