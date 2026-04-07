import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';
import { format } from 'date-fns';

interface UptimeSegment {
  date: string;
  status: 'operational' | 'degraded' | 'outage' | 'maintenance';
  uptime: number;
  incidents?: string[];
}

interface UptimeBarProps {
  segments: UptimeSegment[];
  className?: string;
  showTooltip?: boolean;
}

const statusColors = {
  operational: {
    bg: 'bg-emerald-500',
    hover: 'hover:bg-emerald-400',
    glow: 'shadow-emerald-500/50',
    label: 'Operational',
    description: 'All systems running normally',
  },
  degraded: {
    bg: 'bg-amber-500',
    hover: 'hover:bg-amber-400',
    glow: 'shadow-amber-500/50',
    label: 'Degraded Performance',
    description: 'Slower response times detected',
  },
  outage: {
    bg: 'bg-red-500',
    hover: 'hover:bg-red-400',
    glow: 'shadow-red-500/50',
    label: 'Major Outage',
    description: 'Service disruption in progress',
  },
  maintenance: {
    bg: 'bg-purple-500',
    hover: 'hover:bg-purple-400',
    glow: 'shadow-purple-500/50',
    label: 'Scheduled Maintenance',
    description: 'Planned maintenance window',
  },
};

export function UptimeBar({ segments, className, showTooltip = true }: UptimeBarProps) {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  // Calculate overall uptime
  const operationalSegments = segments.filter(s => s.status === 'operational').length;
  const uptime = Math.round((operationalSegments / segments.length) * 100 * 100) / 100;

  // Calculate day counts for each status
  const statusCounts = segments.reduce((acc, segment) => {
    acc[segment.status] = (acc[segment.status] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const handleMouseMove = (_e: React.MouseEvent, index: number) => {
    setHoveredIndex(index);
  };

  return (
    <div className={cn('space-y-4', className)}>
      {/* Stats header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <span className="text-sm text-text-secondary">
            {segments.length} days history
          </span>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-text-muted">Uptime:</span>
            <motion.span 
              className="font-semibold text-emerald-400"
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              {uptime}%
            </motion.span>
          </div>
        </div>
        
        {/* Status breakdown */}
        <div className="hidden sm:flex items-center gap-3 text-xs">
          {Object.entries(statusCounts).map(([status, count]) => (
            <div key={status} className="flex items-center gap-1.5">
              <div className={cn('w-2 h-2 rounded-full', statusColors[status as keyof typeof statusColors]?.bg)} />
              <span className="text-text-muted capitalize">{status}</span>
              <span className="text-text-secondary">({count})</span>
            </div>
          ))}
        </div>
      </div>

      {/* Uptime bar */}
      <div className="relative">
        <div 
          className={cn(
            'flex h-8 md:h-10 rounded-lg overflow-hidden',
            'bg-bg-tertiary border border-border-subtle',
            'p-1 gap-[2px]'
          )}
        >
          {segments.map((segment, index) => {
            const status = statusColors[segment.status];
            const isHovered = hoveredIndex === index;

            return (
              <motion.div
                key={index}
                initial={{ scaleY: 0 }}
                animate={{ scaleY: 1 }}
                transition={{ duration: 0.3, delay: index * 0.01 }}
                className={cn(
                  'flex-1 min-w-[3px] rounded-sm cursor-pointer relative',
                  'transition-all duration-150 ease-out',
                  status.bg,
                  status.hover,
                  isHovered && 'scale-y-125 z-10',
                  isHovered && 'shadow-sm',
                  isHovered && status.glow
                )}
                onMouseEnter={(e) => handleMouseMove(e, index)}
                onMouseLeave={() => setHoveredIndex(null)}
              >
                {/* Subtle inner glow on hover */}
                <AnimatePresence>
                  {isHovered && (
                    <motion.div
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      className={cn(
                        'absolute inset-0 rounded-sm',
                        'bg-gradient-to-t from-black/20 to-transparent'
                      )}
                    />
                  )}
                </AnimatePresence>
              </motion.div>
            );
          })}
        </div>

        {/* Custom tooltip */}
        <AnimatePresence>
          {showTooltip && hoveredIndex !== null && (
            <motion.div
              initial={{ opacity: 0, y: 10, scale: 0.95 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: 5, scale: 0.95 }}
              transition={{ duration: 0.15 }}
              className="absolute z-50 bottom-full mb-2 pointer-events-none"
              style={{
                left: `${(hoveredIndex / segments.length) * 100}%`,
                transform: 'translateX(-50%)',
              }}
            >
              <div 
                className={cn(
                  'bg-bg-glass-strong backdrop-blur-xl',
                  'border border-border-default rounded-xl',
                  'p-3 shadow-xl whitespace-nowrap',
                  'min-w-[200px]'
                )}
              >
                <div className="flex items-center gap-2 mb-2">
                  <div 
                    className={cn(
                      'w-2.5 h-2.5 rounded-full shadow-lg',
                      statusColors[segments[hoveredIndex].status].bg,
                      statusColors[segments[hoveredIndex].status].glow
                    )} 
                  />
                  <span className="font-semibold text-text-primary text-sm">
                    {statusColors[segments[hoveredIndex].status].label}
                  </span>
                </div>
                <p className="text-xs text-text-secondary mb-2">
                  {statusColors[segments[hoveredIndex].status].description}
                </p>
                <div className="space-y-1 pt-2 border-t border-border-subtle">
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-text-muted">Date</span>
                    <span className="text-text-secondary font-mono">
                      {segments[hoveredIndex].date}
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-text-muted">Uptime</span>
                    <span className={cn(
                      'font-semibold',
                      segments[hoveredIndex].uptime >= 99 
                        ? 'text-emerald-400' 
                        : segments[hoveredIndex].uptime >= 95 
                          ? 'text-amber-400' 
                          : 'text-red-400'
                    )}>
                      {segments[hoveredIndex].uptime.toFixed(2)}%
                    </span>
                  </div>
                </div>
              </div>
              {/* Arrow */}
              <div className="absolute top-full left-1/2 -translate-x-1/2 -mt-1">
                <div className="w-2 h-2 bg-bg-glass-strong border-r border-b border-border-default rotate-45" />
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap items-center gap-4 text-xs text-text-muted">
        {Object.entries(statusColors).map(([key, config]) => (
          <div 
            key={key} 
            className="flex items-center gap-1.5 group cursor-default"
          >
            <div 
              className={cn(
                'w-2.5 h-2.5 rounded-sm transition-transform duration-200',
                config.bg,
                'group-hover:scale-125'
              )} 
            />
            <span className="group-hover:text-text-secondary transition-colors">
              {config.label}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function UptimeMiniBar({ 
  days = 30, 
  uptime = 99.9,
  className 
}: { 
  days?: number; 
  uptime?: number;
  className?: string;
}) {
  // Generate segments based on uptime percentage
  const failedDays = Math.round((100 - uptime) / 100 * days);
  const segments = Array.from({ length: days }, (_, i) => {
    const isFailed = i < failedDays;
    const randomStatus = isFailed 
      ? (Math.random() > 0.3 ? 'outage' : 'degraded')
      : 'operational';
    return {
      date: format(new Date(Date.now() - (days - i - 1) * 24 * 60 * 60 * 1000), 'MMM dd'),
      status: randomStatus as 'operational' | 'degraded' | 'outage' | 'maintenance',
      uptime: isFailed ? Math.random() * 95 : 99 + Math.random(),
    };
  });

  return (
    <div className={cn('flex items-center gap-3', className)}>
      <div className="flex gap-[2px] h-2 rounded overflow-hidden bg-bg-tertiary flex-1">
        {segments.map((segment, index) => (
          <motion.div
            key={index}
            initial={{ scaleY: 0 }}
            animate={{ scaleY: 1 }}
            transition={{ duration: 0.2, delay: index * 0.02 }}
            className={cn(
              'flex-1 min-w-[2px]',
              statusColors[segment.status].bg,
              'hover:opacity-80 transition-opacity'
            )}
            title={`${segment.date}: ${segment.status} (${segment.uptime.toFixed(2)}%)`}
          />
        ))}
      </div>
      <span className="text-xs text-text-muted font-mono tabular-nums">
        {days}d
      </span>
    </div>
  );
}

export function UptimeSparkline({ 
  data,
  className 
}: { 
  data: number[];
  className?: string;
}) {
  const min = Math.min(...data);
  const max = Math.max(...data);
  const range = max - min || 1;
  
  const points = data.map((value, index) => {
    const x = (index / (data.length - 1)) * 100;
    const y = 100 - ((value - min) / range) * 100;
    return `${x},${y}`;
  }).join(' ');

  return (
    <div className={cn('relative h-12 w-full', className)}>
      <svg 
        viewBox="0 0 100 100" 
        preserveAspectRatio="none"
        className="absolute inset-0 w-full h-full overflow-visible"
      >
        <defs>
          <linearGradient id="sparklineGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="rgba(99, 102, 241, 0.3)" />
            <stop offset="100%" stopColor="rgba(99, 102, 241, 0)" />
          </linearGradient>
        </defs>
        
        {/* Area fill */}
        <polygon
          points={`0,100 ${points} 100,100`}
          fill="url(#sparklineGradient)"
          className="transition-all duration-300"
        />
        
        {/* Line */}
        <polyline
          points={points}
          fill="none"
          stroke="rgb(99, 102, 241)"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="transition-all duration-300"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
    </div>
  );
}
