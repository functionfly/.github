import React, { forwardRef } from 'react';
import { Spinner } from './Spinner';

type ButtonSize = 'sm' | 'md' | 'lg';

interface SealedButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  size?: ButtonSize;
  loading?: boolean;
  iconRight?: React.ReactNode;
  iconLeft?: React.ReactNode;
  children: React.ReactNode;
}

/**
 * SealedButton - Primary button with brushed-metal gradient appearance.
 */
export const SealedButton = forwardRef<HTMLButtonElement, SealedButtonProps>(
  (
    {
      size = 'md',
      loading = false,
      iconRight,
      iconLeft,
      children,
      className = '',
      disabled,
      ...props
    },
    ref
  ) => {
    const isDisabled = disabled || loading;

    return (
      <button
        ref={ref}
        disabled={isDisabled}
        aria-disabled={isDisabled}
        aria-busy={loading}
        className={`sealed-button-primary ${size === 'sm' ? 'sealed-btn-sm' : size === 'lg' ? 'sealed-btn-lg' : 'sealed-btn-md'} ${loading ? 'sealed-btn-loading' : ''} ${className}`}
        {...props}
      >
        {loading ? (
          <>
            <Spinner size={size === 'sm' ? 'sm' : 'md'} />
            <span className="sealed-btn-label" aria-hidden="true">Loading</span>
          </>
        ) : (
          <>
            {iconLeft && <span className="sealed-btn-icon sealed-btn-icon-left">{iconLeft}</span>}
            <span className="sealed-btn-label">{children}</span>
            {iconRight && <span className="sealed-btn-icon sealed-btn-icon-right">{iconRight}</span>}
          </>
        )}
      </button>
    );
  }
);

SealedButton.displayName = 'SealedButton';
