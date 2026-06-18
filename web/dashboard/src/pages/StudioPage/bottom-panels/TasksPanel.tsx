import React from "react";
import { AutonomousTaskBoard } from "@functionfly/ui-agent";
import { useStudioAgents } from "@/hooks/useStudio";
import { CheckSquare, Plus, ListTodo } from "lucide-react";

interface Task {
  id: string;
  title: string;
  description: string;
  status: "todo" | "in-progress" | "done" | "blocked";
  priority: "low" | "medium" | "high";
  createdAt: string;
  updatedAt: string;
}

interface TaskActions {
  onCreate: (task: { title: string; description: string }) => void;
  onUpdate: (task: { id: string; updates: Partial<Task> }) => void;
  onDelete: (id: string) => void;
  onAssign: (task: { id: string; agentId: string | null }) => void;
}

interface TasksPanelProps {
  tasks: Task[];
  onTaskCreate: (task: { title: string; description: string }) => void;
  onTaskUpdate: (task: { id: string; updates: Partial<Task> }) => void;
  onTaskDelete: (id: string) => void;
  onTaskAssign: (task: { id: string; agentId: string | null }) => void;
}

export function TasksPanel({
  tasks,
  onTaskCreate,
  onTaskUpdate,
  onTaskDelete,
  onTaskAssign,
}: TasksPanelProps) {
  const { agents: rawAgents } = useStudioAgents();

  const agents = rawAgents.map((agent) => ({
    id: agent.id,
    name: agent.name,
    isAI: true,
  }));

  if (tasks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <div className="w-16 h-16 rounded-full bg-bg-primary flex items-center justify-center mb-4">
          <ListTodo className="size-6 text-text-muted" />
        </div>
        <h3 className="text-sm font-medium mb-2">No Tasks Yet</h3>
        <p className="text-xs text-text-muted mb-4 max-w-[240px]">
          Create tasks to track agent activities and workflow steps
        </p>
        <button
          onClick={() =>
            onTaskCreate({
              title: "Sample Task",
              description: "This is a sample task description",
            })
          }
          className="px-4 py-2 text-xs bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition-colors flex items-center gap-2"
        >
          <Plus className="size-3" />
          Create First Task
        </button>
      </div>
    );
  }

  return (
    <div className="p-3">
      <AutonomousTaskBoard
        tasks={tasks.map((t) => ({
          id: t.id,
          title: t.title,
          description: t.description,
          status: t.status,
          priority: t.priority,
          createdAt: t.createdAt,
          updatedAt: t.updatedAt,
        }))}
        agents={agents}
        onTaskCreate={(task) =>
          onTaskCreate({ title: task.title, description: task.description })
        }
        onTaskUpdate={(id, updates) =>
          onTaskUpdate({ id, updates: { ...updates } as Partial<Task> })
        }
        onTaskDelete={onTaskDelete}
        onTaskAssign={(id, agentId) =>
          onTaskAssign({ id, agentId: agentId || null })
        }
      />
    </div>
  );
}