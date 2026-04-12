'use client';

import * as React from 'react';
import { format, isValid, parse } from 'date-fns';
import { Calendar as CalendarIcon, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from './button';
import { Popover, PopoverContent, PopoverTrigger } from './popover';
import { Calendar } from './calendar';

export interface DatePickerProps {
  value?: Date | null;
  onChange?: (date: Date | null) => void;
  placeholder?: string;
  disabled?: boolean;
  minDate?: Date;
  maxDate?: Date;
  format?: string;
  className?: string;
  showClear?: boolean;
  showToday?: boolean;
  showOutsideDays?: boolean;
  disabledDays?: Date[];
  disabledDayMatcher?: (date: Date) => boolean;
}

const DatePicker = React.forwardRef<HTMLButtonElement, DatePickerProps>(
  (
    {
      value,
      onChange,
      placeholder = 'Pick a date',
      disabled = false,
      minDate,
      maxDate,
      format: formatStr = 'PPP',
      className,
      showClear = true,
      showToday = true,
      showOutsideDays = true,
      disabledDays,
      disabledDayMatcher,
    },
    ref
  ) => {
    const [open, setOpen] = React.useState(false);

    const handleSelect = React.useCallback(
      (date: Date | undefined) => {
        onChange?.(date || null);
        setOpen(false);
      },
      [onChange]
    );

    const handleClear = React.useCallback(
      (e: React.MouseEvent) => {
        e.stopPropagation();
        onChange?.(null);
      },
      [onChange]
    );

    const handleToday = React.useCallback(() => {
      onChange?.(new Date());
      setOpen(false);
    }, [onChange]);

    const displayValue = React.useMemo(() => {
      if (!value || !isValid(value)) return placeholder;
      return format(value, formatStr);
    }, [value, formatStr, placeholder]);

    const disabledMatcher = React.useCallback(
      (date: Date) => {
        if (disabledDayMatcher?.(date)) return true;
        if (disabledDays?.some((d) => d.getTime() === date.getTime())) return true;
        if (minDate && date < minDate) return true;
        if (maxDate && date > maxDate) return true;
        return false;
      },
      [disabledDays, disabledDayMatcher, minDate, maxDate]
    );

    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            ref={ref}
            variant="outline"
            className={cn(
              'w-[280px] justify-start text-left font-normal',
              !value && 'text-text-muted',
              className
            )}
            disabled={disabled}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {displayValue}
            <ChevronDown className="ml-auto h-4 w-4 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={value || undefined}
            onSelect={handleSelect}
            disabled={disabledMatcher}
            showOutsideDays={showOutsideDays}
            initialFocus
          />
          {(showClear || showToday) && (
            <div className="flex items-center justify-between border-t border-border-default p-3">
              {showToday && (
                <Button variant="ghost" size="sm" onClick={handleToday} disabled={disabled}>
                  Today
                </Button>
              )}
              {showClear && value && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleClear}
                  disabled={disabled}
                  className="text-error hover:text-error"
                >
                  Clear
                </Button>
              )}
            </div>
          )}
        </PopoverContent>
      </Popover>
    );
  }
);
DatePicker.displayName = 'DatePicker';

export type DateRangeSelection = { from: Date | null; to: Date | null };

export interface DateRangePickerProps {
  value?: DateRangeSelection;
  onChange?: (range: DateRangeSelection) => void;
  placeholder?: string;
  disabled?: boolean;
  minDate?: Date;
  maxDate?: Date;
  className?: string;
  showClear?: boolean;
  showPresets?: boolean;
  numberOfMonths?: number;
}

const presets = [
  { label: 'Today', days: 0 },
  { label: 'Yesterday', days: -1 },
  { label: 'Last 7 days', days: -7 },
  { label: 'Last 30 days', days: -30 },
  { label: 'This month', type: 'month', offset: 0 },
  { label: 'Last month', type: 'month', offset: -1 },
];

const DateRangePicker = React.forwardRef<HTMLButtonElement, DateRangePickerProps>(
  (
    {
      value,
      onChange,
      placeholder = 'Pick a date range',
      disabled = false,
      minDate,
      maxDate,
      className,
      showClear = true,
      showPresets = true,
      numberOfMonths = 2,
    },
    ref
  ) => {
    const [open, setOpen] = React.useState(false);
    const [tempRange, setTempRange] = React.useState<{ from?: Date; to?: Date }>(
      value?.from && value?.to ? { from: value.from, to: value.to } : {}
    );

    React.useEffect(() => {
      if (value?.from && value?.to) {
        setTempRange({ from: value.from, to: value.to });
      }
    }, [value?.from, value?.to]);

    const displayValue = React.useMemo(() => {
      if (!value?.from || !isValid(value.from)) return placeholder;
      if (!value?.to || !isValid(value.to)) {
        return format(value.from, 'PP');
      }
      return `${format(value.from, 'PP')} - ${format(value.to, 'PP')}`;
    }, [value, placeholder]);

    const handleSelect = React.useCallback(
      (range: { from?: Date; to?: Date }) => {
        setTempRange(range);
        if (range.from && range.to) {
          onChange?.({ from: range.from, to: range.to });
          setOpen(false);
        }
      },
      [onChange]
    );

    const handleClear = React.useCallback(() => {
      setTempRange({});
      onChange?.({ from: null, to: null });
      setOpen(false);
    }, [onChange]);

    const applyPreset = React.useCallback(
      (preset: (typeof presets)[number]) => {
        const today = new Date();
        today.setHours(0, 0, 0, 0);

        let from: Date;
        let to: Date;

        if (preset.type === 'month') {
          const offset = (preset.offset as number) || 0;
          const year = today.getFullYear();
          const month = today.getMonth() + offset;
          from = new Date(year, month, 1);
          to = new Date(year, month + 1, 0);
        } else {
          to = new Date(today);
          from = new Date(today);
          from.setDate(from.getDate() + (preset.days as number));
          if (preset.days === -1) {
            to = new Date(from);
          } else if (preset.days !== 0) {
            from.setDate(from.getDate() - (preset.days as number));
          }
        }

        from.setHours(0, 0, 0, 0);
        to.setHours(23, 59, 59, 999);

        onChange?.({ from, to });
        setOpen(false);
      },
      [onChange]
    );

    const disabledMatcher = React.useCallback(
      (date: Date) => {
        if (minDate && date < minDate) return true;
        if (maxDate && date > maxDate) return true;
        return false;
      },
      [minDate, maxDate]
    );

    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            ref={ref}
            variant="outline"
            className={cn(
              'w-[300px] justify-start text-left font-normal',
              !value?.from && 'text-text-muted',
              className
            )}
            disabled={disabled}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {displayValue}
            <ChevronDown className="ml-auto h-4 w-4 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          {showPresets && (
            <div className="border-b border-border-default p-3">
              <div className="flex flex-wrap gap-2">
                {presets.map((preset) => (
                  <Button
                    key={preset.label}
                    variant="ghost"
                    size="sm"
                    onClick={() => applyPreset(preset)}
                    disabled={disabled}
                  >
                    {preset.label}
                  </Button>
                ))}
              </div>
            </div>
          )}
          <Calendar
            mode="range"
            selected={{
              from: tempRange.from,
              to: tempRange.to,
            }}
            onSelect={handleSelect}
            numberOfMonths={numberOfMonths}
            disabled={disabledMatcher}
            initialFocus
          />
          {showClear && (value?.from || value?.to) && (
            <div className="flex items-center justify-end border-t border-border-default p-3">
              <Button
                variant="ghost"
                size="sm"
                onClick={handleClear}
                disabled={disabled}
                className="text-error hover:text-error"
              >
                Clear
              </Button>
            </div>
          )}
        </PopoverContent>
      </Popover>
    );
  }
);
DateRangePicker.displayName = 'DateRangePicker';

export { DatePicker, DateRangePicker };
