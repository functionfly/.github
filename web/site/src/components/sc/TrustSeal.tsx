import React, { useEffect, useRef, useCallback, useState } from "react";
import { motion, useScroll, useTransform, useMotionValue, useAnimationFrame } from "framer-motion";
import { useReducedMotion } from "./ReducedMotionGate";

interface TrustSealProps {
  label: string;
  variant?: "live" | "verified" | "trust";
  size?: "sm" | "md" | "lg";
  ariaLabel?: string;
}

const SIZE_CONFIG = {
  sm: { outer: 16, inner: 3, inset: 3 },
  md: { outer: 22, inner: 4, inset: 4 },
  lg: { outer: 32, inner: 6, inset: 6 },
} as const;

const POINTER_TRACKING_RADIUS = 80;
const POINTER_SENSITIVITY = 0.4;
const IDLE_DRIFT_SPEED = 1;
const SCROLL_ROTATION_RANGE = 720;

export const TrustSeal: React.FC<TrustSealProps> = ({
  label,
  variant = "verified",
  size = "md",
  ariaLabel,
}) => {
  const prefersReducedMotion = useReducedMotion();
  const containerRef = useRef<HTMLDivElement>(null);
  const sealRef = useRef<HTMLDivElement>(null);
  const pointerAngleRef = useRef<number>(0);
  const isPointerNearRef = useRef(false);
  const idleAngleRef = useRef(140);
  const rafRef = useRef<number>(0);
  const lastIdleUpdateRef = useRef<number>(0);

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start end", "end start"],
  });

  const scrollAngle = useTransform(
    scrollYProgress,
    [0, 1],
    [0, SCROLL_ROTATION_RANGE]
  );

  const idleAngle = useMotionValue(140);

  useAnimationFrame((timestamp) => {
    if (prefersReducedMotion || !sealRef.current) return;

    if (lastIdleUpdateRef.current === 0) {
      lastIdleUpdateRef.current = timestamp;
    }

    const elapsed = timestamp - lastIdleUpdateRef.current;
    lastIdleUpdateRef.current = timestamp;

    idleAngleRef.current += (elapsed / 1000) * IDLE_DRIFT_SPEED;
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
    const distance = Math.sqrt(
      Math.pow(e.clientX - cx, 2) + Math.pow(e.clientY - cy, 2)
    );
    return distance <= POINTER_TRACKING_RADIUS;
  }, []);

  useEffect(() => {
    if (prefersReducedMotion || !containerRef.current) return;

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
    element.addEventListener("pointermove", handlePointerMove);
    element.addEventListener("pointerleave", handlePointerLeave);

    return () => {
      element.removeEventListener("pointermove", handlePointerMove);
      element.removeEventListener("pointerleave", handlePointerLeave);
    };
  }, [prefersReducedMotion, calculatePointerAngle, isPointerInRange]);

  useEffect(() => {
    if (prefersReducedMotion || !sealRef.current) return;

    let currentScrollAngle = 0;
    let currentIdleAngle = 140;
    let combinedAngle = 140;

    const unsubscribeScroll = scrollAngle.on("change", (value) => {
      currentScrollAngle = value;
      updateAngle();
    });

    const unsubscribeIdle = idleAngle.on("change", (value) => {
      currentIdleAngle = value;
      updateAngle();
    });

    const updateAngle = () => {
      if (isPointerNearRef.current) {
        combinedAngle = currentScrollAngle + (pointerAngleRef.current - currentIdleAngle) * POINTER_SENSITIVITY;
      } else {
        combinedAngle = currentScrollAngle + currentIdleAngle;
      }
      sealRef.current!.style.setProperty("--seal-angle", `${combinedAngle}deg`);
    };

    return () => {
      unsubscribeScroll();
      unsubscribeIdle();
    };
  }, [prefersReducedMotion, scrollAngle, idleAngle]);

  const accessibleLabel = ariaLabel ?? `${variant}: ${label}`;
  const sizeConfig = SIZE_CONFIG[size] ?? SIZE_CONFIG['md'];

  return (
    <div
      ref={containerRef}
      role="img"
      aria-label={accessibleLabel}
      className="inline-flex items-center gap-2 cursor-default"
    >
      <div
        ref={sealRef}
        className="relative rounded-full trust-seal-foil"
        style={{
          width: sizeConfig.outer,
          height: sizeConfig.outer,
          background: `conic-gradient(from var(--seal-angle, 140deg), var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-d), var(--foil-a))`,
        }}
      >
        <div
          className="absolute rounded-full"
          style={{
            inset: sizeConfig.inset,
            backgroundColor: 'var(--panel)',
          }}
        />
      </div>

      <span
        className="font-[var(--font-mono)] font-medium uppercase tracking-wider trust-seal-label"
        style={{
          color: 'var(--status-ok)',
          fontSize: size === 'sm' ? '10px' : size === 'lg' ? '12px' : '11px',
        }}
      >
        {label}
      </span>
    </div>
  );
};
