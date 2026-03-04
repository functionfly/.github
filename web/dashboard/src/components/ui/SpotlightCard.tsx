/**
 * SpotlightCard Component
 *
 * An interactive card component featuring a dynamic spotlight effect that follows
 * the user's mouse cursor. Uses CSS custom properties for performant mouse position
 * tracking. Creates an engaging "wow" effect perfect for blog post cards,
 * feature highlights, and premium content sections.
 *
 * @example
 * ```tsx
 * <SpotlightCard spotlightColor="rgba(99, 102, 241, 0.15)">
 *   <img src="/blog-image.jpg" alt="Blog post" />
 *   <h3>Blog Post Title</h3>
 *   <p>Preview text...</p>
 * </SpotlightCard>
 * ```
 */

import * as React from "react";
import { cn } from "@/lib/utils";

export interface SpotlightCardProps
  extends React.HTMLAttributes<HTMLDivElement> {
  /** Content to render inside the card */
  children: React.ReactNode;
  /** Color of the spotlight gradient (supports rgba) */
  spotlightColor?: string;
  /** Size of the spotlight effect in pixels */
  spotlightSize?: number;
  /** Enable spotlight effect on hover only (vs always active) */
  hoverOnly?: boolean;
  /** Optional border radius override */
  borderRadius?: string;
}

/**
 * SpotlightCard - A card with a mouse-following spotlight gradient effect
 *
 * Uses CSS custom properties (--mouse-x, --mouse-y) to track cursor position
 * and applies a radial gradient that follows the mouse. This approach is
 * performant as it avoids React re-renders on mouse move.
 */
const SpotlightCard = React.forwardRef<HTMLDivElement, SpotlightCardProps>(
  (
    {
      className,
      children,
      spotlightColor = "rgba(99, 102, 241, 0.15)",
      spotlightSize = 300,
      hoverOnly = true,
      borderRadius = "1rem",
      ...props
    },
    ref
  ) => {
    const cardRef = React.useRef<HTMLDivElement>(null);
    const [isHovered, setIsHovered] = React.useState(false);

    // Merge refs
    React.useImperativeHandle(
      ref,
      () => cardRef.current as HTMLDivElement,
      []
    );

    /**
     * Update CSS custom properties with mouse position
     */
    const handleMouseMove = React.useCallback(
      (e: React.MouseEvent<HTMLDivElement>) => {
        if (!cardRef.current) return;

        const rect = cardRef.current.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        cardRef.current.style.setProperty("--mouse-x", `${x}px`);
        cardRef.current.style.setProperty("--mouse-y", `${y}px`);
      },
      []
    );

    /**
     * Handle mouse enter
     */
    const handleMouseEnter = React.useCallback(() => {
      setIsHovered(true);
    }, []);

    /**
     * Handle mouse leave - reset spotlight position
     */
    const handleMouseLeave = React.useCallback(() => {
      setIsHovered(false);
      if (cardRef.current) {
        // Reset to center when not hovering
        const rect = cardRef.current.getBoundingClientRect();
        cardRef.current.style.setProperty(
          "--mouse-x",
          `${rect.width / 2}px`
        );
        cardRef.current.style.setProperty(
          "--mouse-y",
          `${rect.height / 2}px`
        );
      }
    }, []);

    // Initialize center position on mount
    React.useEffect(() => {
      if (cardRef.current) {
        const rect = cardRef.current.getBoundingClientRect();
        cardRef.current.style.setProperty(
          "--mouse-x",
          `${rect.width / 2}px`
        );
        cardRef.current.style.setProperty(
          "--mouse-y",
          `${rect.height / 2}px`
        );
      }
    }, []);

    return (
      <div
        ref={cardRef}
        className={cn(
          // Base styles - use theme-aware card background (light in light mode, dark in dark mode)
          "relative overflow-hidden",
          "bg-card",
          "border border-border-default",
          "transition-all duration-300 ease-out",
          "group",
          className
        )}
        style={{
          borderRadius,
          // CSS custom properties for spotlight position
          "--mouse-x": "50%",
          "--mouse-y": "50%",
          "--spotlight-color": spotlightColor,
          "--spotlight-size": `${spotlightSize}px`,
        } as React.CSSProperties}
        onMouseMove={handleMouseMove}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        {...props}
      >
        {/* Spotlight gradient overlay */}
        <div
          className={cn(
            "absolute inset-0 pointer-events-none transition-opacity duration-300",
            hoverOnly ? (isHovered ? "opacity-100" : "opacity-0") : "opacity-100"
          )}
          style={{
            background: `radial-gradient(
              var(--spotlight-size) circle at var(--mouse-x) var(--mouse-y),
              var(--spotlight-color),
              transparent 80%
            )`,
          }}
          aria-hidden="true"
        />

        {/* Subtle border glow on hover */}
        <div
          className={cn(
            "absolute inset-0 pointer-events-none rounded-[inherit]",
            "transition-opacity duration-300",
            "ring-1 ring-inset",
            isHovered
              ? "ring-brand-500/30 opacity-100"
              : "ring-transparent opacity-0"
          )}
          aria-hidden="true"
        />

        {/* Content wrapper */}
        <div className="relative z-10">{children}</div>
      </div>
    );
  }
);

SpotlightCard.displayName = "SpotlightCard";

export { SpotlightCard };
