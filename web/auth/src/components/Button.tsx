import React, { type ButtonHTMLAttributes } from "react";
import { cn } from "../lib/utils";

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "ghost";
  loading?: boolean;
  fullWidth?: boolean;
}

export default function Button({
  variant = "primary",
  loading = false,
  fullWidth = false,
  children,
  disabled,
  className = "",
  ...rest
}: Props) {
  return (
    <button
      disabled={disabled || loading}
      className={cn(
        // Base button styles from component library
        "ff-btn",
        
        // Variant-specific classes
        variant === "primary" && "ff-btn--primary",
        variant === "secondary" && "ff-btn--secondary",
        variant === "ghost" && "ff-btn--ghost",
        
        // State modifiers
        loading && "ff-btn--loading",
        fullWidth && "ff-btn--full",
        
        // External class overrides
        className
      )}
      {...rest}
    >
      {children}
    </button>
  );
}
