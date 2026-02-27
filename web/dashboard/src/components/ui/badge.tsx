import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-(--ring) focus:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-gradient-to-r from-brand-500 to-purple-500 text-white",
        secondary:
          "border-border-subtle bg-bg-tertiary text-text-muted",
        destructive:
          "border-transparent bg-gradient-to-r from-error to-red-400 text-white",
        outline:
          "border-border-default text-text-primary",
        success:
          "border-transparent bg-success-glow text-success",
        warning:
          "border-transparent bg-warning-glow text-warning",
        error:
          "border-transparent bg-error-glow text-error",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
