/**
 * @functionfly/ui-core
 * Loading/spinner component
 */

import * as React from "react";
import { cn } from "./utils";

export interface SpinnerProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: "sm" | "md" | "lg" | "xl";
  variant?: "default" | "brand" | "monochrome";
}

export function Spinner({
  className,
  size = "md",
  variant = "default",
  children,
  ...props
}: SpinnerProps) {
  const sizeClasses = {
    sm: "size-4",
    md: "size-6",
    lg: "size-8",
    xl: "size-10",
  };

  const variantClasses = {
    default: "text-brand-500",
    brand: "text-brand-500",
    monochrome: "text-text-primary",
  };

  return (
    <div
      className={cn("flex flex-col items-center gap-2", className)}
      {...props}
    >
      <div
        className={cn(
          "rounded-full border-2 border-border-subtle border-t-brand-500 animate-spin",
          sizeClasses[size],
          variantClasses[variant]
        )}
        style={{ borderTopColor: "currentColor" }}
        role="progressbar"
        aria-label="Loading"
      />
      {children && <span className="text-sm text-text-secondary">{children}</span>}
    </div>
  );
}