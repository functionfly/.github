import React, { useEffect, useRef, useCallback } from "react";
import { useReducedMotion } from "./ReducedMotionGate";

interface TrustSealProps {
  label: string;
  variant?: "live" | "verified" | "trust";
  size?: "sm" | "md" | "lg";
  ariaLabel?: string;
}

const SIZES = {
  sm: { outer: 16, inner: 3, label: "10px" },
  md: { outer: 22, inner: 4, label: "10px" },
  lg: { outer: 32, inner: 6, label: "12px" },
} as const;

const IDLE_ROTATION_PERIOD = 16000;
const POINTER_TRACKING_RADIUS = 80;
const SHIMMER_DURATION = 800;
const POINTER_SENSITIVITY = 0.4;

export const TrustSeal: React.FC<TrustSealProps> = ({
  label,
  variant = "verified",
  size = "md",
  ariaLabel,
}) => {
  const prefersReducedMotion = useReducedMotion();
  const containerRef = useRef<HTMLDivElement>(null);
  const sealRef = useRef<HTMLDivElement>(null);
  const rafRef = useRef<number>(0);
  const animationStartRef = useRef<number>(0);
  const hasPlayedShimmerRef = useRef(false);
  const pointerAngleRef = useRef<number>(0);
  const isPointerNearRef = useRef(false);
  const shimmerStartRef = useRef<number>(0);
  const shimmerProgressRef = useRef<number>(0);

  const easeOutCubic = (t: number): number => 1 - Math.pow(1 - t, 3);

  const calculatePointerAngle = useCallback(
    (e: PointerEvent, rect: DOMRect): number => {
      const cx = rect.left + rect.width / 2;
      const cy = rect.top + rect.height / 2;
      const dx = e.clientX - cx;
      const dy = e.clientY - cy;
      return (Math.atan2(dy, dx) * 180) / Math.PI + 90;
    },
    [],
  );

  const isPointerInRange = useCallback(
    (e: PointerEvent, rect: DOMRect): boolean => {
      const cx = rect.left + rect.width / 2;
      const cy = rect.top + rect.height / 2;
      const distance = Math.sqrt(
        Math.pow(e.clientX - cx, 2) + Math.pow(e.clientY - cy, 2),
      );
      return distance <= POINTER_TRACKING_RADIUS;
    },
    [],
  );

  const animationLoop = useCallback(
    (timestamp: number) => {
      if (!sealRef.current) return;

      if (animationStartRef.current === 0) {
        animationStartRef.current = timestamp;
      }

      const idleElapsed = timestamp - animationStartRef.current;
      const idleAngle = (idleElapsed / IDLE_ROTATION_PERIOD) * 360;

      let finalAngle: number;
      if (isPointerNearRef.current && !prefersReducedMotion) {
        finalAngle =
          idleAngle +
          (pointerAngleRef.current - idleAngle) * POINTER_SENSITIVITY;
      } else {
        finalAngle = idleAngle;
      }

      if (!hasPlayedShimmerRef.current && !prefersReducedMotion) {
        const shimmerElapsed = timestamp - shimmerStartRef.current;
        shimmerProgressRef.current = Math.min(
          shimmerElapsed / SHIMMER_DURATION,
          1,
        );
        const easedProgress = easeOutCubic(shimmerProgressRef.current);
        const shimmerOffset = (1 - easedProgress) * 180;
        finalAngle += shimmerOffset;

        if (shimmerProgressRef.current >= 1) {
          hasPlayedShimmerRef.current = true;
        }
      }

      sealRef.current.style.setProperty("--seal-angle", `${finalAngle}deg`);
      rafRef.current = requestAnimationFrame(animationLoop);
    },
    [prefersReducedMotion],
  );

  const handlePointerMove = useCallback(
    (e: React.PointerEvent<HTMLDivElement>) => {
      if (!containerRef.current || prefersReducedMotion) return;
      const rect = containerRef.current.getBoundingClientRect();
      isPointerNearRef.current = isPointerInRange(e.nativeEvent, rect);
      if (isPointerNearRef.current) {
        pointerAngleRef.current = calculatePointerAngle(
          e.nativeEvent,
          rect,
        );
      }
    },
    [prefersReducedMotion, calculatePointerAngle, isPointerInRange],
  );

  const handlePointerLeave = useCallback(() => {
    isPointerNearRef.current = false;
  }, []);

  useEffect(() => {
    if (!containerRef.current || prefersReducedMotion) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !hasPlayedShimmerRef.current) {
          shimmerStartRef.current = performance.now();
          shimmerProgressRef.current = 0;
          hasPlayedShimmerRef.current = true;
        }
      },
      { threshold: 0.5 },
    );
    observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, [prefersReducedMotion]);

  useEffect(() => {
    if (prefersReducedMotion) {
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = 0;
      }
      return;
    }
    animationStartRef.current = 0;
    rafRef.current = requestAnimationFrame(animationLoop);
    return () => {
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = 0;
      }
    };
  }, [prefersReducedMotion, animationLoop]);

  const accessibleLabel = ariaLabel ?? `${variant}: ${label}`;
  const s = SIZES[size];

  return (
    <div
      ref={containerRef}
      role="img"
      aria-label={accessibleLabel}
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: "0.5rem",
        cursor: "default",
      }}
      onPointerMove={handlePointerMove}
      onPointerLeave={handlePointerLeave}
    >
      <div
        ref={sealRef}
        className="trust-seal-foil"
        style={{
          position: "relative",
          width: s.outer,
          height: s.outer,
          borderRadius: "50%",
          background: `conic-gradient(from var(--seal-angle, 140deg), var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-a))`,
        }}
      >
        <div
          style={{
            position: "absolute",
            inset: s.inner,
            borderRadius: "50%",
            background: "var(--panel)",
          }}
        />
        {!prefersReducedMotion && (
          <div
            className="trust-seal-shimmer"
            style={{
              position: "absolute",
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              borderRadius: "50%",
            }}
            aria-hidden="true"
          />
        )}
      </div>
      <span
        className="trust-seal-label"
        style={{
          fontFamily: "var(--font-mono)",
          fontWeight: 500,
          textTransform: "uppercase",
          letterSpacing: "0.06em",
          color: "var(--status-ok)",
          fontSize: s.label,
        }}
      >
        {label}
      </span>
    </div>
  );
};
