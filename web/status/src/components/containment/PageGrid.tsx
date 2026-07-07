import React from 'react';

interface PageGridProps {
  className?: string;
}

export const PageGrid: React.FC<PageGridProps> = ({ className = '' }) => {
  return (
    <div
      className={`fixed inset-0 pointer-events-none select-none ${className}`}
      style={{
        backgroundImage:
          'linear-gradient(rgba(255,255,255,0.025) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.025) 1px, transparent 1px)',
        backgroundSize: '48px 48px',
        zIndex: 'var(--z-base)',
      }}
      aria-hidden="true"
    />
  );
};
