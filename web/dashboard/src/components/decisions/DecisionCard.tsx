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
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
    <Card
      className={cn(
        "group relative transition-all hover:border-border/80",
        decision.status === "pending" && "border-yellow-500/20"
      )}
    >
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <DecisionStatusBadge status={decision.status} />
              {decision.source_type === "ai_extracted" && (
                <Badge
                  variant="outline"
                  className="bg-purple-500/10 text-purple-500 border-purple-500/20 text-xs"
                >
                  AI Extracted
                </Badge>
              )}
            </div>
            <CardTitle
              className={cn(
                "text-lg cursor-pointer hover:text-primary transition-colors",
                onView && "underline-offset-2 hover:underline"
              )}
              onClick={() => onView?.(decision)}
            >
              {decision.title}
            </CardTitle>
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
                <DropdownMenuItem
                  onClick={() => onEdit(decision)}
                >
                  <Pencil className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
              )}
              {canApprove && (
                <DropdownMenuItem
                  onClick={() => onApprove(decision)}
                >
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
          <CardDescription className="line-clamp-2 mt-2">
            {decision.description}
          </CardDescription>
        )}
      </CardHeader>

      <CardContent className="pb-2">
        {decision.tags && decision.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {decision.tags.slice(0, 5).map((tag) => (
              <Badge
                key={tag}
                variant="secondary"
                className="text-xs"
              >
                {tag}
              </Badge>
            ))}
            {decision.tags.length > 5 && (
              <Badge variant="secondary" className="text-xs">
                +{decision.tags.length - 5}
              </Badge>
            )}
          </div>
        )}

        {decision.rationale && isExpanded && (
          <div className="mt-3 pt-3 border-t">
            <p className="text-sm font-medium text-muted-foreground mb-1">
              Rationale
            </p>
            <p className="text-sm">{decision.rationale}</p>
          </div>
        )}

        {decision.outcome && isExpanded && (
          <div className="mt-3 pt-3 border-t">
            <p className="text-sm font-medium text-muted-foreground mb-1">
              Outcome
            </p>
            <p className="text-sm">{decision.outcome}</p>
          </div>
        )}

        {decision.alternatives && decision.alternatives.length > 0 && isExpanded && (
          <div className="mt-3 pt-3 border-t">
            <p className="text-sm font-medium text-muted-foreground mb-1">
              Alternatives Considered
            </p>
            <ul className="text-sm space-y-1">
              {decision.alternatives.map((alt, i) => (
                <li key={i} className="flex items-start gap-2">
                  <ArrowRight className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
                  {alt}
                </li>
              ))}
            </ul>
          </div>
        )}
      </CardContent>

      <CardFooter className="pt-2 text-xs text-muted-foreground">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {formatDistanceToNow(new Date(decision.created_at), {
              addSuffix: true,
            })}
          </div>

          {decision.approved_by && decision.approved_at && (
            <div className="flex items-center gap-1 text-green-500">
              <CheckCircle className="h-3 w-3" />
              Approved{" "}
              {formatDistanceToNow(new Date(decision.approved_at), {
                addSuffix: true,
              })}
            </div>
          )}

          {(decision.rationale || decision.outcome || decision.alternatives) && (
            <Button
              variant="ghost"
              size="sm"
              className="ml-auto h-auto p-0 text-xs"
              onClick={() => setIsExpanded(!isExpanded)}
            >
              {isExpanded ? "Show less" : "Show more"}
            </Button>
          )}
        </div>
      </CardFooter>
    </Card>
  );
}