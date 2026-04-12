import React from 'react';
import { cn } from '@/lib/utils';

interface ToggleButtonGroupProps<T extends string> {
  options: {
    value: T;
    label: string;
    icon?: React.ReactNode;
    disabled?: boolean;
  }[];
  value: T;
  onChange: (value: T) => void;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'outline' | 'pills';
  fullWidth?: boolean;
}

export function ToggleButtonGroup<T extends string>({
  options,
  value,
  onChange,
  className,
  size = 'md',
  variant = 'default',
  fullWidth = false,
}: ToggleButtonGroupProps<T>) {
  const sizeClasses = {
    sm: 'h-8 text-xs',
    md: 'h-10 text-sm',
    lg: 'h-12 text-base',
  };

  const variantClasses = {
    default: {
      group: 'bg-gray-100 p-1 rounded-lg',
      button: 'rounded-md transition-all',
      active: 'bg-white text-gray-900 shadow-sm',
      inactive: 'text-gray-600 hover:text-gray-900 hover:bg-gray-200/50',
    },
    outline: {
      group: 'bg-white border border-gray-200 rounded-lg',
      button: 'border-r border-gray-200 last:border-r-0 transition-colors',
      active: 'bg-indigo-50 text-indigo-700 border-indigo-200',
      inactive: 'text-gray-600 hover:bg-gray-50',
    },
    pills: {
      group: 'gap-2',
      button: 'rounded-full border transition-all',
      active: 'bg-indigo-600 text-white border-indigo-600',
      inactive: 'bg-white text-gray-600 border-gray-200 hover:border-gray-300',
    },
  };

  const v = variantClasses[variant];

  return (
    <div
      className={cn(
        'inline-flex',
        v.group,
        fullWidth && 'w-full',
        className
      )}
      role="group"
      aria-label="Toggle options"
    >
      {options.map((option) => {
        const isActive = value === option.value;
        return (
          <button
            key={option.value}
            type="button"
            disabled={option.disabled}
            onClick={() => onChange(option.value)}
            className={cn(
              'inline-flex items-center justify-center gap-2 font-medium focus:outline-none focus:ring-2 focus:ring-indigo-500/50 disabled:opacity-50 disabled:cursor-not-allowed',
              sizeClasses[size],
              v.button,
              isActive ? v.active : v.inactive,
              fullWidth && 'flex-1',
              size === 'sm' ? 'px-2.5' : size === 'lg' ? 'px-5' : 'px-4'
            )}
            aria-pressed={isActive}
          >
            {option.icon && <span className="flex-shrink-0">{option.icon}</span>}
            <span>{option.label}</span>
          </button>
        );
      })}
    </div>
  );
}

// Preset view toggles for common patterns
interface ViewToggleProps {
  value: 'list' | 'grid' | 'calendar' | 'kanban';
  onChange: (value: 'list' | 'grid' | 'calendar' | 'kanban') => void;
  availableViews?: ('list' | 'grid' | 'calendar' | 'kanban')[];
  className?: string;
}

import { List, LayoutGrid, Calendar, Kanban } from 'lucide-react';

export function ViewToggle({
  value,
  onChange,
  availableViews = ['list', 'grid', 'calendar'],
  className,
}: ViewToggleProps) {
  const viewOptions: { value: 'list' | 'grid' | 'calendar' | 'kanban'; label: string; icon: React.ReactNode }[] = [
    { value: 'list', label: 'List', icon: <List className="w-4 h-4" /> },
    { value: 'grid', label: 'Grid', icon: <LayoutGrid className="w-4 h-4" /> },
    { value: 'calendar', label: 'Calendar', icon: <Calendar className="w-4 h-4" /> },
    { value: 'kanban', label: 'Kanban', icon: <Kanban className="w-4 h-4" /> },
  ];

  const options = viewOptions.filter((v) => availableViews.includes(v.value));

  return (
    <ToggleButtonGroup
      options={options}
      value={value}
      onChange={onChange}
      variant="outline"
      size="sm"
      className={className}
    />
  );
}
