import React, { forwardRef, useState, useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { ChevronDown, Check } from 'lucide-react';

interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

interface SelectProps extends Omit<React.SelectHTMLAttributes<HTMLSelectElement>, 'options'> {
  label?: string;
  helperText?: string;
  error?: string;
  options: SelectOption[];
  placeholder?: string;
  fullWidth?: boolean;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  (
    {
      label,
      helperText,
      error,
      options,
      placeholder,
      fullWidth = true,
      className,
      disabled,
      value,
      onChange,
      ...props
    },
    ref
  ) => {
    const [isOpen, setIsOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    const selectedOption = options.find((o) => o.value === value);

    useEffect(() => {
      const handleClickOutside = (e: MouseEvent) => {
        if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
          setIsOpen(false);
        }
      };
      if (isOpen) {
        document.addEventListener('mousedown', handleClickOutside);
      }
      return () => {
        document.removeEventListener('mousedown', handleClickOutside);
      };
    }, [isOpen]);

    const handleSelect = (optionValue: string) => {
      onChange?.({ target: { value: optionValue } } as any);
      setIsOpen(false);
    };

    return (
      <div className={cn('flex flex-col gap-1.5', fullWidth && 'w-full')} ref={containerRef}>
        {label && (
          <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
            {label}
            {props.required && <span className="ml-1 text-red-500">*</span>}
          </label>
        )}
        <div className="relative">
          <button
            type="button"
            onClick={() => !disabled && setIsOpen(!isOpen)}
            disabled={disabled}
            className={cn(
              'w-full flex items-center justify-between px-3 py-2 border rounded-lg bg-white dark:bg-gray-800 appearance-none',
              'focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500',
              'disabled:bg-gray-100 dark:disabled:bg-gray-700 disabled:text-gray-500 disabled:cursor-not-allowed',
              'transition-colors duration-150 pr-10 text-left',
              error ? 'border-red-500 focus:ring-red-500 focus:border-red-500' : 'border-gray-300 dark:border-gray-600',
              className
            )}
          >
            <span className={cn(
              selectedOption ? 'text-gray-900 dark:text-gray-100' : 'text-gray-400 dark:text-gray-500'
            )}>
              {selectedOption?.label || placeholder || 'Select...'}
            </span>
            <ChevronDown className={cn(
              'w-4 h-4 text-gray-400 transition-transform duration-150',
              isOpen && 'rotate-180'
            )} />
          </button>

          {isOpen && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg z-50 py-1 max-h-60 overflow-auto">
              {options.map((option) => {
                const isSelected = option.value === value;
                return (
                  <button
                    key={option.value}
                    type="button"
                    disabled={option.disabled}
                    onClick={() => !option.disabled && handleSelect(option.value)}
                    className={cn(
                      'w-full flex items-center justify-between px-3 py-2 text-left text-sm transition-colors duration-150',
                      'hover:bg-gray-50 dark:hover:bg-gray-700',
                      option.disabled && 'opacity-50 cursor-not-allowed',
                      isSelected ? 'text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30' : 'text-gray-700 dark:text-gray-200'
                    )}
                  >
                    <span>{option.label}</span>
                    {isSelected && <Check className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />}
                  </button>
                );
              })}
            </div>
          )}
        </div>

        <select
          ref={ref}
          value={value}
          onChange={onChange}
          disabled={disabled}
          className="sr-only"
          {...props}
        >
          {placeholder && (
            <option value="" disabled>
              {placeholder}
            </option>
          )}
          {options.map((option) => (
            <option key={option.value} value={option.value} disabled={option.disabled}>
              {option.label}
            </option>
          ))}
        </select>

        {helperText && !error && (
          <p className="text-xs text-gray-500 dark:text-gray-400">{helperText}</p>
        )}
        {error && (
          <p className="text-xs text-red-600 dark:text-red-400">{error}</p>
        )}
      </div>
    );
  }
);

Select.displayName = 'Select';

interface MultiSelectProps {
  label?: string;
  helperText?: string;
  error?: string;
  options: SelectOption[];
  selected: string[];
  onChange: (selected: string[]) => void;
  placeholder?: string;
  fullWidth?: boolean;
}

export function MultiSelect({
  label,
  helperText,
  error,
  options,
  selected,
  onChange,
  placeholder = 'Select options...',
  fullWidth = true,
}: MultiSelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const toggleOption = (value: string) => {
    if (selected.includes(value)) {
      onChange(selected.filter((v) => v !== value));
    } else {
      onChange([...selected, value]);
    }
  };

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  return (
    <div className={cn('flex flex-col gap-1.5', fullWidth && 'w-full')} ref={containerRef}>
      {label && (
        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">{label}</label>
      )}
      <div className="relative">
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className={cn(
            'w-full flex items-center justify-between px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800',
            'focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500',
            'transition-colors duration-150 text-left',
            error && 'border-red-500 focus:ring-red-500 focus:border-red-500'
          )}
        >
          <span className={selected.length === 0 ? 'text-gray-400 dark:text-gray-500' : 'text-gray-900 dark:text-gray-100'}>
            {selected.length === 0 ? placeholder : `${selected.length} selected`}
          </span>
          <ChevronDown className={cn('w-4 h-4 text-gray-400', isOpen && 'rotate-180')} />
        </button>

        {isOpen && (
          <div className="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg z-50 py-1 max-h-48 overflow-auto">
            {options.map((option) => {
              const isSelected = selected.includes(option.value);
              return (
                <label
                  key={option.value}
                  className={cn(
                    'flex items-center gap-2 px-3 py-2 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700',
                    option.disabled && 'opacity-50 cursor-not-allowed'
                  )}
                >
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={() => !option.disabled && toggleOption(option.value)}
                    disabled={option.disabled}
                    className="w-4 h-4 text-indigo-600 border-gray-300 dark:border-gray-600 rounded focus:ring-indigo-500 dark:bg-gray-700"
                  />
                  <span className="text-sm text-gray-700 dark:text-gray-200">{option.label}</span>
                </label>
              );
            })}
          </div>
        )}
      </div>
      {helperText && !error && (
        <p className="text-xs text-gray-500 dark:text-gray-400">{helperText}</p>
      )}
      {error && (
        <p className="text-xs text-red-600 dark:text-red-400">{error}</p>
      )}
    </div>
  );
}