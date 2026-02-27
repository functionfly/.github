import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cn } from "@/lib/utils";

const Input = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement> & {
    asChild?: boolean;
  }
>(({ className, type, asChild = false, ...props }, ref) => {
  const Comp = asChild ? Slot : "input";
  return (
    <Comp
      type={type}
      className={cn(
        "flex h-11 w-full rounded-lg border px-4 py-2 text-sm shadow-sm transition-all duration-200",
        "placeholder:text-text-muted",
        "focus:outline-none focus:border-border-focus focus:ring-2 focus:ring-border-focus",
        "disabled:cursor-not-allowed disabled:opacity-50",
        "file:border-0 file:bg-transparent file:text-sm file:font-medium",
        className
      )}
      style={{
        borderColor: 'var(--border-subtle)',
        backgroundColor: 'var(--input)',
        color: 'var(--input-foreground)',
      } as React.CSSProperties}
      ref={ref}
      {...props}
    />
  );
});
Input.displayName = "Input";

export { Input };
