/**
 * @functionfly/ui-core
 * Glass morphism card component
 */

import * as React from "react";
import { cn } from "./utils";

export interface GlassCardProps extends React.HTMLAttributes<HTMLDivElement> {
  glowColor?: string;
  intensity?: "low" | "medium" | "high";
  animated?: boolean;
  glass?: boolean;
}

export function GlassCard({
  className,
  glowColor = "#f97316",
  intensity = "medium",
  animated = false,
  glass = true,
  children,
  ...props
}: GlassCardProps) {
  const blurValue = intensity === "low" ? 8 : intensity === "high" ? 24 : 16;
  const alphaValue = intensity === "low" ? 0.03 : intensity === "high" ? 0.1 : 0.06;

  return (
    <div
      className={cn(
        "relative rounded-2xl border transition-all duration-300",
        glass && "backdrop-blur-xl",
        animated && "animate-border-glow",
        className
      )}
      style={
        glass
          ? ({
              "--glow-color": glowColor,
              "--blur-value": `${blurValue}px`,
              "--alpha-value": alphaValue,
              background: `rgba(255, 255, 255, var(--alpha-value, ${alphaValue}))`,
              backdropFilter: `blur(var(--blur-value, ${blurValue}px))`,
              WebkitBackdropFilter: `blur(var(--blur-value, ${blurValue}px))`,
              borderColor: `rgba(255, 255, 255, 0.1)`,
              boxShadow: `0 0 20px color-mix(in srgb, ${glowColor} 15%, transparent)`,
            } as React.CSSProperties)
          : {}
      }
      {...props}
    >
      {children}
    </div>
  );
}