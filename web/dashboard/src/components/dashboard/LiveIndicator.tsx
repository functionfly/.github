import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { Activity, Radio, Wifi } from "lucide-react";

export type LiveStatus = "connected" | "connecting" | "disconnected" | "error";

export interface LiveIndicatorProps {
  status: LiveStatus;
  label?: string;
  showPulse?: boolean;
  showIcon?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
  onClick?: () => void;
}

const statusConfig = {
  connected: {
    color: "var(--color-aviation-green, #10b981)",
    bgColor: "bg-[var(--color-aviation-green)]/10",
    icon: Wifi,
    label: "Live",
    animation: "pulse",
  },
  connecting: {
    color: "var(--color-aviation-amber, #f59e0b)",
    bgColor: "bg-[var(--color-aviation-amber)]/10",
    icon: Activity,
    label: "Connecting...",
    animation: "spin",
  },
  disconnected: {
    color: "var(--color-text-muted, #6b7280)",
    bgColor: "bg-bg-tertiary",
    icon: Radio,
    label: "Offline",
    animation: "none",
  },
  error: {
    color: "var(--color-error, #ef4444)",
    bgColor: "bg-[var(--color-error)]/10",
    icon: Activity,
    label: "Error",
    animation: "shake",
  },
};

const sizeConfig = {
  sm: {
    dot: "w-1.5 h-1.5",
    container: "px-2 py-0.5 gap-1.5",
    icon: "w-3 h-3",
    text: "text-[10px]",
  },
  md: {
    dot: "w-2 h-2",
    container: "px-2.5 py-1 gap-2",
    icon: "w-3.5 h-3.5",
    text: "text-xs",
  },
  lg: {
    dot: "w-2.5 h-2.5",
    container: "px-3 py-1.5 gap-2",
    icon: "w-4 h-4",
    text: "text-sm",
  },
};

export function LiveIndicator({
  status,
  label,
  showPulse = true,
  showIcon = true,
  size = "md",
  className,
  onClick,
}: LiveIndicatorProps) {
  const config = statusConfig[status];
  const sizes = sizeConfig[size];
  const StatusIcon = config.icon;
  const displayLabel = label || config.label;

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      className={cn(
        "inline-flex items-center rounded-full border border-border/50",
        sizes.container,
        config.bgColor,
        onClick && "cursor-pointer hover:opacity-80 transition-opacity",
        className
      )}
      onClick={onClick}
    >
      {/* Status Dot with Pulse */}
      <div className="relative flex items-center justify-center">
        {showPulse && status === "connected" && (
          <motion.div
            className="absolute inset-0 rounded-full"
            style={{ backgroundColor: config.color }}
            animate={{
              scale: [1, 2, 2],
              opacity: [0.5, 0.3, 0],
            }}
            transition={{
              duration: 1.5,
              repeat: Infinity,
              ease: "easeOut",
            }}
          />
        )}
        <motion.div
          className={cn("rounded-full relative z-10", sizes.dot)}
          style={{ backgroundColor: config.color }}
          animate={
            config.animation === "spin"
              ? { rotate: 360 }
              : config.animation === "shake"
                ? {
                    x: [0, -2, 2, -2, 2, 0],
                    transition: { repeat: Infinity, duration: 0.5 },
                  }
                : {}
          }
          transition={
            config.animation === "spin"
              ? { duration: 1, repeat: Infinity, ease: "linear" }
              : {}
          }
        />
      </div>

      {/* Icon */}
      {showIcon && (
        <StatusIcon
          className={cn(sizes.icon)}
          style={{ color: config.color }}
        />
      )}

      {/* Label */}
      <span
        className={cn(
          "font-medium whitespace-nowrap",
          sizes.text,
          status === "disconnected" && "text-text-muted",
          status !== "disconnected" && "text-text-primary"
        )}
      >
        {displayLabel}
      </span>
    </motion.div>
  );
}

// Compact version for tight spaces
export function LiveDot({
  status,
  size = "md",
  className,
}: {
  status: LiveStatus;
  size?: "sm" | "md" | "lg";
  className?: string;
}) {
  const config = statusConfig[status];
  const sizeClasses = {
    sm: "w-2 h-2",
    md: "w-2.5 h-2.5",
    lg: "w-3 h-3",
  };

  return (
    <div className={cn("relative inline-flex", className)}>
      {status === "connected" && (
        <motion.div
          className="absolute inset-0 rounded-full"
          style={{ backgroundColor: config.color }}
          animate={{
            scale: [1, 2.5, 2.5],
            opacity: [0.6, 0.2, 0],
          }}
          transition={{
            duration: 1.5,
            repeat: Infinity,
            ease: "easeOut",
          }}
        />
      )}
      <div
        className={cn("rounded-full relative z-10", sizeClasses[size])}
        style={{ backgroundColor: config.color }}
      />
    </div>
  );
}
