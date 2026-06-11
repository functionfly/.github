import { useState } from "react";
import { formatDistanceToNow } from "date-fns";
import {
  CheckCircle,
  AlertCircle,
  Clock,
  MoreHorizontal,
  Pencil,
  Trash2,
  ArrowRight,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { DecisionStatusBadge } from "./DecisionStatusBadge";
import type { TeamDecision, DecisionStatus } from "@/api/decisions";
import { cn } from "@/lib/utils";

interface DecisionCardProps {
  decision: TeamDecision;
  teamId: string;
  onEdit?: (decision: TeamDecision) => void;
  onDelete?: (decisionId: string) => void;
  onApprove?: (decision: TeamDecision) => void;
  onView?: (decision: TeamDecision) => void;
  showTeam?: boolean;
}

export function DecisionCard({
  decision,
  teamId,
  onEdit,
  onDelete,
  onApprove,
  onView,
  showTeam = false,
}: DecisionCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const canApprove =
    decision.status === "pending" &&
    onApprove;

  return (
    <div className={cn("decisions-card", decision.status)}>
      {/* Header */}
      <div className="decisions-card-header">
        <div className="flex-1 min-w-0">
          <div className="decisions-card-title-row">
            <DecisionStatusBadge status={decision.status} />
            {decision.source_type === "ai_extracted" && (
              <span className="decisions-ai-badge">
                AI Extracted
              </span>
            )}
          </div>
          <h3
            className="decisions-card-title"
            onClick={() => onView?.(decision)}
          >
            {decision.title}
          </h3>
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {onEdit && (
              <DropdownMenuItem onClick={() => onEdit(decision)}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </DropdownMenuItem>
            )}
            {canApprove && (
              <DropdownMenuItem onClick={() => onApprove(decision)}>
                <CheckCircle className="mr-2 h-4 w-4" />
                Approve
              </DropdownMenuItem>
            )}
            {onDelete && (
              <DropdownMenuItem
                onClick={() => onDelete(decision.id)}
                className="text-destructive"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {decision.description && (
        <p className="decisions-card-description">
          {decision.description}
        </p>
      )}

      {/* Tags */}
      {decision.tags && decision.tags.length > 0 && (
        <div className="decisions-card-tags">
          {decision.tags.slice(0, 5).map((tag) => (
            <span key={tag} className="decisions-card-tag">
              {tag}
            </span>
          ))}
          {decision.tags.length > 5 && (
            <span className="decisions-card-tag decisions-card-tag-more">
              +{decision.tags.length - 5}
            </span>
          )}
        </div>
      )}

      {/* Expanded Content */}
      {decision.rationale && isExpanded && (
        <div className="decisions-card-expanded">
          <p className="decisions-card-expanded-title">Rationale</p>
          <p className="decisions-card-expanded-text">{decision.rationale}</p>
        </div>
      )}

      {decision.outcome && isExpanded && (
        <div className="decisions-card-expanded">
          <p className="decisions-card-expanded-title">Outcome</p>
          <p className="decisions-card-expanded-text">{decision.outcome}</p>
        </div>
      )}

      {decision.alternatives && decision.alternatives.length > 0 && isExpanded && (
        <div className="decisions-card-alternatives">
          <p className="decisions-card-alternatives-title">Alternatives Considered</p>
          <ul className="decisions-card-alternatives-list">
            {decision.alternatives.map((alt, i) => (
              <li key={i} className="decisions-card-alternatives-item">
                <ArrowRight className="h-4 w-4" />
                {alt}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Footer */}
      <div className="decisions-card-footer">
        <div className="decisions-card-timestamp">
          <Clock className="h-3 w-3" />
          {formatDistanceToNow(new Date(decision.created_at), {
            addSuffix: true,
          })}
        </div>

        {decision.approved_by && decision.approved_at && (
          <div className="decisions-card-approved">
            <CheckCircle className="h-3 w-3" />
            Approved{" "}
            {formatDistanceToNow(new Date(decision.approved_at), {
              addSuffix: true,
            })}
          </div>
        )}

        {(decision.rationale || decision.outcome || decision.alternatives) && (
          <button
            className="decisions-show-more-btn"
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {isExpanded ? "Show less" : "Show more"}
          </button>
        )}
      </div>
    </div>
  );
}