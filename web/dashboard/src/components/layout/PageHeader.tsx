import { ReactNode } from "react";
import { motion } from "framer-motion";
import { ChevronRight, Home } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { ROUTES } from "@/lib/constants";

interface BreadcrumbItem {
  label: string;
  path?: string;
  icon?: React.ComponentType<{ className?: string }>;
}

interface ActionButton {
  label: string;
  onClick: () => void;
  variant?: "default" | "outline" | "ghost" | "secondary";
  size?: "sm" | "md" | "lg";
  icon?: React.ComponentType<{ className?: string }>;
  disabled?: boolean;
}

interface PageHeaderProps {
  title: string;
  subtitle?: string;
  breadcrumbs?: BreadcrumbItem[];
  actions?: ActionButton[];
  badges?: Array<{
    label: string;
    variant?: "default" | "secondary" | "destructive" | "outline";
  }>;
  className?: string;
  animate?: boolean;
}

export function PageHeader({
  title,
  subtitle,
  breadcrumbs,
  actions,
  badges,
  className,
  animate = true,
}: PageHeaderProps) {
  const location = useLocation();

  // Auto-generate breadcrumbs if not provided
  const defaultBreadcrumbs = generateDefaultBreadcrumbs(location.pathname);
  const displayBreadcrumbs = breadcrumbs || defaultBreadcrumbs;

  const Header = animate ? motion.div : "div";
  const headerProps = animate ? {
    initial: { opacity: 0, y: -10 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.3, ease: "easeOut" },
  } : {};

  return (
    <Header
      {...headerProps}
      className={cn("space-y-4", className)}
    >
      {/* Breadcrumbs */}
      {displayBreadcrumbs && displayBreadcrumbs.length > 1 && (
        <nav className="flex items-center gap-2 text-sm text-text-muted">
          {displayBreadcrumbs.map((crumb, index) => {
            const isLast = index === displayBreadcrumbs.length - 1;
            const Icon = crumb.icon;

            return (
              <div key={index} className="flex items-center gap-2">
                {index > 0 && (
                  <ChevronRight className="w-4 h-4" />
                )}
                {crumb.path && !isLast ? (
                  <Link
                    to={crumb.path}
                    className="flex items-center gap-2 hover:text-text-primary transition-colors"
                  >
                    {Icon && <Icon className="w-4 h-4" />}
                    <span>{crumb.label}</span>
                  </Link>
                ) : (
                  <span className={cn(
                    "flex items-center gap-2",
                    isLast ? "text-text-primary font-medium" : "text-text-secondary"
                  )}>
                    {Icon && <Icon className="w-4 h-4" />}
                    <span>{crumb.label}</span>
                  </span>
                )}
              </div>
            );
          })}
        </nav>
      )}

      {/* Header Content */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-text-primary">{title}</h1>
            {badges && badges.length > 0 && (
              <div className="flex gap-2">
                {badges.map((badge, index) => (
                  <Badge key={index} variant={badge.variant}>
                    {badge.label}
                  </Badge>
                ))}
              </div>
            )}
          </div>
          {subtitle && (
            <p className="text-text-secondary">{subtitle}</p>
          )}
        </div>

        {/* Actions */}
        {actions && actions.length > 0 && (
          <div className="flex items-center gap-2">
            {actions.map((action, index) => {
              const Icon = action.icon;
              return (
                <Button
                  key={index}
                  variant={action.variant || "default"}
                  size={action.size || "sm"}
                  onClick={action.onClick}
                  disabled={action.disabled}
                >
                  {Icon && <Icon className="w-4 h-4 mr-2" />}
                  {action.label}
                </Button>
              );
            })}
          </div>
        )}
      </div>
    </Header>
  );
}

// Helper function to generate default breadcrumbs based on current path
function generateDefaultBreadcrumbs(pathname: string): BreadcrumbItem[] {
  const segments = pathname.split('/').filter(Boolean);
  const breadcrumbs: BreadcrumbItem[] = [
    { label: "Dashboard", path: ROUTES.DASHBOARD, icon: Home }
  ];

  if (segments.length === 0 || segments[0] === 'dashboard') {
    return breadcrumbs;
  }

  // Add section breadcrumb
  const section = segments[0];
  switch (section) {
    case 'functions':
      breadcrumbs.push({ label: "Functions", path: ROUTES.FUNCTIONS });
      break;
    case 'providers':
      breadcrumbs.push({ label: "Providers", path: ROUTES.PROVIDERS });
      break;
    case 'analytics':
      breadcrumbs.push({ label: "Analytics", path: ROUTES.ANALYTICS });
      break;
    case 'settings':
      breadcrumbs.push({ label: "Settings", path: ROUTES.SETTINGS });
      break;
    case 'admin':
      breadcrumbs.push({ label: "Admin" });
      break;
  }

  // Add subsection breadcrumb if applicable
  if (segments.length > 1 && section !== 'admin') {
    const subsection = segments[1];
    if (subsection && subsection !== 'new') {
      breadcrumbs.push({ label: subsection.charAt(0).toUpperCase() + subsection.slice(1) });
    } else if (subsection === 'new') {
      breadcrumbs.push({ label: "New" });
    }
  } else if (segments.length > 1 && section === 'admin') {
    const adminSection = segments[1];
    switch (adminSection) {
      case 'tenants':
        breadcrumbs.push({ label: "Tenants", path: ROUTES.ADMIN_TENANTS });
        break;
      case 'users':
        breadcrumbs.push({ label: "Users", path: ROUTES.ADMIN_USERS });
        break;
      case 'billing':
        breadcrumbs.push({ label: "Billing", path: ROUTES.ADMIN_BILLING });
        break;
      case 'audit':
        breadcrumbs.push({ label: "Audit Log", path: ROUTES.ADMIN_AUDIT });
        break;
      case 'system':
        breadcrumbs.push({ label: "System", path: ROUTES.ADMIN_SYSTEM });
        break;
    }
  }

  return breadcrumbs;
}