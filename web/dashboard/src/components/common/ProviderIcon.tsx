import { Terminal, Zap } from "lucide-react";
import { CloudflareIcon, VercelIcon, FlyIoIcon } from "@/pages/LandingPage/components/icons";
import { cn } from "@/lib/utils";

interface ProviderIconProps {
  provider: "workers" | "vercel" | "fly" | "deno" | "functionfly-edge" | string;
  size?: "sm" | "md" | "lg";
  className?: string;
}

const sizeConfig = {
  sm: "w-4 h-4",
  md: "w-5 h-5",
  lg: "w-6 h-6",
};

const providerConfig = {
  workers: {
    Icon: CloudflareIcon,
    color: "#f48120",
    label: "Cloudflare Workers",
  },
  vercel: {
    Icon: VercelIcon,
    color: "#000000",
    label: "Vercel",
  },
  fly: {
    Icon: FlyIoIcon,
    color: "#7b68ee",
    label: "Fly.io",
  },
  deno: {
    Icon: Terminal,
    color: "#ffffff",
    label: "Deno Deploy",
  },
  "functionfly-edge": {
    Icon: Zap,
    color: "#6366f1",
    label: "FunctionFly Edge",
  },
};

export function ProviderIcon({ provider, size = "md", className }: ProviderIconProps) {
  const config = providerConfig[provider as keyof typeof providerConfig] || {
    Icon: Zap,
    color: "#a0a0b0",
    label: provider,
  };

  const Icon = config.Icon;

  return (
    <div
      className={cn(
        "flex items-center justify-center rounded-md shrink-0",
        sizeConfig[size],
        className
      )}
      style={{ color: config.color }}
      title={config.label}
    >
      <Icon className="w-full h-full text-inherit" />
    </div>
  );
}
