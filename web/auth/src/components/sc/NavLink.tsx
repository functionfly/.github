import React from "react";

interface NavLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  active?: boolean;
}

export const NavLink: React.FC<NavLinkProps> = ({
  active = false,
  children,
  className = "",
  style,
  ...props
}) => {
  return (
    <a
      className={`nav-link ${active ? "nav-link-active" : ""} ${className}`}
      style={{
        fontFamily: "var(--font-body)",
        fontSize: "14px",
        lineHeight: 1,
        color: active ? "var(--status-ok)" : "var(--text-dim)",
        padding: "var(--space-2) var(--space-3)",
        borderRadius: "var(--radius-sm)",
        borderBottom: active
          ? "2px solid var(--status-ok)"
          : "2px solid transparent",
        whiteSpace: "nowrap",
        textDecoration: "none",
        transition:
          "color var(--duration-fast) var(--ease-out), background var(--duration-fast) var(--ease-out)",
        ...style,
      }}
      {...props}
    >
      {children}
    </a>
  );
};
