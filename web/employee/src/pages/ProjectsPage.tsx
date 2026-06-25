import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useProjects, useCreateProject } from '@/hooks/useProjects';
import { Plus, Search, FolderKanban, Calendar, User } from 'lucide-react';
import { formatDate } from '@/lib/utils';
import { PROJECT_STATUSES, PROJECT_PRIORITIES } from '@/lib/constants';

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  paused: 'bg-yellow-500/20 text-yellow-400',
  completed: 'bg-blue-500/20 text-blue-400',
  archived: 'bg-gray-500/20 text-gray-400',
};

const priorityColors: Record<string, string> = {
  low: 'text-gray-400',
  medium: 'text-blue-400',
  high: 'text-orange-400',
  critical: 'text-red-400',
};

export function ProjectsPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const { data, isLoading } = useProjects({ status: statusFilter || undefined });
  const createProject = useCreateProject();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');

  const projects = data?.data?.projects || [];

  const filtered = search
    ? projects.filter((p) => p.name.toLowerCase().includes(search.toLowerCase()))
    : projects;

  const handleCreate = async () => {
    if (!newName.trim()) return;
    await createProject.mutateAsync({ name: newName, description: newDesc || undefined });
    setShowCreate(false);
    setNewName('');
    setNewDesc('');
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Projects</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          New Project
        </button>
      </div>

      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search projects..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-gray-700 bg-gray-800 py-2 pl-10 pr-4 text-sm text-gray-100 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Statuses</option>
          {PROJECT_STATUSES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <FolderKanban className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No projects found</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map((project) => (
            <div
              key={project.id}
              onClick={() => navigate(`/projects/${project.id}`)}
              className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
            >
              <div className="mb-3 flex items-start justify-between">
                <h3 className="font-semibold text-gray-100">{project.name}</h3>
                <span className={`rounded-full px-2 py-0.5 text-xs ${statusColors[project.status] || ''}`}>
                  {project.status}
                </span>
              </div>
              {project.description && (
                <p className="mb-3 line-clamp-2 text-sm text-gray-400">{project.description}</p>
              )}
              <div className="flex items-center gap-4 text-xs text-gray-500">
                <span className={priorityColors[project.priority] || ''}>
                  {project.priority}
                </span>
                {project.target_date && (
                  <span className="flex items-center gap-1">
                    <Calendar className="h-3 w-3" />
                    {formatDate(project.target_date)}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Create Project</h2>
            <input
              type="text"
              placeholder="Project name"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              autoFocus
            />
            <textarea
              placeholder="Description (optional)"
              value={newDesc}
              onChange={(e) => setNewDesc(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              rows={3}
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowCreate(false)}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={!newName.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
