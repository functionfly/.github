import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import {
  Activity,
  AlertCircle,
  AlertTriangle,
  CheckCircle,
  Wrench,
} from "lucide-react";

interface StatusBadgeProps {
  status: string;
  size?: "sm" | "md" | "lg";
  showPulse?: boolean;
}

const statusConfig = {
  operational: {
    icon: CheckCircle,
    label: "Operational",
    colors: {
      bg: "bg-emerald-500/10",
      border: "border-emerald-500/30",
      text: "text-emerald-400",
      glow: "shadow-emerald-500/20",
      dot: "bg-emerald-500",
    },
  },
  degraded: {
    icon: AlertTriangle,
    label: "Degraded",
    colors: {
      bg: "bg-amber-500/10",
      border: "border-amber-500/30",
      text: "text-amber-400",
      glow: "shadow-amber-500/20",
      dot: "bg-amber-500",
    },
  },
  degraded_performance: {
    icon: AlertTriangle,
    label: "Degraded",
    colors: {
      bg: "bg-amber-500/10",
      border: "border-amber-500/30",
      text: "text-amber-400",
      glow: "shadow-amber-500/20",
      dot: "bg-amber-500",
    },
  },
  major_outage: {
    icon: AlertCircle,
    label: "Major Outage",
    colors: {
      bg: "bg-red-500/10",
      border: "border-red-500/30",
      text: "text-red-400",
      glow: "shadow-red-500/20",
      dot: "bg-red-500",
    },
  },
  partial_outage: {
    icon: AlertCircle,
    label: "Partial Outage",
    colors: {
      bg: "bg-orange-500/10",
      border: "border-orange-500/30",
      text: "text-orange-400",
      glow: "shadow-orange-500/20",
      dot: "bg-orange-500",
    },
  },
  maintenance: {
    icon: Wrench,
    label: "Maintenance",
    colors: {
      bg: "bg-purple-500/10",
      border: "border-purple-500/30",
      text: "text-purple-400",
      glow: "shadow-purple-500/20",
      dot: "bg-purple-500",
    },
  },
};

const sizeConfig = {
  sm: {
    container: "px-2 py-0.5 text-xs gap-1",
    icon: "w-3 h-3",
    dot: "w-1.5 h-1.5",
  },
  md: {
    container: "px-2.5 py-1 text-sm gap-1.5",
    icon: "w-4 h-4",
    dot: "w-2 h-2",
  },
  lg: {
    container: "px-3 py-1.5 text-base gap-2",
    icon: "w-5 h-5",
    dot: "w-2.5 h-2.5",
  },
};

export function StatusBadge({
  status,
  size = "sm",
  showPulse = true,
}: StatusBadgeProps) {
  const config = statusConfig[status as keyof typeof statusConfig] || {
    icon: Activity,
    label: "Unknown",
    colors: {
      bg: "bg-gray-500/10",
      border: "border-gray-500/30",
      text: "text-gray-400",
      glow: "shadow-gray-500/20",
      dot: "bg-gray-500",
    },
  };

  const sizeClasses = sizeConfig[size];
  const Icon = config.icon;
  const isOperational = status === "operational";

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      whileHover={{ scale: 1.02 }}
      className={cn(
        "inline-flex items-center rounded-full border font-medium",
        "transition-all duration-200",
        config.colors.bg,
        config.colors.border,
        config.colors.text,
        isOperational && showPulse && "shadow-lg",
        isOperational && showPulse && config.colors.glow,
        sizeClasses.container,
      )}
    >
      {/* Status indicator dot with pulse */}
      <div className="relative flex items-center justify-center">
        {isOperational && showPulse && (
          <>
            <motion.span
              className={cn(
                "absolute inline-flex h-full w-full rounded-full opacity-40",
                config.colors.dot,
              )}
              animate={{ scale: [1, 1.8, 2], opacity: [0.4, 0.2, 0] }}
              transition={{ duration: 2, repeat: Infinity, ease: "easeOut" }}
            />
            <motion.span
              className={cn(
                "absolute inline-flex h-full w-full rounded-full opacity-20",
                config.colors.dot,
              )}
              animate={{ scale: [1, 2, 2.5], opacity: [0.3, 0.1, 0] }}
              transition={{
                duration: 2,
                repeat: Infinity,
                ease: "easeOut",
                delay: 0.5,
              }}
            />
          </>
        )}
        <span
          className={cn(
            "relative inline-flex rounded-full",
            config.colors.dot,
            sizeClasses.dot,
            isOperational && showPulse && "shadow-sm",
          )}
        />
      </div>

      <Icon className={cn(sizeClasses.icon, "shrink-0")} />
      <span>{config.label}</span>
    </motion.div>
  );
}

export function StatusDot({
  status,
  size = "md",
  pulse = true,
}: {
  status: string;
  size?: "sm" | "md" | "lg";
  pulse?: boolean;
}) {
  const colors = {
    operational: "bg-emerald-500 shadow-emerald-500/50",
    degraded: "bg-amber-500 shadow-amber-500/50",
    degraded_performance: "bg-amber-500 shadow-amber-500/50",
    major_outage: "bg-red-500 shadow-red-500/50",
    partial_outage: "bg-orange-500 shadow-orange-500/50",
    maintenance: "bg-purple-500 shadow-purple-500/50",
  };

  const sizes = {
    sm: "w-2 h-2",
    md: "w-2.5 h-2.5",
    lg: "w-3 h-3",
  };

  const colorClass = colors[status as keyof typeof colors] || "bg-gray-500";
  const isOperational = status === "operational";

  return (
    <div className="relative flex items-center justify-center">
      {isOperational && pulse && (
        <>
          <motion.span
            className={cn(
              "absolute inline-flex h-full w-full rounded-full",
              colorClass.split(" ")[0],
            )}
            animate={{ scale: [1, 2, 2.5], opacity: [0.5, 0.3, 0] }}
            transition={{ duration: 2, repeat: Infinity, ease: "easeOut" }}
          />
          <motion.span
            className={cn(
              "absolute inline-flex h-full w-full rounded-full",
              colorClass.split(" ")[0],
            )}
            animate={{ scale: [1, 1.5, 2], opacity: [0.3, 0.15, 0] }}
            transition={{
              duration: 2,
              repeat: Infinity,
              ease: "easeOut",
              delay: 0.6,
            }}
          />
        </>
      )}
      <span
        className={cn(
          "relative inline-flex rounded-full shadow-lg",
          colorClass,
          sizes[size],
        )}
      />
    </div>
  );
}
