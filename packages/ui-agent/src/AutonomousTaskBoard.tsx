/**
 * @functionfly/ui-agent
 * Autonomous Task Board - task creation, assignment, and execution tracking
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";

export interface Task {
  id: string;
  title: string;
  description: string;
  status: "todo" | "in-progress" | "review" | "done" | "blocked";
  priority: "low" | "medium" | "high" | "critical";
  assignedAgent?: {
    id: string;
    name: string;
    isAI: boolean;
  };
  createdAt: string;
  updatedAt: string;
  dueDate?: string;
  tags?: string[];
  subtasks?: Task[];
  executionTrace?: Array<{
    id: string;
    action: string;
    timestamp: string;
    agent?: string;
    result?: "success" | "failure" | "pending";
  }>;
}

export interface AutonomousTaskBoardProps {
  tasks: Task[];
  agents: Array<{ id: string; name: string; isAI: boolean }>;
  onTaskCreate?: (task: Omit<Task, "id" | "createdAt" | "updatedAt">) => void;
  onTaskUpdate?: (taskId: string, updates: Partial<Task>) => void;
  onTaskDelete?: (taskId: string) => void;
  onTaskAssign?: (taskId: string, agentId: string | null) => void;
  className?: string;
}

const priorityConfig = {
  low: { color: "#6b7280", label: "Low" },
  medium: { color: "#3b82f6", label: "Medium" },
  high: { color: "#f97316", label: "High" },
  critical: { color: "#ef4444", label: "Critical" },
};

const statusConfig = {
  todo: { color: "#6b7280", label: "To Do", bg: "bg-bg-secondary" },
  "in-progress": { color: "#3b82f6", label: "In Progress", bg: "bg-blue-500/10" },
  review: { color: "#f59e0b", label: "Review", bg: "bg-yellow-500/10" },
  done: { color: "#10b981", label: "Done", bg: "bg-green-500/10" },
  blocked: { color: "#ef4444", label: "Blocked", bg: "bg-red-500/10" },
};

const statusOrder = ["todo", "in-progress", "review", "done", "blocked"] as const;

export function AutonomousTaskBoard({
  tasks,
  agents,
  onTaskCreate,
  onTaskUpdate,
  onTaskDelete,
  onTaskAssign,
  className,
}: AutonomousTaskBoardProps) {
  const [selectedTask, setSelectedTask] = React.useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = React.useState(false);
  const [newTask, setNewTask] = React.useState<Partial<Task>>({
    title: "",
    description: "",
    status: "todo" as const,
    priority: "medium" as const,
  });

  const tasksByStatus = statusOrder.reduce(
    (acc, status) => {
      acc[status] = tasks.filter((t) => t.status === status);
      return acc;
    },
    {} as Record<(typeof statusOrder)[number], Task[]>
  );

  const handleCreateTask = () => {
    if (!newTask.title?.trim()) return;
    onTaskCreate?.({
      ...newTask,
      status: newTask.status || "todo",
      priority: newTask.priority || "medium",
    } as Omit<Task, "id" | "createdAt" | "updatedAt">);
    setNewTask({ title: "", description: "", status: "todo", priority: "medium" });
    setShowCreateForm(false);
  };

  return (
    <div className={cn("flex h-full", className)}>
      {/* Main Kanban board */}
      <div className="flex-1 flex gap-3 overflow-x-auto p-4">
        {statusOrder.map((status) => (
          <div key={status} className="flex-shrink-0 w-72 flex flex-col">
            {/* Column header */}
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <div
                  className="size-2 rounded-full"
                  style={{ backgroundColor: statusConfig[status].color }}
                />
                <span className="text-xs font-semibold text-text-primary">
                  {statusConfig[status].label}
                </span>
                <Badge variant="ghost" size="sm">
                  {tasksByStatus[status].length}
                </Badge>
              </div>
              {status === "todo" && (
                <button
                  onClick={() => setShowCreateForm(true)}
                  className="size-5 rounded text-text-muted hover:text-text-primary hover:bg-bg-hover flex items-center justify-center text-lg"
                >
                  +
                </button>
              )}
            </div>

            {/* Tasks */}
            <div className="flex-1 space-y-2 overflow-y-auto">
              {tasksByStatus[status].map((task) => (
                <TaskCard
                  key={task.id}
                  task={task}
                  isSelected={selectedTask === task.id}
                  onSelect={() => setSelectedTask(task.id)}
                  onStatusChange={(s) => onTaskUpdate?.(task.id, { status: s as Task["status"] })}
                  onPriorityChange={(p) => onTaskUpdate?.(task.id, { priority: p as Task["priority"] })}
                  onAssign={(agentId) => onTaskAssign?.(task.id, agentId)}
                  onDelete={() => onTaskDelete?.(task.id)}
                  agents={agents}
                />
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Task detail panel */}
      {selectedTask && (
        <TaskDetailPanel
          task={tasks.find((t) => t.id === selectedTask)!}
          agents={agents}
          onClose={() => setSelectedTask(null)}
          onUpdate={(updates) => onTaskUpdate?.(selectedTask, updates)}
          onAssign={(agentId) => onTaskAssign?.(selectedTask, agentId)}
        />
      )}

      {/* Create task modal */}
      {showCreateForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-96 bg-bg-primary border border-border-subtle rounded-xl p-4 space-y-4 shadow-xl">
            <h3 className="text-sm font-semibold text-text-primary">Create New Task</h3>

            <input
              type="text"
              value={newTask.title || ""}
              onChange={(e) => setNewTask((t) => ({ ...t, title: e.target.value }))}
              placeholder="Task title..."
              className="w-full px-3 py-2 text-sm bg-bg-secondary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
            />

            <textarea
              value={newTask.description || ""}
              onChange={(e) => setNewTask((t) => ({ ...t, description: e.target.value }))}
              placeholder="Description (optional)..."
              className="w-full px-3 py-2 text-sm bg-bg-secondary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500 resize-none"
              rows={3}
            />

            <div className="flex gap-2">
              <select
                value={newTask.priority || "medium"}
                onChange={(e) => setNewTask((t) => ({ ...t, priority: e.target.value as Task["priority"] }))}
                className="flex-1 px-3 py-2 text-sm bg-bg-secondary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>

              <select
                value={newTask.status || "todo"}
                onChange={(e) => setNewTask((t) => ({ ...t, status: e.target.value as Task["status"] }))}
                className="flex-1 px-3 py-2 text-sm bg-bg-secondary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
              >
                <option value="todo">To Do</option>
                <option value="in-progress">In Progress</option>
                <option value="review">Review</option>
              </select>
            </div>

            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setShowCreateForm(false)}
                className="px-3 py-1.5 text-xs text-text-muted hover:text-text-primary"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateTask}
                className="px-3 py-1.5 text-xs bg-brand-500 hover:bg-brand-600 text-white rounded-lg"
              >
                Create Task
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// --- Task Card ---
function TaskCard({
  task,
  isSelected,
  onSelect,
  onStatusChange,
  onPriorityChange,
  onAssign,
  onDelete,
  agents,
}: {
  task: Task;
  isSelected: boolean;
  onSelect: () => void;
  onStatusChange: (status: string) => void;
  onPriorityChange: (priority: string) => void;
  onAssign: (agentId: string | null) => void;
  onDelete: () => void;
  agents: AutonomousTaskBoardProps["agents"];
}) {
  return (
    <div
      className={cn(
        "p-3 bg-bg-primary border rounded-lg cursor-pointer transition-all group",
        isSelected
          ? "border-brand-500/50 shadow-lg shadow-brand-500/10"
          : "border-border-subtle hover:border-border-default"
      )}
      onClick={onSelect}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-2 mb-2">
        <h4 className="text-sm font-medium text-text-primary line-clamp-1">{task.title}</h4>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="opacity-0 group-hover:opacity-100 p-1 text-text-muted hover:text-error transition-all"
        >
          ×
        </button>
      </div>

      {/* Description */}
      {task.description && (
        <p className="text-[11px] text-text-muted line-clamp-2 mb-2">{task.description}</p>
      )}

      {/* Meta */}
      <div className="flex items-center gap-2 mb-2">
        <span
          className="px-1.5 py-0.5 text-[9px] font-medium rounded"
          style={{
            backgroundColor: priorityConfig[task.priority].color + "20",
            color: priorityConfig[task.priority].color,
          }}
        >
          {priorityConfig[task.priority].label}
        </span>

        {task.dueDate && (
          <span className="text-[10px] text-text-muted">
            Due {new Date(task.dueDate).toLocaleDateString()}
          </span>
        )}
      </div>

      {/* Assigned agent */}
      <div className="flex items-center justify-between">
        {task.assignedAgent ? (
          <div className="flex items-center gap-1.5">
            <div
              className={cn(
                "size-5 rounded-full flex items-center justify-center text-[9px] font-bold",
                task.assignedAgent.isAI ? "bg-purple-500/20 text-purple-400" : "bg-blue-500/20 text-blue-400"
              )}
            >
              {task.assignedAgent.name[0]}
            </div>
            <span className="text-[10px] text-text-muted truncate">{task.assignedAgent.name}</span>
          </div>
        ) : (
          <select
            onChange={(e) => {
              e.stopPropagation();
              onAssign(e.target.value || null);
            }}
            className="text-[10px] bg-bg-secondary border border-border-subtle rounded px-1 py-0.5 text-text-muted focus:outline-none focus:border-brand-500"
            value=""
          >
            <option value="">Assign...</option>
            {agents.map((a) => (
              <option key={a.id} value={a.id}>
                {a.name}
              </option>
            ))}
          </select>
        )}

        {/* Subtask count */}
        {task.subtasks && task.subtasks.length > 0 && (
          <span className="text-[10px] text-text-muted">
            {task.subtasks.filter((s) => s.status === "done").length}/{task.subtasks.length}
          </span>
        )}
      </div>

      {/* Execution trace indicator */}
      {task.executionTrace && task.executionTrace.length > 0 && (
        <div className="flex items-center gap-1 mt-2 pt-2 border-t border-border-subtle">
          <div
            className={cn(
              "size-1.5 rounded-full",
              task.executionTrace[task.executionTrace.length - 1]?.result === "success"
                ? "bg-success"
                : task.executionTrace[task.executionTrace.length - 1]?.result === "failure"
                ? "bg-error"
                : "bg-warning"
            )}
          />
          <span className="text-[9px] text-text-muted">
            {task.executionTrace.length} actions
          </span>
        </div>
      )}
    </div>
  );
}

// --- Task Detail Panel ---
function TaskDetailPanel({
  task,
  agents,
  onClose,
  onUpdate,
  onAssign,
}: {
  task: Task;
  agents: AutonomousTaskBoardProps["agents"];
  onClose: () => void;
  onUpdate: (updates: Partial<Task>) => void;
  onAssign: (agentId: string | null) => void;
}) {
  const [activeTab, setActiveTab] = React.useState<"details" | "trace">("details");

  return (
    <div className="w-80 border-l border-border-subtle bg-bg-secondary flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-border-subtle">
        <h3 className="text-sm font-semibold text-text-primary">Task Details</h3>
        <button
          onClick={onClose}
          className="p-1 text-text-muted hover:text-text-primary"
        >
          ×
        </button>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-border-subtle">
        {(["details", "trace"] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={cn(
              "flex-1 py-2 text-xs font-medium capitalize transition-colors",
              activeTab === tab
                ? "text-brand-400 border-b-2 border-brand-500"
                : "text-text-muted hover:text-text-secondary"
            )}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {activeTab === "details" ? (
          <div className="space-y-4">
            <input
              value={task.title}
              onChange={(e) => onUpdate({ title: e.target.value })}
              className="w-full px-3 py-2 text-sm font-medium bg-transparent border-b border-border-subtle text-text-primary focus:outline-none focus:border-brand-500"
            />

            <textarea
              value={task.description}
              onChange={(e) => onUpdate({ description: e.target.value })}
              className="w-full px-3 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-secondary focus:outline-none focus:border-brand-500 resize-none"
              rows={4}
            />

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-[10px] text-text-muted uppercase">Status</label>
                <select
                  value={task.status}
                  onChange={(e) => onUpdate({ status: e.target.value as Task["status"] })}
                  className="w-full mt-1 px-2 py-1.5 text-xs bg-bg-primary border border-border-subtle rounded text-text-primary focus:outline-none focus:border-brand-500"
                >
                  {statusOrder.map((s) => (
                    <option key={s} value={s}>
                      {statusConfig[s].label}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-[10px] text-text-muted uppercase">Priority</label>
                <select
                  value={task.priority}
                  onChange={(e) => onUpdate({ priority: e.target.value as Task["priority"] })}
                  className="w-full mt-1 px-2 py-1.5 text-xs bg-bg-primary border border-border-subtle rounded text-text-primary focus:outline-none focus:border-brand-500"
                >
                  {Object.entries(priorityConfig).map(([k, v]) => (
                    <option key={k} value={k}>
                      {v.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Assignment */}
            <div>
              <label className="text-[10px] text-text-muted uppercase">Assigned Agent</label>
              <select
                value={task.assignedAgent?.id || ""}
                onChange={(e) => {
                  const agent = agents.find((a) => a.id === e.target.value);
                  onAssign(agent ? e.target.value : null);
                }}
                className="w-full mt-1 px-2 py-1.5 text-xs bg-bg-primary border border-border-subtle rounded text-text-primary focus:outline-none focus:border-brand-500"
              >
                <option value="">Unassigned</option>
                {agents.map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.name} {a.isAI ? "(AI)" : ""}
                  </option>
                ))}
              </select>
            </div>

            {/* Tags */}
            {task.tags && task.tags.length > 0 && (
              <div>
                <label className="text-[10px] text-text-muted uppercase">Tags</label>
                <div className="flex flex-wrap gap-1 mt-1">
                  {task.tags.map((tag) => (
                    <span key={tag} className="px-1.5 py-0.5 text-[10px] bg-bg-tertiary text-text-muted rounded">
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Subtasks */}
            {task.subtasks && task.subtasks.length > 0 && (
              <div>
                <label className="text-[10px] text-text-muted uppercase">
                  Subtasks ({task.subtasks.filter((s) => s.status === "done").length}/{task.subtasks.length})
                </label>
                <div className="mt-1 space-y-1">
                  {task.subtasks.map((subtask) => (
                    <div
                      key={subtask.id}
                      className="flex items-center gap-2 p-2 bg-bg-primary rounded border border-border-subtle"
                    >
                      <input
                        type="checkbox"
                        checked={subtask.status === "done"}
                        onChange={() => {}}
                        className="size-3"
                      />
                      <span className="text-xs text-text-secondary">{subtask.title}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-2">
            {task.executionTrace && task.executionTrace.length > 0 ? (
              task.executionTrace.map((trace, i) => (
                <div key={trace.id} className="flex items-start gap-2 p-2 bg-bg-primary rounded">
                  <div
                    className={cn(
                      "size-1.5 rounded-full mt-1.5",
                      trace.result === "success"
                        ? "bg-success"
                        : trace.result === "failure"
                        ? "bg-error"
                        : "bg-warning"
                    )}
                  />
                  <div className="flex-1 min-w-0">
                    <div className="text-xs text-text-primary">{trace.action}</div>
                    <div className="text-[10px] text-text-muted">
                      {trace.agent && <span>{trace.agent} · </span>}
                      {new Date(trace.timestamp).toLocaleTimeString()}
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8 text-text-muted text-xs">
                No execution trace yet
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}