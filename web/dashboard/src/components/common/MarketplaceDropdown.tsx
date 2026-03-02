import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";
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

export function MarketplaceDropdown({ className }: MarketplaceDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const theme = useThemeStore((state) => state.theme);
  const location = useLocation();

  const isActive = location.pathname.startsWith("/marketplace");

  const marketplaceItems = [
    {
      label: "Function Marketplace",
      description: "Discover and deploy serverless functions",
      path: "/marketplace/functions",
      icon: "⚡",
    },
    {
      label: "Agent Marketplace",
      description: "Browse and hire AI agents",
      path: "/marketplace/agents",
      icon: "🤖",
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
      <DropdownMenuContent align="start" className="w-64">
        <div className="px-2 py-1.5 text-sm font-semibold text-text-primary">
          Marketplace
        </div>
        <DropdownMenuSeparator />
        {marketplaceItems.map((item) => (
          <DropdownMenuItem key={item.path} asChild>
            <Link
              to={item.path}
              className="flex items-start gap-3 p-3 cursor-pointer rounded-md transition-all duration-200 hover:bg-bg-hover/50 hover:shadow-sm focus:bg-bg-hover/30 focus:outline-none focus:ring-2 focus:ring-brand-500/20"
            >
              <span className="text-lg transition-transform duration-200 hover:scale-110">{item.icon}</span>
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
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default MarketplaceDropdown;
