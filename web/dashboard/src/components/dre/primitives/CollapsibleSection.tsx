import { useState } from "react";
import { ChevronDown, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

export interface CollapsibleSectionProps {
  /** Section title */
  title: string;
  /** Whether the section is initially open */
  defaultOpen?: boolean;
  /** Whether the section is controlled */
  open?: boolean;
  /** Callback when open state changes */
  onOpenChange?: (open: boolean) => void;
  /** Optional badge count */
  badge?: number | string;
  /** Optional icon */
  icon?: React.ReactNode;
  /** Section content */
  children: React.ReactNode;
  /** Custom className */
  className?: string;
}

export function CollapsibleSection({
  title,
  defaultOpen = true,
  open: controlledOpen,
  onOpenChange,
  badge,
  icon,
  children,
  className,
}: CollapsibleSectionProps) {
  const [internalOpen, setInternalOpen] = useState(defaultOpen);

  const isControlled = controlledOpen !== undefined;
  const isOpen = isControlled ? controlledOpen : internalOpen;

  const handleToggle = () => {
    if (isControlled) {
      onOpenChange?.(!isOpen);
    } else {
      setInternalOpen(!internalOpen);
      onOpenChange?.(!internalOpen);
    }
  };

  return (
    <div className={cn("border border-border-subtle rounded-lg overflow-hidden", className)}>
      <button
        type="button"
        onClick={handleToggle}
        className="w-full flex items-center justify-between px-4 py-3 bg-bg-secondary/50 hover:bg-bg-secondary transition-colors text-left"
      >
        <div className="flex items-center gap-2">
          {icon && <span className="text-muted-foreground">{icon}</span>}
          <span className="font-medium text-sm">{title}</span>
          {badge !== undefined && (
            <span className="inline-flex items-center justify-center min-w-[20px] h-5 px-1.5 text-xs font-medium bg-bg-tertiary rounded-full">
              {badge}
            </span>
          )}
        </div>
        {isOpen ? (
          <ChevronDown className="h-4 w-4 text-muted-foreground" />
        ) : (
          <ChevronRight className="h-4 w-4 text-muted-foreground" />
        )}
      </button>
      {isOpen && (
        <div className="p-4 border-t border-border-subtle bg-bg-primary/30">
          {children}
        </div>
      )}
    </div>
  );
}
