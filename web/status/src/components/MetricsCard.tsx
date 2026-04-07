import { cn } from "@/lib/utils";
import { TrendingUp, TrendingDown, Minus, type LucideIcon } from "lucide-react";

interface MetricsCardProps {
  title: string;
  value: string | number;
  change?: {
    value: number;
    label: string;
  };
  icon: LucideIcon;
  trend?: "up" | "down" | "neutral";
  className?: string;
  delay?: number;
}

export function MetricsCard({
  title,
  value,
  change,
  icon: Icon,
  trend = "neutral",
  className,
  delay = 0,
}: MetricsCardProps) {
  const trendIcons = {
    up: TrendingUp,
    down: TrendingDown,
    neutral: Minus,
  };
  
  const trendColors = {
    up: "text-emerald-400",
    down: "text-red-400",
    neutral: "text-text-muted",
  };

  const TrendIcon = trendIcons[trend];

  return (
    <div 
      className={cn(
        "glass-card p-6 metric-card",
        "animate-fade-in-up",
        className
      )}
      style={{ animationDelay: `${delay}ms` }}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-text-muted text-sm font-medium mb-1">{title}</p>
          <p className="text-2xl font-bold text-text-primary">{value}</p>
          
          {change && (
            <div className="flex items-center gap-1.5 mt-2">
              <TrendIcon className={cn("w-4 h-4", trendColors[trend])} />
              <span className={cn("text-sm font-medium", trendColors[trend])}>
                {Math.abs(change.value)}%
              </span>
              <span className="text-text-muted text-sm">{change.label}</span>
            </div>
          )}
        </div>
        
        <div className="relative">
          <div className="absolute inset-0 bg-gradient-to-r from-brand-500/20 to-purple-500/20 blur-xl" />
          <div className="relative p-3 rounded-xl bg-gradient-to-br from-brand-500/10 to-purple-500/10 border border-white/10">
            <Icon className="w-6 h-6 text-brand-400" />
          </div>
        </div>
      </div>
    </div>
  );
}

interface ServiceMetricProps {
  name: string;
  value: number;
  max: number;
  unit: string;
  status: "good" | "warning" | "critical";
  className?: string;
  delay?: number;
}

export function ServiceMetric({
  name,
  value,
  max,
  unit,
  status,
  className,
  delay = 0,
}: ServiceMetricProps) {
  const percentage = Math.min((value / max) * 100, 100);
  
  const statusColors = {
    good: "from-emerald-500 to-emerald-400",
    warning: "from-amber-500 to-amber-400",
    critical: "from-red-500 to-red-400",
  };

  const statusGlows = {
    good: "shadow-glow-success",
    warning: "shadow-glow-warning",
    critical: "shadow-glow-error",
  };

  return (
    <div 
      className={cn(
        "glass-card p-4",
        "animate-fade-in-up",
        className
      )}
      style={{ animationDelay: `${delay}ms` }}
    >
      <div className="flex items-center justify-between mb-3">
        <span className="text-text-secondary text-sm font-medium">{name}</span>
        <div className="flex items-baseline gap-1">
          <span className="text-lg font-bold text-text-primary">{value}</span>
          <span className="text-text-muted text-sm">{unit}</span>
        </div>
      </div>
      
      <div className="relative h-2 bg-bg-elevated rounded-full overflow-hidden">
        <div
          className={cn(
            "absolute inset-y-0 left-0 rounded-full transition-all duration-1000 ease-out bg-gradient-to-r",
            statusColors[status]
          )}
          style={{ width: `${percentage}%` }}
        >
          <div className={cn("absolute inset-0 rounded-full animate-glow-pulse", statusGlows[status])} />
        </div>
      </div>
    </div>
  );
}

// Sparkline chart component
interface SparklineProps {
  data: number[];
  className?: string;
  color?: string;
  fillColor?: string;
}

export function Sparkline({ 
  data, 
  className, 
  color = "#6366f1"
}: SparklineProps) {
  const min = Math.min(...data);
  const max = Math.max(...data);
  const range = max - min || 1;
  
  const width = 100;
  const height = 30;
  const padding = 2;
  
  const points = data.map((value, index) => {
    const x = (index / (data.length - 1)) * (width - padding * 2) + padding;
    const y = height - ((value - min) / range) * (height - padding * 2) - padding;
    return `${x},${y}`;
  }).join(" ");
  
  const fillPoints = `${padding},${height} ${points} ${width - padding},${height}`;

  return (
    <svg 
      viewBox={`0 0 ${width} ${height}`} 
      className={cn("w-full h-full", className)}
      preserveAspectRatio="none"
    >
      <defs>
        <linearGradient id="sparklineGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.3" />
          <stop offset="100%" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      
      {/* Fill area */}
      <polygon 
        points={fillPoints} 
        fill="url(#sparklineGradient)"
      />
      
      {/* Line */}
      <polyline
        points={points}
        fill="none"
        stroke={color}
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      
      {/* End point */}
      <circle
        cx={width - padding}
        cy={height - ((data[data.length - 1] - min) / range) * (height - padding * 2) - padding}
        r="3"
        fill={color}
      />
    </svg>
  );
}
