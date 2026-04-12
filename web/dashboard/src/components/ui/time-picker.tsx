'use client';

import * as React from 'react';
import { Clock, ChevronUp, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from './button';
import { Popover, PopoverContent, PopoverTrigger } from './popover';

export interface TimePickerProps {
  value?: string;
  onChange?: (time: string) => void;
  placeholder?: string;
  disabled?: boolean;
  minTime?: string;
  maxTime?: string;
  className?: string;
  showSeconds?: boolean;
  minuteStep?: number;
  hourStep?: number;
  clearable?: boolean;
}

const formatTime = (hours: number, minutes: number, seconds: number = 0, showSeconds: boolean = false) => {
  const pad = (n: number) => n.toString().padStart(2, '0');
  return `${pad(hours)}:${pad(minutes)}${showSeconds ? `:${pad(seconds)}` : ''}`;
};

const parseTime = (time: string) => {
  const parts = time.split(':');
  return {
    hours: parseInt(parts[0] || '0', 10),
    minutes: parseInt(parts[1] || '0', 10),
    seconds: parseInt(parts[2] || '0', 10),
  };
};

const TimePicker = React.forwardRef<HTMLButtonElement, TimePickerProps>(
  (
    {
      value,
      onChange,
      placeholder = 'Pick a time',
      disabled = false,
      minTime,
      maxTime,
      className,
      showSeconds = false,
      minuteStep = 1,
      hourStep = 1,
      clearable = true,
    },
    ref
  ) => {
    const [open, setOpen] = React.useState(false);
    const { hours, minutes, seconds } = React.useMemo(() => {
      return value ? parseTime(value) : { hours: 12, minutes: 0, seconds: 0 };
    }, [value]);

    const isTimeValid = React.useCallback(
      (h: number, m: number, s: number = 0) => {
        const timeStr = formatTime(h, m, s, showSeconds);
        if (minTime && timeStr < minTime) return false;
        if (maxTime && timeStr > maxTime) return false;
        return true;
      },
      [minTime, maxTime, showSeconds]
    );

    const handleHourChange = React.useCallback(
      (delta: number) => {
        const newHours = ((hours + delta + 24) % 24);
        if (isTimeValid(newHours, minutes, seconds)) {
          onChange?.(formatTime(newHours, minutes, seconds, showSeconds));
        }
      },
      [hours, minutes, seconds, isTimeValid, onChange, showSeconds]
    );

    const handleMinuteChange = React.useCallback(
      (delta: number) => {
        const newMinutes = ((minutes + delta + 60) % 60);
        if (isTimeValid(hours, newMinutes, seconds)) {
          onChange?.(formatTime(hours, newMinutes, seconds, showSeconds));
        }
      },
      [hours, minutes, seconds, isTimeValid, onChange, showSeconds]
    );

    const handleSecondChange = React.useCallback(
      (delta: number) => {
        const newSeconds = ((seconds + delta + 60) % 60);
        if (isTimeValid(hours, minutes, newSeconds)) {
          onChange?.(formatTime(hours, minutes, newSeconds, showSeconds));
        }
      },
      [hours, minutes, seconds, isTimeValid, onChange, showSeconds]
    );

    const handleClear = React.useCallback(() => {
      onChange?.('');
      setOpen(false);
    }, [onChange]);

    const displayValue = value || placeholder;

    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            ref={ref}
            variant="outline"
            className={cn(
              'w-[160px] justify-start text-left font-normal',
              !value && 'text-text-muted',
              className
            )}
            disabled={disabled}
          >
            <Clock className="mr-2 h-4 w-4" />
            {displayValue}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-3" align="start">
          <div className="flex items-center justify-center gap-4">
            {/* Hours */}
            <div className="flex flex-col items-center">
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={() => handleHourChange(hourStep)}
                disabled={disabled}
              >
                <ChevronUp className="h-4 w-4" />
              </Button>
              <div className="flex h-10 w-12 items-center justify-center rounded-md bg-bg-secondary text-lg font-semibold">
                {hours.toString().padStart(2, '0')}
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={() => handleHourChange(-hourStep)}
                disabled={disabled}
              >
                <ChevronDown className="h-4 w-4" />
              </Button>
            </div>

            <span className="text-xl font-bold text-text-muted">:</span>

            {/* Minutes */}
            <div className="flex flex-col items-center">
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={() => handleMinuteChange(minuteStep)}
                disabled={disabled}
              >
                <ChevronUp className="h-4 w-4" />
              </Button>
              <div className="flex h-10 w-12 items-center justify-center rounded-md bg-bg-secondary text-lg font-semibold">
                {minutes.toString().padStart(2, '0')}
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={() => handleMinuteChange(-minuteStep)}
                disabled={disabled}
              >
                <ChevronDown className="h-4 w-4" />
              </Button>
            </div>

            {showSeconds && (
              <>
                <span className="text-xl font-bold text-text-muted">:</span>
                {/* Seconds */}
                <div className="flex flex-col items-center">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => handleSecondChange(1)}
                    disabled={disabled}
                  >
                    <ChevronUp className="h-4 w-4" />
                  </Button>
                  <div className="flex h-10 w-12 items-center justify-center rounded-md bg-bg-secondary text-lg font-semibold">
                    {seconds.toString().padStart(2, '0')}
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => handleSecondChange(-1)}
                    disabled={disabled}
                  >
                    <ChevronDown className="h-4 w-4" />
                  </Button>
                </div>
              </>
            )}
          </div>

          {(minTime || maxTime) && (
            <div className="mt-2 text-center text-xs text-text-muted">
              {minTime && `Min: ${minTime}`}
              {minTime && maxTime && ' | '}
              {maxTime && `Max: ${maxTime}`}
            </div>
          )}

          {clearable && value && (
            <div className="mt-3 border-t border-border-default pt-2">
              <Button
                variant="ghost"
                size="sm"
                className="w-full text-error hover:text-error"
                onClick={handleClear}
                disabled={disabled}
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
TimePicker.displayName = 'TimePicker';

export { TimePicker };
