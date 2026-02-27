import { TrendingUp, TrendingDown } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface StatCardProps {
  title: string;
  value: string | number;
  change?: {
    value: number;
    label: string;
  };
  icon: React.ReactNode;
  trend?: "up" | "down" | "neutral";
  className?: string;
}

export function StatCard({
  title,
  value,
  change,
  icon,
  trend = "neutral",
  className,
}: StatCardProps) {
  return (
    <Card className={cn("glass-card glow hover-lift overflow-hidden", className)}>
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <p className="text-sm font-medium text-text-secondary">{title}</p>
            <div className="flex items-baseline gap-2">
              <span className="text-3xl font-bold text-text-primary text-glow">{value}</span>
              {change && (
                <div
                  className={cn(
                    "flex items-center gap-0.5 text-xs font-medium",
                    trend === "up" && "text-emerald-400 glow-success",
                    trend === "down" && "text-red-400 glow-error",
                    trend === "neutral" && "text-text-muted"
                  )}
                >
                  {trend === "up" && <TrendingUp className="w-3 h-3" />}
                  {trend === "down" && <TrendingDown className="w-3 h-3" />}
                  <span>{change.value > 0 ? "+" : ""}{change.value}%</span>
                </div>
              )}
            </div>
            {change && (
              <p className="text-xs text-text-muted">{change.label}</p>
            )}
          </div>
          <div className="p-3 rounded-xl glass-light border border-brand-500/20 hover:border-brand-500/40 transition-all duration-300 shine-effect">
            {icon}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
