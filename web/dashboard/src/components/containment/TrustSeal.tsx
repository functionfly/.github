import { useEffect, useRef, useState, useCallback } from 'react';
import { motion, useScroll, useTransform, useSpring } from 'framer-motion';

export type TrustSealSize = 'sm' | 'md' | 'lg';

export interface TrustSealProps {
  size?: TrustSealSize;
  className?: string;
  showLabel?: boolean;
  label?: string;
}

const SIZE_MAP = {
  sm: 16,
  md: 22,
  lg: 32,
};

const INSET_MAP = {
  sm: 2,
  md: 4,
  lg: 6,
};

export function TrustSeal({ size = 'md', className = '', showLabel, label }: TrustSealProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [isHovered, setIsHovered] = useState(false);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);
  const idleAngleRef = useRef(140);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReducedMotion(mediaQuery.matches);
    const handler = (e: MediaQueryListEvent) => setPrefersReducedMotion(e.matches);
    mediaQuery.addEventListener('change', handler);
    return () => mediaQuery.removeEventListener('change', handler);
  }, []);

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ['start end', 'end start'],
  });

  const rawRotation = useTransform(scrollYProgress, [0, 1], [0, 720]);
  const springRotation = useSpring(rawRotation, { stiffness: 50, damping: 20 });

  const [scrollAngle, setScrollAngle] = useState(140);

  useEffect(() => {
    return springRotation.on('change', (v) => {
      setScrollAngle(v);
    });
  }, [springRotation]);

  useEffect(() => {
    if (prefersReducedMotion) return;

    let lastTime = performance.now();
    let angle = 140;

    const animate = (time: number) => {
      const delta = time - lastTime;
      if (delta > 16) {
        angle += delta * 0.0625;
        if (angle > 720) angle -= 360;
        if (!isHovered) {
          setScrollAngle(angle);
        }
        lastTime = time;
      }
      rafRef.current = requestAnimationFrame(animate);
    };

    rafRef.current = requestAnimationFrame(animate);
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [prefersReducedMotion, isHovered]);

  const handlePointerMove = useCallback(
    (e: React.PointerEvent) => {
      if (prefersReducedMotion || !containerRef.current) return;

      const rect = containerRef.current.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;
      const angle = Math.atan2(e.clientY - centerY, e.clientX - centerX) * (180 / Math.PI) + 90;
      setScrollAngle(angle);
    },
    [prefersReducedMotion]
  );

  const diameter = SIZE_MAP[size];
  const inset = INSET_MAP[size];

  return (
    <div
      ref={containerRef}
      className={`trust-seal trust-seal--${size} ${className}`}
      onPointerEnter={() => setIsHovered(true)}
      onPointerLeave={() => setIsHovered(false)}
      onPointerMove={handlePointerMove}
    >
      <div
        className="trust-seal__outer"
        style={{
          width: diameter,
          height: diameter,
          background: `conic-gradient(from ${scrollAngle}deg, var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-d), var(--foil-a))`,
          boxShadow: 'var(--shadow-seal)',
        }}
      >
        <div
          className="trust-seal__inner"
          style={{
            inset: inset,
            borderRadius: '50%',
            background: 'var(--panel)',
          }}
        />
      </div>
      {showLabel && (
        <span className="trust-seal__label">
          {label || 'Verified'}
        </span>
      )}
    </div>
  );
}