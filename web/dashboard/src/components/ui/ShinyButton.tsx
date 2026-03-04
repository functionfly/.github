/**
 * ShinyButton Component
 *
 * A premium button component with a metallic shine animation on hover.
 * Creates an eye-catching "wow" effect using only Tailwind CSS animations.
 * The shine effect sweeps across the button on hover, giving it a polished,
 * premium feel perfect for CTAs and important actions.
 *
 * @example
 * ```tsx
 * <ShinyButton variant="default" size="lg" onClick={handleClick}>
 *   Get Started
 * </ShinyButton>
 * ```
 */

import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

/**
 * Button variants with shine effect styling
 */
const shinyButtonVariants = cva(
  // Base styles
  [
    "relative inline-flex items-center justify-center gap-2",
    "whitespace-nowrap rounded-lg text-sm font-semibold",
    "transition-all duration-200 ease-out",
    "focus-visible:outline-none focus-visible:ring-2",
    "focus-visible:ring-border-focus focus-visible:ring-offset-2",
    "disabled:pointer-events-none disabled:opacity-50",
    "overflow-hidden",
    "[&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
  ],
  {
    variants: {
      variant: {
        default: [
          // Gradient background
          "bg-gradient-to-r from-brand-500 via-violet-500 to-purple-500",
          "text-white",
          // Shine effect
          "before:absolute before:inset-0",
          "before:bg-gradient-to-r before:from-transparent",
          "before:via-white/30 before:to-transparent",
          "before:-translate-x-full before:animate-shine",
          "hover:before:animate-shine-hover",
          // Hover scale
          "hover:scale-[1.02] hover:shadow-lg hover:shadow-brand-500/25",
          "active:scale-[0.98]",
        ],
        secondary: [
          // Light metallic background
          "bg-gradient-to-b from-white to-gray-100",
          "dark:from-gray-800 dark:to-gray-900",
          "text-gray-900 dark:text-white",
          "border border-gray-200 dark:border-gray-700",
          // Shine effect
          "before:absolute before:inset-0",
          "before:bg-gradient-to-r before:from-transparent",
          "before:via-white/60 dark:before:via-white/20 before:to-transparent",
          "before:-translate-x-full",
          "hover:before:animate-shine-hover",
          // Hover
          "hover:shadow-lg hover:shadow-black/10",
          "active:scale-[0.98]",
        ],
        outline: [
          // Transparent with border
          "bg-transparent",
          "border-2 border-brand-500 text-brand-500",
          "dark:border-brand-400 dark:text-brand-400",
          // Shine effect
          "before:absolute before:inset-0",
          "before:bg-gradient-to-r before:from-transparent",
          "before:via-brand-500/20 before:to-transparent",
          "before:-translate-x-full",
          "hover:before:animate-shine-hover",
          // Hover
          "hover:bg-brand-500/10",
          "active:scale-[0.98]",
        ],
        ghost: [
          // Ghost style
          "bg-transparent text-text-primary",
          // Shine effect
          "before:absolute before:inset-0",
          "before:bg-gradient-to-r before:from-transparent",
          "before:via-black/5 dark:before:via-white/10 before:to-transparent",
          "before:-translate-x-full",
          "hover:before:animate-shine-hover",
          // Hover
          "hover:bg-black/5 dark:hover:bg-white/10",
          "active:scale-[0.98]",
        ],
        destructive: [
          // Red gradient
          "bg-gradient-to-r from-red-500 to-red-600",
          "text-white",
          // Shine effect
          "before:absolute before:inset-0",
          "before:bg-gradient-to-r before:from-transparent",
          "before:via-white/30 before:to-transparent",
          "before:-translate-x-full",
          "hover:before:animate-shine-hover",
          // Hover
          "hover:scale-[1.02] hover:shadow-lg hover:shadow-red-500/25",
          "active:scale-[0.98]",
        ],
      },
      size: {
        default: "h-10 px-6 py-2",
        sm: "h-8 rounded-md px-4 text-xs",
        lg: "h-12 rounded-xl px-8 text-base",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

/**
 * Props for the ShinyButton component
 */
export interface ShinyButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof shinyButtonVariants> {
  /** Render as a child component (useful for links) */
  asChild?: boolean;
  /** Loading state */
  isLoading?: boolean;
}

/**
 * ShinyButton - A button with a metallic shine animation on hover
 *
 * Uses CSS transforms and pseudo-elements to create a sweeping shine effect
 * that moves across the button when hovered, creating a premium metallic look.
 */
const ShinyButton = React.forwardRef<HTMLButtonElement, ShinyButtonProps>(
  (
    {
      className,
      variant,
      size,
      asChild = false,
      isLoading,
      children,
      disabled,
      ...props
    },
    ref
  ) => {
    // Common style element for animations
    const styleElement = (
      <style>{`
        @keyframes shine-hover {
          0% {
            transform: translateX(-100%);
          }
          100% {
            transform: translateX(100%);
          }
        }
        .animate-shine-hover:hover::before {
          animation: shine-hover 0.6s ease-out;
        }
      `}</style>
    );

    if (asChild) {
      const child = React.Children.only(children);
      const childChildren = child.props.children;

      return (
        <>
          {styleElement}
          <Slot
            className={cn(shinyButtonVariants({ variant, size, className }))}
            ref={ref}
            {...props}
          >
            {React.cloneElement(child, {
              children: isLoading ? (
                <>
                  <svg
                    className="animate-spin h-4 w-4 mr-2"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  {childChildren}
                </>
              ) : childChildren
            })}
          </Slot>
        </>
      );
    }

    return (
      <button
        className={cn(shinyButtonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || isLoading}
        {...props}
      >
        {styleElement}
        {/* Content wrapper */}
        <span className="relative z-10 flex items-center gap-2">
          {isLoading && (
            <svg
              className="animate-spin h-4 w-4"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
          )}
          {children}
        </span>
      </button>
    );
  }
);

ShinyButton.displayName = "ShinyButton";

export { ShinyButton, shinyButtonVariants };
