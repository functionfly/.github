import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Loader2 } from 'lucide-react';

export interface RetryButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  loading?: boolean;
  children?: ReactNode;
}

export function RetryButton({
  loading = false,
  disabled,
  children = 'Retry',
  className = '',
  ...props
}: RetryButtonProps) {
  const isDisabled = disabled || loading;

  return (
    <button
      className={`retry-button ${className}`}
      disabled={isDisabled}
      aria-busy={loading}
      {...props}
    >
      {loading ? (
        <Loader2 className="retry-button__spinner" />
      ) : (
        <>
          <svg
            className="retry-button__icon"
            width="14"
            height="14"
            viewBox="0 0 14 14"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M1 7C1 3.68629 3.68629 1 7 1C8.94249 1 10.7053 1.92386 11.8944 3.34927M13 7C13 10.3137 10.3137 13 7 13C5.05751 13 3.29469 12.0761 2.10557 10.6507"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              transform="translate(1, 0)"
            />
            <path
              d="M12 1V4H9"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <span className="retry-button__label">{children}</span>
        </>
      )}
    </button>
  );
}
