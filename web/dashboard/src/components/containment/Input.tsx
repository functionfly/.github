import type { InputHTMLAttributes } from 'react';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  error?: boolean;
  errorMessage?: string;
}

export function Input({
  error,
  errorMessage,
  className = '',
  disabled,
  ...props
}: InputProps) {
  const baseClass = error ? 'input input--error' : 'input';

  return (
    <div className="input-wrapper">
      <input
        className={`${baseClass} ${className}`}
        disabled={disabled}
        aria-invalid={error}
        {...props}
      />
      {error && errorMessage && (
        <span className="input-error-message" role="alert">
          {errorMessage}
        </span>
      )}
    </div>
  );
}
