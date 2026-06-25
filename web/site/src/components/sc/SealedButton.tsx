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
 * Reads as a physical embossed metal button with inset highlight and drop shadow.
 * 
 * Visual spec:
 * - Background: linear-gradient(180deg, #ffffff, #d8dee2)
 * - Text color: --bg (#0a0d11)
 * - Shadow: inset highlight (0 1px 0 rgba(255,255,255,0.4)) + drop shadow (0 6px 16px rgba(0,0,0,0.35))
 * - Font: Inter, 600 weight, 14px
 * - Padding: 13px vertical, 22px horizontal
 * - Border-radius: 4px (--radius token)
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
        className={`
          sealed-button-primary
          ${size === 'sm' ? 'sealed-btn-sm' : size === 'lg' ? 'sealed-btn-lg' : 'sealed-btn-md'}
          ${loading ? 'sealed-btn-loading' : ''}
          ${className}
        `}
        {...props}
      >
        {loading ? (
          <>
            <Spinner size={size === 'sm' ? 'sm' : 'md'} color="var(--status-ok)" />
            <span className="sealed-btn-label" aria-hidden="true">
              Loading
            </span>
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
