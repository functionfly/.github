import type { InputHTMLAttributes } from 'react';
import { useState } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  error?: string;
}

export function Input({ error, disabled, style, className = '', ...props }: InputProps) {
  const [isFocused, setIsFocused] = useState(false);

  return (
    <div>
      <input
        {...props}
        disabled={disabled}
        className={className}
        style={{
          background: 'var(--panel-raised)',
          color: 'var(--text)',
          fontFamily: 'var(--font-body)',
          fontSize: '15px',
          padding: 'var(--space-3) var(--space-4)',
          borderRadius: 'var(--radius)',
          border: '1px solid var(--steel)',
          boxShadow: error ? 'var(--shadow-input-error)' : isFocused ? 'var(--shadow-input-focus)' : 'var(--shadow-input-rest)',
          outline: 'none',
          opacity: disabled ? 0.5 : 1,
          cursor: disabled ? 'not-allowed' : 'text',
          transition: 'border-color var(--duration-fast) var(--ease-out), box-shadow var(--duration-fast) var(--ease-out)',
          width: '100%',
          ...style,
        }}
        onFocus={(e) => {
          setIsFocused(true);
          props.onFocus?.(e);
        }}
        onBlur={(e) => {
          setIsFocused(false);
          props.onBlur?.(e);
        }}
      />
      {error && (
        <p style={{
          color: 'var(--status-revoked)',
          fontSize: '13px',
          marginTop: 'var(--space-1)',
          fontFamily: 'var(--font-body)',
        }}>
          {error}
        </p>
      )}
    </div>
  );
}
