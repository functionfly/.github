import * as React from "react";
import { cn } from "./utils";

export interface SelectProps {
  value?: string;
  onValueChange?: (value: string) => void;
  children: React.ReactNode;
  className?: string;
}

export function Select({ value, onValueChange, children, className }: SelectProps) {
  const [open, setOpen] = React.useState(false);

  return (
    <div className={cn("relative", className)}>
      {React.Children.map(children, (child) => {
        if (React.isValidElement(child)) {
          return React.cloneElement(child as React.ReactElement<any>, {
            open,
            onOpenChange: setOpen,
            value,
            onValueChange,
          });
        }
        return child;
      })}
    </div>
  );
}

export interface SelectTriggerProps {
  children: React.ReactNode;
  className?: string;
  asChild?: boolean;
}

export function SelectTrigger({ children, className, asChild }: SelectTriggerProps) {
  return (
    <button
      type="button"
      className={cn(
        "flex h-9 w-full items-center justify-between whitespace-nowrap rounded-md border border-border-subtle bg-bg-primary px-3 py-2 text-sm text-text-primary hover:bg-bg-hover focus:outline-none focus:ring-2 focus:ring-brand-500 disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
    >
      {children}
    </button>
  );
}

export function SelectValue({ placeholder }: { placeholder?: string }) {
  return <span className="text-text-muted">{placeholder}</span>;
}

export interface SelectContentProps {
  children: React.ReactNode;
  className?: string;
  open?: boolean;
}

export function SelectContent({ children, className, open }: SelectContentProps) {
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        // Would trigger close
      }
    };
    if (open) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [open]);

  if (!open) return null;

  return (
    <div
      ref={ref}
      className={cn(
        "absolute z-50 mt-1 min-w-[8rem] overflow-hidden rounded-md border border-border-subtle bg-bg-primary p-1 shadow-lg",
        className
      )}
    >
      {children}
    </div>
  );
}

export interface SelectItemProps {
  value: string;
  children: React.ReactNode;
  className?: string;
}

export function SelectItem({ value, children, className }: SelectItemProps) {
  return (
    <button
      type="button"
      className={cn(
        "relative flex w-full cursor-pointer select-none items-center rounded-sm py-1.5 px-2 text-sm text-text-primary hover:bg-bg-hover",
        className
      )}
    >
      {children}
    </button>
  );
}