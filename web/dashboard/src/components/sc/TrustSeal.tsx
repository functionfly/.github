import React, { useEffect, useRef, useCallback } from 'react';
import { ReducedMotionGate } from './ReducedMotionGate';

interface TrustSealProps {
  label: string;
  variant?: 'live' | 'verified' | 'trust';
  size?: 'sm' | 'md' | 'lg';
  ariaLabel?: string;
}

const SIZE_CLASSES = {
  sm: { outer: 'w-6 h-6', inner: 'inset-[2px]', label: 'text-[10px]' },
  md: { outer: 'w-8 h-8', inner: 'inset-[3px]', label: 'text-xs' },
  lg: { outer: 'w-10 h-10', inner: 'inset-[4px]', label: 'text-sm' },
} as const;

// Idle rotation period in ms (15s = one full rotation)
const IDLE_ROTATION_PERIOD = 15000;

// Pointer tracking radius in px
const POINTER_TRACKING_RADIUS = 80;

// Animation constants
const SHIMMER_DURATION = 800;
const POINTER_SENSITIVITY = 0.4; // How much the pointer influences angle

/**
 * TrustSeal - Animated trust indicator with holographic foil effect.
 * 
 * Features:
 * - Conic gradient foil that rotates based on pointer position
 * - Idle rotation animation when not interacting
 * - Scroll-triggered shimmer effect
 * - Respects reduced motion preferences
 */
export const TrustSeal: React.FC<TrustSealProps> = ({
  label,
  variant = 'verified',
  size = 'md',
  ariaLabel,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const sealRef = useRef<HTMLDivElement>(null);
  const rafRef = useRef<number>(0);
  const animationStartRef = useRef<number>(0);
  const hasPlayedShimmerRef = useRef(false);
  const pointerAngleRef = useRef<number>(0);
  const isPointerNearRef = useRef(false);
  const shimmerStartRef = useRef<number>(0);
  const shimmerProgressRef = useRef<number>(0);

  // Easing function for shimmer (ease-out cubic)
  const easeOutCubic = (t: number): number => 1 - Math.pow(1 - t, 3);

  // Calculate angle from pointer position relative to seal center
  const calculatePointerAngle = useCallback((e: PointerEvent, rect: DOMRect): number => {
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    const dx = e.clientX - cx;
    const dy = e.clientY - cy;
    return Math.atan2(dy, dx) * (180 / Math.PI) + 90;
  }, []);

  // Check if pointer is within tracking radius of seal center
  const isPointerInRange = useCallback((e: PointerEvent, rect: DOMRect): boolean => {
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    const distance = Math.sqrt(
      Math.pow(e.clientX - cx, 2) + Math.pow(e.clientY - cy, 2)
    );
    return distance <= POINTER_TRACKING_RADIUS;
  }, []);

  // Main animation loop - updates CSS custom property for performance
  const animationLoop = useCallback((timestamp: number) => {
    if (!sealRef.current) return;

    // Initialize animation start time if needed
    if (animationStartRef.current === 0) {
      animationStartRef.current = timestamp;
    }

    // Calculate idle rotation angle (continuous, slow rotation)
    const idleElapsed = timestamp - animationStartRef.current;
    const idleAngle = (idleElapsed / IDLE_ROTATION_PERIOD) * 360;

    // Calculate pointer influence (if pointer is near)
    let finalAngle: number;
    if (isPointerNearRef.current) {
      // Blend between idle rotation and pointer angle
      finalAngle = idleAngle + (pointerAngleRef.current - idleAngle) * POINTER_SENSITIVITY;
    } else {
      finalAngle = idleAngle;
    }

    // Handle shimmer animation if it hasn't played yet
    if (!hasPlayedShimmerRef.current) {
      const shimmerElapsed = timestamp - shimmerStartRef.current;
      shimmerProgressRef.current = Math.min(shimmerElapsed / SHIMMER_DURATION, 1);
      const easedProgress = easeOutCubic(shimmerProgressRef.current);
      
      // Apply shimmer as a rotation offset that settles back
      const shimmerOffset = (1 - easedProgress) * 180;
      finalAngle += shimmerOffset;

      if (shimmerProgressRef.current >= 1) {
        hasPlayedShimmerRef.current = true;
      }
    }

    // Update CSS custom property - this is the key performance optimization
    // No React re-render, just style update on the DOM element
    sealRef.current.style.setProperty('--seal-angle', `${finalAngle}deg`);
    
    // Continue animation loop
    rafRef.current = requestAnimationFrame(animationLoop);
  }, []);

  // Pointer move handler - updates angle based on cursor position
  const handlePointerMove = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    if (!containerRef.current) return;

    const rect = containerRef.current.getBoundingClientRect();
    isPointerNearRef.current = isPointerInRange(e.nativeEvent, rect);
    
    if (isPointerNearRef.current) {
      pointerAngleRef.current = calculatePointerAngle(e.nativeEvent, rect);
    }
  }, [calculatePointerAngle, isPointerInRange]);

  // Pointer leave handler
  const handlePointerLeave = useCallback(() => {
    isPointerNearRef.current = false;
  }, []);

  // Intersection observer for scroll-in shimmer
  useEffect(() => {
    if (!containerRef.current) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !hasPlayedShimmerRef.current) {
          // Reset shimmer state
          shimmerStartRef.current = performance.now();
          shimmerProgressRef.current = 0;
          hasPlayedShimmerRef.current = true;
        }
      },
      { threshold: 0.5 }
    );

    observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, []);

  // Start/stop animation loop
  useEffect(() => {
    animationStartRef.current = 0;
    rafRef.current = requestAnimationFrame(animationLoop);

    return () => {
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = 0;
      }
    };
  }, [animationLoop]);

  const accessibleLabel = ariaLabel ?? `${variant}: ${label}`;
  const sizeClass = SIZE_CLASSES[size];

  return (
    <ReducedMotionGate
      fallback={
        <div role="img" aria-label={accessibleLabel} className="inline-flex items-center gap-2">
          <div
            className={`relative ${sizeClass.outer} rounded-full`}
            style={{
              background: `conic-gradient(from 0deg, var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-a))`,
              boxShadow: 'var(--foil-shadow)',
            }}
          />
          <span
            className={`font-mono font-semibold uppercase tracking-wider text-[var(--status-ok)] ${sizeClass.label}`}
          >
            {label}
          </span>
        </div>
      }
    >
      <div
        ref={containerRef}
        role="img"
        aria-label={accessibleLabel}
        className="inline-flex items-center gap-2 cursor-default"
        onPointerMove={handlePointerMove}
        onPointerLeave={handlePointerLeave}
      >
        {/* Foil seal with CSS custom property for angle */}
        <div
          ref={sealRef}
          className={`relative ${sizeClass.outer} rounded-full`}
          style={{
            background: `conic-gradient(from var(--seal-angle, 0deg), var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-a))`,
          }}
        >
          {/* Punch-out center revealing background */}
          <div
            className={`absolute rounded-full bg-[var(--chamber-bg)] ${sizeClass.inner}`}
          />
          
          {/* Shimmer overlay */}
          <div
            className="absolute inset-0 rounded-full opacity-60 mix-blend-overlay pointer-events-none"
            style={{
              background: 'conic-gradient(from 0deg, transparent 0deg, rgba(255, 255, 255, 0.5) 30deg, transparent 60deg)',
            }}
          />
        </div>
        
        {/* Label text */}
        <span
          className={`font-mono font-semibold uppercase tracking-wider text-[var(--status-ok)] ${sizeClass.label}`}
        >
          {label}
        </span>
      </div>
    </ReducedMotionGate>
  );
};

TrustSeal.displayName = 'TrustSeal';
