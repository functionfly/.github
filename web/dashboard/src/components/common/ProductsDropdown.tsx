import { useState } from "react";
import { Link } from "react-router-dom";
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

interface ProductsDropdownProps {
  className?: string;
}

export function ProductsDropdown({ className }: ProductsDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const theme = useThemeStore((state) => state.theme);

  const productItems = [
    {
      label: "Functions",
      description: "Deploy and manage serverless functions",
      path: "/functions",
      icon: "⚡",
    },
    {
      label: "Providers",
      description: "Manage cloud providers and resources",
      path: "/providers",
      icon: "☁️",
    },
    {
      label: "Analytics",
      description: "Monitor performance and usage",
      path: "/analytics",
      icon: "📊",
    },
    {
      label: "API Gateway",
      description: "Manage APIs and endpoints",
      path: "/api-gateway",
      icon: "🚪",
    },
    {
      label: "State Fabric",
      description: "Manage state and data orchestration",
      path: "/products/state-fabric",
      icon: "🧵",
    },
    {
      label: "Monitoring",
      description: "Real-time monitoring and alerts",
      path: "/monitoring",
      icon: "👁️",
    },
  ];

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <button
          className={cn(
            "flex items-center gap-1 text-text-secondary hover:text-text-primary transition-colors font-medium",
            className
          )}
          style={theme === 'light' ? {
            color: '#1a1a2e',
          } : {}}
        >
          Products
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
          Products
        </div>
        <DropdownMenuSeparator />
        {productItems.map((item) => (
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