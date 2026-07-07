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
    const padding = size === 'sm' ? '9px 14px' : size === 'lg' ? '15px 26px' : '13px 22px';

    return (
      <button
        ref={ref}
        disabled={isDisabled}
        aria-disabled={isDisabled}
        aria-busy={loading}
        className={`inline-flex items-center justify-center ${className}`}
        style={{
          fontFamily: 'var(--font-body)',
          fontWeight: 500,
          fontSize: size === 'sm' ? '12px' : size === 'lg' ? '15px' : '14px',
          lineHeight: 1,
          color: 'var(--text)',
          background: 'transparent',
          border: '1px solid var(--steel)',
          borderRadius: 'var(--radius)',
          boxShadow: 'var(--shadow-btn-secondary-rest)',
          cursor: isDisabled ? 'not-allowed' : 'pointer',
          opacity: isDisabled ? 0.4 : 1,
          padding,
          transition: 'border-color var(--duration-fast) var(--ease-out), box-shadow var(--duration-fast) var(--ease-out)',
          whiteSpace: 'nowrap',
          textDecoration: 'none',
          WebkitFontSmoothing: 'antialiased',
        }}
        onMouseEnter={(e) => {
          if (!isDisabled) {
            (e.currentTarget as HTMLElement).style.borderColor = 'var(--steel-light)';
          }
        }}
        onMouseLeave={(e) => {
          if (!isDisabled) {
            (e.currentTarget as HTMLElement).style.borderColor = 'var(--steel)';
          }
        }}
        onMouseDown={(e) => {
          if (!isDisabled) {
            (e.currentTarget as HTMLElement).style.boxShadow = 'var(--shadow-btn-secondary-active)';
          }
        }}
        onMouseUp={(e) => {
          if (!isDisabled) {
            (e.currentTarget as HTMLElement).style.boxShadow = 'var(--shadow-btn-secondary-rest)';
          }
        }}
        {...props}
      >
        {loading ? (
          <>
            <Spinner size={size === 'sm' ? 'sm' : 'md'} />
            <span style={{ marginLeft: '8px' }} aria-hidden="true">Loading</span>
          </>
        ) : (
          <>
            {iconLeft && <span className="inline-flex items-center" style={{ marginRight: '6px' }}>{iconLeft}</span>}
            <span>{children}</span>
            {iconRight && <span className="inline-flex items-center" style={{ marginLeft: '6px' }}>{iconRight}</span>}
          </>
        )}
      </button>
    );
  }
);

FrameButton.displayName = 'FrameButton';
