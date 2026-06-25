import { useRef, useEffect, useCallback } from 'react';
import { useScroll, useTransform, useMotionValue, useAnimationFrame, useReducedMotion } from 'framer-motion';

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
const IDLE_DRIFT_SPEED = 1;
const SCROLL_ROTATION_RANGE = 720;

export function TrustSeal({ size = 'default', label = 'Verified' }: TrustSealProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const sealRef = useRef<HTMLDivElement>(null);
  const pointerAngleRef = useRef<number>(0);
  const isPointerNearRef = useRef(false);
  const idleAngleRef = useRef(140);
  const prefersReducedMotion = useReducedMotion();

  const sizeValues = sizes[size] ?? sizes['default'];

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ['start end', 'end start'],
  });

  const scrollAngle = useTransform(scrollYProgress, [0, 1], [0, SCROLL_ROTATION_RANGE]);
  const idleAngle = useMotionValue(140);

  useAnimationFrame(() => {
    if (prefersReducedMotion || !sealRef.current) return;

    idleAngleRef.current += (1 / 60) * IDLE_DRIFT_SPEED;
    if (idleAngleRef.current >= 360) {
      idleAngleRef.current -= 360;
    }
    idleAngle.set(idleAngleRef.current);
  });

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

    const handlePointerMove = (e: PointerEvent) => {
      const rect = containerRef.current!.getBoundingClientRect();
      isPointerNearRef.current = isPointerInRange(e, rect);

      if (isPointerNearRef.current) {
        pointerAngleRef.current = calculatePointerAngle(e, rect);
      }
    };

    const handlePointerLeave = () => {
      isPointerNearRef.current = false;
    };

    const element = containerRef.current;
    element.addEventListener('pointermove', handlePointerMove);
    element.addEventListener('pointerleave', handlePointerLeave);

    return () => {
      element.removeEventListener('pointermove', handlePointerMove);
      element.removeEventListener('pointerleave', handlePointerLeave);
    };
  }, [prefersReducedMotion, calculatePointerAngle, isPointerInRange]);

  useEffect(() => {
    if (!sealRef.current) return;

    let currentScrollAngle = 0;
    let currentIdleAngle = 140;
    let combinedAngle = 140;

    const unsubscribeScroll = scrollAngle.on('change', (value) => {
      currentScrollAngle = value;
      updateAngle();
    });

    const unsubscribeIdle = idleAngle.on('change', (value) => {
      currentIdleAngle = value;
      updateAngle();
    });

    const updateAngle = () => {
      if (isPointerNearRef.current) {
        combinedAngle = currentScrollAngle + (pointerAngleRef.current - currentIdleAngle) * POINTER_SENSITIVITY;
      } else {
        combinedAngle = currentScrollAngle + currentIdleAngle;
      }
      sealRef.current!.style.setProperty('--seal-angle', `${combinedAngle}deg`);
    };

    return () => {
      unsubscribeScroll();
      unsubscribeIdle();
    };
  }, [scrollAngle, idleAngle]);

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
      <span className="font-mono text-xs font-medium uppercase tracking-wider" style={{ color: 'var(--status-ok)' }}>
        {label}
      </span>
    </div>
  );
}
