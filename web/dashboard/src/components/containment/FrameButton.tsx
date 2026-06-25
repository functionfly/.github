import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Loader2 } from 'lucide-react';

export interface FrameButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  iconLeft?: ReactNode;
  iconRight?: ReactNode;
}

export function FrameButton({
  children,
  size = 'md',
  loading = false,
  iconLeft,
  iconRight,
  disabled,
  className = '',
  ...props
}: FrameButtonProps) {
  const isDisabled = disabled || loading;

  return (
    <button
      className={`frame-button frame-button--${size} ${className}`}
      disabled={isDisabled}
      aria-busy={loading}
      {...props}
    >
      {loading ? (
        <Loader2 className="frame-button__spinner" />
      ) : (
        <>
          {iconLeft && <span className="frame-button__icon frame-button__icon--left">{iconLeft}</span>}
          <span className="frame-button__label">{children}</span>
          {iconRight && <span className="frame-button__icon frame-button__icon--right">{iconRight}</span>}
        </>
      )}
    </button>
  );
}