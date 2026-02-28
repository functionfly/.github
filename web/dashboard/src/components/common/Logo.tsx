import { cn } from "@/lib/utils";
import { APP_NAME } from "@/lib/constants";

interface LogoProps {
  size?: "sm" | "md" | "lg";
  showText?: boolean;
  className?: string;
  variant?: "default" | "white";
}

const sizeConfig = {
  sm: {
    icon: 20,
    text: "text-lg",
  },
  md: {
    icon: 28,
    text: "text-xl",
  },
  lg: {
    icon: 36,
    text: "text-2xl",
  },
};

export function Logo({ size = "md", showText = true, className, variant = "default" }: LogoProps) {
  const config = sizeConfig[size];
  const logoSrc = variant === "white" ? "/logo/logo-icon-white.svg" : "/logo/logo-icon.svg";

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <img
        src={logoSrc}
        alt="FunctionFly"
        width={config.icon}
        height={config.icon}
        className="shrink-0"
      />
      {showText && (
        <span className={cn("font-bold tracking-tight gradient-text", config.text)}>
          {APP_NAME}
        </span>
      )}
    </div>
  );
}
