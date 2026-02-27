import { Zap } from "lucide-react";
import { cn } from "@/lib/utils";
import { APP_NAME } from "@/lib/constants";

interface LogoProps {
  size?: "sm" | "md" | "lg";
  showText?: boolean;
  className?: string;
}

const sizeConfig = {
  sm: {
    icon: "w-5 h-5",
    text: "text-lg",
  },
  md: {
    icon: "w-6 h-6",
    text: "text-xl",
  },
  lg: {
    icon: "w-8 h-8",
    text: "text-2xl",
  },
};

export function Logo({ size = "md", showText = true, className }: LogoProps) {
  const config = sizeConfig[size];

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <div className={cn(
        "flex items-center justify-center rounded-lg bg-gradient-to-br from-[#6366f1] to-[#8b5cf6] p-1.5",
        config.icon
      )}>
        <Zap className="w-full h-full text-white" fill="currentColor" />
      </div>
      {showText && (
        <span className={cn("font-bold tracking-tight gradient-text", config.text)}>
          {APP_NAME}
        </span>
      )}
    </div>
  );
}
