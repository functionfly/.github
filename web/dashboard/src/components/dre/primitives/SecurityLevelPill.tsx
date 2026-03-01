import { Shield } from "lucide-react";
import { cn } from "@/lib/utils";

export type SecurityLevel = "standard" | "enhanced" | "strict" | "fips";

export interface SecurityLevelPillProps {
  /** Security level */
  level: SecurityLevel;
  /** Custom className */
  className?: string;
  /** Click handler */
  onClick?: () => void;
}

const levelConfig = {
  standard: {
    label: "Standard",
    className: "bg-gray-500/10 text-gray-500 border-gray-500/20",
  },
  enhanced: {
    label: "Enhanced",
    className: "bg-blue-500/10 text-blue-500 border-blue-500/20",
  },
  strict: {
    label: "Strict",
    className: "bg-orange-500/10 text-orange-500 border-orange-500/20",
  },
  fips: {
    label: "FIPS 140-2",
    className: "bg-green-500/10 text-green-500 border-green-500/20",
  },
};

export function SecurityLevelPill({
  level,
  className,
  onClick,
}: SecurityLevelPillProps) {
  const config = levelConfig[level];

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 text-xs font-medium px-2 py-0.5 rounded-md border",
        config.className,
        onClick && "cursor-pointer hover:opacity-80 transition-opacity",
        className
      )}
      onClick={onClick}
    >
      <Shield className="h-3 w-3" />
      {config.label}
    </span>
  );
}
