import { cn } from '@/lib/utils';
import * as React from 'react';

export interface SwitchProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  onCheckedChange?: (checked: boolean) => void;
  checked?: boolean;
}

const Switch = React.forwardRef<HTMLInputElement, SwitchProps>(
  ({ className, onCheckedChange, checked, ...props }, ref) => {
    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
      onCheckedChange?.(event.target.checked);
    };

    return (
      <label className="inline-flex items-center cursor-pointer">
        <input
          type="checkbox"
          className="sr-only"
          ref={ref}
          checked={checked}
          onChange={handleChange}
          {...props}
        />
        <div
          role="switch"
          aria-checked={checked}
          className={cn(
            'relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors',
            checked ? 'bg-brand-500' : 'bg-text-muted',
            className
          )}
          style={
            {
              backgroundColor: checked ? 'var(--status-ok, #22c55e)' : 'var(--text-muted, #6b7280)',
            } as React.CSSProperties
          }
        >
          <span
            className={cn(
              'inline-block h-4 w-4 shrink-0 transform rounded-full bg-white shadow-sm transition-transform',
              checked ? 'translate-x-6' : 'translate-x-1'
            )}
          />
        </div>
      </label>
    );
  }
);

Switch.displayName = 'Switch';

export { Switch };
