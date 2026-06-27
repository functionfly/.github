import React, { useEffect, useRef, useState } from "react";
import { useReducedMotion } from "./ReducedMotionGate";

interface TrustSealProps {
  label: string;
  variant?: "live" | "verified" | "trust";
  size?: "sm" | "md" | "lg";
  ariaLabel?: string;
}

const SIZES = {
  sm: { outer: 16, inner: 3 },
  md: { outer: 22, inner: 4 },
  lg: { outer: 32, inner: 6 },
} as const;

const IDLE_DRIFT_RATE = 1; // degrees per second
const POINTER_TRACKING_RADIUS = 80;
const POINTER_SENSITIVITY = 0.4;
const SCROLL_ROTATION_RANGE = 720; // two full revolutions

// Shared idle drift timer - one driver for all instances
let sharedDriftTime = 0;
let lastTimestamp = 0;
let driftRafId: number | null = null;
const driftListeners = new Set<(angle: number) => void>();

function startSharedDriftTimer() {
  if (driftRafId !== null) return;

  function tick(timestamp: number) {
    if (lastTimestamp === 0) lastTimestamp = timestamp;
    const delta = timestamp - lastTimestamp;
    lastTimestamp = timestamp;

    sharedDriftTime += (delta / 1000) * IDLE_DRIFT_RATE;

    // Notify all listeners with current drift angle
    driftListeners.forEach((listener) => {
      listener(sharedDriftTime % 360);
    });

    driftRafId = requestAnimationFrame(tick);
  }

  driftRafId = requestAnimationFrame(tick);
}

function stopSharedDriftTimer() {
  if (driftRafId !== null) {
    cancelAnimationFrame(driftRafId);
    driftRafId = null;
  }
}

function addDriftListener(listener: (angle: number) => void) {
  if (driftListeners.size === 0) {
    startSharedDriftTimer();
  }
  driftListeners.add(listener);
  return () => {
    driftListeners.delete(listener);
    if (driftListeners.size === 0) {
      stopSharedDriftTimer();
    }
  };
}

// Scroll-bound angle calculator
function useScrollAngle(
  ref: React.RefObject<HTMLElement | null>,
  prefersReducedMotion: boolean,
) {
  const [scrollAngle, setScrollAngle] = useState(140); // default starting angle
  const rafRef = useRef<number>(0);

  useEffect(() => {
    if (prefersReducedMotion) return;

    const element = ref.current;
    if (!element) return;

    const handleScroll = () => {
      if (rafRef.current) return;

      rafRef.current = requestAnimationFrame(() => {
        rafRef.current = 0;

        const rect = element.getBoundingClientRect();
        const viewportHeight = window.innerHeight;

        // Calculate scroll progress: 0 when element enters, 1 when it exits viewport
        let progress = 0;
        const elementTop = rect.top;
        const elementBottom = rect.bottom;

        if (elementTop <= viewportHeight && elementBottom >= 0) {
          const elementCenter = (elementTop + elementBottom) / 2;
          progress = Math.max(
            0,
            Math.min(1, (viewportHeight - elementCenter) / viewportHeight),
          );
        } else if (elementBottom < 0) {
          progress = 1;
        }

        // Map progress to angle: 0 → 720 degrees (2 full rotations)
        setScrollAngle(progress * SCROLL_ROTATION_RANGE);
      });
    };

    window.addEventListener("scroll", handleScroll, { passive: true });
    handleScroll();

    return () => {
      window.removeEventListener("scroll", handleScroll);
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
      }
    };
  }, [ref, prefersReducedMotion]);

  return scrollAngle;
}

// Pointer-reactive angle
function usePointerAngle(
  ref: React.RefObject<HTMLElement | null>,
  prefersReducedMotion: boolean,
) {
  const [pointerAngle, setPointerAngle] = useState(0);
  const [isPointerNear, setIsPointerNear] = useState(false);
  const rafRef = useRef<number>(0);

  useEffect(() => {
    if (prefersReducedMotion) return;

    const element = ref.current;
    if (!element) return;

    const handlePointerMove = (e: PointerEvent) => {
      if (rafRef.current) return;

      rafRef.current = requestAnimationFrame(() => {
        rafRef.current = 0;

        const rect = element.getBoundingClientRect();
        const cx = rect.left + rect.width / 2;
        const cy = rect.top + rect.height / 2;
        const distance = Math.sqrt(
          Math.pow(e.clientX - cx, 2) + Math.pow(e.clientY - cy, 2),
        );

        if (distance <= POINTER_TRACKING_RADIUS) {
          setIsPointerNear(true);
          const angle =
            (Math.atan2(e.clientY - cy, e.clientX - cx) * 180) / Math.PI + 90;
          setPointerAngle(angle);
        } else {
          setIsPointerNear(false);
        }
      });
    };

    window.addEventListener("pointermove", handlePointerMove, {
      passive: true,
    });

    return () => {
      window.removeEventListener("pointermove", handlePointerMove);
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
      }
    };
  }, [ref, prefersReducedMotion]);

  return { pointerAngle, isPointerNear };
}

export const TrustSeal: React.FC<TrustSealProps> = ({
  label,
  variant = "verified",
  size = "md",
  ariaLabel,
}) => {
  const prefersReducedMotion = useReducedMotion();
  const containerRef = useRef<HTMLDivElement>(null);
  const componentMounted = useRef(true);

  const scrollAngle = useScrollAngle(containerRef, prefersReducedMotion);
  const { pointerAngle, isPointerNear } = usePointerAngle(
    containerRef,
    prefersReducedMotion,
  );

  // Idle drift - shared across all instances
  const [idleAngle, setIdleAngle] = useState(0);
  useEffect(() => {
    if (prefersReducedMotion) return;

    const cleanup = addDriftListener((angle) => {
      if (componentMounted.current) {
        setIdleAngle(angle);
      }
    });

    return cleanup;
  }, [prefersReducedMotion]);

  // Calculate final angle
  let finalAngle = 140; // default static angle
  if (!prefersReducedMotion) {
    finalAngle = scrollAngle; // Base scroll angle

    // Blend with idle drift when pointer not near
    if (!isPointerNear) {
      finalAngle += idleAngle * 0.1; // Subtle idle influence
    }

    // Blend with pointer when near
    if (isPointerNear) {
      const pointerInfluence =
        (pointerAngle - finalAngle) * POINTER_SENSITIVITY;
      finalAngle += pointerInfluence;
    }
  }

  const dimensions = SIZES[size];
  const innerSize = dimensions.outer - dimensions.inner;

  // Cleanup on unmount
  useEffect(() => {
    componentMounted.current = true;
    return () => {
      componentMounted.current = false;
    };
  }, []);

  return (
    <div
      ref={containerRef}
      style={{
        position: "relative",
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        width: `${dimensions.outer}px`,
        height: `${dimensions.outer}px`,
        flexShrink: 0,
      }}
      aria-label={ariaLabel || label}
    >
      {/* Foil gradient outer circle */}
      <div
        style={{
          position: "absolute",
          inset: 0,
          borderRadius: "50%",
          background: `conic-gradient(from ${finalAngle}deg, var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-d), var(--foil-a))`,
          boxShadow: "var(--shadow-seal)",
          willChange: "transform",
        }}
      />

      {/* Center punch-out - inherits parent background */}
      <div
        style={{
          position: "absolute",
          inset: `${dimensions.inner}px`,
          borderRadius: "50%",
          background: "var(--panel)",
        }}
      />

      {/* Visually hidden label for accessibility */}
      <span
        style={{
          position: "absolute",
          width: "1px",
          height: "1px",
          padding: 0,
          margin: "-1px",
          overflow: "hidden",
          clip: "rect(0,0,0,0)",
          whiteSpace: "nowrap",
          border: 0,
        }}
      >
        {label}
      </span>
    </div>
  );
};
