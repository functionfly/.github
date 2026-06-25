import React, { useEffect, useState } from 'react';

interface ReducedMotionGateProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * Detects user preference for reduced motion and either renders children
 * (if motion is allowed) or the fallback (if user prefers reduced motion).
 */
export const ReducedMotionGate: React.FC<ReducedMotionGateProps> = ({
  children,
  fallback = null,
}) => {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReducedMotion(mediaQuery.matches);

    const handler = (e: MediaQueryListEvent) => setPrefersReducedMotion(e.matches);
    mediaQuery.addEventListener('change', handler);
    return () => mediaQuery.removeEventListener('change', handler);
  }, []);

  return <>{prefersReducedMotion ? fallback : children}</>;
};

ReducedMotionGate.displayName = 'ReducedMotionGate';
