import React, { createContext, useContext, useState, useCallback } from 'react';
import { cn } from '@/lib/utils';

type TabsContextValue = {
  activeTab: string;
  setActiveTab: (id: string) => void;
};

const TabsContext = createContext<TabsContextValue | null>(null);

function useTabs() {
  const context = useContext(TabsContext);
  if (!context) {
    throw new Error('Tabs components must be used within <Tabs>');
  }
  return context;
}

interface TabsProps {
  children: React.ReactNode;
  defaultTab: string;
  className?: string;
}

export function Tabs({ children, defaultTab, className }: TabsProps) {
  const [activeTab, setActiveTab] = useState(defaultTab);
  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <div className={className}>{children}</div>
    </TabsContext.Provider>
  );
}

interface TabsListProps {
  children: React.ReactNode;
  className?: string;
  variant?: 'default' | 'pills' | 'segmented';
}

export function TabsList({ children, className, variant = 'default' }: TabsListProps) {
  const variantClasses = {
    default: 'flex gap-1 border-b border-admin-200',
    pills: 'flex gap-1 p-1 bg-admin-100/60 rounded-lg',
    segmented: 'flex p-1 bg-admin-100 rounded-xl',
  };

  return (
    <div
      role="tablist"
      className={cn(variantClasses[variant], className)}
    >
      {children}
    </div>
  );
}

interface TabsTriggerProps {
  value: string;
  children: React.ReactNode;
  icon?: React.ComponentType<{ className?: string }>;
  className?: string;
  variant?: 'default' | 'pills' | 'segmented';
}

export function TabsTrigger({
  value,
  children,
  icon: Icon,
  className,
  variant = 'default',
}: TabsTriggerProps) {
  const { activeTab, setActiveTab } = useTabs();
  const isActive = activeTab === value;

  const variantClasses = {
    default: {
      base: 'flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-all duration-200 rounded-t-lg',
      active: 'border-admin-600 text-admin-900 bg-admin-50/50',
      inactive: 'border-transparent text-admin-500 hover:text-admin-700 hover:border-admin-300 hover:bg-admin-50/30',
    },
    pills: {
      base: 'flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md transition-all duration-200',
      active: 'bg-white text-admin-700 shadow-sm border border-admin-200/50',
      inactive: 'text-admin-600 hover:text-admin-800 hover:bg-admin-200/50',
    },
    segmented: {
      base: 'flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-all duration-200 flex-1 justify-center',
      active: 'bg-white text-admin-700 shadow-sm',
      inactive: 'text-admin-600 hover:text-admin-800',
    },
  };

  const classes = variantClasses[variant];

  return (
    <button
      type="button"
      role="tab"
      aria-selected={isActive}
      aria-controls={`panel-${value}`}
      id={`tab-${value}`}
      onClick={() => setActiveTab(value)}
      className={cn(
        classes.base,
        isActive ? classes.active : classes.inactive,
        className
      )}
    >
      {Icon && <Icon className="w-4 h-4 shrink-0" />}
      {children}
    </button>
  );
}

interface TabsContentProps {
  value: string;
  children: React.ReactNode;
  className?: string;
}

export function TabsContent({ value, children, className }: TabsContentProps) {
  const { activeTab } = useTabs();
  const isActive = activeTab === value;

  if (!isActive) return null;

  return (
    <div
      role="tabpanel"
      id={`panel-${value}`}
      aria-labelledby={`tab-${value}`}
      className={cn('animate-in fade-in slide-in-from-bottom-2 duration-300', className)}
    >
      {children}
    </div>
  );
}

interface TabBadgeProps {
  count: number;
  variant?: 'default' | 'active';
}

export function TabBadge({ count, variant = 'default' }: TabBadgeProps) {
  return (
    <span
      className={cn(
        'ml-1.5 inline-flex items-center justify-center min-w-[1.25rem] px-1.5 py-0.5 text-xs font-semibold rounded-full',
        variant === 'active'
          ? 'bg-admin-100 text-admin-700'
          : 'bg-admin-200/60 text-admin-600'
      )}
    >
      {count}
    </span>
  );
}
