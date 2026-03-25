import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type { FunctionConfig } from '@/types';
import {
  Activity,
  ChevronRight,
  Clock,
  Cpu,
  Edit3,
  Globe,
  Layers,
  MoreVertical,
  Play,
  Server,
  Trash2,
  Zap,
} from 'lucide-react';
import { useState } from 'react';

interface AviationFunctionCardProps {
  fn: FunctionConfig;
  onView: (id: string) => void;
  onEdit: (id: string) => void;
  onDelete: (fn: FunctionConfig) => void;
  index?: number;
}

/**
 * Aviation-themed function card styled like aircraft instrumentation
 * Theme-aware: adapts to light/dark mode
 */
export function AviationFunctionCard({
  fn,
  onView,
  onEdit,
  onDelete,
  index = 0,
}: AviationFunctionCardProps) {
  const [isHovered, setIsHovered] = useState(false);

  // Calculate stagger delay based on index
  const animationDelay = `${index * 0.1}s`;

  // Format runtime display
  const runtimeDisplay = 'EDGE';

  // Determine status color based on providers
  const isActive = fn.providers && fn.providers.length > 0;
  const statusColor = isActive ? 'var(--color-aviation-green)' : 'var(--color-aviation-amber)';
  const statusText = isActive ? 'ONLINE' : 'DEPLOYING';

  // Format date
  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '--';
    const date = new Date(dateStr);
    return date
      .toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
      .toUpperCase();
  };

  return (
    <div
      className="aviation-instrument group cursor-pointer aviation-animate-fade-in-up"
      style={{ animationDelay, opacity: 0 }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onClick={() => onView(fn.id)}
    >
      {/* Corner decorations */}
      <div className="aviation-instrument-corner aviation-instrument-corner-tl" />
      <div className="aviation-instrument-corner aviation-instrument-corner-tr" />
      <div className="aviation-instrument-corner aviation-instrument-corner-bl" />
      <div className="aviation-instrument-corner aviation-instrument-corner-br" />

      <div className="p-5 relative">
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-3">
            {/* Status LED */}
            <div
              className={`aviation-status-led ${isActive ? 'active' : 'warning'}`}
              style={{
                boxShadow: isActive
                  ? '0 0 8px var(--color-aviation-green-glow), inset 0 1px 2px rgba(0,0,0,0.3)'
                  : '0 0 8px var(--color-aviation-amber-glow), inset 0 1px 2px rgba(0,0,0,0.3)',
                background: statusColor,
              }}
            />

            <div>
              {/* Function ID label */}
              <div className="aviation-label mb-0.5">
                FUNC-ID: {fn.id.slice(0, 8).toUpperCase()}
              </div>

              {/* Function name */}
              <h4
                className="font-mono font-semibold text-sm tracking-wide uppercase"
                style={{ color: 'var(--color-aviation-text-primary)' }}
              >
                {fn.name}
              </h4>
            </div>
          </div>

          {/* Actions dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                className="p-1.5 rounded transition-colors hover:bg-(--color-aviation-border-panel)"
                onClick={(e) => e.stopPropagation()}
              >
                <MoreVertical
                  className="w-4 h-4"
                  style={{ color: 'var(--color-aviation-text-muted)' }}
                />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="end"
              className="aviation-panel border-(--color-aviation-amber-dim)"
              style={{
                background: 'var(--color-aviation-bg-tertiary)',
                borderColor: 'var(--color-aviation-amber-dim)',
              }}
            >
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  onView(fn.id);
                }}
                className="cursor-pointer font-mono text-xs flex items-center gap-2"
                style={{ color: 'var(--color-aviation-text-primary)' }}
              >
                <Play className="w-3 h-3" style={{ color: 'var(--color-aviation-green)' }} />
                EXECUTE
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  onEdit(fn.id);
                }}
                className="cursor-pointer font-mono text-xs flex items-center gap-2"
                style={{ color: 'var(--color-aviation-text-primary)' }}
              >
                <Edit3 className="w-3 h-3" style={{ color: 'var(--color-aviation-cyan)' }} />
                MODIFY
              </DropdownMenuItem>
              <DropdownMenuSeparator style={{ background: 'var(--color-aviation-border-panel)' }} />
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete(fn);
                }}
                className="cursor-pointer font-mono text-xs flex items-center gap-2"
                style={{ color: 'var(--color-aviation-red)' }}
              >
                <Trash2 className="w-3 h-3" />
                TERMINATE
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Runtime badge */}
        <div className="flex items-center gap-2 mb-4">
          <div
            className="flex items-center gap-1.5 px-2 py-1 rounded border"
            style={{
              background: 'var(--color-aviation-cyan-subtle)',
              borderColor: 'var(--color-aviation-cyan-dim)',
            }}
          >
            <Cpu className="w-3 h-3" style={{ color: 'var(--color-aviation-cyan)' }} />
            <span
              className="font-mono text-xs font-semibold"
              style={{ color: 'var(--color-aviation-cyan)' }}
            >
              {runtimeDisplay}
            </span>
          </div>

          <div
            className="flex items-center gap-1.5 px-2 py-1 rounded border"
            style={{
              background: 'rgba(16, 185, 129, 0.1)',
              borderColor: 'rgba(16, 185, 129, 0.2)',
            }}
          >
            <Layers className="w-3 h-3" style={{ color: 'var(--color-aviation-green)' }} />
            <span
              className="font-mono text-xs font-semibold"
              style={{ color: 'var(--color-aviation-green)' }}
            >
              V{fn.version || '1.0'}
            </span>
          </div>
        </div>

        {/* Metrics grid */}
        <div className="grid grid-cols-2 gap-3 mb-4">
          {/* Region */}
          <div
            className="aviation-data-row p-2! border-b-0! rounded"
            style={{ background: 'rgba(0,0,0,0.2)' }}
          >
            <div className="flex items-center gap-1.5">
              <Globe className="w-3 h-3" style={{ color: 'var(--color-aviation-text-muted)' }} />
              <span className="aviation-data-label">REGION</span>
            </div>
            <span
              className="aviation-data-value text-xs font-semibold uppercase"
              style={{ color: 'var(--color-aviation-cyan)' }}
            >
              {fn.region?.toUpperCase() || 'AUTO'}
            </span>
          </div>

          {/* Providers */}
          <div
            className="aviation-data-row p-2! border-b-0! rounded"
            style={{ background: 'rgba(0,0,0,0.2)' }}
          >
            <div className="flex items-center gap-1.5">
              <Server className="w-3 h-3" style={{ color: 'var(--color-aviation-text-muted)' }} />
              <span className="aviation-data-label">NODES</span>
            </div>
            <span
              className="aviation-data-value text-xs font-semibold"
              style={{ color: 'var(--color-aviation-text-primary)' }}
            >
              {fn.providers?.length || 0}
            </span>
          </div>

          {/* Status */}
          <div
            className="aviation-data-row p-2! border-b-0! rounded"
            style={{ background: 'rgba(0,0,0,0.2)' }}
          >
            <div className="flex items-center gap-1.5">
              <Activity className="w-3 h-3" style={{ color: 'var(--color-aviation-text-muted)' }} />
              <span className="aviation-data-label">STATUS</span>
            </div>
            <span className="font-mono text-xs font-semibold" style={{ color: statusColor }}>
              {statusText}
            </span>
          </div>

          {/* Last updated */}
          <div
            className="aviation-data-row p-2! border-b-0! rounded"
            style={{ background: 'rgba(0,0,0,0.2)' }}
          >
            <div className="flex items-center gap-1.5">
              <Clock className="w-3 h-3" style={{ color: 'var(--color-aviation-text-muted)' }} />
              <span className="aviation-data-label">UPDATED</span>
            </div>
            <span
              className="aviation-data-value text-xs"
              style={{ color: 'var(--color-aviation-text-secondary)' }}
            >
              {formatDate(fn.updatedAt)}
            </span>
          </div>
        </div>

        {/* Provider tags */}
        <div className="flex flex-wrap gap-1.5 mb-4">
          {fn.providers?.slice(0, 3).map((provider) => (
            <span
              key={provider}
              className="px-1.5 py-0.5 rounded font-mono text-[10px] border"
              style={{
                background: 'var(--color-aviation-amber-subtle)',
                borderColor: 'var(--color-aviation-amber-dim)',
                color: 'var(--color-aviation-amber)',
              }}
            >
              {provider.toUpperCase()}
            </span>
          ))}
          {fn.providers && fn.providers.length > 3 && (
            <span
              className="px-1.5 py-0.5 font-mono text-[10px]"
              style={{ color: 'var(--color-aviation-text-muted)' }}
            >
              +{fn.providers.length - 3}
            </span>
          )}
        </div>

        {/* Footer with action hint */}
        <div
          className={`flex items-center justify-between pt-3 border-t transition-all duration-300 ${
            isHovered ? 'opacity-100' : 'opacity-60'
          }`}
          style={{ borderColor: 'var(--color-aviation-border-panel)' }}
        >
          <div className="flex items-center gap-1.5">
            <Zap className="w-3 h-3" style={{ color: 'var(--color-aviation-amber)' }} />
            <span className="aviation-label text-[10px]">CLICK TO EXECUTE</span>
          </div>

          <ChevronRight
            className={`w-4 h-4 transition-transform duration-300 ${
              isHovered ? 'translate-x-1' : ''
            }`}
            style={{ color: 'var(--color-aviation-amber)' }}
          />
        </div>
      </div>

      {/* Hover glow effect - amber */}
      <div
        className={`absolute inset-0 rounded-xl transition-opacity duration-300 pointer-events-none ${
          isHovered ? 'opacity-100' : 'opacity-0'
        }`}
        style={{
          background:
            'radial-gradient(ellipse at center, var(--color-aviation-amber-subtle) 0%, transparent 70%)',
        }}
      />
    </div>
  );
}
