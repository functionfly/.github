import { useEffect, useRef, useState } from 'react';

export type CornerBracePosition = 'tl' | 'tr' | 'bl' | 'br';

export interface CornerBraceProps {
  position: CornerBracePosition;
  className?: string;
}

export function CornerBrace({ position, className = '' }: CornerBraceProps) {
  return (
    <div
      className={`corner-brace corner-brace--${position} ${className}`}
      aria-hidden="true"
    />
  );
}

interface PageGridProps {
  className?: string;
}

export function PageGrid({ className = '' }: PageGridProps) {
  return <div className={`page-grid ${className}`} aria-hidden="true" />;
}