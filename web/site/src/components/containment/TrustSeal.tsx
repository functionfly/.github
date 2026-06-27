import { useRef, useEffect, useCallback } from 'react';
import { useScroll, useTransform, useReducedMotion } from 'framer-motion';
import { useSharedAnimationFrame } from '@/lib/sharedRaf';

interface TrustSealProps {
  size?: 'small' | 'default' | 'large';
  label?: string;
}

const sizes = {
  small: { outer: 16, inset: 3 },
  default: { outer: 22, inset: 4 },
  large: { outer: 32, inset: 6 },
};

const POINTER_TRACKING_RADIUS = 80;
const POINTER_SENSITIVITY = 0.4;
const IDLE_DRIFT_DEG_PER_SEC = 1;
const SCROLL_ROTATION_RANGE = 720;
const DEFAULT_ANGLE = 140;

export function TrustSeal({ size = 'default', label = 'Verified' }: TrustSealProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const sealRef = useRef<HTMLDivElement>(null);
  const pointerAngleRef = useRef<number>(0);
  const isPointerNearRef = useRef(false);
  const idleAngleRef = useRef<number>(DEFAULT_ANGLE);
  const prefersReducedMotion = useReducedMotion();

  const sizeValues = sizes[size] ?? sizes['default'];

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ['start end', 'end start'],
  });

  const scrollAngle = useTransform(scrollYProgress, [0, 1], [0, SCROLL_ROTATION_RANGE]);

  const calculatePointerAngle = useCallback((e: PointerEvent, rect: DOMRect): number => {
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    const dx = e.clientX - cx;
    const dy = e.clientY - cy;
    return Math.atan2(dy, dx) * (180 / Math.PI) + 90;
  }, []);

  const isPointerInRange = useCallback((e: PointerEvent, rect: DOMRect): boolean => {
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    const distance = Math.sqrt(Math.pow(e.clientX - cx, 2) + Math.pow(e.clientY - cy, 2));
    return distance <= POINTER_TRACKING_RADIUS;
  }, []);

  useEffect(() => {
    if (!containerRef.current || prefersReducedMotion) return;

    const element = containerRef.current;

    const handlePointerMove = (e: PointerEvent) => {
      const rect = element.getBoundingClientRect();
      const inRange = isPointerInRange(e, rect);
      isPointerNearRef.current = inRange;
      if (inRange) {
        pointerAngleRef.current = calculatePointerAngle(e, rect);
      }
    };

    const handlePointerLeave = () => {
      isPointerNearRef.current = false;
    };

    element.addEventListener('pointermove', handlePointerMove);
    element.addEventListener('pointerleave', handlePointerLeave);

    return () => {
      element.removeEventListener('pointermove', handlePointerMove);
      element.removeEventListener('pointerleave', handlePointerLeave);
    };
  }, [prefersReducedMotion, calculatePointerAngle, isPointerInRange]);

  useSharedAnimationFrame((deltaSeconds) => {
    if (prefersReducedMotion || !sealRef.current) return;

    if (!isPointerNearRef.current) {
      idleAngleRef.current = (idleAngleRef.current + IDLE_DRIFT_DEG_PER_SEC * deltaSeconds) % 360;
    }

    const scrollVal = scrollAngle.get();
    const idleVal = idleAngleRef.current;

    let combinedAngle: number;
    if (isPointerNearRef.current) {
      combinedAngle = scrollVal + (pointerAngleRef.current - idleVal) * POINTER_SENSITIVITY;
    } else {
      combinedAngle = scrollVal + idleVal;
    }

    sealRef.current.style.setProperty('--seal-angle', `${combinedAngle}deg`);
  });

  return (
    <div
      ref={containerRef}
      role="img"
      aria-label={label}
      className="inline-flex items-center gap-2"
    >
      <div
        ref={sealRef}
        className="relative rounded-full trust-seal-foil"
        style={{
          width: sizeValues.outer,
          height: sizeValues.outer,
          background: `conic-gradient(from var(--seal-angle, 140deg), var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-d), var(--foil-a))`,
        }}
      >
        <div
          className="absolute rounded-full"
          style={{
            inset: sizeValues.inset,
            backgroundColor: 'var(--panel)',
          }}
        />
      </div>
      <span
        className="font-mono text-xs font-medium uppercase tracking-wider"
        style={{ color: 'var(--status-ok)' }}
      >
        {label}
      </span>
    </div>
  );
}
