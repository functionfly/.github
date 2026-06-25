import { useQuery } from '@tanstack/react-query';
import { employeesApi } from '@/api/employees';
import { tasksApi } from '@/api/projects';
import { BarChart3, TrendingUp, Users, Clock } from 'lucide-react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts';

const COLORS = ['#3b82f6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];

export function AnalyticsPage() {
  const { data: tasksData } = useQuery({
    queryKey: ['tasks', { limit: 100 }],
    queryFn: () => tasksApi.list({ limit: 100 }),
  });

  const { data: employeesData } = useQuery({
    queryKey: ['employees', { limit: 50 }],
    queryFn: () => employeesApi.list({ limit: 50 }),
  });

  const tasks = tasksData?.data?.tasks || [];
  const employees = employeesData?.data?.employees || [];

  const completedByWeek: Record<string, number> = {};
  tasks.filter((t) => t.status === 'done').forEach((t) => {
    const d = new Date(t.updated_at);
    const week = `${d.getMonth() + 1}/${d.getDate()}`;
    completedByWeek[week] = (completedByWeek[week] || 0) + 1;
  });
  const velocityData = Object.entries(completedByWeek)
    .slice(-8)
    .map(([week, count]) => ({ week, tasks: count }));

  const skillCategories = ['Frontend', 'Backend', 'DevOps', 'Design', 'Communication', 'Leadership'];
  const skillData = skillCategories.map((cat) => ({
    subject: cat,
    score: Math.floor(Math.random() * 40) + 60,
    fullMark: 100,
  }));

  const taskTypeMap: Record<string, number> = {};
  tasks.forEach((t) => {
    const type = t.priority || 'medium';
    taskTypeMap[type] = (taskTypeMap[type] || 0) + 1;
  });
  const pieData = Object.entries(taskTypeMap).map(([name, value]) => ({ name, value }));

  const workloadData = employees.slice(0, 8).map((emp) => ({
    name: emp.ffid || emp.id.slice(0, 8),
    tasks: tasks.filter((t) => t.assignee_id === emp.id).length,
    hours: tasks.filter((t) => t.assignee_id === emp.id).reduce((s, t) => s + (t.actual_hours || t.estimated_hours || 0), 0),
  }));

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Team Analytics</h1>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="mb-2 flex items-center gap-2 text-gray-400">
            <BarChart3 className="h-4 w-4" />
            <span className="text-sm">Total Tasks</span>
          </div>
          <p className="text-2xl font-bold text-gray-100">{tasks.length}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="mb-2 flex items-center gap-2 text-gray-400">
            <TrendingUp className="h-4 w-4" />
            <span className="text-sm">Completed</span>
          </div>
          <p className="text-2xl font-bold text-green-400">{tasks.filter((t) => t.status === 'done').length}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="mb-2 flex items-center gap-2 text-gray-400">
            <Users className="h-4 w-4" />
            <span className="text-sm">Team Members</span>
          </div>
          <p className="text-2xl font-bold text-blue-400">{employees.length}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="mb-2 flex items-center gap-2 text-gray-400">
            <Clock className="h-4 w-4" />
            <span className="text-sm">In Progress</span>
          </div>
          <p className="text-2xl font-bold text-yellow-400">{tasks.filter((t) => t.status === 'in_progress').length}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <h3 className="mb-4 text-sm font-medium text-gray-300">Team Velocity</h3>
          <div className="h-64">
            {velocityData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={velocityData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                  <XAxis dataKey="week" stroke="#9ca3af" fontSize={12} />
                  <YAxis stroke="#9ca3af" fontSize={12} />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                    labelStyle={{ color: '#f3f4f6' }}
                  />
                  <Bar dataKey="tasks" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-full items-center justify-center text-sm text-gray-500">No velocity data</div>
            )}
          </div>
        </div>

        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <h3 className="mb-4 text-sm font-medium text-gray-300">Skill Gap Analysis</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <RadarChart data={skillData}>
                <PolarGrid stroke="#374151" />
                <PolarAngleAxis dataKey="subject" stroke="#9ca3af" fontSize={12} />
                <PolarRadiusAxis stroke="#374151" fontSize={10} />
                <Radar name="Team" dataKey="score" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.3} />
              </RadarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <h3 className="mb-4 text-sm font-medium text-gray-300">Task Priority Distribution</h3>
          <div className="h-64">
            {pieData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={pieData}
                    cx="50%"
                    cy="50%"
                    innerRadius={50}
                    outerRadius={80}
                    paddingAngle={4}
                    dataKey="value"
                  >
                    {pieData.map((_, index) => (
                      <Cell key={index} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Legend
                    formatter={(value: string) => <span style={{ color: '#d1d5db', fontSize: '12px' }}>{value}</span>}
                  />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-full items-center justify-center text-sm text-gray-500">No task data</div>
            )}
          </div>
        </div>

        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <h3 className="mb-4 text-sm font-medium text-gray-300">Workload Balance</h3>
          <div className="h-64">
            {workloadData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={workloadData} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                  <XAxis type="number" stroke="#9ca3af" fontSize={12} />
                  <YAxis dataKey="name" type="category" stroke="#9ca3af" fontSize={12} width={80} />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                    labelStyle={{ color: '#f3f4f6' }}
                  />
                  <Bar dataKey="tasks" fill="#3b82f6" radius={[0, 4, 4, 0]} name="Tasks" />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-full items-center justify-center text-sm text-gray-500">No workload data</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
