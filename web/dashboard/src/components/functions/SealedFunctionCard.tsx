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

interface SealedFunctionCardProps {
  fn: FunctionConfig;
  onView: (id: string) => void;
  onEdit: (id: string) => void;
  onDelete?: (fn: FunctionConfig) => void;
  index?: number;
}

function formatDate(dateStr?: string): string {
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
}

export function SealedFunctionCard({
  fn,
  onView,
  onEdit,
  onDelete,
  index = 0,
}: SealedFunctionCardProps) {
  const [isHovered, setIsHovered] = useState(false);
  const animationDelay = `${index * 0.05}s`;

  const isActive = fn.providers && fn.providers.length > 0;
  const statusColor = isActive ? 'var(--status-ok)' : 'var(--status-pending)';
  const statusText = isActive ? 'ONLINE' : 'DEPLOYING';

  return (
    <div
      className={`sc-function-card sc-animate-fade-in-up sc-stagger-${(index % 6) + 1}`}
      style={{ animationDelay, opacity: 0 }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onClick={() => onView(fn.id)}
    >
      {/* Corner Braces */}
      <div className="sc-function-card__brace-tl" />
      <div className="sc-function-card__brace-br" />

      {/* Header */}
      <div className="sc-function-card__header">
        <div>
          <div className="sc-function-card__id">
            FUNC-ID: {fn.id.slice(0, 8).toUpperCase()}
          </div>
          <div className="sc-function-card__name">{fn.name}</div>
        </div>

        {/* Actions Dropdown */}
        <div className="sc-function-card__actions">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                onClick={(e) => e.stopPropagation()}
                aria-label="Function actions"
              >
                <MoreVertical className="w-4 h-4" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="end"
              className="sc-function-card__dropdown"
              style={{
                background: 'var(--panel-raised)',
                borderColor: 'var(--panel-edge)',
              }}
            >
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  onView(fn.id);
                }}
                className="sc-function-card__dropdown-item"
              >
                <Play className="w-3 h-3" style={{ color: 'var(--status-ok)' }} />
                EXECUTE
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation();
                  onEdit(fn.id);
                }}
                className="sc-function-card__dropdown-item"
              >
                <Edit3 className="w-3 h-3" style={{ color: 'var(--foil-a)' }} />
                MODIFY
              </DropdownMenuItem>
              {onDelete && (
                <>
                  <DropdownMenuSeparator
                    style={{ background: 'var(--panel-edge)' }}
                  />
                  <DropdownMenuItem
                    onClick={(e) => {
                      e.stopPropagation();
                      onDelete(fn);
                    }}
                    className="sc-function-card__dropdown-item sc-function-card__dropdown-item--danger"
                  >
                    <Trash2 className="w-3 h-3" />
                    TERMINATE
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Badges */}
      <div className="sc-function-card__badges">
        <span className="sc-function-card__badge">
          <Layers className="w-3 h-3" style={{ marginRight: '4px' }} />
          EDGE
        </span>
        <span className="sc-function-card__badge">
          V{fn.version || '1.0'}
        </span>
      </div>

      {/* Metrics Grid */}
      <div className="sc-function-card__metrics">
        <div className="sc-function-card__metric">
          <div className="sc-function-card__metric-label">
            <Globe className="w-3 h-3" />
            REGION
          </div>
          <div className="sc-function-card__metric-value sc-function-card__metric-value--highlight">
            {fn.region?.toUpperCase() || 'AUTO'}
          </div>
        </div>

        <div className="sc-function-card__metric">
          <div className="sc-function-card__metric-label">
            <Server className="w-3 h-3" />
            NODES
          </div>
          <div className="sc-function-card__metric-value">
            {fn.providers?.length || 0}
          </div>
        </div>

        <div className="sc-function-card__metric">
          <div className="sc-function-card__metric-label">
            <Activity className="w-3 h-3" />
            STATUS
          </div>
          <div
            className="sc-function-card__metric-value"
            style={{ color: statusColor }}
          >
            {statusText}
          </div>
        </div>

        <div className="sc-function-card__metric">
          <div className="sc-function-card__metric-label">
            <Clock className="w-3 h-3" />
            UPDATED
          </div>
          <div className="sc-function-card__metric-value">
            {formatDate(fn.updatedAt)}
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="sc-function-card__footer">
        <div className="sc-function-card__meta">
          <Zap className="w-3 h-3" />
          {formatDate(fn.createdAt)}
        </div>
        <button
          className="sc-function-card__execute-btn"
          onClick={(e) => {
            e.stopPropagation();
            onView(fn.id);
          }}
        >
          EXECUTE
          <ChevronRight className="w-3 h-3" />
        </button>
      </div>
    </div>
  );
}
