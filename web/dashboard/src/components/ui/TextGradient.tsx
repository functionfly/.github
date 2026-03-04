/**
 * TextGradient Component
 *
 * A stunning text component with animated gradient colors that smoothly shift
 * through a customizable color palette. Perfect for headlines, CTAs, and
 * attention-grabbing text elements. Supports both heading and body text sizes
 * with smooth, performant CSS animations.
 *
 * @example
 * ```tsx
 * <TextGradient
 *   animate={true}
 *   colors={["#6366f1", "#8b5cf6", "#d946ef", "#6366f1"]}
 *   className="text-4xl font-bold"
 * >
 *   Stunning Gradient Text
 * </TextGradient>
 * ```
 */

import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

/**
 * Default gradient colors (brand palette)
 */
const DEFAULT_COLORS = [
  "#6366f1", // indigo-500
  "#8b5cf6", // violet-500
  "#d946ef", // fuchsia-500
  "#6366f1", // back to indigo for seamless loop
];

/**
 * Text size variants
 */
const textGradientVariants = cva(
  // Base styles
  [
    "inline-block",
    "bg-clip-text text-transparent",
    "transition-all duration-300",
  ],
  {
    variants: {
      size: {
        heading: "text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight",
        subheading: "text-2xl md:text-3xl lg:text-4xl font-semibold",
        body: "text-base md:text-lg font-medium",
        small: "text-sm font-medium",
      },
    },
    defaultVariants: {
      size: "heading",
    },
  }
);

export interface TextGradientProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof textGradientVariants> {
  /** Text content */
  children: React.ReactNode;
  /** Enable gradient animation */
  animate?: boolean;
  /** Custom gradient colors (hex format recommended) */
  colors?: string[];
  /** Animation duration in seconds */
  animationDuration?: number;
  /** Direction of gradient flow */
  direction?: "horizontal" | "vertical" | "diagonal";
  /** Render as different element */
  as?: "span" | "h1" | "h2" | "h3" | "h4" | "p";
}

/**
 * TextGradient - Animated gradient text component
 *
 * Uses CSS background-clip: text with an animated gradient background.
 * The gradient smoothly shifts through the specified colors creating
 * a mesmerizing text effect. Fully performant using CSS animations.
 */
const TextGradient = React.forwardRef<HTMLSpanElement, TextGradientProps>(
  (
    {
      className,
      children,
      size = "heading",
      animate = true,
      colors = DEFAULT_COLORS,
      animationDuration = 3,
      direction = "horizontal",
      as: Component = "span",
      ...props
    },
    ref
  ) => {
    // Generate gradient string from colors
    const gradientString = React.useMemo(() => {
      const percentageStep = 100 / (colors.length - 1);
      return colors
        .map((color, index) => `${color} ${index * percentageStep}%`)
        .join(", ");
    }, [colors]);

    // Determine gradient direction
    const gradientDirection = React.useMemo(() => {
      switch (direction) {
        case "vertical":
          return "to bottom";
        case "diagonal":
          return "135deg";
        case "horizontal":
        default:
          return "to right";
      }
    }, [direction]);

    // Unique ID for this instance's animation
    const animationId = React.useMemo(
      () => `text-gradient-${Math.random().toString(36).substr(2, 9)}`,
      []
    );

    return (
      <>
        <Component
          ref={ref as React.Ref<HTMLHeadingElement & HTMLParagraphElement>}
          className={cn(textGradientVariants({ size, className }))}
          style={{
            backgroundImage: `linear-gradient(${gradientDirection}, ${gradientString})`,
            backgroundSize: animate ? "200% 200%" : "100% 100%",
            animation: animate
              ? `${animationId} ${animationDuration}s ease infinite`
              : undefined,
          }}
          {...props}
        >
          {children}
        </Component>

        {/* CSS animation for this instance */}
        {animate && (
          <style>{`
            @keyframes ${animationId} {
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
          `}</style>
        )}
      </>
    );
  }
);

TextGradient.displayName = "TextGradient";

export { TextGradient, DEFAULT_COLORS };
