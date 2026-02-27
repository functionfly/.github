import { Cloud, Triangle, Plane, Terminal, Zap } from "lucide-react";
import { cn } from "@/lib/utils";

interface ProviderIconProps {
  provider: "workers" | "vercel" | "fly" | "deno" | "functionfly-edge" | string;
  size?: "sm" | "md" | "lg";
  className?: string;
}

const providerConfig = {
  workers: {
    icon: Cloud,
    color: "#f48120",
    label: "Cloudflare Workers",
  },
  vercel: {
    icon: Triangle,
    color: "#ffffff",
    label: "Vercel",
  },
  fly: {
    icon: Plane,
    color: "#7b68ee",
    label: "Fly.io",
  },
  deno: {
    icon: Terminal,
    color: "#ffffff",
    label: "Deno Deploy",
  },
  "functionfly-edge": {
    icon: Zap,
    color: "#6366f1",
    label: "FunctionFly Edge",
  },
};

const sizeConfig = {
  sm: "w-4 h-4",
  md: "w-5 h-5",
  lg: "w-6 h-6",
};

export function ProviderIcon({ provider, size = "md", className }: ProviderIconProps) {
  const config = providerConfig[provider as keyof typeof providerConfig] || {
    icon: Cloud,
    color: "#a0a0b0",
    label: provider,
  };

  const Icon = config.icon;

  return (
    <div
      className={cn(
        "flex items-center justify-center rounded-md",
        sizeConfig[size],
        className
      )}
      style={{ color: config.color }}
      title={config.label}
    >
      <Icon className="w-full h-full" />
    </div>
  );
}
