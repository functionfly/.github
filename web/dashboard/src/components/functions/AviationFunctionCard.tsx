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
  onDelete?: (fn: FunctionConfig) => void;
  index?: number;
  isNewStyle?: boolean;
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
  isNewStyle = false,
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

  // New style specific classes
  const cardClass = isNewStyle ? 'new-function-card new-animate-fade-in-up' : 'aviation-function-card aviation-animate-fade-in-up';
  const staggerClass = isNewStyle ? `new-stagger-${(index % 6) + 1}` : '';

  return (
    <div
      className={`${cardClass} ${staggerClass}`}
      style={{ animationDelay, opacity: 0 }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onClick={() => onView(fn.id)}
    >
      {isNewStyle && (
        <>
          {/* New badge */}
          <span className="new-badge">NEW</span>
          {/* Corner decorations */}
          <div className="corner-decoration corner-tl" />
          <div className="corner-decoration corner-tr" />
          <div className="corner-decoration corner-bl" />
          <div className="corner-decoration corner-br" />
          {/* Hover glow */}
          <div className="hover-glow" />
        </>
      )}

      {!isNewStyle && (
        <>
          {/* Corner decorations */}
          <div className="aviation-card-corner aviation-card-corner-tl" />
          <div className="aviation-card-corner aviation-card-corner-tr" />
          <div className="aviation-card-corner aviation-card-corner-bl" />
          <div className="aviation-card-corner aviation-card-corner-br" />
          {/* Hover glow effect */}
          <div className="aviation-card-glow" />
        </>
      )}

      <div className={isNewStyle ? 'card-body relative' : 'aviation-card-body relative'}>
        {/* Header */}
        <div className={isNewStyle ? 'card-header' : 'aviation-card-header'}>
          <div className="flex items-center gap-3">
            {/* Status LED */}
            {!isNewStyle && <div className={`aviation-status-led ${isActive ? 'active' : 'warning'}`} />}

            <div>
              {/* Function ID label */}
              <div className={isNewStyle ? 'card-id' : 'aviation-card-id'}>
                FUNC-ID: {fn.id.slice(0, 8).toUpperCase()}
              </div>

              {/* Function name */}
              <h4 className={isNewStyle ? 'card-title' : 'aviation-card-title'}>
                {fn.name}
              </h4>
            </div>
          </div>

          {/* Actions dropdown - only for non-new styles */}
          {!isNewStyle && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  className="p-1.5 rounded transition-colors hover:bg-[var(--color-aviation-border-panel)]"
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
                className="aviation-panel"
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
                {onDelete && (
                  <>
                    <DropdownMenuSeparator style={{ background: 'var(--color-aviation-border-panel)' }} />
                    <DropdownMenuItem
                      onClick={(e) => {
                        e.stopPropagation();
                        onDelete(fn);
                      }}
                      className="cursor-pointer font-mono text-xs flex items-center gap-2 focus!bg-red-500/30"
                    >
                      <Trash2 className="w-3 h-3" />
                      <span style={{ color: '#ef4444' }} className="focus:!text-white">TERMINATE</span>
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>

        {/* Badges */}
        <div className={isNewStyle ? 'card-meta' : 'flex items-center gap-2 mb-4'}>
          {isNewStyle ? (
            <>
              <span className="meta-badge">{runtimeDisplay}</span>
              <span className="meta-badge">V{fn.version || '1.0'}</span>
            </>
          ) : (
            <>
              <span className="aviation-badge aviation-badge-runtime">
                <Cpu className="w-3 h-3" />
                {runtimeDisplay}
              </span>
              <span className="aviation-badge aviation-badge-version">
                <Layers className="w-3 h-3" />
                V{fn.version || '1.0'}
              </span>
            </>
          )}
        </div>

        {/* Metrics grid */}
        <div className={isNewStyle ? 'stats-grid' : 'aviation-metrics-grid'}>
          {/* Region */}
          <div className={isNewStyle ? 'stat-item' : 'aviation-metric'}>
            <div className={isNewStyle ? 'stat-label' : 'aviation-metric-label'}>
              {!isNewStyle && <Globe className="w-3 h-3" />}
              REGION
            </div>
            <span className={isNewStyle ? 'stat-value highlight' : 'aviation-metric-value highlight'}>
              {fn.region?.toUpperCase() || 'AUTO'}
            </span>
          </div>

          {/* Providers */}
          <div className={isNewStyle ? 'stat-item' : 'aviation-metric'}>
            <div className={isNewStyle ? 'stat-label' : 'aviation-metric-label'}>
              {!isNewStyle && <Server className="w-3 h-3" />}
              NODES
            </div>
            <span className={isNewStyle ? 'stat-value' : 'aviation-metric-value'}>
              {fn.providers?.length || 0}
            </span>
          </div>

          {/* Status */}
          <div className={isNewStyle ? 'stat-item' : 'aviation-metric'}>
            <div className={isNewStyle ? 'stat-label' : 'aviation-metric-label'}>
              {!isNewStyle && <Activity className="w-3 h-3" />}
              STATUS
            </div>
            <span className={isNewStyle ? 'stat-value' : 'aviation-metric-value'} style={!isNewStyle ? { color: statusColor } : undefined}>
              {statusText}
            </span>
          </div>

          {/* Last updated */}
          <div className={isNewStyle ? 'stat-item' : 'aviation-metric'}>
            <div className={isNewStyle ? 'stat-label' : 'aviation-metric-label'}>
              {!isNewStyle && <Clock className="w-3 h-3" />}
              UPDATED
            </div>
            <span className={isNewStyle ? 'stat-value' : 'aviation-metric-value'}>
              {formatDate(fn.updatedAt)}
            </span>
          </div>
        </div>

        {/* Provider tags */}
        {!isNewStyle && (
          <div className="aviation-provider-tags">
            {fn.providers?.slice(0, 3).map((provider) => (
              <span
                key={provider}
                className={`aviation-provider-tag ${provider === 'registry' ? 'registry' : ''}`}
              >
                {provider.toUpperCase()}
              </span>
            ))}
            {fn.providers && fn.providers.length > 3 && (
              <span className="aviation-provider-more">
                +{fn.providers.length - 3}
              </span>
            )}
          </div>
        )}

        {/* Footer */}
        <div className={isNewStyle ? 'card-footer' : 'aviation-card-footer'}>
          <div className={isNewStyle ? 'created-at' : 'aviation-card-hint'}>
            {isNewStyle ? (
              <>
                <Clock className="w-3 h-3" />
                {formatDate(fn.createdAt)}
              </>
            ) : (
              <>
                <Zap className="w-3 h-3" />
                CLICK TO EXECUTE
              </>
            )}
          </div>
          {isNewStyle ? (
            <button className="action-btn">
              EXECUTE
              <ChevronRight className="w-3 h-3" />
            </button>
          ) : (
            <ChevronRight className="aviation-card-arrow" />
          )}
        </div>
      </div>
    </div>
  );
}
