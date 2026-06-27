import { type ButtonHTMLAttributes } from "react";
import { cn } from "../lib/utils";
import { FrameButton, SealedButton } from "./sc";

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
  // Map variants to SC button components
  if (variant === "secondary") {
    return (
      <FrameButton
        loading={loading}
        disabled={disabled}
        className={cn(fullWidth && "full-width")}
        style={fullWidth ? { width: "100%" } : undefined}
        {...rest}
      >
        {children}
      </FrameButton>
    );
  }

  if (variant === "ghost") {
    // Ghost uses FrameButton with no visible border
    return (
      <FrameButton
        loading={loading}
        disabled={disabled}
        className={cn("ghost-btn", fullWidth && "full-width", className)}
        style={
          fullWidth
            ? { width: "100%", border: "none", boxShadow: "none" }
            : { border: "none", boxShadow: "none" }
        }
        {...rest}
      >
        {children}
      </FrameButton>
    );
  }

  // Primary uses SealedButton
  return (
    <SealedButton
      loading={loading}
      disabled={disabled}
      className={cn(fullWidth && "full-width", className)}
      style={fullWidth ? { width: "100%" } : undefined}
      {...rest}
    >
      {children}
    </SealedButton>
  );
}
