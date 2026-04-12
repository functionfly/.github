'use client';

import * as React from 'react';
import * as TabsPrimitive from '@radix-ui/react-tabs';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const toggleGroupVariants = cva('flex', {
  variants: {
    variant: {
      default: 'bg-bg-secondary rounded-lg p-1',
      outline: 'border border-border-default rounded-lg p-1 bg-bg-primary',
      ghost: 'gap-1',
    },
    orientation: {
      horizontal: 'flex-row',
      vertical: 'flex-col',
    },
    size: {
      default: '',
      sm: 'p-0.5',
      lg: 'p-1.5',
    },
  },
  defaultVariants: {
    variant: 'default',
    orientation: 'horizontal',
    size: 'default',
  },
});

const toggleGroupItemVariants = cva(
  'inline-flex items-center justify-center rounded-md text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=off]:hover:bg-bg-primary/50',
  {
    variants: {
      variant: {
        default: 'data-[state=on]:bg-bg-primary data-[state=on]:text-text-primary data-[state=on]:shadow-sm text-text-secondary',
        outline: 'border border-transparent data-[state=on]:border-border-default data-[state=on]:bg-bg-secondary data-[state=on]:text-text-primary text-text-secondary',
        ghost: 'data-[state=on]:bg-brand-500/10 data-[state=on]:text-brand-500 text-text-secondary',
      },
      size: {
        default: 'h-9 px-3',
        sm: 'h-7 px-2 text-xs',
        lg: 'h-11 px-4 text-base',
        icon: 'h-9 w-9',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

export interface ToggleButtonOption {
  value: string;
  label?: string;
  icon?: React.ReactNode;
  disabled?: boolean;
  tooltip?: string;
}

export interface ToggleButtonGroupProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof toggleGroupVariants> {
  options: ToggleButtonOption[];
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  itemSize?: 'default' | 'sm' | 'lg' | 'icon';
  itemVariant?: 'default' | 'outline' | 'ghost';
}

const ToggleButtonGroup = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Root>,
  ToggleButtonGroupProps
>(
  (
    {
      className,
      variant,
      orientation,
      size,
      options,
      value,
      defaultValue,
      onValueChange,
      itemSize,
      itemVariant,
    },
    ref
  ) => {
    return (
      <div className={cn(toggleGroupVariants({ variant, orientation, size, className }))}>
        <TabsPrimitive.Root
          ref={ref}
          value={value}
          defaultValue={defaultValue}
          onValueChange={onValueChange}
        >
          <TabsPrimitive.List className="flex">
            {options.map((option) => (
              <TabsPrimitive.Trigger
                key={option.value}
                value={option.value}
                disabled={option.disabled}
                title={option.tooltip}
              className={cn(toggleGroupItemVariants({ variant: itemVariant || variant, size: itemSize }))}
            >
              {option.icon && <span className={cn(option.label && 'mr-2')}>{option.icon}</span>}
              {option.label}
            </TabsPrimitive.Trigger>
          ))}
          </TabsPrimitive.List>
        </TabsPrimitive.Root>
      </div>
    );
    }
  );
  ToggleButtonGroup.displayName = 'ToggleButtonGroup';

export { ToggleButtonGroup, toggleGroupVariants, toggleGroupItemVariants };
