import React from 'react';

type Position = 'tl' | 'tr' | 'bl' | 'br';

interface CornerBraceProps {
  position: Position;
  className?: string;
}

const STYLES: Record<Position, React.CSSProperties> = {
  tl: {
    top: '14px',
    left: '14px',
    width: '22px',
    height: '22px',
    borderTop: '2px solid var(--steel-light)',
    borderLeft: '2px solid var(--steel-light)',
  },
  tr: {
    top: '14px',
    right: '14px',
    width: '22px',
    height: '22px',
    borderTop: '2px solid var(--steel-light)',
    borderRight: '2px solid var(--steel-light)',
  },
  bl: {
    bottom: '14px',
    left: '14px',
    width: '22px',
    height: '22px',
    borderBottom: '2px solid var(--steel-light)',
    borderLeft: '2px solid var(--steel-light)',
  },
  br: {
    bottom: '14px',
    right: '14px',
    width: '22px',
    height: '22px',
    borderBottom: '2px solid var(--steel-light)',
    borderRight: '2px solid var(--steel-light)',
  },
};

export const CornerBrace: React.FC<CornerBraceProps> = ({
  position,
  className = '',
}) => {
  return (
    <div
      className={`absolute pointer-events-none ${className}`}
      style={STYLES[position]}
      aria-hidden="true"
    />
  );
};
