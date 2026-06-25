import React, { forwardRef, type ReactNode } from 'react';

export type TactileButtonVariant = 'default' | 'primary' | 'ghost';
export type TactileButtonSize = 'sm' | 'md' | 'lg';

interface BaseTactileButtonProps {
  children?: ReactNode;
  variant?: TactileButtonVariant;
  size?: TactileButtonSize;
  /** Show a left-side icon (ReactNode) */
  iconLeft?: ReactNode;
  /** Show a right-side icon (ReactNode) */
  iconRight?: ReactNode;
}

export type TactileButtonProps = BaseTactileButtonProps &
  (
    | (Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, 'children'> & { href?: undefined })
    | (Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'children'> & { href: string })
  );

const sizeStyles: Record<TactileButtonSize, React.CSSProperties> = {
  sm: { padding: '0.375rem 0.875rem', fontSize: '0.8125rem' },
  md: { padding: '0.625rem 1.25rem', fontSize: '0.875rem' },
  lg: { padding: '0.75rem 1.5rem', fontSize: '0.9375rem' },
};

/**
 * TactileButton - neumorphic button with press animation.
 * Press state animates shadow depth (NOT just opacity).
 * Focus ring is independent of shadow for accessibility.
 *
 * When `href` is provided, renders as an anchor (`<a>`) styled as a button.
 */
export const TactileButton = forwardRef<HTMLButtonElement | HTMLAnchorElement, TactileButtonProps>(
  function TactileButton(
    { children, variant = 'default', size = 'md', iconLeft, iconRight, className = '', style, href, ...rest },
    ref
  ) {
    const classes = [
      'hs-btn',
      variant === 'primary' ? 'hs-btn--primary' : '',
      className,
    ]
      .filter(Boolean)
      .join(' ');

    const sharedProps = {
      className: classes,
      style: { ...sizeStyles[size], ...style },
    };

    const content = (
      <>
        {iconLeft ? <span aria-hidden="true">{iconLeft}</span> : null}
        {children}
        {iconRight ? <span aria-hidden="true">{iconRight}</span> : null}
      </>
    );

    if (href) {
      return (
        <a ref={ref as React.Ref<HTMLAnchorElement>} href={href} {...sharedProps} {...(rest as React.AnchorHTMLAttributes<HTMLAnchorElement>)}>
          {content}
        </a>
      );
    }

    return (
      <button ref={ref as React.Ref<HTMLButtonElement>} {...sharedProps} {...(rest as React.ButtonHTMLAttributes<HTMLButtonElement>)}>
        {content}
      </button>
    );
  }
);
