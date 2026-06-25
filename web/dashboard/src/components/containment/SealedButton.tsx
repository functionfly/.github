import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Loader2 } from 'lucide-react';

export interface SealedButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  iconLeft?: ReactNode;
  iconRight?: ReactNode;
}

export function SealedButton({
  children,
  size = 'md',
  loading = false,
  iconLeft,
  iconRight,
  disabled,
  className = '',
  ...props
}: SealedButtonProps) {
  const isDisabled = disabled || loading;

  return (
    <button
      className={`sealed-button sealed-button--${size} ${className}`}
      disabled={isDisabled}
      aria-busy={loading}
      {...props}
    >
      {loading ? (
        <Loader2 className="sealed-button__spinner" />
      ) : (
        <>
          {iconLeft && <span className="sealed-button__icon sealed-button__icon--left">{iconLeft}</span>}
          <span className="sealed-button__label">{children}</span>
          {iconRight && <span className="sealed-button__icon sealed-button__icon--right">{iconRight}</span>}
        </>
      )}
    </button>
  );
}