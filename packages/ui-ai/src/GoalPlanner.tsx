/**
 * @functionfly/ui-ai
 * Goal Planner - AI goal decomposition and tracking
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Target, Plus, Trash2, ChevronDown, ChevronRight, CheckCircle2, Circle, AlertCircle } from "lucide-react";

export interface Milestone {
  id: string;
  label: string;
  status: "pending" | "in_progress" | "completed" | "blocked";
  description?: string;
  tasks: Array<{
    id: string;
    label: string;
    completed: boolean;
  }>;
}

export interface Goal {
  id: string;
  title: string;
  description?: string;
  status: "active" | "completed" | "abandoned";
  priority: "critical" | "high" | "medium" | "low";
  progress: number;
  milestones: Milestone[];
  createdAt: number;
  deadline?: number;
}

export interface GoalPlannerProps {
  goals: Goal[];
  onGoalAdd?: (goal: Omit<Goal, "id" | "createdAt">) => void;
  onGoalUpdate?: (id: string, updates: Partial<Goal>) => void;
  onGoalDelete?: (id: string) => void;
  onMilestoneToggle?: (goalId: string, milestoneId: string) => void;
  onTaskToggle?: (goalId: string, milestoneId: string, taskId: string) => void;
  className?: string;
}

const statusConfig = {
  pending: { icon: Circle, color: "text-text-muted" },
  in_progress: { icon: Circle, color: "text-brand-500", fill: "bg-brand-500" },
  completed: { icon: CheckCircle2, color: "text-success" },
  blocked: { icon: AlertCircle, color: "text-error" },
};

const priorityColors = {
  critical: "text-error",
  high: "text-warning",
  medium: "text-brand-500",
  low: "text-text-muted",
};

export function GoalPlanner({
  goals,
  onGoalAdd,
  onGoalUpdate,
  onGoalDelete,
  onMilestoneToggle,
  onTaskToggle,
  className,
}: GoalPlannerProps) {
  const [expandedGoals, setExpandedGoals] = React.useState<Set<string>>(new Set());
  const [showCompleted, setShowCompleted] = React.useState(true);

  const toggleGoal = (id: string) => {
    setExpandedGoals(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const getProgress = (goal: Goal): number => {
    const allTasks = goal.milestones.flatMap(m => m.tasks);
    if (allTasks.length === 0) return goal.progress;
    const completed = allTasks.filter(t => t.completed).length;
    return Math.round((completed / allTasks.length) * 100);
  };

  const activeGoals = goals.filter(g => g.status !== "completed" || showCompleted);

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
        <div className="flex items-center gap-2">
          <Target className="size-4 text-brand-500" />
          <span className="text-sm font-medium text-text-primary">Goal Planner</span>
          <Badge variant="brand" size="sm">{activeGoals.length} goals</Badge>
        </div>
        <button
          onClick={() => setShowCompleted(s => !s)}
          className={cn(
            "text-xs px-2 py-1 rounded transition-colors",
            showCompleted ? "bg-brand-500/10 text-brand-500" : "bg-bg-tertiary text-text-muted"
          )}
        >
          {showCompleted ? "Hide completed" : "Show completed"}
        </button>
      </div>

      {/* Goals List */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {activeGoals.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-text-muted">
            <Target className="size-12 mb-3 opacity-50" />
            <p className="text-sm">No goals yet</p>
            <p className="text-xs mt-1">Create a goal to get started</p>
          </div>
        ) : (
          activeGoals.map(goal => {
            const isExpanded = expandedGoals.has(goal.id);
            const progress = getProgress(goal);

            return (
              <div
                key={goal.id}
                className={cn(
                  "rounded-lg border transition-colors",
                  goal.status === "completed"
                    ? "bg-bg-tertiary/50 border-transparent"
                    : "bg-bg-secondary border-border-subtle"
                )}
              >
                {/* Goal Header */}
                <div
                  onClick={() => toggleGoal(goal.id)}
                  className="flex items-center gap-3 p-3 cursor-pointer hover:bg-bg-hover rounded-t-lg"
                >
                  {isExpanded ? (
                    <ChevronDown className="size-4 text-text-muted shrink-0" />
                  ) : (
                    <ChevronRight className="size-4 text-text-muted shrink-0" />
                  )}

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-text-primary">{goal.title}</span>
                      <span className={cn("text-xs font-medium", priorityColors[goal.priority])}>
                        {goal.priority}
                      </span>
                    </div>
                    {goal.description && (
                      <p className="text-xs text-text-muted mt-0.5 line-clamp-1">{goal.description}</p>
                    )}
                  </div>

                  {/* Progress */}
                  <div className="flex items-center gap-2 shrink-0">
                    <div className="w-20 h-2 rounded-full bg-bg-tertiary overflow-hidden">
                      <div
                        className={cn(
                          "h-full rounded-full transition-all",
                          progress === 100 ? "bg-success" : "bg-brand-500"
                        )}
                        style={{ width: `${progress}%` }}
                      />
                    </div>
                    <span className="text-xs text-text-muted w-8 text-right">{progress}%</span>
                  </div>
                </div>

                {/* Milestones */}
                {isExpanded && (
                  <div className="px-3 pb-3 space-y-2">
                    {goal.milestones.map(milestone => {
                      const config = statusConfig[milestone.status];
                      const Icon = config.icon;
                      const completedTasks = milestone.tasks.filter(t => t.completed).length;

                      return (
                        <div
                          key={milestone.id}
                          className="pl-6 p-2 rounded bg-bg-tertiary/50 border border-border-subtle"
                        >
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => onMilestoneToggle?.(goal.id, milestone.id)}
                              className={cn("shrink-0", config.color)}
                            >
                              <Icon className="size-4" />
                            </button>
                            <span className="text-sm text-text-primary flex-1">{milestone.label}</span>
                            <Badge variant="ghost" size="sm">
                              {completedTasks}/{milestone.tasks.length}
                            </Badge>
                          </div>

                          {/* Tasks */}
                          <div className="pl-6 mt-2 space-y-1">
                            {milestone.tasks.map(task => (
                              <div
                                key={task.id}
                                onClick={() => onTaskToggle?.(goal.id, milestone.id, task.id)}
                                className="flex items-center gap-2 py-1 cursor-pointer hover:bg-bg-hover rounded px-2"
                              >
                                <div className={cn(
                                  "size-4 rounded border flex items-center justify-center",
                                  task.completed
                                    ? "bg-success border-success text-white"
                                    : "border-border-subtle"
                                )}>
                                  {task.completed && <CheckCircle2 className="size-3" />}
                                </div>
                                <span className={cn(
                                  "text-xs",
                                  task.completed ? "text-text-muted line-through" : "text-text-secondary"
                                )}>
                                  {task.label}
                                </span>
                              </div>
                            ))}
                          </div>
                        </div>
                      );
                    })}

                    {/* Add Milestone */}
                    <button
                      onClick={() => {
                        // Add new milestone logic would go here
                      }}
                      className="flex items-center gap-2 pl-6 py-1 text-xs text-text-muted hover:text-text-primary transition-colors"
                    >
                      <Plus className="size-3" />
                      Add milestone
                    </button>
                  </div>
                )}
              </div>
            );
          })
        )}

        {/* Add Goal Button */}
        <button
          onClick={() => onGoalAdd?.({
            title: "New Goal",
            status: "active",
            priority: "medium",
            progress: 0,
            milestones: [],
          })}
          className="w-full flex items-center justify-center gap-2 p-3 rounded-lg border border-dashed border-border-subtle text-text-muted hover:text-text-primary hover:border-brand-500/30 transition-colors"
        >
          <Plus className="size-4" />
          <span className="text-sm">Add Goal</span>
        </button>
      </div>
    </div>
  );
}
