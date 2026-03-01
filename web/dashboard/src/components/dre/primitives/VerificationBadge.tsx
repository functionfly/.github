import { CheckCircle, XCircle, Clock, AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";

export type VerificationStatus = "verified" | "failed" | "pending" | "warning";

export interface VerificationBadgeProps {
  /** Verification status */
  status: VerificationStatus;
  /** Label text */
  label?: string;
  /** Show icon */
  showIcon?: boolean;
  /** Size variant */
  size?: "sm" | "md" | "lg";
  /** Custom className */
  className?: string;
  /** Click handler */
  onClick?: () => void;
}

const statusConfig = {
  verified: {
    icon: CheckCircle,
    label: "Verified",
    className: "bg-green-500/10 text-green-500 border-green-500/20",
    iconClassName: "text-green-500",
  },
  failed: {
    icon: XCircle,
    label: "Failed",
    className: "bg-red-500/10 text-red-500 border-red-500/20",
    iconClassName: "text-red-500",
  },
  pending: {
    icon: Clock,
    label: "Pending",
    className: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
    iconClassName: "text-yellow-500",
  },
  warning: {
    icon: AlertTriangle,
    label: "Warning",
    className: "bg-orange-500/10 text-orange-500 border-orange-500/20",
    iconClassName: "text-orange-500",
  },
};

const sizeConfig = {
  sm: {
    badge: "text-xs px-2 py-0.5 gap-1",
    icon: "h-3 w-3",
    text: "text-xs",
  },
  md: {
    badge: "text-sm px-2.5 py-1 gap-1.5",
    icon: "h-4 w-4",
    text: "text-sm",
  },
  lg: {
    badge: "text-base px-3 py-1.5 gap-2",
    icon: "h-5 w-5",
    text: "text-base",
  },
};

export function VerificationBadge({
  status,
  label,
  showIcon = true,
  size = "md",
  className,
  onClick,
}: VerificationBadgeProps) {
  const config = statusConfig[status];
  const sizeStyles = sizeConfig[size];
  const Icon = config.icon;

  return (
    <span
      className={cn(
        "inline-flex items-center font-medium rounded-md border",
        config.className,
        sizeStyles.badge,
        onClick && "cursor-pointer hover:opacity-80 transition-opacity",
        className
      )}
      onClick={onClick}
    >
      {showIcon && <Icon className={cn(sizeStyles.icon, config.iconClassName)} />}
      {label ?? config.label}
    </span>
  );
}
