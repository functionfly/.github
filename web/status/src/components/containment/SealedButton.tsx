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
          fontWeight: 600,
          fontSize: size === 'sm' ? '12px' : size === 'lg' ? '15px' : '14px',
          lineHeight: 1,
          color: 'var(--text-on-light)',
          background: 'linear-gradient(180deg, #ffffff 0%, #d8dee2 100%)',
          border: 'none',
          borderRadius: 'var(--radius)',
          boxShadow: isDisabled
            ? 'var(--shadow-btn-primary-disabled)'
            : 'var(--shadow-btn-primary-rest)',
          cursor: isDisabled ? 'not-allowed' : 'pointer',
          opacity: isDisabled ? 0.4 : 1,
          padding,
          transition: 'box-shadow var(--duration-fast) var(--ease-out), transform var(--duration-fast) var(--ease-out)',
          whiteSpace: 'nowrap',
          textDecoration: 'none',
          WebkitFontSmoothing: 'antialiased',
        }}
        onMouseEnter={(e) => {
          if (!isDisabled) {
            (e.currentTarget as HTMLElement).style.boxShadow = 'var(--shadow-btn-primary-hover)';
          }
        }}
        onMouseLeave={(e) => {
          if (!isDisabled) {
            (e.currentTarget as HTMLElement).style.boxShadow = 'var(--shadow-btn-primary-rest)';
          }
        }}
        onMouseDown={(e) => {
          if (!isDisabled) {
            (e.currentTarget as HTMLElement).style.boxShadow = 'var(--shadow-btn-primary-active)';
            (e.currentTarget as HTMLElement).style.transform = 'translateY(1px)';
          }
        }}
        onMouseUp={(e) => {
          if (!isDisabled) {
            (e.currentTarget as HTMLElement).style.boxShadow = 'var(--shadow-btn-primary-rest)';
            (e.currentTarget as HTMLElement).style.transform = '';
          }
        }}
        {...props}
      >
        {loading ? (
          <>
            <Spinner size={size === 'sm' ? 'sm' : 'md'} color="var(--status-ok)" />
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

SealedButton.displayName = 'SealedButton';
