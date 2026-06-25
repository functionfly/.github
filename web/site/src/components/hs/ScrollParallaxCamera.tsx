import React, { useEffect, useRef, useState, type ReactNode } from 'react';

export interface ScrollParallaxCameraProps {
  children: ReactNode;
  /** Max rotation in degrees (default 3) */
  maxRotation?: number;
  /** Max translation in pixels (default 10) */
  maxTranslation?: number;
  className?: string;
  style?: React.CSSProperties;
}

/**
 * ScrollParallaxCamera - subtle scroll-linked camera movement.
 *
 * The point: refractive objects visibly bend differently as the user
 * scrolls. Keep displacement small (a few degrees, a few pixels) - this
 * is atmosphere, not a scroll-jacking gimmick.
 *
 * Pure DOM/CSS implementation - no Three.js needed for the parallax
 * effect itself. Pair with <RefractiveObject> for maximum effect.
 */
export function ScrollParallaxCamera({
  children,
  maxRotation = 3,
  maxTranslation = 10,
  className,
  style,
}: ScrollParallaxCameraProps): React.ReactElement {
  const ref = useRef<HTMLDivElement>(null);
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    let rafId = 0;
    let lastScrollY = 0;

    const onScroll = () => {
      lastScrollY = window.scrollY;
      if (rafId) return;
      rafId = requestAnimationFrame(() => {
        rafId = 0;
        if (!ref.current) return;
        const rect = ref.current.getBoundingClientRect();
        const vh = window.innerHeight;
        // 0 when element enters viewport from bottom, 1 when it leaves from top
        const p = 1 - (rect.top + rect.height / 2) / (vh + rect.height);
        setProgress(Math.max(0, Math.min(1, p)));
      });
    };

    window.addEventListener('scroll', onScroll, { passive: true });
    onScroll();
    return () => {
      window.removeEventListener('scroll', onScroll);
      if (rafId) cancelAnimationFrame(rafId);
    };
  }, []);

  const rotX = (progress - 0.5) * 2 * maxRotation;
  const translateY = (progress - 0.5) * 2 * maxTranslation;

  return (
    <div
      ref={ref}
      className={className}
      style={{
        perspective: '1000px',
        perspectiveOrigin: '50% 50%',
        ...style,
      }}
    >
      <div
        style={{
          transform: `translateY(${translateY}px) rotateX(${-rotX}deg)`,
          transformStyle: 'preserve-3d',
          transition: 'transform 80ms linear',
          willChange: 'transform',
        }}
      >
        {children}
      </div>
    </div>
  );
}
