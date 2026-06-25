import React from 'react';

interface ChamberProps {
  children: React.ReactNode;
  variant?: 'default' | 'ribs';
  nested?: boolean;
  className?: string;
}

export function Chamber({
  children,
  variant = 'default',
  nested = false,
  className = '',
}: ChamberProps) {
  const classes = [
    'chamber',
    variant === 'ribs' ? 'chamber--ribs' : '',
    nested ? 'chamber--nested' : '',
    className,
  ].filter(Boolean).join(' ');

  return (
    <div className={classes}>
      {children}
    </div>
  );
}
