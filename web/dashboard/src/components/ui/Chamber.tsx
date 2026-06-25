import React from 'react';

type CornerPosition = 'tl' | 'tr' | 'bl' | 'br';

interface CornerBraceProps {
  position: CornerPosition;
  className?: string;
}

export function CornerBrace({ position, className = '' }: CornerBraceProps) {
  return <div className={`corner-brace corner-brace--${position} ${className}`} />;
}

interface ChamberProps {
  children: React.ReactNode;
  size?: 'sm' | 'md' | 'lg';
  ribs?: boolean;
  corners?: CornerPosition[];
  className?: string;
  annotation?: string;
  annotationClassName?: string;
  as?: 'div' | 'section' | 'article';
}

export function Chamber({
  children,
  size = 'md',
  ribs = false,
  corners = [],
  className = '',
  annotation,
  annotationClassName = '',
  as: Tag = 'div',
}: ChamberProps) {
  const sizeClasses = {
    sm: 'founders-chamber--small',
    md: 'founders-chamber--medium',
    lg: 'founders-chamber--large',
  };

  return (
    <Tag
      className={`founders-chamber ${sizeClasses[size]} ${ribs ? 'founders-chamber--ribbed' : ''} ${className}`}
      style={{ position: 'relative' }}
    >
      {corners.map((pos) => (
        <CornerBrace key={pos} position={pos} />
      ))}
      {annotation && <span className={`annotation-tag ${annotationClassName}`}>{annotation}</span>}
      {children}
    </Tag>
  );
}
