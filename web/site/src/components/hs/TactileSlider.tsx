import React, { forwardRef, useCallback, useRef } from 'react';

export interface TactileSliderProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  /** Current value */
  value: number;
  /** Min/max for display scaling (default 0-100) */
  min?: number;
  max?: number;
  /** Called when value changes */
  onValueChange: (next: number) => void;
}

/**
 * TactileSlider - neumorphic range slider. Native <input type="range">
 * is restyled with the tactile shadow language. The track is inset
 * (pressed in), the thumb is raised.
 */
export const TactileSlider = forwardRef<HTMLInputElement, TactileSliderProps>(
  function TactileSlider(
    { value, min = 0, max = 100, onValueChange, className = '', ...rest },
    ref
  ) {
    const internalRef = useRef<HTMLInputElement | null>(null);
    const setRefs = useCallback(
      (el: HTMLInputElement | null) => {
        internalRef.current = el;
        if (typeof ref === 'function') ref(el);
        else if (ref) (ref as React.MutableRefObject<HTMLInputElement | null>).current = el;
      },
      [ref]
    );

    return (
      <input
        ref={setRefs}
        type="range"
        value={value}
        min={min}
        max={max}
        className={`hs-slider ${className}`}
        onChange={(e) => onValueChange(Number(e.target.value))}
        {...rest}
      />
    );
  }
);
