import React, { type ReactNode, Suspense, useEffect, useState } from 'react';
import { StaticHolographicBackdrop } from './fallbacks/StaticHolographicBackdrop';

export type HolographicBackdropVariant = 'gradient-mesh' | 'particles' | 'grid';

export interface HolographicBackdropProps {
  /**
   * Spline scene URL. Either:
   * - A path string like "/scenes/hero.splinecode"
   * - A Spline scene object (use Spline's createSplineScene)
   *
   * When provided AND the splinetool runtime is loaded, renders Spline.
   * Otherwise falls back to a static gradient.
   */
  scene?: string | object;
  /** Visual variant for static fallback */
  variant?: HolographicBackdropVariant;
  /** Children to overlay on top of the backdrop */
  children?: ReactNode;
  className?: string;
  style?: React.CSSProperties;
  /** Intensity of the backdrop (affects opacity in fallback) */
  intensity?: 'subtle' | 'medium' | 'strong';
}

const intensityOpacity: Record<NonNullable<HolographicBackdropVariant> extends never ? never : 'subtle' | 'medium' | 'strong', number> = {
  subtle: 0.4,
  medium: 0.7,
  strong: 1,
};

/**
 * HolographicBackdrop - ambient WebGPU/WebGL background.
 *
 * Tries to load Spline (smaller bundle, designer-friendly, designer-art-directed)
 * for the ambient 3D scene. If Spline is unavailable or fails, falls back
 * to a pre-baked CSS gradient.
 *
 * Wrap usage in <ReducedMotionGate> for low-GPU / reduced-motion users.
 */
export function HolographicBackdrop({
  scene,
  children,
  className,
  style,
  intensity = 'medium',
}: HolographicBackdropProps): React.ReactElement {
  const [SplineComponent, setSplineComponent] = useState<React.ComponentType<{ scene: string | object; style?: React.CSSProperties }> | null>(null);
  const [loadFailed, setLoadFailed] = useState(false);

  useEffect(() => {
    if (!scene) return;
    // Lazy-load the splinetool runtime + react-spline wrapper
    let cancelled = false;
    (async () => {
      try {
        const mod = await import(/* @vite-ignore */ '@splinetool/react-spline');
        if (cancelled) return;
        const Cmp = (mod.default ?? mod) as React.ComponentType<{ scene: string | object; style?: React.CSSProperties }>;
        setSplineComponent(() => Cmp);
      } catch {
        if (!cancelled) setLoadFailed(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [scene]);

  if (!scene || loadFailed || !SplineComponent) {
    return (
      <div className={className} style={{ position: 'relative', ...style }}>
        <StaticHolographicBackdrop style={{ opacity: intensityOpacity[intensity] }} />
        {children}
      </div>
    );
  }

  return (
    <div className={className} style={{ position: 'relative', ...style }}>
      <Suspense
        fallback={
          <StaticHolographicBackdrop style={{ opacity: intensityOpacity[intensity] }} />
        }
      >
        <SplineComponent scene={scene} style={{ position: 'absolute', inset: 0, zIndex: 'var(--hs-z-backdrop)' }} />
      </Suspense>
      {children}
    </div>
  );
}
