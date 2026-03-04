/**
 * GlassmorphismCard Component
 *
 * A modern glass-like card component featuring backdrop blur, subtle borders,
 * and transparency effects. Perfect for creating depth and layering in modern UI designs.
 * Works beautifully in both light and dark modes with appropriate opacity adjustments.
 *
 * @example
 * ```tsx
 * <GlassmorphismCard blurAmount={12}>
 *   <h3>Glass Card</h3>
 *   <p>Content with a beautiful glass morphism effect.</p>
 * </GlassmorphismCard>
 * ```
 */

import * as React from "react";
import { cn } from "@/lib/utils";

export interface GlassmorphismCardProps
  extends React.HTMLAttributes<HTMLDivElement> {
  /** Content to render inside the card */
  children: React.ReactNode;
  /** Amount of backdrop blur (px) - higher values create stronger glass effect */
  blurAmount?: number;
  /** Optional header content */
  header?: React.ReactNode;
  /** Optional footer content */
  footer?: React.ReactNode;
}

/**
 * GlassmorphismCard - A card with glass-like transparency and blur effects
 *
 * Uses CSS backdrop-filter for the blur effect and rgba colors for transparency.
 * Automatically adapts to light/dark mode with appropriate background and border colors.
 */
const GlassmorphismCard = React.forwardRef<
  HTMLDivElement,
  GlassmorphismCardProps
>(({ className, children, blurAmount = 12, header, footer, ...props }, ref) => {
  return (
    <div
      ref={ref}
      className={cn(
        // Base styles
        "relative overflow-hidden rounded-2xl",
        // Glass effect background
        "bg-white/70 dark:bg-white/5",
        // Backdrop blur
        "backdrop-blur-md",
        // Border - subtle in light mode, slightly more visible in dark
        "border border-white/40 dark:border-white/10",
        // Shadow for depth
        "shadow-xl shadow-black/5 dark:shadow-black/20",
        // Hover effect
        "transition-all duration-300 ease-out",
        "hover:bg-white/80 dark:hover:bg-white/10",
        "hover:shadow-2xl hover:shadow-black/10 dark:hover:shadow-black/30",
        className
      )}
      style={{
        backdropFilter: `blur(${blurAmount}px)`,
        WebkitBackdropFilter: `blur(${blurAmount}px)`,
      }}
      {...props}
    >
      {/* Subtle gradient overlay for depth */}
      <div
        className={cn(
          "absolute inset-0 pointer-events-none",
          "bg-gradient-to-br from-white/50 via-transparent to-transparent",
          "dark:from-white/10 dark:via-transparent dark:to-transparent"
        )}
        aria-hidden="true"
      />

      {/* Inner glow effect */}
      <div
        className={cn(
          "absolute inset-0 pointer-events-none rounded-2xl",
          "ring-1 ring-inset ring-white/60 dark:ring-white/10"
        )}
        aria-hidden="true"
      />

      {/* Content container */}
      <div className="relative z-10 flex flex-col">
        {/* Header */}
        {header && (
          <div
            className={cn(
              "px-6 py-4 border-b",
              "border-black/5 dark:border-white/5"
            )}
          >
            {header}
          </div>
        )}

        {/* Main content */}
        <div className={cn("flex-1", !header ? "p-6" : "px-6 py-4")}>
          {children}
        </div>

        {/* Footer */}
        {footer && (
          <div
            className={cn(
              "px-6 py-4 border-t",
              "border-black/5 dark:border-white/5",
              "bg-black/[0.02] dark:bg-white/[0.02]"
            )}
          >
            {footer}
          </div>
        )}
      </div>
    </div>
  );
});

GlassmorphismCard.displayName = "GlassmorphismCard";

export { GlassmorphismCard };
