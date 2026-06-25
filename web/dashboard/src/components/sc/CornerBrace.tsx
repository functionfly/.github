import React from 'react';

type Position = 'tl' | 'tr' | 'bl' | 'br';

interface CornerBraceProps {
  position: Position;
  className?: string;
}

const STYLES: Record<Position, React.CSSProperties> = {
  tl: {
    top: -1,
    left: -1,
    borderTop: '2px solid var(--steel-light)',
    borderLeft: '2px solid var(--steel-light)',
  },
  tr: {
    top: -1,
    right: -1,
    borderTop: '2px solid var(--steel-light)',
    borderRight: '2px solid var(--steel-light)',
  },
  bl: {
    bottom: -1,
    left: -1,
    borderBottom: '2px solid var(--steel-light)',
    borderLeft: '2px solid var(--steel-light)',
  },
  br: {
    bottom: -1,
    right: -1,
    borderBottom: '2px solid var(--steel-light)',
    borderRight: '2px solid var(--steel-light)',
  },
};

/**
 * Decorative L-shaped corner bracket.
 * Used to frame panels and cards in the Sealed Containment design.
 */
export const CornerBrace: React.FC<CornerBraceProps> = ({
  position,
  className = '',
}) => {
  return (
    <div
      className={`absolute w-4 h-4 pointer-events-none ${className}`}
      style={STYLES[position]}
      aria-hidden="true"
    />
  );
};

CornerBrace.displayName = 'CornerBrace';
