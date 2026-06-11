/**
 * @functionfly/ui-core
 * Tab component for Workspace sections
 */

import * as React from "react";
import { cn } from "./utils";

export interface TabsProps {
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  children: React.ReactNode;
  className?: string;
}

export interface TabsListProps {
  children: React.ReactNode;
  className?: string;
}

export interface TabsTriggerProps {
  value: string;
  children: React.ReactNode;
  disabled?: boolean;
  className?: string;
  icon?: React.ReactNode;
}

export interface TabsContentProps {
  value: string;
  children: React.ReactNode;
  className?: string;
}

const TabsContext = React.createContext<{
  value: string;
  onValueChange: (value: string) => void;
} | null>(null);

export function Tabs({ value, defaultValue, onValueChange, children, className }: TabsProps) {
  const [internalValue, setInternalValue] = React.useState(defaultValue || value);

  const currentValue = value ?? internalValue;

  const handleValueChange = (newValue: string) => {
    setInternalValue(newValue);
    onValueChange?.(newValue);
  };

  const contextValue = React.useMemo(() => ({ value: currentValue ?? "", onValueChange: handleValueChange }), [currentValue]);
  return (
    <TabsContext.Provider value={contextValue}>
      <div className={cn("flex flex-col", className)}>{children}</div>
    </TabsContext.Provider>
  );
}

export function TabsList({ children, className }: TabsListProps) {
  return (
    <div
      className={cn(
        "flex items-center gap-1 rounded-t-lg border-b border-border-subtle bg-bg-secondary/50 p-1",
        className
      )}
    >
      {children}
    </div>
  );
}

export function TabsTrigger({ value, children, disabled, className, icon }: TabsTriggerProps) {
  const context = React.useContext(TabsContext);
  if (!context) throw new Error("TabsTrigger must be used within Tabs");

  const isActive = context.value === value;

  return (
    <button
      disabled={disabled}
      onClick={() => context.onValueChange(value)}
      className={cn(
        "inline-flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-all duration-200",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-border-focus focus-visible:ring-offset-2",
        isActive
          ? "bg-bg-primary text-text-primary shadow-sm border border-border-default"
          : "text-text-secondary hover:text-text-primary hover:bg-bg-hover",
        disabled && "opacity-50 cursor-not-allowed",
        className
      )}
    >
      {icon && <span className="size-4">{icon}</span>}
      {children}
    </button>
  );
}

export function TabsContent({ value, children, className }: TabsContentProps) {
  const context = React.useContext(TabsContext);
  if (!context) throw new Error("TabsContent must be used within Tabs");

  if (context.value !== value) return null;

  return <div className={cn("mt-2", className)}>{children}</div>;
}