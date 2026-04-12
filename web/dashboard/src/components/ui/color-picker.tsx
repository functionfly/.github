'use client';

import * as React from 'react';
import { Palette, Check, Copy, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from './button';
import { Popover, PopoverContent, PopoverTrigger } from './popover';
import { Input } from './input';

export interface ColorPickerProps {
  value?: string;
  onChange?: (color: string) => void;
  disabled?: boolean;
  className?: string;
  showPresets?: boolean;
  showInput?: boolean;
  presets?: string[];
  label?: string;
}

const defaultPresets = [
  '#ef4444', // red-500
  '#f97316', // orange-500
  '#f59e0b', // amber-500
  '#84cc16', // lime-500
  '#22c55e', // green-500
  '#10b981', // emerald-500
  '#06b6d4', // cyan-500
  '#0ea5e9', // sky-500
  '#3b82f6', // blue-500
  '#6366f1', // indigo-500 (brand)
  '#8b5cf6', // violet-500
  '#a855f7', // purple-500
  '#d946ef', // fuchsia-500
  '#ec4899', // pink-500
  '#f43f5e', // rose-500
  '#64748b', // slate-500
];

const isValidHex = (color: string) => /^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$/.test(color);
const normalizeHex = (color: string) => {
  if (!color.startsWith('#')) return `#${color}`;
  return color;
};

const hexToRgb = (hex: string) => {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : null;
};

const ColorPicker = React.forwardRef<HTMLButtonElement, ColorPickerProps>(
  (
    {
      value = '#6366f1',
      onChange,
      disabled = false,
      className,
      showPresets = true,
      showInput = true,
      presets = defaultPresets,
      label,
    },
    ref
  ) => {
    const [open, setOpen] = React.useState(false);
    const [inputValue, setInputValue] = React.useState(value);
    const [copied, setCopied] = React.useState(false);

    const normalizedValue = normalizeHex(value);

    React.useEffect(() => {
      setInputValue(normalizedValue);
    }, [normalizedValue]);

    const handleColorChange = React.useCallback(
      (color: string) => {
        const normalized = normalizeHex(color);
        if (isValidHex(normalized)) {
          onChange?.(normalized);
          setInputValue(normalized);
        }
      },
      [onChange]
    );

    const handleInputChange = React.useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const val = e.target.value;
        setInputValue(val);
        const normalized = normalizeHex(val);
        if (isValidHex(normalized)) {
          onChange?.(normalized);
        }
      },
      [onChange]
    );

    const handleCopy = React.useCallback(async () => {
      try {
        await navigator.clipboard.writeText(normalizedValue);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch {
        // Ignore copy errors
      }
    }, [normalizedValue]);

    const handleRandom = React.useCallback(() => {
      const randomColor = `#${Math.floor(Math.random() * 16777215).toString(16).padStart(6, '0')}`;
      handleColorChange(randomColor);
    }, [handleColorChange]);

    const rgb = hexToRgb(normalizedValue);

    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            ref={ref}
            variant="outline"
            className={cn(
              'w-[140px] justify-between text-left font-normal',
              !value && 'text-text-muted',
              className
            )}
            disabled={disabled}
          >
            <div className="flex items-center gap-2">
              <div
                className="h-5 w-5 rounded border border-border-subtle"
                style={{ backgroundColor: normalizedValue }}
              />
              <span className="uppercase">{normalizedValue.slice(1)}</span>
            </div>
            <Palette className="h-4 w-4 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[280px] p-3" align="start">
          <div className="space-y-3">
            {/* Color Preview */}
            <div className="flex items-center gap-3">
              <div
                className="h-12 w-12 rounded-lg border border-border-subtle shadow-inner"
                style={{ backgroundColor: normalizedValue }}
              />
              <div className="flex-1 space-y-1">
                <div className="text-sm font-medium">{normalizedValue.toUpperCase()}</div>
                {rgb && (
                  <div className="text-xs text-text-muted">
                    RGB: {rgb.r}, {rgb.g}, {rgb.b}
                  </div>
                )}
              </div>
              <div className="flex gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={handleCopy}
                  title="Copy to clipboard"
                >
                  {copied ? <Check className="h-4 w-4 text-success" /> : <Copy className="h-4 w-4" />}
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={handleRandom}
                  title="Random color"
                >
                  <RefreshCw className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Saturation/Brightness Picker */}
            <div className="relative h-32 w-full overflow-hidden rounded-md">
              <div
                className="absolute inset-0"
                style={{
                  background: `linear-gradient(to top, #000, transparent), linear-gradient(to right, #fff, ${normalizedValue})`,
                }}
              />
            </div>

            {/* Hue Slider */}
            <div className="relative h-4 w-full overflow-hidden rounded-full">
              <div
                className="absolute inset-0"
                style={{
                  background:
                    'linear-gradient(to right, #f00 0%, #ff0 17%, #0f0 33%, #0ff 50%, #00f 67%, #f0f 83%, #f00 100%)',
                }}
              />
              <div
                className="absolute top-0 h-full w-2 -translate-x-1/2 rounded-full border-2 border-white shadow"
                style={{ left: '50%' }}
              />
            </div>

            {/* Manual Input */}
            {showInput && (
              <div className="flex items-center gap-2">
                <span className="text-sm text-text-muted">#</span>
                <Input
                  value={inputValue.replace('#', '')}
                  onChange={handleInputChange}
                  className="h-8 flex-1 font-mono uppercase"
                  maxLength={6}
                />
              </div>
            )}

            {/* Presets */}
            {showPresets && (
              <>
                <div className="border-t border-border-default pt-2">
                  <div className="mb-2 text-xs font-medium text-text-muted">Presets</div>
                  <div className="grid grid-cols-8 gap-1">
                    {presets.map((color) => (
                      <button
                        key={color}
                        className={cn(
                          'h-6 w-6 rounded border border-border-subtle transition-all hover:scale-110 focus:outline-none focus:ring-2 focus:ring-brand-500',
                          normalizedValue === color && 'ring-2 ring-brand-500'
                        )}
                        style={{ backgroundColor: color }}
                        onClick={() => handleColorChange(color)}
                        title={color}
                      />
                    ))}
                  </div>
                </div>
              </>
            )}
          </div>
        </PopoverContent>
      </Popover>
    );
  }
);
ColorPicker.displayName = 'ColorPicker';

export { ColorPicker };
