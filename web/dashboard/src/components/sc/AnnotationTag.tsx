import React from 'react';

interface AnnotationTagProps {
  children: React.ReactNode;
  position?: 'tl' | 'tr' | 'bl' | 'br';
  className?: string;
}

const POSITION_STYLES: Record<string, React.CSSProperties> = {
  tl: { top: 6, left: 6 },
  tr: { top: 6, right: 6 },
  bl: { bottom: 6, left: 6 },
  br: { bottom: 6, right: 6 },
};

/**
 * AnnotationTag - Small fixed-position monospace label.
 * Engineering dimension callout detail for chambers.
 */
export const AnnotationTag: React.FC<AnnotationTagProps> = ({
  children,
  position = 'tl',
  className = '',
}) => {
  return (
    <span
      className={`absolute font-mono text-[9px] font-medium uppercase tracking-widest text-[var(--text-faint)] pointer-events-none ${className}`}
      style={POSITION_STYLES[position]}
      aria-hidden="true"
    >
      {children}
    </span>
  );
};

AnnotationTag.displayName = 'AnnotationTag';
