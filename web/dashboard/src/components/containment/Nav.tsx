import type { ReactNode } from 'react';

export interface NavProps {
  children?: ReactNode;
  className?: string;
}

export function Nav({ children, className = '' }: NavProps) {
  return (
    <nav className={`nav ${className}`}>
      <div className="nav__inner">
        <div className="nav__brand">
          <div className="nav__logo" aria-hidden="true" />
          <span className="nav__wordmark">FunctionFly</span>
        </div>
        <div className="nav__links">{children}</div>
      </div>
    </nav>
  );
}

export interface NavLinkProps {
  href: string;
  children: ReactNode;
  isActive?: boolean;
  className?: string;
}

export function NavLink({ href, children, isActive = false, className = '' }: NavLinkProps) {
  return (
    <a href={href} className={`nav__link ${isActive ? 'nav__link--active' : ''} ${className}`}>
      {children}
    </a>
  );
}
