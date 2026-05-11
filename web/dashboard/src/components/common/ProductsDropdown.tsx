import { useState } from "react";
import { Link } from "react-router-dom";
import { ChevronDown, Zap, Cloud, BarChart3, Network, Boxes, Eye } from "lucide-react";
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

interface ProductItem {
  label: string;
  description: string;
  path: string;
  icon: React.ComponentType<{ className?: string }>;
}

export function ProductsDropdown({ className }: ProductsDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const theme = useThemeStore((state) => state.theme);

  const productItems: ProductItem[] = [
    {
      label: "Functions",
      description: "Deploy and manage serverless functions",
      path: "/functions",
      icon: Zap,
    },
    {
      label: "Providers",
      description: "Manage cloud providers and resources",
      path: "/providers",
      icon: Cloud,
    },
    {
      label: "Analytics",
      description: "Monitor performance and usage",
      path: "/analytics",
      icon: BarChart3,
    },
    {
      label: "API Gateway",
      description: "Manage APIs and endpoints",
      path: "/api-gateway",
      icon: Network,
    },
    {
      label: "State Fabric",
      description: "Manage state and data orchestration",
      path: "/products/state-fabric",
      icon: Boxes,
    },
    {
      label: "Monitoring",
      description: "Real-time monitoring and alerts",
      path: "/monitoring",
      icon: Eye,
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
      <DropdownMenuContent align="start" className="w-72">
        <div className="px-2 py-1.5 text-sm font-semibold text-text-primary">
          Products
        </div>
        <DropdownMenuSeparator />
        {productItems.map((item) => {
          const Icon = item.icon;
          return (
            <DropdownMenuItem key={item.path} asChild>
              <Link
                to={item.path}
                className="flex items-center gap-3 p-3 cursor-pointer rounded-md transition-all duration-200 hover:bg-bg-hover/50 hover:shadow-sm focus:bg-bg-hover/30 focus:outline-none focus:ring-2 focus:ring-brand-500/20"
              >
                <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-gradient-to-br from-indigo-500/10 to-purple-500/10 border border-indigo-500/20">
                  <Icon className="w-5 h-5 text-indigo-500" />
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
