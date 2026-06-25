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
          className={`
            sealed-input
            w-full
            px-[var(--space-4)]
            py-[var(--space-3)]
            text-[var(--text)]
            text-[15px]
            bg-[var(--panel-raised)]
            border
            border-[var(--steel)]
            rounded-[var(--radius)]
            font-[var(--font-body)]
            placeholder:text-[var(--text-faint)]
            transition-shadow duration-[var(--duration-fast)] ease-[var(--ease-out)]
            focus:outline-none
            focus:border-[var(--steel-light)]
            focus:shadow-[var(--shadow-input-focus)]
            disabled:opacity-50
            disabled:cursor-not-allowed
            ${error ? 'border-[var(--status-revoked)] shadow-[var(--shadow-input-error)]' : ''}
            ${className}
          `}
          {...props}
        />
        {error && errorMessage && (
          <p
            className="mt-[var(--space-1)] text-[13px] text-[var(--status-revoked)]"
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
