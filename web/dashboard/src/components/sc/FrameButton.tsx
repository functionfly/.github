import React, { forwardRef } from 'react';
import { Spinner } from './Spinner';

type ButtonSize = 'sm' | 'md' | 'lg';

interface FrameButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  size?: ButtonSize;
  loading?: boolean;
  iconRight?: React.ReactNode;
  iconLeft?: React.ReactNode;
  children: React.ReactNode;
}

/**
 * FrameButton - Secondary/outline button with steel border styling.
 * Paired with SealedButton for primary actions.
 */
export const FrameButton = forwardRef<HTMLButtonElement, FrameButtonProps>(
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
    const sizeClasses = {
      sm: 'frame-btn-sm',
      md: 'frame-btn-md',
      lg: 'frame-btn-lg',
    };

    return (
      <button
        ref={ref}
        disabled={isDisabled}
        aria-disabled={isDisabled}
        aria-busy={loading}
        className={`frame-button-secondary ${sizeClasses[size]} ${loading ? 'frame-btn-loading' : ''} ${className}`}
        {...props}
      >
        {loading ? (
          <>
            <Spinner size={size === 'sm' ? 'sm' : 'md'} />
            <span className="frame-btn-label" aria-hidden="true">Loading</span>
          </>
        ) : (
          <>
            {iconLeft && <span className="frame-btn-icon frame-btn-icon-left">{iconLeft}</span>}
            <span className="frame-btn-label">{children}</span>
            {iconRight && <span className="frame-btn-icon frame-btn-icon-right">{iconRight}</span>}
          </>
        )}
      </button>
    );
  }
);

FrameButton.displayName = 'FrameButton';
