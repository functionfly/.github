import { Link, useLocation } from "react-router-dom";
import { Store } from "lucide-react";
import { cn } from "@/lib/utils";
import { ROUTES } from "@/lib/constants";
import { useThemeStore } from "@/stores/themeStore";

interface MarketplaceDropdownProps {
  className?: string;
  style?: React.CSSProperties;
}

export function MarketplaceDropdown({ className, style }: MarketplaceDropdownProps) {
  const theme = useThemeStore((state) => state.theme);
  const location = useLocation();

  const isActive = location.pathname.startsWith("/marketplace");

  return (
    <Link
      to={ROUTES.MARKETPLACE}
      className={cn(
        "flex items-center gap-1.5 transition-colors font-medium",
        isActive
          ? "text-text-primary"
          : "text-text-secondary hover:text-text-primary",
        className
      )}
      style={{
        ...(theme === 'light' ? {
          color: isActive ? '#7c3aed' : '#1a1a2e',
        } : {}),
        ...style,
      }}
    >
      <Store style={{ width: 14, height: 14 }} />
      Marketplace
    </Link>
  );
}

export default MarketplaceDropdown;
