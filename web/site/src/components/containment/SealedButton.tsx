import type { ReactNode, ButtonHTMLAttributes } from 'react';

interface SealedButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  loading?: boolean;
  size?: 'sm' | 'md' | 'lg';
  iconRight?: React.ReactNode;
  iconLeft?: React.ReactNode;
}

export function SealedButton({
  children,
  loading = false,
  disabled,
  size = 'md',
  iconRight,
  iconLeft,
  className = '',
  ...props
}: SealedButtonProps) {
  const isDisabled = disabled || loading;

  const sizeClass = size === 'sm' ? 'sealed-btn-sm' : size === 'lg' ? 'sealed-btn-lg' : 'sealed-btn-md';

  return (
    <button
      {...props}
      disabled={isDisabled}
      aria-disabled={isDisabled}
      aria-busy={loading}
      className={`sealed-button-primary ${sizeClass} ${loading ? 'sealed-btn-loading' : ''} ${className}`}
    >
      {loading ? (
        <>
          <span className="trust-seal-foil" style={{ width: 14, height: 14, borderRadius: '50%', display: 'inline-block' }} />
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
