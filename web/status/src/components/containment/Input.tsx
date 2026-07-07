import React, { forwardRef } from 'react';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: boolean;
  errorMessage?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ error = false, errorMessage, className = '', ...props }, ref) => {
    return (
      <div className="w-full">
        <input
          ref={ref}
          className={`w-full ${className}`}
          style={{
            fontFamily: 'var(--font-body)',
            fontSize: '15px',
            color: 'var(--text)',
            backgroundColor: 'var(--panel-raised)',
            border: `1px solid ${error ? 'var(--status-revoked)' : 'var(--steel)'}`,
            borderRadius: 'var(--radius)',
            padding: 'var(--space-3) var(--space-4)',
            boxShadow: error ? 'var(--shadow-input-error)' : 'var(--shadow-input-rest)',
            outline: 'none',
            transition: 'border-color var(--duration-fast) var(--ease-out), box-shadow var(--duration-fast) var(--ease-out)',
          }}
          onFocus={(e) => {
            if (!error) {
              e.currentTarget.style.borderColor = 'var(--steel-light)';
              e.currentTarget.style.boxShadow = 'var(--shadow-input-focus)';
            }
          }}
          onBlur={(e) => {
            e.currentTarget.style.borderColor = error ? 'var(--status-revoked)' : 'var(--steel)';
            e.currentTarget.style.boxShadow = error ? 'var(--shadow-input-error)' : 'var(--shadow-input-rest)';
          }}
          {...props}
        />
        {error && errorMessage && (
          <p
            className="mt-1"
            style={{ fontSize: '13px', color: 'var(--status-revoked)' }}
            role="alert"
          >
            {errorMessage}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';
