import { cn } from "@/lib/utils";
import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-border- focus:ring-offset-2",
  {
    variants: {
      variant: {
        default: "border-transparent bg-brand-500 text-white shadow-sm",
        secondary: "border-border-subtle bg-bg-tertiary text-text-secondary",
        destructive: "border-transparent bg-red-500 text-white shadow-sm",
        outline: "border-border-default text-text-primary bg-transparent",
        success: "border-transparent bg-emerald-500 text-white shadow-sm",
        warning: "border-transparent bg-amber-500 text-white shadow-sm",
        error: "border-transparent bg-red-500 text-white shadow-sm",
        info: "border-transparent bg-blue-500 text-white shadow-sm",
        ghost:
          "border-transparent bg-transparent text-text-muted hover:bg-bg-hover hover:text-text-primary",
      },
      size: {
        default: "px-2.5 py-0.5",
        sm: "px-2 py-0.5 text-[10px]",
        lg: "px-3 py-1 text-sm",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

export interface BadgeProps
  extends
    React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, size, ...props }: BadgeProps) {
  return (
    <div
      className={cn(badgeVariants({ variant, size, className }))}
      {...props}
    />
  );
}

export { Badge, badgeVariants };
