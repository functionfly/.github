/**
 * ParticleBackground Component
 *
 * A subtle, performant floating particles animation for hero sections and
 * background decoration. Uses pure CSS keyframes instead of heavy JavaScript
 * libraries for optimal performance. Particles gently float upward with
 * varying speeds and sizes for a mesmerizing ambient effect.
 *
 * @example
 * ```tsx
 * <div className="relative h-screen">
 *   <ParticleBackground particleCount={30} color="rgba(99, 102, 241, 0.3)" />
 *   <div className="relative z-10">Your content here</div>
 * </div>
 * ```
 */

import * as React from "react";
import { cn } from "@/lib/utils";

export interface ParticleBackgroundProps
  extends React.HTMLAttributes<HTMLDivElement> {
  /** Number of particles to render (recommended: 15-50) */
  particleCount?: number;
  /** Color of particles (supports rgba for transparency) */
  color?: string;
  /** Minimum particle size in pixels */
  minSize?: number;
  /** Maximum particle size in pixels */
  maxSize?: number;
  /** Animation duration range in seconds [min, max] */
  durationRange?: [number, number];
  /** Enable connections between nearby particles */
  enableConnections?: boolean;
}

/**
 * Generate random number between min and max
 */
const random = (min: number, max: number): number =>
  Math.random() * (max - min) + min;

/**
 * Particle interface for type safety
 */
interface Particle {
  id: number;
  x: number;
  y: number;
  size: number;
  duration: number;
  delay: number;
  opacity: number;
}

/**
 * ParticleBackground - Floating particles animation for backgrounds
 *
 * Creates a subtle, ambient particle effect using CSS animations.
 * Each particle floats upward at different speeds creating a dynamic,
 * organic feel perfect for hero sections and feature highlights.
 */
const ParticleBackground = React.forwardRef<
  HTMLDivElement,
  ParticleBackgroundProps
>(
  (
    {
      className,
      particleCount = 25,
      color = "rgba(99, 102, 241, 0.4)",
      minSize = 2,
      maxSize = 6,
      durationRange = [15, 30],
      enableConnections = false,
      ...props
    },
    ref
  ) => {
    // Generate particles with stable random values
    const particles = React.useMemo<Particle[]>(() => {
      return Array.from({ length: particleCount }, (_, i) => ({
        id: i,
        x: random(0, 100),
        y: random(0, 100),
        size: random(minSize, maxSize),
        duration: random(durationRange[0], durationRange[1]),
        delay: random(0, 20),
        opacity: random(0.3, 0.8),
      }));
    }, [particleCount, minSize, maxSize, durationRange]);

    return (
      <div
        ref={ref}
        className={cn(
          "absolute inset-0 overflow-hidden pointer-events-none",
          className
        )}
        {...props}
      >
        {/* Particles container */}
        <div className="absolute inset-0">
          {particles.map((particle) => (
            <div
              key={particle.id}
              className="absolute rounded-full animate-float-particle"
              style={{
                left: `${particle.x}%`,
                top: `${particle.y}%`,
                width: `${particle.size}px`,
                height: `${particle.size}px`,
                backgroundColor: color,
                opacity: particle.opacity,
                animationDuration: `${particle.duration}s`,
                animationDelay: `${particle.delay}s`,
                boxShadow: `0 0 ${particle.size * 2}px ${color}`,
              }}
            />
          ))}
        </div>

        {/* Optional connection lines overlay */}
        {enableConnections && (
          <svg
            className="absolute inset-0 w-full h-full opacity-20"
            xmlns="http://www.w3.org/2000/svg"
          >
            <defs>
              <linearGradient id="lineGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                <stop offset="0%" stopColor={color} stopOpacity="0" />
                <stop offset="50%" stopColor={color} stopOpacity="0.5" />
                <stop offset="100%" stopColor={color} stopOpacity="0" />
              </linearGradient>
            </defs>
          </svg>
        )}

        {/* CSS animations */}
        <style>{`
          @keyframes float-particle {
            0% {
              transform: translateY(100vh) translateX(0);
              opacity: 0;
            }
            10% {
              opacity: var(--particle-opacity, 0.5);
            }
            90% {
              opacity: var(--particle-opacity, 0.5);
            }
            100% {
              transform: translateY(-100vh) translateX(50px);
              opacity: 0;
            }
          }
          .animate-float-particle {
            animation: float-particle linear infinite;
          }
        `}</style>
      </div>
    );
  }
);

ParticleBackground.displayName = "ParticleBackground";

export { ParticleBackground };
