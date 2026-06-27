import React from 'react';

export interface ChamberProps {
  children: React.ReactNode;
  /** Show ribbed texture overlay */
  ribs?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

/**
 * Chamber - A container with containment glass styling.
 *
 * Features:
 * - Radial gradient background with chamber-bg
 * - Inset highlight and box-shadow for depth
 * - Optional ribbed texture overlay for mechanical feel
 */
export const Chamber: React.FC<ChamberProps> = ({
  children,
  ribs = false,
  className = '',
  style,
}) => {
  return (
    <div
      className={`relative rounded ${className}`}
      style={{
        background: `radial-gradient(ellipse at top left, rgba(110, 190, 170, 0.05) 0%, var(--chamber-bg)) 60%), var(--chamber-bg)`,
        boxShadow: 'var(--chamber-shadow)',
        borderRadius: 'var(--radius)',
        ...style,
      }}
    >
      {ribs && (
        <div
          className="absolute inset-0 pointer-events-none select-none"
          style={{
            opacity: 0.03,
            backgroundImage:
              'repeating-linear-gradient(90deg, #4a565f 0px, #4a565f 1px, transparent 1px, transparent 24px)',
          }}
          aria-hidden="true"
        />
      )}
      {children}
    </div>
  );
};

Chamber.displayName = 'Chamber';
