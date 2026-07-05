import { cn } from "@/lib/utils";

interface StatusBadgeProps {
  status: "online" | "offline" | "degraded" | "pending" | "error";
  showPulse?: boolean;
  className?: string;
}

const statusConfig = {
  online: {
    bg: "bg-emerald-500",
    text: "text-emerald-400",
    border: "border-emerald-500/30",
    label: "Online",
  },
  offline: {
    bg: "bg-red-500",
    text: "text-red-400",
    border: "border-red-500/30",
    label: "Offline",
  },
  degraded: {
    bg: "bg-amber-500",
    text: "text-amber-400",
    border: "border-amber-500/30",
    label: "Degraded",
  },
  pending: {
    bg: "bg-gray-500",
    text: "text-gray-400",
    border: "border-gray-500/30",
    label: "Pending",
  },
  error: {
    bg: "bg-red-500",
    text: "text-red-400",
    border: "border-red-500/30",
    label: "Error",
  },
};

export function StatusBadge({ status, showPulse = true, className }: StatusBadgeProps) {
  const config = statusConfig[status];

  return (
    <div
      className={cn(
        "inline-flex items-center gap-2 px-2.5 py-1 rounded-full border bg-opacity-10",
        config.border,
        className
      )}
    >
      <span className="relative flex h-2 w-2">
        {showPulse && status === "online" && (
          <span
            className={cn(
              "animate-ping absolute inline-flex h-full w-full rounded-full opacity-75",
              config.bg
            )}
          />
        )}
        <span className={cn("relative inline-flex rounded-full h-2 w-2", config.bg)} />
      </span>
      <span className={cn("text-xs font-medium", config.text)}>{config.label}</span>
    </div>
  );
}
