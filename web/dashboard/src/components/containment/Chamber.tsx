import type { HTMLAttributes, ReactNode } from 'react';

export interface ChamberProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  ribs?: boolean;
  nested?: boolean;
}

export function Chamber({ children, ribs, nested, className = '', ...props }: ChamberProps) {
  const baseClasses = 'chamber';
  const ribsClass = ribs ? 'chamber--ribs' : '';
  const nestedClass = nested ? 'chamber--nested' : '';

  return (
    <div className={`${baseClasses} ${ribsClass} ${nestedClass} ${className}`} {...props}>
      {children}
    </div>
  );
}