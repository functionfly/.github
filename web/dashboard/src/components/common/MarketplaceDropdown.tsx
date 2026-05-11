import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { ChevronDown, Store, Bot, ShoppingBag, Plus } from "lucide-react";
import { cn } from "@/lib/utils";
import { ROUTES } from "@/lib/constants";
import { useThemeStore } from "@/stores/themeStore";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface MarketplaceDropdownProps {
  className?: string;
}

interface MarketplaceItem {
  label: string;
  description: string;
  path: string;
  icon: React.ComponentType<{ className?: string }>;
}

export function MarketplaceDropdown({ className }: MarketplaceDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const theme = useThemeStore((state) => state.theme);
  const location = useLocation();

  const isActive =
    location.pathname.startsWith("/marketplace") || 
    location.pathname === ROUTES.DASHBOARD ||
    location.pathname.startsWith("/functions/discovery");

  const marketplaceItems: MarketplaceItem[] = [
    {
      label: "Browse Agents",
      description: "Find AI agents for any task",
      path: "/marketplace/agents",
      icon: Bot,
    },
    {
      label: "Function Marketplace",
      description: "Discover serverless functions",
      path: "/functions/discovery",
      icon: Store,
    },
    {
      label: "My Purchases",
      description: "Your bought agents & functions",
      path: "/marketplace/purchases",
      icon: ShoppingBag,
    },
    {
      label: "Create Listing",
      description: "Sell your work",
      path: "/marketplace/new",
      icon: Plus,
    },
  ];

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <button
          className={cn(
            "flex items-center gap-1 transition-colors font-medium",
            isActive
              ? "text-text-primary"
              : "text-text-secondary hover:text-text-primary",
            className
          )}
          style={theme === 'light' ? {
            color: isActive ? '#7c3aed' : '#1a1a2e',
          } : {}}
        >
          Marketplace
          <ChevronDown
            className={cn(
              "w-4 h-4 transition-transform duration-200",
              isOpen && "rotate-180"
            )}
          />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-80">
        <div className="px-3 py-3">
          <p className="text-sm font-semibold text-text-primary">Marketplace</p>
          <p className="text-xs text-text-secondary mt-0.5">Discover and deploy AI agents and functions</p>
        </div>
        <DropdownMenuSeparator />
        {marketplaceItems.map((item) => {
          const Icon = item.icon;
          const isItemActive = location.pathname === item.path;
          return (
            <DropdownMenuItem key={item.path} asChild>
              <Link
                to={item.path}
                className={cn(
                  "flex items-center gap-3 p-3 cursor-pointer rounded-lg transition-all duration-200 mx-2 my-1",
                  "hover:bg-bg-hover/50 focus:bg-bg-hover/30 focus:outline-none focus:ring-2 focus:ring-brand-500/20",
                  isItemActive && "bg-brand-500/10 border border-brand-500/20"
                )}
              >
                <div className={cn(
                  "flex items-center justify-center w-10 h-10 rounded-lg border",
                  isItemActive 
                    ? "bg-brand-500/10 border-brand-500/30" 
                    : "bg-bg-secondary border-border-subtle"
                )}>
                  <Icon className={cn(
                    "w-5 h-5",
                    isItemActive ? "text-brand-500" : "text-text-secondary"
                  )} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-text-primary transition-colors duration-200">
                    {item.label}
                  </div>
                  <div className="text-xs text-text-secondary mt-0.5 transition-colors duration-200">
                    {item.description}
                  </div>
                </div>
              </Link>
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default MarketplaceDropdown;
