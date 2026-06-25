import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useProject } from '@/hooks/useProjects';
import { useTasks, useCreateTask, useUpdateTask } from '@/hooks/useTasks';
import { Plus, ArrowLeft, Calendar, Clock } from 'lucide-react';
import { formatDate } from '@/lib/utils';
import { TASK_STATUSES, TASK_PRIORITIES } from '@/lib/constants';

const columns = [
  { id: 'todo', label: 'To Do' },
  { id: 'in_progress', label: 'In Progress' },
  { id: 'in_review', label: 'In Review' },
  { id: 'done', label: 'Done' },
];

const priorityColors: Record<string, string> = {
  low: 'border-l-gray-500',
  medium: 'border-l-blue-500',
  high: 'border-l-orange-500',
  critical: 'border-l-red-500',
};

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: projectData } = useProject(id!);
  const { data: tasksData } = useTasks({ project_id: id });
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();

  const [showCreate, setShowCreate] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [newPriority, setNewPriority] = useState('medium');

  const project = projectData?.data?.project;
  const tasks = tasksData?.data?.tasks || [];

  const handleCreate = async () => {
    if (!newTitle.trim()) return;
    await createTask.mutateAsync({
      project_id: id!,
      title: newTitle,
      priority: newPriority,
    });
    setShowCreate(false);
    setNewTitle('');
  };

  const handleStatusChange = async (taskId: string, newStatus: string) => {
    await updateTask.mutateAsync({ id: taskId, data: { status: newStatus } });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <button onClick={() => navigate('/projects')} className="text-gray-400 hover:text-gray-200">
          <ArrowLeft className="h-5 w-5" />
        </button>
        <div>
          <h1 className="text-2xl font-bold">{project?.name || 'Project'}</h1>
          {project?.description && <p className="text-sm text-gray-400">{project.description}</p>}
        </div>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex gap-2">
          {project?.target_date && (
            <span className="flex items-center gap-1 text-sm text-gray-400">
              <Calendar className="h-4 w-4" />
              Due {formatDate(project.target_date)}
            </span>
          )}
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Add Task
        </button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        {columns.map((col) => {
          const colTasks = tasks.filter((t) => t.status === col.id);
          return (
            <div key={col.id} className="rounded-xl border border-gray-800 bg-gray-900/50 p-3">
              <div className="mb-3 flex items-center justify-between">
                <h3 className="text-sm font-medium text-gray-300">{col.label}</h3>
                <span className="rounded-full bg-gray-800 px-2 py-0.5 text-xs text-gray-500">
                  {colTasks.length}
                </span>
              </div>
              <div className="space-y-2">
                {colTasks.map((task) => (
                  <div
                    key={task.id}
                    className={`rounded-lg border border-gray-800 bg-gray-900 p-3 border-l-2 ${priorityColors[task.priority] || ''}`}
                  >
                    <p className="text-sm font-medium text-gray-200">{task.title}</p>
                    {task.description && (
                      <p className="mt-1 line-clamp-2 text-xs text-gray-500">{task.description}</p>
                    )}
                    <div className="mt-2 flex items-center justify-between">
                      <div className="flex gap-1">
                        {task.due_date && (
                          <span className="flex items-center gap-1 text-xs text-gray-500">
                            <Clock className="h-3 w-3" />
                            {formatDate(task.due_date)}
                          </span>
                        )}
                      </div>
                      <select
                        value={task.status}
                        onChange={(e) => handleStatusChange(task.id, e.target.value)}
                        className="rounded border border-gray-700 bg-gray-800 px-1.5 py-0.5 text-xs text-gray-300"
                      >
                        {TASK_STATUSES.map((s) => (
                          <option key={s} value={s}>{s.replace('_', ' ')}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                ))}
                {colTasks.length === 0 && (
                  <p className="py-4 text-center text-xs text-gray-600">No tasks</p>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Add Task</h2>
            <input
              type="text"
              placeholder="Task title"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              autoFocus
            />
            <select
              value={newPriority}
              onChange={(e) => setNewPriority(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              {TASK_PRIORITIES.map((p) => (
                <option key={p} value={p}>{p}</option>
              ))}
            </select>
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowCreate(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={!newTitle.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Add
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
