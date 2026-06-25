import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { employeesApi } from '@/api/employees';
import { projectsApi } from '@/api/projects';
import { tasksApi } from '@/api/projects';
import { learningApi } from '@/api/learning';
import {
  FolderKanban,
  CheckSquare,
  GraduationCap,
  Bell,
  ArrowRight,
  Clock,
  AlertTriangle,
} from 'lucide-react';
import { formatDate } from '@/lib/utils';
import { useEffect } from 'react';

export function DashboardPage() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { unreadCount, fetchUnreadCount } = useNotificationStore();

  const { data: employeeData } = useQuery({
    queryKey: ['employee', 'me'],
    queryFn: () => employeesApi.list({ limit: 1 }),
  });

  const { data: tasksData } = useQuery({
    queryKey: ['tasks', { limit: 5 }],
    queryFn: () => tasksApi.list({ limit: 5 }),
  });

  const { data: projectsData } = useQuery({
    queryKey: ['projects', { status: 'active', limit: 5 }],
    queryFn: () => projectsApi.list({ status: 'active', limit: 5 }),
  });

  const { data: learningData } = useQuery({
    queryKey: ['learning', 'progress'],
    queryFn: () => learningApi.getMyProgress(),
  });

  useEffect(() => {
    fetchUnreadCount();
  }, [fetchUnreadCount]);

  const tasks = tasksData?.data?.tasks || [];
  const projects = projectsData?.data?.projects || [];
  const learning = learningData?.data?.progress || [];
  const activeTasks = tasks.filter((t) => t.status !== 'done');
  const overdueTasks = tasks.filter(
    (t) => t.due_date && new Date(t.due_date) < new Date() && t.status !== 'done'
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">
            Welcome back, {user?.name || 'Employee'}
          </h1>
          <p className="text-sm text-gray-400">Here's your workspace overview</p>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        <div
          onClick={() => navigate('/tasks')}
          className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
        >
          <div className="mb-2 flex items-center justify-between">
            <CheckSquare className="h-5 w-5 text-blue-400" />
            <ArrowRight className="h-4 w-4 text-gray-600" />
          </div>
          <p className="text-2xl font-bold text-gray-100">{activeTasks.length}</p>
          <p className="text-sm text-gray-400">Active Tasks</p>
        </div>

        <div
          onClick={() => navigate('/projects')}
          className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
        >
          <div className="mb-2 flex items-center justify-between">
            <FolderKanban className="h-5 w-5 text-green-400" />
            <ArrowRight className="h-4 w-4 text-gray-600" />
          </div>
          <p className="text-2xl font-bold text-gray-100">{projects.length}</p>
          <p className="text-sm text-gray-400">Active Projects</p>
        </div>

        <div
          onClick={() => navigate('/learning')}
          className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
        >
          <div className="mb-2 flex items-center justify-between">
            <GraduationCap className="h-5 w-5 text-purple-400" />
            <ArrowRight className="h-4 w-4 text-gray-600" />
          </div>
          <p className="text-2xl font-bold text-gray-100">{learning.length}</p>
          <p className="text-sm text-gray-400">Courses Enrolled</p>
        </div>

        <div
          onClick={() => navigate('/notifications')}
          className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
        >
          <div className="mb-2 flex items-center justify-between">
            <Bell className="h-5 w-5 text-yellow-400" />
            <ArrowRight className="h-4 w-4 text-gray-600" />
          </div>
          <p className="text-2xl font-bold text-gray-100">{unreadCount}</p>
          <p className="text-sm text-gray-400">Unread Notifications</p>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {overdueTasks.length > 0 && (
          <div className="rounded-xl border border-red-600/30 bg-red-600/10 p-4">
            <div className="mb-3 flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-red-400" />
              <h3 className="text-sm font-medium text-red-400">Overdue Tasks</h3>
            </div>
            <div className="space-y-2">
              {overdueTasks.slice(0, 3).map((task) => (
                <div key={task.id} className="flex items-center justify-between rounded bg-gray-900/50 p-2">
                  <span className="text-sm text-gray-200">{task.title}</span>
                  <span className="flex items-center gap-1 text-xs text-red-400">
                    <Clock className="h-3 w-3" />
                    {formatDate(task.due_date!)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <h3 className="mb-3 text-sm font-medium text-gray-300">Recent Tasks</h3>
          {tasks.length === 0 ? (
            <p className="py-4 text-center text-sm text-gray-500">No tasks yet</p>
          ) : (
            <div className="space-y-2">
              {tasks.slice(0, 5).map((task) => (
                <div key={task.id} className="flex items-center gap-3 rounded p-2">
                  <span
                    className={`h-2 w-2 rounded-full ${
                      task.status === 'done'
                        ? 'bg-green-500'
                        : task.status === 'in_progress'
                        ? 'bg-blue-500'
                        : task.status === 'blocked'
                        ? 'bg-red-500'
                        : 'bg-gray-500'
                    }`}
                  />
                  <span className="flex-1 text-sm text-gray-200">{task.title}</span>
                  <span className="text-xs text-gray-500 capitalize">{task.status.replace('_', ' ')}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <h3 className="mb-3 text-sm font-medium text-gray-300">Active Projects</h3>
          {projects.length === 0 ? (
            <p className="py-4 text-center text-sm text-gray-500">No active projects</p>
          ) : (
            <div className="space-y-2">
              {projects.slice(0, 5).map((project) => (
                <div
                  key={project.id}
                  onClick={() => navigate(`/projects/${project.id}`)}
                  className="flex cursor-pointer items-center justify-between rounded p-2 hover:bg-gray-800/50"
                >
                  <span className="text-sm text-gray-200">{project.name}</span>
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs ${
                      project.priority === 'critical'
                        ? 'bg-red-500/20 text-red-400'
                        : project.priority === 'high'
                        ? 'bg-orange-500/20 text-orange-400'
                        : 'bg-gray-500/20 text-gray-400'
                    }`}
                  >
                    {project.priority}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
