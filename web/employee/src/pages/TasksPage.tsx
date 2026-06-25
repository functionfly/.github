import { useState } from 'react';
import { useTasks, useUpdateTask } from '@/hooks/useTasks';
import { Search, Filter } from 'lucide-react';
import { TASK_STATUSES } from '@/lib/constants';

const statusColumns = [
  { id: 'todo', label: 'To Do', color: 'bg-gray-500' },
  { id: 'in_progress', label: 'In Progress', color: 'bg-blue-500' },
  { id: 'in_review', label: 'In Review', color: 'bg-yellow-500' },
  { id: 'done', label: 'Done', color: 'bg-green-500' },
  { id: 'blocked', label: 'Blocked', color: 'bg-red-500' },
];

export function TasksPage() {
  const [statusFilter, setStatusFilter] = useState('');
  const { data, isLoading } = useTasks({ status: statusFilter || undefined });
  const updateTask = useUpdateTask();
  const tasks = data?.data?.tasks || [];

  const handleStatusChange = async (taskId: string, newStatus: string) => {
    await updateTask.mutateAsync({ id: taskId, data: { status: newStatus } });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">My Tasks</h1>
        <div className="flex gap-2">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
          >
            <option value="">All Statuses</option>
            {TASK_STATUSES.map((s) => (
              <option key={s} value={s}>{s.replace('_', ' ')}</option>
            ))}
          </select>
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : tasks.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <p className="text-gray-400">No tasks found</p>
        </div>
      ) : (
        <div className="space-y-2">
          {tasks.map((task) => (
            <div
              key={task.id}
              className="flex items-center gap-4 rounded-lg border border-gray-800 bg-gray-900 p-4"
            >
              <select
                value={task.status}
                onChange={(e) => handleStatusChange(task.id, e.target.value)}
                className="rounded border border-gray-700 bg-gray-800 px-2 py-1 text-xs text-gray-300"
              >
                {TASK_STATUSES.map((s) => (
                  <option key={s} value={s}>{s.replace('_', ' ')}</option>
                ))}
              </select>
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-200">{task.title}</p>
                {task.description && (
                  <p className="mt-0.5 text-xs text-gray-500">{task.description}</p>
                )}
              </div>
              <span className={`rounded px-2 py-0.5 text-xs ${
                task.priority === 'critical' ? 'bg-red-500/20 text-red-400' :
                task.priority === 'high' ? 'bg-orange-500/20 text-orange-400' :
                task.priority === 'medium' ? 'bg-blue-500/20 text-blue-400' :
                'bg-gray-500/20 text-gray-400'
              }`}>
                {task.priority}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
