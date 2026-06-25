import React, { forwardRef } from 'react';

export interface TactileToggleProps
  extends Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, 'onChange' | 'children'> {
  /** Whether the toggle is on */
  on: boolean;
  /** Called when user toggles */
  onChange: (next: boolean) => void;
  /** Accessible label (required) */
  'aria-label': string;
}

/**
 * TactileToggle - neumorphic toggle switch.
 * Recessed track, raised thumb. Thumb slides on toggle.
 * The track fills with accent gradient when on.
 */
export const TactileToggle = forwardRef<HTMLButtonElement, TactileToggleProps>(
  function TactileToggle({ on, onChange, className = '', ...rest }, ref) {
    return (
      <button
        ref={ref}
        type="button"
        role="switch"
        aria-checked={on}
        data-on={on}
        className={`hs-toggle ${className}`}
        onClick={() => onChange(!on)}
        {...rest}
      >
        <span className="hs-toggle__thumb" aria-hidden="true" />
      </button>
    );
  }
);
