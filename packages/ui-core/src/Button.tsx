import * as React from "react";
import { cn } from "./utils";

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "secondary" | "ghost" | "destructive" | "outline" | "success" | "warning";
  size?: "default" | "sm" | "lg" | "icon";
  asChild?: boolean;
}

const variantClasses = {
  default: "bg-brand-500 text-white hover:bg-brand-600",
  secondary: "bg-bg-secondary text-text-primary hover:bg-bg-hover",
  ghost: "text-text-muted hover:text-text-primary hover:bg-white/10",
  destructive: "bg-red-500 text-white hover:bg-red-600",
  outline: "border border-border-subtle text-text-secondary hover:bg-bg-hover",
  success: "bg-emerald-500 text-white hover:bg-emerald-600",
  warning: "bg-amber-500 text-white hover:bg-amber-600",
};

const sizeClasses = {
  default: "h-10 px-4 py-2",
  sm: "h-8 px-3 text-sm",
  lg: "h-12 px-6 text-lg",
  icon: "h-8 w-8",
};

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", ...props }, ref) => {
    return (
      <button
        className={cn(
          "inline-flex items-center justify-center gap-2 rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-brand-500 focus:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
          variantClasses[variant],
          sizeClasses[size],
          className
        )}
        ref={ref}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";