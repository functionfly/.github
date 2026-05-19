/**
 * @functionfly/ui-core
 * Badge component
 */

import * as React from "react";
import { cn } from "./utils";

export type BadgeVariant =
  | "default"
  | "brand"
  | "success"
  | "error"
  | "warning"
  | "info"
  | "ghost"
  | "outline";

export interface BadgeProps {
  variant?: BadgeVariant;
  size?: "sm" | "md" | "lg";
  dot?: boolean;
  pulse?: boolean;
  className?: string;
  children?: React.ReactNode;
  [key: string]: unknown;
}

export function Badge({
  className,
  variant = "default",
  size = "md",
  dot,
  pulse,
  children,
  ...props
}: BadgeProps) {
  const sizeClasses = {
    sm: "px-1.5 py-0.5 text-[10px]",
    md: "px-2 py-0.5 text-[11px]",
    lg: "px-2.5 py-1 text-xs",
  };

  const variantClasses = {
    default: "bg-bg-tertiary text-text-secondary",
    brand: "bg-brand-500/20 text-brand-400 border border-brand-500/30",
    success: "bg-success/20 text-success border border-success/30",
    error: "bg-error/20 text-error border border-error/30",
    warning: "bg-warning/20 text-warning border border-warning/30",
    info: "bg-info/20 text-info border border-info/30",
    ghost: "bg-transparent text-text-secondary border border-border-subtle",
    outline: "bg-transparent text-text-primary border border-border-default",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full font-medium transition-all duration-200",
        sizeClasses[size],
        variantClasses[variant],
        pulse && "relative",
        className
      )}
      {...props}
    >
      {dot && (
        <span
          className={cn(
            "absolute -top-0.5 -right-0.5 size-2 rounded-full border-2",
            "border-bg-primary",
            pulse ? "animate-pulse" : ""
          )}
          style={{
            backgroundColor:
              variant === "success" ? "#10b981" : variant === "error" ? "#ef4444" : variant === "warning" ? "#f59e0b" : variant === "info" ? "#3b82f6" : "#f97316",
          }}
        />
      )}
      {children}
    </span>
  );
}