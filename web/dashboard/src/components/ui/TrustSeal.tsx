'use client';

import React, { useState, useRef, useCallback, useEffect } from 'react';
import { Check } from 'lucide-react';

interface TrustSealProps {
  size?: 'sm' | 'md' | 'lg';
  showIcon?: boolean;
  animate?: boolean;
  className?: string;
}

export function TrustSeal({
  size = 'md',
  showIcon = true,
  animate = true,
  className = '',
}: TrustSealProps) {
  const [angle, setAngle] = useState(0);
  const [isHovering, setIsHovering] = useState(false);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);
  const sealRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReducedMotion(mediaQuery.matches);

    const handler = (e: MediaQueryListEvent) => setPrefersReducedMotion(e.matches);
    mediaQuery.addEventListener('change', handler);
    return () => mediaQuery.removeEventListener('change', handler);
  }, []);

  const handlePointerMove = useCallback(
    (e: React.PointerEvent<HTMLDivElement>) => {
      if (!prefersReducedMotion && isHovering && sealRef.current) {
        const rect = sealRef.current.getBoundingClientRect();
        const centerX = rect.left + rect.width / 2;
        const centerY = rect.top + rect.height / 2;
        const deltaX = e.clientX - centerX;
        const deltaY = e.clientY - centerY;
        const newAngle = Math.atan2(deltaY, deltaX) * (180 / Math.PI) + 90;
        setAngle(newAngle);
      }
    },
    [isHovering, prefersReducedMotion]
  );

  const sizeClasses = {
    sm: 'trust-seal--sm w-8 h-8',
    md: 'trust-seal--md w-12 h-12',
    lg: 'trust-seal--lg w-16 h-16',
  };

  const iconSizes = {
    sm: 14,
    md: 18,
    lg: 24,
  };

  return (
    <div
      ref={sealRef}
      className={`trust-seal ${sizeClasses[size]} ${className}`}
      style={{
        background: `conic-gradient(from ${angle}deg, var(--seal-a), var(--seal-b), var(--seal-c), var(--seal-a))`,
      }}
      onPointerEnter={() => setIsHovering(true)}
      onPointerLeave={() => {
        setIsHovering(false);
        if (!prefersReducedMotion) setAngle(0);
      }}
      onPointerMove={handlePointerMove}
      role="img"
      aria-label="Verified founder status"
    >
      <div className="trust-seal__inner">
        {showIcon && (
          <Check
            size={iconSizes[size]}
            className="trust-seal__icon"
            strokeWidth={2.5}
          />
        )}
      </div>
    </div>
  );
}

/**
 * Reduced motion fallback - static seal without pointer tracking
 */
export function TrustSealStatic({ size = 'md', showIcon = true }: Omit<TrustSealProps, 'animate'>) {
  const sizeClasses = {
    sm: 'trust-seal--sm w-8 h-8',
    md: 'trust-seal--md w-12 h-12',
    lg: 'trust-seal--lg w-16 h-16',
  };

  const iconSizes = {
    sm: 14,
    md: 18,
    lg: 24,
  };

  return (
    <div
      className={`trust-seal ${sizeClasses[size]}`}
      style={{ background: 'conic-gradient(from 0deg, var(--seal-a), var(--seal-b), var(--seal-c), var(--seal-a))' }}
      role="img"
      aria-label="Verified founder status"
    >
      <div className="trust-seal__inner">
        {showIcon && (
          <Check
            size={iconSizes[size]}
            className="trust-seal__icon"
            strokeWidth={2.5}
          />
        )}
      </div>
    </div>
  );
}
