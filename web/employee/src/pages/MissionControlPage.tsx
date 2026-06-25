import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { missionControlApi } from '@/api/mission_control';
import { Gauge, Users, FolderKanban, CheckSquare, Lightbulb, RefreshCw, TrendingUp, TrendingDown, Clock, BookOpen, Flame, CalendarDays } from 'lucide-react';
import { AreaChart, Area, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

const sparkData = [
  { d: 'W1', v: 42 }, { d: 'W2', v: 48 }, { d: 'W3', v: 45 }, { d: 'W4', v: 51 },
  { d: 'W5', v: 55 }, { d: 'W6', v: 53 }, { d: 'W7', v: 58 }, { d: 'W8', v: 62 },
];

export function MissionControlPage() {
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['mission-control'],
    queryFn: () => missionControlApi.get(),
  });

  const refreshMutation = useMutation({
    mutationFn: () => missionControlApi.refresh(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['mission-control'] }),
  });

  const s = data?.data?.snapshot;

  const burnoutPct = s ? Math.round(s.avg_burnout_risk * 100) : 0;
  const burnoutColor = burnoutPct > 70 ? 'text-red-400' : burnoutPct > 40 ? 'text-yellow-400' : 'text-green-400';

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Gauge className="h-6 w-6 text-blue-400" />
          <h1 className="text-2xl font-bold">Mission Control</h1>
        </div>
        <div className="flex items-center gap-3">
          {s?.snapshot_date && (
            <span className="text-xs text-gray-500">
              Last refreshed: {new Date(s.snapshot_date).toLocaleString()}
            </span>
          )}
          <button
            onClick={() => refreshMutation.mutate()}
            disabled={refreshMutation.isPending}
            className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            <RefreshCw className={`h-4 w-4 ${refreshMutation.isPending ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : !s ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Gauge className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No mission control data available</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <div className="mb-3 flex items-center gap-2 text-gray-400">
                <Users className="h-5 w-5" />
                <span className="text-sm font-medium">Employees</span>
              </div>
              <p className="text-3xl font-bold text-blue-400">{s.total_employees}</p>
              <div className="mt-2 flex items-center gap-2 text-xs text-gray-500">
                <span>{s.active_employees} active</span>
                <span>·</span>
                <span className="text-green-400 flex items-center gap-1"><TrendingUp className="h-3 w-3" />+{s.new_hires_30d} new</span>
                <span>·</span>
                <span className="text-red-400 flex items-center gap-1"><TrendingDown className="h-3 w-3" />-{s.departures_30d} left</span>
              </div>
              <div className="mt-3 h-12">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={sparkData}>
                    <Area type="monotone" dataKey="v" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.1} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <div className="mb-3 flex items-center gap-2 text-gray-400">
                <FolderKanban className="h-5 w-5" />
                <span className="text-sm font-medium">Projects</span>
              </div>
              <p className="text-3xl font-bold text-purple-400">{s.total_projects}</p>
              <div className="mt-2 flex items-center gap-2 text-xs text-gray-500">
                <span>{s.active_projects} active</span>
                <span>·</span>
                <span className="text-green-400">{s.completed_projects_30d} completed 30d</span>
              </div>
              <div className="mt-3 h-12">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={sparkData}>
                    <Area type="monotone" dataKey="v" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.1} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <div className="mb-3 flex items-center gap-2 text-gray-400">
                <CheckSquare className="h-5 w-5" />
                <span className="text-sm font-medium">Tasks</span>
              </div>
              <p className="text-3xl font-bold text-green-400">{s.total_tasks}</p>
              <div className="mt-2 flex items-center gap-2 text-xs text-gray-500">
                <span>{s.completed_tasks_30d} completed 30d</span>
                <span>·</span>
                <span className="flex items-center gap-1"><Clock className="h-3 w-3" />{s.avg_task_completion_days.toFixed(1)}d avg</span>
              </div>
              <div className="mt-3 h-12">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={sparkData}>
                    <Area type="monotone" dataKey="v" stroke="#22c55e" fill="#22c55e" fillOpacity={0.1} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <div className="mb-3 flex items-center gap-2 text-gray-400">
                <Lightbulb className="h-5 w-5" />
                <span className="text-sm font-medium">Innovation</span>
              </div>
              <p className="text-3xl font-bold text-yellow-400">{s.innovation_grants_submitted}</p>
              <div className="mt-2 flex items-center gap-2 text-xs text-gray-500">
                <span>{s.innovation_grants_funded} funded</span>
              </div>
              <div className="mt-3 h-12">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={sparkData}>
                    <Area type="monotone" dataKey="v" stroke="#f59e0b" fill="#f59e0b" fillOpacity={0.1} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-400">New Hires (30d)</span>
                <TrendingUp className="h-4 w-4 text-green-400" />
              </div>
              <p className="mt-1 text-2xl font-bold text-gray-100">{s.new_hires_30d}</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-400">Departures (30d)</span>
                <TrendingDown className="h-4 w-4 text-red-400" />
              </div>
              <p className="mt-1 text-2xl font-bold text-gray-100">{s.departures_30d}</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-400">Task Completion Rate</span>
                <CheckSquare className="h-4 w-4 text-blue-400" />
              </div>
              <p className="mt-1 text-2xl font-bold text-gray-100">
                {s.total_tasks > 0 ? Math.round((s.completed_tasks_30d / s.total_tasks) * 100) : 0}%
              </p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-400">Learning Hours</span>
                <BookOpen className="h-4 w-4 text-purple-400" />
              </div>
              <p className="mt-1 text-2xl font-bold text-gray-100">{s.total_learning_hours.toLocaleString()}</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-400">Skill Proficiency (avg)</span>
                <TrendingUp className="h-4 w-4 text-green-400" />
              </div>
              <p className="mt-1 text-2xl font-bold text-gray-100">{(s.avg_skill_proficiency * 100).toFixed(0)}%</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-400">PTO Used (30d)</span>
                <CalendarDays className="h-4 w-4 text-cyan-400" />
              </div>
              <p className="mt-1 text-2xl font-bold text-gray-100">{s.pto_days_used_30d} days</p>
            </div>
          </div>

          <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
            <div className="mb-4 flex items-center gap-2">
              <Flame className="h-5 w-5 text-orange-400" />
              <h3 className="text-sm font-medium text-gray-300">Burnout Risk</h3>
            </div>
            <div className="flex items-center gap-6">
              <div className="relative h-32 w-32">
                <svg className="h-32 w-32 -rotate-90" viewBox="0 0 120 120">
                  <circle cx="60" cy="60" r="50" fill="none" stroke="#374151" strokeWidth="10" />
                  <circle
                    cx="60" cy="60" r="50" fill="none"
                    stroke={burnoutPct > 70 ? '#ef4444' : burnoutPct > 40 ? '#f59e0b' : '#22c55e'}
                    strokeWidth="10"
                    strokeDasharray={`${burnoutPct * 3.14} 314`}
                    strokeLinecap="round"
                  />
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                  <span className={`text-2xl font-bold ${burnoutColor}`}>{burnoutPct}%</span>
                  <span className="text-xs text-gray-500">risk</span>
                </div>
              </div>
              <div>
                <p className="text-sm text-gray-400">
                  {burnoutPct > 70 ? 'Critical — immediate intervention recommended' :
                   burnoutPct > 40 ? 'Elevated — monitor team workload closely' :
                   'Healthy — team is within sustainable limits'}
                </p>
                <p className="mt-2 text-xs text-gray-500">
                  Based on workload, overtime, PTO utilization, and velocity metrics
                </p>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
