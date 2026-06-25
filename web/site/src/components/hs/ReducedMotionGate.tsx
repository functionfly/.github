import React, { type ReactNode, useEffect, useState } from 'react';
import {
  detectGPULevel,
  getEnhancedVisualsPreference,
  getReducedMotionPreference,
  type GPULevel,
} from './utils/gpuDetect';

export interface ReducedMotionGateProps {
  children: ReactNode;
  /** Static fallback when reduced motion is requested or GPU is low */
  fallback?: ReactNode;
  /** Override auto-detected GPU level */
  gpuLevel?: GPULevel;
  /** Override reduced-motion preference (e.g. user toggle) */
  prefersReducedMotion?: boolean;
  /** Minimum GPU level required to render `children`. Default: 'medium' */
  minLevel?: GPULevel;
}

const LEVEL_ORDER: Record<GPULevel, number> = { low: 0, medium: 1, high: 2 };

/**
 * ReducedMotionGate wraps any shader/Spline/Framer Motion content.
 * - Checks prefers-reduced-motion (OS or user override)
 * - Checks GPU level (heuristic: cores + screen size + device memory)
 * - Renders `fallback` (static) when either condition fails
 * - Renders `children` (full effects) when both pass
 *
 * Every 3D or shader component MUST be a child of this.
 */
export function ReducedMotionGate({
  children,
  fallback = null,
  gpuLevel,
  prefersReducedMotion,
  minLevel = 'medium',
}: ReducedMotionGateProps): React.ReactElement {
  // SSR default: assume medium GPU, enhanced visuals on, no reduced motion
  const [effectiveLevel, setEffectiveLevel] = useState<GPULevel>(gpuLevel ?? 'medium');
  const [effectivePrefersReduced, setEffectivePrefersReduced] = useState<boolean>(
    prefersReducedMotion ?? false
  );
  const [enhancedVisuals, setEnhancedVisuals] = useState<boolean>(true);

  useEffect(() => {
    // Only run detection on client
    setEffectiveLevel(gpuLevel ?? detectGPULevel());
    setEffectivePrefersReduced(prefersReducedMotion ?? getReducedMotionPreference());
    setEnhancedVisuals(getEnhancedVisualsPreference());

    // Listen for OS reduced-motion changes
    if (typeof window === 'undefined' || !window.matchMedia) return;
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handler = (e: MediaQueryListEvent) => {
      if (prefersReducedMotion === undefined) setEffectivePrefersReduced(e.matches);
    };
    mq.addEventListener('change', handler);

    // Listen for user toggle changes
    const onEnhancedChange = (e: CustomEvent<boolean>) => setEnhancedVisuals(e.detail);
    window.addEventListener('ff-enhanced-visuals-change', onEnhancedChange as EventListener);

    return () => {
      mq.removeEventListener('change', handler);
      window.removeEventListener('ff-enhanced-visuals-change', onEnhancedChange as EventListener);
    };
  }, [gpuLevel, prefersReducedMotion]);

  const shouldFallback =
    effectivePrefersReduced || !enhancedVisuals || LEVEL_ORDER[effectiveLevel] < LEVEL_ORDER[minLevel];

  return <>{shouldFallback ? fallback : children}</>;
}
