import React, { forwardRef } from 'react';

export interface TactileInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  /** Forwarded className */
  className?: string;
}

/**
 * TactileInput - neumorphic text input. Inset (pressed in) shadow
 * treatment. Focus state adds a 2px accent ring (not shadow).
 */
export const TactileInput = forwardRef<HTMLInputElement, TactileInputProps>(
  function TactileInput({ className = '', ...rest }, ref) {
    return <input ref={ref} className={`hs-input ${className}`} {...rest} />;
  }
);
