/**
 * AnimatedGradientCard Component
 *
 * A visually stunning card component with animated gradient borders that smoothly
 * shift between brand colors (indigo-500, violet-500, purple-500). Creates a
 * mesmerizing "wow" effect perfect for highlighting premium content.
 *
 * @example
 * ```tsx
 * <AnimatedGradientCard intensity="medium">
 *   <h3>Premium Feature</h3>
 *   <p>This content is highlighted with an animated gradient border.</p>
 * </AnimatedGradientCard>
 * ```
 */

import * as React from "react";
import { cn } from "@/lib/utils";

export interface AnimatedGradientCardProps
  extends React.HTMLAttributes<HTMLDivElement> {
  /** Content to render inside the card */
  children: React.ReactNode;
  /** Visual intensity of the gradient animation */
  intensity?: "subtle" | "medium" | "strong";
}

/**
 * Intensity configurations for the gradient border
 */
const intensityConfig = {
  subtle: {
    gradient: "from-indigo-500/30 via-violet-500/30 to-purple-500/30",
    glow: "opacity-20",
    width: "1px",
  },
  medium: {
    gradient: "from-indigo-500/60 via-violet-500/60 to-purple-500/60",
    glow: "opacity-40",
    width: "2px",
  },
  strong: {
    gradient: "from-indigo-500 via-violet-500 to-purple-500",
    glow: "opacity-60",
    width: "3px",
  },
} as const;

/**
 * AnimatedGradientCard - A card with animated gradient borders
 *
 * Uses CSS keyframes for smooth color transitions without heavy JavaScript.
 * The gradient rotates continuously creating a flowing border effect.
 */
const AnimatedGradientCard = React.forwardRef<
  HTMLDivElement,
  AnimatedGradientCardProps
>(({ className, children, intensity = "medium", ...props }, ref) => {
  const config = intensityConfig[intensity];

  return (
    <div
      ref={ref}
      className={cn("relative group", className)}
      {...props}
    >
      {/* Animated gradient border */}
      <div
        className={cn(
          "absolute -inset-[1px] rounded-xl bg-gradient-to-r",
          config.gradient,
          "animate-gradient-rotate",
          "dark:brightness-100 brightness-90"
        )}
        style={{
          backgroundSize: "200% 200%",
        }}
        aria-hidden="true"
      />

      {/* Glow effect */}
      <div
        className={cn(
          "absolute -inset-1 rounded-xl bg-gradient-to-r from-indigo-500 to-purple-500",
          "blur-lg transition-opacity duration-500",
          config.glow,
          "group-hover:opacity-100",
          "dark:group-hover:opacity-80"
        )}
        aria-hidden="true"
      />

      {/* Inner card content */}
      <div
        className={cn(
          "relative rounded-xl bg-bg-secondary dark:bg-bg-secondary",
          "border border-border-default",
          "p-6 h-full"
        )}
      >
        {children}
      </div>

      {/* CSS for animation */}
      <style>{`
        @keyframes gradient-rotate {
          0% {
            background-position: 0% 50%;
          }
          50% {
            background-position: 100% 50%;
          }
          100% {
            background-position: 0% 50%;
          }
        }
        .animate-gradient-rotate {
          animation: gradient-rotate 3s ease infinite;
        }
      `}</style>
    </div>
  );
});

AnimatedGradientCard.displayName = "AnimatedGradientCard";

export { AnimatedGradientCard };
